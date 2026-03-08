package minio

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	aferos3 "github.com/futureharmony/afero-aws-s3"
	madmingo "github.com/minio/madmin-go/v3"
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
	afs         afero.Fs
	Cfg         Config
	AwsCfg      aws.Config
	adminClient *madmingo.AdminClient
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

	// Initialize MinIO Admin Client for quota management
	// Note: madmin.New expects host:port format, not http://host:port
	adminEndpoint := Cfg.Endpoint
	if strings.HasPrefix(adminEndpoint, "http://") {
		adminEndpoint = strings.TrimPrefix(adminEndpoint, "http://")
	} else if strings.HasPrefix(adminEndpoint, "https://") {
		adminEndpoint = strings.TrimPrefix(adminEndpoint, "https://")
	}

	log.Printf("[MINIO] Init: creating admin client with endpoint=%s (original: %s), accessKey=%s",
		adminEndpoint, Cfg.Endpoint, Cfg.AccessKey)
	adminClient, err = madmingo.New(adminEndpoint, Cfg.AccessKey, Cfg.SecretKey, false)
	if err != nil {
		log.Printf("[MINIO] Init: failed to create admin client: %v", err)
		return err
	}

	log.Printf("[MINIO] Init: admin client created, setting custom transport")
	// Set custom transport to match S3 client configuration
	adminClient.SetCustomTransport(&http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
	})
	log.Printf("[MINIO] Admin client initialized: %v", adminClient != nil)

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
	QuotaStorageMB int64  `json:"quotaStorageMB"`  // Storage quota in MB
	QuotaObjects   int64  `json:"quotaObjects"`
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

	if settings.QuotaStorageMB > 0 || settings.QuotaObjects > 0 {
		if err := SetBucketQuota(name, settings.QuotaStorageMB, settings.QuotaObjects); err != nil {
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

	// Get quota settings using dedicated function
	quotaStorage, quotaObjects, quotaErr := GetBucketQuota(name)
	if quotaErr == nil {
		settings.QuotaStorageMB = quotaStorage
		settings.QuotaObjects = quotaObjects
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

	// If already enabled, can only update retention settings
	if currentEnabled && lockConfig.ObjectLockConfiguration.Rule != nil && lockConfig.ObjectLockConfiguration.Rule.DefaultRetention != nil {
		_, err := client.PutObjectLockConfiguration(ctx, &s3.PutObjectLockConfigurationInput{
			Bucket: aws.String(name),
			ObjectLockConfiguration: &types.ObjectLockConfiguration{
				ObjectLockEnabled: types.ObjectLockEnabledEnabled,
				Rule: &types.ObjectLockRule{
					DefaultRetention: &types.DefaultRetention{
						Days: aws.Int32(int32(days)),
						Mode: mode,
					},
				},
			},
		})
		return err
	}

	return nil
}

func SetBucketTags(name string, quotaStorageGB int64, quotaObjects int64) error {
	client := GetS3Client()
	ctx := context.Background()

	var tagSet []types.Tag
	if quotaStorageGB > 0 {
		tagSet = append(tagSet, types.Tag{
			Key:   aws.String("QuotaStorageGB"),
			Value: aws.String(fmt.Sprintf("%d", quotaStorageGB)),
		})
	}
	if quotaObjects > 0 {
		tagSet = append(tagSet, types.Tag{
			Key:   aws.String("QuotaObjects"),
			Value: aws.String(fmt.Sprintf("%d", quotaObjects)),
		})
	}

	if len(tagSet) == 0 {
		_, err := client.DeleteBucketTagging(ctx, &s3.DeleteBucketTaggingInput{
			Bucket: aws.String(name),
		})
		return err
	}

	_, err := client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
		Bucket: aws.String(name),
		Tagging: &types.Tagging{
			TagSet: tagSet,
		},
	})
	return err
}

