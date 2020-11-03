package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"io/ioutil"
	"os"
)

var _ = Describe("AssignMutliStemcell", func() {
	var (
		fakeService *fakes.AssignMultiStemcellService
		logger      *fakes.Logger
		command     *commands.AssignMultiStemcell
	)

	BeforeEach(func() {
		fakeService = &fakes.AssignMultiStemcellService{}
		fakeService.InfoReturns(api.Info{Version: "2.6.0"}, nil)
		logger = &fakes.Logger{}
		command = commands.NewAssignMultiStemcell(fakeService, logger)
	})

	When("--stemcell exists for the specified product", func() {
		BeforeEach(func() {
			fakeService.ListMultiStemcellsReturns(api.ProductMultiStemcells{
				Products: []api.ProductMultiStemcell{
					{
						GUID:              "cf-guid",
						ProductName:       "cf",
						StagedForDeletion: false,
						StagedStemcells:   []api.StemcellObject{},
						AvailableVersions: []api.StemcellObject{
							{OS: "ubuntu-trusty", Version: "1234.5"},
							{OS: "ubuntu-xenial", Version: "1234.6"},
							{OS: "ubuntu-trusty", Version: "1234.6"},
							{OS: "ubuntu-xenial", Version: "1234.67"},
							{OS: "ubuntu-trusty", Version: "1234.99"},
						},
					},
				},
			}, nil)
		})

		It("assigns the stemcell", func() {
			err := executeCommand(command,[]string{"--product", "cf", "--stemcell", "ubuntu-trusty:1234.6"})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(1))

			Expect(fakeService.AssignMultiStemcellArgsForCall(0)).To(Equal(api.ProductMultiStemcells{
				Products: []api.ProductMultiStemcell{
					{
						GUID: "cf-guid",
						StagedStemcells: []api.StemcellObject{
							{OS: "ubuntu-trusty", Version: "1234.6"},
						},
					},
				},
			}))
		})

		It("assigns multiple stemcells", func() {
			err := executeCommand(command,[]string{"--product", "cf", "--stemcell", "ubuntu-trusty:1234.6", "--stemcell", "ubuntu-xenial:1234.67"})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(1))

			Expect(fakeService.AssignMultiStemcellArgsForCall(0)).To(Equal(api.ProductMultiStemcells{
				Products: []api.ProductMultiStemcell{
					{
						GUID: "cf-guid",
						StagedStemcells: []api.StemcellObject{
							{OS: "ubuntu-trusty", Version: "1234.6"},
							{OS: "ubuntu-xenial", Version: "1234.67"},
						},
					},
				},
			}))
		})

		It("assigns multiple stemcells with the same version number", func() {
			err := executeCommand(command,[]string{"--product", "cf", "--stemcell", "ubuntu-trusty:1234.6", "--stemcell", "ubuntu-xenial:1234.6"})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(1))

			Expect(fakeService.AssignMultiStemcellArgsForCall(0)).To(Equal(api.ProductMultiStemcells{
				Products: []api.ProductMultiStemcell{
					{
						GUID: "cf-guid",
						StagedStemcells: []api.StemcellObject{
							{OS: "ubuntu-trusty", Version: "1234.6"},
							{OS: "ubuntu-xenial", Version: "1234.6"},
						},
					},
				},
			}))
		})

		When("--stemcell latest is used", func() {
			It("assign the latest stemcell available", func() {
				err := executeCommand(command,[]string{"--product", "cf", "--stemcell", "ubuntu-trusty:latest"})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
				Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(1))

				Expect(fakeService.AssignMultiStemcellArgsForCall(0)).To(Equal(api.ProductMultiStemcells{
					Products: []api.ProductMultiStemcell{
						{
							GUID: "cf-guid",
							StagedStemcells: []api.StemcellObject{
								{OS: "ubuntu-trusty", Version: "1234.99"},
							},
						},
					},
				}))
			})
		})
	})

	When("config file is provided", func() {
		var configFile *os.File

		BeforeEach(func() {
			var err error

			fakeService.ListMultiStemcellsReturns(api.ProductMultiStemcells{
				Products: []api.ProductMultiStemcell{
					{
						GUID:              "cf-guid",
						ProductName:       "cf",
						StagedForDeletion: false,
						StagedStemcells:   []api.StemcellObject{},
						AvailableVersions: []api.StemcellObject{
							{OS: "ubuntu-trusty", Version: "1234.5"},
							{OS: "ubuntu-trusty", Version: "1234.6"},
							{OS: "ubuntu-xenial", Version: "1234.67"},
							{OS: "ubuntu-trusty", Version: "1234.99"},
						},
					},
				},
			}, nil)

			configContent := `
product: cf
stemcell: [ "ubuntu-trusty:1234.6", "ubuntu-xenial:latest" ]
`
			configFile, err = ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			_, err = configFile.WriteString(configContent)
			Expect(err).ToNot(HaveOccurred())
		})

		It("reads configuration from config file", func() {
			err := executeCommand(command,[]string{"--config", configFile.Name()})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(1))

			Expect(fakeService.AssignMultiStemcellArgsForCall(0)).To(Equal(api.ProductMultiStemcells{
				Products: []api.ProductMultiStemcell{
					{
						GUID: "cf-guid",
						StagedStemcells: []api.StemcellObject{
							{OS: "ubuntu-trusty", Version: "1234.6"},
							{OS: "ubuntu-xenial", Version: "1234.67"},
						},
					},
				},
			}))
		})
	})

	When("given stemcell version is not available", func() {
		BeforeEach(func() {
			fakeService.ListMultiStemcellsReturns(api.ProductMultiStemcells{
				Products: []api.ProductMultiStemcell{
					{
						GUID:              "cf-guid",
						ProductName:       "cf",
						StagedForDeletion: false,
						StagedStemcells:   []api.StemcellObject{},
						AvailableVersions: []api.StemcellObject{
							{OS: "ubuntu-trusty", Version: "1234.5"},
							{OS: "ubuntu-trusty", Version: "1234.6"},
						},
					},
				},
			}, nil)
		})
		It("returns an error with the available stemcells", func() {
			err := executeCommand(command,[]string{"--product", "cf", "--stemcell", "ubuntu-trusty:1234.1"})
			Expect(err).To(MatchError(ContainSubstring("stemcell version 1234.1 for ubuntu-trusty not found in Ops Manager")))
			Expect(err).To(MatchError(ContainSubstring("Available Stemcells for \"cf\": ubuntu-trusty 1234.5, ubuntu-trusty 1234.6")))

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(0))
		})
	})

	When("the product is not found but the stemcell exists", func() {
		BeforeEach(func() {
			fakeService.ListMultiStemcellsReturns(api.ProductMultiStemcells{
				Products: []api.ProductMultiStemcell{
					{
						GUID:              "cf-guid",
						ProductName:       "not-cf",
						StagedForDeletion: false,
						StagedStemcells:   []api.StemcellObject{},
						AvailableVersions: []api.StemcellObject{
							{OS: "ubuntu-trusty", Version: "1234.5"},
						},
					},
				},
			}, nil)
		})

		It("returns an error", func() {
			err := executeCommand(command,[]string{"--product", "cf", "--stemcell", "ubuntu-trusty:1234.5"})
			Expect(err).To(MatchError(ContainSubstring("could not list product stemcell: product \"cf\" not found")))

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(0))
		})
	})

	When("the product is staged for deletion", func() {
		BeforeEach(func() {
			fakeService.ListMultiStemcellsReturns(api.ProductMultiStemcells{
				Products: []api.ProductMultiStemcell{
					{
						GUID:              "cf-guid",
						ProductName:       "cf",
						StagedForDeletion: true,
						StagedStemcells:   []api.StemcellObject{},
						AvailableVersions: []api.StemcellObject{},
					},
				},
			}, nil)
		})

		It("returns an error", func() {
			err := executeCommand(command,[]string{"--product", "cf", "--stemcell", "ubuntu-trusty:1234.5"})
			Expect(err).To(MatchError(ContainSubstring("could not assign stemcell: product \"cf\" is staged for deletion")))

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(0))
		})
	})

	When("no available stemcell returned from api", func() {
		BeforeEach(func() {
			fakeService.ListMultiStemcellsReturns(api.ProductMultiStemcells{
				Products: []api.ProductMultiStemcell{
					{
						GUID:              "cf-guid",
						ProductName:       "cf",
						StagedForDeletion: false,
						StagedStemcells:   []api.StemcellObject{},
						AvailableVersions: []api.StemcellObject{},
						RequiredStemcells: []api.StemcellObject{
							{OS: "ubuntu-xenial", Version: "1234.9"},
						},
					},
				},
			}, nil)
		})

		It("returns an error", func() {
			err := executeCommand(command,[]string{"--product", "cf", "--stemcell", "ubuntu-xenial:1234.5"})
			Expect(err).To(MatchError(ContainSubstring("no stemcells are available for \"cf\".")))
			Expect(err).To(MatchError(ContainSubstring("minimum required stemcells are: ubuntu-xenial 1234.9")))
			Expect(err).To(MatchError(ContainSubstring("upload-stemcell, and try again")))

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(0))
		})
	})

	When("no available stemcell of the particular OS returned from api", func() {
		BeforeEach(func() {
			fakeService.ListMultiStemcellsReturns(api.ProductMultiStemcells{
				Products: []api.ProductMultiStemcell{
					{
						GUID:              "cf-guid",
						ProductName:       "cf",
						StagedForDeletion: false,
						StagedStemcells:   []api.StemcellObject{},
						AvailableVersions: []api.StemcellObject{
							{OS: "ubuntu-trusty", Version: "1234.9"},
						},
						RequiredStemcells: []api.StemcellObject{
							{OS: "ubuntu-xenial", Version: "1234.9"},
						},
					},
				},
			}, nil)
		})

		It("returns an error", func() {
			err := executeCommand(command,[]string{"--product", "cf", "--stemcell", "ubuntu-xenial:1234.5"})
			Expect(err).To(MatchError(ContainSubstring(`stemcell version 1234.5 for ubuntu-xenial not found in Ops Manager.`)))
			Expect(err).To(MatchError(ContainSubstring(`there are no available stemcells to for "cf"`)))
			Expect(err).To(MatchError(ContainSubstring("upload-stemcell, and try again")))

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(0))
		})
	})

	When("the product flag is not provided", func() {
		It("returns an error", func() {
			err := executeCommand(command,[]string{})
			Expect(err).To(MatchError("could not parse assign-stemcell flags: missing required flag \"--product\""))
		})
	})

	When("there is no --stemcell provided", func() {
		It("returns an error", func() {
			err := executeCommand(command,[]string{"--product", "cf"})
			Expect(err).To(MatchError(ContainSubstring(`missing required flag "--stemcell"`)))
		})
	})

	When("incorrect os and version are entered", func() {
		It("returns an error", func() {
			err := executeCommand(command,[]string{"--product", "cf", "--stemcell", "ubuntu    1234.5"})
			Expect(err).To(MatchError(ContainSubstring(`could not parse assign-stemcell arguments: expected "--stemcell" format value as "operating-system=version"`)))
		})
	})

	When("OpsManager is not 2.6+", func() {
		It("returns an error", func() {
			fakeService.InfoReturns(api.Info{Version: "2.2-build.1"}, nil)

			err := executeCommand(command,[]string{"--product", "cf", "--stemcell", "ubuntu=1234.5"})
			Expect(err).To(MatchError(ContainSubstring("this command can only be used with OpsManager 2.6+")))
		})
	})
})
