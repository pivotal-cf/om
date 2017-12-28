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
		directorService *fakes.DirectorService
		command         commands.ConfigureDirector
		logger          *fakes.Logger
	)

	BeforeEach(func() {
		directorService = &fakes.DirectorService{}
		logger = &fakes.Logger{}
		command = commands.NewConfigureDirector(directorService, logger)
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
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(directorService.AZConfigurationCallCount()).To(Equal(1))
			Expect(directorService.AZConfigurationArgsForCall(0)).To(Equal(api.AZConfiguration{
				AvailabilityZones: json.RawMessage(`[{"some-az-assignment": "az"}]`),
			}))

			Expect(directorService.NetworksConfigurationCallCount()).To(Equal(1))
			Expect(directorService.NetworksConfigurationArgsForCall(0)).To(Equal(json.RawMessage(`{"network": "network-1"}`)))

			Expect(directorService.NetworkAndAZCallCount()).To(Equal(1))
			Expect(directorService.NetworkAndAZArgsForCall(0)).To(Equal(api.NetworkAndAZConfiguration{
				NetworkAZ: json.RawMessage(networkAssignmentJSON),
			}))

			Expect(directorService.PropertiesCallCount()).To(Equal(1))
			Expect(directorService.PropertiesArgsForCall(0)).To(Equal(api.DirectorProperties{
				DirectorConfiguration: json.RawMessage(`{"some-director-assignment": "director"}`),
				IAASConfiguration:     json.RawMessage(`{"some-iaas-assignment": "iaas"}`),
				SecurityConfiguration: json.RawMessage(`{"some-security-assignment": "security"}`),
				SyslogConfiguration:   json.RawMessage(`{"some-syslog-assignment": "syslog"}`),
			}))

			Expect(logger.PrintfCallCount()).To(Equal(8))
			Expect(logger.PrintfArgsForCall(0)).To(Equal("started configuring director options for bosh tile"))
			Expect(logger.PrintfArgsForCall(1)).To(Equal("finished configuring director options for bosh tile"))
			Expect(logger.PrintfArgsForCall(2)).To(Equal("started configuring availability zone options for bosh tile"))
			Expect(logger.PrintfArgsForCall(3)).To(Equal("finished configuring availability zone options for bosh tile"))
			Expect(logger.PrintfArgsForCall(4)).To(Equal("started configuring network options for bosh tile"))
			Expect(logger.PrintfArgsForCall(5)).To(Equal("finished configuring network options for bosh tile"))
			Expect(logger.PrintfArgsForCall(6)).To(Equal("started configuring network assignment options for bosh tile"))
			Expect(logger.PrintfArgsForCall(7)).To(Equal("finished configuring network assignment options for bosh tile"))
		})

		Context("when no director configuration flags are provided", func() {
			It("only calls the properties function once", func() {
				err := command.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())
				Expect(directorService.AZConfigurationCallCount()).To(Equal(0))
				Expect(directorService.NetworksConfigurationCallCount()).To(Equal(0))
				Expect(directorService.NetworkAndAZCallCount()).To(Equal(0))
				Expect(directorService.PropertiesCallCount()).To(Equal(1))
				Expect(directorService.PropertiesArgsForCall(0)).To(Equal(api.DirectorProperties{
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
					directorService.AZConfigurationReturns(errors.New("az endpoint failed"))
					err := command.Execute([]string{"--az-configuration", `{}`})
					Expect(err).To(MatchError("availability zones configuration could not be applied: az endpoint failed"))
				})
			})

			Context("when configuring networks fails", func() {
				It("returns an error", func() {
					directorService.NetworksConfigurationReturns(errors.New("networks endpoint failed"))
					err := command.Execute([]string{"--networks-configuration", `{}`})
					Expect(err).To(MatchError("networks configuration could not be applied: networks endpoint failed"))
				})
			})

			Context("when configuring networks fails", func() {
				It("returns an error", func() {
					directorService.NetworkAndAZReturns(errors.New("director service failed"))
					err := command.Execute([]string{"--network-assignment", `{}`})
					Expect(err).To(MatchError("network and AZs could not be applied: director service failed"))
				})
			})

			Context("when configuring properties fails", func() {
				It("returns an error", func() {
					directorService.PropertiesReturns(errors.New("properties end point failed"))
					err := command.Execute([]string{"--director-configuration", `{}`})
					Expect(err).To(MatchError("properties could not be applied: properties end point failed"))
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
