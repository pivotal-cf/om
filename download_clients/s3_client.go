package download_clients

import (
	"fmt"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/s3"
	"github.com/pivotal-cf/om/commands"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"log"
	"strconv"
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

func NewS3Client(stower Stower, config S3Configuration, progressWriter io.Writer) (stowClient, error) {
	validate := validator.New()
	err := validate.Struct(config)
	if err != nil {
		return stowClient{}, err
	}

	disableSSL := strconv.FormatBool(config.DisableSSL)
	enableV2Signing := strconv.FormatBool(config.EnableV2Signing)
	if config.AuthType == "" {
		config.AuthType = "accesskey"
	}

	err = validateAccessKeyAuthType(config)
	if err != nil {
		return stowClient{}, err
	}

	stowConfig := stow.ConfigMap{
		s3.ConfigAccessKeyID: config.AccessKeyID,
		s3.ConfigSecretKey:   config.SecretAccessKey,
		s3.ConfigRegion:      config.RegionName,
		s3.ConfigEndpoint:    config.Endpoint,
		s3.ConfigDisableSSL:  disableSSL,
		s3.ConfigV2Signing:   enableV2Signing,
		s3.ConfigAuthType:    config.AuthType,
	}

	return NewStowClient(
		stower,
		config.Bucket,
		stowConfig,
		progressWriter,
		config.ProductPath,
		config.StemcellPath,
		"s3",
	), nil
}

func validateAccessKeyAuthType(config S3Configuration) error {
	if config.AuthType == "iam" {
		return nil
	}

	if config.AccessKeyID == "" || config.SecretAccessKey == "" {
		return fmt.Errorf("the flags \"s3-access-key-id\" and \"s3-secret-access-key\" are required when the \"auth-type\" is \"accesskey\"")
	}

	return nil
}

func init() {
	initializer := func(
		c commands.DownloadProductOptions,
		progressWriter io.Writer,
		_ *log.Logger,
		_ *log.Logger,
	) (commands.ProductDownloader, error) {
		config := S3Configuration{
			Bucket:          c.S3Bucket,
			AccessKeyID:     c.S3AccessKeyID,
			AuthType:        c.S3AuthType,
			SecretAccessKey: c.S3SecretAccessKey,
			RegionName:      c.S3RegionName,
			Endpoint:        c.S3Endpoint,
			DisableSSL:      c.S3DisableSSL,
			EnableV2Signing: c.S3EnableV2Signing,
			ProductPath:     c.S3ProductPath,
			StemcellPath:    c.S3StemcellPath,
		}

		return NewS3Client(wrapStow{}, config, progressWriter)
	}

	commands.RegisterProductClient("s3", initializer)
}
