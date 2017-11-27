package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/jhanda/flags"
	"github.com/pivotal-cf/om/api"
)

type Credentials struct {
	service   credentialsService
	lister    deployedProductsLister
	presenter Presenter
	logger    logger
	Options   struct {
		Product             string `short:"p"  long:"product-name"  description:"name of deployed product"`
		CredentialReference string `short:"c"  long:"credential-reference"  description:"name of credential reference"`
		CredentialField     string `short:"f"  long:"credential-field"  description:"single credential field to output"`
	}
}

//go:generate counterfeiter -o ./fakes/credentials_service.go --fake-name CredentialsService . credentialsService
type credentialsService interface {
	Fetch(deployedProductGUID, credentialReference string) (api.CredentialOutput, error)
}

func NewCredentials(csService credentialsService, dpLister deployedProductsLister, presenter Presenter, logger logger) Credentials {
	return Credentials{service: csService, lister: dpLister, presenter: presenter, logger: logger}
}

func (cs Credentials) Execute(args []string) error {
	_, err := flags.Parse(&cs.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse credential-references flags: %s", err)
	}

	if cs.Options.Product == "" {
		return errors.New("error: product-name is missing. Please see usage for more information.")
	}

	if cs.Options.CredentialReference == "" {
		return errors.New("error: credential-reference is missing. Please see usage for more information.")
	}

	deployedProductGUID := ""
	deployedProducts, err := cs.lister.DeployedProducts()
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

func (cs Credentials) Usage() commands.Usage {
	return commands.Usage{
		Description:      "This authenticated command fetches credentials for deployed products.",
		ShortDescription: "fetch credentials for a deployed product",
		Flags:            cs.Options,
	}
}
