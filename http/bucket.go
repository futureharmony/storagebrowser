package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/futureharmony/storagebrowser/v2/minio"
	"github.com/futureharmony/storagebrowser/v2/users"
	"github.com/gorilla/mux"
)

var errStorageType = errors.New("bucket operations require S3 storage type")

func listBucketsHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		log.Printf("[BUCKET] listBucketsHandler: request received")
		if d.server.StorageType != "s3" {
			log.Printf("[BUCKET] listBucketsHandler: not S3 storage type")
			return http.StatusBadRequest, errStorageType
		}

		buckets, err := minio.ListBuckets()
		if err != nil {
			log.Printf("[BUCKET] listBucketsHandler: failed to list buckets: %v", err)
			return http.StatusInternalServerError, err
		}

		log.Printf("[BUCKET] listBucketsHandler: found %d buckets", len(buckets))
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
		log.Printf("[BUCKET] createBucketHandler: request received")
		if d.server.StorageType != "s3" {
			log.Printf("[BUCKET] createBucketHandler: not S3 storage type")
			return http.StatusBadRequest, errStorageType
		}

		if !d.user.Perm.Admin {
			log.Printf("[BUCKET] createBucketHandler: user %s is not admin", d.user.Username)
			return http.StatusForbidden, errors.New("admin permission required")
		}

		var req createBucketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[BUCKET] createBucketHandler: failed to decode request: %v", err)
			return http.StatusBadRequest, err
		}

		if req.Name == "" {
			log.Printf("[BUCKET] createBucketHandler: bucket name is empty")
			return http.StatusBadRequest, errors.New("bucket name is required")
		}

		log.Printf("[BUCKET] createBucketHandler: creating bucket %s", req.Name)
		if err := minio.CreateBucket(req.Name, &minio.BucketSettings{
			Name:           req.Name,
			Versioning:     req.Versioning,
			ObjectLock:     req.ObjectLock,
			ObjectLockDays: req.ObjectLockDays,
			RetentionMode:  req.RetentionMode,
			QuotaStorageGB: req.QuotaStorageGB,
			QuotaObjects:   req.QuotaObjects,
		}); err != nil {
			log.Printf("[BUCKET] createBucketHandler: failed to create bucket: %v", err)
			return http.StatusInternalServerError, err
		}

		log.Printf("[BUCKET] createBucketHandler: bucket %s created successfully", req.Name)
		return renderJSON(w, r, map[string]string{"name": req.Name})
	})
}

func deleteBucketHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		log.Printf("[BUCKET] deleteBucketHandler: request received")
		if d.server.StorageType != "s3" {
			log.Printf("[BUCKET] deleteBucketHandler: not S3 storage type")
			return http.StatusBadRequest, errStorageType
		}

		if !d.user.Perm.Admin {
			log.Printf("[BUCKET] deleteBucketHandler: user %s is not admin", d.user.Username)
			return http.StatusForbidden, errors.New("admin permission required")
		}

		vars := mux.Vars(r)
		name := vars["name"]

		if name == "" {
			log.Printf("[BUCKET] deleteBucketHandler: bucket name is empty")
			return http.StatusBadRequest, errors.New("bucket name is required")
		}

		log.Printf("[BUCKET] deleteBucketHandler: deleting bucket %s", name)
		if err := minio.DeleteBucket(name); err != nil {
			log.Printf("[BUCKET] deleteBucketHandler: failed to delete bucket: %v", err)
			return http.StatusInternalServerError, err
		}

		// Update all users to remove the deleted bucket from their AvailableScopes
		log.Printf("[BUCKET] deleteBucketHandler: updating users to remove bucket %s from scopes", name)
		allUsers, err := d.store.Users.Gets(d.server.Root)
		if err != nil {
			log.Printf("[BUCKET] deleteBucketHandler: failed to get users: %v", err)
		} else {
			for _, user := range allUsers {
				updated := false
				newScopes := make([]users.Scope, 0)
				for _, scope := range user.AvailableScopes {
					if scope.Name != name {
						newScopes = append(newScopes, scope)
					} else {
						updated = true
					}
				}
				if updated {
					user.AvailableScopes = newScopes
					// If current scope was the deleted bucket, switch to first available
					if user.CurrentScope.Name == name {
						if len(newScopes) > 0 {
							user.CurrentScope = newScopes[0]
						} else {
							user.CurrentScope = users.Scope{}
						}
					}
					err := d.store.Users.Update(user)
					if err != nil {
						log.Printf("[BUCKET] deleteBucketHandler: failed to update user %s: %v", user.Username, err)
					} else {
						log.Printf("[BUCKET] deleteBucketHandler: updated user %s, removed bucket %s", user.Username, name)
					}
				}
			}
		}

		log.Printf("[BUCKET] deleteBucketHandler: bucket %s deleted successfully", name)
		return http.StatusNoContent, nil
	})
}

func getBucketSettingsHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		log.Printf("[BUCKET] getBucketSettingsHandler: request received")
		if d.server.StorageType != "s3" {
			log.Printf("[BUCKET] getBucketSettingsHandler: not S3 storage type")
			return http.StatusBadRequest, errStorageType
		}

		if !d.user.Perm.Admin {
			log.Printf("[BUCKET] getBucketSettingsHandler: user %s is not admin", d.user.Username)
			return http.StatusForbidden, errors.New("admin permission required")
		}

		vars := mux.Vars(r)
		name := vars["name"]

		if name == "" {
			log.Printf("[BUCKET] getBucketSettingsHandler: bucket name is empty")
			return http.StatusBadRequest, errors.New("bucket name is required")
		}

		log.Printf("[BUCKET] getBucketSettingsHandler: getting settings for bucket %s", name)
		settings, err := minio.GetBucketSettings(name)
		if err != nil {
			log.Printf("[BUCKET] getBucketSettingsHandler: failed to get settings: %v", err)
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
		log.Printf("[BUCKET] updateBucketSettingsHandler: request received")
		if d.server.StorageType != "s3" {
			log.Printf("[BUCKET] updateBucketSettingsHandler: not S3 storage type")
			return http.StatusBadRequest, errStorageType
		}

		if !d.user.Perm.Admin {
			log.Printf("[BUCKET] updateBucketSettingsHandler: user %s is not admin", d.user.Username)
			return http.StatusForbidden, errors.New("admin permission required")
		}

		vars := mux.Vars(r)
		name := vars["name"]

		if name == "" {
			log.Printf("[BUCKET] updateBucketSettingsHandler: bucket name is empty")
			return http.StatusBadRequest, errors.New("bucket name is required")
		}

		var req updateBucketSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[BUCKET] updateBucketSettingsHandler: failed to decode request: %v", err)
			return http.StatusBadRequest, err
		}

		log.Printf("[BUCKET] updateBucketSettingsHandler: updating bucket %s (versioning=%v, objectLock=%v)", name, req.Versioning, req.ObjectLock)
		if err := minio.SetBucketVersioning(name, req.Versioning); err != nil {
			log.Printf("[BUCKET] updateBucketSettingsHandler: failed to set versioning: %v", err)
			return http.StatusInternalServerError, err
		}

		mode := types.ObjectLockRetentionModeGovernance
		if req.RetentionMode == "COMPLIANCE" {
			mode = types.ObjectLockRetentionModeCompliance
		}
		if err := minio.SetBucketObjectLock(name, req.ObjectLock, req.ObjectLockDays, mode); err != nil {
			log.Printf("[BUCKET] updateBucketSettingsHandler: failed to set object lock: %v", err)
			return http.StatusInternalServerError, err
		}

		if err := minio.SetBucketTags(name, req.QuotaStorageGB, req.QuotaObjects); err != nil {
			log.Printf("[BUCKET] updateBucketSettingsHandler: failed to set tags: %v", err)
			return http.StatusInternalServerError, err
		}

		settings, err := minio.GetBucketSettings(name)
		if err != nil {
			log.Printf("[BUCKET] updateBucketSettingsHandler: failed to get settings: %v", err)
			return http.StatusInternalServerError, err
		}

		log.Printf("[BUCKET] updateBucketSettingsHandler: bucket %s updated successfully", name)
		return renderJSON(w, r, settings)
	})
}
