package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
)

var _ = Describe("DeployedProducts", func() {
	var (
		presenter   *presenterfakes.FormattedPresenter
		fakeService *fakes.DeployedProductsService
		command     commands.DeployedProducts
	)

	BeforeEach(func() {
		presenter = &presenterfakes.FormattedPresenter{}
		fakeService = &fakes.DeployedProductsService{}
		command = commands.NewDeployedProducts(presenter, fakeService)
	})

	Describe("Execute", func() {
		var deployedProducts []api.DiagnosticProduct

		BeforeEach(func() {
			deployedProducts = []api.DiagnosticProduct{
				{
					Name:    "nonsense-product",
					Version: "nonsense-number",
				},
				{
					Name:    "acme-product",
					Version: "googleplex",
				},
			}

			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{
				DeployedProducts: deployedProducts,
			}, nil)
		})

		It("lists the deployed products", func() {
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeService.GetDiagnosticReportCallCount()).To(Equal(1))

			Expect(presenter.SetFormatArgsForCall(0)).To(Equal("table"))

			Expect(presenter.PresentDeployedProductsCallCount()).To(Equal(1))
			Expect(presenter.PresentDeployedProductsArgsForCall(0)).To(Equal(deployedProducts))
		})

		Context("when the format flag is provided", func() {
			It("sets the format on the presenter", func() {
				err := command.Execute([]string{"--format", "json"})
				Expect(err).NotTo(HaveOccurred())
				Expect(presenter.SetFormatArgsForCall(0)).To(Equal("json"))
			})
		})

		Context("failure cases", func() {
			Context("when an unknown flag is passed", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--unknown-flag"})
					Expect(err).To(MatchError("could not parse deployed-products flags: flag provided but not defined: -unknown-flag"))
				})
			})

			Context("when fetching the diagnostic report fails", func() {
				It("returns an error", func() {
					fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{}, errors.New("beep boop"))

					err := command.Execute([]string{})
					Expect(err).To(MatchError("failed to retrieve deployed products beep boop"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewDeployedProducts(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command lists all deployed products.",
				ShortDescription: "lists deployed products",
			}))
		})
	})
})
