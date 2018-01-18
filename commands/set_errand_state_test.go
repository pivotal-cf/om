package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("Set errand state", func() {
	var (
		stagedProductsFinder *fakes.StagedProductsFinder
		errandsService       *fakes.ErrandsService
		command              commands.SetErrandState
	)

	BeforeEach(func() {
		stagedProductsFinder = &fakes.StagedProductsFinder{}
		errandsService = &fakes.ErrandsService{}

		stagedProductsFinder.FindReturns(api.StagedProductsFindOutput{
			Product: api.StagedProduct{GUID: "some-product-guid", Type: "some-type"},
		}, nil)

		command = commands.NewSetErrandState(errandsService, stagedProductsFinder)
	})

	Describe("Execute", func() {
		It("set errand state for given errand in product", func() {
			err := command.Execute([]string{
				"--product-name", "some-product-name",
				"--errand-name", "some-errand",
				"--post-deploy-state", "enabled",
				"--pre-delete-state", "disabled",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(stagedProductsFinder.FindCallCount()).To(Equal(1))
			Expect(stagedProductsFinder.FindArgsForCall(0)).To(Equal("some-product-name"))

			Expect(errandsService.SetStateCallCount()).To(Equal(1))

			productGUID, errandName, postDeployState, preDeleteState := errandsService.SetStateArgsForCall(0)

			Expect(productGUID).To(Equal("some-product-guid"))
			Expect(errandName).To(Equal("some-errand"))
			Expect(postDeployState).To(BeTrue())
			Expect(preDeleteState).To(BeFalse())
		})

		DescribeTable("when user sets one state", func(postDeploy, preDelete string, desiredPostDeploy, desiredPreDelete interface{}) {
			err := command.Execute([]string{
				"--product-name", "some-product-name",
				"--errand-name", "some-errand",
				"--post-deploy-state", postDeploy,
				"--pre-delete-state", preDelete,
			})

			Expect(err).NotTo(HaveOccurred())

			Expect(errandsService.SetStateCallCount()).To(Equal(1))

			productGUID, errandName, postDeployState, preDeleteState := errandsService.SetStateArgsForCall(0)

			Expect(productGUID).To(Equal("some-product-guid"))
			Expect(errandName).To(Equal("some-errand"))

			if desiredPostDeploy != nil {
				Expect(postDeployState).To(Equal(desiredPostDeploy))
			} else {
				Expect(postDeployState).To(BeNil())
			}

			if desiredPreDelete != nil {
				Expect(preDeleteState).To(Equal(desiredPreDelete))
			} else {
				Expect(preDeleteState).To(BeNil())
			}
		},
			Entry("when only post deploy is given", "when-changed", "", "when-changed", nil),
			Entry("when only pre delete is given", "", "disabled", nil, false),
			Entry("when default states are desired", "default", "default", "default", "default"),
		)

		Context("failures", func() {
			Context("when invalid states have been given", func() {
				It("returns an error", func() {
					err := command.Execute([]string{
						"--product-name", "some-product-name",
						"--errand-name", "some-errand",
						"--post-deploy-state", "foo",
						"--pre-delete-state", "bar",
					})

					Expect(err).To(MatchError(`post-deploy-state "foo" is invalid, pre-delete-state "bar" is invalid`))
				})
			})

			Context("when an unknown flag is passed", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--unknown-flag"})
					Expect(err).To(MatchError("could not parse set-errand-state flags: flag provided but not defined: -unknown-flag"))
				})
			})

			Context("when the staged products finder fails", func() {
				It("returns an error", func() {
					stagedProductsFinder.FindReturns(api.StagedProductsFindOutput{}, errors.New("there was an error"))

					err := command.Execute([]string{
						"--product-name", "some-product",
						"--errand-name", "some-errand",
					})
					Expect(err).To(MatchError("failed to find staged product \"some-product\": there was an error"))
				})
			})

			Context("when no errand name is passed", func() {
				It("returns an error", func() {
					err := command.Execute([]string{
						"--product-name", "some-product-name",
						"--post-deploy-state", "enabled",
					})

					Expect(err).To(MatchError("could not parse set-errand-state flags: missing required flag \"--errand-name\""))
				})
			})

			Context("when the errands service fails", func() {
				It("returns an error", func() {
					errandsService.SetStateReturns(errors.New("there was an error"))

					err := command.Execute([]string{
						"--product-name", "some-product",
						"--errand-name", "some-errand",
					})
					Expect(err).To(MatchError("failed to set errand state: there was an error"))
				})
			})

			Context("when the product name is missing", func() {
				It("returns an error", func() {
					err := command.Execute([]string{})
					Expect(err).To(MatchError("could not parse set-errand-state flags: missing required flag \"--product-name\""))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewSetErrandState(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command sets the state of a product's errand.",
				ShortDescription: "sets state for a product's errand",
				Flags:            command.Options,
			}))
		})
	})
})
