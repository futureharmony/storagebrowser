package http

import (
	"encoding/json"

	"net/http"

	"github.com/futureharmony/storagebrowser/v2/minio"
	"github.com/futureharmony/storagebrowser/v2/users"
)

var bucketListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if d.server.StorageType != "s3" {
		return renderJSON(w, r, map[string]interface{}{
			"availableScopes": d.user.AvailableScopes,
			"currentScope":    d.user.CurrentScope,
		})
	}

	// Admin users can see all buckets as available scopes
	if d.user.Perm.Admin {
		// Convert cached buckets to scopes
		scopes := make([]users.Scope, len(minio.CachedBuckets))
		for i, bucket := range minio.CachedBuckets {
			scopes[i] = users.Scope{
				Name:       bucket.Name,
				RootPrefix: "/",
			}
		}

		// Set current scope if not already set
		currentScope := d.user.CurrentScope
		if currentScope.Name == "" && len(scopes) > 0 {
			currentScope = scopes[0]
		}

		return renderJSON(w, r, map[string]interface{}{
			"availableScopes": scopes,
			"currentScope":    currentScope,
		})
	} else {
		// Non-admin users use their existing available scopes
		scopes := d.user.AvailableScopes
		if len(scopes) == 0 {
			// If no available scopes set, return empty array
			scopes = []users.Scope{}
		}

		currentScope := d.user.CurrentScope
		if currentScope.Name == "" && len(scopes) > 0 {
			currentScope = scopes[0]
		}

		return renderJSON(w, r, map[string]interface{}{
			"availableScopes": scopes,
			"currentScope":    currentScope,
		})
	}
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

	// Update the user's current scope in the database
	currentUser, err := d.store.Users.Get(d.server.Root, d.user.ID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Find the corresponding scope for the selected bucket
	// If the bucket matches a scope name, update the current scope
	for _, scope := range currentUser.AvailableScopes {
		if scope.Name == req.Bucket || scope.Name == req.Bucket+"/" {
			currentUser.CurrentScope = scope
			break
		}
	}

	// If no specific scope found for the bucket, create a default one
	if currentUser.CurrentScope.Name == "" || currentUser.CurrentScope.Name == req.Bucket {
		currentUser.CurrentScope = users.Scope{
			Name:       req.Bucket,
			RootPrefix: "/",
		}
	}

	err = d.store.Users.Update(currentUser, "CurrentScope")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Update the user's filesystem to use the new bucket and scope
	currentUser.Fs = minio.CreateUserFs(req.Bucket, currentUser.CurrentScope.RootPrefix)
	d.user = currentUser // Update the current request's user reference

	return renderJSON(w, r, map[string]interface{}{
		"bucket":       req.Bucket,
		"currentScope": currentUser.CurrentScope,
	})
})
