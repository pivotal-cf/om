package commands

import (
	"errors"
	"github.com/fatih/color"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
	"regexp"
)

//counterfeiter:generate -o ./fakes/expiring_licenses_service.go --fake-name ExpiringLicensesService . expiringLicensesService
type expiringLicensesService interface {
	ListExpiringLicenses(string, bool, bool) ([]api.ExpiringLicenseOutPut, error)
}

type ExpiringLicenses struct {
	logger    logger
	presenter presenters.FormattedPresenter
	api       expiringLicensesService
	Options   struct {
		Staged        bool   `long:"staged" short:"s" description:"Specify to include staged products. Can be used with other options."`
		Deployed      bool   `long:"deployed" short:"d" description:"Specify to deployed products. Can be used with other options."`
		ExpiresWithin string `long:"expires-within"  short:"e"  description:"timeframe in which to check expiration. Default: \"90d\".\n\t\t\t\tdays(d), weeks(w), months(m) and years(y) supported."`
		Format        string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
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
		e.Options.ExpiresWithin = "3m"
	}
	err := e.validateConfig()
	if err != nil {
		return err
	}

	e.logger.Println("Getting expiring licenses...")
	expiringLicenses, _ := e.api.ListExpiringLicenses(e.Options.ExpiresWithin, e.Options.Staged, e.Options.Deployed)

	if len(expiringLicenses) == 0 {
		e.logger.Printf(color.GreenString("[✓] No licenses are expiring in %s\n"), e.Options.ExpiresWithin)
	}
	e.presenter.PresentLicensedProducts(expiringLicenses)
	return nil
}

func (e ExpiringLicenses) validateConfig() error {
	matched, err := regexp.MatchString("^[1-9]+\\d*[dwmy]$", e.Options.ExpiresWithin)
	if err != nil {
		return err
	}

	if !matched {
		return errors.New("only d,w,m, or y are supported. Default is \"3m\"")
	}
	return nil
}
