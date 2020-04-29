package commands_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		command       commands.Installations
		fakeService   *fakes.InstallationsService
		fakePresenter *presenterfakes.FormattedPresenter
	)

	BeforeEach(func() {
		fakePresenter = &presenterfakes.FormattedPresenter{}
		fakeService = &fakes.InstallationsService{}
		command = commands.NewInstallations(fakeService, fakePresenter)
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			fakeService.ListInstallationsReturns([]api.InstallationsServiceOutput{
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
		})

		It("lists recent installations as a table", func() {
			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())

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

		When("the format flag is provided", func() {
			It("sets the format on the presenter", func() {
				err := command.Execute([]string{"--format", "json"})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
				Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("json"))
			})
		})

		Context("Failure cases", func() {
			When("an unknown flag is passed", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--unknown-flag"})
					Expect(err).To(MatchError("could not parse installations flags: flag provided but not defined: -unknown-flag"))
				})
			})

			When("the api fails to list installations", func() {
				It("returns an error", func() {
					fakeService.ListInstallationsReturns([]api.InstallationsServiceOutput{}, errors.New("failed to retrieve installations"))

					err := command.Execute([]string{})
					Expect(err).To(MatchError("failed to retrieve installations"))
				})
			})
		})
	})
})
