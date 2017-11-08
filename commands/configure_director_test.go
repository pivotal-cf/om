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
		logger          *fakes.Logger
	)

	BeforeEach(func() {
		directorService = &fakes.DirectorService{}
		logger = &fakes.Logger{}
		command = commands.NewConfigureDirector(directorService, logger)
	})

	Describe("Execute", func() {

		It("configures the director", func() {
			err := command.Execute([]string{
				"--network-assignment", "some-network-assignment",
				"--director-configuration", "some-director-configuration",
				"--iaas-configuration", "some-iaas-configuration",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(directorService.NetworkAndAZCallCount()).To(Equal(1))
			Expect(directorService.NetworkAndAZArgsForCall(0)).To(Equal("some-network-assignment"))

			Expect(directorService.PropertiesCallCount()).To(Equal(2))
			Expect(directorService.PropertiesArgsForCall(0)).To(Equal("some-director-configuration"))
			Expect(directorService.PropertiesArgsForCall(1)).To(Equal("some-iaas-configuration"))
		})

		Context("when the iaas-configuration flag is not provided", func() {
			It("only calls the properties function once", func() {
				err := command.Execute([]string{
					"--network-assignment", "some-network-assignment",
					"--director-configuration", "some-director-configuration",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(directorService.PropertiesCallCount()).To(Equal(1))
				Expect(directorService.PropertiesArgsForCall(0)).To(Equal("some-director-configuration"))
			})
		})

		Context("when the network-assignment flag is not provided", func() {
			It("calls the network_and_az function once", func() {
				err := command.Execute([]string{
					"--iaas-configuration", "some-iaas-configuration",
					"--director-configuration", "some-director-configuration",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(directorService.NetworkAndAZCallCount()).To(Equal(0))
			})
		})

		Context("when the director-configuration flag is not provided", func() {
			It("calls the properties function once", func() {
				err := command.Execute([]string{
					"--iaas-configuration", "some-iaas-configuration",
					"--network-assignment", "some-network-assignment",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(directorService.PropertiesCallCount()).To(Equal(1))
				Expect(directorService.PropertiesArgsForCall(0)).To(Equal("some-iaas-configuration"))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the flag parser fails", func() {
				err := command.Execute([]string{"--foo", "bar"})
				Expect(err).To(MatchError("could not parse configure-director flags: flag provided but not defined: -foo"))
			})

			It("returns an error when the director service fails", func() {
				directorService.NetworkAndAZReturns(errors.New("director service failed"))
				err := command.Execute([]string{"--network-assignment", `{}`})
				Expect(err).To(MatchError("network and AZs couldn't be applied: director service failed"))
			})

			It("returns an error when the properties end point fails", func() {
				directorService.PropertiesReturns(errors.New("properties end point failed"))
				err := command.Execute([]string{"--director-configuration", `{}`})
				Expect(err).To(MatchError("properties couldn't be applied: properties end point failed"))
			})

			It("returns an error when the iaas configuration end point fails", func() {
				directorService.PropertiesReturns(errors.New("iaas configuration end point failed"))
				err := command.Execute([]string{"--iaas-configuration", `{}`})
				Expect(err).To(MatchError("iaas configuration couldn't be applied: iaas configuration end point failed"))
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
