package configparser

import (
	"fmt"
	"github.com/pivotal-cf/om/api"
	"strconv"
	"strings"
)

type getCredential interface {
	GetDeployedProductCredential(input api.GetDeployedProductCredentialInput) (api.GetDeployedProductCredentialOutput, error)
}

type CredentialHandler func(name PropertyName, property api.ResponseProperty) (map[string]interface{}, error)

type PropertyName struct {
	prefix         string
	index          int
	collectionName string
}

func NewPropertyName(prefix string) PropertyName {
	return PropertyName{
		prefix: prefix,
	}
}

func (n *PropertyName) credentialName() string {
	if n.collectionName != "" {
		return n.prefix + "[" + strconv.Itoa(n.index) + "]." + n.collectionName
	}

	return n.prefix
}

func (n *PropertyName) placeholderName() string {
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

type configParser struct{}

func NewConfigParser() *configParser {
	return &configParser{}
}

func (p *configParser) ParseProperties(name PropertyName, property api.ResponseProperty, handler CredentialHandler) (map[string]interface{}, error) {
	if !property.Configurable {
		return nil, nil
	}
	if property.IsCredential {
		return handler(name, property)
	}
	if property.Type == "collection" {
		return p.handleCollection(name, property, handler)
	}
	if property.SelectedOption != "" {
		return map[string]interface{}{"value": property.Value, "selected_option": property.SelectedOption}, nil
	}
	return map[string]interface{}{"value": property.Value}, nil
}

func (p *configParser) handleCollection(name PropertyName, property api.ResponseProperty, handler CredentialHandler) (map[string]interface{}, error) {
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
			returnValue, err := p.ParseProperties(PropertyName{
				prefix:         name.prefix,
				index:          index,
				collectionName: innerKey.(string),
			}, innerValueProperty, handler)
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

func NilHandler() CredentialHandler {
	return func(name PropertyName, property api.ResponseProperty) (map[string]interface{}, error) {
		return nil, nil
	}
}

func KeyOnlyHandler() CredentialHandler {
	var output map[string]interface{}

	return func(name PropertyName, property api.ResponseProperty) (map[string]interface{}, error) {
		switch property.Type {
		case "secret":
			output = map[string]interface{}{
				"value": map[string]string{
					"secret": "",
				},
			}
		case "simple_credentials":
			output = map[string]interface{}{
				"value": map[string]string{
					"identity": "",
					"password": "",
				},
			}
		case "rsa_cert_credentials":
			output = map[string]interface{}{
				"value": map[string]string{
					"cert_pem":        "",
					"private_key_pem": "",
				},
			}
		case "rsa_pkey_credentials":
			output = map[string]interface{}{
				"value": map[string]string{
					"public_key_pem":  "",
					"private_key_pem": "",
				},
			}
		case "salted_credentials":
			output = map[string]interface{}{
				"value": map[string]string{
					"identity": "",
					"password": "",
					"salt":     "",
				},
			}
		}
		return output, nil
	}
}

func PlaceholderHandler() CredentialHandler {
	var output map[string]interface{}

	return func(name PropertyName, property api.ResponseProperty) (map[string]interface{}, error) {
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
					"public_key_pem":  fmt.Sprintf("((%s.public_key_pem))", name.placeholderName()),
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
}

func GetCredentialHandler(productGUID string, apiService getCredential) CredentialHandler {
	var output map[string]interface{}

	return func(name PropertyName, property api.ResponseProperty) (map[string]interface{}, error) {
		apiOutput, err := apiService.GetDeployedProductCredential(api.GetDeployedProductCredentialInput{
			DeployedGUID:        productGUID,
			CredentialReference: name.credentialName(),
		})
		if err != nil {
			return nil, err
		}
		output = map[string]interface{}{"value": apiOutput.Credential.Value}
		return output, nil
	}
}
