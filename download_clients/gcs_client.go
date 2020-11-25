package download_clients

import (
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/google"
	storage "google.golang.org/api/storage/v1beta2"
	"gopkg.in/go-playground/validator.v9"
	"log"
)

type GCSConfiguration struct {
	Bucket             string `validate:"required"`
	ServiceAccountJSON string `validate:"required"`
	ProjectID          string `validate:"required"`
	ProductPath        string
	StemcellPath       string
}

func NewGCSClient(stower Stower, config GCSConfiguration, stderr *log.Logger) (stowClient, error) {
	validate := validator.New()
	err := validate.Struct(config)
	if err != nil {
		return stowClient{}, err
	}

	stowConfig := stow.ConfigMap{
		google.ConfigJSON:      config.ServiceAccountJSON,
		google.ConfigProjectId: config.ProjectID,
		google.ConfigScopes:    storage.DevstorageReadOnlyScope,
	}

	return NewStowClient(stower, stderr, stowConfig, config.ProductPath, config.StemcellPath, "google", config.Bucket), nil
}
