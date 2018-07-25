package commands

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/config"
	"gopkg.in/yaml.v2"
)

//go:generate counterfeiter -o ./fakes/create_vm_extension_service.go --fake-name CreateVMExtensionService . createVMExtensionService
type createVMExtensionService interface {
	CreateStagedVMExtension(api.CreateVMExtension) error
}

type CreateVMExtension struct {
	service createVMExtensionService
	logger  logger
	Options struct {
		Name            string   `long:"name"               short:"n"   description:"VM extension name"`
		ConfigFile      string   `long:"config"             short:"c"   description:"path to yml file containing all config fields (see docs/create-vm-extension/README.md for format)"`
		VarsFile        []string `long:"vars-file"          short:"l"   description:"Load variables from a YAML file"`
		OpsFile         []string `long:"ops-file"           short:"o"   description:"YAML operations file"`
		CloudProperties string   `long:"cloud-properties"   short:"cp"  description:"cloud properties in JSON format"`
	}
}

func NewCreateVMExtension(service createVMExtensionService, logger logger) CreateVMExtension {
	return CreateVMExtension{
		service: service,
		logger:  logger,
	}
}

func (c CreateVMExtension) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse create-vm-extension flags: %s", err)
	}

	var (
		name            string
		cloudProperties json.RawMessage
	)
	if c.Options.ConfigFile != "" {
		var cfg config.VMExtenstionConfig
		configContents, err := interpolate(c.Options.ConfigFile, c.Options.VarsFile, c.Options.OpsFile)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(configContents, &cfg)
		if err != nil {
			return fmt.Errorf("%s could not be parsed as valid configuration: %s", c.Options.ConfigFile, err)
		}

		if cfg.VMExtension.Name == "" {
			return errors.New("Config file must contain name element")
		}
		name = cfg.VMExtension.Name

		cp, err := getJSONProperties(cfg.VMExtension.CloudProperties)
		if err != nil {
			return err
		}
		cloudProperties = json.RawMessage(cp)

	} else {
		if c.Options.Name == "" {
			return errors.New("VM Extension name must provide name via --name flag")
		}
		name = c.Options.Name
		cloudProperties = json.RawMessage(c.Options.CloudProperties)
	}

	err := c.service.CreateStagedVMExtension(api.CreateVMExtension{
		Name:            name,
		CloudProperties: cloudProperties,
	})

	if err != nil {
		return err
	}

	c.logger.Printf("VM Extension '%s' created/updated\n", name)

	return nil
}

func (c CreateVMExtension) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This creates/updates a VM extension",
		ShortDescription: "creates/updates a VM extension",
		Flags:            c.Options,
	}
}
