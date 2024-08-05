package commands_test

import (
	"errors"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StagedManifest", func() {
	var (
		command     *commands.StagedManifest
		logger      *fakes.Logger
		fakeService *fakes.StagedManifestService
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		fakeService = &fakes.StagedManifestService{}
		fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{
			Product: api.StagedProduct{GUID: "some-product-guid", Type: "some-product"},
		}, nil)
		fakeService.GetStagedProductManifestReturns(`---
name: some-product
key: value
`, nil)

		command = commands.NewStagedManifest(fakeService, logger)
	})

	It("prints the manifest of the staged product", func() {
		err := executeCommand(command, []string{
			"--product-name", "some-product",
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(fakeService.GetStagedProductByNameCallCount()).To(Equal(1))
		Expect(fakeService.GetStagedProductByNameArgsForCall(0)).To(Equal("some-product"))

		Expect(fakeService.GetStagedProductManifestCallCount()).To(Equal(1))
		Expect(fakeService.GetStagedProductManifestArgsForCall(0)).To(Equal("some-product-guid"))

		Expect(logger.PrintCallCount()).To(Equal(1))
		Expect(logger.PrintArgsForCall(0)[0]).To(MatchYAML(`---
name: some-product
key: value
`))
	})

	When("the staged products service find call fails", func() {
		It("returns an error", func() {
			fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{}, errors.New("product find failed"))

			err := executeCommand(command, []string{
				"--product-name", "some-product",
			})
			Expect(err).To(MatchError(ContainSubstring("failed to find product: product find failed")))
		})
	})

	When("the staged products service manifest call fails", func() {
		It("returns an error", func() {
			fakeService.GetStagedProductManifestReturns("", errors.New("product manifest failed"))

			err := executeCommand(command, []string{
				"--product-name", "some-product",
			})
			Expect(err).To(MatchError(ContainSubstring("failed to fetch product manifest: product manifest failed")))
		})
	})
})
