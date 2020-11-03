package commands_test

import (
	"errors"
	"github.com/pivotal-cf/om/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
)

var _ = Describe("Products", func() {
	var (
		presenter          *presenterfakes.FormattedPresenter
		fakeProductService *fakes.ProductService
		command            *commands.Products
	)

	BeforeEach(func() {
		presenter = &presenterfakes.FormattedPresenter{}
		fakeProductService = &fakes.ProductService{}
		command = commands.NewProducts(presenter, fakeProductService)
	})

	Describe("Execute", func() {
		var (
			stagedProducts    []api.DiagnosticProduct
			deployedProducts  []api.DiagnosticProduct
			availableProducts api.AvailableProductsOutput
		)

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

			deployedProducts = []api.DiagnosticProduct{
				{
					Name:    "some-product",
					Version: "some-version",
				},
				{
					Name:    "acme-product",
					Version: "another-version",
				},
			}

			availableProducts = api.AvailableProductsOutput{
				ProductsList: []api.ProductInfo{
					api.ProductInfo{
						Name:    "some-product",
						Version: "1.2.3",
					},
					api.ProductInfo{
						Name:    "acme-product",
						Version: "another-version",
					},
					api.ProductInfo{
						Name:    "acme-product",
						Version: "version-infinity",
					},
				},
			}

			fakeProductService.GetDiagnosticReportReturns(api.DiagnosticReport{
				StagedProducts:   stagedProducts,
				DeployedProducts: deployedProducts,
			}, nil)

			fakeProductService.ListAvailableProductsReturns(availableProducts, nil)
		})

		It("lists the products and their available, staged, and deployed versions", func() {
			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeProductService.GetDiagnosticReportCallCount()).To(Equal(1))
			Expect(fakeProductService.ListAvailableProductsCallCount()).To(Equal(1))

			products := models.ProductsVersionsDisplay{
				ProductVersions: []models.ProductVersions{{
					Name:      "some-product",
					Available: []string{"1.2.3"},
					Staged:    "some-version",
					Deployed:  "some-version",
				}, {
					Name:      "acme-product",
					Available: []string{"another-version", "version-infinity"},
					Staged:    "version-infinity",
					Deployed:  "another-version",
				}},
				Available: true,
				Staged:    true,
				Deployed:  true,
			}

			Expect(presenter.SetFormatArgsForCall(0)).To(Equal("table"))
			Expect(presenter.PresentProductsCallCount()).To(Equal(1))
			Expect(presenter.PresentProductsArgsForCall(0).Available).To(Equal(true))
			Expect(presenter.PresentProductsArgsForCall(0).Staged).To(Equal(true))
			Expect(presenter.PresentProductsArgsForCall(0).Deployed).To(Equal(true))
			Expect(presenter.PresentProductsArgsForCall(0).ProductVersions).To(ContainElements(products.ProductVersions))

		})

		When("the available flag is provided", func() {
			It("sets the available flag for the presenter", func() {
				err := executeCommand(command, []string{"--available"})
				Expect(err).ToNot(HaveOccurred())

				Expect(presenter.SetFormatArgsForCall(0)).To(Equal("table"))
				Expect(presenter.PresentProductsCallCount()).To(Equal(1))
				Expect(presenter.PresentProductsArgsForCall(0).Available).To(Equal(true))
				Expect(presenter.PresentProductsArgsForCall(0).Staged).To(Equal(false))
				Expect(presenter.PresentProductsArgsForCall(0).Deployed).To(Equal(false))
			})
		})

		When("the staged flag is provided", func() {
			It("sets the staged flag for the presenter", func() {
				err := executeCommand(command, []string{"--staged"})
				Expect(err).ToNot(HaveOccurred())

				Expect(presenter.SetFormatArgsForCall(0)).To(Equal("table"))
				Expect(presenter.PresentProductsCallCount()).To(Equal(1))
				Expect(presenter.PresentProductsArgsForCall(0).Available).To(Equal(false))
				Expect(presenter.PresentProductsArgsForCall(0).Staged).To(Equal(true))
				Expect(presenter.PresentProductsArgsForCall(0).Deployed).To(Equal(false))
			})
		})

		When("the deployed flag is provided", func() {
			It("sets the deployed flag for the presenter", func() {
				err := executeCommand(command, []string{"--deployed"})
				Expect(err).ToNot(HaveOccurred())

				Expect(presenter.SetFormatArgsForCall(0)).To(Equal("table"))
				Expect(presenter.PresentProductsCallCount()).To(Equal(1))
				Expect(presenter.PresentProductsArgsForCall(0).Available).To(Equal(false))
				Expect(presenter.PresentProductsArgsForCall(0).Staged).To(Equal(false))
				Expect(presenter.PresentProductsArgsForCall(0).Deployed).To(Equal(true))
			})
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
				fakeProductService.GetDiagnosticReportReturns(api.DiagnosticReport{}, errors.New("beep boop"))

				err := executeCommand(command, []string{})
				Expect(err).To(MatchError("failed to retrieve staged and deployed products beep boop"))
			})
		})

		When("fetching the available products fails", func() {
			It("returns an error", func() {
				fakeProductService.GetDiagnosticReportReturns(api.DiagnosticReport{}, nil)
				fakeProductService.ListAvailableProductsReturns(api.AvailableProductsOutput{}, errors.New("beep boop"))

				err := executeCommand(command, []string{})
				Expect(err).To(MatchError("failed to retrieve available products beep boop"))
			})
		})
	})
})
