package commands

import (
	"encoding/json"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

//go:generate counterfeiter -o ./fakes/vm_extension_creator.go --fake-name VMExtensionCreator . vmExtensionCreator
type vmExtensionCreator interface {
	Create(api.CreateVMExtension) error
}

type CreateVMExtension struct {
	service vmExtensionCreator
	logger  logger
	Options struct {
		Name            string `long:"name"             short:"n"  required:"true" description:"VM extension name"`
		CloudProperties string `long:"cloud-properties" short:"cp" required:"true" description:"cloud properties in JSON format"`
	}
}

func NewCreateVMExtension(service vmExtensionCreator, logger logger) CreateVMExtension {
	return CreateVMExtension{
		service: service,
		logger:  logger,
	}
}

func (c CreateVMExtension) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse create-vm-extension flags: %s", err)
	}

	err := c.service.Create(api.CreateVMExtension{
		Name:            c.Options.Name,
		CloudProperties: json.RawMessage(c.Options.CloudProperties),
	})

	if err != nil {
		return err
	}

	c.logger.Printf("VM Extension '%s' created\n", c.Options.Name)

	return nil
}

func (c CreateVMExtension) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This creates a VM extension",
		ShortDescription: "creates a VM extension",
		Flags:            c.Options,
	}
}
