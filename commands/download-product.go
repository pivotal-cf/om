package commands

import (
	"fmt"
	"github.com/pivotal-cf/jhanda"
)

type DownloadProduct struct {
	logger logger
	Options struct {
		ConfigFile        string `long:"config"               short:"c"   description:"path to yml file for configuration (keys must match the following command line flags)"`
		PivnetAPIToken    string `long:"pivnet-api-token" required:"true"`
		PivnetFileGlob    string `long:"pivnet-file-glob"     short:"f"   description:"Glob to match files within Pivotal Network product to be downloaded." required:"true"`
		PivnetProductSlug string `long:"pivnet-product-slug"  short:"p"   description:"path to product" required:"true"`
		Version           string `long:"product-version"                  description:"version of the provided product file to be used for validation" required:"true"`
	}
}

func NewDownloadProduct(logger logger) DownloadProduct {
	return DownloadProduct{
		logger: logger,
	}
}

func (c DownloadProduct) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command attempts to download a single product file from Pivotal Network. The API token used must be associated with a user account that has already accepted the EULA for the specified product",
		ShortDescription: "downloads a specified product file from Pivotal Network",
		Flags:            c.Options,
	}
}

func (c DownloadProduct) Execute(args []string) error {
	err := loadConfigFile(args, &c.Options)
	if err != nil {
		return fmt.Errorf("could not parse download-product flags: %s", err)
	}
    c.logger.Printf("Not yet implemented. Have a nice day!")
	return nil
}
