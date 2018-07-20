package commands_test

import (
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/configparser"
	"reflect"
)

var _ = FDescribe("StagedConfig", func() {
	var (
		logger         *fakes.Logger
		fakeConfParser *fakes.ConfigParser
		fakeService    *fakes.StagedConfigService
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}

		fakeConfParser = &fakes.ConfigParser{}

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
		It("uses the default config parser creds handler", func() {
			command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
			err := command.Execute([]string{
				"--product-name", "some-product",
			})
			Expect(err).NotTo(HaveOccurred())

			_, _, _, handler := fakeConfParser.ParsePropertiesArgsForCall(0)
			Expect(reflect.ValueOf(handler)).To(Equal(reflect.ValueOf(configparser.DefaultHandleCredentialFunc())))
		})
	})

	Context("when --include-placeholder is used", func() {
		It("uses the with placeholder config parser creds handler", func() {
			command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
			err := command.Execute([]string{
				"--product-name", "some-product",
				"--include-placeholder",
			})
			Expect(err).NotTo(HaveOccurred())

			_, _, _, handler := fakeConfParser.ParsePropertiesArgsForCall(0)
			Expect(reflect.ValueOf(handler)).To(Equal(reflect.ValueOf(configparser.WithPlaceholderHandleCredentialFunc())))
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

		It("uses the with credential config parser creds handler", func() {
			command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
			err := command.Execute([]string{
				"--product-name", "some-product",
				"--include-credentials",
			})
			Expect(err).NotTo(HaveOccurred())

			_, _, _, handler := fakeConfParser.ParsePropertiesArgsForCall(0)
			Expect(reflect.ValueOf(handler)).To(Equal(reflect.ValueOf(configparser.WithCredentialHandleCredentialFunc(nil))))
		})

		//Context("and the product has not yet been deployed", func() {
		//	BeforeEach(func() {
		//		fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{}, nil)
		//	})
		//	It("errors with a helpful message to the operator", func() {
		//		command := commands.NewStagedConfig(fakeService, logger)
		//		err := command.Execute([]string{
		//			"--product-name", "some-product",
		//			"--include-credentials",
		//		})
		//		Expect(err).To(MatchError("cannot retrieve credentials for product 'some-product': deploy the product and retry"))
		//	})
		//})
		//
		//Context("and listing deployed products fails", func() {
		//	BeforeEach(func() {
		//		fakeService.ListDeployedProductsReturns(
		//			[]api.DeployedProductOutput{},
		//			errors.New("some-error"),
		//		)
		//	})
		//
		//	It("returns an error", func() {
		//		command := commands.NewStagedConfig(fakeService, logger)
		//		err := command.Execute([]string{
		//			"--product-name", "some-product",
		//			"--include-credentials",
		//		})
		//		Expect(err).To(MatchError("some-error"))
		//	})
		//})
		//
		//Context("and looking up a credential fails", func() {
		//	BeforeEach(func() {
		//		fakeService.GetDeployedProductCredentialReturns(
		//			api.GetDeployedProductCredentialOutput{},
		//			errors.New("some-error"),
		//		)
		//	})
		//
		//	It("returns an error", func() {
		//		command := commands.NewStagedConfig(fakeService, logger)
		//		err := command.Execute([]string{
		//			"--product-name", "some-product",
		//			"--include-credentials",
		//		})
		//		Expect(err).To(MatchError("some-error"))
		//	})
		//})

	})
//
//	Context("failure cases", func() {
//		Context("when an unknown flag is provided", func() {
//			It("returns an error", func() {
//				command := commands.NewStagedConfig(fakeService, logger)
//				err := command.Execute([]string{"--badflag"})
//				Expect(err).To(MatchError("could not parse staged-config flags: flag provided but not defined: -badflag"))
//			})
//		})
//
//		Context("when product name is not provided", func() {
//			It("returns an error and prints out usage", func() {
//				command := commands.NewStagedConfig(fakeService, logger)
//				err := command.Execute([]string{})
//				Expect(err).To(MatchError("could not parse staged-config flags: missing required flag \"--product-name\""))
//			})
//		})
//
//		Context("when looking up the product GUID fails", func() {
//			BeforeEach(func() {
//				fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{}, errors.New("some-error"))
//			})
//
//			It("returns an error", func() {
//				command := commands.NewStagedConfig(fakeService, logger)
//				err := command.Execute([]string{
//					"--product-name", "some-product",
//				})
//				Expect(err).To(MatchError("some-error"))
//			})
//		})
//
//		Context("when looking up the product properties fails", func() {
//			BeforeEach(func() {
//				fakeService.GetStagedProductPropertiesReturns(nil, errors.New("some-error"))
//			})
//
//			It("returns an error", func() {
//				command := commands.NewStagedConfig(fakeService, logger)
//				err := command.Execute([]string{
//					"--product-name", "some-product",
//				})
//				Expect(err).To(MatchError("some-error"))
//			})
//		})
//
//		Context("when looking up the network fails", func() {
//			BeforeEach(func() {
//				fakeService.GetStagedProductNetworksAndAZsReturns(nil, errors.New("some-error"))
//			})
//
//			It("returns an error", func() {
//				command := commands.NewStagedConfig(fakeService, logger)
//				err := command.Execute([]string{
//					"--product-name", "some-product",
//				})
//				Expect(err).To(MatchError("some-error"))
//			})
//		})
//
//		Context("when listing jobs fails", func() {
//			BeforeEach(func() {
//				fakeService.ListStagedProductJobsReturns(nil, errors.New("some-error"))
//			})
//
//			It("returns an error", func() {
//				command := commands.NewStagedConfig(fakeService, logger)
//				err := command.Execute([]string{
//					"--product-name", "some-product",
//				})
//				Expect(err).To(MatchError("some-error"))
//			})
//		})
//
//		Context("when looking up the job fails", func() {
//			BeforeEach(func() {
//				fakeService.GetStagedProductJobResourceConfigReturns(api.JobProperties{}, errors.New("some-error"))
//			})
//
//			It("returns an error", func() {
//				command := commands.NewStagedConfig(fakeService, logger)
//				err := command.Execute([]string{
//					"--product-name", "some-product",
//				})
//				Expect(err).To(MatchError("some-error"))
//			})
//		})
//
//	})
//
//	Describe("Usage", func() {
//		It("returns usage information for the command", func() {
//
//			command := commands.NewStagedConfig(nil, nil)
//
//			Expect(command.Usage()).To(Equal(jhanda.Usage{
//				Description:      "This command generates a config from a staged product that can be passed in to om configure-product (Note: credentials are not available and will appear as '***')",
//				ShortDescription: "**EXPERIMENTAL** generates a config from a staged product",
//				Flags:            command.Options,
//			}))
//		})
	//})
})
