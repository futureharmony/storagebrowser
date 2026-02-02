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
	"time"

	aferos3 "github.com/futureharmony/afero-aws-s3"
	"github.com/jellydator/ttlcache/v3"
	"github.com/spf13/afero"

	"github.com/futureharmony/storagebrowser/v2/files"
	"github.com/futureharmony/storagebrowser/v2/minio"
)

const maxUploadWait = 3 * time.Minute

// UploadState tracks the state of an active upload
type UploadState struct {
	UploadLength int64
	UploadID     string                  // For S3 multipart uploads
	Parts        []aferos3.CompletedPart // For S3 multipart uploads
}

// Tracks active uploads along with their respective upload lengths
var activeUploads = initActiveUploads()

func initActiveUploads() *ttlcache.Cache[string, *UploadState] {
	cache := ttlcache.New[string, *UploadState]()
	cache.OnEviction(func(_ context.Context, reason ttlcache.EvictionReason, item *ttlcache.Item[string, *UploadState]) {
		if reason == ttlcache.EvictionReasonExpired {
			state := item.Value()
			if state != nil && state.UploadID != "" {
				// For S3, abort the multipart upload
				fmt.Printf("aborting incomplete S3 multipart upload: \"%s\" (uploadID: %s)", item.Key(), state.UploadID)
				// Note: Aborting multipart upload requires the bucket and key
				// For now, we just log it. In a full implementation, we'd store bucket/key or abort here.
			} else {
				fmt.Printf("deleting incomplete upload file: \"%s\"", item.Key())
				os.Remove(item.Key())
			}
		}
	})
	go cache.Start()

	return cache
}

func registerUpload(filePath string, fileSize int64) {
	state := &UploadState{
		UploadLength: fileSize,
		Parts:        make([]aferos3.CompletedPart, 0),
	}
	activeUploads.Set(filePath, state, maxUploadWait)
}

func completeUpload(filePath string) {
	activeUploads.Delete(filePath)
}

func getActiveUploadLength(filePath string) (int64, error) {
	item := activeUploads.Get(filePath)
	if item == nil {
		return 0, fmt.Errorf("no active upload found for the given path")
	}

	return item.Value().UploadLength, nil
}

func getUploadState(filePath string) (*UploadState, error) {
	item := activeUploads.Get(filePath)
	if item == nil {
		return nil, fmt.Errorf("no active upload found for the given path")
	}

	return item.Value(), nil
}

func updateUploadState(filePath string, state *UploadState) {
	activeUploads.Set(filePath, state, maxUploadWait)
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
		// Get path from query parameter
		path := r.URL.Query().Get("path")
		if path == "" {
			return http.StatusBadRequest, nil
		}

		if !d.user.Perm.Create || !d.Check(path) {
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
		switch {
		case errors.Is(err, afero.ErrFileNotFound):
			// s3 file system no need to create parent directories
			// afero-s3 handles this internally
			// for other file systems, we can create parent directories if needed
			// Uncomment the following lines if you want to create parent directories

			// dirPath := filepath.Dir(r.URL.Path)
			// if _, statErr := d.requestFs.Stat(dirPath); os.IsNotExist(statErr) {
			// 	if mkdirErr := d.requestFs.MkdirAll(dirPath, d.settings.DirMode); mkdirErr != nil {
			// 		return http.StatusInternalServerError, err
			// 	}
			// }
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

		openFile, err := d.requestFs.OpenFile(path, fileFlags, d.settings.FileMode)
		if err != nil {
			return errToStatus(err), err
		}
		defer openFile.Close()

		// For afero-s3 compatibility, we need to handle the case where the file
		// may not be immediately visible after creation
		file, err = files.NewFileInfo(&files.FileOptions{
			Fs:         d.requestFs,
			Path:       path,
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
					Fs:        d.requestFs,
					Path:      path,
					Name:      filepath.Base(path),
					IsDir:     false,
					Mode:      d.settings.FileMode,
					Size:      0, // New file size is 0
					ModTime:   time.Now(),
					Extension: filepath.Ext(path),
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

		// Check if it's an S3 filesystem to handle it differently
		if s3wrapper, ok := d.requestFs.(*aferos3.FsWrapper); ok {
			// Initiate multipart upload for S3
			uploadID, initErr := s3wrapper.InitiateMultipartUpload(path)
			if initErr != nil {
				return http.StatusInternalServerError, fmt.Errorf("failed to initiate multipart upload: %w", initErr)
			}

			// Update upload state with upload ID
			state, _ := getUploadState(file.RealPath())
			if state != nil {
				state.UploadID = uploadID
				updateUploadState(file.RealPath(), state)
			}
		}

		locationPath, err := url.JoinPath("/", d.server.BaseURL, "/api/tus", path)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid path: %w", err)
		}

		w.Header().Set("Location", locationPath)
		return http.StatusCreated, nil
	})
}

func tusHeadHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		w.Header().Set("Cache-Control", "no-store")

		// Get path from query parameter
		path := r.URL.Query().Get("path")
		if path == "" {
			return http.StatusBadRequest, nil
		}

		if !d.user.Perm.Create || !d.Check(path) {
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

		uploadLength, err := getActiveUploadLength(file.RealPath())
		if err != nil {
			return http.StatusNotFound, err
		}

		// Check if S3
		offset := file.Size
		if minio.IsS3FileSystem(d.requestFs) {
			state, err := getUploadState(file.RealPath())
			if err == nil && state != nil {
				offset = 0
				for _, part := range state.Parts {
					offset += part.Size
				}
			}
		}

		w.Header().Set("Upload-Offset", strconv.FormatInt(offset, 10))
		w.Header().Set("Upload-Length", strconv.FormatInt(uploadLength, 10))

		return http.StatusOK, nil
	})
}

func tusPatchHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		// Get path from query parameter
		path := r.URL.Query().Get("path")
		if path == "" {
			return http.StatusBadRequest, nil
		}

		if !d.user.Perm.Create || !d.Check(path) {
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
			Fs:         d.requestFs,
			Path:       path,
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

		if file.IsDir {
			return http.StatusBadRequest, fmt.Errorf("cannot upload to a directory %s", file.RealPath())
		}

		// Check if it's an S3 filesystem to handle it differently
		if s3wrapper, ok := d.requestFs.(*aferos3.FsWrapper); ok {
			// Handle S3 multipart upload
			state, err := getUploadState(file.RealPath())
			if err != nil || state == nil || state.UploadID == "" {
				return http.StatusNotFound, fmt.Errorf("no active S3 multipart upload found")
			}

			// Calculate current offset from uploaded parts
			currentOffset := int64(0)
			for _, part := range state.Parts {
				currentOffset += part.Size
			}

			// Check if upload offset matches current offset
			if currentOffset != uploadOffset {
				return http.StatusConflict, fmt.Errorf(
					"%s upload offset doesn't match current offset: expected %d, got %d",
					file.RealPath(),
					currentOffset,
					uploadOffset,
				)
			}

			// Read the request body
			bodyBytes, readErr := io.ReadAll(r.Body)
			if readErr != nil {
				return http.StatusInternalServerError, fmt.Errorf("could not read request body: %w", readErr)
			}

			// Upload part
			partNumber := int32(len(state.Parts) + 1) // #nosec G115 -- number of parts is small
			etag, uploadErr := s3wrapper.UploadPart(path, state.UploadID, partNumber, bodyBytes)
			if uploadErr != nil {
				return http.StatusInternalServerError, fmt.Errorf("could not upload part: %w", uploadErr)
			}

			// Add part to state
			state.Parts = append(state.Parts, aferos3.CompletedPart{
				PartNumber: partNumber,
				ETag:       etag,
				Size:       int64(len(bodyBytes)),
			})
			updateUploadState(file.RealPath(), state)

			// Calculate new offset
			newOffset := uploadOffset + int64(len(bodyBytes))
			w.Header().Set("Upload-Offset", strconv.FormatInt(newOffset, 10))

			if newOffset >= uploadLength {
				// Complete the multipart upload
				err = s3wrapper.CompleteMultipartUpload(path, state.UploadID, state.Parts)
				if err != nil {
					return http.StatusInternalServerError, fmt.Errorf("could not complete multipart upload: %w", err)
				}

				completeUpload(file.RealPath())
				_ = d.RunHook(func() error { return nil }, "upload", path, "", d.user)
			}

			return http.StatusNoContent, nil
		}

		// Traditional filesystem approach
		if file.Size != uploadOffset {
			return http.StatusConflict, fmt.Errorf(
				"%s file size doesn't match the provided offset: %d",
				file.RealPath(),
				uploadOffset,
			)
		}

		openFile, err := d.requestFs.OpenFile(path, os.O_WRONLY|os.O_APPEND, d.settings.FileMode)
		if err != nil {
			return errToStatus(err), err
		}
		defer openFile.Close()

		defer r.Body.Close()
		bytesWritten, err := io.Copy(openFile, r.Body)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("could not write to file: %w", err)
		}

		newOffset := uploadOffset + bytesWritten
		w.Header().Set("Upload-Offset", strconv.FormatInt(newOffset, 10))

		if newOffset >= uploadLength {
			completeUpload(file.RealPath())
			_ = d.RunHook(func() error { return nil }, "upload", path, "", d.user)
		}

		return http.StatusNoContent, nil
	})
}

func tusDeleteHandler() handleFunc {
	return withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
		// Get path from query parameter
		path := r.URL.Query().Get("path")
		if path == "" || path == "/" || !d.user.Perm.Create {
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

		_, err = getActiveUploadLength(file.RealPath())
		if err != nil {
			return http.StatusNotFound, err
		}

		err = d.requestFs.RemoveAll(path)
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
