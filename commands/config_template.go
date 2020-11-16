package commands

import (
	"fmt"
	"os"

	"github.com/pivotal-cf/om/configtemplate/generator"
	"github.com/pivotal-cf/om/configtemplate/metadata"
)

type ConfigTemplate struct {
	environFunc   envProvider
	buildProvider buildProvider
	Options       struct {
		InterpolateOptions interpolateConfigFileOptions `group:"config file interpolation"`

		PivnetApiToken    string `long:"pivnet-api-token"`
		PivnetProductSlug string `long:"pivnet-product-slug"                          description:"the product name in pivnet"`
		ProductVersion    string `long:"product-version"                              description:"the version of the product from which to generate a template"`
		PivnetHost        string `long:"pivnet-host" description:"the API endpoint for Pivotal Network" default:"https://network.pivotal.io"`
		FileGlob          string `long:"file-glob" short:"f" description:"a glob to match exactly one file in the pivnet product slug"  default:"*.pivotal"`
		PivnetDisableSSL  bool   `long:"pivnet-disable-ssl"                           description:"whether to disable ssl validation when contacting the Pivotal Network"`

		ProductPath string `long:"product-path" description:"path to product file"`

		OutputDirectory   string `long:"output-directory" description:"a directory to create templates under. must already exist." required:"true"`
		ExcludeVersion    bool   `long:"exclude-version"  description:"if set, will not output a version-specific directory"`
		SizeOfCollections int    `long:"size-of-collections"`

		PivnetFileGlobSupport string `long:"pivnet-file-glob" hidden:"true"`
	}
}

//counterfeiter:generate -o ./fakes/metadata_provider.go --fake-name MetadataProvider . MetadataProvider
type MetadataProvider interface {
	MetadataBytes() ([]byte, error)
}

var DefaultProvider = func() func(c *ConfigTemplate) MetadataProvider {
	return func(c *ConfigTemplate) MetadataProvider {
		options := c.Options
		if options.ProductPath != "" {
			return metadata.NewFileProvider(options.ProductPath)
		}
		return metadata.NewPivnetProvider(options.PivnetHost, options.PivnetApiToken, options.PivnetProductSlug, options.ProductVersion, options.FileGlob, options.PivnetDisableSSL)
	}
}

type buildProvider func(*ConfigTemplate) MetadataProvider
type envProvider func() []string

func NewConfigTemplate(bp buildProvider) *ConfigTemplate {
	return NewConfigTemplateWithEnvironment(bp, os.Environ)
}

func NewConfigTemplateWithEnvironment(bp buildProvider, environFunc envProvider) *ConfigTemplate {
	return &ConfigTemplate{
		environFunc:   environFunc,
		buildProvider: bp,
	}
}

//Execute - generates config template and ops files
func (c *ConfigTemplate) Execute(args []string) error {
	_, err := os.Stat(c.Options.OutputDirectory)
	if os.IsNotExist(err) {
		return fmt.Errorf("output-directory does not exist: %s", c.Options.OutputDirectory)
	}

	err = c.Validate()
	if err != nil {
		return err
	}

	var userSetSizeOfCollections bool
	if c.Options.SizeOfCollections > 0 {
		userSetSizeOfCollections = true
	} else {
		c.Options.SizeOfCollections = 10
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
		c.Options.SizeOfCollections,
		userSetSizeOfCollections,
	).Generate()
}

func (c *ConfigTemplate) newMetadataSource() (metadataSource MetadataProvider) {
	return c.buildProvider(c)
}

func (c *ConfigTemplate) Validate() error {
	if c.Options.PivnetFileGlobSupport != "" {
		c.Options.FileGlob = c.Options.PivnetFileGlobSupport
	}

	if c.Options.PivnetApiToken != "" && c.Options.PivnetProductSlug != "" && c.Options.ProductVersion != "" && c.Options.ProductPath == "" {
		return nil
	}

	if c.Options.PivnetApiToken == "" && c.Options.PivnetProductSlug == "" && c.Options.ProductVersion == "" && c.Options.ProductPath != "" {
		return nil
	}

	return fmt.Errorf("cannot load tile metadata: please provide either pivnet flags OR product-path")
}
