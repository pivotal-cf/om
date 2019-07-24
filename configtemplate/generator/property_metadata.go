package generator

import (
	"fmt"
	"reflect"
	"strings"
)

type PropertyBlueprint struct {
	Configurable       string              `yaml:"configurable"`
	Default            interface{}         `yaml:"default"`
	Optional           bool                `yaml:"optional"`
	Name               string              `yaml:"name"`
	Type               string              `yaml:"type"`
	Options            []Option            `yaml:"options"`
	OptionTemplates    []OptionTemplate    `yaml:"option_templates"`
	PropertyBlueprints []PropertyBlueprint `yaml:"property_blueprints"`
}

type OptionTemplate struct {
	Name               string              `yaml:"name"`
	SelectValue        string              `yaml:"select_value"`
	PropertyBlueprints []PropertyBlueprint `yaml:"property_blueprints"`
}

func (p *PropertyBlueprint) IsConfigurable() bool {
	return strings.EqualFold(p.Configurable, "true")
}

func (p *PropertyBlueprint) DefaultSelectorPath(property string) string {
	return fmt.Sprintf("%s.%s", property, p.DefaultSelector())
}

func (p *PropertyBlueprint) DefaultSelector() string {
	defaultAsString := fmt.Sprintf("%v", p.Default)
	for _, optiontemplate := range p.OptionTemplates {
		if strings.EqualFold(optiontemplate.SelectValue, defaultAsString) {
			return optiontemplate.Name
		}
	}
	return defaultAsString
}

func (p *PropertyBlueprint) HasDefault() bool {
	return p.Default != nil
}

func (p *PropertyBlueprint) IsRequired() bool {
	return !p.Optional
}

func (p *PropertyBlueprint) OptionTemplate(selectorReference string) *OptionTemplate {
	for _, option := range p.OptionTemplates {
		if strings.EqualFold(option.Name, selectorReference) {
			return &option
		}
	}
	return nil
}

func (p *PropertyBlueprint) PropertyType(propertyName string) PropertyValue {
	propertyName = strings.Replace(propertyName, "properties.", "", 1)
	propertyName = strings.Replace(propertyName, ".", "_", -1)
	if p.IsSelector() {
		if p.Default != nil {
			return &SelectorValue{
				Value: fmt.Sprintf("%v", p.Default),
			}
		} else {
			return nil
		}
	}
	if p.IsMultiSelect() {
		if len(p.Options) == 1 {
			return &MultiSelectorValue{
				Value: []string{fmt.Sprintf("%v", p.Options[0].Name)},
			}
		}

		if p.Default == nil {
			return nil
		}
		rt := reflect.TypeOf(p.Default)
		switch rt.Kind() {
		case reflect.Slice:
			values := []string{}
			for _, option := range p.Default.([]interface{}) {
				values = append(values, fmt.Sprintf("%v", option))
			}
			return &MultiSelectorValue{
				Value: values,
			}

		default:
			return nil
		}

	}
	if p.IsCertificate() {
		return &CertificateValueHolder{
			Value: NewCertificateValue(propertyName),
		}
	}
	if p.IsSecret() {
		return &SecretValueHolder{
			Value: &SecretValue{
				Value: fmt.Sprintf("((%s))", propertyName),
			},
		}
	}

	if p.IsSimpleCredentials() {
		return &SimpleCredentialValueHolder{
			Value: &SimpleCredentialValue{
				Password: fmt.Sprintf("((%s_password))", propertyName),
				Identity: fmt.Sprintf("((%s_identity))", propertyName),
			},
		}
	}
	return &SimpleValue{
		Value: fmt.Sprintf("((%s))", propertyName),
	}
}

func (p *PropertyBlueprint) IsString() bool {
	if p.Type == "dropdown_select" {
		_, ok := p.Options[0].Name.(string)
		return ok
	} else {
		return p.Type == "string" || p.Type == "text" ||
			p.Type == "ip_ranges" || p.Type == "string_list" ||
			p.Type == "network_address" || p.Type == "wildcard_domain" ||
			p.Type == "email" || p.Type == "ca_certificate" || p.Type == "http_url" ||
			p.Type == "ldap_url" || p.Type == "service_network_az_single_select" || p.Type == "vm_type_dropdown" || p.Type == "disk_type_dropdown"
	}
}
func (p *PropertyBlueprint) IsInt() bool {
	if p.Type == "dropdown_select" {
		_, ok := p.Options[0].Name.(int)
		return ok
	} else {
		return p.Type == "port" || p.Type == "integer"
	}
}

func (p *PropertyBlueprint) IsBool() bool {
	return p.Type == "boolean"
}

func (p *PropertyBlueprint) IsSecret() bool {
	return p.Type == "secret"
}
func (p *PropertyBlueprint) IsSimpleCredentials() bool {
	return p.Type == "simple_credentials"
}

func (p *PropertyBlueprint) IsCollection() bool {
	return p.Type == "collection"
}

func (p *PropertyBlueprint) IsRequiredCollection() bool {
	return p.IsRequired()
}

func (p *PropertyBlueprint) IsSelector() bool {
	return p.Type == "selector"
}

func (p *PropertyBlueprint) IsMultiSelect() bool {
	return p.Type == "multi_select_options"
}

func (p *PropertyBlueprint) IsCertificate() bool {
	return p.Type == "rsa_cert_credentials"
}

func (p *PropertyBlueprint) IsDropdown() bool {
	return p.Type == "vm_type_dropdown" || p.Type == "disk_type_dropdown"
}

func (p *PropertyBlueprint) IsAZList() bool {
	return p.Type == "service_network_az_multi_select"
}

func (p *PropertyBlueprint) DataType() string {
	if p.IsString() {
		return "string"
	} else if p.IsInt() {
		return "int"
	} else if p.IsBool() {
		return "bool"
	} else {
		panic("Type " + p.Type + " not recongnized")
	}
}
