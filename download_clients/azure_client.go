package download_clients

import (
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/azure"
	"github.com/pivotal-cf/om/commands"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"log"
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

func init() {
	initializer := func(
		c commands.DownloadProductOptions,
		progressWriter io.Writer,
		_ *log.Logger,
		_ *log.Logger,
	) (commands.ProductDownloader, error) {
		config := AzureConfiguration{
			Container:      c.Bucket,
			StorageAccount: c.AzureStorageAccount,
			Key:            c.AzureKey,
			ProductPath:    c.ProductPath,
			StemcellPath:   c.StemcellPath,
		}

		return NewAzureClient(wrapStow{}, config, progressWriter)
	}

	commands.RegisterProductClient("azure", initializer)
}
