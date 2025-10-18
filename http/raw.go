package http

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	gopath "path"
	"path/filepath"
	"strconv"
	"strings"

	aferos3 "github.com/futureharmony/afero-aws-s3"
	"github.com/mholt/archives"
	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/fileutils"
	"github.com/filebrowser/filebrowser/v2/users"
)

func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return gopath.Clean(name)
}

func parseQueryFiles(r *http.Request, f *files.FileInfo, _ *users.User) ([]string, error) {
	var fileSlice []string
	names := strings.Split(r.URL.Query().Get("files"), ",")

	if len(names) == 0 {
		fileSlice = append(fileSlice, f.Path)
	} else {
		for _, name := range names {
			name, err := url.QueryUnescape(strings.Replace(name, "+", "%2B", -1)) //nolint:govet
			if err != nil {
				return nil, err
			}

			name = slashClean(name)
			fileSlice = append(fileSlice, filepath.Join(f.Path, name))
		}
	}

	return fileSlice, nil
}

func parseQueryAlgorithm(r *http.Request) (string, archives.Archival, error) {
	switch r.URL.Query().Get("algo") {
	case "zip", "true", "":
		return ".zip", archives.Zip{}, nil
	case "tar":
		return ".tar", archives.Tar{}, nil
	case "targz":
		return ".tar.gz", archives.CompressedArchive{Compression: archives.Gz{}, Archival: archives.Tar{}}, nil
	case "tarbz2":
		return ".tar.bz2", archives.CompressedArchive{Compression: archives.Bz2{}, Archival: archives.Tar{}}, nil
	case "tarxz":
		return ".tar.xz", archives.CompressedArchive{Compression: archives.Xz{}, Archival: archives.Tar{}}, nil
	case "tarlz4":
		return ".tar.lz4", archives.CompressedArchive{Compression: archives.Lz4{}, Archival: archives.Tar{}}, nil
	case "tarsz":
		return ".tar.sz", archives.CompressedArchive{Compression: archives.Sz{}, Archival: archives.Tar{}}, nil
	case "tarbr":
		return ".tar.br", archives.CompressedArchive{Compression: archives.Brotli{}, Archival: archives.Tar{}}, nil
	case "tarzst":
		return ".tar.zst", archives.CompressedArchive{Compression: archives.Zstd{}, Archival: archives.Tar{}}, nil
	default:
		return "", nil, errors.New("format not implemented")
	}
}

func setContentDisposition(w http.ResponseWriter, r *http.Request, file *files.FileInfo) {
	if r.URL.Query().Get("inline") == "true" {
		w.Header().Set("Content-Disposition", "inline")
	} else {
		// As per RFC6266 section 4.3
		w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(file.Name))
	}
}

var rawHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download {
		return http.StatusAccepted, nil
	}

	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.user.Fs,
		Path:       r.URL.Path,
		Modify:     d.user.Perm.Modify,
		Expand:     false,
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
	})
	if err != nil {
		return errToStatus(err), err
	}

	if files.IsNamedPipe(file.Mode) {
		setContentDisposition(w, r, file)
		return 0, nil
	}

	if !file.IsDir {
		return rawFileHandler(w, r, file)
	}

	return rawDirHandler(w, r, d, file)
})

func getFiles(d *data, path, commonPath string) ([]archives.FileInfo, error) {
	if !d.Check(path) {
		return nil, nil
	}

	info, err := d.user.Fs.Stat(path)
	if err != nil {
		return nil, err
	}

	var archiveFiles []archives.FileInfo

	if path != commonPath {
		nameInArchive := strings.TrimPrefix(path, commonPath)
		// Use forward slash separator for nameInArchive to be consistent across filesystems
		nameInArchive = strings.TrimPrefix(nameInArchive, "/")
		if !isS3Fs(d.user.Fs) {
			nameInArchive = strings.TrimPrefix(nameInArchive, string(filepath.Separator))
		}

		archiveFiles = append(archiveFiles, archives.FileInfo{
			FileInfo:      info,
			NameInArchive: nameInArchive,
			Open: func() (fs.File, error) {
				return d.user.Fs.Open(path)
			},
		})
	}

	if info.IsDir() {
		f, err := d.user.Fs.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		names, err := f.Readdirnames(0)
		if err != nil {
			return nil, err
		}

		for _, name := range names {
			var fPath string
			if isS3Fs(d.user.Fs) {
				// For S3 filesystems, use forward slash path separators
				fPath = gopath.Join(path, name)
			} else {
				// For regular filesystems, use OS-specific separators
				fPath = filepath.Join(path, name)
			}
			
			subFiles, err := getFiles(d, fPath, commonPath)
			if err != nil {
				log.Printf("Failed to get files from %s: %v", fPath, err)
				continue
			}
			archiveFiles = append(archiveFiles, subFiles...)
		}
	}

	return archiveFiles, nil
}

