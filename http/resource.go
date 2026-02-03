package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/spf13/afero"

	fbErrors "github.com/futureharmony/storagebrowser/v2/errors"
	"github.com/futureharmony/storagebrowser/v2/files"
	"github.com/futureharmony/storagebrowser/v2/fileutils"
	"github.com/futureharmony/storagebrowser/v2/minio"
)

var resourceGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	// Get path from query parameter
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	// Get limit parameter to control listing size
	limitStr := r.URL.Query().Get("limit")

	// Parse limit (-1 for no limit, default 1000 for performance)
	limit := 1000
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	log.Printf("[RESOURCE] GET request path: %s, user: %s, currentScope: %s (rootPrefix: %s), storageType: %s, limit: %d",
		path, d.user.Username, d.user.CurrentScope.Name, d.user.CurrentScope.RootPrefix, d.server.StorageType, limit)

	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.requestFs,
		Path:       path,
		Modify:     d.user.Perm.Modify,
		Expand:     true, // Always expand for now, but limit controls size
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
		Content:    true,
		Limit:      limit,
	})
	if err != nil {
		return errToStatus(err), err
	}

	if file.IsDir && file.Listing != nil {
		// Apply limit to listing if specified
		if limit > 0 && len(file.Listing.Items) > limit {
			file.Listing.Items = file.Listing.Items[:limit]
			file.Listing.HasMore = true
			// Recalculate counts
			file.Listing.NumDirs = 0
			file.Listing.NumFiles = 0
			for _, item := range file.Listing.Items {
				if item.IsDir {
					file.Listing.NumDirs++
				} else {
					file.Listing.NumFiles++
				}
			}
		}

		file.Listing.Sorting = d.user.Sorting
		file.Listing.ApplySort()
		return renderJSON(w, r, file)
	}

	if checksum := r.URL.Query().Get("checksum"); checksum != "" {
		err := file.Checksum(checksum)
		if errors.Is(err, fbErrors.ErrInvalidOption) {
			return http.StatusBadRequest, nil
		} else if err != nil {
			return http.StatusInternalServerError, err
		}

		// do not waste bandwidth if we just want the checksum
		file.Content = ""
	}

	return renderJSON(w, r, file)
})

func resourceDeleteHandler(fileCache FileCache) handleFunc {
	return withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
		// Get path from query parameter and decode any URL-encoded characters
		path := decodePath(r.URL.Query().Get("path"))
		if path == "" || path == "/" || !d.user.Perm.Delete {
			return http.StatusForbidden, nil
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

		err = d.store.Share.DeleteWithPathPrefix(file.Path)
		if err != nil {
			log.Printf("WARNING: Error(s) occurred while deleting associated shares with file: %s", err)
		}

		// delete thumbnails
		err = delThumbs(r.Context(), fileCache, file)
		if err != nil {
			return errToStatus(err), err
		}

		err = d.RunHook(func() error {
			return d.requestFs.RemoveAll(path)
		}, "delete", path, "", d.user)

		if err != nil {
			return errToStatus(err), err
		}

		return http.StatusNoContent, nil
	})
}

func resourcePostHandler(fileCache FileCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		// Get path from query parameter and decode any URL-encoded characters
		path := decodePath(r.URL.Query().Get("path"))
		if path == "" {
			return http.StatusBadRequest, nil
		}

		if !d.user.Perm.Create || !d.Check(path) {
			return http.StatusForbidden, nil
		}

		// Directories creation on POST.
		if strings.HasSuffix(path, "/") {
			// For S3 filesystems, we need to handle the root directory specially since MkdirAll("/") fails
			if minio.IsS3FileSystem(d.requestFs) {
				if path != "/" {
					err := d.requestFs.MkdirAll(path, d.settings.DirMode)
					return errToStatus(err), err
				} else {
					// The root directory always exists in S3, so just return success
					return http.StatusOK, nil
				}
			} else {
				err := d.requestFs.MkdirAll(path, d.settings.DirMode)
				return errToStatus(err), err
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
		if err == nil {
			if r.URL.Query().Get("override") != "true" {
				return http.StatusConflict, nil
			}

			// Permission for overwriting the file
			if !d.user.Perm.Modify {
				return http.StatusForbidden, nil
			}

			err = delThumbs(r.Context(), fileCache, file)
			if err != nil {
				return errToStatus(err), err
			}
		}

		err = d.RunHook(func() error {
			info, writeErr := writeFile(d.requestFs, path, r.Body, d.settings.FileMode, d.settings.DirMode)
			if writeErr != nil {
				return writeErr

			}

			etag := fmt.Sprintf(`"%x%x"`, info.ModTime().UnixNano(), info.Size())
			w.Header().Set("ETag", etag)
			return nil
		}, "upload", path, "", d.user)

		if err != nil {
			_ = d.requestFs.RemoveAll(path)
		}

		return errToStatus(err), err
	})
}

var resourcePutHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	// Get path from query parameter and decode any URL-encoded characters
	path := decodePath(r.URL.Query().Get("path"))
	if path == "" {
		return http.StatusBadRequest, nil
	}

	if !d.user.Perm.Modify || !d.Check(path) {
		return http.StatusForbidden, nil
	}

	// Only allow PUT for files.
	if strings.HasSuffix(path, "/") {
		return http.StatusMethodNotAllowed, nil
	}

	exists, err := afero.Exists(d.requestFs, path)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if !exists {
		return http.StatusNotFound, nil
	}

	err = d.RunHook(func() error {
		info, writeErr := writeFile(d.requestFs, path, r.Body, d.settings.FileMode, d.settings.DirMode)
		if writeErr != nil {
			return writeErr
		}

		etag := fmt.Sprintf(`"%x%x"`, info.ModTime().UnixNano(), info.Size())
		w.Header().Set("ETag", etag)
		return nil
	}, "save", path, "", d.user)

	return errToStatus(err), err
})

