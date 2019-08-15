package commands

import (
	"fmt"
	"os"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/configtemplate/generator"
	"github.com/pivotal-cf/om/configtemplate/metadata"
)

type ConfigTemplate struct {
	environFunc   func() []string
	buildProvider buildProvider
	Options       struct {
		OutputDirectory   string `long:"output-directory"    description:"a directory to create templates under. must already exist."                       required:"true"`
		PivnetApiToken    string `long:"pivnet-api-token"                                                                                                   required:"true"`
		PivnetProductSlug string `long:"pivnet-product-slug" description:"the product name in pivnet"                                                       required:"true"`
		ProductVersion    string `long:"product-version"     description:"the version of the product from which to generate a template"                     required:"true"`
		ProductFileGlob   string `long:"product-file-glob"   description:"a glob to match exactly one file in the pivnet product slug"  default:"*.pivotal" `
		ExcludeVersion    bool   `long:"exclude-version"     description:"if set, will not output a version-specific directory"`
	}
}

//go:generate counterfeiter -o ./fakes/metadata_provider.go --fake-name MetadataProvider . MetadataProvider
type MetadataProvider interface {
	MetadataBytes() ([]byte, error)
}

var pivnetHost = pivnet.DefaultHost
var DefaultProvider = func() func(c *ConfigTemplate) MetadataProvider {
	return func(c *ConfigTemplate) MetadataProvider {
		options := c.Options
		return metadata.NewPivnetProvider(
			pivnetHost,
			options.PivnetApiToken,
			options.PivnetProductSlug,
			options.ProductVersion,
			options.ProductFileGlob,
		)
	}
}

type buildProvider func(*ConfigTemplate) MetadataProvider

func NewConfigTemplate(bp buildProvider) *ConfigTemplate {
	return &ConfigTemplate{
		environFunc:   os.Environ,
		buildProvider: bp,
	}
}

//Execute - generates config template and ops files
func (c *ConfigTemplate) Execute(args []string) error {
	_, err := jhanda.Parse(
		&c.Options,
		args)
	if err != nil {
		return fmt.Errorf("could not parse config-template flags: %s", err.Error())
	}

	_, err = os.Stat(c.Options.OutputDirectory)
	if os.IsNotExist(err) {
		return fmt.Errorf("output-directory does not exist: %s", c.Options.OutputDirectory)
	}

	metadataSource := c.newMetadataSource()
	metadataBytes, err := metadataSource.MetadataBytes()
	if err != nil {
		return fmt.Errorf("error getting metadata for %s at version %s: %s", c.Options.PivnetProductSlug, c.Options.ProductVersion, err)
	}

	return generator.NewExecutor(
		metadataBytes,
		c.Options.OutputDirectory,
		c.Options.ExcludeVersion,
		true,
	).Generate()
}

func (c *ConfigTemplate) newMetadataSource() (metadataSource MetadataProvider) {
	return c.buildProvider(c)
}

func (c *ConfigTemplate) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "**EXPERIMENTAL** this command generates a product configuration template from a .pivotal file on Pivnet",
		ShortDescription: "**EXPERIMENTAL** generates a config template from a Pivnet product",
		Flags:            c.Options,
	}
}
