package commands_test

import (
	"errors"
	"log"
	"regexp"

	"github.com/fatih/color"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
)

var _ = Describe("PreDeployCheck.Execute", func() {
	var (
		presenter *presenterfakes.FormattedPresenter
		service   *fakes.PreDeployCheckService
		stdout    *gbytes.Buffer
		logger    *log.Logger
	)

	BeforeEach(func() {
		presenter = &presenterfakes.FormattedPresenter{}
		service = &fakes.PreDeployCheckService{}
		stdout = gbytes.NewBuffer()
		logger = log.New(stdout, "", 0)

		// Default to working cases of director and product changes for separate testing
		service.ListPendingDirectorChangesReturns(api.PendingDirectorChangesOutput{
			EndpointResults: api.PreDeployCheck{
				Identifier: "p-bosh-guid",
				Complete:   true,
				Network: api.PreDeployNetwork{
					Assigned: true,
				},
				AvailabilityZone: api.PreDeployAvailabilityZone{
					Assigned: true,
				},
				Stemcells: []api.PreDeployStemcells{
					{
						Assigned:                true,
						RequiredStemcellVersion: "250.2",
						RequiredStemcellOS:      "ubuntu-xenial",
					},
				},
				Properties: []api.PreDeployProperty{},
				Resources:  api.PreDeployResources{},
				Verifiers:  []api.PreDeployVerifier{},
			},
		}, nil)

		service.ListAllPendingProductChangesReturns([]api.PendingProductChangesOutput{
			{
				EndpointResults: api.PreDeployCheck{
					Identifier: "p-guid",
					Complete:   true,
					Network: api.PreDeployNetwork{
						Assigned: true,
					},
					AvailabilityZone: api.PreDeployAvailabilityZone{
						Assigned: true,
					},
					Stemcells: []api.PreDeployStemcells{
						{
							Assigned:                true,
							RequiredStemcellVersion: "250.2",
							RequiredStemcellOS:      "ubuntu-xenial",
						},
					},
					Properties: []api.PreDeployProperty{},
					Resources:  api.PreDeployResources{},
					Verifiers:  []api.PreDeployVerifier{},
				},
			},
		}, nil)

		service.InfoReturns(api.Info{Version: "2.6.0"}, nil)
	})

	When("getting information about the director fails", func() {
		It("displays the error", func() {
			service.ListPendingDirectorChangesReturns(api.PendingDirectorChangesOutput{}, errors.New("something bad happened with the director"))

			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})
Expect(err).To(MatchError(ContainSubstring("something bad happened with the director")))
		})
	})

	When("getting information about the product fails", func() {
		It("displays the error", func() {
			service.ListAllPendingProductChangesReturns([]api.PendingProductChangesOutput{}, errors.New("something bad happened with the product"))
			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})
Expect(err).To(MatchError(ContainSubstring("something bad happened with the product")))
		})
	})

	When("the director and product are complete", func() {
		It("displays a helpful message", func() {
			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("Scanning OpsManager now ...")))
			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("[✓] director: p-bosh-guid")))
			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("[✓] product: p-guid")))
		})
	})

	When("the director is incomplete but the product is complete", func() {
		It("displays status and returns an error", func() {
			service.ListPendingDirectorChangesReturns(api.PendingDirectorChangesOutput{
				EndpointResults: api.PreDeployCheck{
					Identifier: "p-bosh-guid",
					Complete:   false,
				},
			}, nil)
			service.ListAllPendingProductChangesReturns([]api.PendingProductChangesOutput{
				{
					EndpointResults: api.PreDeployCheck{
						Identifier: "another-p-guid",
						Complete:   true,
					},
				},
			}, nil)
			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})
			Expect(err).To(MatchError(ContainSubstring("OpsManager is not fully configured")))

			Expect(string(stdout.Contents())).To(ContainSubstring("[X] director: p-bosh-guid"))
			Expect(string(stdout.Contents())).To(ContainSubstring("[✓] product: another-p-guid"))
		})

		It("does not display bosh director as a product", func() {
			service.ListPendingDirectorChangesReturns(api.PendingDirectorChangesOutput{
				EndpointResults: api.PreDeployCheck{
					Identifier: "p-bosh-guid",
					Complete:   false,
				},
			}, nil)
			service.ListAllPendingProductChangesReturns([]api.PendingProductChangesOutput{
				{
					EndpointResults: api.PreDeployCheck{
						Identifier: "p-bosh-guid",
						Complete:   false,
					},
				},
				{
					EndpointResults: api.PreDeployCheck{
						Identifier: "another-p-guid",
						Complete:   true,
					},
				},
			}, nil)
			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})
			Expect(err).To(MatchError(ContainSubstring("OpsManager is not fully configured")))

			Expect(string(stdout.Contents())).To(ContainSubstring("[X] director: p-bosh-guid"))
			Expect(string(stdout.Contents())).To(ContainSubstring("[✓] product: another-p-guid"))
			Expect(string(stdout.Contents())).ToNot(ContainSubstring("[X] product: p-bosh-guid"))
		})

		When("the summary has errors", func() {
			It("prints a list of all errors", func() {
				service.ListPendingDirectorChangesReturns(api.PendingDirectorChangesOutput{
					EndpointResults: api.PreDeployCheck{
						Identifier: "p-bosh-guid",
						Complete:   false,
						Network: api.PreDeployNetwork{
							Assigned: false,
						},
						AvailabilityZone: api.PreDeployAvailabilityZone{
							Assigned: false,
						},
						Stemcells: []api.PreDeployStemcells{
							{
								Assigned:                false,
								RequiredStemcellOS:      "ubuntu-trusty",
								RequiredStemcellVersion: "93.17",
							},
						},
						Properties: []api.PreDeployProperty{
							{
								Name:   "some-property",
								Type:   "string",
								Errors: []string{"can't be blank", "must be more than 0 characters"},
							},
						},
						Resources: api.PreDeployResources{
							Jobs: []api.PreDeployJob{
								{
									Identifier: "some-job",
									GUID:       "some-guid",
									Errors:     []string{"can't be blank", "must be greater than 0"},
								},
								{
									Identifier: "some-other-job",
									GUID:       "some-other-guid",
									Errors:     []string{"some-error"},
								},
							},
						},
						Verifiers: []api.PreDeployVerifier{
							{
								Type:      "WildcardDomainVerifier",
								Errors:    []string{"domain failed to resolve", "dns is bad"},
								Ignorable: false,
							},
							{
								Type:      "AZVerifier",
								Errors:    []string{"az is wrong"},
								Ignorable: false,
							},
						},
					},
				}, nil)
				command := commands.NewPreDeployCheck(presenter, service, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError(ContainSubstring("OpsManager is not fully configured")))

				contents := string(stdout.Contents())
				boldErr := color.New(color.Bold)

				Expect(contents).To(ContainSubstring("[X] director: p-bosh-guid"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " Network is not assigned"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " Availability Zone is not assigned"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " Availability Zone is not assigned"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " missing stemcell"))
				Expect(contents).To(ContainSubstring("Why: Required stemcell OS: ubuntu-trusty version 93.17"))
				Expect(contents).To(ContainSubstring("Fix: Download ubuntu-trusty version 93.17 from Pivnet and upload to OpsManager"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " property: some-property"))
				Expect(contents).To(ContainSubstring("Why: can't be blank"))
				Expect(contents).To(ContainSubstring("Why: must be more than 0 characters"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " resource: some-job"))
				Expect(contents).To(ContainSubstring("Why: can't be blank"))
				Expect(contents).To(ContainSubstring("Why: must be greater than 0"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " resource: some-other-job"))
				Expect(contents).To(ContainSubstring("Why: some-error"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " verifier: WildcardDomainVerifier"))
				Expect(contents).To(ContainSubstring("Why: domain failed to resolve"))
				Expect(contents).To(ContainSubstring("Why: dns is bad"))
				Expect(contents).To(ContainSubstring("Disable: `om disable-director-verifiers --type WildcardDomainVerifier`"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " verifier: AZVerifier"))
				Expect(contents).To(ContainSubstring("Why: az is wrong"))
			})
		})
	})

	When("the director is incomplete, and a product is incomplete", func() {
		It("displays status and returns an error", func() {
			service.ListPendingDirectorChangesReturns(api.PendingDirectorChangesOutput{
				EndpointResults: api.PreDeployCheck{
					Identifier: "p-bosh-guid",
					Complete:   false,
				},
			}, nil)
			service.ListAllPendingProductChangesReturns([]api.PendingProductChangesOutput{
				{
					EndpointResults: api.PreDeployCheck{
						Identifier: "another-p-guid",
						Complete:   false,
					},
				},
			}, nil)
			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})
			Expect(err).To(MatchError(ContainSubstring("OpsManager is not fully configured")))

			Expect(string(stdout.Contents())).To(ContainSubstring("[X] director: p-bosh-guid"))
			Expect(string(stdout.Contents())).To(ContainSubstring("[X] product: another-p-guid"))
		})
	})

	When("the director is complete, but a product is incomplete", func() {
		It("displays a message and returns an error", func() {
			service.ListAllPendingProductChangesReturns([]api.PendingProductChangesOutput{
				{
					EndpointResults: api.PreDeployCheck{
						Identifier: "p-guid",
						Complete:   false,
					},
				},
				{
					EndpointResults: api.PreDeployCheck{
						Identifier: "another-p-guid",
						Complete:   false,
					},
				},
			}, nil)
			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})
			Expect(err).To(MatchError(ContainSubstring("OpsManager is not fully configured")))

			Expect(string(stdout.Contents())).To(ContainSubstring("[✓] director: p-bosh-guid"))
			Expect(string(stdout.Contents())).To(ContainSubstring("[X] product: p-guid"))
			Expect(string(stdout.Contents())).To(ContainSubstring("[X] product: another-p-guid"))
		})

		When("the summary has errors", func() {
			It("prints a list of all errors", func() {
				service.ListAllPendingProductChangesReturns([]api.PendingProductChangesOutput{
					{
						EndpointResults: api.PreDeployCheck{
							Identifier: "p-guid",
							Complete:   false,
							Network: api.PreDeployNetwork{
								Assigned: false,
							},
							AvailabilityZone: api.PreDeployAvailabilityZone{
								Assigned: false,
							},
							Stemcells: []api.PreDeployStemcells{
								{
									Assigned:                false,
									RequiredStemcellOS:      "ubuntu-trusty",
									RequiredStemcellVersion: "93.17",
								},
							},
							Properties: []api.PreDeployProperty{
								{
									Name:   "some-property",
									Type:   "string",
									Errors: []string{"can't be blank", "must be more than 0 characters"},
								},
							},
							Resources: api.PreDeployResources{
								Jobs: []api.PreDeployJob{
									{
										Identifier: "some-job",
										GUID:       "some-guid",
										Errors:     []string{"can't be blank", "must be greater than 0"},
									},
									{
										Identifier: "some-other-job",
										GUID:       "some-other-guid",
										Errors:     []string{"some-error"},
									},
								},
							},
							Verifiers: []api.PreDeployVerifier{
								{
									Type:      "WildcardDomainVerifier",
									Errors:    []string{"domain failed to resolve", "dns is bad"},
									Ignorable: false,
								},
								{
									Type:      "AZVerifier",
									Errors:    []string{"az is wrong"},
									Ignorable: false,
								},
							},
						},
					},
				}, nil)
				command := commands.NewPreDeployCheck(presenter, service, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError(ContainSubstring("OpsManager is not fully configured")))

				contents := string(stdout.Contents())
				boldErr := color.New(color.Bold)

				Expect(contents).To(ContainSubstring("[X] product: p-guid"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " Network is not assigned"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " Availability Zone is not assigned"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " Availability Zone is not assigned"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " missing stemcell"))
				Expect(contents).To(ContainSubstring("Why: Required stemcell OS: ubuntu-trusty version 93.17"))
				Expect(contents).To(ContainSubstring("Fix: Download ubuntu-trusty version 93.17 from Pivnet and upload to OpsManager"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " property: some-property"))
				Expect(contents).To(ContainSubstring("Why: can't be blank"))
				Expect(contents).To(ContainSubstring("Why: must be more than 0 characters"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " resource: some-job"))
				Expect(contents).To(ContainSubstring("Why: can't be blank"))
				Expect(contents).To(ContainSubstring("Why: must be greater than 0"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " resource: some-other-job"))
				Expect(contents).To(ContainSubstring("Why: some-error"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " verifier: WildcardDomainVerifier"))
				Expect(contents).To(ContainSubstring("Why: domain failed to resolve"))
				Expect(contents).To(ContainSubstring("Why: dns is bad"))
				//Expect(contents).To(ContainSubstring("Disable: `om disable-product-verifiers --type WildcardDomainVerifier`"))
				Expect(contents).To(ContainSubstring(boldErr.Sprintf("    Error:") + " verifier: AZVerifier"))
				Expect(contents).To(ContainSubstring("Why: az is wrong"))
			})
		})
	})

	It("only works for version 2.6+", func() {
		for _, validVersion := range []string{"2.6.0", "2.7.0", "2.8.0"} {
			service.InfoReturns(api.Info{Version: validVersion}, nil)
			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())
		}
		for _, invalidVersion := range []string{"2.3.0", "2.4.0", "2.5.0"} {
			service.InfoReturns(api.Info{Version: invalidVersion}, nil)
			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
		}
	})
})
