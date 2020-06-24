package download_clients

import (
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/azure"
	"gopkg.in/go-playground/validator.v9"
	"io"
)

type AzureConfiguration struct {
	StorageAccount string `validate:"required"`
	Key            string `validate:"required"`
	Container      string `validate:"required"`
	ProductPath    string
	StemcellPath   string
}

func NewAzureClient(stower Stower, config AzureConfiguration, progressWriter io.Writer) (stowClient, error) {
	validate := validator.New()
	err := validate.Struct(config)
	if err != nil {
		return stowClient{}, err
	}

	stowConfig := stow.ConfigMap{
		azure.ConfigAccount: config.StorageAccount,
		azure.ConfigKey:     config.Key,
	}

	return NewStowClient(
		stower,
		config.Container,
		stowConfig,
		progressWriter,
		config.ProductPath,
		config.StemcellPath,
		"azure",
	), nil
}
