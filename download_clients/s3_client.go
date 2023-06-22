package download_clients

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Configuration struct {
	Bucket          string `validate:"required"`
	AccessKeyID     string
	SecretAccessKey string
	RegionName      string `validate:"required"`
	Endpoint        string
	DisableSSL      bool
	EnableV2Signing bool
	ProductPath     string
	StemcellPath    string
	AuthType        string
}

func (s S3Configuration) ResolveEndpoint(service, region string, options ...interface{}) (aws.Endpoint, error) {
	return aws.Endpoint{
		URL: s.Endpoint,
	}, nil
}

//counterfeiter:generate -o ./fakes/s3_client.go . AWSS3Client
type AWSS3Client interface {
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type s3Client struct {
	S3Configuration
	aws.Config
	Client AWSS3Client
}

func NewS3Client(c S3Configuration, stderr *log.Logger) (s3Client, error) {
	if err := validateAccessKeyAuthType(c); err != nil {
		return s3Client{}, err
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(c.RegionName),
		loadAuthConfig(c),
		config.WithEndpointResolverWithOptions(c),
	)

	if err != nil {
		return s3Client{}, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = c.RegionName
		o.UseAccelerate = true
	})

	return s3Client{
		S3Configuration: c,
		Config:          cfg,
		Client:          client,
	}, nil
}

func loadAuthConfig(c S3Configuration) func(*config.LoadOptions) error {
	if c.AuthType == "iam" {
		// use default credentials
		return func(_ *config.LoadOptions) error { return nil }
	}

	return config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, ""))
}

func validateAccessKeyAuthType(c S3Configuration) error {
	if c.AuthType == "iam" {
		return nil
	}

	if c.AccessKeyID == "" || c.SecretAccessKey == "" {
		return errors.New("the flags \"s3-access-key-id\" and \"s3-secret-access-key\" are required when the \"auth-type\" is \"accesskey\"")
	}

	return nil
}

func (s s3Client) Name() string {
	return fmt.Sprintf("s3://%s", s.S3Configuration.Bucket)
}

func (s s3Client) GetAllProductVersions(slug string) ([]string, error) {
	r, _ := s.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(slug),
	})

	returns := make([]string, len(r.Contents))

	for i, o := range r.Contents {
		returns[i] = *o.Key
	}

	return returns, nil
}

func (s s3Client) GetLatestProductFile(slug, version, glob string) (FileArtifacter, error) {
	return nil, nil
}

func (s s3Client) DownloadProductToFile(fa FileArtifacter, file *os.File) error {
	s.Client.GetObject(context.TODO(), &s3.GetObjectInput{})
	return nil
}

func (s s3Client) GetLatestStemcellForProduct(fa FileArtifacter, downloadedProductFileName string) (StemcellArtifacter, error) {
	return nil, nil
}
