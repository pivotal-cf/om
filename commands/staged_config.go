package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	yaml "gopkg.in/yaml.v2"
)

type StagedConfig struct {
	logger  logger
	service stagedConfigService
	Options struct {
		Product            string `long:"product-name" short:"p" required:"true" description:"name of product"`
		IncludeCredentials bool   `short:"c" long:"include-credentials" description:"include credentials. note: requires product to have been deployed"`
	}
}

//go:generate counterfeiter -o ./fakes/staged_config_service.go --fake-name StagedConfigService . stagedConfigService
type stagedConfigService interface {
	GetDeployedProductCredential(input api.GetDeployedProductCredentialInput) (api.GetDeployedProductCredentialOutput, error)
	GetStagedProductByName(product string) (api.StagedProductsFindOutput, error)
	GetStagedProductJobResourceConfig(productGUID, jobGUID string) (api.JobProperties, error)
	GetStagedProductNetworksAndAZs(product string) (map[string]interface{}, error)
	GetStagedProductProperties(product string) (map[string]api.ResponseProperty, error)
	ListDeployedProducts() ([]api.DeployedProductOutput, error)
	ListStagedProductJobs(productGUID string) (map[string]string, error)
}

func NewStagedConfig(service stagedConfigService, logger logger) StagedConfig {
	return StagedConfig{
		logger:  logger,
		service: service,
	}
}

func (ec StagedConfig) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command generates a config from a staged product that can be passed in to om configure-product (Note: credentials are not available and will appear as '***')",
		ShortDescription: "**EXPERIMENTAL** generates a config from a staged product",
		Flags:            ec.Options,
	}
}

func (ec StagedConfig) Execute(args []string) error {
	if _, err := jhanda.Parse(&ec.Options, args); err != nil {
		return fmt.Errorf("could not parse staged-config flags: %s", err)
	}

	if ec.Options.IncludeCredentials {
		deployedProducts, err := ec.service.ListDeployedProducts()
		if err != nil {
			return err
		}
		var productDeployed bool
		for _, p := range deployedProducts {
			if p.Type == ec.Options.Product {
				productDeployed = true
				break
			}
		}
		if !productDeployed {
			return fmt.Errorf("cannot retrieve credentials for product '%s': deploy the product and retry", ec.Options.Product)
		}
	}

	findOutput, err := ec.service.GetStagedProductByName(ec.Options.Product)
	if err != nil {
		return err
	}
	productGUID := findOutput.Product.GUID

	properties, err := ec.service.GetStagedProductProperties(productGUID)
	if err != nil {
		return err
	}

	configurableProperties := map[string]interface{}{}

	for name, property := range properties {
		if property.Configurable && property.Value != nil {
			if property.IsCredential && ec.Options.IncludeCredentials {
				output, err := ec.service.GetDeployedProductCredential(api.GetDeployedProductCredentialInput{
					DeployedGUID:        productGUID,
					CredentialReference: name,
				})
				if err != nil {
					return err
				}
				configurableProperties[name] = map[string]interface{}{"value": output.Credential.Value}
			} else {
				configurableProperties[name] = map[string]interface{}{"value": property.Value}
			}
		}
	}

	networks, err := ec.service.GetStagedProductNetworksAndAZs(productGUID)
	if err != nil {
		return err
	}

	jobs, err := ec.service.ListStagedProductJobs(productGUID)
	if err != nil {
		return err
	}

	resourceConfig := map[string]api.JobProperties{}

	for name, jobGUID := range jobs {
		jobProperties, err := ec.service.GetStagedProductJobResourceConfig(productGUID, jobGUID)
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
