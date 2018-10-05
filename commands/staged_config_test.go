package commands_test

import (
	"errors"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StagedConfig", func() {
	var (
		logger      *fakes.Logger
		fakeService *fakes.StagedConfigService
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}

		fakeService = &fakes.StagedConfigService{}
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
			InstanceType: api.InstanceType{
				ID: "automatic",
			},
			Instances: 1,
		}, nil)
	})

	Describe("Execute", func() {
		It("writes a config file to stdout", func() {
			command := commands.NewStagedConfig(fakeService, logger)
			err := command.Execute([]string{
				"--product-name", "some-product",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeService.GetStagedProductByNameCallCount()).To(Equal(1))
			Expect(fakeService.GetStagedProductByNameArgsForCall(0)).To(Equal("some-product"))

			Expect(fakeService.GetStagedProductPropertiesCallCount()).To(Equal(1))
			Expect(fakeService.GetStagedProductPropertiesArgsForCall(0)).To(Equal("some-product-guid"))

			Expect(fakeService.GetStagedProductNetworksAndAZsCallCount()).To(Equal(1))
			Expect(fakeService.GetStagedProductNetworksAndAZsArgsForCall(0)).To(Equal("some-product-guid"))

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
    instances: 1
    instance_type:
      id: automatic
errand-config:
  first-errand:
    post-deploy-state: true
    pre-delete-state: do-something
  second-errand:
    post-deploy-state: false
`)))
		})
	})

	Context("when --include-placeholders is used", func() {
		It("replaces *** with interpolatable placeholders and removes non-configurable properties", func() {
			command := commands.NewStagedConfig(fakeService, logger)
			err := command.Execute([]string{
				"--product-name", "some-product",
				"--include-placeholders",
			})
			Expect(err).NotTo(HaveOccurred())

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
    instances: 1
    instance_type:
      id: automatic
errand-config:
  first-errand:
    post-deploy-state: true
    pre-delete-state: do-something
  second-errand:
    post-deploy-state: false
`)))
		})
	})

	Context("when --include-credentials is used", func() {
		BeforeEach(func() {
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

		It("includes secret values in the output", func() {
			command := commands.NewStagedConfig(fakeService, logger)
			err := command.Execute([]string{
				"--product-name", "some-product",
				"--include-credentials",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeService.GetDeployedProductCredentialCallCount()).To(Equal(7))

			apiInputs := []api.GetDeployedProductCredentialInput{}
			for i := 0; i < 7; i++ {
				apiInputs = append(apiInputs, fakeService.GetDeployedProductCredentialArgsForCall(i))
			}
			Expect(apiInputs).To(ConsistOf([]api.GetDeployedProductCredentialInput{
				{
					DeployedGUID:        "some-product-guid",
					CredentialReference: ".properties.some-secret-property",
				},
				{
					DeployedGUID:        "some-product-guid",
					CredentialReference: ".properties.salted-credentials",
				},
				{
					DeployedGUID:        "some-product-guid",
					CredentialReference: ".properties.collection[0].certificate",
				},
				{
					DeployedGUID:        "some-product-guid",
					CredentialReference: ".properties.collection[1].certificate2",
				},
				{
					DeployedGUID:        "some-product-guid",
					CredentialReference: ".properties.simple-credentials",
				},
				{
					DeployedGUID:        "some-product-guid",
					CredentialReference: ".properties.rsa-cert-credentials",
				},
				{
					DeployedGUID:        "some-product-guid",
					CredentialReference: ".properties.rsa-pkey-credentials",
				},
			}))

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			output := logger.PrintlnArgsForCall(0)
			Expect(output).To(ContainElement(MatchYAML(`
product-name: some-product
product-properties:
  .properties.collection:
    value:
    - certificate:
        some-secret-key: some-secret-value
      name: Certificate
    - certificate2:
        some-secret-key: some-secret-value
  .properties.some-string-property:
    value: some-value
  .properties.some-secret-property:
    value:
      some-secret-key: some-secret-value
  .properties.simple-credentials:
    value:
      some-secret-key: some-secret-value
  .properties.rsa-cert-credentials:
    value:
      some-secret-key: some-secret-value
  .properties.rsa-pkey-credentials:
    value:
      some-secret-key: some-secret-value
  .properties.salted-credentials:
    value:
      some-secret-key: some-secret-value
  .properties.some-selector:
    value: internal
network-properties:
  singleton_availability_zone:
    name: az-one
resource-config:
  some-job:
    instances: 1
    instance_type:
      id: automatic
errand-config:
  first-errand:
    post-deploy-state: true
    pre-delete-state: do-something
  second-errand:
    post-deploy-state: false
`)))
		})

		Context("and the product has not yet been deployed", func() {
			BeforeEach(func() {
				fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{}, nil)
			})
			It("errors with a helpful message to the operator", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := command.Execute([]string{
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
				err := command.Execute([]string{
					"--product-name", "some-product",
					"--include-credentials",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("and looking up a credential fails", func() {
			BeforeEach(func() {
				fakeService.GetDeployedProductCredentialReturns(
					api.GetDeployedProductCredentialOutput{},
					errors.New("some-error"),
				)
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
					"--include-credentials",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

	})

	Context("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse staged-config flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when product name is not provided", func() {
			It("returns an error and prints out usage", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not parse staged-config flags: missing required flag \"--product-name\""))
			})
		})

		Context("when looking up the product GUID fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{}, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the product properties fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductPropertiesReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the network fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductNetworksAndAZsReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when listing jobs fails", func() {
			BeforeEach(func() {
				fakeService.ListStagedProductJobsReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the job fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductJobResourceConfigReturns(api.JobProperties{}, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {

			command := commands.NewStagedConfig(nil, nil)

			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command generates a config from a staged product that can be passed in to om configure-product (Note: credentials are not available and will appear as '***')",
				ShortDescription: "**EXPERIMENTAL** generates a config from a staged product",
				Flags:            command.Options,
			}))
		})
	})
})
