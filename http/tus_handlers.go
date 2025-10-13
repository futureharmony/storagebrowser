package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/files"
)

const maxUploadWait = 3 * time.Minute

// Tracks active uploads along with their respective upload lengths
var activeUploads = initActiveUploads()

func initActiveUploads() *ttlcache.Cache[string, int64] {
	cache := ttlcache.New[string, int64]()
	cache.OnEviction(func(_ context.Context, reason ttlcache.EvictionReason, item *ttlcache.Item[string, int64]) {
		if reason == ttlcache.EvictionReasonExpired {
			fmt.Printf("deleting incomplete upload file: \"%s\"", item.Key())
			os.Remove(item.Key())
		}
	})
	go cache.Start()

	return cache
}

func registerUpload(filePath string, fileSize int64) {
	activeUploads.Set(filePath, fileSize, maxUploadWait)
}

func completeUpload(filePath string) {
	activeUploads.Delete(filePath)
}

func getActiveUploadLength(filePath string) (int64, error) {
	item := activeUploads.Get(filePath)
	if item == nil {
		return 0, fmt.Errorf("no active upload found for the given path")
	}

	return item.Value(), nil
}

func keepUploadActive(filePath string) func() {
	stop := make(chan bool)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				activeUploads.Touch(filePath)
			}
		}
	}()

	return func() {
		close(stop)
	}
}

func tusPostHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Create || !d.Check(r.URL.Path) {
			return http.StatusForbidden, nil
		}
		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       r.URL.Path,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		switch {
		case errors.Is(err, afero.ErrFileNotFound):
			dirPath := filepath.Dir(r.URL.Path)
			if _, statErr := d.user.Fs.Stat(dirPath); os.IsNotExist(statErr) {
				if mkdirErr := d.user.Fs.MkdirAll(dirPath, d.settings.DirMode); mkdirErr != nil {
					return http.StatusInternalServerError, err
				}
			}
		case err != nil:
			return errToStatus(err), err
		}

		fileFlags := os.O_CREATE | os.O_WRONLY

		// if file exists
		if file != nil {
			if file.IsDir {
				return http.StatusBadRequest, fmt.Errorf("cannot upload to a directory %s", file.RealPath())
			}

			// Existing files will remain untouched unless explicitly instructed to override
			if r.URL.Query().Get("override") != "true" {
				return http.StatusConflict, nil
			}

			// Permission for overwriting the file
			if !d.user.Perm.Modify {
				return http.StatusForbidden, nil
			}

			fileFlags |= os.O_TRUNC
		}

		openFile, err := d.user.Fs.OpenFile(r.URL.Path, fileFlags, d.settings.FileMode)
		if err != nil {
			return errToStatus(err), err
		}
		defer openFile.Close()

		// For afero-s3 compatibility, we need to handle the case where the file
		// may not be immediately visible after creation
		file, err = files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       r.URL.Path,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: false,
			Checker:    d,
			Content:    false,
		})
		if err != nil {
			// If the file was just created but not visible yet (common with S3),
			// we can still proceed by creating a basic file info structure
			if errors.Is(err, afero.ErrFileNotFound) {
				// Create a basic file info for the new file
				file = &files.FileInfo{
					Fs:        d.user.Fs,
					Path:      r.URL.Path,
					Name:      filepath.Base(r.URL.Path),
					IsDir:     false,
					Mode:      d.settings.FileMode,
					Size:      0, // New file size is 0
					ModTime:   time.Now(),
					Extension: filepath.Ext(r.URL.Path),
				}
			} else {
				return errToStatus(err), err
			}
		}

		uploadLength, err := getUploadLength(r)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid upload length: %w", err)
		}

		// Enables the user to utilize the PATCH endpoint for uploading file data
		registerUpload(file.RealPath(), uploadLength)

		path, err := url.JoinPath("/", d.server.BaseURL, "/api/tus", r.URL.Path)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid path: %w", err)
		}

		w.Header().Set("Location", path)
		return http.StatusCreated, nil
	})
}

func tusHeadHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		w.Header().Set("Cache-Control", "no-store")
		if !d.user.Perm.Create || !d.Check(r.URL.Path) {
			return http.StatusForbidden, nil
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

		uploadLength, err := getActiveUploadLength(file.RealPath())
		if err != nil {
			return http.StatusNotFound, err
		}

		w.Header().Set("Upload-Offset", strconv.FormatInt(file.Size, 10))
		w.Header().Set("Upload-Length", strconv.FormatInt(uploadLength, 10))

		return http.StatusOK, nil
	})
}

func tusPatchHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Create || !d.Check(r.URL.Path) {
			return http.StatusForbidden, nil
		}
		if r.Header.Get("Content-Type") != "application/offset+octet-stream" {
			return http.StatusUnsupportedMediaType, nil
		}

		uploadOffset, err := getUploadOffset(r)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid upload offset")
		}

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       r.URL.Path,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})

		switch {
		case errors.Is(err, afero.ErrFileNotFound):
			return http.StatusNotFound, nil
		case err != nil:
			return errToStatus(err), err
		}

		uploadLength, err := getActiveUploadLength(file.RealPath())
		if err != nil {
			return http.StatusNotFound, err
		}

		// Prevent the upload from being evicted during the transfer
		stop := keepUploadActive(file.RealPath())
		defer stop()

		switch {
		case file.IsDir:
			return http.StatusBadRequest, fmt.Errorf("cannot upload to a directory %s", file.RealPath())
		case file.Size != uploadOffset:
			return http.StatusConflict, fmt.Errorf(
				"%s file size doesn't match the provided offset: %d",
				file.RealPath(),
				uploadOffset,
			)
		}

		// Check if we're dealing with an S3-like filesystem that doesn't support
		// the traditional file operations needed for tus protocol
		// Instead of using O_APPEND (which S3 doesn't support), we'll try to detect
		// S3 filesystems and handle the upload differently
		openFile, err := d.user.Fs.OpenFile(r.URL.Path, os.O_WRONLY|os.O_APPEND, d.settings.FileMode)
		if err != nil {
			// If O_APPEND failed, it might be an S3 filesystem, try different approach
			if err.Error() == "s3 doesn't support this operation" || 
				err.Error() == "not implemented" || 
				strings.Contains(err.Error(), "not support") {
				
				// For S3, we can't do traditional tus chunked uploads.
				// Instead, for the first chunk (when offset=0), we upload the entire file content
				if uploadOffset == 0 {
					// Read the entire request body into memory
					bodyBytes, err := io.ReadAll(r.Body)
					if err != nil {
						return http.StatusInternalServerError, fmt.Errorf("could not read request body: %w", err)
					}
					
					// Write the entire content to the file
					err = afero.WriteFile(d.user.Fs, r.URL.Path, bodyBytes, d.settings.FileMode)
					if err != nil {
						return http.StatusInternalServerError, fmt.Errorf("could not write to file: %w", err)
					}
					
					// Update the file info after write
					file, err = files.NewFileInfo(&files.FileOptions{
						Fs:         d.user.Fs,
						Path:       r.URL.Path,
						Modify:     d.user.Perm.Modify,
						Expand:     false,
						ReadHeader: false,
						Checker:    d,
						Content:    false,
					})
					if err != nil {
						return errToStatus(err), err
					}
					
					newOffset := int64(len(bodyBytes))
					w.Header().Set("Upload-Offset", strconv.FormatInt(newOffset, 10))

					if newOffset >= uploadLength {
						completeUpload(file.RealPath())
						_ = d.RunHook(func() error { return nil }, "upload", r.URL.Path, "", d.user)
					}

					return http.StatusNoContent, nil
				} else {
					// S3 doesn't properly support chunked uploads beyond the first chunk
					return http.StatusNotImplemented, fmt.Errorf("S3 filesystem does not support chunked uploads beyond first chunk")
				}
			} else {
				return errToStatus(err), err
			}
		} else {
			// Traditional filesystem approach
			defer openFile.Close()
			
			// For S3 filesystems that don't support seeking, write directly without seeking
			// Since we've validated that the offset matches the file size, we can append directly
			defer r.Body.Close()
			bytesWritten, err := io.Copy(openFile, r.Body)
			if err != nil {
				return http.StatusInternalServerError, fmt.Errorf("could not write to file: %w", err)
			}

			newOffset := uploadOffset + bytesWritten
			w.Header().Set("Upload-Offset", strconv.FormatInt(newOffset, 10))

			if newOffset >= uploadLength {
				completeUpload(file.RealPath())
				_ = d.RunHook(func() error { return nil }, "upload", r.URL.Path, "", d.user)
			}

			return http.StatusNoContent, nil
		}
	})
}

func tusDeleteHandler() handleFunc {
	return withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.URL.Path == "/" || !d.user.Perm.Create {
			return http.StatusForbidden, nil
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

		_, err = getActiveUploadLength(file.RealPath())
		if err != nil {
			return http.StatusNotFound, err
		}

		err = d.user.Fs.RemoveAll(r.URL.Path)
		if err != nil {
			return errToStatus(err), err
		}

		completeUpload(file.RealPath())

		return http.StatusNoContent, nil
	})
}

func getUploadLength(r *http.Request) (int64, error) {
	uploadOffset, err := strconv.ParseInt(r.Header.Get("Upload-Length"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid upload length: %w", err)
	}
	return uploadOffset, nil
}

func getUploadOffset(r *http.Request) (int64, error) {
	uploadOffset, err := strconv.ParseInt(r.Header.Get("Upload-Offset"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid upload offset: %w", err)
	}
	return uploadOffset, nil
}
