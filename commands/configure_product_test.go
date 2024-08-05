package commands_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigureProduct", func() {
	Describe("Execute", func() {
		var (
			service    *fakes.ConfigureProductService
			logger     *fakes.Logger
			config     string
			configFile *os.File
			err        error
		)

		BeforeEach(func() {
			service = &fakes.ConfigureProductService{}
			logger = &fakes.Logger{}
		})

		JustBeforeEach(func() {
			configFile, err = os.CreateTemp("", "config.yml")
			Expect(err).ToNot(HaveOccurred())
			defer configFile.Close()

			_, err = configFile.WriteString(config)
			Expect(err).ToNot(HaveOccurred())
		})

		When("product properties are provided", func() {
			BeforeEach(func() {
				config = fmt.Sprintf(`{"product-name": "cf", "product-properties": %s}`, productProperties)
			})

			It("configures the given product's properties", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)

				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
						{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
					},
				}, nil)

				err := executeCommand(client, []string{
					"--config", configFile.Name(),
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.ListStagedProductsCallCount()).To(Equal(1))
				actual := service.UpdateStagedProductPropertiesArgsForCall(0)
				Expect(actual.GUID).To(Equal("some-product-guid"))
				Expect(actual.Properties).To(MatchJSON(productProperties))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring cf..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("setting properties"))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished setting properties"))

				format, content = logger.PrintfArgsForCall(3)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring product"))
			})

			It("check configuration is complete after configuring", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, "example.com", logger)

				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
						{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
					},
				}, nil)

				service.ListStagedPendingChangesReturns(api.PendingChangesOutput{
					ChangeList: []api.ProductChange{
						{
							GUID:    "some-product-guid",
							Action:  "install",
							Errands: nil,
							CompletenessChecks: &api.CompletenessChecks{
								ConfigurationComplete:       false,
								StemcellPresent:             false,
								ConfigurablePropertiesValid: true,
							},
						},
					},
				}, nil)

				err := executeCommand(client, []string{
					"--config", configFile.Name(),
				})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("configuration not complete.\nThe properties you provided have been set,\nbut some required properties or configuration details are still missing.\nVisit the Ops Manager for details: example.com"))

				Expect(service.ListStagedProductsCallCount()).To(Equal(1))
				actual := service.UpdateStagedProductPropertiesArgsForCall(0)
				Expect(actual.GUID).To(Equal("some-product-guid"))
				Expect(actual.Properties).To(MatchJSON(productProperties))

				Expect(service.ListStagedPendingChangesCallCount()).To(Equal(1))
			})

			It("returns a helpful error message if configuration completeness cannot be validated", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, "example.com", logger)

				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
						{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
					},
				}, nil)

				service.ListStagedPendingChangesReturns(api.PendingChangesOutput{
					ChangeList: []api.ProductChange{
						{
							GUID:               "some-product-guid",
							Action:             "install",
							Errands:            nil,
							CompletenessChecks: nil,
						},
					},
				}, nil)

				err := executeCommand(client, []string{
					"--config", configFile.Name(),
				})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("configuration completeness could not be determined.\nThis feature is only supported for OpsMan 2.2+\nIf you're on older version of OpsMan add the line `validate-config-complete: false` to your config file."))

				Expect(service.ListStagedProductsCallCount()).To(Equal(1))
				actual := service.UpdateStagedProductPropertiesArgsForCall(0)
				Expect(actual.GUID).To(Equal("some-product-guid"))
				Expect(actual.Properties).To(MatchJSON(productProperties))

				Expect(service.ListStagedPendingChangesCallCount()).To(Equal(1))
			})
		})

		When("product network is provided", func() {
			BeforeEach(func() {
				config = fmt.Sprintf(`{"product-name": "cf", "network-properties": %s}`, networkProperties)
			})

			It("configures a product's network", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)

				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
						{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
					},
				}, nil)

				err := executeCommand(client, []string{
					"--config", configFile.Name(),
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.ListStagedProductsCallCount()).To(Equal(1))
				actual := service.UpdateStagedProductNetworksAndAZsArgsForCall(0)
				Expect(actual.GUID).To(Equal("some-product-guid"))
				Expect(actual.NetworksAndAZs).To(MatchJSON(networkProperties))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring cf..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("setting up network"))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished setting up network"))

				format, content = logger.PrintfArgsForCall(3)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring product"))
			})
		})

		When("product syslog is provided", func() {
			BeforeEach(func() {
				config = fmt.Sprintf(`{"product-name": "cf", "syslog-properties": %s}`, syslogProperties)
			})

			It("configures a product's syslog", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)

				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
						{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
					},
				}, nil)

				err := executeCommand(client, []string{
					"--config", configFile.Name(),
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.ListStagedProductsCallCount()).To(Equal(1))
				actual := service.UpdateSyslogConfigurationArgsForCall(0)
				Expect(actual.GUID).To(Equal("some-product-guid"))
				Expect(actual.SyslogConfiguration).To(MatchJSON(syslogProperties))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring cf..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("setting up syslog"))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished setting up syslog"))

				format, content = logger.PrintfArgsForCall(3)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring product"))
			})
		})

		When("product resources are provided", func() {
			BeforeEach(func() {
				config = fmt.Sprintf(`{"product-name": "cf", "resource-config": %s}`, resourceConfig)
			})

			It("configures the resource that is provided", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
						{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
					},
				}, nil)

				service.ListStagedProductJobsReturns(map[string]string{
					"some-job":       "a-guid",
					"some-other-job": "a-different-guid",
					"bad":            "do-not-use",
				}, nil)

				err := executeCommand(client, []string{
					"--config", configFile.Name(),
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.ListStagedProductsCallCount()).To(Equal(1))
				Expect(service.ListStagedProductJobsArgsForCall(0)).To(Equal("some-product-guid"))
				Expect(service.ConfigureJobResourceConfigCallCount()).To(Equal(1))
				productGUID, userConfig := service.ConfigureJobResourceConfigArgsForCall(0)
				Expect(productGUID).To(Equal("some-product-guid"))
				payload, err := json.Marshal(userConfig)
				Expect(err).ToNot(HaveOccurred())
				Expect(payload).To(MatchJSON(`{
          		    "some-job": {
          		        "persistent_disk": {"size_mb": "20480"},
          		        "elb_names": ["some-lb"],
          		        "instance_type": {"id": "m1.medium"},
          		        "instances": 1,
          		        "internet_connected": true,
          		        "max_in_flight": "20%"
          		    },
          		    "some-other-job": {
          		        "persistent_disk": {"size_mb": "20480"},
          		        "instance_type": {"id": "m1.medium"},
          		        "max_in_flight": 1
          		    }
          		}`))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring cf..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("applying resource configurations..."))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished applying resource configurations"))
			})

			It("sets the max in flight for all jobs", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
						{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
					},
				}, nil)

				service.ListStagedProductJobsReturns(map[string]string{
					"some-job":       "a-guid",
					"some-other-job": "a-different-guid",
					"bad":            "do-not-use",
				}, nil)

				err := executeCommand(client, []string{
					"--config", configFile.Name(),
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.UpdateStagedProductJobMaxInFlightCallCount()).To(Equal(1))
				productGUID, payload := service.UpdateStagedProductJobMaxInFlightArgsForCall(0)
				Expect(productGUID).To(Equal("some-product-guid"))
				Expect(payload).To(Equal(map[string]interface{}{
					"a-guid":           "20%",
					"a-different-guid": 1,
				}))

				format, content := logger.PrintfArgsForCall(3)
				Expect(fmt.Sprintf(format, content...)).To(Equal("applying max in flight for the following jobs:"))
			})
		})

		When("interpolating", func() {
			var (
				configFile *os.File
				err        error
			)

			BeforeEach(func() {
				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
						{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
					},
				}, nil)
				service.ListStagedProductJobsReturns(map[string]string{
					"some-job":       "a-guid",
					"some-other-job": "a-different-guid",
					"bad":            "do-not-use",
				}, nil)
			})

			AfterEach(func() {
				os.RemoveAll(configFile.Name())
			})

			When("the config file contains variables", func() {
				Context("passed in a vars-file", func() {
					It("can interpolate variables into the configuration", func() {
						client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)

						configFile, err = os.CreateTemp("", "")
						Expect(err).ToNot(HaveOccurred())

						_, err = configFile.WriteString(productPropertiesWithVariableTemplate)
						Expect(err).ToNot(HaveOccurred())

						varsFile, err := os.CreateTemp("", "")
						Expect(err).ToNot(HaveOccurred())

						_, err = varsFile.WriteString(`password: something-secure`)
						Expect(err).ToNot(HaveOccurred())

						err = executeCommand(client, []string{
							"--config", configFile.Name(),
							"--vars-file", varsFile.Name(),
						})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("given vars", func() {
					It("can interpolate variables into the configuration", func() {
						client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)

						configFile, err = os.CreateTemp("", "")
						Expect(err).ToNot(HaveOccurred())

						_, err = configFile.WriteString(productPropertiesWithVariableTemplate)
						Expect(err).ToNot(HaveOccurred())

						err = executeCommand(client, []string{
							"--config", configFile.Name(),
							"--var", "password=something-secure",
						})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("passed as environment variables", func() {
					It("can interpolate variables into the configuration", func() {
						client := commands.NewConfigureProduct(func() []string { return []string{"OM_VAR_password=something-secure"} }, service, "", logger)

						configFile, err = os.CreateTemp("", "")
						Expect(err).ToNot(HaveOccurred())

						_, err = configFile.WriteString(productPropertiesWithVariableTemplate)
						Expect(err).ToNot(HaveOccurred())

						err = executeCommand(client, []string{
							"--config", configFile.Name(),
							"--vars-env", "OM_VAR",
						})
						Expect(err).ToNot(HaveOccurred())
					})

					It("supports the experimental feature of OM_VARS_ENV", func() {
						os.Setenv("OM_VARS_ENV", "OM_VAR")
						defer os.Unsetenv("OM_VARS_ENV")

						client := commands.NewConfigureProduct(func() []string { return []string{"OM_VAR_password=something-secure"} }, service, "", logger)

						configFile, err = os.CreateTemp("", "")
						Expect(err).ToNot(HaveOccurred())

						_, err = configFile.WriteString(productPropertiesWithVariableTemplate)
						Expect(err).ToNot(HaveOccurred())

						err = executeCommand(client, []string{
							"--config", configFile.Name(),
						})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				It("returns an error if missing variables", func() {
					client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)

					configFile, err = os.CreateTemp("", "")
					Expect(err).ToNot(HaveOccurred())

					_, err = configFile.WriteString(productPropertiesWithVariableTemplate)
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(client, []string{
						"--config", configFile.Name(),
					})
					Expect(err).To(MatchError(ContainSubstring("Expected to find variables")))
				})
			})

			When("an ops-file is provided", func() {
				It("can interpolate ops-files into the configuration", func() {
					client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)

					configFile, err = os.CreateTemp("", "")
					Expect(err).ToNot(HaveOccurred())

					_, err = configFile.WriteString(ymlProductProperties)
					Expect(err).ToNot(HaveOccurred())

					opsFile, err := os.CreateTemp("", "")
					Expect(err).ToNot(HaveOccurred())

					_, err = opsFile.WriteString(productOpsFile)
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(client, []string{
						"--config", configFile.Name(),
						"--ops-file", opsFile.Name(),
					})
					Expect(err).ToNot(HaveOccurred())

					Expect(service.ListStagedProductsCallCount()).To(Equal(1))
					Expect(service.UpdateStagedProductPropertiesCallCount()).To(Equal(1))
					Expect(service.UpdateStagedProductPropertiesArgsForCall(0).GUID).To(Equal("some-product-guid"))
					Expect(service.UpdateStagedProductPropertiesArgsForCall(0).Properties).To(MatchJSON(productPropertiesWithOpsFileInterpolated))
				})

				It("returns an error if the ops file is invalid", func() {
					client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)

					configFile, err = os.CreateTemp("", "")
					Expect(err).ToNot(HaveOccurred())

					_, err = configFile.WriteString(ymlProductProperties)
					Expect(err).ToNot(HaveOccurred())

					opsFile, err := os.CreateTemp("", "")
					Expect(err).ToNot(HaveOccurred())

					_, err = opsFile.WriteString(`%%%`)
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(client, []string{
						"-c", configFile.Name(),
						"-o", opsFile.Name(),
					})
					Expect(err).To(MatchError(ContainSubstring("could not find expected directive name")))
				})
			})
		})

		When("GetStagedProductJobResourceConfig returns an error", func() {
			BeforeEach(func() {
				config = fmt.Sprintf(`{"product-name": "cf", "resource-config": %s}`, resourceConfig)
			})
			It("returns an error", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
						{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
					},
				}, nil)

				service.ListStagedProductJobsReturns(map[string]string{
					"some-job":       "a-guid",
					"some-other-job": "a-different-guid",
					"bad":            "do-not-use",
				}, nil)

				service.ConfigureJobResourceConfigReturns(errors.New("some error"))
				err := executeCommand(client, []string{
					"--config", configFile.Name(),
				})

				Expect(err).To(MatchError("failed to configure resources: some error"))
			})
		})

		When("certain fields are not provided in the config", func() {
			BeforeEach(func() {
				config = `{"product-name": "cf"}`
				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
					},
				}, nil)
			})

			It("logs and then does nothing if they are empty", func() {
				command := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)

				err := executeCommand(command, []string{
					"--config", configFile.Name(),
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.ListStagedProductsCallCount()).To(Equal(1))

				Expect(logger.PrintlnCallCount()).To(Equal(6))
				msg := logger.PrintlnArgsForCall(0)[0]
				Expect(msg).To(Equal("network properties are not provided, nothing to do here"))
				msg = logger.PrintlnArgsForCall(1)[0]
				Expect(msg).To(Equal("product properties are not provided, nothing to do here"))
				msg = logger.PrintlnArgsForCall(2)[0]
				Expect(msg).To(Equal("resource config properties are not provided, nothing to do here"))
				msg = logger.PrintlnArgsForCall(3)[0]
				Expect(msg).To(Equal("max in flight properties are not provided, nothing to do here"))
				msg = logger.PrintlnArgsForCall(4)[0]
				Expect(msg).To(Equal("syslog configuration is not provided, nothing to do here"))
				msg = logger.PrintlnArgsForCall(5)[0]
				Expect(msg).To(Equal("errands are not provided, nothing to do here"))
				format, content := logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(ContainSubstring("finished configuring product"))
			})
		})

		When("there is a running installation", func() {
			BeforeEach(func() {
				service.ListInstallationsReturns([]api.InstallationsServiceOutput{
					{
						ID:         999,
						Status:     "running",
						Logs:       "",
						StartedAt:  nil,
						FinishedAt: nil,
						UserName:   "admin",
					},
				}, nil)

				config = `{"product-name": "cf"}`
			})

			It("returns an error", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
				err := executeCommand(client, []string{"--config", configFile.Name()})
				Expect(err).To(MatchError("OpsManager does not allow configuration or staging changes while apply changes are running to prevent data loss for configuration and/or staging changes"))
				Expect(service.ListInstallationsCallCount()).To(Equal(1))
			})
		})

		When("product-version is provided in the config", func() {
			BeforeEach(func() {
				config = fmt.Sprintf(`{
					"product-name": "cf", 
					"product-properties": %s,
					"product-version": 1.2.3
				}`, productProperties)
			})

			It("does not return an error", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)

				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
						{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
					},
				}, nil)

				err := executeCommand(client, []string{
					"--config", configFile.Name(),
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("an error occurs", func() {
			BeforeEach(func() {
				config = `{"product-name": "cf"}`
			})

			When("the product does not exist", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)

					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
						},
					}, nil)

					err := executeCommand(command, []string{
						"--config", configFile.Name(),
					})
					Expect(err).To(MatchError(`could not find product "cf"`))
				})
			})

			When("the product resources cannot be decoded", func() {
				BeforeEach(func() {
					config = `{"product-name": "cf", "resource-config": "%%%%%"}`
				})

				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "cf"},
						},
					}, nil)

					err := executeCommand(command, []string{"--config", configFile.Name()})
					Expect(err).To(MatchError(ContainSubstring("could not be parsed as valid configuration: yaml: unmarshal errors")))
				})
			})

			When("the jobs cannot be fetched", func() {
				BeforeEach(func() {
					config = fmt.Sprintf(`{"product-name": "cf", "resource-config": %s}`, resourceConfig)
				})

				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "cf"},
						},
					}, nil)

					service.ListStagedProductJobsReturns(
						map[string]string{
							"some-job": "a-guid",
						}, errors.New("boom"))

					err := executeCommand(command, []string{"--config", configFile.Name()})
					Expect(err).To(MatchError("failed to fetch jobs: boom"))
				})
			})

			When("the product-name is missing from config", func() {
				BeforeEach(func() {
					config = `{}`
				})

				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
					err := executeCommand(command, []string{"--config", configFile.Name()})
					Expect(err).To(MatchError("could not parse configure-product config: \"product-name\" is required"))
				})
			})

			When("the --config flag is passed", func() {
				When("the provided config path does not exist", func() {
					It("returns an error", func() {
						command := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
						service.ListStagedProductsReturns(api.StagedProductsOutput{
							Products: []api.StagedProduct{
								{GUID: "some-product-guid", Type: "cf"},
							},
						}, nil)
						err := executeCommand(command, []string{"--config", "some/non-existant/path.yml"})
						Expect(err).To(MatchError(ContainSubstring("open some/non-existant/path.yml: no such file or directory")))
					})
				})

				When("the provided config file is not valid yaml", func() {
					var (
						configFile *os.File
						err        error
					)

					AfterEach(func() {
						os.RemoveAll(configFile.Name())
					})

					It("returns an error", func() {
						invalidConfig := "this is not a valid config"
						client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
						service.ListStagedProductsReturns(api.StagedProductsOutput{
							Products: []api.StagedProduct{
								{GUID: "some-product-guid", Type: "cf"},
							},
						}, nil)

						configFile, err = os.CreateTemp("", "")
						Expect(err).ToNot(HaveOccurred())

						_, err = configFile.WriteString(invalidConfig)
						Expect(err).ToNot(HaveOccurred())

						err = executeCommand(client, []string{"--config", configFile.Name()})
						Expect(err).To(MatchError(ContainSubstring("could not be parsed as valid configuration")))

						os.RemoveAll(configFile.Name())
					})
				})
			})

			When("the properties cannot be configured", func() {
				BeforeEach(func() {
					config = `{"product-name": "some-product", "product-properties": {}, "network-properties": {}}`
				})

				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
					service.UpdateStagedProductPropertiesReturns(errors.New("some product error"))

					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "some-product"},
						},
					}, nil)

					err := executeCommand(command, []string{"--config", configFile.Name()})
					Expect(err).To(MatchError("failed to configure product: some product error"))
				})
			})

			When("the networks cannot be configured", func() {
				BeforeEach(func() {
					config = `{"product-name": "some-product", "product-properties": {}, "network-properties": {}}`
				})

				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
					service.UpdateStagedProductNetworksAndAZsReturns(errors.New("some product error"))

					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "some-product"},
						},
					}, nil)

					err := executeCommand(command, []string{"--config", configFile.Name()})
					Expect(err).To(MatchError("failed to configure product: some product error"))
				})
			})

			When("the syslog cannot be configured", func() {
				BeforeEach(func() {
					config = `{"product-name": "some-product", "product-properties": {}, "syslog-properties": {}}`
				})

				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
					service.UpdateSyslogConfigurationReturns(errors.New("some product error"))

					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "some-product"},
						},
					}, nil)

					err := executeCommand(command, []string{"--config", configFile.Name()})
					Expect(err).To(MatchError("failed to configure product: some product error"))
				})
			})

			When("when the syslog cannot be configured", func() {
				BeforeEach(func() {
					config = `{"product-name": "some-product", "product-properties": {}, "syslog-properties": {}}`
				})

				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
					service.UpdateSyslogConfigurationReturns(errors.New("some product error"))

					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "some-product"},
						},
					}, nil)

					err := executeCommand(command, []string{"--config", configFile.Name()})
					Expect(err).To(MatchError("failed to configure product: some product error"))
				})
			})

			When("errand config errors", func() {
				var (
					configFile *os.File
					err        error
				)
				BeforeEach(func() {
					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "cf"},
							{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
						},
					}, nil)
				})
				AfterEach(func() {
					os.RemoveAll(configFile.Name())
				})
				It("errors when calling api", func() {
					service.UpdateStagedProductErrandsReturns(errors.New("error configuring errand"))
					client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)

					configFile, err = os.CreateTemp("", "")
					Expect(err).ToNot(HaveOccurred())

					_, err = configFile.WriteString(errandConfigFile)
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(client, []string{
						"--config", configFile.Name(),
					})
					Expect(err).To(MatchError("failed to set errand state for errand push-usage-service: error configuring errand"))
				})
			})

			Context("with unrecognized top-level-keys", func() {
				It("returns error saying the specified key", func() {
					configYAML := `{"product-name": "cf", "unrecognized-other-key": {}, "unrecognized-key": {"some-attr1": "some-val1"}}`
					configFile, err := os.CreateTemp("", "config.yaml")
					Expect(err).ToNot(HaveOccurred())
					_, err = configFile.WriteString(configYAML)
					Expect(err).ToNot(HaveOccurred())
					Expect(configFile.Close()).ToNot(HaveOccurred())

					client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
					err = executeCommand(client, []string{
						"--config", configFile.Name(),
					})
					Expect(err).To(MatchError(ContainSubstring(`the config file contains unrecognized keys: unrecognized-key, unrecognized-other-key`)))
				})
			})
		})
	})
})

