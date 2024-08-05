package commands_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("AssignStemcell", func() {
	var (
		fakeService *fakes.AssignStemcellService
		logger      *fakes.Logger
		command     *commands.AssignStemcell
	)

	BeforeEach(func() {
		fakeService = &fakes.AssignStemcellService{}
		logger = &fakes.Logger{}
		command = commands.NewAssignStemcell(fakeService, logger)
	})

	When("--stemcell exists for the specified product", func() {
		BeforeEach(func() {
			fakeService.ListStemcellsReturns(api.ProductStemcells{
				Products: []api.ProductStemcell{
					{
						GUID:                  "cf-guid",
						ProductName:           "cf",
						StagedForDeletion:     false,
						StagedStemcellVersion: "",
						AvailableVersions: []string{
							"1234.5", "1234.6", "1234.99",
						},
					},
				},
			}, nil)
		})

		It("assigns the stemcell", func() {
			err := executeCommand(command, []string{"--product", "cf", "--stemcell", "1234.6"})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.ListStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignStemcellCallCount()).To(Equal(1))

			Expect(fakeService.AssignStemcellArgsForCall(0)).To(Equal(api.ProductStemcells{
				Products: []api.ProductStemcell{
					{
						GUID:                  "cf-guid",
						StagedStemcellVersion: "1234.6",
					},
				},
			}))
		})

		When("--stemcell latest is used", func() {
			It("assign the latest stemcell available", func() {
				err := executeCommand(command, []string{"--product", "cf", "--stemcell", "latest"})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeService.ListStemcellsCallCount()).To(Equal(1))
				Expect(fakeService.AssignStemcellCallCount()).To(Equal(1))

				Expect(fakeService.AssignStemcellArgsForCall(0)).To(Equal(api.ProductStemcells{
					Products: []api.ProductStemcell{
						{
							GUID:                  "cf-guid",
							StagedStemcellVersion: "1234.99",
						},
					},
				}))
			})
		})
	})

	When("there is no --stemcell provided", func() {
		BeforeEach(func() {
			fakeService.ListStemcellsReturns(api.ProductStemcells{
				Products: []api.ProductStemcell{
					{
						GUID:                  "cf-guid",
						ProductName:           "cf",
						StagedForDeletion:     false,
						StagedStemcellVersion: "",
						AvailableVersions: []string{
							"1234.5", "1234.6", "1234.99",
						},
					},
				},
			}, nil)
		})

		It("defaults to latest", func() {
			err := executeCommand(command, []string{"--product", "cf"})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.ListStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignStemcellCallCount()).To(Equal(1))

			Expect(fakeService.AssignStemcellArgsForCall(0)).To(Equal(api.ProductStemcells{
				Products: []api.ProductStemcell{
					{
						GUID:                  "cf-guid",
						StagedStemcellVersion: "1234.99",
					},
				},
			}))
		})
	})

	When("given stemcell version is not available", func() {
		BeforeEach(func() {
			fakeService.ListStemcellsReturns(api.ProductStemcells{
				Products: []api.ProductStemcell{
					{
						GUID:                  "cf-guid",
						ProductName:           "cf",
						StagedForDeletion:     false,
						StagedStemcellVersion: "",
						AvailableVersions: []string{
							"1234.5", "1234.6",
						},
					},
				},
			}, nil)
		})
		It("returns an error with the available stemcells", func() {
			err := executeCommand(command, []string{"--product", "cf", "--stemcell", "1234.1"})
			Expect(err).To(MatchError(ContainSubstring("stemcell version 1234.1 not found in Ops Manager")))
			Expect(err).To(MatchError(ContainSubstring("Available Stemcells for \"cf\": 1234.5, 1234.6")))

			Expect(fakeService.ListStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignStemcellCallCount()).To(Equal(0))
		})
	})

	When("the product is not found but the stemcell exists", func() {
		BeforeEach(func() {
			fakeService.ListStemcellsReturns(api.ProductStemcells{
				Products: []api.ProductStemcell{
					{
						GUID:                  "different-product-guid",
						ProductName:           "different-product",
						StagedForDeletion:     false,
						StagedStemcellVersion: "",
						AvailableVersions:     []string{"1234.5"},
					},
				},
			}, nil)
		})

		It("returns an error", func() {
			err := executeCommand(command, []string{"--product", "cf", "--stemcell", "1234.5"})
			Expect(err).To(MatchError(ContainSubstring("could not list product stemcell: product \"cf\" not found")))

			Expect(fakeService.ListStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignStemcellCallCount()).To(Equal(0))
		})
	})

	When("the product is staged for deletion", func() {
		BeforeEach(func() {
			fakeService.ListStemcellsReturns(api.ProductStemcells{
				Products: []api.ProductStemcell{
					{
						GUID:                  "cf-guid",
						ProductName:           "cf",
						StagedForDeletion:     true,
						StagedStemcellVersion: "",
						AvailableVersions:     []string{},
					},
				},
			}, nil)
		})

		It("returns an error", func() {
			err := executeCommand(command, []string{"--product", "cf", "--stemcell", "1234.5"})
			Expect(err).To(MatchError(ContainSubstring("could not assign stemcell: product \"cf\" is staged for deletion")))

			Expect(fakeService.ListStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignStemcellCallCount()).To(Equal(0))
		})
	})

	When("no available stemcell returned from api", func() {
		BeforeEach(func() {
			fakeService.ListStemcellsReturns(api.ProductStemcells{
				Products: []api.ProductStemcell{
					{
						GUID:                    "cf-guid",
						ProductName:             "cf",
						StagedForDeletion:       false,
						StagedStemcellVersion:   "",
						RequiredStemcellVersion: "1234.9",
						AvailableVersions:       []string{},
					},
				},
			}, nil)
		})

		It("returns an error", func() {
			err := executeCommand(command, []string{"--product", "cf", "--stemcell", "1234.5"})
			Expect(err).To(MatchError(ContainSubstring("no stemcells are available for \"cf\".")))
			Expect(err).To(MatchError(ContainSubstring("minimum required stemcell version is: 1234.9")))
			Expect(err).To(MatchError(ContainSubstring("upload-stemcell, and try again")))

			Expect(fakeService.ListStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignStemcellCallCount()).To(Equal(0))
		})
	})
})
