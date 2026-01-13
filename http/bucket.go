package http

import (
	"encoding/json"

	"net/http"

	"github.com/filebrowser/filebrowser/v2/minio"
)

var bucketListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if d.server.StorageType != "s3" {
		return renderJSON(w, r, []minio.BucketInfo{})
	}

	return renderJSON(w, r, minio.CachedBuckets)
})

type bucketSwitchRequest struct {
	Bucket string `json:"bucket"`
}

var bucketSwitchHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if d.server.StorageType != "s3" {
		return http.StatusBadRequest, nil
	}

	var req bucketSwitchRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if req.Bucket == "" {
		return http.StatusBadRequest, nil
	}

	err = minio.SwitchBucket(req.Bucket)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, map[string]string{"bucket": req.Bucket})
})
