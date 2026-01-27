package http

import (
	"encoding/json"

	"net/http"

	"github.com/futureharmony/storagebrowser/v2/minio"
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

	err = minio.SwitchBase(req.Bucket, "")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Update the user's bucket in the database
	currentUser, err := d.store.Users.Get(d.server.Root, d.user.ID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	currentUser.Bucket = req.Bucket
	err = d.store.Users.Update(currentUser, "Bucket")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Update the user's filesystem to use the new bucket and scope
	currentUser.Fs = minio.CreateUserFs(req.Bucket, currentUser.Scope)
	d.user = currentUser // Update the current request's user reference

	return renderJSON(w, r, map[string]string{"bucket": req.Bucket})
})
