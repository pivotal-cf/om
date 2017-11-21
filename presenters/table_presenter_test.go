package presenters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/models"
	"github.com/pivotal-cf/om/presenters"
)

var _ = Describe("TablePresenter", func() {
	var (
		tablePresenter  presenters.TablePresenter
		fakeTableWriter *fakes.TableWriter
	)

	BeforeEach(func() {
		fakeTableWriter = &fakes.TableWriter{}
		tablePresenter = presenters.NewTablePresenter(fakeTableWriter)
	})

	Describe("PresentInstallations", func() {
		var installations []models.Installation

		BeforeEach(func() {
			installations = []models.Installation{
				{
					Id:         "some-id",
					User:       "some-user",
					Status:     "some-status",
					StartedAt:  "some-started-at",
					FinishedAt: "some-finished-at",
				},
			}
		})

		It("creates a table", func() {
			tablePresenter.PresentInstallations(installations)
			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))

			headers := fakeTableWriter.SetHeaderArgsForCall(0)
			Expect(headers).To(Equal([]string{"ID", "User", "Status", "Started At", "Finished At"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(1))
			values := fakeTableWriter.AppendArgsForCall(0)
			Expect(values).To(Equal([]string{
				installations[0].Id,
				installations[0].User,
				installations[0].Status,
				installations[0].StartedAt,
				installations[0].FinishedAt,
			}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})

		Context("when there are no installations", func() {
			BeforeEach(func() {
				installations = []models.Installation{}
			})

			It("creates an empty table when no installations are present", func() {
				tablePresenter.PresentInstallations(installations)
				Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))

				headers := fakeTableWriter.SetHeaderArgsForCall(0)
				Expect(headers).To(ConsistOf("ID", "User", "Status", "Started At", "Finished At"))

				Expect(fakeTableWriter.AppendCallCount()).To(Equal(0))

				Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
			})
		})
	})
})
