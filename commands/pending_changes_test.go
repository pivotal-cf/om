package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	jhandacommands "github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("PendingChanges", func() {
	var (
		tableWriter *fakes.TableWriter
		pcService   *fakes.PendingChangesService
		command     commands.PendingChanges
	)

	BeforeEach(func() {
		tableWriter = &fakes.TableWriter{}
		pcService = &fakes.PendingChangesService{}
		command = commands.NewPendingChanges(tableWriter, pcService)
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

		Expect(tableWriter.SetHeaderCallCount()).To(Equal(1))
		Expect(tableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"PRODUCT", "ACTION", "ERRANDS"}))

		Expect(tableWriter.AppendCallCount()).To(Equal(3))
		Expect(tableWriter.AppendArgsForCall(0)).To(Equal([]string{"some-product", "update", "some-errand"}))
		Expect(tableWriter.AppendArgsForCall(1)).To(Equal([]string{"", "", "some-errand-2"}))
		Expect(tableWriter.AppendArgsForCall(2)).To(Equal([]string{"some-product-without-errand", "install", ""}))
	})

	Context("failure cases", func() {
		Context("when fetching the pending changes fails", func() {
			It("returns an error", func() {
				command := commands.NewPendingChanges(tableWriter, pcService)

				pcService.ListReturns(api.PendingChangesOutput{}, errors.New("beep boop"))

				err := command.Execute([]string{})
				Expect(err).To(MatchError("failed to retrieve pending changes beep boop"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewPendingChanges(nil, nil)
			Expect(command.Usage()).To(Equal(jhandacommands.Usage{
				Description:      "This authenticated command lists all pending changes.",
				ShortDescription: "lists pending changes",
			}))
		})
	})
})
