package commands_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pivotal-cf/jhanda"
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
			configFile, err = ioutil.TempFile("", "config.yml")
			Expect(err).NotTo(HaveOccurred())
			defer configFile.Close()

			_, err = configFile.WriteString(config)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when product properties is provided", func() {
			BeforeEach(func() {
				config = fmt.Sprintf(`{"product-name": "cf", "product-properties": %s}`, productProperties)
			})

			It("configures a product's properties", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
						{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
					},
				}, nil)

				err := client.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.ListStagedProductsCallCount()).To(Equal(1))
				actual := service.UpdateStagedProductPropertiesArgsForCall(0)
				Expect(actual.GUID).To(Equal("some-product-guid"))
				Expect(actual.Properties).To(MatchJSON(productProperties))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring product..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("setting properties"))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished setting properties"))

				format, content = logger.PrintfArgsForCall(3)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring product"))
			})
		})

		Context("when product network is provided", func() {
			BeforeEach(func() {
				config = fmt.Sprintf(`{"product-name": "cf", "network-properties": %s}`, networkProperties)
			})

			It("configures a product's network", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
						{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
					},
				}, nil)

				err := client.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.ListStagedProductsCallCount()).To(Equal(1))
				actual := service.UpdateStagedProductNetworksAndAZsArgsForCall(0)
				Expect(actual.GUID).To(Equal("some-product-guid"))
				Expect(actual.NetworksAndAZs).To(MatchJSON(networkProperties))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring product..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("setting up network"))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished setting up network"))

				format, content = logger.PrintfArgsForCall(3)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring product"))
			})
		})

		Context("when product resources are provided", func() {
			BeforeEach(func() {
				config = fmt.Sprintf(`{"product-name": "cf", "resource-config": %s}`, resourceConfig)
			})

			It("configures the resource that is provided", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
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

				service.GetStagedProductJobResourceConfigStub = func(productGUID, jobGUID string) (api.JobProperties, error) {
					if productGUID == "some-product-guid" {
						switch jobGUID {
						case "a-guid":
							apiReturn := api.JobProperties{
								Instances:         0,
								PersistentDisk:    &api.Disk{Size: "000"},
								InstanceType:      api.InstanceType{ID: "t2.micro"},
								InternetConnected: new(bool),
								LBNames:           []string{"pre-existing-1"},
							}

							return apiReturn, nil
						case "a-different-guid":
							apiReturn := api.JobProperties{
								Instances:         2,
								PersistentDisk:    &api.Disk{Size: "20480"},
								InstanceType:      api.InstanceType{ID: "m1.medium"},
								InternetConnected: new(bool),
								LBNames:           []string{"pre-existing-2"},
							}

							*apiReturn.InternetConnected = true

							return apiReturn, nil
						default:
							return api.JobProperties{}, nil
						}
					}
					return api.JobProperties{}, errors.New("guid not found")
				}

				err := client.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.ListStagedProductsCallCount()).To(Equal(1))
				Expect(service.ListStagedProductJobsArgsForCall(0)).To(Equal("some-product-guid"))
				Expect(service.UpdateStagedProductJobResourceConfigCallCount()).To(Equal(2))

				argProductGUID, argJobGUID, argProperties := service.UpdateStagedProductJobResourceConfigArgsForCall(0)
				Expect(argProductGUID).To(Equal("some-product-guid"))
				Expect(argJobGUID).To(Equal("a-guid"))

				jobProperties := api.JobProperties{
					Instances:         float64(1),
					PersistentDisk:    &api.Disk{Size: "20480"},
					InstanceType:      api.InstanceType{ID: "m1.medium"},
					InternetConnected: new(bool),
					LBNames:           []string{"some-lb"},
				}

				*jobProperties.InternetConnected = true

				argProductGUID, argJobGUID, argProperties = service.UpdateStagedProductJobResourceConfigArgsForCall(1)
				Expect(argProductGUID).To(Equal("some-product-guid"))
				Expect(argJobGUID).To(Equal("a-different-guid"))

				jobProperties = api.JobProperties{
					Instances:         2,
					PersistentDisk:    &api.Disk{Size: "20480"},
					InstanceType:      api.InstanceType{ID: "m1.medium"},
					InternetConnected: new(bool),
					LBNames:           []string{"pre-existing-2"},
				}

				*jobProperties.InternetConnected = true

				Expect(argProperties).To(Equal(jobProperties))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring product..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("applying resource configuration for the following jobs:"))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("\tsome-job"))

				format, content = logger.PrintfArgsForCall(3)
				Expect(fmt.Sprintf(format, content...)).To(Equal("\tsome-other-job"))

				format, content = logger.PrintfArgsForCall(4)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring product"))
			})
		})

		Context("when interpolating", func() {
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

				service.GetStagedProductJobResourceConfigStub = func(productGUID, jobGUID string) (api.JobProperties, error) {
					if productGUID == "some-product-guid" {
						switch jobGUID {
						case "a-guid":
							apiReturn := api.JobProperties{
								Instances:         0,
								PersistentDisk:    &api.Disk{Size: "000"},
								InstanceType:      api.InstanceType{ID: "t2.micro"},
								InternetConnected: new(bool),
								LBNames:           []string{"pre-existing-1"},
							}

							return apiReturn, nil
						case "a-different-guid":
							apiReturn := api.JobProperties{
								Instances:         2,
								PersistentDisk:    &api.Disk{Size: "20480"},
								InstanceType:      api.InstanceType{ID: "m1.medium"},
								InternetConnected: new(bool),
								LBNames:           []string{"pre-existing-2"},
							}

							*apiReturn.InternetConnected = true

							return apiReturn, nil
						default:
							return api.JobProperties{}, nil
						}
					}
					return api.JobProperties{}, errors.New("guid not found")
				}
			})

			AfterEach(func() {
				os.RemoveAll(configFile.Name())
			})

			Context("when the config file contains variables", func() {
				Context("passed in a vars-file", func() {
					It("can interpolate variables into the configuration", func() {
						client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

						configFile, err = ioutil.TempFile("", "")
						Expect(err).NotTo(HaveOccurred())

						_, err = configFile.WriteString(productPropertiesWithVariables)
						Expect(err).NotTo(HaveOccurred())

						varsFile, err := ioutil.TempFile("", "")
						Expect(err).NotTo(HaveOccurred())

						_, err = varsFile.WriteString(`password: something-secure`)
						Expect(err).NotTo(HaveOccurred())

						err = client.Execute([]string{
							"--config", configFile.Name(),
							"--vars-file", varsFile.Name(),
						})
						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("passed as environment variables", func() {
					It("can interpolate variables into the configuration", func() {
						client := commands.NewConfigureProduct(
							func() []string { return []string{"OM_VAR_password=something-secure"} },
							service,
							logger)

						configFile, err = ioutil.TempFile("", "")
						Expect(err).NotTo(HaveOccurred())

						_, err = configFile.WriteString(productPropertiesWithVariables)
						Expect(err).NotTo(HaveOccurred())

						err = client.Execute([]string{
							"--config", configFile.Name(),
							"--vars-env", "OM_VAR",
						})
						Expect(err).NotTo(HaveOccurred())
					})
				})

				It("returns an error if missing variables", func() {
					client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(productPropertiesWithVariables)
					Expect(err).NotTo(HaveOccurred())

					err = client.Execute([]string{
						"--config", configFile.Name(),
					})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Expected to find variables"))
				})
			})

			Context("when an ops-file is provided", func() {
				It("can interpolate ops-files into the configuration", func() {
					client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(ymlProductProperties)
					Expect(err).NotTo(HaveOccurred())

					opsFile, err := ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = opsFile.WriteString(productOpsFile)
					Expect(err).NotTo(HaveOccurred())

					err = client.Execute([]string{
						"--config", configFile.Name(),
						"--ops-file", opsFile.Name(),
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(service.ListStagedProductsCallCount()).To(Equal(1))
					Expect(service.UpdateStagedProductPropertiesCallCount()).To(Equal(1))
					Expect(service.UpdateStagedProductPropertiesArgsForCall(0).GUID).To(Equal("some-product-guid"))
					Expect(service.UpdateStagedProductPropertiesArgsForCall(0).Properties).To(MatchJSON(productPropertiesWithOpsFileInterpolated))
				})

				It("returns an error if the ops file is invalid", func() {
					client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(ymlProductProperties)
					Expect(err).NotTo(HaveOccurred())

					opsFile, err := ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = opsFile.WriteString(`%%%`)
					Expect(err).NotTo(HaveOccurred())

					err = client.Execute([]string{
						"-c", configFile.Name(),
						"-o", opsFile.Name(),
					})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("could not find expected directive name"))
				})
			})
		})

		Context("when the instance count is not an int", func() {
			BeforeEach(func() {
				config = fmt.Sprintf(`{"product-name": "cf", "resource-config": %s}`, automaticResourceConfig)
			})

			It("configures the resource that is provided", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
					},
				}, nil)

				service.ListStagedProductJobsReturns(map[string]string{
					"some-job": "a-guid",
				}, nil)

				service.GetStagedProductJobResourceConfigStub = func(productGUID, jobGUID string) (api.JobProperties, error) {
					if productGUID == "some-product-guid" {
						switch jobGUID {
						case "a-guid":
							apiReturn := api.JobProperties{
								Instances:         0,
								PersistentDisk:    &api.Disk{Size: "000"},
								InstanceType:      api.InstanceType{ID: "t2.micro"},
								InternetConnected: new(bool),
								LBNames:           []string{"pre-existing-1"},
							}

							return apiReturn, nil
						default:
							return api.JobProperties{}, nil
						}
					}
					return api.JobProperties{}, errors.New("guid not found")
				}

				err := client.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).NotTo(HaveOccurred())

				_, _, argProperties := service.UpdateStagedProductJobResourceConfigArgsForCall(0)

				jobProperties := api.JobProperties{
					Instances:         "automatic",
					PersistentDisk:    &api.Disk{Size: "20480"},
					InstanceType:      api.InstanceType{ID: "m1.medium"},
					InternetConnected: new(bool),
					LBNames:           []string{"some-lb"},
				}

				*jobProperties.InternetConnected = true

				Expect(argProperties).To(Equal(jobProperties))
			})
		})

		Context("when GetStagedProductJobResourceConfig returns an error", func() {
			BeforeEach(func() {
				config = fmt.Sprintf(`{"product-name": "cf", "resource-config": %s}`, resourceConfig)
			})
			It("returns an error", func() {
				client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
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

				service.GetStagedProductJobResourceConfigReturns(api.JobProperties{}, errors.New("some error"))
				err := client.Execute([]string{
					"--config", configFile.Name(),
				})

				Expect(err).To(MatchError("could not fetch existing job configuration: some error"))
			})
		})

		Context("when certain fields are not provided in the config", func() {
			BeforeEach(func() {
				config = `{"product-name": "cf"}`
				service.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "some-product-guid", Type: "cf"},
					},
				}, nil)
			})

			It("logs and then does nothing if network is empty", func() {
				command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

				err := command.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.ListStagedProductsCallCount()).To(Equal(1))

				msg := logger.PrintlnArgsForCall(0)[0]
				Expect(msg).To(Equal("network properties are not provided, nothing to do here"))
				msg = logger.PrintlnArgsForCall(1)[0]
				Expect(msg).To(Equal("product properties are not provided, nothing to do here"))
				msg = logger.PrintlnArgsForCall(2)[0]
				Expect(msg).To(Equal("resource config properties are not provided, nothing to do here"))
				msg = logger.PrintlnArgsForCall(3)[0]
				Expect(msg).To(Equal("errands are not provided, nothing to do here"))
				format, content := logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(ContainSubstring("finished configuring product"))
			})
		})

		Context("when there is a running installation", func() {
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
				client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
				err := client.Execute([]string{"--config", configFile.Name()})
				Expect(err).To(MatchError("OpsManager does not allow configuration or staging changes while apply changes are running to prevent data loss for configuration and/or staging changes"))
				Expect(service.ListInstallationsCallCount()).To(Equal(1))
			})
		})

		Context("when an error occurs", func() {
			BeforeEach(func() {
				config = `{"product-name": "cf"}`
			})

			Context("when the product does not exist", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
						},
					}, nil)

					err := command.Execute([]string{
						"--config", configFile.Name(),
					})
					Expect(err).To(MatchError(`could not find product "cf"`))
				})
			})

			Context("when the product resources cannot be decoded", func() {
				BeforeEach(func() {
					config = `{"product-name": "cf", "resource-config": "%%%%%"}`
				})

				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "cf"},
						},
					}, nil)

					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError(ContainSubstring("could not be parsed as valid configuration: yaml: unmarshal errors")))
				})
			})

			Context("when the jobs cannot be fetched", func() {
				BeforeEach(func() {
					config = fmt.Sprintf(`{"product-name": "cf", "resource-config": %s}`, resourceConfig)
				})

				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "cf"},
						},
					}, nil)

					service.ListStagedProductJobsReturns(
						map[string]string{
							"some-job": "a-guid",
						}, errors.New("boom"))

					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError("failed to fetch jobs: boom"))
				})
			})

			Context("when resources fail to configure", func() {
				BeforeEach(func() {
					config = fmt.Sprintf(`{"product-name": "cf", "resource-config": %s}`, resourceConfig)
				})

				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "cf"},
						},
					}, nil)

					service.ListStagedProductJobsReturns(
						map[string]string{
							"some-job": "a-guid",
						}, nil)

					service.UpdateStagedProductJobResourceConfigReturns(errors.New("bad things happened"))

					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError("failed to configure resources: bad things happened"))
				})
			})

			Context("when an unknown flag is provided", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
					err := command.Execute([]string{"--badflag"})
					Expect(err).To(MatchError("could not parse configure-product flags: flag provided but not defined: -badflag"))
				})
			})

			Context("when the product-name is missing from config", func() {
				BeforeEach(func() {
					config = `{}`
				})

				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError("could not parse configure-product config: \"product-name\" is required"))
				})
			})

			Context("when the --config flag is passed", func() {
				Context("when the provided config path does not exist", func() {
					It("returns an error", func() {
						command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
						service.ListStagedProductsReturns(api.StagedProductsOutput{
							Products: []api.StagedProduct{
								{GUID: "some-product-guid", Type: "cf"},
							},
						}, nil)
						err := command.Execute([]string{"--config", "some/non-existant/path.yml"})
						Expect(err.Error()).To(ContainSubstring("open some/non-existant/path.yml: no such file or directory"))
					})
				})

				Context("when the provided config file is not valid yaml", func() {
					var (
						configFile *os.File
						err        error
					)

					AfterEach(func() {
						os.RemoveAll(configFile.Name())
					})

					It("returns an error", func() {
						invalidConfig := "this is not a valid config"
						client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
						service.ListStagedProductsReturns(api.StagedProductsOutput{
							Products: []api.StagedProduct{
								{GUID: "some-product-guid", Type: "cf"},
							},
						}, nil)

						configFile, err = ioutil.TempFile("", "")
						Expect(err).NotTo(HaveOccurred())

						_, err = configFile.WriteString(invalidConfig)
						Expect(err).NotTo(HaveOccurred())

						err = client.Execute([]string{"--config", configFile.Name()})
						Expect(err).To(MatchError(ContainSubstring("could not be parsed as valid configuration")))

						os.RemoveAll(configFile.Name())
					})
				})
			})

			Context("when the properties cannot be configured", func() {
				BeforeEach(func() {
					config = `{"product-name": "some-product", "product-properties": {}, "network-properties": {}}`
				})

				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
					service.UpdateStagedProductPropertiesReturns(errors.New("some product error"))

					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "some-product"},
						},
					}, nil)

					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError("failed to configure product: some product error"))
				})
			})

			Context("when the networks cannot be configured", func() {
				BeforeEach(func() {
					config = `{"product-name": "some-product", "product-properties": {}, "network-properties": {}}`
				})

				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
					service.UpdateStagedProductNetworksAndAZsReturns(errors.New("some product error"))

					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "some-product"},
						},
					}, nil)

					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError("failed to configure product: some product error"))
				})
			})

			Context("when errand config errors", func() {
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
					client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(errandConfigFile)
					Expect(err).NotTo(HaveOccurred())

					err = client.Execute([]string{
						"--config", configFile.Name(),
					})
					Expect(err).To(MatchError("failed to set errand state for errand push-usage-service: error configuring errand"))
				})
			})

			Context("with unrecognized top-level-keys", func() {
				It("returns error saying the specified key", func() {
					configYAML := `{"product-name": "cf", "unrecognized-other-key": {}, "unrecognized-key": {"some-attr1": "some-val1"}}`
					configFile, err := ioutil.TempFile("", "config.yaml")
					Expect(err).ToNot(HaveOccurred())
					_, err = configFile.WriteString(configYAML)
					Expect(err).ToNot(HaveOccurred())
					Expect(configFile.Close()).ToNot(HaveOccurred())

					client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
					err = client.Execute([]string{
						"--config", configFile.Name(),
					})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`the config file contains unrecognized keys: unrecognized-key, unrecognized-other-key`))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewConfigureProduct(nil, nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command configures a staged product",
				ShortDescription: "configures a staged product",
				Flags:            command.Options,
			}))
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

