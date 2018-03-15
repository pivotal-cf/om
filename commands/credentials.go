package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type Credentials struct {
	service   credentialsService
	lister    deployedProductsLister
	presenter presenters.Presenter
	logger    logger
	Options   struct {
		Product             string `long:"product-name"         short:"p" required:"true" description:"name of deployed product"`
		CredentialReference string `long:"credential-reference" short:"c" required:"true" description:"name of credential reference"`
		CredentialField     string `long:"credential-field"     short:"f"                 description:"single credential field to output"`
	}
}

//go:generate counterfeiter -o ./fakes/credentials_service.go --fake-name CredentialsService . credentialsService
type credentialsService interface {
	Fetch(deployedProductGUID, credentialReference string) (api.CredentialOutput, error)
}

func NewCredentials(csService credentialsService, dpLister deployedProductsLister, presenter presenters.Presenter, logger logger) Credentials {
	return Credentials{service: csService, lister: dpLister, presenter: presenter, logger: logger}
}

func (cs Credentials) Execute(args []string) error {
	if _, err := jhanda.Parse(&cs.Options, args); err != nil {
		return fmt.Errorf("could not parse credential-references flags: %s", err)
	}

	deployedProductGUID := ""
	deployedProducts, err := cs.lister.List()
	if err != nil {
		return fmt.Errorf("failed to fetch credential: %s", err)
	}
	for _, deployedProduct := range deployedProducts {
		if deployedProduct.Type == cs.Options.Product {
			deployedProductGUID = deployedProduct.GUID
			break
		}
	}

	if deployedProductGUID == "" {
		return fmt.Errorf("failed to fetch credential: %q is not deployed", cs.Options.Product)
	}

	output, err := cs.service.Fetch(deployedProductGUID, cs.Options.CredentialReference)
	if err != nil {
		return fmt.Errorf("failed to fetch credential for %q: %s", cs.Options.CredentialReference, err)
	}

	if len(output.Credential.Value) == 0 {
		return fmt.Errorf("failed to fetch credential for %q", cs.Options.CredentialReference)
	}

	if cs.Options.CredentialField == "" {
		cs.presenter.PresentCredentials(output.Credential.Value)
	} else {
		if value, ok := output.Credential.Value[cs.Options.CredentialField]; ok {
			cs.logger.Println(value)
		} else {
			return fmt.Errorf("credential field %q not found", cs.Options.CredentialField)
		}
	}

	return nil
}

func (cs Credentials) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command fetches credentials for deployed products.",
		ShortDescription: "fetch credentials for a deployed product",
		Flags:            cs.Options,
	}
}
