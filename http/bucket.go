package http

import (
	"net/http"

	"github.com/futureharmony/storagebrowser/v2/minio"
)

func listBucketsHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if d.server.StorageType != "s3" {
			return http.StatusBadRequest, nil
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
