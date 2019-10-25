package commands_test

import (
	"errors"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"io/ioutil"
	"log"
	"os"

	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DisableDirectorVerifiers", func() {
	var (
		presenter *presenterfakes.FormattedPresenter
		service   *fakes.DisableDirectorVerifiersService
		command   commands.DisableDirectorVerifiers
		stderr    *gbytes.Buffer
		logger    *log.Logger
	)

	BeforeEach(func() {
		presenter = &presenterfakes.FormattedPresenter{}
		service = &fakes.DisableDirectorVerifiersService{}
		stderr = gbytes.NewBuffer()
		logger = log.New(stderr, "", 0)
		command = commands.NewDisableDirectorVerifiers(presenter, service, logger)
	})

	When("all provided verifiers exist", func() {
		It("disables all the provided verifiers", func() {
			verifierType1 := "some-verifier-type"
			verifierType2 := "another-verifier-type"

			service.ListDirectorVerifiersReturns([]api.Verifier{
				{
					Type:    verifierType1,
					Enabled: true,
				},
				{
					Type:    verifierType2,
					Enabled: false,
				},
			}, nil)
			service.DisableDirectorVerifiersReturns(nil)

			err := command.Execute([]string{"--type", verifierType1, "-t", verifierType2})
			Expect(err).ToNot(HaveOccurred())

			Expect(service.ListDirectorVerifiersCallCount()).To(Equal(1))
			Expect(service.DisableDirectorVerifiersCallCount()).To(Equal(1))

			verifierTypes := service.DisableDirectorVerifiersArgsForCall(0)
			Expect(verifierTypes).To(Equal([]string{verifierType1, verifierType2}))
		})
	})

	When("listing the available verifiers fails", func() {
		It("returns an error", func() {
			service.ListDirectorVerifiersReturns(nil, errors.New("some error occurred"))

			err := command.Execute([]string{"--type", "failing-verifier-type"})
			Expect(err).To(MatchError("could not get available verifiers from Ops Manager: some error occurred"))
		})
	})

	When("disabling verifiers fails", func() {
		It("returns an error", func() {
			service.ListDirectorVerifiersReturns([]api.Verifier{
				{
					Type:    "some-verifier-type",
					Enabled: true,
				},
			}, nil)

			service.DisableDirectorVerifiersReturns(errors.New("some error occurred"))

			err := command.Execute([]string{"--type", "some-verifier-type"})
			Expect(err).To(MatchError("could not disable verifiers in Ops Manager: some error occurred"))
		})
	})

	When("some of the provided verifiers don't exist", func() {
		It("returns a list of the verifiers that weren't found", func() {
			service.ListDirectorVerifiersReturns([]api.Verifier{{
				Type:    "some-verifier-type",
				Enabled: true,
			}}, nil)

			err := command.Execute([]string{"--type", "missing-verifier-type", "-t", "another-missing-verifier-type"})
			Expect(err).To(MatchError(ContainSubstring("verifier does not exist for director")))

			Expect(service.DisableDirectorVerifiersCallCount()).To(Equal(0))

			Expect(string(stderr.Contents())).To(ContainSubstring("The following verifiers do not exist:"))
			Expect(string(stderr.Contents())).To(ContainSubstring("- missing-verifier-type"))
			Expect(string(stderr.Contents())).To(ContainSubstring("- another-missing-verifier-type"))
			Expect(string(stderr.Contents())).To(ContainSubstring("No changes were made."))
		})
	})

	When("config file is provided", func() {
		var configFile *os.File

		BeforeEach(func() {
			var err error

			service.ListDirectorVerifiersReturns([]api.Verifier{
				{
					Type:    "some-verifier-type",
					Enabled: false,
				},
				{
					Type:    "another-verifier-type",
					Enabled: true,
				},
			}, nil)

			configFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.Remove(configFile.Name())
			Expect(err).ToNot(HaveOccurred())
		})

		It("reads configuration from config file", func() {
			configContent := `type: [ "some-verifier-type", "another-verifier-type" ]`

			_, err := configFile.WriteString(configContent)
			Expect(err).NotTo(HaveOccurred())

			err = command.Execute([]string{"--config", configFile.Name()})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.ListDirectorVerifiersCallCount()).To(Equal(1))
			Expect(service.DisableDirectorVerifiersCallCount()).To(Equal(1))
		})

		It("returns an error if the config file is malformed", func() {
			_, err := configFile.WriteString("malformed-yaml")
			Expect(err).NotTo(HaveOccurred())

			err = command.Execute([]string{"--config", configFile.Name()})
			Expect(err).To(HaveOccurred())
		})
	})

	When("flags are provided", func() {
		It("returns an error if an unknown flag is provided", func() {
			err := command.Execute([]string{"--badflag"})
			Expect(err).To(MatchError("could not parse disable-director-verifiers flags: flag provided but not defined: -badflag"))
		})

		It("returns an error if there is no --type provided", func() {
			err := command.Execute([]string{})
			Expect(err).To(MatchError(ContainSubstring(`missing required flag "--type"`)))
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewDisableDirectorVerifiers(nil, nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command disables director verifiers",
				ShortDescription: "disables director verifiers",
				Flags:            command.Options,
			}))
		})
	})
})
