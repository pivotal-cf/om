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

var _ = Describe("StagedProducts", func() {
	var (
		presenter         *presenterfakes.Presenter
		diagnosticService *fakes.DiagnosticService
		command           commands.StagedProducts
	)

	BeforeEach(func() {
		presenter = &presenterfakes.Presenter{}
		diagnosticService = &fakes.DiagnosticService{}
		command = commands.NewStagedProducts(presenter, diagnosticService)
	})

	It("lists the staged products", func() {
		stagedProducts := []api.DiagnosticProduct{
			{
				Name:    "some-product",
				Version: "some-version",
			},
			{
				Name:    "acme-product",
				Version: "version-infinity",
			},
		}
		diagnosticService.ReportReturns(api.DiagnosticReport{
			StagedProducts: stagedProducts,
		}, nil)

		err := command.Execute([]string{})
		Expect(err).NotTo(HaveOccurred())

		Expect(diagnosticService.ReportCallCount()).To(Equal(1))

		Expect(presenter.PresentStagedProductsCallCount()).To(Equal(1))
		Expect(presenter.PresentStagedProductsArgsForCall(0)).To(Equal(stagedProducts))
	})

	Context("failure cases", func() {
		Context("when fetching the diagnostic report fails", func() {
			It("returns an error", func() {
				diagnosticService.ReportReturns(api.DiagnosticReport{}, errors.New("beep boop"))

				err := command.Execute([]string{})
				Expect(err).To(MatchError("failed to retrieve staged products beep boop"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewStagedProducts(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command lists all staged products.",
				ShortDescription: "lists staged products",
			}))
		})
	})
})
