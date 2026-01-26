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
	afs    afero.Fs
	Cfg    Config
	AwsCfg aws.Config
)

func Init(config *Config) error {
	Cfg = *config

	var err error
	AwsCfg, err = awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(Cfg.AccessKey, Cfg.SecretKey, ""),
		),
		awsconfig.WithRegion(Cfg.Region),
		awsconfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:           Cfg.Endpoint,
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

	afs = aferos3.NewFs(AwsCfg)
	err = SetupBucket()
	if err != nil {
		return err
	}
	return nil
}

func GetCurrenBucket() string {
	return Cfg.Bucket
}

func SwitchBucket(bucket string) error {
	Cfg.Bucket = bucket
	if s3fs, ok := GetS3FileSystem(afs); ok {
		s3fs.SetBucket(bucket)
		return nil
	}
	return fmt.Errorf("underlying filesystem is not an S3 filesystem")
}

func ListBuckets() ([]string, error) {
	if s3fs, ok := GetS3FileSystem(afs); ok {
		return s3fs.ListBuckets()
	}
	return nil, fmt.Errorf("underlying filesystem is not an S3 filesystem")
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
	return afs
}

func GetS3Client() *s3.Client {
	return s3.NewFromConfig(AwsCfg)
}

func GetAWSConfig() aws.Config {
	return AwsCfg
}

func FullPath(_ afero.Fs, relativePath string) string {
	return relativePath
}

// GetS3FileSystem returns the underlying S3 filesystem if it exists, accounting for wrappers
func GetS3FileSystem(fs afero.Fs) (*aferos3.Fs, bool) {
	if s3fs, ok := fs.(*aferos3.Fs); ok {
		return s3fs, true
	}

	if wrapper, ok := fs.(*S3PermissionWrapper); ok {
		return wrapper.GetUnderlyingS3Fs()
	}

	return nil, false
}

// IsS3FileSystem checks if the given filesystem is an S3 filesystem, accounting for wrappers
func IsS3FileSystem(fs afero.Fs) bool {
	if _, ok := fs.(*aferos3.Fs); ok {
		return true
	}

	if wrapper, ok := fs.(*S3PermissionWrapper); ok {
		return wrapper.IsS3Fs()
	}

	return false
}
