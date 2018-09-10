package commands

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/config"

	yamlConverter "github.com/ghodss/yaml"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type ConfigureProduct struct {
	environFunc func() []string
	service     configureProductService
	logger      logger
	Options     struct {
		ProductName       string   `long:"product-name"       short:"n"  required:"true" description:"name of the product being configured"`
		ConfigFile        string   `long:"config"             short:"c"                  description:"path to yml file containing all config fields (see docs/configure-product/README.md for format)"`
		VarsFile          []string `long:"vars-file"          short:"l"                  description:"Load variables from a YAML file"`
		VarsEnv           []string `long:"vars-env"                                      description:"Load variables from environment variables (e.g.: 'MY' to load MY_var=value)"`
		OpsFile           []string `long:"ops-file"           short:"o"                  description:"YAML operations file"`
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
	UpdateStagedProductErrands(productID, errandName string, postDeployState, preDeleteState interface{}) error
}

func NewConfigureProduct(environFunc func() []string, service configureProductService, logger logger) ConfigureProduct {
	return ConfigureProduct{
		environFunc: environFunc,
		service:     service,
		logger:      logger,
	}
}

func (cp ConfigureProduct) Execute(args []string) error {
	args, err := cp.loadProductName(args)
	if err != nil {
		return fmt.Errorf("could not parse configure-product flags: %s", err)
	}

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
		errandConfigs     map[string]config.ErrandConfig
	)

	if cp.Options.ConfigFile != "" {
		var cfg config.ProductConfiguration
		configContents, err := interpolate(interpolateOptions{
			templateFile: cp.Options.ConfigFile,
			varsFiles:    cp.Options.VarsFile,
			environFunc:  cp.environFunc,
			varsEnvs:     cp.Options.VarsEnv,
			opsFiles:     cp.Options.OpsFile,
		})
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(configContents, &cfg)
		if err != nil {
			return fmt.Errorf("%s could not be parsed as valid configuration: %s", cp.Options.ConfigFile, err)
		}

		if cfg.NetworkProperties != nil {
			networkProperties, err = getJSONProperties(cfg.NetworkProperties)
			if err != nil {
				return err
			}
		}

		if cfg.ProductProperties != nil {
			productProperties, err = getJSONProperties(cfg.ProductProperties)
			if err != nil {
				return err
			}
		}

		if cfg.ResourceConfigProperties != nil {
			productResources, err = getJSONProperties(cfg.ResourceConfigProperties)
			if err != nil {
				return err
			}
		}

		if cfg.ErrandConfigs != nil {
			errandConfigs = cfg.ErrandConfigs
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

	if len(errandConfigs) > 0 {
		err = cp.configureErrands(errandConfigs, productGUID)
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

func (cp ConfigureProduct) configureErrands(errandConfigs map[string]config.ErrandConfig, productGUID string) error {
	var names []string
	for name := range errandConfigs {
		names = append(names, name)
	}

	sort.Strings(names)

	cp.logger.Printf("applying errand configuration for the following errands:")
	for _, name := range names {
		cp.logger.Printf("\t%s", name)

		errandConfig := errandConfigs[name]
		err := cp.service.UpdateStagedProductErrands(productGUID, name, errandConfig.PostDeployState, errandConfig.PreDeleteState)
		if err != nil {
			return fmt.Errorf("failed to set errand state for errand %s: %s", name, err)
		}
	}
	return nil
}

func (cp ConfigureProduct) loadProductName(args []string) ([]string, error) {
	type productName struct {
		Name string `yaml:"product-name"`
	}

	jhanda.Parse(&cp.Options, args)

	flagSet := cp.checkIfProductNameFlagSet()
	if flagSet {
		return args, nil
	}

	configSet := cp.checkConfigSet()
	if !configSet {
		return args, nil
	}

	configContent, err := ioutil.ReadFile(cp.Options.ConfigFile)
	if err != nil {
		return args, err
	}

	name := productName{}
	err = yaml.Unmarshal(configContent, &name)
	if err != nil {
		return args, fmt.Errorf("%s could not be parsed as valid configuration: %s", cp.Options.ConfigFile, err)
	}
	if name.Name != "" {
		return append(args, "--product-name", name.Name), nil

	}
	return args, nil
}

func (cp ConfigureProduct) checkIfProductNameFlagSet() bool {
	return cp.Options.ProductName != ""
}

func (cp ConfigureProduct) checkConfigSet() bool {
	return cp.Options.ConfigFile != ""
}
