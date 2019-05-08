package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
)

var _ = Describe("PreDeployCheck", func() {
	var (
		presenter *presenterfakes.FormattedPresenter
		service   *fakes.PreDeployCheckService
	)

	BeforeEach(func() {
		presenter = &presenterfakes.FormattedPresenter{}
		service = &fakes.PreDeployCheckService{}

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
			command := commands.NewPreDeployCheck(presenter, service)
			err := command.Execute([]string{})

			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("director is incomplete", func() {
		It("returns an error", func() {
			service.ListPendingDirectorChangesReturns(api.PendingDirectorChangesOutput{
				EndpointResults: api.PreDeployCheck{
					Complete: false,
				},
			}, nil)

			command := commands.NewPreDeployCheck(presenter, service)
			err := command.Execute([]string{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("director configuration incomplete"))
			Expect(err.Error()).To(ContainSubstring("Please validate your Ops Manager installation in the UI"))
		})
	})

	When("product is complete", func() {
		//TODO: get command STDOUT
		It("returns a 'the product with guid 'p-guid' is configured correctly' message", func() {
			// Default case should be success
			command := commands.NewPreDeployCheck(presenter, service)
			err := command.Execute([]string{})

			Expect(err).NotTo(HaveOccurred())
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

			command := commands.NewPreDeployCheck(presenter, service)
			err := command.Execute([]string{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("product configuration incomplete for product with guid 'p-guid'"))
			Expect(err.Error()).To(ContainSubstring("product configuration incomplete for product with guid 'another-p-guid'"))
			Expect(err.Error()).To(ContainSubstring("Please validate your Ops Manager installation in the UI"))
		})
	})
})
