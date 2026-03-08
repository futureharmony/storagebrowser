package minio

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
	log.Printf("[MINIO] Init: starting initialization with endpoint=%s, region=%s, bucket=%s",
		config.Endpoint, config.Region, config.Bucket)

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: nil,
		},
	}

	var err error
	AwsCfg, err = awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(Cfg.AccessKey, Cfg.SecretKey, ""),
		),
		awsconfig.WithRegion(Cfg.Region),
		awsconfig.WithHTTPClient(httpClient),
		awsconfig.WithRetryer(func() aws.Retryer {
			return aws.NopRetryer{}
		}),
		awsconfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					log.Printf("[MINIO] Using endpoint: %s, region: %s", Cfg.Endpoint, region)
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
		log.Printf("[MINIO] Init: failed to load AWS config: %v", err)
		return err
	}

	log.Printf("[MINIO] Initialized with endpoint: %s, region: %s, bucket: %s", Cfg.Endpoint, Cfg.Region, Cfg.Bucket)

	afs = aferos3.NewFsWrapper(AwsCfg, "", "")
	err = SetupBucket()
	if err != nil {
		log.Printf("[MINIO] Init: failed to setup bucket: %v", err)
		return err
	}

	log.Printf("[MINIO] Init: initialization complete")
	return nil
}

func CreateUserFs(bucket, scope string) afero.Fs {
	if bucket == "" {
		// If no bucket is specified, use the first available bucket
		bucketNames, err := ListBuckets()
		if err != nil || len(bucketNames) == 0 {
			// Fallback to empty bucket if no buckets are available
			return aferos3.NewFsWrapper(AwsCfg, "", scope)
		}
		return aferos3.NewFsWrapper(AwsCfg, bucketNames[0], scope)
	}
	return aferos3.NewFsWrapper(AwsCfg, bucket, scope)
}

func ListBuckets() ([]string, error) {
	start := time.Now()
	log.Printf("[MINIO] ListBuckets: starting")

	// Use S3 client directly to list all buckets
	client := GetS3Client()
	ctx := context.Background()

	output, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Printf("[MINIO] ListBuckets: failed to list buckets: %v", err)
		return nil, err
	}

	buckets := make([]string, 0, len(output.Buckets))
	for _, bucket := range output.Buckets {
		if bucket.Name != nil {
			buckets = append(buckets, *bucket.Name)
		}
	}

	log.Printf("[MINIO] ListBuckets: took %v, count=%d", time.Since(start), len(buckets))
	return buckets, nil
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
	return s3.NewFromConfig(AwsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
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

type BucketSettings struct {
	Name           string `json:"name"`
	Versioning     bool   `json:"versioning"`
	ObjectLock     bool   `json:"objectLock"`
	ObjectLockDays int    `json:"objectLockDays"`
	RetentionMode  string `json:"retentionMode"`
}

func CreateBucket(name string, settings *BucketSettings) error {
	client := GetS3Client()
	ctx := context.Background()

	log.Printf("[MINIO] CreateBucket: creating bucket %s with region %s", name, Cfg.Region)

	input := &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}

	// Only set LocationConstraint if region is not us-east-1
	if Cfg.Region != "" && Cfg.Region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(Cfg.Region),
		}
	}

	// Enable object lock at creation time if requested
	if settings.ObjectLock {
		input.ObjectLockEnabledForBucket = aws.Bool(true)
	}

	log.Printf("[MINIO] CreateBucket: input = %+v", input)

	_, err := client.CreateBucket(ctx, input)
	if err != nil {
		log.Printf("[MINIO] CreateBucket: failed: %v", err)
		var bucketExistsErr *types.BucketAlreadyExists
		var bucketOwnedErr *types.BucketAlreadyOwnedByYou
		switch {
		case errors.As(err, &bucketExistsErr):
			return fmt.Errorf("bucket %s already exists", name)
		case errors.As(err, &bucketOwnedErr):
			return fmt.Errorf("bucket %s already exists and is owned by you", name)
		default:
			return err
		}
	}

	log.Printf("[MINIO] CreateBucket: success, now setting properties")

	if settings.Versioning {
		if err := SetBucketVersioning(name, true); err != nil {
			return err
		}
	}

	// Set retention settings after bucket is created (object lock must be enabled at creation time)
	if settings.ObjectLock && settings.ObjectLockDays > 0 {
		mode := types.ObjectLockRetentionModeGovernance
		if settings.RetentionMode == "COMPLIANCE" {
			mode = types.ObjectLockRetentionModeCompliance
		}
		if err := SetBucketObjectLock(name, true, settings.ObjectLockDays, mode); err != nil {
			return err
		}
	}

	return nil
}

func DeleteBucket(name string) error {
	client := GetS3Client()
	ctx := context.Background()

	_, err := client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	})
	return err
}

