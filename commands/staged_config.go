package commands

import (
	"fmt"

	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/config"
	"github.com/pivotal-cf/om/configparser"
	"gopkg.in/yaml.v2"
)

type StagedConfig struct {
	service stagedConfigService
	logger  logger
	Options struct {
		Product             string `long:"product-name" short:"p" required:"true" description:"name of product"`
		IncludeCredentials  bool   `long:"include-credentials" short:"c" description:"include credentials. note: requires product to have been deployed"`
		IncludePlaceholders bool   `long:"include-placeholders" short:"r" description:"replace obscured credentials with interpolatable placeholders"`
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
	ListStagedProductErrands(productID string) (api.ErrandsListOutput, error)
}

func NewStagedConfig(service stagedConfigService, logger logger) StagedConfig {
	return StagedConfig{
		service: service,
		logger:  logger,
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
	selectorProperties := map[string]string{}

	for name, property := range properties {
		if property.Value == nil {
			continue
		}
		if property.Type == "selector" {
			value := property.SelectedOption
			if value == "" {
				value = property.Value.(string)
			}
			selectorProperties[name] = value
		}
		var output map[string]interface{}

		parser := configparser.NewConfigParser()
		propertyName := configparser.NewPropertyName(name)
		output, err = parser.ParseProperties(propertyName, property, ec.chooseCredentialHandler(productGUID))

		if err != nil {
			return err
		}
		if output != nil && len(output) > 0 {
			configurableProperties[name] = output
		}
	}

	for name := range configurableProperties {
		components := strings.Split(name, ".")[1:] // the 0th item is an empty string due to `.some.other`
		if len(components) == 2 {
			continue
		}
		selector := "." + strings.Join(components[:2], ".")
		if val, ok := selectorProperties[selector]; ok && components[2] != val {
			delete(configurableProperties, name)
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

	resourceConfig := map[string]interface{}{}

	for name, jobGUID := range jobs {
		jobProperties, err := ec.service.GetStagedProductJobResourceConfig(productGUID, jobGUID)
		if err != nil {
			return err
		}
		resourceConfig[name] = jobProperties
	}

	errandsListOutput, err := ec.service.ListStagedProductErrands(productGUID)
	if err != nil {
		return err
	}

	errandConfigs := map[string]config.ErrandConfig{}

	for _, errand := range errandsListOutput.Errands {
		errandConfig := config.ErrandConfig{}
		errandConfig.PostDeployState = errand.PostDeploy
		errandConfig.PreDeleteState = errand.PreDelete

		errandConfigs[errand.Name] = errandConfig
	}

	config := config.ProductConfiguration{
		ProductName:              ec.Options.Product,
		ProductProperties:        configurableProperties,
		NetworkProperties:        networks,
		ResourceConfigProperties: resourceConfig,
		ErrandConfigs:            errandConfigs,
	}

	output, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %s", err) // un-tested
	}

	ec.logger.Println(string(output))
	return nil
}

func (ec StagedConfig) chooseCredentialHandler(productGUID string) configparser.CredentialHandler {
	if ec.Options.IncludePlaceholders {
		return configparser.PlaceholderHandler()
	}

	if ec.Options.IncludeCredentials {
		return configparser.GetCredentialHandler(productGUID, ec.service)
	}

	return configparser.NilHandler()
}
