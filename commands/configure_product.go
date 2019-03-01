package commands

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/config"

	yamlConverter "github.com/ghodss/yaml"
	"gopkg.in/yaml.v2"
)

type ConfigureProduct struct {
	environFunc func() []string
	service     configureProductService
	logger      logger
	target      string
	Options     struct {
		ConfigFile string   `long:"config"    short:"c" description:"path to yml file containing all config fields (see docs/configure-product/README.md for format)" required:"true"`
		VarsFile   []string `long:"vars-file" short:"l" description:"Load variables from a YAML file"`
		VarsEnv    []string `long:"vars-env"            description:"Load variables from environment variables (e.g.: 'MY' to load MY_var=value)"`
		OpsFile    []string `long:"ops-file"  short:"o" description:"YAML operations file"`
	}
}

//go:generate counterfeiter -o ./fakes/configure_product_service.go --fake-name ConfigureProductService . configureProductService
type configureProductService interface {
	GetStagedProductJobResourceConfig(productGUID, jobGUID string) (api.JobProperties, error)
	ListInstallations() ([]api.InstallationsServiceOutput, error)
	ListStagedPendingChanges() (api.PendingChangesOutput, error)
	ListStagedProductJobs(productGUID string) (map[string]string, error)
	ListStagedProducts() (api.StagedProductsOutput, error)
	UpdateStagedProductErrands(productID, errandName string, postDeployState, preDeleteState interface{}) error
	UpdateStagedProductJobResourceConfig(productGUID, jobGUID string, jobProperties api.JobProperties) error
	UpdateStagedProductNetworksAndAZs(api.UpdateStagedProductNetworksAndAZsInput) error
	UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput) error
}

type configureProduct struct {
	config.ProductConfiguration `yaml:",inline"`
	ValidateConfigComplete      bool                   `yaml:"validate-config-complete"`
	Field                       map[string]interface{} `yaml:",inline"`
}

func NewConfigureProduct(environFunc func() []string, service configureProductService, target string, logger logger) ConfigureProduct {
	return ConfigureProduct{
		environFunc: environFunc,
		service:     service,
		target:      target,
		logger:      logger,
	}
}

