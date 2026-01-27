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

	afs = aferos3.NewFsWrapper(AwsCfg, Cfg.Bucket, "") // Using the wrapper with fixed bucket
	err = SetupBucket()
	if err != nil {
		return err
	}
	return nil
}

func GetCurrenBucket() string {
	return Cfg.Bucket
}

func SwitchBase(bucket, scope string) error {
	Cfg.Bucket = bucket
	if _, ok := GetS3FileSystem(afs); ok {
		// With the new API, we need to create a new wrapper with the new bucket
		afs = aferos3.NewFsWrapper(AwsCfg, bucket, scope)
		return nil
	}
	return fmt.Errorf("underlying filesystem is not an S3 filesystem")
}

// SwitchUserBase switches the bucket and scope for a user's specific filesystem
func SwitchUserBase(userFs *afero.Fs, bucket, scope string) {
	// Create a new wrapper with the updated bucket and scope
	*userFs = aferos3.NewFsWrapper(AwsCfg, bucket, scope)
	return
}

func CreateUserFs(bucket, scope string) afero.Fs {
	if bucket == "" {
		bucketNames, err := ListBuckets()
		if err != nil {
			return nil
		}

		if len(bucketNames) == 0 {
			return nil
		}
		return aferos3.NewFsWrapper(AwsCfg, bucketNames[0], scope)
	}
	return aferos3.NewFsWrapper(AwsCfg, bucket, scope)
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
	SwitchBase(buckets[0].Name, "")

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
	if s3wrapper, ok := fs.(*aferos3.FsWrapper); ok {
		return s3wrapper.Fs, true
	}

	return nil, false
}

// IsS3FileSystem checks if the given filesystem is an S3 filesystem, accounting for wrappers
func IsS3FileSystem(fs afero.Fs) bool {
	if _, ok := fs.(*aferos3.FsWrapper); ok {
		return true
	}

	return false
}
