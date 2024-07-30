package commands_test

import (
	"errors"
	"log"

	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DisableProductVerifiers", func() {
	var (
		presenter *presenterfakes.FormattedPresenter
		service   *fakes.DisableProductVerifiersService
		command   *commands.DisableProductVerifiers
		stderr    *gbytes.Buffer
		logger    *log.Logger
	)

	BeforeEach(func() {
		presenter = &presenterfakes.FormattedPresenter{}
		service = &fakes.DisableProductVerifiersService{}
		stderr = gbytes.NewBuffer()
		logger = log.New(stderr, "", 0)
		command = commands.NewDisableProductVerifiers(presenter, service, logger)
	})

	When("all provided verifiers exist", func() {
		It("disables all the provided verifiers", func() {
			verifierType1 := "some-verifier-type"
			verifierType2 := "another-verifier-type"

			service.ListProductVerifiersReturns([]api.Verifier{
				{
					Type:    verifierType1,
					Enabled: true,
				},
				{
					Type:    verifierType2,
					Enabled: false,
				},
			}, "cf-guid", nil)
			service.DisableProductVerifiersReturns(nil)

			err := executeCommand(command, []string{"--product-name", "cf", "--type", verifierType1, "-t", verifierType2})
			Expect(err).ToNot(HaveOccurred())

			Expect(service.ListProductVerifiersCallCount()).To(Equal(1))
			Expect(service.ListProductVerifiersArgsForCall(0)).To(Equal("cf"))
			Expect(service.DisableProductVerifiersCallCount()).To(Equal(1))
			verifierTypes, guid := service.DisableProductVerifiersArgsForCall(0)
			Expect(guid).To(Equal("cf-guid"))
			Expect(verifierTypes).To(Equal([]string{verifierType1, verifierType2}))
		})
	})

	When("listing the available verifiers fails", func() {
		It("returns an error", func() {
			service.ListProductVerifiersReturns(nil, "", errors.New("some error occurred"))

			err := executeCommand(command, []string{"--product-name", "cf", "--type", "failing-verifier-type"})
			Expect(err).To(MatchError("could not get available verifiers from Ops Manager: some error occurred"))
		})
	})

	When("disabling verifiers fails", func() {
		It("returns an error", func() {
			service.ListProductVerifiersReturns([]api.Verifier{
				{
					Type:    "some-verifier-type",
					Enabled: true,
				},
			}, "cf-guid", nil)

			service.DisableProductVerifiersReturns(errors.New("some error occurred"))

			err := executeCommand(command, []string{"--product-name", "cf", "--type", "some-verifier-type"})
			Expect(err).To(MatchError("could not disable verifiers in Ops Manager: some error occurred"))
		})
	})

	When("some of the provided verifiers don't exist", func() {
		It("returns a list of the verifiers that weren't found", func() {
			service.ListProductVerifiersReturns([]api.Verifier{{
				Type:    "some-verifier-type",
				Enabled: true,
			}}, "cf-guid", nil)

			err := executeCommand(command, []string{"--product-name", "cf", "--type", "missing-verifier-type", "-t", "another-missing-verifier-type"})
			Expect(err).To(MatchError(ContainSubstring("verifier does not exist for product")))

			Expect(service.DisableProductVerifiersCallCount()).To(Equal(0))

			Expect(string(stderr.Contents())).To(ContainSubstring("The following verifiers do not exist for cf:"))
			Expect(string(stderr.Contents())).To(ContainSubstring("- missing-verifier-type"))
			Expect(string(stderr.Contents())).To(ContainSubstring("- another-missing-verifier-type"))
			Expect(string(stderr.Contents())).To(ContainSubstring("No changes were made."))
		})
	})
})
