package commands_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/models"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
)

func parseTime(timeString string) *time.Time {
	timeValue, err := time.Parse(time.RFC3339, timeString)

	if err != nil {
		return nil
	}

	return &timeValue
}

var _ = Describe("Installations", func() {
	var (
		command                  commands.Installations
		fakeInstallationsService *fakes.InstallationsService
		fakePresenter            *presenterfakes.Presenter
	)

	BeforeEach(func() {
		fakePresenter = &presenterfakes.Presenter{}
		fakeInstallationsService = &fakes.InstallationsService{}
		command = commands.NewInstallations(fakeInstallationsService, fakePresenter)
	})

	Describe("Execute", func() {
		It("lists recent installations as a table", func() {
			fakeInstallationsService.ListInstallationsReturns([]api.InstallationsServiceOutput{
				{
					ID:         1,
					UserName:   "some-user",
					Status:     "succeeded",
					StartedAt:  parseTime("2017-05-24T23:38:37.316Z"),
					FinishedAt: parseTime("2017-05-24T23:39:37.316Z"),
				},
				{
					ID:        2,
					UserName:  "some-user2",
					Status:    "failed",
					StartedAt: parseTime("2017-05-25T23:38:37.316Z"),
				},
			}, nil)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakePresenter.PresentInstallationsCallCount()).To(Equal(1))
			installations := fakePresenter.PresentInstallationsArgsForCall(0)
			Expect(installations).To(ConsistOf(
				models.Installation{
					Id:         1,
					User:       "some-user",
					Status:     "succeeded",
					StartedAt:  parseTime("2017-05-24T23:38:37.316Z"),
					FinishedAt: parseTime("2017-05-24T23:39:37.316Z"),
				},
				models.Installation{
					Id:        2,
					User:      "some-user2",
					Status:    "failed",
					StartedAt: parseTime("2017-05-25T23:38:37.316Z"),
				}))
		})

		Context("Failure cases", func() {
			Context("when the api fails to list installations", func() {
				It("returns an error", func() {
					fakeInstallationsService.ListInstallationsReturns([]api.InstallationsServiceOutput{}, errors.New("failed to retrieve installations"))

					err := command.Execute([]string{})
					Expect(err).To(MatchError("failed to retrieve installations"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewInstallations(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command lists all recent installation events.",
				ShortDescription: "list recent installation events",
			}))
		})
	})
})
