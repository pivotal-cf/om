package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type Credentials struct {
	service   credentialsService
	presenter presenters.FormattedPresenter
	logger    logger
	Options   struct {
		Product             string `long:"product-name"         short:"p" required:"true" description:"name of deployed product"`
		CredentialReference string `long:"credential-reference" short:"c" required:"true" description:"name of credential reference"`
		CredentialField     string `long:"credential-field"     short:"f"                 description:"single credential field to output"`
		Format              string `long:"format"               short:"t" default:"table" description:"Format to print as (options: table,json)"`
	}
}

//go:generate counterfeiter -o ./fakes/credentials_service.go --fake-name CredentialsService . credentialsService
type credentialsService interface {
	GetDeployedProductCredential(api.GetDeployedProductCredentialInput) (api.GetDeployedProductCredentialOutput, error)
	ListDeployedProducts() ([]api.DeployedProductOutput, error)
}

func NewCredentials(csService credentialsService, presenter presenters.FormattedPresenter, logger logger) Credentials {
	return Credentials{service: csService, presenter: presenter, logger: logger}
}

func (cs Credentials) Execute(args []string) error {
	if _, err := jhanda.Parse(&cs.Options, args); err != nil {
		return fmt.Errorf("could not parse credential-references flags: %s", err)
	}

	deployedProductGUID := ""
	deployedProducts, err := cs.service.ListDeployedProducts()
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

	output, err := cs.service.GetDeployedProductCredential(api.GetDeployedProductCredentialInput{
		DeployedGUID:        deployedProductGUID,
		CredentialReference: cs.Options.CredentialReference,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch credential for %q: %s", cs.Options.CredentialReference, err)
	}

	if len(output.Credential.Value) == 0 {
		return fmt.Errorf("failed to fetch credential for %q", cs.Options.CredentialReference)
	}

	if cs.Options.CredentialField == "" {
		cs.presenter.SetFormat(cs.Options.Format)
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
