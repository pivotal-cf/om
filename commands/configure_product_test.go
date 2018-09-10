package commands_test

import (
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
product-properties:
  .properties.something:
    value: configure-me
  .a-job.job-property:
    value:
      identity: username
      password: ((password))`

const networkPropertiesFile = `---
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
errand-config:
  smoke_tests:
    post-deploy-state: true
    pre-delete-state: default
  push-usage-service:
    post-deploy-state: false
    pre-delete-state: when-changed
`

var _ = Describe("ConfigureProduct", func() {
	Describe("Execute", func() {
		var (
			service *fakes.ConfigureProductService
			logger  *fakes.Logger
		)

		BeforeEach(func() {
			service = &fakes.ConfigureProductService{}
			logger = &fakes.Logger{}
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
				"--product-name", "cf",
				"--product-properties", productProperties,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.ListStagedProductsCallCount()).To(Equal(1))
			Expect(service.UpdateStagedProductPropertiesArgsForCall(0)).To(Equal(api.UpdateStagedProductPropertiesInput{
				GUID:       "some-product-guid",
				Properties: productProperties,
			}))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuring product..."))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("setting properties"))

			format, content = logger.PrintfArgsForCall(2)
			Expect(fmt.Sprintf(format, content...)).To(Equal("finished setting properties"))

			format, content = logger.PrintfArgsForCall(3)
			Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring product"))
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
				"--product-name", "cf",
				"--product-network", networkProperties,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.ListStagedProductsCallCount()).To(Equal(1))
			Expect(service.UpdateStagedProductNetworksAndAZsArgsForCall(0)).To(Equal(api.UpdateStagedProductNetworksAndAZsInput{
				GUID:           "some-product-guid",
				NetworksAndAZs: networkProperties,
			}))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuring product..."))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("setting up network"))

			format, content = logger.PrintfArgsForCall(2)
			Expect(fmt.Sprintf(format, content...)).To(Equal("finished setting up network"))

			format, content = logger.PrintfArgsForCall(3)
			Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring product"))
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
				"--product-name", "cf",
				"--product-resources", resourceConfig,
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

		Context("when the --config flag is passed", func() {
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

			Context("when the config file contains product-name", func() {
				It("reads product-name from config file", func() {
					client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(productPropertiesFile)
					Expect(err).NotTo(HaveOccurred())

					err = client.Execute([]string{
						"--config", configFile.Name(),
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(service.UpdateStagedProductPropertiesArgsForCall(0).GUID).To(Equal("some-product-guid"))
					Expect(service.UpdateStagedProductPropertiesArgsForCall(0).Properties).To(MatchJSON(productProperties))

				})
			})

			Context("when the config file contains product-name and is passed as a flag", func() {
				It("overrides the config value with the flag value", func() {
					client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(productPropertiesFile)
					Expect(err).NotTo(HaveOccurred())

					err = client.Execute([]string{
						"--config", configFile.Name(),
						"--product-name", "something-else",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(service.UpdateStagedProductPropertiesArgsForCall(0).GUID).To(Equal("not-the-guid-you-are-looking-for"))
				})
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
							"--product-name", "cf",
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
							"--product-name", "cf",
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
						"--product-name", "cf",
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
						"--product-name", "cf",
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
						"-n", "cf",
						"-c", configFile.Name(),
						"-o", opsFile.Name(),
					})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("could not find expected directive name"))
				})
			})

			Context("when the config file only contains product properties", func() {
				It("configures only the product properties", func() {
					client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(productPropertiesFile)
					Expect(err).NotTo(HaveOccurred())

					err = client.Execute([]string{
						"--product-name", "cf",
						"--config", configFile.Name(),
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(service.ListStagedProductsCallCount()).To(Equal(1))
					Expect(service.UpdateStagedProductPropertiesCallCount()).To(Equal(1))
					Expect(service.UpdateStagedProductPropertiesArgsForCall(0).GUID).To(Equal("some-product-guid"))
					Expect(service.UpdateStagedProductPropertiesArgsForCall(0).Properties).To(MatchJSON(productProperties))
					Expect(service.UpdateStagedProductNetworksAndAZsCallCount()).To(Equal(0))
					Expect(service.UpdateStagedProductJobResourceConfigCallCount()).To(Equal(0))

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

			Context("when the config file only contains network properties", func() {
				It("configures only the network properties", func() {
					client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(networkPropertiesFile)
					Expect(err).NotTo(HaveOccurred())

					err = client.Execute([]string{
						"--product-name", "cf",
						"--config", configFile.Name(),
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(service.ListStagedProductsCallCount()).To(Equal(1))
					Expect(service.UpdateStagedProductNetworksAndAZsCallCount()).To(Equal(1))
					Expect(service.UpdateStagedProductNetworksAndAZsArgsForCall(0).GUID).To(Equal("some-product-guid"))
					Expect(service.UpdateStagedProductNetworksAndAZsArgsForCall(0).NetworksAndAZs).To(MatchJSON(networkProperties))
					Expect(service.UpdateStagedProductPropertiesCallCount()).To(Equal(0))
					Expect(service.UpdateStagedProductJobResourceConfigCallCount()).To(Equal(0))

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

			Context("when the config file contains only resource properties", func() {
				It("configures only the resource properties", func() {
					client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(resourceConfigFile)
					Expect(err).NotTo(HaveOccurred())

					err = client.Execute([]string{
						"--product-name", "cf",
						"--config", configFile.Name(),
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(service.ListStagedProductsCallCount()).To(Equal(1))
					Expect(service.UpdateStagedProductPropertiesCallCount()).To(Equal(0))
					Expect(service.UpdateStagedProductNetworksAndAZsCallCount()).To(Equal(0))

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

			Context("when the config file contains only errand properties", func() {
				It("configures only the errand properties", func() {
					client := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(errandConfigFile)
					Expect(err).NotTo(HaveOccurred())

					err = client.Execute([]string{
						"--product-name", "cf",
						"--config", configFile.Name(),
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(service.ListStagedProductsCallCount()).To(Equal(1))
					Expect(service.UpdateStagedProductPropertiesCallCount()).To(Equal(0))
					Expect(service.UpdateStagedProductNetworksAndAZsCallCount()).To(Equal(0))
					Expect(service.UpdateStagedProductJobResourceConfigCallCount()).To(Equal(0))

					Expect(service.UpdateStagedProductErrandsCallCount()).To(Equal(2))

					argProductGUID, argErrandName, argPostDeployState, argPreDeleteState := service.UpdateStagedProductErrandsArgsForCall(0)
					Expect(argProductGUID).To(Equal("some-product-guid"))
					Expect(argErrandName).To(Equal("push-usage-service"))
					Expect(argPostDeployState).To(Equal(false))
					Expect(argPreDeleteState).To(Equal("when-changed"))

					argProductGUID, argErrandName, argPostDeployState, argPreDeleteState = service.UpdateStagedProductErrandsArgsForCall(1)
					Expect(argProductGUID).To(Equal("some-product-guid"))
					Expect(argErrandName).To(Equal("smoke_tests"))
					Expect(argPostDeployState).To(Equal(true))
					Expect(argPreDeleteState).To(Equal("default"))

					format, content := logger.PrintfArgsForCall(0)
					Expect(fmt.Sprintf(format, content...)).To(Equal("configuring product..."))

					format, content = logger.PrintfArgsForCall(1)
					Expect(fmt.Sprintf(format, content...)).To(Equal("applying errand configuration for the following errands:"))

					format, content = logger.PrintfArgsForCall(2)
					Expect(fmt.Sprintf(format, content...)).To(Equal("\tpush-usage-service"))

					format, content = logger.PrintfArgsForCall(3)
					Expect(fmt.Sprintf(format, content...)).To(Equal("\tsmoke_tests"))

					format, content = logger.PrintfArgsForCall(4)
					Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring product"))
				})
			})
		})

		Context("when the instance count is not an int", func() {
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
					"--product-name", "cf",
					"--product-resources", automaticResourceConfig,
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
					"--product-name", "cf",
					"--product-resources", resourceConfig,
				})

				Expect(err).To(MatchError("could not fetch existing job configuration: some error"))
			})
		})

		Context("when neither the product-properties, product-network or product-resources flag is provided", func() {
			It("logs and then does nothing", func() {
				command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
				err := command.Execute([]string{"--product-name", "cf"})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.ListStagedProductsCallCount()).To(Equal(0))

				format, content := logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("Provided properties are empty, nothing to do here"))
			})
		})

		Context("when an error occurs", func() {
			Context("when the product does not exist", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)

					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
						},
					}, nil)

					err := command.Execute([]string{
						"--product-name", "cf",
						"--product-properties", productProperties,
					})
					Expect(err).To(MatchError(`could not find product "cf"`))
				})
			})

			Context("when the product resources cannot be decoded", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "cf"},
						},
					}, nil)

					err := command.Execute([]string{"--product-name", "cf", "--product-resources", "%%%%%"})
					Expect(err).To(MatchError(ContainSubstring("could not decode product-resource json")))
				})
			})

			Context("when the jobs cannot be fetched", func() {
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

					err := command.Execute([]string{"--product-name", "cf", "--product-resources", resourceConfig})
					Expect(err).To(MatchError("failed to fetch jobs: boom"))
				})
			})

			Context("when resources fail to configure", func() {
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

					err := command.Execute([]string{"--product-name", "cf", "--product-resources", resourceConfig})
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

			Context("when the --product-name flag is missing", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
					err := command.Execute([]string{})
					Expect(err).To(MatchError("could not parse configure-product flags: missing required flag \"--product-name\""))
				})
			})

			Context("when the --config flag is passed", func() {
				Context("when the config flag is passed with the product-properties, product-network or product-resources flag", func() {
					It("returns an error", func() {
						file, err := ioutil.TempFile("", "")
						Expect(err).NotTo(HaveOccurred())
						command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
						service.ListStagedProductsReturns(api.StagedProductsOutput{
							Products: []api.StagedProduct{
								{GUID: "some-product-guid", Type: "cf"},
							},
						}, nil)
						err = command.Execute([]string{"--product-name", "cf", "--product-resources", resourceConfig, "--config", file.Name()})
						Expect(err).To(MatchError("config flag can not be passed with the product-properties, product-network or product-resources flag"))
					})
				})

				Context("when the provided config path does not exist", func() {
					It("returns an error", func() {
						command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
						service.ListStagedProductsReturns(api.StagedProductsOutput{
							Products: []api.StagedProduct{
								{GUID: "some-product-guid", Type: "cf"},
							},
						}, nil)
						err := command.Execute([]string{"--product-name", "cf", "--config", "some/non-existant/path.yml"})
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

						err = client.Execute([]string{"--product-name", "cf", "--config", configFile.Name()})
						Expect(err).To(MatchError(ContainSubstring("could not be parsed as valid configuration")))

						os.RemoveAll(configFile.Name())
					})
				})
			})

			Context("when the properties cannot be configured", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
					service.UpdateStagedProductPropertiesReturns(errors.New("some product error"))

					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "some-product"},
						},
					}, nil)

					err := command.Execute([]string{"--product-name", "some-product", "--product-properties", "{}", "--product-network", "anything"})
					Expect(err).To(MatchError("failed to configure product: some product error"))
				})
			})

			Context("when the networks cannot be configured", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(func() []string { return nil }, service, logger)
					service.UpdateStagedProductNetworksAndAZsReturns(errors.New("some product error"))

					service.ListStagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "some-product"},
						},
					}, nil)

					err := command.Execute([]string{"--product-name", "some-product", "--product-properties", "{}", "--product-network", "anything"})
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
						"--product-name", "cf",
						"--config", configFile.Name(),
					})
					Expect(err).To(MatchError("failed to set errand state for errand push-usage-service: error configuring errand"))
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
