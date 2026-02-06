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
	"regexp"
	"strconv"
	"strings"

	"github.com/mholt/archives"
	"github.com/spf13/afero"

	"github.com/futureharmony/storagebrowser/v2/files"
	"github.com/futureharmony/storagebrowser/v2/fileutils"
	"github.com/futureharmony/storagebrowser/v2/minio"
	"github.com/futureharmony/storagebrowser/v2/users"
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

// getContentTypeForExtension returns the appropriate Content-Type for a file extension
func getContentTypeForExtension(extension string) string {
	if extension == "" {
		return "application/octet-stream"
	}

	// First try Go's built-in MIME type detection
	if mimeType := mime.TypeByExtension(extension); mimeType != "" {
		return mimeType
	}

	// Custom MIME type mappings for common video/audio formats
	// that might not be in Go's default database
	customMimeTypes := map[string]string{
		".mkv":  "video/mp4",                     // Matroska video, treat as MP4
		".webm": "video/webm",                    // WebM video
		".mp4":  "video/mp4",                     // MPEG-4 video
		".m4v":  "video/mp4",                     // MPEG-4 video (Apple variant)
		".avi":  "video/x-msvideo",               // AVI video
		".mov":  "video/quicktime",               // QuickTime video
		".qt":   "video/quicktime",               // QuickTime video
		".flv":  "video/x-flv",                   // Flash video
		".wmv":  "video/x-ms-wmv",                // Windows Media Video
		".ogg":  "video/ogg",                     // Ogg video
		".ogv":  "video/ogg",                     // Ogg video
		".mpeg": "video/mpeg",                    // MPEG video
		".mpg":  "video/mpeg",                    // MPEG video
		".mpe":  "video/mpeg",                    // MPEG video
		".m2v":  "video/mpeg",                    // MPEG-2 video
		".m1v":  "video/mpeg",                    // MPEG-1 video
		".3gp":  "video/3gpp",                    // 3GPP video
		".3g2":  "video/3gpp2",                   // 3GPP2 video
		".asf":  "video/x-ms-asf",                // Advanced Systems Format
		".swf":  "application/x-shockwave-flash", // Shockwave Flash
	}

	// Ensure extension starts with dot
	normalizedExt := strings.ToLower(extension)
	if !strings.HasPrefix(normalizedExt, ".") {
		normalizedExt = "." + normalizedExt
	}

	if mimeType, exists := customMimeTypes[normalizedExt]; exists {
		return mimeType
	}

	// Default for unknown types
	if strings.HasPrefix(normalizedExt, ".") {
		// Check if it might be a video/audio file based on common patterns
		if strings.Contains("mp4,m4v,avi,mov,flv,wmv,ogg,ogv,mkv,mpeg,mpg,mpe,3gp,3g2,asf,swf", normalizedExt[1:]) {
			return "video/mp4"
		}
	}

	return "application/octet-stream"
}

var rawHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download {
		return http.StatusAccepted, nil
	}

	// Check for scope and path query parameters (for S3 storage)
	path := r.URL.Path
	scopeParam := r.URL.Query().Get("scope")
	if scopeParam != "" && d.server.StorageType == "s3" {
		// Use scope parameter to get the path within the bucket
		pathParam := r.URL.Query().Get("path")
		if pathParam != "" {
			path = decodePath(pathParam)
		} else {
			// If no path param, strip bucket prefix from URL path
			bucketMatch := regexp.MustCompile(`^/buckets/[^/]+(.*)$`).FindStringSubmatch(path)
			if bucketMatch != nil {
				path = bucketMatch[1]
				if path == "" {
					path = "/"
				}
			}
		}
	} else {
		// For non-S3 or no scope param, strip bucket prefix if present
		if d.server.StorageType == "s3" {
			bucketMatch := regexp.MustCompile(`^/buckets/[^/]+(.*)$`).FindStringSubmatch(path)
			if bucketMatch != nil {
				path = bucketMatch[1]
				if path == "" {
					path = "/"
				}
			}
		}
	}

	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.requestFs,
		Path:       path,
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

	info, err := d.requestFs.Stat(path)
	if err != nil {
		return nil, err
	}

	var archiveFiles []archives.FileInfo

	if path != commonPath {
		nameInArchive := strings.TrimPrefix(path, commonPath)
		// Use forward slash separator for nameInArchive to be consistent across filesystems
		nameInArchive = strings.TrimPrefix(nameInArchive, "/")
		if !isS3Fs(d.requestFs) {
			nameInArchive = strings.TrimPrefix(nameInArchive, string(filepath.Separator))
		}

		archiveFiles = append(archiveFiles, archives.FileInfo{
			FileInfo:      info,
			NameInArchive: nameInArchive,
			Open: func() (fs.File, error) {
				return d.requestFs.Open(path)
			},
		})
	}

	if info.IsDir() {
		f, err := d.requestFs.Open(path)
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
			if isS3Fs(d.requestFs) {
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
	return minio.IsS3FileSystem(fs)
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

	var commonDir string
	if isS3Fs(d.requestFs) {
		// For S3 filesystems, use forward slash as separator for CommonPrefix
		commonDir = fileutils.CommonPrefix('/', filenames...)
	} else {
		// For regular filesystems, use OS-specific separator
		commonDir = fileutils.CommonPrefix(filepath.Separator, filenames...)
	}

	var allFiles []archives.FileInfo
	for _, fname := range filenames {
		// For S3 filesystems, we need to normalize the path format
		normalizedFname := fname
		if isS3Fs(d.requestFs) {
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

	// Always set Content-Type based on file extension for all files
	if file.Extension != "" {
		mimeType := getContentTypeForExtension(file.Extension)
		w.Header().Set("Content-Type", mimeType)
	}

	// Check if filesystem is S3 to handle differently since S3 files are not seekable
	// We need to get the actual filesystem from the data object in the rawHandler function
	// Since we cannot import aferos3 in files package, we check by name or type
	// We'll pass filesystem type information through the request context or check differently
	if minio.IsS3FileSystem(file.Fs) {
		// For S3 files, we can't use http.ServeContent because it requires seeking
		// Instead, we'll set headers manually and copy the content

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
