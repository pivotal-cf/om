package commands_test

import (
	"log"
	"time"

	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("ExpiringLicenses", func() {
	var (
		presenter *presenterfakes.FormattedPresenter
		service   *fakes.ExpiringLicensesService
		stdout    *gbytes.Buffer
		logger    *log.Logger
	)

	BeforeEach(func() {
		service = &fakes.ExpiringLicensesService{}
		stdout = gbytes.NewBuffer()
		logger = log.New(stdout, "", 0)
		presenter = &presenterfakes.FormattedPresenter{}
	})

	When("there are no expiring licenses", func() {
		It("presents empty list", func() {
			service.ListExpiringLicensesReturns([]api.ExpiringLicenseOutPut{}, nil)

			command := commands.NewExpiringLicenses(presenter, service, logger)
			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(presenter.SetFormatCallCount()).To(Equal(1))
			Expect(presenter.SetFormatArgsForCall(0)).To(Equal("table"))

			Expect(presenter.PresentLicensedProductsCallCount()).To(Equal(1))
			Expect(presenter.PresentLicensedProductsArgsForCall(0)).To(Equal([]api.ExpiringLicenseOutPut{}))
		})
	})

	When("there are expiring licenses", func() {
		It("displays the licenses correctly", func() {
			expiringLicenses := []api.ExpiringLicenseOutPut{
				{
					ProductName: "pivotal-container-service",
					GUID:        "pks-guid-123",
					ExpiresAt:   time.Now().AddDate(0, 1, 0), // expires in 1 month
				},
				{
					ProductName: "pivotal-application-service",
					GUID:        "pas-guid-456",
					ExpiresAt:   time.Now().AddDate(0, 2, 0), // expires in 2 months
				},
			}

			service.ListExpiringLicensesReturns(expiringLicenses, nil)

			command := commands.NewExpiringLicenses(presenter, service, logger)
			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(presenter.SetFormatCallCount()).To(Equal(1))
			Expect(presenter.SetFormatArgsForCall(0)).To(Equal("table"))

			Expect(presenter.PresentLicensedProductsCallCount()).To(Equal(1))
			Expect(presenter.PresentLicensedProductsArgsForCall(0)).To(Equal(expiringLicenses))
		})
	})
})