const resourceConfig = `{
  "some-job": {
    "instances": 1,
    "persistent_disk": { "size_mb": "20480" },
    "instance_type": { "id": "m1.medium" },
    "internet_connected": true,
    "elb_names": ["some-lb"]
  },
  "some-other-job": {
    "persistent_disk": { "size_mb": "20480" },
    "instance_type": { "id": "m1.medium" }
  }
}`

const automaticResourceConfig = `{
  "some-job": {
    "instances": "automatic",
    "persistent_disk": { "size_mb": "20480" },
    "instance_type": { "id": "m1.medium" },
    "internet_connected": true,
    "elb_names": ["some-lb"]
  }
}`

const productPropertiesFile = `---
product-name: cf
product-properties:
  .properties.something:
    value: configure-me
  .a-job.job-property:
    value:
      identity: username
      password: example-new-password
`

const productPropertiesWithVariables = `---
product-name: cf
product-properties:
  .properties.something:
    value: configure-me
  .a-job.job-property:
    value:
      identity: username
      password: ((password))`

const networkPropertiesFile = `---
product-name: cf
network-properties:
  singleton_availability_zone:
    name: az-one
  other_availability_zones:
    - name: az-two
    - name: az-three
  network:
    name: network-one
product-properties:
`

const resourceConfigFile = `---
product-name: cf
resource-config:
  some-job:
    instances: 1
    persistent_disk:
      size_mb: "20480"
    instance_type:
      id: m1.medium
    internet_connected: true
    elb_names:
      - some-lb
  some-other-job:
    persistent_disk:
      size_mb: "20480"
    instance_type:
      id: m1.medium
`

const ymlProductProperties = `---
product-name: cf
product-properties:
  .properties.something:
    value: configure-me
  .a-job.job-property:
    value:
      identity: username
      password: example-new-password
`

const productOpsFile = `---
- type: replace
  path: /product-properties?/.some.property/value
  value: some-value
`

const productPropertiesWithOpsFileInterpolated = `{
  ".properties.something": {"value": "configure-me"},
  ".a-job.job-property": {"value": {"identity": "username", "password": "example-new-password"} },
  ".some.property": {"value": "some-value"}
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

func marshallResponse(input interface{}) []byte {
	resp, err := json.Marshal(input)
	Expect(err).ToNot(HaveOccurred())
	return resp
}
