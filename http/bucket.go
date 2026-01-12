package http

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/filebrowser/filebrowser/v2/minio"
	"net/http"
)

type bucketInfo struct {
	Name string `json:"name"`
}

var bucketListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if d.server.StorageType != "s3" {
		return renderJSON(w, r, []bucketInfo{})
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(d.server.S3AccessKey, d.server.S3SecretKey, ""),
		),
		awsconfig.WithRegion(d.server.S3Region),
		awsconfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:           d.server.S3Endpoint,
						SigningRegion: region,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			}),
		),
	)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	s3Client := s3.NewFromConfig(awsCfg)

	listBucketsOutput, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	buckets := make([]bucketInfo, 0, len(listBucketsOutput.Buckets))
	for _, bucket := range listBucketsOutput.Buckets {
		buckets = append(buckets, bucketInfo{
			Name: *bucket.Name,
		})
	}

	return renderJSON(w, r, buckets)
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
