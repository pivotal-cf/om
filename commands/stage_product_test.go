package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StageProduct", func() {
	var (
		fakeService *fakes.StageProductService
		logger      *fakes.Logger
	)

	BeforeEach(func() {
		fakeService = &fakes.StageProductService{}
		logger = &fakes.Logger{}
	})

	It("stages a product", func() {
		fakeService.CheckProductAvailabilityReturns(true, nil)

		command := commands.NewStageProduct(fakeService, logger)

		fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{
			api.DeployedProductOutput{
				Type: "some-other-product",
				GUID: "deployed-product-guid",
			},
		}, nil)

		err := executeCommand(command, []string{
			"--product-name", "some-product",
			"--product-version", "some-version",
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(fakeService.ListDeployedProductsCallCount()).To(Equal(1))

		Expect(fakeService.StageCallCount()).To(Equal(1))
		stageProductInput, deployedProductGUID := fakeService.StageArgsForCall(0)
		Expect(stageProductInput).To(Equal(api.StageProductInput{
			ProductName:    "some-product",
			ProductVersion: "some-version",
		}))
		Expect(deployedProductGUID).To(BeEmpty())

		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("staging some-product some-version"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("finished staging"))
	})

	When("the product-version is `latest`", func() {
		It("uses the latest available product version", func() {
			fakeService.CheckProductAvailabilityReturns(true, nil)

			command := commands.NewStageProduct(fakeService, logger)

			fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{
				api.DeployedProductOutput{
					Type: "some-other-product",
					GUID: "deployed-product-guid",
				},
			}, nil)

			fakeService.GetLatestAvailableVersionReturns("1.1.1", nil)

			err := executeCommand(command, []string{
				"--product-name", "some-product",
				"--product-version", "latest",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.ListDeployedProductsCallCount()).To(Equal(1))

			Expect(fakeService.StageCallCount()).To(Equal(1))
			stageProductInput, _ := fakeService.StageArgsForCall(0)
			Expect(stageProductInput).To(Equal(api.StageProductInput{
				ProductName:    "some-product",
				ProductVersion: "1.1.1",
			}))
		})

		When("there is not latest version", func() {
			It("errors with a useful message", func() {
				fakeService.CheckProductAvailabilityReturns(true, nil)

				command := commands.NewStageProduct(fakeService, logger)

				fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{
					api.DeployedProductOutput{
						Type: "some-other-product",
						GUID: "deployed-product-guid",
					},
				}, nil)

				fakeService.GetLatestAvailableVersionReturns("", errors.New("some error"))

				err := executeCommand(command, []string{
					"--product-name", "some-product",
					"--product-version", "latest",
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find latest version: some error"))
			})
		})
	})

	When("a product has already been deployed", func() {
		It("stages the product", func() {
			fakeService.CheckProductAvailabilityReturns(true, nil)

			command := commands.NewStageProduct(fakeService, logger)

			fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{
				api.DeployedProductOutput{
					Type: "some-other-product",
					GUID: "other-deployed-product-guid",
				},
				api.DeployedProductOutput{
					Type: "some-product",
					GUID: "deployed-product-guid",
				},
			}, nil)

			err := executeCommand(command, []string{
				"--product-name", "some-product",
				"--product-version", "some-version",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.ListDeployedProductsCallCount()).To(Equal(1))

			Expect(fakeService.StageCallCount()).To(Equal(1))
			stageProductInput, deployedProductGUID := fakeService.StageArgsForCall(0)
			Expect(stageProductInput).To(Equal(api.StageProductInput{
				ProductName:    "some-product",
				ProductVersion: "some-version",
			}))
			Expect(deployedProductGUID).To(Equal("deployed-product-guid"))

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(Equal("staging some-product some-version"))

			format, v = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, v...)).To(Equal("finished staging"))
		})
	})

	When("the product version has already been staged", func() {
		It("no-ops and returns successfully", func() {
			fakeService.CheckProductAvailabilityReturns(true, nil)
			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{
				StagedProducts: []api.DiagnosticProduct{
					{
						Name:    "some-product",
						Version: "some-version",
					},
				},
			}, nil)

			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "some-product",
				"--product-version", "some-version",
			})
			Expect(err).ToNot(HaveOccurred())

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(Equal("some-product some-version is already staged"))

			Expect(fakeService.StageCallCount()).To(Equal(0))
		})
	})

	When("there is a running installation", func() {
		BeforeEach(func() {
			fakeService.ListInstallationsReturns([]api.InstallationsServiceOutput{
				{
					ID:         999,
					Status:     "running",
					Logs:       "",
					StartedAt:  nil,
					FinishedAt: nil,
					UserName:   "admin",
				},
			}, nil)
		})
		It("returns an error", func() {
			command := commands.NewStageProduct(fakeService, logger)
			err := executeCommand(command, []string{"--product-name", "cf", "--product-version", "some-version"})
			Expect(err).To(MatchError("OpsManager does not allow configuration or staging changes while apply changes are running to prevent data loss for configuration and/or staging changes"))
			Expect(fakeService.ListInstallationsCallCount()).To(Equal(1))
		})
	})

	When("the product is not available", func() {
		BeforeEach(func() {
			fakeService.CheckProductAvailabilityReturns(false, nil)
		})

		It("returns an error", func() {
			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "some-product",
				"--product-version", "some-version",
			})
			Expect(err).To(MatchError(ContainSubstring("failed to stage product: cannot find product")))
		})
	})

	When("the product availability cannot be determined", func() {
		BeforeEach(func() {
			fakeService.CheckProductAvailabilityReturns(false, errors.New("failed to check availability"))
		})

		It("returns an error", func() {
			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "some-product",
				"--product-version", "some-version",
			})
			Expect(err).To(MatchError(ContainSubstring("failed to stage product: cannot check availability")))
		})
	})

	When("the product cannot be staged", func() {
		It("returns an error", func() {
			command := commands.NewStageProduct(fakeService, logger)
			fakeService.CheckProductAvailabilityReturns(true, nil)
			fakeService.StageReturns(errors.New("some product error"))

			err := executeCommand(command, []string{"--product-name", "some-product", "--product-version", "some-version"})
			Expect(err).To(MatchError("failed to stage product: some product error"))
		})
	})

	When("the diagnostic report cannot be fetched", func() {
		It("returns an error", func() {
			command := commands.NewStageProduct(fakeService, logger)
			fakeService.CheckProductAvailabilityReturns(true, nil)
			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{}, errors.New("bad diagnostic report"))

			err := executeCommand(command, []string{"--product-name", "some-product", "--product-version", "some-version"})
			Expect(err).To(MatchError("failed to stage product: bad diagnostic report"))
		})
	})

	When("the deployed products cannot be fetched", func() {
		BeforeEach(func() {
			fakeService.ListDeployedProductsReturns(
				[]api.DeployedProductOutput{},
				errors.New("could not fetch deployed products"))
		})

		It("returns an error", func() {
			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "some-product",
				"--product-version", "some-version",
			})
			Expect(err).To(MatchError(ContainSubstring("failed to stage product: could not fetch deployed products")))
		})
	})
})
