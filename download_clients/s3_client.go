package download_clients

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Configuration struct {
	Bucket          string `validate:"required"`
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	RoleARN         string
	RegionName      string `validate:"required"`
	Endpoint        string
	DisableSSL      bool
	EnableV2Signing bool
	ProductPath     string
	StemcellPath    string
	AuthType        string
}

func withEndpointIfNeeded(c S3Configuration) func(*config.LoadOptions) error {
	if c.Endpoint != "" {
		// Create an endpoint resolver that returns the endpoint from the user configuration
		return config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: c.Endpoint}, nil
			}),
		)
	}
	return func(_ *config.LoadOptions) error { return nil }
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
		withEndpointIfNeeded(c),
		withAuthConfig(c),
		withARNConfig(c),
	)

	if err != nil {
		return s3Client{}, err
	}

	client := s3.NewFromConfig(cfg)

	return s3Client{
		S3Configuration: c,
		Config:          cfg,
		Client:          client,
	}, nil
}

func withAuthConfig(c S3Configuration) func(*config.LoadOptions) error {
	if c.AuthType == "iam" {
		// use default credentials
		return func(_ *config.LoadOptions) error { return nil }
	}

	return config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, c.SessionToken))
}

func withARNConfig(c S3Configuration) func(*config.LoadOptions) error {
	if c.RoleARN != "" {
		return config.WithAssumeRoleCredentialOptions(func(options *stscreds.AssumeRoleOptions) {
			options.RoleARN = c.RoleARN
		})
	}
	return func(_ *config.LoadOptions) error { return nil }
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
	fileRegex := regexp.MustCompile(fmt.Sprintf(`^%s/%s-([0-9]*\.[0-9]*\.*[0-9]*)`,
		regexp.QuoteMeta(strings.Trim(s.ProductPath, "/")),
		slug,
	))
	r, err := s.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(s.ProductPath),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in bucket %s: %w", s.Bucket, err)
	}

	files := make([]string, len(r.Contents))

	for i, o := range r.Contents {
		files[i] = aws.ToString(o.Key)

		// fmt.Println(aws.ToString(o.Key))
	}

	var versions []string
	versionFound := make(map[string]bool)
	for _, f := range files {
		fmt.Println(f)
		match := fileRegex.FindStringSubmatch(f)
		if match != nil {
			version := match[1]
			if !versionFound[version] {
				versions = append(versions, version)
				versionFound[version] = true
			}
		}
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no files matching pivnet-product-slug %s found", slug)
	}

	return versions, nil
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
