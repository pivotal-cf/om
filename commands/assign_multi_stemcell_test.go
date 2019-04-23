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
		command     commands.AssignMultiStemcell
	)

	BeforeEach(func() {
		fakeService = &fakes.AssignMultiStemcellService{}
		fakeService.InfoReturns(api.Info{Version: "2.6.0"}, nil)
		logger = &fakes.Logger{}
		command = commands.NewAssignMultiStemcell(fakeService, logger)
	})

	Context("when --stemcell exists for the specified product", func() {
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
			err := command.Execute([]string{"--product", "cf", "--stemcell", "ubuntu-trusty=1234.6"})
			Expect(err).NotTo(HaveOccurred())

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
			err := command.Execute([]string{"--product", "cf", "--stemcell", "ubuntu-trusty=1234.6", "--stemcell", "ubuntu-xenial=1234.67"})
			Expect(err).NotTo(HaveOccurred())

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
			err := command.Execute([]string{"--product", "cf", "--stemcell", "ubuntu-trusty=1234.6", "--stemcell", "ubuntu-xenial=1234.6"})
			Expect(err).NotTo(HaveOccurred())

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

		Context("when --stemcell latest is used", func() {
			It("assign the latest stemcell available", func() {
				err := command.Execute([]string{"--product", "cf", "--stemcell", "ubuntu-trusty=latest"})
				Expect(err).NotTo(HaveOccurred())

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

	Context("when config file is provided", func() {
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
stemcell: [ "ubuntu-trusty=1234.6", "ubuntu-xenial=latest" ]
`
			configFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = configFile.WriteString(configContent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("reads configuration from config file", func() {
			err := command.Execute([]string{"--config", configFile.Name()})
			Expect(err).NotTo(HaveOccurred())

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

	Context("when given stemcell version is not available", func() {
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
			err := command.Execute([]string{"--product", "cf", "--stemcell", "ubuntu-trusty=1234.1"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stemcell version 1234.1 for ubuntu-trusty not found in Ops Manager"))
			Expect(err.Error()).To(ContainSubstring("Available Stemcells for \"cf\": ubuntu-trusty 1234.5, ubuntu-trusty 1234.6"))

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(0))
		})
	})

	Context("when the product is not found but the stemcell exists", func() {
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
			err := command.Execute([]string{"--product", "cf", "--stemcell", "ubuntu-trusty=1234.5"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not list product stemcell: product \"cf\" not found"))

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(0))
		})
	})

	Context("when the product is staged for deletion", func() {
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
			err := command.Execute([]string{"--product", "cf", "--stemcell", "ubuntu-trusty=1234.5"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not assign stemcell: product \"cf\" is staged for deletion"))

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(0))
		})
	})

	Context("when no available stemcell returned from api", func() {
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
			err := command.Execute([]string{"--product", "cf", "--stemcell", "ubuntu-xenial=1234.5"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no stemcells are available for \"cf\"."))
			Expect(err.Error()).To(ContainSubstring("minimum required stemcells are: ubuntu-xenial 1234.9"))
			Expect(err.Error()).To(ContainSubstring("upload-stemcell, and try again"))

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(0))
		})
	})

	Context("when no available stemcell of the particular OS returned from api", func() {
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
			err := command.Execute([]string{"--product", "cf", "--stemcell", "ubuntu-xenial=1234.5"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`stemcell version 1234.5 for ubuntu-xenial not found in Ops Manager.`))
			Expect(err.Error()).To(ContainSubstring(`there are no available stemcells to for "cf" choose from`))
			Expect(err.Error()).To(ContainSubstring("upload-stemcell, and try again"))

			Expect(fakeService.ListMultiStemcellsCallCount()).To(Equal(1))
			Expect(fakeService.AssignMultiStemcellCallCount()).To(Equal(0))
		})
	})

	Context("when an unknown flag is provided", func() {
		It("returns an error", func() {
			err := command.Execute([]string{"--badflag"})
			Expect(err).To(MatchError("could not parse assign-stemcell flags: flag provided but not defined: -badflag"))
		})
	})

	Context("when the product flag is not provided", func() {
		It("returns an error", func() {
			err := command.Execute([]string{})
			Expect(err).To(MatchError("could not parse assign-stemcell flags: missing required flag \"--product\""))
		})
	})

	Context("when there is no --stemcell provided", func() {
		It("returns an error", func() {
			err := command.Execute([]string{"--product", "cf"})
			Expect(err.Error()).To(ContainSubstring(`missing required flag "--stemcell"`))
		})
	})

	Context("when incorrect os and version are entered", func() {
		It("returns an error", func() {
			err := command.Execute([]string{"--product", "cf", "--stemcell", "ubuntu    1234.5"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`could not parse assign-stemcell arguments: expected "--stemcell" format value as "operating-system=version"`))
		})
	})

	Context("when OpsManager is not 2.6+", func() {
		It("returns an error", func() {
			fakeService.InfoReturns(api.Info{Version: "2.2-build.1"}, nil)

			err := command.Execute([]string{"--product", "cf", "--stemcell", "ubuntu=1234.5"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("this command can only be used with OpsManager 2.6+"))
		})
	})
})
