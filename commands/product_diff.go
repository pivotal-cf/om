package commands

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/pivotal-cf/om/api"
	"strings"

	"github.com/pivotal-cf/jhanda"
)

type ProductDiff struct {
	service productDiffService
	logger  logger
	Options struct {
		Product []string `long:"product" short:"p" required:"true" description:"Product to get diff for"`
	}
}

//counterfeiter:generate -o ./fakes/diff_service.go --fake-name ProductDiffService . productDiffService
type productDiffService interface {
	ProductDiff(productName string) (api.ProductDiff, error)
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
	for _, product := range c.Options.Product {

		diff, err := c.service.ProductDiff(product)
		if err != nil {
			return err
		}

		c.logger.Printf("## Product Manifest for %s\n\n", product)

		switch diff.Manifest.Status {
		case "same":
			c.logger.Println("no changes\n")
		case "does_not_exist":
			c.logger.Println("no manifest for this product\n")
		case "different":
			c.logger.Printf("%s\n\n", c.colorize(diff.Manifest.Diff))
		case "to_be_installed":
			c.logger.Println("This product is not yet deployed, so the product and runtime diffs are not available.")
			return nil
		default:
			c.logger.Printf("unrecognized product status: %s\n\n%s\n\n", diff.Manifest.Status, diff.Manifest.Diff)
		}
		c.printRuntimeConfigs(diff, product)
	}
	return nil
}

func (c ProductDiff) printRuntimeConfigs(diff api.ProductDiff, product string) {
	c.logger.Printf("## Runtime Configs for %s\n\n", product)

	noneChanged := true

	for _, config := range diff.RuntimeConfigs {
		if config.Status == "same" {
			continue
		}

		noneChanged = false

		c.logger.Printf("### %s\n\n", config.Name)
		c.logger.Printf("%s\n\n", c.colorize(config.Diff))
	}

	if noneChanged {
		c.logger.Printf("no changes")
	}
}

func (c ProductDiff) colorize(diff string) string {
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
		Description:      "**EXPERIMENTAL** This command displays the bosh manifest diff for a product (Note: secret values are replaced with double-paren variable names)",
		ShortDescription: "**EXPERIMENTAL** displays BOSH manifest diff for a product",
		Flags:            c.Options,
	}
}
