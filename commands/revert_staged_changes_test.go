package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RevertStagedChanges", func() {
	var (
		service *fakes.DashboardService
		logger  *fakes.Logger
	)

	BeforeEach(func() {
		service = &fakes.DashboardService{}
		logger = &fakes.Logger{}
	})

	Describe("Execute", func() {
		It("reverts staged changes on the targeted OpsMan", func() {
			command := commands.NewRevertStagedChanges(service, logger)

			service.GetRevertFormReturns(api.Form{
				Action:            "/installation",
				AuthenticityToken: "some-auth-token",
				RailsMethod:       "the-rails",
			}, nil)

			err := command.Execute([]string{})

			Expect(err).NotTo(HaveOccurred())

			Expect(service.PostInstallFormArgsForCall(0)).To(Equal(api.PostFormInput{
				Form: api.Form{
					Action:            "/installation",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				},
				EncodedPayload: "_method=delete&authenticity_token=some-auth-token&commit=Confirm",
			}))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("reverting staged changes on the targeted Ops Manager"))
			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("done"))
		})

		Context("when there are no staged changes to revert", func() {
			It("returns without error", func() {
				command := commands.NewRevertStagedChanges(service, logger)
				service.GetRevertFormReturns(api.Form{}, nil)
				err := command.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())
				Expect(service.PostInstallFormCallCount()).To(Equal(0))
			})
		})

		Context("error cases", func() {
			Context("when the form can't be fetched", func() {
				It("returns an error", func() {
					service.GetRevertFormReturns(api.Form{}, errors.New("meow meow meow"))

					command := commands.NewRevertStagedChanges(service, logger)

					err := command.Execute([]string{""})
					Expect(err).To(MatchError("could not fetch form: meow meow meow"))
				})
			})

			Context("when the form can't be posted", func() {
				It("returns an error", func() {
					service.GetRevertFormReturns(api.Form{
						Action:            "/installation",
						AuthenticityToken: "some-auth-token",
						RailsMethod:       "the-rails",
					}, nil)

					service.PostInstallFormReturns(errors.New("meow meow meow"))

					command := commands.NewRevertStagedChanges(service, logger)

					err := command.Execute([]string{""})
					Expect(err).To(MatchError("failed to revert staged changes: meow meow meow"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns the usage for the command", func() {
			command := commands.NewRevertStagedChanges(nil, nil)

			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "reverts staged changes on the installation dashboard page in the target Ops Manager",
				ShortDescription: "reverts staged changes on the Ops Manager targeted",
			}))
		})
	})
})
