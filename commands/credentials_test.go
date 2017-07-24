package commands_test

import (
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Credentials", func() {
	var (
		csService   *fakes.CredentialsService
		dpLister    *fakes.DeployedProductsLister
		tableWriter *fakes.TableWriter
		logger      *fakes.Logger
	)

	BeforeEach(func() {
		csService = &fakes.CredentialsService{}
		dpLister = &fakes.DeployedProductsLister{}
		tableWriter = &fakes.TableWriter{}
		logger = &fakes.Logger{}
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			dpLister.DeployedProductsReturns([]api.DeployedProductOutput{
				api.DeployedProductOutput{
					Type: "some-product",
					GUID: "other-deployed-product-guid",
				}}, nil)
		})

		It("fetches the credential alphabetically", func() {
			command := commands.NewCredentials(csService, dpLister, tableWriter, logger)

			csService.FetchReturns(api.CredentialOutput{
				Credential: api.Credential{
					Type: "simple_credentials",
					Value: map[string]string{
						"password": "some-password",
						"identity": "some-identity",
					},
				},
			}, nil)

			err := command.Execute([]string{
				"--product-name", "some-product",
				"--credential-reference", ".properties.some-credentials",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(tableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"identity", "password"}))

			Expect(tableWriter.AppendCallCount()).To(Equal(1))
			Expect(tableWriter.AppendArgsForCall(0)).To(Equal([]string{"some-identity", "some-password"}))

			Expect(tableWriter.RenderCallCount()).To(Equal(1))
		})

		Context("when the credential reference cannot be found", func() {
			BeforeEach(func() {
				csService.FetchReturns(api.CredentialOutput{}, nil)
			})

			It("returns an error", func() {
				command := commands.NewCredentials(csService, dpLister, tableWriter, logger)

				err := command.Execute([]string{
					"--product-name", "some-product",
					"--credential-reference", "some-credential",
				})
				Expect(err).To(MatchError(ContainSubstring("failed to fetch credential")))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewCredentials(nil, nil, nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This authenticated command fetches credentials for deployed products.",
				ShortDescription: "fetch credentials for a deployed product",
				Flags:            command.Options,
			}))
		})
	})
})
