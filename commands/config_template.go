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
		Product             string `long:"product"  short:"p"  required:"true" description:"path to product to generate config template for"`
		IncludePlaceholders bool   `long:"include-placeholders" short:"r" description:"replace obscured credentials with interpolatable placeholders"`
	}
}

type propertyBluePrintPair struct {
	proofing.NormalizedPropertyBlueprint
	proofing.PropertyBlueprint
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

	propertyPairs := makePropertyBluePrintPair(&template)
	nameApiResponseMaps := transformAll(propertyPairs)

	configTemplateProperties := map[string]interface{}{}

	parser := configparser.NewConfigParser()

	for name, prop := range nameApiResponseMaps {
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
		return fmt.Errorf("could not marshal config template: %s", err)
	}

	// post-processing
	ct.logger.Println(concatenateRequiredProperties(output, propertyPairs))

	return nil
}

func concatenateRequiredProperties(output []byte, propertyPairs []propertyBluePrintPair) string {
	// convert list to map to avoid n^2 double for loop
	namePropertyMaps := map[string]proofing.NormalizedPropertyBlueprint{}
	for _, pbp := range propertyPairs {
		namePropertyMaps[pbp.Property] = pbp.NormalizedPropertyBlueprint
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		propertyName := strings.TrimSpace(strings.Split(line, ":")[0])
		if v, ok := namePropertyMaps[propertyName]; ok && v.Required {
			lines[i+1] = lines[i+1] + " # required"
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
	if ct.Options.IncludePlaceholders {
		return configparser.PlaceholderHandler()
	}

	return configparser.KeyOnlyHandler()
}

func makePropertyBluePrintPair(template *proofing.ProductTemplate) []propertyBluePrintPair {
	var output []propertyBluePrintPair

	for _, pb := range template.PropertyBlueprints {
		normalizedPBs := pb.Normalize(".properties")
		for _, normalizedPB := range normalizedPBs {
			output = append(output, propertyBluePrintPair{
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
				output = append(output, propertyBluePrintPair{
					NormalizedPropertyBlueprint: normalizedPB,
					PropertyBlueprint:           pb,
				})
			}
		}
	}

	return output
}

func isCredential(t string) bool {
	switch t {
	case "secret", "simple_credentials", "rsa_cert_credentials", "rsa_pkey_credentials", "salted_credentials":
		return true
	}
	return false
}

func transformCollection(pbp propertyBluePrintPair) (string, api.ResponseProperty) {
	collectionPB := pbp.PropertyBlueprint.(proofing.CollectionPropertyBlueprint)

	name := pbp.Property
	prop := api.ResponseProperty{
		Configurable: pbp.Configurable,
		Type:         pbp.Type,
		IsCredential: isCredential(pbp.Type),
	}

	propertyBlueprintList := collectionPB.PropertyBlueprints

	values := make([]interface{}, 0)

	if pbp.Default == nil {
		value := map[interface{}]interface{}{}
		for _, pb := range propertyBlueprintList {
			value[pb.Name] = map[interface{}]interface{}{
				"value":        pb.Default,
				"configurable": true,
				"type":         pb.Type,
				"credential":   isCredential(pb.Type),
			}
		}
		values = append(values, value)
	} else {
		for _, defaultValue := range pbp.Default.([]interface{}) {
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

func transform(pbp propertyBluePrintPair) (string, api.ResponseProperty) {
	if pbp.Type == "collection" {
		return transformCollection(pbp)
	}

	name := pbp.Property
	prop := api.ResponseProperty{
		Value:        pbp.Default,
		Configurable: pbp.Configurable,
		Type:         pbp.Type,
		IsCredential: isCredential(pbp.Type),
	}

	return name, prop
}

func transformAll(pbps []propertyBluePrintPair) map[string]api.ResponseProperty {
	properties := map[string]api.ResponseProperty{}
	for _, propTuple := range pbps {
		name, prop := transform(propTuple)
		properties[name] = prop
	}
	return properties
}
