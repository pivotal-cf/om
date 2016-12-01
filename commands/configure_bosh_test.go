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

var _ = Describe("ConfigureBosh", func() {
	var (
		service *fakes.BoshFormService
		logger  *fakes.Logger
	)

	BeforeEach(func() {
		service = &fakes.BoshFormService{}
		logger = &fakes.Logger{}
	})

	Describe("Execute", func() {
		It("configures the bosh", func() {
			command := commands.NewConfigureBosh(service, logger)

			service.GetFormReturns(api.Form{
				Action:            "form-action",
				AuthenticityToken: "some-auth-token",
				RailsMethod:       "the-rails",
			}, nil)

			err := command.Execute([]string{
				"--iaas-configuration",
				`{
				"project": "some-project",
				"default_deployment_tag": "my-vms",
				"auth_json": "{\"service_account_key\": \"some-service-key\",\"private_key\": \"some-key\"}"
			}`})
			Expect(err).NotTo(HaveOccurred())

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuring iaas specific options for bosh tile"))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring bosh tile"))

			Expect(service.GetFormArgsForCall(0)).To(Equal("/infrastructure/iaas_configuration/edit"))

			Expect(service.ConfigureIAASArgsForCall(0)).To(Equal(api.ConfigureIAASInput{
				Form: api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				},
				EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&iaas_configuration%5Bauth_json%5D=%7B%22service_account_key%22%3A+%22some-service-key%22%2C%22private_key%22%3A+%22some-key%22%7D&iaas_configuration%5Bdefault_deployment_tag%5D=my-vms&iaas_configuration%5Bproject%5D=some-project",
			}))
		})

		Context("error cases", func() {
			Context("when an invalid flag is passed", func() {
				It("returns an error", func() {
					command := commands.NewConfigureBosh(service, logger)

					err := command.Execute([]string{"--not-a-real-flag"})
					Expect(err).To(MatchError("flag provided but not defined: -not-a-real-flag"))
				})
			})

			Context("when the form can't be fetched", func() {
				It("returns an error", func() {
					service.GetFormReturns(api.Form{}, errors.New("meow meow meow"))

					command := commands.NewConfigureBosh(service, logger)

					err := command.Execute([]string{""})
					Expect(err).To(MatchError("could not fetch form: meow meow meow"))
				})
			})

			Context("when the json can't be decoded", func() {
				It("returns an error", func() {
					command := commands.NewConfigureBosh(service, logger)

					err := command.Execute([]string{"--iaas-configuration", "%$#@#$"})
					Expect(err).To(MatchError("could not decode json: invalid character '%' looking for beginning of value"))
				})
			})

			Context("when configuring the tile fails", func() {
				It("returns an error", func() {
					service.ConfigureIAASReturns(errors.New("NOPE"))

					command := commands.NewConfigureBosh(service, logger)

					err := command.Execute([]string{"--iaas-configuration", "{}"})
					Expect(err).To(MatchError("tile failed to configure: NOPE"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns the usage for the command", func() {
			command := commands.NewConfigureBosh(nil, nil)

			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "configures the bosh director that is deployed by the Ops Manager",
				ShortDescription: "configures Ops Manager deployed bosh director",
				Flags:            command.Options,
			}))
		})
	})
})
