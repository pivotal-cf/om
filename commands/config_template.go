package commands

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/kiln/proofing"
	"github.com/pivotal-cf/om/config"
	"gopkg.in/yaml.v2"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/configparser"
	"log"
)

type ConfigTemplate struct {
	metadataExtractor metadataExtractor
	logger            logger
	Options struct {
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

	properties := map[string]api.ResponseProperty{}
	for _, pb := range template.AllPropertyBlueprints() {

		name, prop := transform(pb)
		properties[name] = prop
	}

	configTemplateProperties := map[string]interface{}{}

	parser := configparser.NewConfigParser()

	for name, prop := range properties {
		if !prop.Configurable {
			continue
		}

		log.Println(name)

		propertyName := configparser.NewPropertyName(name)
		output, err := parser.ParseProperties("", propertyName, prop, ct.chooseCredentialHandler())
		if err != nil {
			return err
		}

		if output != nil && len(output) > 0 {
			configTemplateProperties[name] = output
		}
	}

	configTemplate := config.ProductConfiguration{
		ProductProperties: configTemplateProperties,
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

func (ct ConfigTemplate) chooseCredentialHandler() configparser.CredentialHandler {
	if ct.Options.IncludePlaceholder {
		return configparser.PlaceholderHandler()
	}

	return configparser.KeyOnlyHandler()
}

func isCredential(t string) bool {
	switch t {
	case "secret":
		return true
	case "simple_credentials":
		return true
	case "rsa_cert_credentials":
		return true
	case "rsa_pkey_credentials":
		return true
	case "salted_credentials":
		return true
	}
	return false
}

func transformCollection(raw proofing.NormalizedPropertyBlueprint) (string, api.ResponseProperty) {
	name := raw.Property

	prop := api.ResponseProperty{
		Value:        make([]interface{}, 0),
		Configurable: raw.Configurable,
		Type:         raw.Type,
		IsCredential: isCredential(raw.Type),
	}

	//
	//
	//if raw.Default != nil {
	//	prop.Value = raw.Default
	//	return name, prop
	//}

	return name, prop
}

func transform(raw proofing.NormalizedPropertyBlueprint) (string, api.ResponseProperty) {
	if raw.Type == "collection" {
		return transformCollection(raw)
	}

	name := raw.Property
	prop := api.ResponseProperty{
		Value:        raw.Default,
		Configurable: raw.Configurable,
		Type:         raw.Type,
		IsCredential: isCredential(raw.Type),
	}

	return name, prop
}
