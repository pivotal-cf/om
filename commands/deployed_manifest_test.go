package commands_test

import (
	"errors"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeployedManifest", func() {
	var (
		command     *commands.DeployedManifest
		logger      *fakes.Logger
		fakeService *fakes.DeployedManifestService
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		fakeService = &fakes.DeployedManifestService{}
		fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{
			{Type: "other-product", GUID: "other-product-guid"},
			{Type: "some-product", GUID: "some-product-guid"},
		}, nil)
		fakeService.GetDeployedProductManifestReturns(`---
name: some-product
key: value
`, nil)

		command = commands.NewDeployedManifest(fakeService, logger)
	})

	It("prints the manifest of the deployed product", func() {
		err := executeCommand(command, []string{
			"--product-name", "some-product",
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(fakeService.ListDeployedProductsCallCount()).To(Equal(1))

		Expect(fakeService.GetDeployedProductManifestCallCount()).To(Equal(1))
		Expect(fakeService.GetDeployedProductManifestArgsForCall(0)).To(Equal("some-product-guid"))

		Expect(logger.PrintCallCount()).To(Equal(1))
		Expect(logger.PrintArgsForCall(0)[0]).To(MatchYAML(`---
name: some-product
key: value
`))
	})

	When("the deployed products cannot be listed", func() {
		It("returns an error", func() {
			fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{}, errors.New("deployed products cannot be listed"))

			err := executeCommand(command, []string{
				"--product-name", "some-product",
			})
			Expect(err).To(MatchError(ContainSubstring("deployed products cannot be listed")))
		})
	})

	When("the guid is not found", func() {
		It("returns an error", func() {
			err := executeCommand(command, []string{
				"--product-name", "unknown-product",
			})
			Expect(err).To(MatchError(ContainSubstring("could not find given product")))
		})
	})

	When("the manifest cannot be returned", func() {
		It("returns an error", func() {
			fakeService.GetDeployedProductManifestReturns("", errors.New("manifest could not be retrieved"))
			err := executeCommand(command, []string{
				"--product-name", "some-product",
			})
			Expect(err).To(MatchError(ContainSubstring("manifest could not be retrieved")))
		})
	})
})
