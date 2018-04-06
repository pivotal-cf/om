package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	yaml "gopkg.in/yaml.v2"
)

type ExportConfig struct {
	logger              logger
	exportConfigService exportConfigService
	Options             struct {
		Product string `long:"product-name"    short:"p" required:"true" description:"name of product"`
	}
}

//go:generate counterfeiter -o ./fakes/export_config_service.go --fake-name ExportConfigService . exportConfigService
type exportConfigService interface {
	ExportConfig(product string) (api.ExportConfigOutput, error)
}

func NewExportConfig(exportConfigService exportConfigService, logger logger) ExportConfig {
	return ExportConfig{
		logger:              logger,
		exportConfigService: exportConfigService,
	}
}

func (ec ExportConfig) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command generates a config from a staged product that can be passed in to om configure-product",
		ShortDescription: "generates a config from a staged product",
		Flags:            ec.Options,
	}
}

func (ec ExportConfig) Execute(args []string) error {
	if _, err := jhanda.Parse(&ec.Options, args); err != nil {
		return fmt.Errorf("could not parse export-config flags: %s", err)
	}

	config, err := ec.exportConfigService.ExportConfig(ec.Options.Product)
	if err != nil {
		return fmt.Errorf("failed to export config: %s", err)
	}
	output, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %s", err) // un-tested
	}
	ec.logger.Println(string(output))

	return nil
}
