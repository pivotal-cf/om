package commands

import (
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"os"
	//"github.com/pivotalservices/tile-config-generator/generator"
)

type ConfigTemplate struct {
	environFunc func() []string
	Options     struct {
		ConfigFile      string `long:"config" short:"c" description:"path to yml file for configuration (keys must match the following command line flags)"`
		OutputDirectory string `long:"output-directory" required:"true"`
	}
}

func NewConfigTemplate() *ConfigTemplate {
	return &ConfigTemplate{
		environFunc: os.Environ,
	}
}

//Execute - generates config template and ops files
func (c *ConfigTemplate) Execute(flags []string) error {
	err := loadConfigFile(flags, &c.Options, c.environFunc)
	if err != nil {
		return fmt.Errorf("could not parse config-template flags: %s", err.Error())
	}
	//metadataSource := c.newPivnetMetadataSource()

	//metadataBytes, err := metadataSource.MetadataBytes()
	//if err != nil {
	//	return err
	//}
	//return generator.NewExecutor(metadataBytes, c.BaseDirectory, c.DoNotIncludeProductVersion, c.IncludeErrands).Generate()
	return nil
}

//func (c *ConfigTemplate) newPivnetMetadataSource() (metadataSource Provider) {
//	return metadataSource
//}

//type Provider interface {
//	MetadataBytes() ([]byte, error)
//}

func (c *ConfigTemplate) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command generates a product configuration template from a .pivotal file on Pivnet",
		ShortDescription: "generates a config template from a Pivnet product",
		Flags:            c.Options,
	}
}
