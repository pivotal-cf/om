package commands

import (
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/common"
	"github.com/pivotal-cf/om/flags"
)

type UploadProduct struct {
	multipart      multipart
	logger         common.Logger
	productService productService
	Options        struct {
		Product string `short:"p"  long:"product"  description:"path to product"`
	}
}

//go:generate counterfeiter -o ./fakes/product.go --fake-name ProductService . productService
type productService interface {
	Upload(api.UploadProductInput) (api.UploadProductOutput, error)
	Stage(api.StageProductInput) error
}

func NewUploadProduct(multipart multipart, productService productService, logger common.Logger) UploadProduct {
	return UploadProduct{
		multipart:      multipart,
		logger:         logger,
		productService: productService,
	}
}

func (up UploadProduct) Usage() Usage {
	return Usage{
		Description:      "This command attempts to upload a product to the Ops Manager",
		ShortDescription: "uploads a given product to the Ops Manager targeted",
		Flags:            up.Options,
	}
}

func (up UploadProduct) Execute(args []string) error {
	_, err := flags.Parse(&up.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse upload-product flags: %s", err)
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

	_, err = up.productService.Upload(api.UploadProductInput{
		ContentLength: submission.Length,
		Product:       submission.Content,
		ContentType:   submission.ContentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload product: %s", err)
	}

	up.logger.Printf("finished upload")

	return nil
}
