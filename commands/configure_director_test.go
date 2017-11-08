package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("ConfigureDirector", func() {
	var (
		directorService *fakes.DirectorService
		command         commands.ConfigureDirector
	)

	BeforeEach(func() {
		directorService = &fakes.DirectorService{}
		command = commands.NewConfigureDirector(directorService)
	})

	Describe("Execute", func() {
		It("configures the director with both network-assignment and director-configuration", func() {
			err := command.Execute([]string{"--network-assignment",
				`{"network_and_az": {"network": { "name": "network_name"},"singleton_availability_zone": {"name": "availability_zone_name"}}}`, "--director-configuration", `{
				"director_configuration": {
					"ntp_servers_string": "us.example.org, time.something.com",
					"resurrector_enabled": false,
					"director_hostname": "foo.example.com",
					"max_threads": 5
				}
			 }`})
			Expect(err).NotTo(HaveOccurred())
			Expect(directorService.NetworkAndAZCallCount()).To(Equal(1))

			jsonBody := directorService.NetworkAndAZArgsForCall(0)
			Expect(jsonBody).To(Equal(`{"network_and_az": {"network": { "name": "network_name"},"singleton_availability_zone": {"name": "availability_zone_name"}}}`))

			Expect(directorService.PropertiesCallCount()).To(Equal(1))

			jsonBody = directorService.PropertiesArgsForCall(0)
			Expect(jsonBody).To(Equal(`{
				"director_configuration": {
					"ntp_servers_string": "us.example.org, time.something.com",
					"resurrector_enabled": false,
					"director_hostname": "foo.example.com",
					"max_threads": 5
				}
			 }`))
		})

		It("configures the director without director-configuration properties", func() {
			err := command.Execute([]string{"--network-assignment",
				`{"network_and_az": {"network": { "name": "network_name"},"singleton_availability_zone": {"name": "availability_zone_name"}}}`})
			Expect(err).NotTo(HaveOccurred())
			Expect(directorService.NetworkAndAZCallCount()).To(Equal(1))

			jsonBody := directorService.NetworkAndAZArgsForCall(0)
			Expect(jsonBody).To(Equal(`{"network_and_az": {"network": { "name": "network_name"},"singleton_availability_zone": {"name": "availability_zone_name"}}}`))

			Expect(directorService.PropertiesCallCount()).To(Equal(0))
		})

		It("configures the director without network assignment", func() {
			err := command.Execute([]string{"--director-configuration", `{
				"director_configuration": {
					"ntp_servers_string": "us.example.org, time.something.com",
					"resurrector_enabled": false,
					"director_hostname": "foo.example.com",
					"max_threads": 5
				}
			 }`})
			Expect(err).NotTo(HaveOccurred())
			Expect(directorService.PropertiesCallCount()).To(Equal(1))

			jsonBody := directorService.PropertiesArgsForCall(0)
			Expect(jsonBody).To(Equal(`{
				"director_configuration": {
					"ntp_servers_string": "us.example.org, time.something.com",
					"resurrector_enabled": false,
					"director_hostname": "foo.example.com",
					"max_threads": 5
				}
			 }`))

			Expect(directorService.NetworkAndAZCallCount()).To(Equal(0))
		})

		Context("failure cases", func() {
			It("returns an error when the flag parser fails", func() {
				err := command.Execute([]string{"--foo", "bar"})
				Expect(err).To(MatchError("could not parse configure-director flags: flag provided but not defined: -foo"))
			})

			It("returns an error when the director service fails", func() {
				directorService.NetworkAndAZReturns(errors.New("director service failed"))
				err := command.Execute([]string{"--network-assignment",
					`{"network_and_az": {"network": { "name": "network_name"},"singleton_availability_zone": {"name": "availability_zone_name"}}}`})
				Expect(err).To(MatchError("network and AZs couldn't be applied: director service failed"))
			})

			It("returns an error when the properties end point fails", func() {
				directorService.PropertiesReturns(errors.New("properties end point failed"))
				err := command.Execute([]string{"--director-configuration", `{}`})
				Expect(err).To(MatchError("properties couldn't be applied: properties end point failed"))
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
