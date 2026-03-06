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

	When("--stage-all-replicas is set", func() {
		BeforeEach(func() {
			fakeService.InfoReturns(api.Info{Version: "3.3.0"}, nil)
			fakeService.CheckProductAvailabilityReturns(true, nil)
		})

		It("stages all products matching the product_template_name including the primary", func() {
			fakeService.ListStagedProductsReturns(api.StagedProductsOutput{
				Products: []api.StagedProduct{
					{GUID: "ist-primary-guid", Type: "p-isolation-segment", ProductTemplateName: "p-isolation-segment"},
					{GUID: "ist-replica1-guid", Type: "p-isolation-segment-replica1", ProductTemplateName: "p-isolation-segment"},
					{GUID: "ist-replica2-guid", Type: "p-isolation-segment-replica2", ProductTemplateName: "p-isolation-segment"},
				},
			}, nil)

			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "10.4.0-build.7",
				"--stage-all-replicas",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.StageCallCount()).To(Equal(3))

			stageInput0, guid0 := fakeService.StageArgsForCall(0)
			Expect(stageInput0).To(Equal(api.StageProductInput{
				ProductName:    "p-isolation-segment",
				ProductVersion: "10.4.0-build.7",
			}))
			Expect(guid0).To(Equal("ist-primary-guid"))

			stageInput1, guid1 := fakeService.StageArgsForCall(1)
			Expect(stageInput1).To(Equal(api.StageProductInput{
				ProductName:    "p-isolation-segment",
				ProductVersion: "10.4.0-build.7",
			}))
			Expect(guid1).To(Equal("ist-replica1-guid"))

			stageInput2, guid2 := fakeService.StageArgsForCall(2)
			Expect(stageInput2).To(Equal(api.StageProductInput{
				ProductName:    "p-isolation-segment",
				ProductVersion: "10.4.0-build.7",
			}))
			Expect(guid2).To(Equal("ist-replica2-guid"))

			lastFormat, lastV := logger.PrintfArgsForCall(logger.PrintfCallCount() - 1)
			Expect(fmt.Sprintf(lastFormat, lastV...)).To(Equal("finished staging replicas"))
		})

		It("stages only the primary when no replicas exist", func() {
			fakeService.ListStagedProductsReturns(api.StagedProductsOutput{
				Products: []api.StagedProduct{
					{GUID: "ist-primary-guid", Type: "p-isolation-segment", ProductTemplateName: "p-isolation-segment"},
				},
			}, nil)

			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "10.4.0-build.7",
				"--stage-all-replicas",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeService.StageCallCount()).To(Equal(1))

			_, guid0 := fakeService.StageArgsForCall(0)
			Expect(guid0).To(Equal("ist-primary-guid"))
		})

		When("listing staged products fails", func() {
			It("returns an error", func() {
				fakeService.ListStagedProductsReturns(api.StagedProductsOutput{}, errors.New("staged products error"))

				command := commands.NewStageProduct(fakeService, logger)

				err := executeCommand(command, []string{
					"--product-name", "p-isolation-segment",
					"--product-version", "10.4.0-build.7",
					"--stage-all-replicas",
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to list staged products"))
			})
		})

		When("staging a replica fails", func() {
			It("returns an error", func() {
				fakeService.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "replica-guid", Type: "p-isolation-segment-replica1", ProductTemplateName: "p-isolation-segment"},
					},
				}, nil)
				fakeService.StageReturnsOnCall(0, errors.New("replica stage error"))

				command := commands.NewStageProduct(fakeService, logger)

				err := executeCommand(command, []string{
					"--product-name", "p-isolation-segment",
					"--product-version", "10.4.0-build.7",
					"--stage-all-replicas",
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to stage replica"))
			})
		})

		When("a replica is already at the requested version", func() {
			It("logs a friendly message and skips staging that replica", func() {
				fakeService.ListStagedProductsReturns(api.StagedProductsOutput{
					Products: []api.StagedProduct{
						{GUID: "ist-primary-guid", Type: "p-isolation-segment", ProductTemplateName: "p-isolation-segment", ProductVersion: "10.4.0-build.6"},
						{GUID: "ist-replica1-guid", Type: "p-isolation-segment-replica1", ProductTemplateName: "p-isolation-segment", ProductVersion: "10.4.0-build.7"},
						{GUID: "ist-replica2-guid", Type: "p-isolation-segment-replica2", ProductTemplateName: "p-isolation-segment", ProductVersion: "10.4.0-build.6"},
					},
				}, nil)

				command := commands.NewStageProduct(fakeService, logger)

				err := executeCommand(command, []string{
					"--product-name", "p-isolation-segment",
					"--product-version", "10.4.0-build.7",
					"--stage-all-replicas",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeService.StageCallCount()).To(Equal(2))

				var logMessages []string
				for i := 0; i < logger.PrintfCallCount(); i++ {
					format, v := logger.PrintfArgsForCall(i)
					logMessages = append(logMessages, fmt.Sprintf(format, v...))
				}
				Expect(logMessages).To(ContainElement("p-isolation-segment-replica1 10.4.0-build.7 is already staged"))
				Expect(logMessages).To(ContainElement("finished staging replicas"))
			})
		})
	})

	When("--stage-replicas is set", func() {
		BeforeEach(func() {
			fakeService.InfoReturns(api.Info{Version: "3.3.0"}, nil)
			fakeService.CheckProductAvailabilityReturns(true, nil)
			fakeService.ListStagedProductsReturns(api.StagedProductsOutput{
				Products: []api.StagedProduct{
					{GUID: "ist-primary-guid", Type: "p-isolation-segment", ProductTemplateName: "p-isolation-segment"},
					{GUID: "ist-replica1-guid", Type: "p-isolation-segment-replica1", ProductTemplateName: "p-isolation-segment"},
					{GUID: "ist-replica2-guid", Type: "p-isolation-segment-replica2", ProductTemplateName: "p-isolation-segment"},
				},
			}, nil)
		})

		It("stages only the specified replicas", func() {
			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "10.4.0-build.7",
				"--stage-replicas", "p-isolation-segment-replica2",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.StageCallCount()).To(Equal(1))

			_, guid0 := fakeService.StageArgsForCall(0)
			Expect(guid0).To(Equal("ist-replica2-guid"))
		})

		It("stages multiple comma-separated replicas", func() {
			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "10.4.0-build.7",
				"--stage-replicas", "p-isolation-segment-replica1,p-isolation-segment-replica2",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.StageCallCount()).To(Equal(2))
		})

		It("stages the main tile when it is included in --stage-replicas", func() {
			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "10.4.0-build.14",
				"--stage-replicas", "p-isolation-segment,p-isolation-segment-replica1",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.StageCallCount()).To(Equal(2))

			_, guid0 := fakeService.StageArgsForCall(0)
			_, guid1 := fakeService.StageArgsForCall(1)
			Expect([]string{guid0, guid1}).To(ConsistOf("ist-primary-guid", "ist-replica1-guid"))
		})

		It("still stages requested replicas when the main tile is already at the requested version", func() {
			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{
				StagedProducts: []api.DiagnosticProduct{
					{Name: "p-isolation-segment", Version: "10.4.0-build.14"},
				},
			}, nil)

			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "10.4.0-build.14",
				"--stage-replicas", "p-isolation-segment-replica1",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.StageCallCount()).To(Equal(1))
			_, guid := fakeService.StageArgsForCall(0)
			Expect(guid).To(Equal("ist-replica1-guid"))
		})

		When("a requested replica name is not found", func() {
			It("returns an error", func() {
				command := commands.NewStageProduct(fakeService, logger)

				err := executeCommand(command, []string{
					"--product-name", "p-isolation-segment",
					"--product-version", "10.4.0-build.7",
					"--stage-replicas", "p-isolation-segment-nonexistent",
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find replicas with type(s)"))
				Expect(err.Error()).To(ContainSubstring("p-isolation-segment-nonexistent"))
			})
		})
	})

	When("--stage-all-replicas and --stage-replicas are both set", func() {
		It("returns a mutual exclusion error", func() {
			fakeService.InfoReturns(api.Info{Version: "3.3.0"}, nil)

			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "10.4.0-build.7",
				"--stage-all-replicas",
				"--stage-replicas", "p-isolation-segment-replica1",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("--stage-all-replicas and --stage-replicas are mutually exclusive"))
			Expect(fakeService.StageCallCount()).To(Equal(0))
		})
	})

	When("Ops Manager version is older than 3.3 and --stage-all-replicas is set", func() {
		It("returns an error with a clear message", func() {
			fakeService.InfoReturns(api.Info{Version: "3.2.0"}, nil)

			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "10.4.0-build.7",
				"--stage-all-replicas",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stage-product replica flags require Ops Manager 3.3 or newer"))
			Expect(err.Error()).To(ContainSubstring("3.2.0"))
			Expect(fakeService.StageCallCount()).To(Equal(0))
		})
	})

	When("Ops Manager version is older than 3.3 and --stage-replicas is set", func() {
		It("returns an error with a clear message", func() {
			fakeService.InfoReturns(api.Info{Version: "3.2.0"}, nil)

			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "10.4.0-build.7",
				"--stage-replicas", "p-isolation-segment-replica1",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stage-product replica flags require Ops Manager 3.3 or newer"))
			Expect(err.Error()).To(ContainSubstring("3.2.0"))
			Expect(fakeService.StageCallCount()).To(Equal(0))
		})
	})

	When("fetching the Ops Manager version fails with replica flags", func() {
		It("returns an error and does not call Stage", func() {
			fakeService.InfoReturns(api.Info{}, errors.New("could not make request to info endpoint"))

			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "p-isolation-segment",
				"--product-version", "10.4.0-build.7",
				"--stage-all-replicas",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get Ops Manager version"))
			Expect(err.Error()).To(ContainSubstring("could not make request to info endpoint"))
			Expect(fakeService.StageCallCount()).To(Equal(0))
		})
	})

	When("no replica flags are set", func() {
		It("does not call Info()", func() {
			fakeService.CheckProductAvailabilityReturns(true, nil)

			command := commands.NewStageProduct(fakeService, logger)

			err := executeCommand(command, []string{
				"--product-name", "some-product",
				"--product-version", "some-version",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeService.InfoCallCount()).To(Equal(0))
			Expect(fakeService.ListStagedProductsCallCount()).To(Equal(0))
		})
	})
})