func resourcePatchHandler(fileCache FileCache) handleFunc {
	return withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
		// Get path from query parameter and decode any URL-encoded characters
		src := decodePath(r.URL.Query().Get("path"))
		dst := decodePath(r.URL.Query().Get("destination"))
		action := r.URL.Query().Get("action")

		if src == "" {
			return http.StatusBadRequest, nil
		}

		if !d.Check(src) || !d.Check(dst) {
			return http.StatusForbidden, nil
		}
		if dst == "/" || src == "/" {
			return http.StatusForbidden, nil
		}

		err := checkParent(src, dst)
		if err != nil {
			return http.StatusBadRequest, err
		}

		override := r.URL.Query().Get("override") == "true"
		rename := r.URL.Query().Get("rename") == "true"
		if !override && !rename {
			if _, err = d.requestFs.Stat(dst); err == nil {
				return http.StatusConflict, nil
			}
		}
		if rename {
			dst = addVersionSuffix(dst, d.requestFs)
		}

		// Permission for overwriting the file
		if override && !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}

		err = d.RunHook(func() error {
			return patchAction(r.Context(), action, src, dst, d, fileCache)
		}, action, src, dst, d.user)

		return errToStatus(err), err
	})
}

func checkParent(src, dst string) error {
	rel, err := filepath.Rel(src, dst)
	if err != nil {
		return err
	}

	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, "../") && rel != ".." && rel != "." {
		return fbErrors.ErrSourceIsParent
	}

	return nil
}

func addVersionSuffix(source string, afs afero.Fs) string {
	counter := 1
	dir, name := path.Split(source)
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	for {
		if _, err := afs.Stat(source); err != nil {
			break
		}
		renamed := fmt.Sprintf("%s(%d)%s", base, counter, ext)
		source = path.Join(dir, renamed)
		counter++
	}

	return source
}

func writeFile(afs afero.Fs, dst string, in io.Reader, fileMode, dirMode fs.FileMode) (os.FileInfo, error) {
	dir, _ := path.Split(dst)

	// For S3 filesystems, skip MkdirAll if the directory is the root "/"
	if minio.IsS3FileSystem(afs) {
		if dir != "/" {
			err := afs.MkdirAll(dir, dirMode)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// For non-S3 filesystems, proceed as before
		err := afs.MkdirAll(dir, dirMode)
		if err != nil {
			return nil, err
		}
	}

	var file afero.File
	var err error
	// s3 does not support os.O_RDWR, so we have to use os.O_WRONLY or os.O_CREATE|os.O_TRUNC
	if minio.IsS3FileSystem(afs) {
		// For S3, get the underlying S3 filesystem to call Create
		// Use the filesystem directly since it's already wrapped with bucket/prefix
		file, err = afs.Create(dst)
	} else {
		file, err = afs.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(file, in)
	if err != nil {
		return nil, err
	}

	// Gets the info about the file.
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return info, nil
}

func delThumbs(ctx context.Context, fileCache FileCache, file *files.FileInfo) error {
	for _, previewSizeName := range PreviewSizeNames() {
		size, _ := ParsePreviewSize(previewSizeName)
		if err := fileCache.Delete(ctx, previewCacheKey(file, size)); err != nil {
			return err
		}
	}

	return nil
}

func patchAction(ctx context.Context, action, src, dst string, d *data, fileCache FileCache) error {
	switch action {
	case "copy":
		if !d.user.Perm.Create {
			return fbErrors.ErrPermissionDenied
		}

		return fileutils.Copy(d.requestFs, src, dst, d.settings.FileMode, d.settings.DirMode)
	case "rename":
		if !d.user.Perm.Rename {
			return fbErrors.ErrPermissionDenied
		}
		src = path.Clean("/" + src)
		dst = path.Clean("/" + dst)

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.requestFs,
			Path:       src,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: false,
			Checker:    d,
		})
		if err != nil {
			return err
		}

		// delete thumbnails
		err = delThumbs(ctx, fileCache, file)
		if err != nil {
			return err
		}

		return fileutils.MoveFile(d.requestFs, src, dst, d.settings.FileMode, d.settings.DirMode)
	default:
		return fmt.Errorf("unsupported action %s: %w", action, fbErrors.ErrInvalidRequestParams)
	}
}

type DiskUsageResponse struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
}

var diskUsage = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	// Get path from query parameter
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.requestFs,
		Path:       path,
		Modify:     d.user.Perm.Modify,
		Expand:     false,
		ReadHeader: false,
		Checker:    d,
		Content:    false,
	})
	if err != nil {
		return errToStatus(err), err
	}
	if minio.IsS3FileSystem(d.requestFs) {
		return renderJSON(w, r, &DiskUsageResponse{
			Total: uint64(file.Size),
			Used:  uint64(file.Size),
		})

	}
	fPath := file.RealPath()
	if !file.IsDir {
		return renderJSON(w, r, &DiskUsageResponse{
			Total: 0,
			Used:  0,
		})
	}

	usage, err := disk.UsageWithContext(r.Context(), fPath)
	if err != nil {
		return errToStatus(err), err
	}
	return renderJSON(w, r, &DiskUsageResponse{
		Total: usage.Total,
		Used:  usage.Used,
	})
})
