package commands

import (
	"fmt"

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
	Name         string      `yaml:"name"`
	Optional     bool        `yaml:"optional"`
	Configurable bool        `yaml:"configurable"`
	Type         string      `yaml:"type"`
	Default      interface{} `yaml:"default"`
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

func (ct ConfigTemplate) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command generates a configuration template that can be passed in to om configure-product",
		ShortDescription: "generates a config template for the product",
		Flags:            ct.Options,
	}
}

func (ct ConfigTemplate) Execute(args []string) error {
	if _, err := jhanda.Parse(&ct.Options, args); err != nil {
		return fmt.Errorf("could not parse config-template flags: %s", err)
	}

	metadata, err := ct.metadataExtractor.ExtractMetadata(ct.Options.Product)
	if err != nil {
		return fmt.Errorf("could not extract metadata: %s", err)
	}

	var template proofing.ProductTemplate
	err = yaml.Unmarshal(metadata.Raw, &template)
	if err != nil {
		return fmt.Errorf("could not parse metadata: %s", err)
	}

	productProperties := map[string]interface{}{}
	for _, pb := range template.AllPropertyBlueprints() {
		if pb.Configurable {
			productProperties[pb.Property] = map[string]interface{}{
				"value": pb.Default,
			}
		}
	}

	configTemplate := map[string]interface{}{
		"product-properties": productProperties,
	}

	output, err := yaml.Marshal(configTemplate)
	if err != nil {
		return fmt.Errorf("could not marshal config template: %s", err) // NOTE: this cannot happen
	}

	ct.logger.Println(string(output))

	return nil
}
