package commands_test

import (
	"errors"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeployedManifest", func() {
	var (
		command                commands.DeployedManifest
		logger                 *fakes.Logger
		deployedProductsLister *fakes.DeployedProductsLister
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		deployedProductsLister = &fakes.DeployedProductsLister{}
		deployedProductsLister.ListReturns([]api.DeployedProductOutput{
			{Type: "other-product", GUID: "other-product-guid"},
			{Type: "some-product", GUID: "some-product-guid"},
		}, nil)
		deployedProductsLister.ManifestReturns(`---
name: some-product
key: value
`, nil)

		command = commands.NewDeployedManifest(deployedProductsLister, logger)
	})

	It("prints the manifest of the deployed product", func() {
		err := command.Execute([]string{
			"--product-name", "some-product",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(deployedProductsLister.ListCallCount()).To(Equal(1))

		Expect(deployedProductsLister.ManifestCallCount()).To(Equal(1))
		Expect(deployedProductsLister.ManifestArgsForCall(0)).To(Equal("some-product-guid"))

		Expect(logger.PrintCallCount()).To(Equal(1))
		Expect(logger.PrintArgsForCall(0)[0]).To(MatchYAML(`---
name: some-product
key: value
`))
	})

	Context("failure cases", func() {
		Context("when the flags cannot be parsed", func() {
			It("returns an error", func() {
				err := command.Execute([]string{
					"--unknown-flag", "unknown-value",
				})
				Expect(err).To(MatchError(ContainSubstring("flag provided but not defined")))
			})
		})

		Context("when the deployed products cannot be listed", func() {
			It("returns an error", func() {
				deployedProductsLister.ListReturns([]api.DeployedProductOutput{}, errors.New("deployed products cannot be listed"))

				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError(ContainSubstring("deployed products cannot be listed")))
			})
		})

		Context("when the guid is not found", func() {
			It("returns an error", func() {
				err := command.Execute([]string{
					"--product-name", "unknown-product",
				})
				Expect(err).To(MatchError(ContainSubstring("could not find given product")))
			})
		})

		Context("when the manifest cannot be returned", func() {
			It("returns an error", func() {
				deployedProductsLister.ManifestReturns("", errors.New("manifest could not be retrieved"))
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError(ContainSubstring("manifest could not be retrieved")))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage info", func() {
			usage := command.Usage()
			Expect(usage).To(Equal(jhanda.Usage{
				Description:      "This authenticated command prints the deployed manifest for a product",
				ShortDescription: "prints the deployed manifest for a product",
				Flags:            command.Options,
			}))
		})
	})
})