func (cp ConfigureProduct) Execute(args []string) error {
	if _, err := jhanda.Parse(&cp.Options, args); err != nil {
		return fmt.Errorf("could not parse configure-product flags: %s", err)
	}

	err := checkRunningInstallation(cp.service.ListInstallations)
	if err != nil {
		return err
	}

	cp.logger.Printf("configuring product...")

	cfg := configureProduct{ValidateConfigComplete: true}

	cfg, err = cp.interpolateConfig(cfg)
	if err != nil {
		return err
	}

	err = cp.validateConfig(cfg)
	if err != nil {
		return err
	}

	productGUID, err := cp.getProductGUID(cfg)
	if err != nil {
		return err
	}

	err = cp.configureNetwork(cfg, productGUID)
	if err != nil {
		return err
	}

	err = cp.configureProperties(cfg, productGUID)
	if err != nil {
		return err
	}

	err = cp.configureResources(cfg, productGUID)
	if err != nil {
		return err
	}

	err = cp.configureErrands(cfg, productGUID)
	if err != nil {
		return err
	}

	if cfg.ValidateConfigComplete {
		if err := cp.validateConfigComplete(productGUID); err != nil {
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

func (cp *ConfigureProduct) configureResources(cfg configureProduct, productGUID string) error {
	if cfg.ResourceConfigProperties == nil {
		cp.logger.Println("resource config properties are not provided, nothing to do here")
		return nil
	}

	productResources, err := getJSONProperties(cfg.ResourceConfigProperties)
	if err != nil {
		return err
	}

	var userProvidedConfig map[string]json.RawMessage
	err = json.Unmarshal([]byte(productResources), &userProvidedConfig)
	if err != nil {
		return fmt.Errorf("could not decode product-resource json: %s", err)
	}

	jobs, err := cp.service.ListStagedProductJobs(productGUID)
	if err != nil {
		return fmt.Errorf("failed to fetch jobs: %s", err)
	}

	var names []string
	for name := range userProvidedConfig {
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

func (cp *ConfigureProduct) configureProperties(cfg configureProduct, productGUID string) error {
	if cfg.ProductProperties == nil {
		cp.logger.Println("product properties are not provided, nothing to do here")
		return nil
	}

	productProperties := cfg.ProductProperties
	for name, value := range productProperties {
		switch value.(type) {
		case map[interface{}]interface{}:
			v := value.(map[interface{}]interface{})
			// This is here:
			// * the GET /properties returns the value as a field named `selected_options`.
			// * the PUT /properties expects the filed to be named `option_value`.
			// We are future-proofing and migrating until the issue is resolved.
			// See for more information [#163833845]
			if v["selected_option"] == nil && v["option_value"] != nil {
				v["selected_option"] = v["option_value"]
			} else if v["option_value"] == nil && v["selected_option"] != nil {
				v["option_value"] = v["selected_option"]
			}
			productProperties[name] = value
		}
	}

	productPropertiesJSON, err := getJSONProperties(cfg.ProductProperties)
	if err != nil {
		return err
	}

	cp.logger.Printf("setting properties")
	err = cp.service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
		GUID:       productGUID,
		Properties: productPropertiesJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to configure product: %s", err)
	}
	cp.logger.Printf("finished setting properties")

	return nil
}

func (cp *ConfigureProduct) configureNetwork(cfg configureProduct, productGUID string) error {
	if cfg.NetworkProperties == nil {
		cp.logger.Println("network properties are not provided, nothing to do here")
		return nil
	}

	networkProperties, err := getJSONProperties(cfg.NetworkProperties)
	if err != nil {
		return err
	}

	cp.logger.Printf("setting up network")
	err = cp.service.UpdateStagedProductNetworksAndAZs(api.UpdateStagedProductNetworksAndAZsInput{
		GUID:           productGUID,
		NetworksAndAZs: networkProperties,
	})

	if err != nil {
		return fmt.Errorf("failed to configure product: %s", err)
	}
	cp.logger.Printf("finished setting up network")

	return nil
}

func (cp *ConfigureProduct) configureErrands(cfg configureProduct, productGUID string) error {
	if cfg.ErrandConfigs == nil || len(cfg.ErrandConfigs) == 0 {
		cp.logger.Println("errands are not provided, nothing to do here")
		return nil
	}

	var names []string
	for name := range cfg.ErrandConfigs {
		names = append(names, name)
	}

	sort.Strings(names)

	cp.logger.Printf("applying errand configuration for the following errands:")
	for _, name := range names {
		cp.logger.Printf("\t%s", name)

		errandConfig := cfg.ErrandConfigs[name]
		err := cp.service.UpdateStagedProductErrands(productGUID, name, errandConfig.PostDeployState, errandConfig.PreDeleteState)
		if err != nil {
			return fmt.Errorf("failed to set errand state for errand %s: %s", name, err)
		}
	}

	return nil
}

func (cp *ConfigureProduct) interpolateConfig(cfg configureProduct) (configureProduct, error) {
	configContents, err := interpolate(interpolateOptions{
		templateFile: cp.Options.ConfigFile,
		varsFiles:    cp.Options.VarsFile,
		environFunc:  cp.environFunc,
		varsEnvs:     cp.Options.VarsEnv,
		opsFiles:     cp.Options.OpsFile,
	}, "")
	if err != nil {
		return configureProduct{}, err
	}

	err = yaml.UnmarshalStrict(configContents, &cfg)
	if err != nil {
		return configureProduct{}, fmt.Errorf("%s could not be parsed as valid configuration: %s", cp.Options.ConfigFile, err)
	}

	return cfg, nil
}

func (cp ConfigureProduct) validateConfig(cfg configureProduct) error {
	if cfg.ProductName == "" {
		return fmt.Errorf("could not parse configure-product config: \"product-name\" is required")
	}

	if len(cfg.Field) > 0 {
		var unrecognizedKeys []string
		for key := range cfg.Field {
			unrecognizedKeys = append(unrecognizedKeys, key)
		}
		sort.Strings(unrecognizedKeys)

		return fmt.Errorf("the config file contains unrecognized keys: %s", strings.Join(unrecognizedKeys, ", "))
	}
	return nil
}

func (cp ConfigureProduct) getProductGUID(cfg configureProduct) (string, error) {
	stagedProducts, err := cp.service.ListStagedProducts()
	if err != nil {
		return "", err
	}

	var productGUID string
	for _, sp := range stagedProducts.Products {
		if sp.Type == cfg.ProductName {
			productGUID = sp.GUID
			break
		}
	}

	if productGUID == "" {
		return "", fmt.Errorf(`could not find product "%s"`, cfg.ProductName)
	}

	return productGUID, nil
}

func (cp ConfigureProduct) validateConfigComplete(productGUID string) error {
	pendingChanges, err := cp.service.ListStagedPendingChanges()
	if err != nil {
		return err
	}
	for _, changeList := range pendingChanges.ChangeList {
		if changeList.GUID == productGUID {
			completenessCheck := changeList.CompletenessChecks
			if completenessCheck == nil {
				return fmt.Errorf("configuration completeness could not be determined.\nThis feature is only supported for OpsMan 2.2+\nIf you're on older version of OpsMan add the line `validate-config-complete: false` to your config file.")
			}
			if !completenessCheck.ConfigurationComplete {
				return fmt.Errorf("configuration not complete.\nThe properties you provided have been set,\nbut some required properties or configuration details are still missing.\nVisit the Ops Manager for details: %s", cp.target)
			}
		}
	}
	return nil
}
