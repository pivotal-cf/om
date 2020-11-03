package commands_test

import (
	"errors"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/models"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errands", func() {
	var (
		fakePresenter *presenterfakes.FormattedPresenter
		fakeService   *fakes.ErrandsService
		command       *commands.Errands
	)

	BeforeEach(func() {
		fakePresenter = &presenterfakes.FormattedPresenter{}
		fakeService = &fakes.ErrandsService{}
		command = commands.NewErrands(fakePresenter, fakeService)
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			fakeService.ListStagedProductErrandsReturns(api.ErrandsListOutput{
				Errands: []api.Errand{
					{Name: "first-errand", PostDeploy: "true"},
					{Name: "second-errand", PostDeploy: "false"},
					{Name: "third-errand", PreDelete: true},
					{Name: "will-not-appear", PreDelete: nil},
					{Name: "also-bad", PostDeploy: nil},
				},
			}, nil)

			fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{
				Product: api.StagedProduct{
					Type: "some-product-name",
					GUID: "some-product-id",
				},
			}, nil)
		})

		It("lists the available products", func() {
			err := executeCommand(command, []string{"--product-name", "some-product-name"})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.GetStagedProductByNameCallCount()).To(Equal(1))
			Expect(fakeService.GetStagedProductByNameArgsForCall(0)).To(Equal("some-product-name"))

			Expect(fakeService.ListStagedProductErrandsCallCount()).To(Equal(1))
			Expect(fakeService.ListStagedProductErrandsArgsForCall(0)).To(Equal("some-product-id"))

			Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("table"))
			Expect(fakePresenter.PresentErrandsCallCount()).To(Equal(1))
			errands := fakePresenter.PresentErrandsArgsForCall(0)
			Expect(errands).To(ConsistOf(
				models.Errand{Name: "first-errand", PostDeployEnabled: "true"},
				models.Errand{Name: "second-errand", PostDeployEnabled: "false"},
				models.Errand{Name: "third-errand", PreDeleteEnabled: "true"},
				models.Errand{Name: "will-not-appear"},
				models.Errand{Name: "also-bad"},
			))
		})

		When("the format flag is provided", func() {
			It("sets the format on the presenter", func() {
				err := executeCommand(command, []string{
					"--product-name", "some-product-name",
					"--format", "json",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
				Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("json"))
			})
		})

		When("the staged products finder fails", func() {
			It("returns an error", func() {
				fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{}, errors.New("there was an error"))
				err := executeCommand(command, []string{"--product-name", "some-product"})
				Expect(err).To(MatchError("failed to find staged product \"some-product\": there was an error"))
			})
		})

		When("the errands service fails", func() {
			It("returns an error", func() {
				fakeService.ListStagedProductErrandsReturns(api.ErrandsListOutput{}, errors.New("there was an error"))
				err := executeCommand(command, []string{"--product-name", "some-product"})
				Expect(err).To(MatchError("failed to list errands: there was an error"))
			})
		})

		When("the product name is missing", func() {
			It("returns an error", func() {
				err := executeCommand(command, []string{})
				Expect(err).To(MatchError("could not parse errands flags: missing required flag \"--product-name\""))
			})
		})
	})
})
