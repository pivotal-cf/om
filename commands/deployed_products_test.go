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
		presenter         *presenterfakes.Presenter
		diagnosticService *fakes.DiagnosticService
		command           commands.DeployedProducts
	)

	BeforeEach(func() {
		presenter = &presenterfakes.Presenter{}
		diagnosticService = &fakes.DiagnosticService{}
		command = commands.NewDeployedProducts(presenter, diagnosticService)
	})

	It("lists the deployed products", func() {
		deployedProducts := []api.DiagnosticProduct{
			{
				Name:    "nonsense-product",
				Version: "nonsense-number",
			},
			{
				Name:    "acme-product",
				Version: "googleplex",
			},
		}

		diagnosticService.ReportReturns(api.DiagnosticReport{
			DeployedProducts: deployedProducts,
		}, nil)

		err := command.Execute([]string{})
		Expect(err).NotTo(HaveOccurred())

		Expect(diagnosticService.ReportCallCount()).To(Equal(1))

		Expect(presenter.PresentDeployedProductsCallCount()).To(Equal(1))
		Expect(presenter.PresentDeployedProductsArgsForCall(0)).To(Equal(deployedProducts))
	})

	Context("failure cases", func() {
		Context("when fetching the diagnostic report fails", func() {
			It("returns an error", func() {
				diagnosticService.ReportReturns(api.DiagnosticReport{}, errors.New("beep boop"))

				err := command.Execute([]string{})
				Expect(err).To(MatchError("failed to retrieve deployed products beep boop"))
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
