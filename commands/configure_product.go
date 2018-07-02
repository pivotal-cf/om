package commands

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"

	yamlConverter "github.com/ghodss/yaml"
	"gopkg.in/yaml.v2"
)

type ConfigureProduct struct {
	service configureProductService
	logger  logger
	Options struct {
		ProductName       string   `long:"product-name"       short:"n"  required:"true" description:"name of the product being configured"`
		ConfigFile        string   `long:"config"             short:"c"                  description:"path to yml file containing all config fields (see docs/configure-product/README.md for format)"`
		VarsFile          []string `long:"vars-file"  description:"Load variables from a YAML file"`
		OpsFile           []string `long:"ops-file"  description:"YAML operations file"`
		ProductProperties string   `long:"product-properties" short:"p"                  description:"properties to be configured in JSON format"`
		NetworkProperties string   `long:"product-network"    short:"pn"                 description:"network properties in JSON format"`
		ProductResources  string   `long:"product-resources"  short:"pr"                 description:"resource configurations in JSON format"`
	}
}

//go:generate counterfeiter -o ./fakes/configure_product_service.go --fake-name ConfigureProductService . configureProductService
type configureProductService interface {
	ListStagedProducts() (api.StagedProductsOutput, error)
	ListStagedProductJobs(productGUID string) (map[string]string, error)
	GetStagedProductJobResourceConfig(productGUID, jobGUID string) (api.JobProperties, error)
	UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput) error
	UpdateStagedProductNetworksAndAZs(api.UpdateStagedProductNetworksAndAZsInput) error
	UpdateStagedProductJobResourceConfig(productGUID, jobGUID string, jobProperties api.JobProperties) error
}

func NewConfigureProduct(service configureProductService, logger logger) ConfigureProduct {
	return ConfigureProduct{
		service: service,
		logger:  logger,
	}
}

func (cp ConfigureProduct) Execute(args []string) error {
	if _, err := jhanda.Parse(&cp.Options, args); err != nil {
		return fmt.Errorf("could not parse configure-product flags: %s", err)
	}

	cp.logger.Printf("configuring product...")

	if cp.Options.ConfigFile != "" {
		if cp.Options.ProductProperties != "" || cp.Options.NetworkProperties != "" || cp.Options.ProductResources != "" {
			return fmt.Errorf("config flag can not be passed with the product-properties, product-network or product-resources flag")
		}
	} else {
		if cp.Options.ProductProperties == "" && cp.Options.NetworkProperties == "" && cp.Options.ProductResources == "" {
			cp.logger.Printf("Provided properties are empty, nothing to do here")
			return nil
		}
	}

	stagedProducts, err := cp.service.ListStagedProducts()
	if err != nil {
		return err
	}

	var productGUID string
	for _, sp := range stagedProducts.Products {
		if sp.Type == cp.Options.ProductName {
			productGUID = sp.GUID
			break
		}
	}

	if productGUID == "" {
		return fmt.Errorf(`could not find product "%s"`, cp.Options.ProductName)
	}

	var (
		networkProperties string
		productProperties string
		productResources  string
	)

	if cp.Options.ConfigFile != "" {
		var config map[string]interface{}
		configContents, err := interpolate(cp.Options.ConfigFile, cp.Options.VarsFile, cp.Options.OpsFile)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(configContents, &config)
		if err != nil {
			return fmt.Errorf("%s could not be parsed as valid configuration: %s", cp.Options.ConfigFile, err)
		}

		if config["network-properties"] != nil {
			networkProperties, err = getJSONProperties(config["network-properties"])
			if err != nil {
				return err
			}
		}

		if config["product-properties"] != nil {
			productProperties, err = getJSONProperties(config["product-properties"])
			if err != nil {
				return err
			}
		}

		if config["resource-config"] != nil {
			productResources, err = getJSONProperties(config["resource-config"])
			if err != nil {
				return err
			}
		}
	} else {
		if cp.Options.NetworkProperties != "" {
			networkProperties = cp.Options.NetworkProperties
		}

		if cp.Options.ProductProperties != "" {
			productProperties = cp.Options.ProductProperties
		}

		if cp.Options.ProductResources != "" {
			productResources = cp.Options.ProductResources
		}
	}

	if networkProperties != "" {
		err = cp.configureNetwork(networkProperties, productGUID)
		if err != nil {
			return err
		}
	}

	if productProperties != "" {
		err = cp.configureProperties(productProperties, productGUID)
		if err != nil {
			return err
		}
	}

	if productResources != "" {
		err = cp.configureResources(productResources, productGUID)
		if err != nil {
			return err
		}
	}

	cp.logger.Printf("finished configuring product")

	return nil
}

func (cp ConfigureProduct) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command configures a staged product",
		ShortDescription: "configures a staged product",
		Flags:            cp.Options,
	}
}

func getJSONProperties(properties interface{}) (string, error) {
	yamlProperties, err := yaml.Marshal(properties)
	if err != nil {
		return "", err
	}

	jsonProperties, err := yamlConverter.YAMLToJSON(yamlProperties)
	if err != nil {
		return "", err
	}

	return string(jsonProperties), nil
}

func (cp ConfigureProduct) configureResources(productResources string, productGUID string) error {
	var userProvidedConfig map[string]json.RawMessage
	err := json.Unmarshal([]byte(productResources), &userProvidedConfig)
	if err != nil {
		return fmt.Errorf("could not decode product-resource json: %s", err)
	}

	jobs, err := cp.service.ListStagedProductJobs(productGUID)
	if err != nil {
		return fmt.Errorf("failed to fetch jobs: %s", err)
	}

	var names []string
	for name, _ := range userProvidedConfig {
		names = append(names, name)
	}

	sort.Strings(names)

	cp.logger.Printf("applying resource configuration for the following jobs:")
	for _, name := range names {
		cp.logger.Printf("\t%s", name)
		jobProperties, err := cp.service.GetStagedProductJobResourceConfig(productGUID, jobs[name])
		if err != nil {
			return fmt.Errorf("could not fetch existing job configuration: %s", err)
		}

		err = json.Unmarshal(userProvidedConfig[name], &jobProperties)
		if err != nil {
			return err
		}

		err = cp.service.UpdateStagedProductJobResourceConfig(productGUID, jobs[name], jobProperties)
		if err != nil {
			return fmt.Errorf("failed to configure resources: %s", err)
		}
	}
	return nil
}

func (cp ConfigureProduct) configureProperties(productProperties string, productGUID string) error {
	cp.logger.Printf("setting properties")
	err := cp.service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
		GUID:       productGUID,
		Properties: productProperties,
	})
	if err != nil {
		return fmt.Errorf("failed to configure product: %s", err)
	}
	cp.logger.Printf("finished setting properties")
	return nil
}

func (cp ConfigureProduct) configureNetwork(networkProperties string, productGUID string) error {
	cp.logger.Printf("setting up network")
	err := cp.service.UpdateStagedProductNetworksAndAZs(api.UpdateStagedProductNetworksAndAZsInput{
		GUID:           productGUID,
		NetworksAndAZs: networkProperties,
	})
	if err != nil {
		return fmt.Errorf("failed to configure product: %s", err)
	}
	cp.logger.Printf("finished setting up network")
	return nil
}
