package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
)

var _ = Describe("PendingChanges.Execute", func() {
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

	When("the check flag is provided", func() {
		var options []string

		BeforeEach(func() {
			options = []string{"--check"}
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

	When("the format flag is provided", func() {
		It("sets the format on the presenter", func() {
			err := command.Execute([]string{"--format", "json"})
			Expect(err).NotTo(HaveOccurred())

			Expect(presenter.SetFormatArgsForCall(0)).To(Equal("json"))
		})
	})

	Describe("failure cases", func() {
		When("an unknown flag is passed", func() {
			It("returns an error", func() {
				err := command.Execute([]string{"--unknown-flag"})
				Expect(err).To(MatchError("could not parse pending-changes flags: flag provided but not defined: -unknown-flag"))
			})
		})

		When("fetching the pending changes fails", func() {
			It("returns an error", func() {
				command := commands.NewPendingChanges(presenter, pcService)

				pcService.ListStagedPendingChangesReturns(api.PendingChangesOutput{}, errors.New("beep boop"))

				err := command.Execute([]string{})
				Expect(err).To(MatchError("failed to retrieve pending changes beep boop"))
			})
		})

		Describe("Ops Man 2.5 and earlier", func() {
			When("completeness_check returns any false values", func() {
				It("returns an error for configuration_complete: false", func() {
					pcService.ListStagedPendingChangesReturns(api.PendingChangesOutput{
						ChangeList: []api.ProductChange{
							{
								GUID:   "some-product-without-errands",
								Action: "unchanged",
								CompletenessChecks: &api.CompletenessChecks{
									ConfigurationComplete:       false,
									StemcellPresent:             true,
									ConfigurablePropertiesValid: true,
								},
							},
						},
					}, nil)

					err := command.Execute([]string{})
					Expect(presenter.PresentPendingChangesCallCount()).To(Equal(1))
					Expect(err).To(MatchError(ContainSubstring("configuration is incomplete for guid some-product-without-errands")))
					Expect(err).To(MatchError(ContainSubstring("Please validate your Ops Manager installation in the UI")))
				})

				It("returns an error for stemcell_present: false", func() {
					pcService.ListStagedPendingChangesReturns(api.PendingChangesOutput{
						ChangeList: []api.ProductChange{
							{
								GUID:   "some-product-without-errands",
								Action: "unchanged",
								CompletenessChecks: &api.CompletenessChecks{
									ConfigurationComplete:       true,
									StemcellPresent:             false,
									ConfigurablePropertiesValid: true,
								},
							},
						},
					}, nil)

					err := command.Execute([]string{})
					Expect(presenter.PresentPendingChangesCallCount()).To(Equal(1))
					Expect(err).To(MatchError(ContainSubstring("stemcell is missing for one or more products for guid some-product-without-errands")))
					Expect(err).To(MatchError(ContainSubstring("Please validate your Ops Manager installation in the UI")))
				})

				It("returns an error for configurable_properties_valid: false", func() {
					pcService.ListStagedPendingChangesReturns(api.PendingChangesOutput{
						ChangeList: []api.ProductChange{
							{

								GUID:   "some-product-without-errands",
								Action: "unchanged",
								CompletenessChecks: &api.CompletenessChecks{
									ConfigurationComplete:       true,
									StemcellPresent:             true,
									ConfigurablePropertiesValid: false,
								},
							},
						},
					}, nil)

					err := command.Execute([]string{})
					Expect(presenter.PresentPendingChangesCallCount()).To(Equal(1))
					Expect(err).To(MatchError(ContainSubstring("one or more properties are invalid for guid some-product-without-errands")))
					Expect(err).To(MatchError(ContainSubstring("Please validate your Ops Manager installation in the UI")))
				})

				When("multiple products fail completeness_checks", func() {
					It("concatenates errors for multiple products", func() {
						pcService.ListStagedPendingChangesReturns(api.PendingChangesOutput{
							ChangeList: []api.ProductChange{
								{

									GUID:   "some-product-without-errands",
									Action: "unchanged",
									CompletenessChecks: &api.CompletenessChecks{
										ConfigurationComplete:       false,
										StemcellPresent:             false,
										ConfigurablePropertiesValid: false,
									},
								},
								{

									GUID:   "second-product-without-errands",
									Action: "unchanged",
									CompletenessChecks: &api.CompletenessChecks{
										ConfigurationComplete:       false,
										StemcellPresent:             false,
										ConfigurablePropertiesValid: false,
									},
								},
							},
						}, nil)

						err := command.Execute([]string{})
						Expect(presenter.PresentPendingChangesCallCount()).To(Equal(1))
						Expect(err).To(MatchError(ContainSubstring("one or more properties are invalid for guid some-product-without-errands")))
						Expect(err).To(MatchError(ContainSubstring("stemcell is missing for one or more products for guid some-product-without-errands")))
						Expect(err).To(MatchError(ContainSubstring("configuration is incomplete for guid some-product-without-errands")))
						Expect(err).To(MatchError(ContainSubstring("one or more properties are invalid for guid second-product-without-errands")))
						Expect(err).To(MatchError(ContainSubstring("stemcell is missing for one or more products for guid second-product-without-errands")))
						Expect(err).To(MatchError(ContainSubstring("configuration is incomplete for guid second-product-without-errands")))
						Expect(err).To(MatchError(ContainSubstring("Please validate your Ops Manager installation in the UI")))
					})
				})

				When("multiple completeness_checks fail for a single product", func() {
					It("concatenates errors into the same error string", func() {
						pcService.ListStagedPendingChangesReturns(api.PendingChangesOutput{
							ChangeList: []api.ProductChange{
								{

									GUID:   "some-product-without-errands",
									Action: "unchanged",
									CompletenessChecks: &api.CompletenessChecks{
										ConfigurationComplete:       false,
										StemcellPresent:             false,
										ConfigurablePropertiesValid: false,
									},
								},
							},
						}, nil)

						err := command.Execute([]string{})
						Expect(presenter.PresentPendingChangesCallCount()).To(Equal(1))
						Expect(err).To(MatchError(ContainSubstring("one or more properties are invalid for guid some-product-without-errands")))
						Expect(err).To(MatchError(ContainSubstring("stemcell is missing for one or more products for guid some-product-without-errands")))
						Expect(err).To(MatchError(ContainSubstring("configuration is incomplete for guid some-product-without-errands")))
						Expect(err).To(MatchError(ContainSubstring("Please validate your Ops Manager installation in the UI")))
					})
				})
			})
		})

		Describe("Ops Man 2.6 and later", func() {

		})
	})
})
