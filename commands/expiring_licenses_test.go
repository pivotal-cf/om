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
			service.ListExpiringLicensesReturns([]api.ExpiringLicenseOutput{}, nil)

			command := commands.NewExpiringLicenses(presenter, service, logger)
			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(presenter.SetFormatCallCount()).To(Equal(1))
			Expect(presenter.SetFormatArgsForCall(0)).To(Equal("table"))

			Expect(presenter.PresentLicensedProductsCallCount()).To(Equal(1))
			Expect(presenter.PresentLicensedProductsArgsForCall(0)).To(Equal([]api.ExpiringLicenseOutput{}))
		})
	})

	When("there are expiring licenses", func() {
		It("displays the licenses correctly", func() {
			expiringLicenses := []api.ExpiringLicenseOutput{
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

	When("validating timeframe formats", func() {
		It("accepts valid formats", func() {
			validFormats := []string{"1d", "2w", "3m", "1y", "10d", "20w", "30m", "5y"}
			for _, format := range validFormats {
				command := commands.NewExpiringLicenses(presenter, service, logger)
				err := executeCommand(command, []string{"--expires-within", format})
				Expect(err).ToNot(HaveOccurred(), "format %s should be valid", format)
			}
		})

		It("rejects invalid formats", func() {
			invalidFormats := []string{"0d", "0w", "0m", "0y", "1", "d", "w", "m", "y", "1x", "1.5d", "abc"}
			for _, format := range invalidFormats {
				command := commands.NewExpiringLicenses(presenter, service, logger)
				err := executeCommand(command, []string{"--expires-within", format})
				Expect(err).To(MatchError("only d,w,m, or y are supported. Default is \"3m\""), "format %s should be invalid", format)
			}
		})
	})

	When("different format options are specified", func() {
		var expiringLicenses []api.ExpiringLicenseOutput

		BeforeEach(func() {
			expiringLicenses = []api.ExpiringLicenseOutput{
				{
					ProductName: "pivotal-container-service",
					GUID:        "pks-guid-123",
					ExpiresAt:   time.Now().AddDate(0, 1, 0),
				},
			}
			service.ListExpiringLicensesReturns(expiringLicenses, nil)
		})

		It("defaults to table format", func() {
			command := commands.NewExpiringLicenses(presenter, service, logger)
			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(presenter.SetFormatCallCount()).To(Equal(1))
			Expect(presenter.SetFormatArgsForCall(0)).To(Equal("table"))
		})

		It("can output in JSON format", func() {
			command := commands.NewExpiringLicenses(presenter, service, logger)
			err := executeCommand(command, []string{"--format", "json"})
			Expect(err).ToNot(HaveOccurred())

			Expect(presenter.SetFormatCallCount()).To(Equal(1))
			Expect(presenter.SetFormatArgsForCall(0)).To(Equal("json"))

			Expect(presenter.PresentLicensedProductsCallCount()).To(Equal(1))
			Expect(presenter.PresentLicensedProductsArgsForCall(0)).To(Equal(expiringLicenses))
		})
	})

	When("there are expiring licenses with different product states", func() {
		It("displays staged and deployed licenses correctly", func() {
			expiryDate, _ := time.Parse("2006-01-02", "2026-03-20")
			expiringLicenses := []api.ExpiringLicenseOutput{
				{
					ProductName:  "cf",
					GUID:         "cf-staged",
					ExpiresAt:    expiryDate,
					ProductState: "staged",
				},
				{
					ProductName:  "p-bosh",
					GUID:         "p-bosh-deployed",
					ExpiresAt:    expiryDate,
					ProductState: "deployed",
				},
			}
			service.ListExpiringLicensesReturns(expiringLicenses, nil)

			command := commands.NewExpiringLicenses(presenter, service, logger)
			err := executeCommand(command, []string{"--expires-within", "3y"})
			Expect(err).ToNot(HaveOccurred())

			Expect(service.ListExpiringLicensesCallCount()).To(Equal(1))
			timeWindow, staged, deployed := service.ListExpiringLicensesArgsForCall(0)
			Expect(timeWindow).To(Equal("3y"))
			Expect(staged).To(BeFalse())
			Expect(deployed).To(BeFalse())

			Expect(presenter.PresentLicensedProductsCallCount()).To(Equal(1))
			presentedLicenses := presenter.PresentLicensedProductsArgsForCall(0)
			Expect(presentedLicenses).To(Equal(expiringLicenses))
		})

		It("filters by staged products when --staged flag is used", func() {
			expiryDate, _ := time.Parse("2006-01-02", "2026-03-20")
			expiringLicenses := []api.ExpiringLicenseOutput{
				{
					ProductName:  "cf",
					GUID:         "cf-staged",
					ExpiresAt:    expiryDate,
					ProductState: "staged",
				},
			}
			service.ListExpiringLicensesReturns(expiringLicenses, nil)

			command := commands.NewExpiringLicenses(presenter, service, logger)
			err := executeCommand(command, []string{"--staged"})
			Expect(err).ToNot(HaveOccurred())

			Expect(service.ListExpiringLicensesCallCount()).To(Equal(1))
			_, staged, deployed := service.ListExpiringLicensesArgsForCall(0)
			Expect(staged).To(BeTrue())
			Expect(deployed).To(BeFalse())

			Expect(presenter.PresentLicensedProductsCallCount()).To(Equal(1))
			presentedLicenses := presenter.PresentLicensedProductsArgsForCall(0)
			Expect(presentedLicenses).To(Equal(expiringLicenses))
		})

		It("filters by deployed products when --deployed flag is used", func() {
			expiryDate, _ := time.Parse("2006-01-02", "2026-03-20")
			expiringLicenses := []api.ExpiringLicenseOutput{
				{
					ProductName:  "p-bosh",
					GUID:         "p-bosh-deployed",
					ExpiresAt:    expiryDate,
					ProductState: "deployed",
				},
			}
			service.ListExpiringLicensesReturns(expiringLicenses, nil)

			command := commands.NewExpiringLicenses(presenter, service, logger)
			err := executeCommand(command, []string{"--deployed"})
			Expect(err).ToNot(HaveOccurred())

			Expect(service.ListExpiringLicensesCallCount()).To(Equal(1))
			_, staged, deployed := service.ListExpiringLicensesArgsForCall(0)
			Expect(staged).To(BeFalse())
			Expect(deployed).To(BeTrue())

			Expect(presenter.PresentLicensedProductsCallCount()).To(Equal(1))
			presentedLicenses := presenter.PresentLicensedProductsArgsForCall(0)
			Expect(presentedLicenses).To(Equal(expiringLicenses))
		})
	})
})
