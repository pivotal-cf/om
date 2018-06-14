package commands

import (
	"fmt"

	"strconv"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"gopkg.in/yaml.v2"
)

type StagedConfig struct {
	logger  logger
	service stagedConfigService
	Options struct {
		Product            string `long:"product-name" short:"p" required:"true" description:"name of product"`
		IncludeCredentials bool   `short:"c" long:"include-credentials" description:"include credentials. note: requires product to have been deployed"`
		IncludePlaceholder bool   `short:"r" long:"include-placeholder" description:"replace obscured credentials to interpolatable placeholder"`
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
	selectorProperties := map[string]string{}

	for name, property := range properties {
		if property.Value == nil {
			continue
		}
		if property.Type == "selector" {
			selectorProperties[name] = property.Value.(string)
		}
		output, err := ec.parseProperties(productGUID, propertyName{
			prefix: name,
		}, property)
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

type propertyName struct {
	prefix         string
	index          int
	collectionName string
}

func (n *propertyName) credentialName() string {
	if n.collectionName != "" {
		return n.prefix + "[" + strconv.Itoa(n.index) + "]." + n.collectionName
	}

	return n.prefix
}

func (n *propertyName) placeholderName() string {
	name := n.prefix
	if n.collectionName != "" {
		name = n.prefix + "_" + strconv.Itoa(n.index) + "_" + n.collectionName
	}

	return strings.Replace(
		strings.TrimLeft(
			name,
			".",
		),
		".",
		"_",
		-1,
	)
}

func (ec StagedConfig) parseProperties(productGUID string, name propertyName, property api.ResponseProperty) (map[string]interface{}, error) {
	if !property.Configurable {
		return nil, nil
	}
	if property.IsCredential {
		return ec.handleCredential(productGUID, name, property)
	}
	if property.Type == "collection" {
		return ec.handleCollection(productGUID, name, property)
	}
	return map[string]interface{}{"value": property.Value}, nil
}

func (ec StagedConfig) handleCollection(productGUID string, name propertyName, property api.ResponseProperty) (map[string]interface{}, error) {
	var valueItems []map[string]interface{}

	for index, item := range property.Value.([]interface{}) {
		innerProperties := make(map[string]interface{})
		for innerKey, innerVal := range item.(map[interface{}]interface{}) {
			typeAssertedInnerValue := innerVal.(map[interface{}]interface{})

			innerValueProperty := api.ResponseProperty{
				Value:        typeAssertedInnerValue["value"],
				Configurable: typeAssertedInnerValue["configurable"].(bool),
				IsCredential: typeAssertedInnerValue["credential"].(bool),
				Type:         typeAssertedInnerValue["type"].(string),
			}
			returnValue, err := ec.parseProperties(productGUID, propertyName{
				prefix:         name.prefix,
				index:          index,
				collectionName: innerKey.(string),
			}, innerValueProperty)
			if err != nil {
				return nil, err
			}
			if returnValue != nil && len(returnValue) > 0 {
				innerProperties[innerKey.(string)] = returnValue["value"]
			}
		}
		if len(innerProperties) > 0 {
			valueItems = append(valueItems, innerProperties)
		}
	}
	if len(valueItems) > 0 {
		return map[string]interface{}{"value": valueItems}, nil
	}
	return nil, nil
}

func (ec StagedConfig) handleCredential(productGUID string, name propertyName, property api.ResponseProperty) (map[string]interface{}, error) {
	var output map[string]interface{}

	if ec.Options.IncludeCredentials {
		apiOutput, err := ec.service.GetDeployedProductCredential(api.GetDeployedProductCredentialInput{
			DeployedGUID:        productGUID,
			CredentialReference: name.credentialName(),
		})
		if err != nil {
			return nil, err
		}
		output = map[string]interface{}{"value": apiOutput.Credential.Value}
		return output, nil
	}
	if ec.Options.IncludePlaceholder {
		switch property.Type {
		case "secret":
			output = map[string]interface{}{
				"value": map[string]string{
					"secret": fmt.Sprintf("((%s.secret))", name.placeholderName()),
				},
			}
		case "simple_credentials":
			output = map[string]interface{}{
				"value": map[string]string{
					"identity": fmt.Sprintf("((%s.identity))", name.placeholderName()),
					"password": fmt.Sprintf("((%s.password))", name.placeholderName()),
				},
			}
		case "rsa_cert_credentials":
			output = map[string]interface{}{
				"value": map[string]string{
					"cert_pem":        fmt.Sprintf("((%s.cert_pem))", name.placeholderName()),
					"private_key_pem": fmt.Sprintf("((%s.private_key_pem))", name.placeholderName()),
				},
			}
		case "rsa_pkey_credentials":
			output = map[string]interface{}{
				"value": map[string]string{
					"private_key_pem": fmt.Sprintf("((%s.private_key_pem))", name.placeholderName()),
				},
			}
		case "salted_credentials":
			output = map[string]interface{}{
				"value": map[string]string{
					"identity": fmt.Sprintf("((%s.identity))", name.placeholderName()),
					"password": fmt.Sprintf("((%s.password))", name.placeholderName()),
					"salt":     fmt.Sprintf("((%s.salt))", name.placeholderName()),
				},
			}
		}
		return output, nil
	}

	return nil, nil
}
