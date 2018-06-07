package commands

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/kiln/proofing"
	"gopkg.in/yaml.v2"
)

type ConfigTemplate struct {
	metadataExtractor metadataExtractor
	logger            logger
	Options           struct {
		Product            string `long:"product"  short:"p"  required:"true" description:"path to product to generate config template for"`
		IncludePlaceholder bool   `short:"r" long:"include-placeholder" description:"replace obscured credentials to interpolatable placeholder"`
	}
}

type propertyBlueprint struct {
	Name         string              `yaml:"name"`
	Optional     bool                `yaml:"optional"`
	Configurable bool                `yaml:"configurable"`
	Type         string              `yaml:"type"`
	Default      interface{}         `yaml:"default"`
	Properties   []propertyBlueprint `yaml:"property_blueprints"`
}

type instanceGroup struct {
	Name       string              `yaml:"name"`
	Properties []propertyBlueprint `yaml:"property_blueprints"`
}

type metadata struct {
	Properties     []propertyBlueprint `yaml:"property_blueprints"`
	InstanceGroups []instanceGroup     `yaml:"job_types"`
}

func NewConfigTemplate(metadataExtractor metadataExtractor, logger logger) ConfigTemplate {
	return ConfigTemplate{
		metadataExtractor: metadataExtractor,
		logger:            logger,
	}
}

func (ct ConfigTemplate) Execute(args []string) error {
	if _, err := jhanda.Parse(&ct.Options, args); err != nil {
		return fmt.Errorf("could not parse config-template flags: %s", err)
	}

	extractedMetadata, err := ct.metadataExtractor.ExtractMetadata(ct.Options.Product)
	if err != nil {
		return fmt.Errorf("could not extract metadata: %s", err)
	}

	var template proofing.ProductTemplate
	err = yaml.Unmarshal(extractedMetadata.Raw, &template)
	if err != nil {
		return fmt.Errorf("could not parse metadata: %s", err)
	}

	configTemplateProperties := map[string]interface{}{}
	for _, pb := range template.AllPropertyBlueprints() {
		if !pb.Configurable {
			continue
		}

		if ct.Options.IncludePlaceholder {
			addSecretPlaceholder(pb.Default, pb.Type, configTemplateProperties, pb.Property)
			continue
		}

		switch pb.Type {
		case "simple_credentials":
			configTemplateProperties[pb.Property] = map[string]map[string]string{
				"value": map[string]string{
					"identity": "",
					"password": "",
				},
			}
		default:
			configTemplateProperties[pb.Property] = map[string]interface{}{
				"value": pb.Default,
			}
		}

	}

	configTemplate := map[string]interface{}{
		"product-properties": configTemplateProperties,
	}

	output, err := yaml.Marshal(configTemplate)
	if err != nil {
		return fmt.Errorf("could not marshal config template: %s", err) // NOTE: this cannot happen
	}

	ct.logger.Println(ct.concatenateRequiredProperties(output, template))

	return nil
}

func (ConfigTemplate) concatenateRequiredProperties(output []byte, template proofing.ProductTemplate) string {
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		for _, pb := range template.AllPropertyBlueprints() {
			propertyName := strings.TrimSpace(strings.Split(line, ":")[0])
			if pb.Property == propertyName {
				if pb.Required {
					lines[i+1] = lines[i+1] + " # required"
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

func (ct ConfigTemplate) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "**EXPERIMENTAL** This command generates a configuration template that can be passed in to om configure-product",
		ShortDescription: "**EXPERIMENTAL** generates a config template for the product",
		Flags:            ct.Options,
	}
}

func addSecretPlaceholder(value interface{}, t string, configurableProperties map[string]interface{}, name string) {
	switch t {
	case "secret":
		configurableProperties[name] = map[string]interface{}{
			"value": map[string]string{
				"secret": fmt.Sprintf("((%s.secret))", name),
			},
		}
	case "simple_credentials":
		configurableProperties[name] = map[string]interface{}{
			"value": map[string]string{
				"identity": fmt.Sprintf("((%s.identity))", name),
				"password": fmt.Sprintf("((%s.password))", name),
			},
		}
	case "rsa_cert_credentials":
		configurableProperties[name] = map[string]interface{}{
			"value": map[string]string{
				"cert_pem":        fmt.Sprintf("((%s.cert_pem))", name),
				"private_key_pem": fmt.Sprintf("((%s.private_key_pem))", name),
			},
		}
	case "rsa_pkey_credentials":
		configurableProperties[name] = map[string]interface{}{
			"value": map[string]string{
				"private_key_pem": fmt.Sprintf("((%s.private_key_pem))", name),
			},
		}
	case "salted_credentials":
		configurableProperties[name] = map[string]interface{}{
			"value": map[string]string{
				"identity": fmt.Sprintf("((%s.identity))", name),
				"password": fmt.Sprintf("((%s.password))", name),
				"salt":     fmt.Sprintf("((%s.salt))", name),
			},
		}
	default:
		configurableProperties[name] = map[string]interface{}{"value": value}
	}
}
