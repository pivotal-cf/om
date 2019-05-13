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

		service.InfoReturns(api.Info{Version: "2.6.0"}, nil)
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

	When("the director and product are complete", func() {
		It("displays a helpful message", func() {
			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout).To(gbytes.Say("The director and products are configured correctly."))
		})
	})

	When("the director is incomplete", func() {
		It("displays a message and returns an error", func() {
			service.ListPendingDirectorChangesReturns(api.PendingDirectorChangesOutput{
				EndpointResults: api.PreDeployCheck{
					Identifier: "p-guid",
					Complete:   false,
				},
			}, nil)
			command := commands.NewPreDeployCheck(presenter, service, logger)
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())

			Expect(string(stdout.Contents())).To(Equal("The director is not configured correctly.\n"))
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
			Expect(err).To(HaveOccurred())

			Expect(stdout).To(gbytes.Say(`The director is configured correctly, but the following product\(s\) are not.`))
			Expect(stdout).To(gbytes.Say(`\[X\] p-guid`))
			Expect(stdout).To(gbytes.Say(`\[X\] another-p-guid`))
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