// Helper function to check if the filesystem is an S3 filesystem
func isS3Fs(fs afero.Fs) bool {
	_, ok := fs.(*aferos3.Fs)
	return ok
}

func rawDirHandler(w http.ResponseWriter, r *http.Request, d *data, file *files.FileInfo) (int, error) {
	filenames, err := parseQueryFiles(r, file, d.user)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	extension, archiver, err := parseQueryAlgorithm(r)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	commonDir := fileutils.CommonPrefix(filepath.Separator, filenames...)

	var allFiles []archives.FileInfo
	for _, fname := range filenames {
		// For S3 filesystems, we need to normalize the path format
		normalizedFname := fname
		if isS3Fs(d.user.Fs) {
			// Convert OS-specific paths to S3-compatible paths with forward slashes
			normalizedFname = gopath.Clean(strings.ReplaceAll(fname, string(filepath.Separator), "/"))
			// Ensure it starts with a forward slash for S3
			if normalizedFname != "" && normalizedFname[0] != '/' {
				normalizedFname = "/" + normalizedFname
			}
		}
		
		archiveFiles, err := getFiles(d, normalizedFname, commonDir)
		if err != nil {
			log.Printf("Failed to get files from %s: %v", normalizedFname, err)
			continue
		}
		allFiles = append(allFiles, archiveFiles...)
	}

	name := filepath.Base(commonDir)
	if name == "." || name == "" || name == string(filepath.Separator) {
		if file.Name != "" {
			name = file.Name
		} else {
			// For S3 filesystems, we can't use "." as a path for Stat operation
			// Instead, we should get the name from the root path directly
			var actual os.FileInfo
			var statErr error
			
			if isS3Fs(file.Fs) {
				// For S3, use the current directory path instead of "."
				actual, statErr = file.Fs.Stat(file.Path)
			} else {
				// For regular filesystems, continue with the original logic
				actual, statErr = file.Fs.Stat(".")
			}
			
			if statErr != nil {
				return http.StatusInternalServerError, statErr
			}
			name = actual.Name()
		}
	}
	if len(filenames) > 1 {
		name = "_" + name
	}
	name += extension
	w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(name))

	if err := archiver.Archive(r.Context(), w, allFiles); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

func rawFileHandler(w http.ResponseWriter, r *http.Request, file *files.FileInfo) (int, error) {
	fd, err := file.Fs.Open(file.Path)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer fd.Close()

	setContentDisposition(w, r, file)
	w.Header().Add("Content-Security-Policy", `script-src 'none';`)
	w.Header().Set("Cache-Control", "private")

	// Check if filesystem is S3 to handle differently since S3 files are not seekable
	// We need to get the actual filesystem from the data object in the rawHandler function
	// Since we cannot import aferos3 in files package, we check by name or type
	// We'll pass filesystem type information through the request context or check differently
	if _, ok := file.Fs.(*aferos3.Fs); ok {
		// For S3 files, we can't use http.ServeContent because it requires seeking
		// Instead, we'll set headers manually and copy the content
		w.Header().Set("Content-Type", "application/octet-stream") // Default to binary
		if file.Extension != "" {
			if mimeType := mime.TypeByExtension(file.Extension); mimeType != "" {
				w.Header().Set("Content-Type", mimeType)
			}
		}

		// Set content length if available
		if file.Size > 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
		}

		// Copy the file content to response
		_, err := io.Copy(w, fd)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		return 0, nil
	}

	http.ServeContent(w, r, file.Name, file.ModTime, fd)
	return 0, nil
}
