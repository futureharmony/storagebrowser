package minio

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	aferos3 "github.com/futureharmony/afero-aws-s3"
	"github.com/spf13/afero"
)

var CachedBuckets []BucketInfo

type Config struct {
	Bucket    string
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
}

var (
	fs     afero.Fs
	cfg    Config
	awsCfg aws.Config
)

func Init(config *Config) error {
	cfg = *config

	var err error
	awsCfg, err = awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:           cfg.Endpoint,
						SigningRegion: region,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			}),
		),
	)

	if err != nil {
		return err
	}

	fs = aferos3.NewFs(awsCfg)
	err = SetupBucket()
	if err != nil {
		return err
	}
	return nil
}

func GetCurrenBucket() string {
	return cfg.Bucket
}

func SwitchBucket(bucket string) error {
	cfg.Bucket = bucket
	s3Fs := fs.(*aferos3.Fs)
	s3Fs.SetBucket(bucket)
	return nil
}

func ListBuckets() ([]string, error) {
	s3Fs := fs.(*aferos3.Fs)
	return s3Fs.ListBuckets()
}

type BucketInfo struct {
	Name string `json:"name"`
}

func SetupBucket() error {

	bucketNames, err := ListBuckets()
	if err != nil {
		return err
	}

	if len(bucketNames) == 0 {
		return fmt.Errorf("no available S3 buckets found")
	}

	buckets := make([]BucketInfo, 0, len(bucketNames))
	for _, name := range bucketNames {
		buckets = append(buckets, BucketInfo{
			Name: name,
		})
	}

	SwitchBucket(buckets[0].Name)
	CachedBuckets = buckets
	return nil
}

func NewBasePathFs() afero.Fs {
	return fs
}

func FullPath(_ afero.Fs, relativePath string) string {
	return relativePath
}
