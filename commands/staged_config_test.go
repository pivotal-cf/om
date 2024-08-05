package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StagedConfig", func() {
	var (
		logger           *fakes.Logger
		fakeService      *fakes.StagedConfigService
		internalSelector api.ResponseProperty
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
	})

	Context("using configs without selected_options", func() {
		BeforeEach(func() {
			internalSelector = api.ResponseProperty{
				Value:        "internal",
				Type:         "selector",
				Configurable: true,
			}
			fakeService = setFakeService(internalSelector, true)
		})

		It("writes a config file to stdout", func() {
			command := commands.NewStagedConfig(fakeService, logger)
			err := executeCommand(command, []string{
				"--product-name", "some-product",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.GetStagedProductByNameCallCount()).To(Equal(1))
			Expect(fakeService.GetStagedProductByNameArgsForCall(0)).To(Equal("some-product"))

			Expect(fakeService.GetStagedProductPropertiesCallCount()).To(Equal(1))
			name, redact := fakeService.GetStagedProductPropertiesArgsForCall(0)
			Expect(name).To(Equal("some-product-guid"))
			Expect(redact).To(BeTrue())

			Expect(fakeService.GetStagedProductNetworksAndAZsCallCount()).To(Equal(1))
			Expect(fakeService.GetStagedProductNetworksAndAZsArgsForCall(0)).To(Equal("some-product-guid"))

			Expect(fakeService.GetStagedProductSyslogConfigurationCallCount()).To(Equal(1))
			Expect(fakeService.GetStagedProductSyslogConfigurationArgsForCall(0)).To(Equal("some-product-guid"))

			Expect(fakeService.ListStagedProductJobsCallCount()).To(Equal(1))
			Expect(fakeService.ListStagedProductJobsArgsForCall(0)).To(Equal("some-product-guid"))

			Expect(fakeService.GetStagedProductJobResourceConfigCallCount()).To(Equal(1))
			productGuid, jobsGuid := fakeService.GetStagedProductJobResourceConfigArgsForCall(0)
			Expect(productGuid).To(Equal("some-product-guid"))
			Expect(jobsGuid).To(Equal("some-job-guid"))

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			output := logger.PrintlnArgsForCall(0)
			Expect(output).To(ContainElement(MatchYAML(`---
product-name: some-product
product-properties:
  .properties.collection:
    value:
    - name: Certificate
  .properties.some-string-property:
    value: some-value
  .properties.some-selector:
    value: internal
network-properties:
  singleton_availability_zone:
    name: az-one
resource-config:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
    max_in_flight: 2
errand-config:
  first-errand:
    post-deploy-state: true
    pre-delete-state: do-something
  second-errand:
    post-deploy-state: false
syslog-properties:
  enabled: true
  host: tcp://1.1.1.1
`)))
		})

		When("--include-placeholders is used", func() {
			It("replaces *** with interpolatable placeholders and removes non-configurable properties", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := executeCommand(command, []string{
					"--product-name", "some-product",
					"--include-placeholders",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(logger.PrintlnCallCount()).To(Equal(1))
				output := logger.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`---
product-name: some-product
product-properties:
  ".properties.some-string-property":
    value: some-value
  ".properties.some-secret-property":
    value:
      secret: "((properties_some-secret-property.secret))"
  ".properties.some-selector":
    value: internal
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
network-properties:
  singleton_availability_zone:
    name: az-one
resource-config:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
    max_in_flight: 2
errand-config:
  first-errand:
    post-deploy-state: true
    pre-delete-state: do-something
  second-errand:
    post-deploy-state: false
syslog-properties:
  enabled: true
  host: tcp://1.1.1.1
`)))
			})
		})

		When("--include-credentials is used", func() {
			BeforeEach(func() {
				fakeService = setFakeService(internalSelector, false)
				fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{
					{
						Type: "some-product",
						GUID: "some-product-guid",
					},
				}, nil)
			})

			It("includes secret values in the output", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := executeCommand(command, []string{
					"--product-name", "some-product",
					"--include-credentials",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeService.GetStagedProductPropertiesCallCount()).To(Equal(1))
				name, redact := fakeService.GetStagedProductPropertiesArgsForCall(0)
				Expect(name).To(Equal("some-product-guid"))
				Expect(redact).To(BeFalse())

				Expect(logger.PrintlnCallCount()).To(Equal(1))
				output := logger.PrintlnArgsForCall(0)
				Expect(output[0]).To(MatchYAML(`
product-name: some-product
product-properties:
  .properties.collection:
    value:
    - certificate:
        cert_pem: some-secret-value
        private_key_pem: some-secret-value
      name: Certificate
    - certificate2:
        cert_pem: some-secret-value
        private_key_pem: some-secret-value
  .properties.some-string-property:
    value: some-value
  .properties.some-secret-property:
    value:
      secret: some-secret-value
  .properties.simple-credentials:
    value:
      identity: some-secret-value
      password: some-secret-value
  .properties.rsa-cert-credentials:
    value:
      cert_pem: some-secret-value
      private_key_pem: some-secret-value
  .properties.rsa-pkey-credentials:
    value:
      private_key_pem: some-secret-value
  .properties.salted-credentials:
    value:
      identity: some-secret-value
      salt: some-secret-value
      password: some-secret-value
  .properties.some-selector:
    value: internal
network-properties:
  singleton_availability_zone:
    name: az-one
resource-config:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
    max_in_flight: 2
errand-config:
  first-errand:
    post-deploy-state: true
    pre-delete-state: do-something
  second-errand:
    post-deploy-state: false
syslog-properties:
  enabled: true
  host: tcp://1.1.1.1
`))
			})

			Context("and the product has not yet been deployed", func() {
				BeforeEach(func() {
					fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{}, nil)
				})
				It("errors with a helpful message to the operator", func() {
					command := commands.NewStagedConfig(fakeService, logger)
					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--include-credentials",
					})
					Expect(err).To(MatchError("cannot retrieve credentials for product 'some-product': deploy the product and retry"))
				})
			})

			Context("and listing deployed products fails", func() {
				BeforeEach(func() {
					fakeService.ListDeployedProductsReturns(
						[]api.DeployedProductOutput{},
						errors.New("some-error"),
					)
				})

				It("returns an error", func() {
					command := commands.NewStagedConfig(fakeService, logger)
					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--include-credentials",
					})
					Expect(err).To(MatchError("some-error"))
				})
			})
		})

		When("Ops Manager is <=2.3>", func() {
			BeforeEach(func() {
				fakeService.InfoReturns(
					api.Info{
						Version: "2.3",
					}, nil)
				fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{
					{
						Type: "some-product",
						GUID: "some-product-guid",
					},
				}, nil)
				fakeService.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
					Credential: api.Credential{
						Type: "some-secret-type",
						Value: map[string]string{
							"some-secret-key": "some-secret-value",
						},
					},
				}, nil)
			})

			It("does not call the syslog configuration endpoint", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := executeCommand(command, []string{
					"--product-name", "some-product",
					"--include-credentials",
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeService.InfoCallCount()).To(Equal(1))
				Expect(fakeService.GetStagedProductSyslogConfigurationCallCount()).To(Equal(0))
			})
		})

		When("getting Ops Manager info fails", func() {
			BeforeEach(func() {
				fakeService.InfoReturns(
					api.Info{
						Version: "2.4",
					}, errors.New("info error"))
			})

			It("returns an error before making any other calls", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := executeCommand(command, []string{
					"--product-name", "some-product",
					"--include-credentials",
				})
				Expect(err).To(MatchError(ContainSubstring("info error")))
			})
		})

		When("looking up the product GUID fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{}, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := executeCommand(command, []string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		When("looking up the product properties fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductPropertiesReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := executeCommand(command, []string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		When("looking up the network fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductNetworksAndAZsReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := executeCommand(command, []string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		When("listing jobs fails", func() {
			BeforeEach(func() {
				fakeService.ListStagedProductJobsReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := executeCommand(command, []string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		When("looking up the job fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductJobResourceConfigReturns(api.JobProperties{}, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := executeCommand(command, []string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		When("syslog properties returns an error", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductSyslogConfigurationReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := executeCommand(command, []string{
					"--product-name", "some-product",
				})
				Expect(err.Error()).To(ContainSubstring("some-error"))
			})
		})
	})

	When("an arbitrarily long non-selector property path is present", func() {
		It("preserves that path", func() {
			fakeService := &fakes.StagedConfigService{}
			fakeService.GetStagedProductPropertiesReturns(
				map[string]api.ResponseProperty{
					".properties.rds_mysql_plans[1].CreateDBInstance_EngineVersion.asdf": {
						Value:        "some-value",
						Configurable: true,
					},
				}, nil)

			command := commands.NewStagedConfig(fakeService, logger)
			err := executeCommand(command, []string{
				"--product-name", "some-product",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			output := logger.PrintlnArgsForCall(0)
			manifest := `
product-name: some-product
product-properties:
  .properties.rds_mysql_plans[1].CreateDBInstance_EngineVersion.asdf:
    value: "some-value"
`
			Expect(output).To(ContainElement(MatchYAML(manifest)))

		})
	})

	When("a selector is present", func() {
		stagedConfigTemplate := `---
product-name: some-product
product-properties:
  .properties.collection:
    value:
    - name: Certificate
  .properties.some-string-property:
    value: some-value
%s
network-properties:
  singleton_availability_zone:
    name: az-one
resource-config:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
    max_in_flight: 2
errand-config:
  first-errand:
    post-deploy-state: true
    pre-delete-state: do-something
  second-errand:
    post-deploy-state: false
syslog-properties:
  enabled: true
  host: tcp://1.1.1.1
`
		Context("and both selected_option and value are available", func() {
			BeforeEach(func() {
				internalSelector = api.ResponseProperty{
					Value:          "internal",
					SelectedOption: "internal-option",
					Type:           "selector",
					Configurable:   true,
				}
				fakeService = setFakeService(internalSelector, true)
			})

			It("will include selected_option if available", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := executeCommand(command, []string{
					"--product-name", "some-product",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(logger.PrintlnCallCount()).To(Equal(1))
				output := logger.PrintlnArgsForCall(0)
				selectedAndValue := `
  .properties.some-selector:
    selected_option: internal-option
    value: internal
  .properties.some-selector.internal-option.name:
    value: "Hello World"`
				Expect(output).To(ContainElement(MatchYAML(fmt.Sprintf(stagedConfigTemplate, selectedAndValue))))
			})
		})

		Context("and only value is available", func() {
			BeforeEach(func() {
				internalSelector = api.ResponseProperty{
					Value:        "internal-option",
					Type:         "selector",
					Configurable: true,
				}
				fakeService = setFakeService(internalSelector, true)
			})

			It("will include properties dependent on the selected selector if available", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := executeCommand(command, []string{
					"--product-name", "some-product",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(logger.PrintlnCallCount()).To(Equal(1))
				output := logger.PrintlnArgsForCall(0)
				selectedAndValue := `
  .properties.some-selector:
    value: internal-option
  .properties.some-selector.internal-option.name:
    value: "Hello World"`
				Expect(output).To(ContainElement(MatchYAML(fmt.Sprintf(stagedConfigTemplate, selectedAndValue))))
			})
		})
	})
})

func setFakeService(internalSelector api.ResponseProperty, redact bool) *fakes.StagedConfigService {
	fakeService := &fakes.StagedConfigService{}
	secretValue := "some-secret-value"
	if redact {
		secretValue = "***"
	}
	fakeService.GetStagedProductPropertiesReturns(
		map[string]api.ResponseProperty{
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
					"secret": secretValue,
				},
				IsCredential: true,
				Configurable: true,
			},
			".properties.simple-credentials": {
				Type: "simple_credentials",
				Value: map[string]interface{}{
					"identity": secretValue,
					"password": secretValue,
				},
				IsCredential: true,
				Configurable: true,
			},
			".properties.rsa-cert-credentials": {
				Type: "rsa_cert_credentials",
				Value: map[string]interface{}{
					"cert_pem":        secretValue,
					"private_key_pem": secretValue,
				},
				IsCredential: true,
				Configurable: true,
			},
			".properties.rsa-pkey-credentials": {
				Type: "rsa_pkey_credentials",
				Value: map[string]interface{}{
					"private_key_pem": secretValue,
				},
				IsCredential: true,
				Configurable: true,
			},
			".properties.salted-credentials": {
				Type: "salted_credentials",
				Value: map[string]interface{}{
					"identity": secretValue,
					"salt":     secretValue,
					"password": secretValue,
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
								"cert_pem":        secretValue,
								"private_key_pem": secretValue,
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
								"cert_pem":        secretValue,
								"private_key_pem": secretValue,
							},
						},
					},
				},

				IsCredential: false,
				Configurable: true,
			},
			".properties.some-non-configurable-secret-property": {
				Value: map[string]interface{}{
					"some-secret-type": secretValue,
				},
				IsCredential: true,
				Configurable: false,
			},
			".properties.some-null-property": {
				Value:        nil,
				Configurable: true,
			},
			".properties.some-selector": internalSelector,
			".properties.some-selector.not-internal.some-string-property": {
				Value:        "some-value",
				Configurable: true,
			},
			".properties.some-selector.internal-option.name": {
				Value:        "Hello World",
				Configurable: true,
			},
		}, nil)
	fakeService.GetStagedProductNetworksAndAZsReturns(
		map[string]interface{}{
			"singleton_availability_zone": map[string]string{
				"name": "az-one",
			},
		}, nil)

	fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{
		Product: api.StagedProduct{
			GUID: "some-product-guid",
		},
	}, nil)
	fakeService.ListStagedProductErrandsReturns(api.ErrandsListOutput{
		Errands: []api.Errand{
			{Name: "first-errand", PostDeploy: true, PreDelete: "do-something"},
			{Name: "second-errand", PostDeploy: false},
		},
	}, nil)
	fakeService.ListStagedProductJobsReturns(map[string]string{
		"some-job": "some-job-guid",
	}, nil)
	fakeService.GetStagedProductJobResourceConfigReturns(api.JobProperties{
		"instances": 1.0,
		"instance_type": map[string]interface{}{
			"id": "automatic",
		},
		"additional_vm_extensions": []string{"some-vm-extension"},
	}, nil)
	fakeService.GetStagedProductJobMaxInFlightReturns(map[string]interface{}{
		"some-job-guid": 2,
	}, nil)
	fakeService.GetStagedProductSyslogConfigurationReturns(map[string]interface{}{
		"enabled": true,
		"host":    "tcp://1.1.1.1",
	}, nil)
	fakeService.InfoReturns(
		api.Info{
			Version: "2.4",
		}, nil)
	return fakeService
}
