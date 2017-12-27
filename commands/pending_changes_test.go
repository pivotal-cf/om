package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
)

var _ = Describe("PendingChanges", func() {
	var (
		presenter *presenterfakes.Presenter
		pcService *fakes.PendingChangesService
		command   commands.PendingChanges
	)

	BeforeEach(func() {
		presenter = &presenterfakes.Presenter{}
		pcService = &fakes.PendingChangesService{}
		command = commands.NewPendingChanges(presenter, pcService)
	})

	It("lists the pending changes", func() {
		pcService.ListReturns(api.PendingChangesOutput{
			ChangeList: []api.ProductChange{
				{
					Product: "some-product",
					Action:  "update",
					Errands: []api.Errand{
						{
							Name:       "some-errand",
							PostDeploy: "on",
							PreDelete:  "false",
						},
						{
							Name:       "some-errand-2",
							PostDeploy: "when-change",
							PreDelete:  "false",
						},
					},
				},
				{
					Product: "some-product-without-errand",
					Action:  "install",
					Errands: []api.Errand{},
				},
			},
		}, nil)

		err := command.Execute([]string{})
		Expect(err).NotTo(HaveOccurred())

		Expect(presenter.PresentPendingChangesCallCount()).To(Equal(1))
	})

	Context("failure cases", func() {
		Context("when fetching the pending changes fails", func() {
			It("returns an error", func() {
				command := commands.NewPendingChanges(presenter, pcService)

				pcService.ListReturns(api.PendingChangesOutput{}, errors.New("beep boop"))

				err := command.Execute([]string{})
				Expect(err).To(MatchError("failed to retrieve pending changes beep boop"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewPendingChanges(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command lists all pending changes.",
				ShortDescription: "lists pending changes",
			}))
		})
	})
})
