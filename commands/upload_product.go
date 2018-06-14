package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/extractor"
	"github.com/pivotal-cf/om/validator"
)

type UploadProduct struct {
	multipart multipart
	logger    logger
	service   uploadProductService
	Options struct {
		Product         string `long:"product"          short:"p"  required:"true" description:"path to product"`
		PollingInterval int    `long:"polling-interval" short:"pi"                 description:"interval (in seconds) at which to print status" default:"1"`
		Shasum          string `long:"shasum" short:"sha" description:"shasum of the provided product file to be used for validation"`
	}
	metadataExtractor metadataExtractor
}

//go:generate counterfeiter -o ./fakes/upload_product_service.go --fake-name UploadProductService . uploadProductService
type uploadProductService interface {
	UploadAvailableProduct(api.UploadAvailableProductInput) (api.UploadAvailableProductOutput, error)
	CheckProductAvailability(string, string) (bool, error)
}

//go:generate counterfeiter -o ./fakes/metadata_extractor.go --fake-name MetadataExtractor . metadataExtractor
type metadataExtractor interface {
	ExtractMetadata(string) (extractor.Metadata, error)
}

func NewUploadProduct(multipart multipart, metadataExtractor metadataExtractor, service uploadProductService, logger logger) UploadProduct {
	return UploadProduct{
		multipart:         multipart,
		metadataExtractor: metadataExtractor,
		logger:            logger,
		service:           service,
	}
}

func (up UploadProduct) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command attempts to upload a product to the Ops Manager",
		ShortDescription: "uploads a given product to the Ops Manager targeted",
		Flags:            up.Options,
	}
}

func (up UploadProduct) Execute(args []string) error {
	if _, err := jhanda.Parse(&up.Options, args); err != nil {
		return fmt.Errorf("could not parse upload-product flags: %s", err)
	}

	if up.Options.Shasum != "" {
		shaValidator := validator.NewSHA256Calculator()
		shasum, err := shaValidator.Checksum(up.Options.Product)

		if err != nil {
			return err
		}

		if shasum != up.Options.Shasum {
			return fmt.Errorf("expected shasum %s does not match file shasum %s", up.Options.Shasum, shasum)
		}

		up.logger.Printf("expected shasum matches product shasum.")
	}

	metadata, err := up.metadataExtractor.ExtractMetadata(up.Options.Product)
	if err != nil {
		return fmt.Errorf("failed to extract product metadata: %s", err)
	}

	prodAvailable, err := up.service.CheckProductAvailability(metadata.Name, metadata.Version)
	if err != nil {
		return fmt.Errorf("failed to check product availability: %s", err)
	}

	if prodAvailable {
		up.logger.Printf("product %s %s is already uploaded, nothing to be done.", metadata.Name, metadata.Version)
		return nil
	}

	up.logger.Printf("processing product")
	err = up.multipart.AddFile("product[file]", up.Options.Product)
	if err != nil {
		return fmt.Errorf("failed to load product: %s", err)
	}

	submission, err := up.multipart.Finalize()
	if err != nil {
		return fmt.Errorf("failed to create multipart form: %s", err)
	}

	up.logger.Printf("beginning product upload to Ops Manager")

	_, err = up.service.UploadAvailableProduct(api.UploadAvailableProductInput{
		ContentLength:   submission.Length,
		Product:         submission.Content,
		ContentType:     submission.ContentType,
		PollingInterval: up.Options.PollingInterval,
	})
	if err != nil {
		return fmt.Errorf("failed to upload product: %s", err)
	}

	up.logger.Printf("finished upload")

	return nil
}
