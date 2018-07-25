package commands

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/kiln/proofing"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/config"
	"github.com/pivotal-cf/om/configparser"
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

	propertiesPair := makeProp(&template)

	properties := map[string]api.ResponseProperty{}
	for _, propTuple := range propertiesPair {
		name, prop := transform(propTuple)
		properties[name] = prop
	}

	configTemplateProperties := map[string]interface{}{}

	parser := configparser.NewConfigParser()

	for name, prop := range properties {
		if !prop.Configurable {
			continue
		}

		propertyName := configparser.NewPropertyName(name)
		output, err := parser.ParseProperties(propertyName, prop, ct.chooseCredentialHandler())
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

type propertyBlueprintTuple struct {
	proofing.NormalizedPropertyBlueprint
	proofing.PropertyBlueprint
}

func makeProp(template *proofing.ProductTemplate) []propertyBlueprintTuple {
	var output []propertyBlueprintTuple

	for _, pb := range template.PropertyBlueprints {
		normalizedPBs := pb.Normalize(".properties")
		for _, normalizedPB := range normalizedPBs {
			output = append(output, propertyBlueprintTuple{
				NormalizedPropertyBlueprint: normalizedPB,
				PropertyBlueprint:           pb,
			})
		}
	}

	for _, jobType := range template.JobTypes {
		for _, pb := range jobType.PropertyBlueprints {
			prefix := fmt.Sprintf(".%s", jobType.Name)
			normalizedPBs := pb.Normalize(prefix)
			for _, normalizedPB := range normalizedPBs {
				output = append(output, propertyBlueprintTuple{
					NormalizedPropertyBlueprint: normalizedPB,
					PropertyBlueprint:           pb,
				})
			}
		}
	}

	return output
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

func transformCollection(pbt propertyBlueprintTuple) (string, api.ResponseProperty) {
	collectionPB := pbt.PropertyBlueprint.(proofing.CollectionPropertyBlueprint)

	name := pbt.Property
	prop := api.ResponseProperty{
		Configurable: pbt.Configurable,
		Type:         pbt.Type,
		IsCredential: isCredential(pbt.Type),
	}

	propertyBlueprintList := collectionPB.PropertyBlueprints

	values := make([]interface{}, 0)

	if pbt.Default == nil {
		value := map[interface{}]interface{}{}
		for _, pb := range propertyBlueprintList {
			value[pb.Name] = map[interface{}]interface{}{
				"value":        nil,
				"configurable": true,
				"type":         pb.Type,
				"credential":   isCredential(pb.Type),
			}
		}
		values = append(values, value)
	} else {
		for _, defaultValue := range pbt.Default.([]interface{}) {
			value := map[interface{}]interface{}{}
			innerMap := defaultValue.(map[interface{}]interface{})
			for _, pb := range propertyBlueprintList {
				if innerValue, ok := innerMap[pb.Name]; ok {
					value[pb.Name] = map[interface{}]interface{}{
						"value":        innerValue,
						"configurable": true,
						"type":         pb.Type,
						"credential":   isCredential(pb.Type),
					}
				} else {
					value[pb.Name] = map[interface{}]interface{}{
						"value":        nil,
						"configurable": true,
						"type":         pb.Type,
						"credential":   isCredential(pb.Type),
					}
				}
			}
			values = append(values, value)
		}
	}

	prop.Value = values

	return name, prop
}

func transform(pbt propertyBlueprintTuple) (string, api.ResponseProperty) {
	if pbt.Type == "collection" {
		return transformCollection(pbt)
	}

	name := pbt.Property
	prop := api.ResponseProperty{
		Value:        pbt.Default,
		Configurable: pbt.Configurable,
		Type:         pbt.Type,
		IsCredential: isCredential(pbt.Type),
	}

	return name, prop
}
