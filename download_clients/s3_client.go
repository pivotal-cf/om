package download_clients

import (
	"errors"
	"fmt"
	"log"
	"os"
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

func NewS3Client(stower Stower, config S3Configuration, stderr *log.Logger) (s3Client, error) {
	return s3Client{
		S3Configuration: config,
	}, nil
}

func validateAccessKeyAuthType(config S3Configuration) error {
	if config.AuthType == "iam" {
		return nil
	}

	if config.AccessKeyID == "" || config.SecretAccessKey == "" {
		return errors.New("the flags \"s3-access-key-id\" and \"s3-secret-access-key\" are required when the \"auth-type\" is \"accesskey\"")
	}

	return nil
}

type s3Client struct {
	S3Configuration
}

func (t s3Client) Name() string {
	return fmt.Sprintf("s3://%s", t.S3Configuration.Bucket)
}

func (t s3Client) GetAllProductVersions(slug string) ([]string, error) {
	return []string{}, nil

}
func (t s3Client) GetLatestProductFile(slug, version, glob string) (FileArtifacter, error) {
	return nil, nil
}

func (t s3Client) DownloadProductToFile(fa FileArtifacter, file *os.File) error {
	return nil
}

func (t s3Client) GetLatestStemcellForProduct(fa FileArtifacter, downloadedProductFileName string) (StemcellArtifacter, error) {
	return nil, nil
}
