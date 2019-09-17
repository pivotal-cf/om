package download_clients

import (
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/google"
	"github.com/pivotal-cf/om/commands"
	storage "google.golang.org/api/storage/v1beta2"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"log"
)

type GCSConfiguration struct {
	Bucket             string `validate:"required"`
	ServiceAccountJSON string `validate:"required"`
	ProjectID          string `validate:"required"`
	ProductPath        string
	StemcellPath       string
}

func NewGCSClient(stower Stower, config GCSConfiguration, progressWriter io.Writer) (stowClient, error) {
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

	return NewStowClient(
		stower,
		config.Bucket,
		stowConfig,
		progressWriter,
		config.ProductPath,
		config.StemcellPath,
		"google",
	), nil
}

func init() {
	initializer := func(
		c commands.DownloadProductOptions,
		progressWriter io.Writer,
		_ *log.Logger,
		_ *log.Logger,
	) (commands.ProductDownloader, error) {
		config := GCSConfiguration{
			Bucket:             c.GCSBucket,
			ProjectID:          c.GCSProjectID,
			ServiceAccountJSON: c.GCSServiceAccountJSON,
			ProductPath:        c.GCSProductPath,
			StemcellPath:       c.GCSStemcellPath,
		}

		return NewGCSClient(wrapStow{}, config, progressWriter)
	}

	commands.RegisterProductClient("gcs", initializer)
}
