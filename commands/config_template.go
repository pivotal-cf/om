package commands

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/kiln/proofing"
	yaml "gopkg.in/yaml.v2"
)

type ConfigTemplate struct {
	logger            logger
	metadataExtractor metadataExtractor
	Options           struct {
		Product string `long:"product"  short:"p"  required:"true" description:"path to product to generate config template for"`
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

func NewConfigTemplate(logger logger, metadataExtractor metadataExtractor) ConfigTemplate {
	return ConfigTemplate{
		logger:            logger,
		metadataExtractor: metadataExtractor,
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
