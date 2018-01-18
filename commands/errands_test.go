package commands_test

import (
	"errors"

	"github.com/pivotal-cf/jhanda"
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
		fakePresenter        *presenterfakes.Presenter
		stagedProductsFinder *fakes.StagedProductsFinder
		errandsService       *fakes.ErrandsService
		command              commands.Errands
	)

	BeforeEach(func() {
		fakePresenter = &presenterfakes.Presenter{}
		stagedProductsFinder = &fakes.StagedProductsFinder{}
		errandsService = &fakes.ErrandsService{}
		command = commands.NewErrands(fakePresenter, errandsService, stagedProductsFinder)
	})

	Describe("Execute", func() {
		It("lists the available products", func() {
			errandsService.ListReturns(api.ErrandsListOutput{
				Errands: []api.Errand{
					{Name: "first-errand", PostDeploy: "true"},
					{Name: "second-errand", PostDeploy: "false"},
					{Name: "third-errand", PreDelete: true},
					{Name: "will-not-appear", PreDelete: nil},
					{Name: "also-bad", PostDeploy: nil},
				},
			}, nil)

			stagedProductsFinder.FindReturns(api.StagedProductsFindOutput{
				Product: api.StagedProduct{
					Type: "some-product-name",
					GUID: "some-product-id",
				},
			}, nil)

			err := command.Execute([]string{"--product-name", "some-product-name"})
			Expect(err).NotTo(HaveOccurred())

			Expect(stagedProductsFinder.FindCallCount()).To(Equal(1))
			Expect(stagedProductsFinder.FindArgsForCall(0)).To(Equal("some-product-name"))

			Expect(errandsService.ListCallCount()).To(Equal(1))
			Expect(errandsService.ListArgsForCall(0)).To(Equal("some-product-id"))

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

		Context("failure cases", func() {
			Context("when an unknown flag is passed", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--unknown-flag"})
					Expect(err).To(MatchError("could not parse errands flags: flag provided but not defined: -unknown-flag"))
				})
			})

			Context("when the staged products finder fails", func() {
				It("returns an error", func() {
					stagedProductsFinder.FindReturns(api.StagedProductsFindOutput{}, errors.New("there was an error"))
					err := command.Execute([]string{"--product-name", "some-product"})
					Expect(err).To(MatchError("failed to find staged product \"some-product\": there was an error"))
				})
			})

			Context("when the errands service fails", func() {
				It("returns an error", func() {
					errandsService.ListReturns(api.ErrandsListOutput{}, errors.New("there was an error"))
					err := command.Execute([]string{"--product-name", "some-product"})
					Expect(err).To(MatchError("failed to list errands: there was an error"))
				})
			})

			Context("when the product name is missing", func() {
				It("returns an error", func() {
					err := command.Execute([]string{})
					Expect(err).To(MatchError("could not parse errands flags: missing required flag \"--product-name\""))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewErrands(nil, nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command lists all errands for a product.",
				ShortDescription: "list errands for a product",
				Flags:            command.Options,
			}))
		})
	})
})
