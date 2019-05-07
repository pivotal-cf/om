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
		presenter *presenterfakes.FormattedPresenter
		pcService *fakes.PendingChangesService
		command   commands.PendingChanges
	)

	BeforeEach(func() {
		presenter = &presenterfakes.FormattedPresenter{}
		pcService = &fakes.PendingChangesService{}
		command = commands.NewPendingChanges(presenter, pcService)
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			pcService.ListStagedPendingChangesReturns(api.PendingChangesOutput{
				ChangeList: []api.ProductChange{
					{
						GUID:   "some-product",
						Action: "update",
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
						GUID:    "some-product-without-errand",
						Action:  "install",
						Errands: []api.Errand{},
					},
				},
			}, nil)
		})

		It("lists the pending changes", func() {
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(presenter.SetFormatArgsForCall(0)).To(Equal("table"))
			Expect(presenter.PresentPendingChangesCallCount()).To(Equal(1))
		})

		Context("when the validate-opsman-clean flag is provided", func() {
			var options []string

			BeforeEach(func() {
				options = []string{"--validate-opsman-clean"}
			})
			When("there are pending changes", func() {
				BeforeEach(func() {
					pcService.ListStagedPendingChangesReturns(api.PendingChangesOutput{
						ChangeList: []api.ProductChange{
							{
								GUID:   "some-product",
								Action: "unchanged",
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
								GUID:    "some-other-product-without-errand",
								Action:  "install",
								Errands: []api.Errand{},
							},
							{
								GUID:   "some-other-product",
								Action: "install",
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
								GUID:    "some-other-product-without-errand",
								Action:  "install",
								Errands: []api.Errand{},
							},
						},
					}, nil)
				})

				It("lists change information for all products and returns an error", func() {
					err := command.Execute(options)
					Expect(presenter.PresentPendingChangesCallCount()).To(Equal(1))
					Expect(err).To(HaveOccurred())
				})
			})

			When("there are no pending changes", func() {
				BeforeEach(func() {
					pcService.ListStagedPendingChangesReturns(api.PendingChangesOutput{
						ChangeList: []api.ProductChange{
							{
								GUID:   "some-product-without-errands",
								Action: "unchanged",
							},
						},
					}, nil)
				})
				It("lists change information for all products and does not return an error", func() {
					err := command.Execute(options)
					Expect(presenter.PresentPendingChangesCallCount()).To(Equal(1))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when the format flag is provided", func() {
			It("sets the format on the presenter", func() {
				err := command.Execute([]string{"--format", "json"})
				Expect(err).NotTo(HaveOccurred())

				Expect(presenter.SetFormatArgsForCall(0)).To(Equal("json"))
			})
		})

		Describe("failure cases", func() {
			Context("when an unknown flag is passed", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--unknown-flag"})
					Expect(err).To(MatchError("could not parse pending-changes flags: flag provided but not defined: -unknown-flag"))
				})
			})

			Context("when fetching the pending changes fails", func() {
				It("returns an error", func() {
					command := commands.NewPendingChanges(presenter, pcService)

					pcService.ListStagedPendingChangesReturns(api.PendingChangesOutput{}, errors.New("beep boop"))

					err := command.Execute([]string{})
					Expect(err).To(MatchError("failed to retrieve pending changes beep boop"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewPendingChanges(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command lists all pending changes.",
				ShortDescription: "lists pending changes",
				Flags:            command.Options,
			}))
		})
	})
})
