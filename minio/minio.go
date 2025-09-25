package minio

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	aferos3 "github.com/futureharmony/afero-aws-s3"
	"github.com/spf13/afero"
)

var bucketName = ""
var fs afero.Fs = nil

// TODO read config from config file or env
func init() {

	endpoint := "" // Custom S3-compatible service endpoint

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("ak", "sk", ""),
		),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolver(
			aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:           endpoint,
						SigningRegion: region,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			}),
		),
	)

	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	// sess, errSession := session.NewSession(&aws.Config{
	// 	Credentials:      credentials.NewStaticCredentials("testuser", "test123123", ""),
	// 	Endpoint:         aws.String("https://io.lifelib.org/testbucket"),
	// 	Region:           aws.String("cn-north-1"),
	// 	DisableSSL:       aws.Bool(false),
	// 	S3ForcePathStyle: aws.Bool(true),
	// })

	fs = aferos3.NewFs(bucketName, cfg)
}

func NewBasePathFs() afero.Fs {
	return fs
}
