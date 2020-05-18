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
	environFunc   envProvider
	buildProvider buildProvider
	Options       struct {
		ConfigFile string   `long:"config"                     short:"c" description:"path to yml file for configuration (keys must match the following command line flags)"`
		VarsEnv    []string `long:"vars-env" env:"OM_VARS_ENV"           description:"load variables from environment variables matching the provided prefix (e.g.: 'MY' to load MY_var=value)"`
		VarsFile   []string `long:"vars-file"                  short:"l" description:"load variables from a YAML file"`
		Vars       []string `long:"var"                        short:"v" description:"Load variable from the command line. Format: VAR=VAL"`

		PivnetApiToken    string `long:"pivnet-api-token"`
		PivnetProductSlug string `long:"pivnet-product-slug"                          description:"the product name in pivnet"`
		ProductVersion    string `long:"product-version"                              description:"the version of the product from which to generate a template"`
		FileGlob          string `long:"file-glob" short:"f" alias:"pivnet-file-glob" description:"a glob to match exactly one file in the pivnet product slug"  default:"*.pivotal"`
		PivnetDisableSSL  bool   `long:"pivnet-disable-ssl"                           description:"whether to disable ssl validation when contacting the Pivotal Network"`

		ProductPath string `long:"product-path" description:"path to product file"`

		OutputDirectory   string `long:"output-directory" description:"a directory to create templates under. must already exist." required:"true"`
		ExcludeVersion    bool   `long:"exclude-version"  description:"if set, will not output a version-specific directory"`
		SizeOfCollections int    `long:"size-of-collections" default:"10"`
	}
}

//counterfeiter:generate -o ./fakes/metadata_provider.go --fake-name MetadataProvider . MetadataProvider
type MetadataProvider interface {
	MetadataBytes() ([]byte, error)
}

var pivnetHost = pivnet.DefaultHost
var DefaultProvider = func() func(c *ConfigTemplate) MetadataProvider {
	return func(c *ConfigTemplate) MetadataProvider {
		options := c.Options
		if options.ProductPath != "" {
			return metadata.NewFileProvider(options.ProductPath)
		}
		return metadata.NewPivnetProvider(pivnetHost, options.PivnetApiToken, options.PivnetProductSlug, options.ProductVersion, options.FileGlob, options.PivnetDisableSSL)
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
	err := loadConfigFile(args, &c.Options, c.environFunc)
	if err != nil {
		return fmt.Errorf("could not parse config-template flags: %s", err.Error())
	}

	_, err = os.Stat(c.Options.OutputDirectory)
	if os.IsNotExist(err) {
		return fmt.Errorf("output-directory does not exist: %s", c.Options.OutputDirectory)
	}

	err = c.Validate()
	if err != nil {
		return err
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
	).Generate()
}

func (c *ConfigTemplate) newMetadataSource() (metadataSource MetadataProvider) {
	return c.buildProvider(c)
}

func (c *ConfigTemplate) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "this command generates a product configuration template from a .pivotal file on Pivnet",
		ShortDescription: "generates a config template from a Pivnet product",
		Flags:            c.Options,
	}
}

func (c *ConfigTemplate) Validate() error {
	if c.Options.PivnetApiToken != "" && c.Options.PivnetProductSlug != "" && c.Options.ProductVersion != "" && c.Options.ProductPath == "" {
		return nil
	}

	if c.Options.PivnetApiToken == "" && c.Options.PivnetProductSlug == "" && c.Options.ProductVersion == "" && c.Options.ProductPath != "" {
		return nil
	}

	return fmt.Errorf("cannot load tile metadata: please provide either pivnet flags OR product-path")
}
