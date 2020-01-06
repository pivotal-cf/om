package commands

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/pivotal-cf/om/api"
	"sort"
	"strings"

	"github.com/pivotal-cf/jhanda"
)

type ProductDiff struct {
	service productDiffService
	logger  logger
	Options struct {
		Product  []string `long:"product" short:"p" description:"Product to get diff for. Pass repeatedly for multiple products. If excluded, all staged non-director products will be shown."`
		Director bool     `long:"director" short:"d" description:"Include director diffs. Can be combined with --product."`
	}
}

//counterfeiter:generate -o ./fakes/diff_service.go --fake-name ProductDiffService . productDiffService
type productDiffService interface {
	DirectorDiff() (api.DirectorDiff, error)
	ProductDiff(productName string) (api.ProductDiff, error)
	ListStagedProducts() (api.StagedProductsOutput, error)
}

func NewProductDiff(service productDiffService, logger logger) ProductDiff {
	return ProductDiff{
		service: service,
		logger:  logger,
	}
}

func (c ProductDiff) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse product-diff flags: %s", err)
	}

	var diffableProducts []string

	showDirectorAndProducts := !c.Options.Director && len(c.Options.Product) == 0

	if c.Options.Director || showDirectorAndProducts {
		diff, err := c.service.DirectorDiff()
		if err != nil {
			panic(err)
		}
		c.logger.Println("## Director Manifest\n")
		notInstalled := c.printManifestDiff(diff.Manifest)
		if !notInstalled {
			c.logger.Println("## Director Cloud Config\n")
			c.printManifestDiff(diff.CloudConfig)
			c.logger.Println("## Director Runtime Configs\n")
			c.printRuntimeConfigs(diff.RuntimeConfigs)
			c.logger.Println("## Director CPI Configs\n")
			c.printCPIConfigs(diff.CPIConfigs)
		}
	}

	if showDirectorAndProducts {
		stagedProducts, err := c.service.ListStagedProducts()
		if err != nil {
			return fmt.Errorf("could not discover staged products to diff: %s", err)
		}

		for _, product := range stagedProducts.Products {
			if product.Type != "p-bosh" {
				diffableProducts = append(diffableProducts, product.Type)
			}
		}
		sort.Strings(diffableProducts)
	} else {
		diffableProducts = c.Options.Product
	}

	for _, product := range diffableProducts {
		diff, err := c.service.ProductDiff(product)
		if err != nil {
			return err
		}

		c.logger.Printf("## Product Manifest for %s\n\n", product)

		notInstalled := c.printManifestDiff(diff.Manifest)
		if notInstalled {
			continue
		}
		c.logger.Printf("## Runtime Configs for %s\n\n", product)
		c.printRuntimeConfigs(diff.RuntimeConfigs)
	}
	return nil
}

func (c ProductDiff) printManifestDiff(diff api.ManifestDiff) (bool) {
	switch diff.Status {
	case "same":
		c.logger.Println("no changes\n")
	case "does_not_exist":
		c.logger.Println("no manifest for this product\n")
	case "different":
		c.logger.Printf("%s\n\n", c.colorizeDiff(diff.Diff))
	case "to_be_installed":
		c.logger.Println("This product is not yet deployed, so the product and runtime diffs are not available.")
		return true
	default:
		c.logger.Printf("unrecognized product status: %s\n\n%s\n\n", diff.Status, diff.Diff)
	}
	return false
}

func (c ProductDiff) printRuntimeConfigs(configs []api.RuntimeConfigsDiff) {
	noneChanged := true

	for _, config := range configs {
		if config.Status == "same" {
			continue
		}

		noneChanged = false

		c.logger.Printf("### %s\n\n", config.Name)
		c.logger.Printf("%s\n\n", c.colorizeDiff(config.Diff))
	}

	if noneChanged {
		c.logger.Println("no changes\n")
	}
}

func (c ProductDiff) printCPIConfigs(configs []api.CPIConfigsDiff) {
	noneChanged := true

	for _, config := range configs {
		if config.Status == "same" {
			continue
		}

		noneChanged = false

		c.logger.Printf("### %s\n\n", config.IAASConfigurationName)
		c.logger.Printf("%s\n\n", c.colorizeDiff(config.Diff))
	}

	if noneChanged {
		c.logger.Println("no changes\n")
	}
}

func (c ProductDiff) colorizeDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	for index, line := range lines {
		if strings.HasPrefix(line, "-") {
			lines[index] = color.RedString(line)
		}
		if strings.HasPrefix(line, "+") {
			lines[index] = color.GreenString(line)
		}
	}
	return strings.Join(lines, "\n")
}

func (c ProductDiff) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "**EXPERIMENTAL** This command displays the bosh manifest diff for products (Note: secret values are replaced with double-paren variable names)",
		ShortDescription: "**EXPERIMENTAL** displays BOSH manifest diff for products",
		Flags:            c.Options,
	}
}
