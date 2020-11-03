package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
)

var _ = Describe("StagedProducts", func() {
	var (
		presenter   *presenterfakes.FormattedPresenter
		fakeService *fakes.StagedProductsService
		command     *commands.StagedProducts
	)

	BeforeEach(func() {
		presenter = &presenterfakes.FormattedPresenter{}
		fakeService = &fakes.StagedProductsService{}
		command = commands.NewStagedProducts(presenter, fakeService)
	})

	Describe("Execute", func() {
		var stagedProducts []api.DiagnosticProduct

		BeforeEach(func() {
			stagedProducts = []api.DiagnosticProduct{
				{
					Name:    "some-product",
					Version: "some-version",
				},
				{
					Name:    "acme-product",
					Version: "version-infinity",
				},
			}
			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{
				StagedProducts: stagedProducts,
			}, nil)
		})

		It("lists the staged products", func() {
			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.GetDiagnosticReportCallCount()).To(Equal(1))

			Expect(presenter.SetFormatArgsForCall(0)).To(Equal("table"))
			Expect(presenter.PresentStagedProductsCallCount()).To(Equal(1))
			Expect(presenter.PresentStagedProductsArgsForCall(0)).To(Equal(stagedProducts))
		})

		When("the format flag is provided", func() {
			It("sets the format on the presenter", func() {
				err := executeCommand(command, []string{"--format", "json"})
				Expect(err).ToNot(HaveOccurred())

				Expect(presenter.SetFormatCallCount()).To(Equal(1))
				Expect(presenter.SetFormatArgsForCall(0)).To(Equal("json"))
			})
		})

		When("fetching the diagnostic report fails", func() {
			It("returns an error", func() {
				fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{}, errors.New("beep boop"))

				err := executeCommand(command, []string{})
				Expect(err).To(MatchError("failed to retrieve staged products beep boop"))
			})
		})
	})
})
