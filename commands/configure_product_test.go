package commands_test

import (
	"errors"
	"fmt"

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
	  "instances": 1,
		"persistent_disk": { "size_mb": "20480" },
    "instance_type": { "id": "m1.medium" }
  }
}`

var _ = Describe("ConfigureProduct", func() {
	Describe("Execute", func() {
		var (
			productsService *fakes.ProductConfigurer
			jobsService     *fakes.JobsConfigurer
			logger          *fakes.Logger
		)

		BeforeEach(func() {
			productsService = &fakes.ProductConfigurer{}
			jobsService = &fakes.JobsConfigurer{}
			logger = &fakes.Logger{}
		})

		It("configures a product's properties", func() {
			client := commands.NewConfigureProduct(productsService, jobsService, logger)

			productsService.StagedProductsReturns(api.StagedProductsOutput{
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

			Expect(jobsService.JobsArgsForCall(0)).To(Equal("some-product-guid"))
			Expect(jobsService.ConfigureArgsForCall(0)).To(Equal(api.JobConfigurationInput{
				ProductGUID: "some-product-guid",
				Jobs:        api.JobConfig{},
			}))

			Expect(productsService.StagedProductsCallCount()).To(Equal(1))
			Expect(productsService.ConfigureArgsForCall(0)).To(Equal(api.ProductsConfigurationInput{
				GUID:          "some-product-guid",
				Configuration: productProperties,
			}))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("setting properties"))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("finished setting properties"))
		})

		It("configures a product's network", func() {
			client := commands.NewConfigureProduct(productsService, jobsService, logger)

			productsService.StagedProductsReturns(api.StagedProductsOutput{
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

			Expect(jobsService.JobsArgsForCall(0)).To(Equal("some-product-guid"))
			Expect(jobsService.ConfigureArgsForCall(0)).To(Equal(api.JobConfigurationInput{
				ProductGUID: "some-product-guid",
				Jobs:        api.JobConfig{},
			}))

			Expect(productsService.StagedProductsCallCount()).To(Equal(1))
			Expect(productsService.ConfigureArgsForCall(0)).To(Equal(api.ProductsConfigurationInput{
				GUID:    "some-product-guid",
				Network: networkProperties,
			}))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("setting properties"))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("finished setting properties"))
		})

		It("configures the resource that is provided", func() {
			client := commands.NewConfigureProduct(productsService, jobsService, logger)
			productsService.StagedProductsReturns(api.StagedProductsOutput{
				Products: []api.StagedProduct{
					{GUID: "some-product-guid", Type: "cf"},
					{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
				},
			}, nil)

			jobsService.JobsReturns(api.JobsOutput{
				Jobs: []api.Job{
					{Name: "some-job", GUID: "a-guid"},
					{Name: "some-other-job", GUID: "a-different-guid"},
					{Name: "bad", GUID: "do-not-use"},
				},
			}, nil)

			err := client.Execute([]string{
				"--product-name", "cf",
				"--product-resources", resourceConfig,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(productsService.StagedProductsCallCount()).To(Equal(1))
			Expect(jobsService.JobsArgsForCall(0)).To(Equal("some-product-guid"))
			Expect(jobsService.ConfigureArgsForCall(0)).To(Equal(api.JobConfigurationInput{
				ProductGUID: "some-product-guid",
				Jobs: api.JobConfig{
					"a-guid": api.JobProperties{
						Instances:         1,
						PersistentDisk:    api.Disk{Size: "20480"},
						InstanceType:      api.InstanceType{ID: "m1.medium"},
						InternetConnected: true,
						LBNames:           []string{"some-lb"},
					},
					"a-different-guid": api.JobProperties{
						Instances:         1,
						PersistentDisk:    api.Disk{Size: "20480"},
						InstanceType:      api.InstanceType{ID: "m1.medium"},
						InternetConnected: false,
						LBNames:           nil,
					},
				},
			}))
		})

		Context("when neither the product-properties or product-network flag is provided", func() {
			It("logs and then does nothing", func() {
				command := commands.NewConfigureProduct(productsService, jobsService, logger)
				err := command.Execute([]string{"--product-name", "cf"})
				Expect(err).NotTo(HaveOccurred())

				Expect(productsService.StagedProductsCallCount()).To(Equal(0))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("Provided properties are empty, nothing to do here"))
			})
		})

		Context("when an error occurs", func() {
			Context("when the product resources cannot be decoded", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(productsService, jobsService, logger)
					productsService.StagedProductsReturns(api.StagedProductsOutput{
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
					command := commands.NewConfigureProduct(productsService, jobsService, logger)
					productsService.StagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "cf"},
						},
					}, nil)

					jobsService.JobsReturns(api.JobsOutput{
						Jobs: []api.Job{
							{Name: "some-job", GUID: "a-guid"},
						},
					}, errors.New("boom"))

					err := command.Execute([]string{"--product-name", "cf", "--product-resources", resourceConfig})
					Expect(err).To(MatchError("failed to fetch jobs: boom"))
				})
			})

			Context("when resources fail to configure", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(productsService, jobsService, logger)
					productsService.StagedProductsReturns(api.StagedProductsOutput{
						Products: []api.StagedProduct{
							{GUID: "some-product-guid", Type: "cf"},
						},
					}, nil)

					jobsService.JobsReturns(api.JobsOutput{
						Jobs: []api.Job{
							{Name: "some-job", GUID: "a-guid"},
						},
					}, nil)

					jobsService.ConfigureReturns(errors.New("bad things happened"))

					err := command.Execute([]string{"--product-name", "cf", "--product-resources", resourceConfig})
					Expect(err).To(MatchError("failed to configure resources: bad things happened"))
				})
			})

			Context("when an unknown flag is provided", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(productsService, jobsService, logger)
					err := command.Execute([]string{"--badflag"})
					Expect(err).To(MatchError("could not parse configure-product flags: flag provided but not defined: -badflag"))
				})
			})

			Context("when the product cannot be configured", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(productsService, jobsService, logger)
					productsService.ConfigureReturns(errors.New("some product error"))

					err := command.Execute([]string{"--product-name", "some-product", "--product-properties", "{}", "--product-network", "anything"})
					Expect(err).To(MatchError("failed to configure product: some product error"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewConfigureProduct(nil, nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This authenticated command configures a staged product",
				ShortDescription: "configures a staged product",
				Flags:            command.Options,
			}))
		})
	})
})
