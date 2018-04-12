package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	yaml "gopkg.in/yaml.v2"
)

type StagedConfig struct {
	logger              logger
	stagedConfigService stagedConfigService
	Options             struct {
		Product string `long:"product-name"    short:"p" required:"true" description:"name of product"`
	}
}

//go:generate counterfeiter -o ./fakes/export_config_service.go --fake-name StagedConfigService . stagedConfigService
type stagedConfigService interface {
	Find(product string) (api.StagedProductsFindOutput, error)
	Jobs(productGUID string) (map[string]string, error)
	GetExistingJobConfig(productGUID, jobGUID string) (api.JobProperties, error)
	Properties(product string) (map[string]api.ResponseProperty, error)
	NetworksAndAZs(product string) (map[string]interface{}, error)
}

func NewStagedConfig(stagedConfigService stagedConfigService, logger logger) StagedConfig {
	return StagedConfig{
		logger:              logger,
		stagedConfigService: stagedConfigService,
	}
}

func (ec StagedConfig) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command generates a config from a staged product that can be passed in to om configure-product (Note: credentials are not available and will appear as '***')",
		ShortDescription: "generates a config from a staged product",
		Flags:            ec.Options,
	}
}

func (ec StagedConfig) Execute(args []string) error {
	if _, err := jhanda.Parse(&ec.Options, args); err != nil {
		return fmt.Errorf("could not parse staged-config flags: %s", err)
	}

	findOutput, err := ec.stagedConfigService.Find(ec.Options.Product)
	if err != nil {
		return err
	}
	productGUID := findOutput.Product.GUID

	properties, err := ec.stagedConfigService.Properties(productGUID)
	if err != nil {
		return err
	}

	configurableProperties := map[string]interface{}{}

	for name, property := range properties {
		if property.Configurable && property.Value != nil {
			configurableProperties[name] = map[string]interface{}{"value": property.Value}
		}
	}

	networks, err := ec.stagedConfigService.NetworksAndAZs(productGUID)
	if err != nil {
		return err
	}

	jobs, err := ec.stagedConfigService.Jobs(productGUID)
	if err != nil {
		return err
	}

	resourceConfig := map[string]api.JobProperties{}

	for name, jobGUID := range jobs {
		jobProperties, err := ec.stagedConfigService.GetExistingJobConfig(productGUID, jobGUID)
		if err != nil {
			return err
		}

		resourceConfig[name] = jobProperties
	}

	config := struct {
		Properties               map[string]interface{}       `yaml:"product-properties"`
		NetworkProperties        map[string]interface{}       `yaml:"network-properties"`
		ResourceConfigProperties map[string]api.JobProperties `yaml:"resource-config"`
	}{
		Properties:               configurableProperties,
		NetworkProperties:        networks,
		ResourceConfigProperties: resourceConfig,
	}

	output, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %s", err) // un-tested
	}
	ec.logger.Println(string(output))

	return nil
}
