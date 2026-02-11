package commands_test

import (
	"errors"
	"fmt"
	"os"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReplicateProduct", func() {
	var (
		fakeService *fakes.StageProductService
		logger      *fakes.Logger
	)

	BeforeEach(func() {
		fakeService = &fakes.StageProductService{}
		logger = &fakes.Logger{}
	})

	It("replicates a product with product-name, product-version, and replica-suffix", func() {
		fakeService.CheckProductAvailabilityReturns(true, nil)
		fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{StagedProducts: []api.DiagnosticProduct{}}, nil)

		command := commands.NewReplicateProduct(fakeService, logger)

		err := executeCommand(command, []string{
			"--product-name", "p-isolation-segment",
			"--product-version", "10.4.0-build.7",
			"--replica-suffix", "fun-suffix-2",
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(fakeService.StageCallCount()).To(Equal(1))
		stageProductInput, deployedProductGUID := fakeService.StageArgsForCall(0)
		Expect(stageProductInput).To(Equal(api.StageProductInput{
			ProductName:    "p-isolation-segment",
			ProductVersion: "10.4.0-build.7",
			Replicate:      true,
			ReplicaSuffix:  "fun-suffix-2",
		}))
		Expect(deployedProductGUID).To(BeEmpty())

		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("replicating p-isolation-segment 10.4.0-build.7 with suffix fun-suffix-2"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("finished replicating"))
	})

	When("replica-suffix is missing", func() {
		It("returns an error", func() {
			command := commands.NewReplicateProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "10.4.0-build.7",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("--replica-suffix"))
			Expect(err.Error()).To(ContainSubstring("are required"))
			Expect(fakeService.StageCallCount()).To(Equal(0))
		})
	})

	When("product-name is missing", func() {
		It("returns an error", func() {
			command := commands.NewReplicateProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-version", "10.4.0-build.7",
				"--replica-suffix", "fun-suffix-2",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("are required"))
			Expect(fakeService.StageCallCount()).To(Equal(0))
		})
	})

	When("product-version is missing", func() {
		It("returns an error", func() {
			command := commands.NewReplicateProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--replica-suffix", "fun-suffix-2",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("are required"))
			Expect(fakeService.StageCallCount()).To(Equal(0))
		})
	})

	When("a config file is provided", func() {
		It("loads product-name, product-version, and replica-suffix from the config file", func() {
			fakeService.CheckProductAvailabilityReturns(true, nil)
			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{StagedProducts: []api.DiagnosticProduct{}}, nil)

			configFile, err := os.CreateTemp("", "replicate-config.yml")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(configFile.Name())
			_, err = configFile.WriteString(`product-name: p-isolation-segment
product-version: 10.4.0-build.7
replica-suffix: my-suffix
`)
			Expect(err).ToNot(HaveOccurred())
			Expect(configFile.Close()).ToNot(HaveOccurred())

			command := commands.NewReplicateProduct(fakeService, logger)

			err = executeCommand(command, []string{
				"--config", configFile.Name(),
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.StageCallCount()).To(Equal(1))
			stageProductInput, _ := fakeService.StageArgsForCall(0)
			Expect(stageProductInput.ProductName).To(Equal("p-isolation-segment"))
			Expect(stageProductInput.ProductVersion).To(Equal("10.4.0-build.7"))
			Expect(stageProductInput.ReplicaSuffix).To(Equal("my-suffix"))
			Expect(stageProductInput.Replicate).To(BeTrue())
		})
	})

	When("the product-version is latest", func() {
		It("uses the latest available product version", func() {
			fakeService.CheckProductAvailabilityReturns(true, nil)
			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{StagedProducts: []api.DiagnosticProduct{}}, nil)
			fakeService.GetLatestAvailableVersionReturns("10.4.0-build.9", nil)

			command := commands.NewReplicateProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "latest",
				"--replica-suffix", "fun-suffix-2",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.GetLatestAvailableVersionCallCount()).To(Equal(1))
			Expect(fakeService.GetLatestAvailableVersionArgsForCall(0)).To(Equal("p-isolation-segment"))

			Expect(fakeService.StageCallCount()).To(Equal(1))
			stageProductInput, _ := fakeService.StageArgsForCall(0)
			Expect(stageProductInput.ProductVersion).To(Equal("10.4.0-build.9"))
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
			command := commands.NewReplicateProduct(fakeService, logger)
			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "10.4.0-build.7",
				"--replica-suffix", "fun-suffix-2",
			})
			Expect(err).To(MatchError(ContainSubstring("OpsManager does not allow configuration or staging changes")))
			Expect(fakeService.StageCallCount()).To(Equal(0))
		})
	})

	When("the product is not available", func() {
		BeforeEach(func() {
			fakeService.CheckProductAvailabilityReturns(false, nil)
			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{StagedProducts: []api.DiagnosticProduct{}}, nil)
		})
		It("returns an error", func() {
			command := commands.NewReplicateProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "some-product",
				"--product-version", "some-version",
				"--replica-suffix", "my-suffix",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to replicate product: cannot find product"))
			Expect(fakeService.StageCallCount()).To(Equal(0))
		})
	})

	When("the product availability cannot be determined", func() {
		BeforeEach(func() {
			fakeService.CheckProductAvailabilityReturns(false, errors.New("failed to check availability"))
			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{StagedProducts: []api.DiagnosticProduct{}}, nil)
		})
		It("returns an error", func() {
			command := commands.NewReplicateProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "some-product",
				"--product-version", "some-version",
				"--replica-suffix", "my-suffix",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to replicate product: cannot check availability"))
			Expect(fakeService.StageCallCount()).To(Equal(0))
		})
	})

	When("Staging the product returns an error", func() {
		It("returns the error", func() {
			fakeService.CheckProductAvailabilityReturns(true, nil)
			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{StagedProducts: []api.DiagnosticProduct{}}, nil)
			fakeService.StageReturns(errors.New("some product error"))

			command := commands.NewReplicateProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "some-product",
				"--product-version", "some-version",
				"--replica-suffix", "my-suffix",
			})
			Expect(err).To(MatchError("failed to replicate product: some product error"))
		})
	})

	When("the replica is already staged", func() {
		It("no-ops and returns successfully", func() {
			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{
				StagedProducts: []api.DiagnosticProduct{
					{
						Name:    "p-isolation-segment-fun-suffix-2",
						Version: "10.4.0-build.7",
					},
				},
			}, nil)

			command := commands.NewReplicateProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "10.4.0-build.7",
				"--replica-suffix", "fun-suffix-2",
			})
			Expect(err).ToNot(HaveOccurred())

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(Equal("p-isolation-segment 10.4.0-build.7 with suffix fun-suffix-2 is already staged"))

			Expect(fakeService.StageCallCount()).To(Equal(0))
		})
	})
})