// SetBucketQuota sets the bucket quota using MinIO Admin API
// This actually enforces the quota on the bucket
func SetBucketQuota(name string, quotaStorageMB int64, quotaObjects int64) error {
	ctx := context.Background()

	log.Printf("[MINIO] SetBucketQuota: called for %s, storage=%d MB, objects=%d, adminClient=%v",
		name, quotaStorageMB, quotaObjects, adminClient != nil)

	// Try to use MinIO Admin API if available
	if adminClient != nil {
		// MinIO madmin-go supports storage quota enforcement
		// Note: Object quota is stored in tags but not enforced by MinIO (API limitation)

		// Set storage quota (convert MB to bytes)
		var storageQuota uint64 = 0
		if quotaStorageMB > 0 {
			storageQuota = uint64(quotaStorageMB * 1024 * 1024)
		}

		// Use MinIO's bucket quota API
		quota := madmingo.BucketQuota{
			Quota: storageQuota,
			Type:  madmingo.HardQuota, // Hard quota means strict enforcement
		}

		log.Printf("[MINIO] SetBucketQuota: calling admin API for %s with storage=%d bytes (%d MB)",
			name, storageQuota, quotaStorageMB)
		err := adminClient.SetBucketQuota(ctx, name, &quota)
		if err != nil {
			log.Printf("[MINIO] SetBucketQuota (storage) failed for %s: %v", name, err)
			// Continue to store quota in tags as fallback
		} else {
			log.Printf("[MINIO] SetBucketQuota (storage) set successfully for %s: %d MB", name, quotaStorageMB)
		}
	} else {
		log.Printf("[MINIO] SetBucketQuota: admin client not available, quota will not be enforced for %s", name)
	}

	// Note: Object quota requires newer version of madmin-go or MinIO server
	// For now, we only support storage quota enforcement
	// Object quota is still stored in tags for reference

	// Store quota settings in tags for reference and GetBucketSettings
	tagSet := []types.Tag{}
	if quotaStorageMB > 0 {
		tagSet = append(tagSet, types.Tag{
			Key:   aws.String("QuotaStorageMB"),
			Value: aws.String(strconv.FormatInt(quotaStorageMB, 10)),
		})
	}
	if quotaObjects > 0 {
		tagSet = append(tagSet, types.Tag{
			Key:   aws.String("QuotaObjects"),
			Value: aws.String(strconv.FormatInt(quotaObjects, 10)),
		})
	}

	s3Client := GetS3Client()
	var err error
	if len(tagSet) == 0 {
		_, err = s3Client.DeleteBucketTagging(ctx, &s3.DeleteBucketTaggingInput{
			Bucket: aws.String(name),
		})
	} else {
		_, err = s3Client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
			Bucket: aws.String(name),
			Tagging: &types.Tagging{
				TagSet: tagSet,
			},
		})
	}

	return err
}

// GetBucketQuota retrieves the bucket quota settings using MinIO Admin API
func GetBucketQuota(name string) (storageQuota int64, objectsQuota int64, err error) {
	ctx := context.Background()

	// Try to get quota from Admin API if available
	if adminClient != nil {
		quota, err := adminClient.GetBucketQuota(ctx, name)
		if err != nil {
			log.Printf("[MINIO] GetBucketQuota (admin API) failed for %s: %v", name, err)
			// Continue to try tags as fallback
		} else {
			// Convert bytes to MB
			storageQuota = int64(quota.Quota) / (1024 * 1024)
			log.Printf("[MINIO] GetBucketQuota (admin API) returned storage=%d MB for %s",
				storageQuota, name)
		}
	}

	// Get quota from tags (MB or legacy GB)
	s3Client := GetS3Client()
	tagsOut, err := s3Client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
		Bucket: aws.String(name),
	})
	if err == nil && tagsOut.TagSet != nil {
		for _, tag := range tagsOut.TagSet {
			if tag.Key == nil || tag.Value == nil {
				continue
			}
			switch *tag.Key {
			case "QuotaStorageMB":
				fmt.Sscanf(*tag.Value, "%d", &storageQuota)
			case "QuotaStorageGB":
				// Legacy support: convert GB to MB
				var gb int64
				fmt.Sscanf(*tag.Value, "%d", &gb)
				if storageQuota == 0 {
					storageQuota = gb * 1024
				}
			case "QuotaObjects":
				fmt.Sscanf(*tag.Value, "%d", &objectsQuota)
			}
		}
	}

	return storageQuota, objectsQuota, nil
}
