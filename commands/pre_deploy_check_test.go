package commands_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
	"log"
)

var _ = Describe("PreDeployCheck", func() {
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
	})

	When("director is complete", func() {
		//TODO: get command STDOUT
		It("returns a 'the director is configured correctly' message", func() {
			// Default case should be success
			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(gbytes.Say("the director is configured correctly"))
		})
	})

	When("director is incomplete", func() {
		It("returns an error", func() {
			service.ListPendingDirectorChangesReturns(api.PendingDirectorChangesOutput{
				EndpointResults: api.PreDeployCheck{
					Complete: false,
				},
			}, nil)

			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("director configuration incomplete"))
			Expect(err.Error()).To(ContainSubstring("Please validate your Ops Manager installation in the UI"))
		})
	})

	When("getting information about the director fails", func() {
		It("displays the error", func() {
			service.ListPendingDirectorChangesReturns(api.PendingDirectorChangesOutput{}, errors.New("something bad happened with the director"))

			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("something bad happened with the director"))
		})
	})

	When("getting information about the product fails", func() {
		It("displays the error", func() {
			service.ListAllPendingProductChangesReturns([]api.PendingProductChangesOutput{}, errors.New("something bad happened with the product"))
			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("something bad happened with the product"))
		})
	})

	When("product is complete", func() {
		//TODO: get command STDOUT
		It("returns a 'the product with guid 'p-guid' is configured correctly' message", func() {
			// Default case should be success
			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(gbytes.Say("the product with guid 'p-guid' is configured correctly"))
		})
	})

	When("products are incomplete", func() {
		It("returns an error", func() {
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

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("product configuration incomplete for product with guid 'p-guid'"))
			Expect(err.Error()).To(ContainSubstring("product configuration incomplete for product with guid 'another-p-guid'"))
			Expect(err.Error()).To(ContainSubstring("Please validate your Ops Manager installation in the UI"))
		})
	})
})