func GetBucketSettings(name string) (*BucketSettings, error) {
	client := GetS3Client()
	ctx := context.Background()

	settings := &BucketSettings{Name: name}

	log.Printf("[MINIO] GetBucketSettings: starting for bucket %s", name)
	start := time.Now()

	var wg sync.WaitGroup
	var versioningErr, lockErr error
	var versioningOut *s3.GetBucketVersioningOutput
	var lockConfigOut *s3.GetObjectLockConfigurationOutput

	wg.Add(2)

	go func() {
		defer wg.Done()
		v1Start := time.Now()
		versioningOut, versioningErr = client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
			Bucket: aws.String(name),
		})
		log.Printf("[MINIO] GetBucketSettings: GetBucketVersioning took %v, err=%v", time.Since(v1Start), versioningErr)
	}()

	go func() {
		defer wg.Done()
		v2Start := time.Now()
		lockConfigOut, lockErr = client.GetObjectLockConfiguration(ctx, &s3.GetObjectLockConfigurationInput{
			Bucket: aws.String(name),
		})
		log.Printf("[MINIO] GetBucketSettings: GetObjectLockConfiguration took %v, err=%v", time.Since(v2Start), lockErr)
	}()

	wg.Wait()

	if versioningErr == nil && versioningOut != nil && versioningOut.Status == types.BucketVersioningStatusEnabled {
		settings.Versioning = true
	}

	if lockErr == nil && lockConfigOut != nil && lockConfigOut.ObjectLockConfiguration != nil && lockConfigOut.ObjectLockConfiguration.ObjectLockEnabled == types.ObjectLockEnabledEnabled {
		settings.ObjectLock = true
		if lockConfigOut.ObjectLockConfiguration.Rule != nil && lockConfigOut.ObjectLockConfiguration.Rule.DefaultRetention != nil {
			retention := lockConfigOut.ObjectLockConfiguration.Rule.DefaultRetention
			if retention.Days != nil {
				settings.ObjectLockDays = int(*retention.Days)
			}
			settings.RetentionMode = string(retention.Mode)
		}
	}

	log.Printf("[MINIO] GetBucketSettings: total time %v", time.Since(start))
	return settings, nil
}

func SetBucketVersioning(name string, enabled bool) error {
	client := GetS3Client()
	ctx := context.Background()

	status := types.BucketVersioningStatusSuspended
	if enabled {
		status = types.BucketVersioningStatusEnabled
	}

	_, err := client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(name),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: status,
		},
	})
	return err
}

func SetBucketObjectLock(name string, enabled bool, days int, mode types.ObjectLockRetentionMode) error {
	client := GetS3Client()
	ctx := context.Background()

	// Get current object lock config
	lockConfig, err := client.GetObjectLockConfiguration(ctx, &s3.GetObjectLockConfigurationInput{
		Bucket: aws.String(name),
	})

	currentEnabled := false
	if err == nil && lockConfig.ObjectLockConfiguration != nil {
		currentEnabled = lockConfig.ObjectLockConfiguration.ObjectLockEnabled == types.ObjectLockEnabledEnabled
	}

	// Cannot enable object lock on existing bucket if not already enabled
	if enabled && !currentEnabled {
		log.Printf("[MINIO] SetBucketObjectLock: cannot enable object lock on existing bucket %s, must be set at creation time", name)
		return fmt.Errorf("object lock cannot be enabled on existing buckets, it must be set at bucket creation time")
	}

	// Cannot disable object lock if it's enabled
	if !enabled && currentEnabled {
		log.Printf("[MINIO] SetBucketObjectLock: cannot disable object lock on existing bucket %s", name)
		return fmt.Errorf("object lock cannot be disabled once enabled")
	}

	// Update retention settings if object lock is enabled
	if currentEnabled || enabled {
		log.Printf("[MINIO] SetBucketObjectLock: updating retention settings for %s (days=%d, mode=%s)", name, days, mode)

		input := &s3.PutObjectLockConfigurationInput{
			Bucket: aws.String(name),
			ObjectLockConfiguration: &types.ObjectLockConfiguration{
				ObjectLockEnabled: types.ObjectLockEnabledEnabled,
			},
		}

		// Only set rule if days > 0
		if days > 0 {
			input.ObjectLockConfiguration.Rule = &types.ObjectLockRule{
				DefaultRetention: &types.DefaultRetention{
					Days: aws.Int32(int32(days)),
					Mode: mode,
				},
			}
		}

		_, err := client.PutObjectLockConfiguration(ctx, input)
		if err != nil {
			log.Printf("[MINIO] SetBucketObjectLock: failed to update configuration: %v", err)
			return err
		}
		log.Printf("[MINIO] SetBucketObjectLock: configuration updated successfully")
		return nil
	}

	return nil
}
