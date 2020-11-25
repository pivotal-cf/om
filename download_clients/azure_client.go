package download_clients

import (
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/azure"
	"gopkg.in/go-playground/validator.v9"
	"log"
)

type AzureConfiguration struct {
	StorageAccount string `validate:"required"`
	Key            string `validate:"required"`
	Container      string `validate:"required"`
	ProductPath    string
	StemcellPath   string
}

func NewAzureClient(stower Stower, config AzureConfiguration, stderr *log.Logger) (stowClient, error) {
	validate := validator.New()
	err := validate.Struct(config)
	if err != nil {
		return stowClient{}, err
	}

	stowConfig := stow.ConfigMap{
		azure.ConfigAccount: config.StorageAccount,
		azure.ConfigKey:     config.Key,
	}

	return NewStowClient(stower, stderr, stowConfig, config.ProductPath, config.StemcellPath, "azure", config.Container), nil
}
