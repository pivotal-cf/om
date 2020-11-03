package commands_test

import (
	"errors"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DisableDirectorVerifiers", func() {
	var (
		presenter *presenterfakes.FormattedPresenter
		service   *fakes.DisableDirectorVerifiersService
		command   *commands.DisableDirectorVerifiers
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

			err := executeCommand(command, []string{"--type", verifierType1, "-t", verifierType2})
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

			err := executeCommand(command, []string{"--type", "failing-verifier-type"})
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

			err := executeCommand(command, []string{"--type", "some-verifier-type"})
			Expect(err).To(MatchError("could not disable verifiers in Ops Manager: some error occurred"))
		})
	})

	When("some of the provided verifiers don't exist", func() {
		It("returns a list of the verifiers that weren't found", func() {
			service.ListDirectorVerifiersReturns([]api.Verifier{{
				Type:    "some-verifier-type",
				Enabled: true,
			}}, nil)

			err := executeCommand(command, []string{"--type", "missing-verifier-type", "-t", "another-missing-verifier-type"})
			Expect(err).To(MatchError(ContainSubstring("verifier does not exist for director")))

			Expect(service.DisableDirectorVerifiersCallCount()).To(Equal(0))

			Expect(string(stderr.Contents())).To(ContainSubstring("The following verifiers do not exist:"))
			Expect(string(stderr.Contents())).To(ContainSubstring("- missing-verifier-type"))
			Expect(string(stderr.Contents())).To(ContainSubstring("- another-missing-verifier-type"))
			Expect(string(stderr.Contents())).To(ContainSubstring("No changes were made."))
		})
	})
})
