package configparser_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/configparser"
)

var _ = Describe("Config Parser", func() {
	var (
		properties map[string]api.ResponseProperty
	)

	getOutput := func(handler configparser.CredentialHandler) (string, error) {
		configurableProperties := map[string]interface{}{}
		for name, property := range properties {
			if property.Value == nil {
				continue
			}
			var output map[string]interface{}

			propertyName := configparser.NewPropertyName(name)
			parser := configparser.NewConfigParser()
			output, err := parser.ParseProperties(propertyName, property, handler)
			if err != nil {
				return "", err
			}

			if len(output) > 0 {
				configurableProperties[name] = output
			}
		}

		bytes, err := yaml.Marshal(configurableProperties)
		return string(bytes), err
	}

	BeforeEach(func() {
		properties = map[string]api.ResponseProperty{
			".properties.some-string-property": {
				Value:        "some-value",
				Configurable: true,
			},
			".properties.some-non-configurable-property": {
				Value:        "some-value",
				Configurable: false,
			},
			".properties.some-secret-property": {
				Type: "secret",
				Value: map[string]interface{}{
					"secret": "***",
				},
				IsCredential: true,
				Configurable: true,
			},
			".properties.simple-credentials": {
				Type: "simple_credentials",
				Value: map[string]interface{}{
					"identity": "***",
					"password": "***",
				},
				IsCredential: true,
				Configurable: true,
			},
			".properties.rsa-cert-credentials": {
				Type: "rsa_cert_credentials",
				Value: map[string]interface{}{
					"cert_pem":        "***",
					"private_key_pem": "***",
				},
				IsCredential: true,
				Configurable: true,
			},
			".properties.rsa-pkey-credentials": {
				Type: "rsa_pkey_credentials",
				Value: map[string]interface{}{
					"private_key_pem": "***",
				},
				IsCredential: true,
				Configurable: true,
			},
			".properties.salted-credentials": {
				Type: "salted_credentials",
				Value: map[string]interface{}{
					"identity": "***",
					"salt":     "***",
					"password": "***",
				},
				IsCredential: true,
				Configurable: true,
			},
			".properties.collection": {
				Type: "collection",
				Value: []interface{}{
					map[interface{}]interface{}{
						"certificate": map[interface{}]interface{}{
							"type":         "rsa_cert_credentials",
							"configurable": true,
							"credential":   true,
							"value": map[interface{}]interface{}{
								"cert_pem":        "***",
								"private_key_pem": "***",
							},
						},
						"name": map[interface{}]interface{}{
							"type":         "string",
							"configurable": true,
							"credential":   false,
							"value":        "Certificate",
						},
						"non-configurable": map[interface{}]interface{}{
							"type":         "string",
							"configurable": false,
							"credential":   false,
							"value":        "non-configurable",
						},
					},
					map[interface{}]interface{}{
						"certificate2": map[interface{}]interface{}{
							"type":         "rsa_cert_credentials",
							"configurable": true,
							"credential":   true,
							"value": map[interface{}]interface{}{
								"cert_pem":        "***",
								"private_key_pem": "***",
							},
						},
					},
				},

				IsCredential: false,
				Configurable: true,
			},
			".properties.some-non-configurable-secret-property": {
				Value: map[string]interface{}{
					"some-secret-type": "***",
				},
				IsCredential: true,
				Configurable: false,
			},
			".properties.some-null-property": {
				Value:        nil,
				Configurable: true,
			},
			".properties.some-selector": api.ResponseProperty{
				Value:        "internal",
				Type:         "selector",
				Configurable: true,
			},
			".properties.some-selector.not-internal.some-string-property": api.ResponseProperty{
				Value:        "some-value",
				Configurable: true,
			},
			".properties.some-selector-with-selected-value": api.ResponseProperty{
				Value:          "Hello World",
				SelectedOption: "beginner",
				Type:           "selector",
				Configurable:   true,
			},
		}
	})

	Context("given nil handler", func() {
		It("removes all the credential types from the payload", func() {
			output, err := getOutput(configparser.NewNilHandler())
			Expect(err).ToNot(HaveOccurred())

			Expect(output).To(MatchYAML(`---
.properties.collection:
  value:
  - name: Certificate
.properties.some-string-property:
  value: some-value
.properties.some-selector:
  value: internal
.properties.some-selector.not-internal.some-string-property:
  value: some-value
.properties.some-selector-with-selected-value:
  selected_option: beginner
  value: Hello World

`))
		})
	})

	Context("given placeholder handler", func() {
		It("replace all the credential types to placeholders", func() {
			output, err := getOutput(configparser.NewPlaceholderHandler())
			Expect(err).ToNot(HaveOccurred())

			Expect(output).To(MatchYAML(`---
".properties.some-string-property":
  value: some-value
".properties.some-secret-property":
  value:
    secret: "((properties_some-secret-property.secret))"
".properties.some-selector":
  value: internal
".properties.some-selector.not-internal.some-string-property":
  value: some-value
".properties.simple-credentials":
  value:
    identity: "((properties_simple-credentials.identity))"
    password: "((properties_simple-credentials.password))"
".properties.rsa-cert-credentials":
  value:
    cert_pem: "((properties_rsa-cert-credentials.cert_pem))"
    private_key_pem: "((properties_rsa-cert-credentials.private_key_pem))"
".properties.rsa-pkey-credentials":
  value:
    public_key_pem: "((properties_rsa-pkey-credentials.public_key_pem))"
    private_key_pem: "((properties_rsa-pkey-credentials.private_key_pem))"
".properties.salted-credentials":
  value:
    identity: "((properties_salted-credentials.identity))"
    password: "((properties_salted-credentials.password))"
    salt: "((properties_salted-credentials.salt))"
".properties.collection":
  value:
  - certificate:
      private_key_pem: "((properties_collection_0_certificate.private_key_pem))"
      cert_pem: "((properties_collection_0_certificate.cert_pem))"
    name: Certificate
  - certificate2:
      private_key_pem: "((properties_collection_1_certificate2.private_key_pem))"
      cert_pem: "((properties_collection_1_certificate2.cert_pem))"
.properties.some-selector-with-selected-value:
  selected_option: beginner
  value: Hello World
`))

		})
	})
})