const productProperties = `{
  ".properties.something": {"value": "configure-me"},
  ".a-job.job-property": {"value": {"identity": "username", "password": "example-new-password"} }
}`

const networkProperties = `{
  "singleton_availability_zone": {"name": "az-one"},
  "other_availability_zones": [{"name": "az-two" }, {"name": "az-three"}],
  "network": {"name": "network-one"}
}`

const syslogProperties = `{
    "enabled": true,
    "address": "example.com",
    "port": 514,
    "transport_protocol": "tcp"
}`

const resourceConfig = `{
  "some-job": {
    "instances": 1,
    "persistent_disk": { "size_mb": "20480" },
    "instance_type": { "id": "m1.medium" },
    "internet_connected": true,
    "elb_names": ["some-lb"],
	"max_in_flight": "20%"
  },
  "some-other-job": {
    "persistent_disk": { "size_mb": "20480" },
    "instance_type": { "id": "m1.medium" },
    "max_in_flight": 1
  }
}`

const productPropertiesWithVariableTemplate = `---
product-name: cf
product-properties:
  .properties.something:
    value: configure-me
  .a-job.job-property:
    value:
      identity: username
      password: ((password))`

const ymlProductProperties = `---
product-name: cf
product-properties:
  .properties.something:
    value: configure-me
  .a-job.job-property:
    value:
      identity: username
      password: example-new-password
  .properties.selector:
    value: "Hello World"
    option_value: "hello"
  .properties.another-selector:
    selected_option: "bye"
`

const productOpsFile = `---
- type: replace
  path: /product-properties?/.some.property/value
  value: some-value
`

const productPropertiesWithOpsFileInterpolated = `{
  ".properties.something": {"value": "configure-me"},
  ".a-job.job-property": {"value": {"identity": "username", "password": "example-new-password"} },
  ".some.property": {"value": "some-value"},
  ".properties.selector": {"value": "Hello World", "option_value": "hello", "selected_option":"hello"},
  ".properties.another-selector": {"option_value": "bye", "selected_option":"bye"}
}`

const errandConfigFile = `---
product-name: cf
errand-config:
  smoke_tests:
    post-deploy-state: true
    pre-delete-state: default
  push-usage-service:
    post-deploy-state: false
    pre-delete-state: when-changed
`
