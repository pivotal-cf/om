package download_clients

import (
	"fmt"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/s3"
	"gopkg.in/go-playground/validator.v9"
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

func NewS3Client(stower Stower, config S3Configuration, stderr *log.Logger) (stowClient, error) {
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

	return NewStowClient(stower, stderr, stowConfig, config.ProductPath, config.StemcellPath, "s3", config.Bucket), nil
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
