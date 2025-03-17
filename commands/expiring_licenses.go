package commands

import (
	"github.com/fatih/color"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

//counterfeiter:generate -o ./fakes/expiring_licenses_service.go --fake-name ExpiringLicensesService . expiringLicensesService
type expiringLicensesService interface {
	ListExpiringLicenses(string) ([]api.ExpiringLicenseOutPut, error)
}

type ExpiringLicenses struct {
	logger    logger
	presenter presenters.FormattedPresenter
	api       expiringLicensesService
	Options   struct {
		//		Staged        bool   `long:"staged" short:"s" description:"Specify to include staged products. Can be used with other options."`
		//		Deployed      bool   `long:"deployed" short:"d" description:"Specify to deployed products. Can be used with other options."`
		ExpiresWithin string `long:"expires-within"  short:"e"  description:"timeframe in which to check expiration. Default: \"90d\".\n\t\t\t\tdays(d), weeks(w), months(m) and years(y) supported."`
		//		Format        string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

func NewExpiringLicenses(presenter presenters.FormattedPresenter, service expiringLicensesService, logger logger) *ExpiringLicenses {
	return &ExpiringLicenses{
		presenter: presenter,
		api:       service,
		logger:    logger,
	}
}

func (e *ExpiringLicenses) Execute(args []string) error {
	if e.Options.ExpiresWithin == "" {
		e.Options.ExpiresWithin = "90d"
	}

	e.logger.Println("Getting expiring licenses...")
	expiringLicenses, _ := e.api.ListExpiringLicenses(e.Options.ExpiresWithin)

	if len(expiringLicenses) == 0 {
		e.logger.Printf(color.GreenString("[✓] No licenses are expiring in %s\n"), e.Options.ExpiresWithin)
	}
	e.presenter.PresentLicensedProducts(expiringLicenses)
	return nil
}
