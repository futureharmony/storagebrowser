package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/futureharmony/storagebrowser/v2/minio"
	"github.com/gorilla/mux"
)

var errStorageType = errors.New("bucket operations require S3 storage type")

func listBucketsHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if d.server.StorageType != "s3" {
			return http.StatusBadRequest, errStorageType
		}

		buckets, err := minio.ListBuckets()
		if err != nil {
			return http.StatusInternalServerError, err
		}

		bucketInfos := make([]map[string]string, len(buckets))
		for i, bucket := range buckets {
			bucketInfos[i] = map[string]string{
				"name": bucket,
			}
		}

		return renderJSON(w, r, bucketInfos)
	})
}

type createBucketRequest struct {
	Name           string `json:"name"`
	Versioning     bool   `json:"versioning"`
	ObjectLock     bool   `json:"objectLock"`
	ObjectLockDays int    `json:"objectLockDays"`
	RetentionMode  string `json:"retentionMode"`
	QuotaStorageGB int64  `json:"quotaStorageGB"`
	QuotaObjects   int64  `json:"quotaObjects"`
}

func createBucketHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if d.server.StorageType != "s3" {
			return http.StatusBadRequest, errStorageType
		}

		if !d.user.Perm.Admin {
			return http.StatusForbidden, errors.New("admin permission required")
		}

		var req createBucketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return http.StatusBadRequest, err
		}

		if req.Name == "" {
			return http.StatusBadRequest, errors.New("bucket name is required")
		}

		if err := minio.CreateBucket(req.Name, &minio.BucketSettings{
			Name:           req.Name,
			Versioning:     req.Versioning,
			ObjectLock:     req.ObjectLock,
			ObjectLockDays: req.ObjectLockDays,
			RetentionMode:  req.RetentionMode,
			QuotaStorageGB: req.QuotaStorageGB,
			QuotaObjects:   req.QuotaObjects,
		}); err != nil {
			return http.StatusInternalServerError, err
		}

		return renderJSON(w, r, map[string]string{"name": req.Name})
	})
}

func deleteBucketHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if d.server.StorageType != "s3" {
			return http.StatusBadRequest, errStorageType
		}

		if !d.user.Perm.Admin {
			return http.StatusForbidden, errors.New("admin permission required")
		}

		vars := mux.Vars(r)
		name := vars["name"]

		if name == "" {
			return http.StatusBadRequest, errors.New("bucket name is required")
		}

		if err := minio.DeleteBucket(name); err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusNoContent, nil
	})
}

func getBucketSettingsHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if d.server.StorageType != "s3" {
			return http.StatusBadRequest, errStorageType
		}

		if !d.user.Perm.Admin {
			return http.StatusForbidden, errors.New("admin permission required")
		}

		vars := mux.Vars(r)
		name := vars["name"]

		if name == "" {
			return http.StatusBadRequest, errors.New("bucket name is required")
		}

		settings, err := minio.GetBucketSettings(name)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return renderJSON(w, r, settings)
	})
}

type updateBucketSettingsRequest struct {
	Versioning     bool   `json:"versioning"`
	ObjectLock     bool   `json:"objectLock"`
	ObjectLockDays int    `json:"objectLockDays"`
	RetentionMode  string `json:"retentionMode"`
	QuotaStorageGB int64  `json:"quotaStorageGB"`
	QuotaObjects   int64  `json:"quotaObjects"`
}

func updateBucketSettingsHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if d.server.StorageType != "s3" {
			return http.StatusBadRequest, errStorageType
		}

		if !d.user.Perm.Admin {
			return http.StatusForbidden, errors.New("admin permission required")
		}

		vars := mux.Vars(r)
		name := vars["name"]

		if name == "" {
			return http.StatusBadRequest, errors.New("bucket name is required")
		}

		var req updateBucketSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return http.StatusBadRequest, err
		}

		if err := minio.SetBucketVersioning(name, req.Versioning); err != nil {
			return http.StatusInternalServerError, err
		}

		mode := types.ObjectLockRetentionModeGovernance
		if req.RetentionMode == "COMPLIANCE" {
			mode = types.ObjectLockRetentionModeCompliance
		}
		if err := minio.SetBucketObjectLock(name, req.ObjectLock, req.ObjectLockDays, mode); err != nil {
			return http.StatusInternalServerError, err
		}

		if err := minio.SetBucketTags(name, req.QuotaStorageGB, req.QuotaObjects); err != nil {
			return http.StatusInternalServerError, err
		}

		settings, err := minio.GetBucketSettings(name)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return renderJSON(w, r, settings)
	})
}
