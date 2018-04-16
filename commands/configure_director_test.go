package commands_test

import (
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("ConfigureDirector", func() {
	var (
		directorService       *fakes.DirectorService
		jobsService           *fakes.JobsService
		stagedProductsService *fakes.StagedProductsService
		command               commands.ConfigureDirector
		logger                *fakes.Logger
	)

	BeforeEach(func() {
		directorService = &fakes.DirectorService{}
		jobsService = &fakes.JobsService{}
		stagedProductsService = &fakes.StagedProductsService{}
		logger = &fakes.Logger{}
		stagedProductsService.FindReturns(api.StagedProductsFindOutput{
			Product: api.StagedProduct{
				GUID: "p-bosh-guid",
			},
		}, nil)
		jobsService.ListStagedProductJobsReturns(map[string]string{
			"resource": "some-resource-guid",
		}, nil)
		jobsService.GetStagedProductJobResourceConfigReturns(api.JobProperties{
			InstanceType: api.InstanceType{
				ID: "automatic",
			},
			FloatingIPs: "1.2.3.4",
		}, nil)

		command = commands.NewConfigureDirector(directorService, jobsService, stagedProductsService, logger)
	})

	Describe("Execute", func() {
		It("configures the director", func() {
			networkAssignmentJSON := `{
				"network": {"name": "network"},
				"singleton_availability_zone": {"name": "singleton"}
			}`

			err := command.Execute([]string{
				"--network-assignment", networkAssignmentJSON,
				"--az-configuration", `[{"some-az-assignment": "az"}]`,
				"--networks-configuration", `{"network": "network-1"}`,
				"--director-configuration", `{"some-director-assignment": "director"}`,
				"--iaas-configuration", `{"some-iaas-assignment": "iaas"}`,
				"--security-configuration", `{"some-security-assignment": "security"}`,
				"--syslog-configuration", `{"some-syslog-assignment": "syslog"}`,
				"--resource-configuration", `{"resource": {"instance_type": {"id": "some-type"}}}`,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(directorService.UpdateStagedDirectorAvailabilityZonesCallCount()).To(Equal(1))
			Expect(directorService.UpdateStagedDirectorAvailabilityZonesArgsForCall(0)).To(Equal(api.AvailabilityZoneInput{
				AvailabilityZones: json.RawMessage(`[{"some-az-assignment": "az"}]`),
			}))

			Expect(directorService.UpdateStagedDirectorNetworksCallCount()).To(Equal(1))
			Expect(directorService.UpdateStagedDirectorNetworksArgsForCall(0)).To(Equal(json.RawMessage(`{"network": "network-1"}`)))

			Expect(directorService.UpdateStagedDirectorNetworkAndAZCallCount()).To(Equal(1))
			Expect(directorService.UpdateStagedDirectorNetworkAndAZArgsForCall(0)).To(Equal(api.NetworkAndAZConfiguration{
				NetworkAZ: json.RawMessage(networkAssignmentJSON),
			}))

			Expect(directorService.UpdateStagedDirectorPropertiesCallCount()).To(Equal(1))
			Expect(directorService.UpdateStagedDirectorPropertiesArgsForCall(0)).To(Equal(api.DirectorProperties{
				DirectorConfiguration: json.RawMessage(`{"some-director-assignment": "director"}`),
				IAASConfiguration:     json.RawMessage(`{"some-iaas-assignment": "iaas"}`),
				SecurityConfiguration: json.RawMessage(`{"some-security-assignment": "security"}`),
				SyslogConfiguration:   json.RawMessage(`{"some-syslog-assignment": "syslog"}`),
			}))

			Expect(stagedProductsService.FindCallCount()).To(Equal(1))
			Expect(stagedProductsService.FindArgsForCall(0)).To(Equal("p-bosh"))

			Expect(jobsService.ListStagedProductJobsCallCount()).To(Equal(1))
			Expect(jobsService.ListStagedProductJobsArgsForCall(0)).To(Equal("p-bosh-guid"))

			Expect(jobsService.GetStagedProductJobResourceConfigCallCount()).To(Equal(1))
			productGUID, instanceGroupGUID := jobsService.GetStagedProductJobResourceConfigArgsForCall(0)
			Expect(productGUID).To(Equal("p-bosh-guid"))
			Expect(instanceGroupGUID).To(Equal("some-resource-guid"))

			Expect(jobsService.UpdateStagedProductJobResourceConfigCallCount()).To(Equal(1))
			productGUID, instanceGroupGUID, jobConfiguration := jobsService.UpdateStagedProductJobResourceConfigArgsForCall(0)
			Expect(productGUID).To(Equal("p-bosh-guid"))
			Expect(instanceGroupGUID).To(Equal("some-resource-guid"))
			Expect(jobConfiguration).To(Equal(api.JobProperties{
				InstanceType: api.InstanceType{
					ID: "some-type",
				},
				FloatingIPs: "1.2.3.4",
			}))

			Expect(logger.PrintfCallCount()).To(Equal(12))
			Expect(logger.PrintfArgsForCall(0)).To(Equal("started configuring director options for bosh tile"))
			Expect(logger.PrintfArgsForCall(1)).To(Equal("finished configuring director options for bosh tile"))
			Expect(logger.PrintfArgsForCall(2)).To(Equal("started configuring availability zone options for bosh tile"))
			Expect(logger.PrintfArgsForCall(3)).To(Equal("finished configuring availability zone options for bosh tile"))
			Expect(logger.PrintfArgsForCall(4)).To(Equal("started configuring network options for bosh tile"))
			Expect(logger.PrintfArgsForCall(5)).To(Equal("finished configuring network options for bosh tile"))
			Expect(logger.PrintfArgsForCall(6)).To(Equal("started configuring network assignment options for bosh tile"))
			Expect(logger.PrintfArgsForCall(7)).To(Equal("finished configuring network assignment options for bosh tile"))
			Expect(logger.PrintfArgsForCall(8)).To(Equal("started configuring resource options for bosh tile"))
			Expect(logger.PrintfArgsForCall(9)).To(Equal("applying resource configuration for the following jobs:"))
			formatStr, formatArg := logger.PrintfArgsForCall(10)
			Expect([]interface{}{formatStr, formatArg}).To(Equal([]interface{}{"\t%s", []interface{}{"resource"}}))
			Expect(logger.PrintfArgsForCall(11)).To(Equal("finished configuring resource options for bosh tile"))
		})

		Context("when no director configuration flags are provided", func() {
			It("only calls the properties function once", func() {
				err := command.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())
				Expect(directorService.UpdateStagedDirectorAvailabilityZonesCallCount()).To(Equal(0))
				Expect(directorService.UpdateStagedDirectorNetworksCallCount()).To(Equal(0))
				Expect(directorService.UpdateStagedDirectorNetworkAndAZCallCount()).To(Equal(0))
				Expect(directorService.UpdateStagedDirectorPropertiesCallCount()).To(Equal(1))
				Expect(directorService.UpdateStagedDirectorPropertiesArgsForCall(0)).To(Equal(api.DirectorProperties{
					IAASConfiguration:     json.RawMessage(``),
					DirectorConfiguration: json.RawMessage(``),
					SecurityConfiguration: json.RawMessage(``),
					SyslogConfiguration:   json.RawMessage(``),
				}))
			})
		})

		Context("when an error occurs", func() {
			Context("when flag parser fails", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--foo", "bar"})
					Expect(err).To(MatchError("could not parse configure-director flags: flag provided but not defined: -foo"))
				})
			})

			Context("when configuring availability_zones fails", func() {
				It("returns an error", func() {
					directorService.UpdateStagedDirectorAvailabilityZonesReturns(errors.New("az endpoint failed"))
					err := command.Execute([]string{"--az-configuration", `{}`})
					Expect(err).To(MatchError("availability zones configuration could not be applied: az endpoint failed"))
				})
			})

			Context("when configuring networks fails", func() {
				It("returns an error", func() {
					directorService.UpdateStagedDirectorNetworksReturns(errors.New("networks endpoint failed"))
					err := command.Execute([]string{"--networks-configuration", `{}`})
					Expect(err).To(MatchError("networks configuration could not be applied: networks endpoint failed"))
				})
			})

			Context("when configuring networks fails", func() {
				It("returns an error", func() {
					directorService.UpdateStagedDirectorNetworkAndAZReturns(errors.New("director service failed"))
					err := command.Execute([]string{"--network-assignment", `{}`})
					Expect(err).To(MatchError("network and AZs could not be applied: director service failed"))
				})
			})

			Context("when configuring properties fails", func() {
				It("returns an error", func() {
					directorService.UpdateStagedDirectorPropertiesReturns(errors.New("properties end point failed"))
					err := command.Execute([]string{"--director-configuration", `{}`})
					Expect(err).To(MatchError("properties could not be applied: properties end point failed"))
				})
			})

			Context("when retrieving staged products fails", func() {
				It("returns an error", func() {
					stagedProductsService.FindReturns(api.StagedProductsFindOutput{}, errors.New("some-error"))
					err := command.Execute([]string{"--resource-configuration", `{}`})
					Expect(err).To(MatchError(ContainSubstring("some-error")))
				})
			})

			Context("when user-provided top-level resource config is not valid JSON", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--resource-configuration", `{{{`})
					Expect(err).To(MatchError(ContainSubstring("resource-configuration")))
				})
			})

			Context("when retrieving jobs for product fails", func() {
				It("returns an error", func() {
					jobsService.ListStagedProductJobsReturns(nil, errors.New("some-error"))
					err := command.Execute([]string{"--resource-configuration", `{}`})
					Expect(err).To(MatchError(ContainSubstring("some-error")))
				})
			})

			Context("when user-provided job does not exist", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--resource-configuration", `{"invalid-resource": {}}`})
					Expect(err).To(MatchError(ContainSubstring("invalid-resource")))
				})
			})

			Context("when retrieving existing job config fails", func() {
				It("returns an error", func() {
					jobsService.GetStagedProductJobResourceConfigReturns(api.JobProperties{}, errors.New("some-error"))
					err := command.Execute([]string{"--resource-configuration", `{"resource": {}}`})
					Expect(err).To(MatchError(ContainSubstring("some-error")))
				})
			})

			Context("when user-provided nested resource config is not valid JSON", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--resource-configuration", `{"resource": "%%%"}`})
					Expect(err).To(MatchError(ContainSubstring("resource-configuration")))
				})
			})

			Context("when configuring the job fails", func() {
				It("returns an error", func() {
					jobsService.UpdateStagedProductJobResourceConfigReturns(errors.New("some-error"))
					err := command.Execute([]string{"--resource-configuration", `{"resource": {}}`})
					Expect(err).To(MatchError(ContainSubstring("some-error")))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns the usage", func() {
			usage := command.Usage()

			Expect(usage.Description).To(Equal("This authenticated command configures the director."))
			Expect(usage.ShortDescription).To(Equal("configures the director"))
			Expect(usage.Flags).To(Equal(command.Options))
		})
	})
})
