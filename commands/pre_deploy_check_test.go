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
	})

	When("director is complete", func() {
		It("returns a 'the director is configured correctly' message", func() {
			service.ListPendingDirectorChangesReturns(api.PendingDirectorChangesOutput{
				EndpointResults: api.PreDeployCheck{
					Complete: true,
				},
			}, nil)
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
		})
	})
})
