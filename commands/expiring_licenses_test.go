package commands_test

import (
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
	"log"
	"regexp"

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

	When("there are no expiring licenses in the time range", func() {
		It("displays a helpful message", func() {
			command := commands.NewExpiringLicenses(presenter, service, logger)
			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("Getting expiring licenses...")))
			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("[✓] No licenses are expiring in 90d")))
		})
	})
})
