package commands

import (
	"errors"
	"fmt"
	"sort"

	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/jhanda/flags"
	"github.com/pivotal-cf/om/api"
)

type CredentialReferences struct {
	service   credentialReferencesService
	lister    deployedProductsLister
	presenter Presenter
	logger    logger
	Options   struct {
		Product string `short:"p"  long:"product-name"  description:"name of deployed product"`
	}
}

//go:generate counterfeiter -o ./fakes/credential_references_service.go --fake-name CredentialReferencesService . credentialReferencesService
type credentialReferencesService interface {
	List(deployedProductGUID string) (api.CredentialReferencesOutput, error)
}

func NewCredentialReferences(crService credentialReferencesService, dpLister deployedProductsLister, presenter Presenter, logger logger) CredentialReferences {
	return CredentialReferences{service: crService, lister: dpLister, presenter: presenter, logger: logger}
}

func (cr CredentialReferences) Execute(args []string) error {
	_, err := flags.Parse(&cr.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse credential-references flags: %s", err)
	}

	if cr.Options.Product == "" {
		return errors.New("error: product-name is missing. Please see usage for more information.")
	}

	deployedProductGUID := ""
	deployedProducts, err := cr.lister.DeployedProducts()
	if err != nil {
		return fmt.Errorf("failed to list credential references: %s", err)
	}
	for _, deployedProduct := range deployedProducts {
		if deployedProduct.Type == cr.Options.Product {
			deployedProductGUID = deployedProduct.GUID
			break
		}
	}

	if deployedProductGUID == "" {
		return fmt.Errorf("failed to list credential references: %s is not deployed", cr.Options.Product)
	}

	output, err := cr.service.List(deployedProductGUID)
	sort.Strings(output.Credentials)
	if err != nil {
		return fmt.Errorf("failed to list credential references: %s", err)
	}

	if len(output.Credentials) == 0 {
		cr.logger.Printf("no credential references found")
		return nil
	}

	cr.presenter.PresentCredentials(output.Credentials)

	return nil
}

func (cr CredentialReferences) Usage() commands.Usage {
	return commands.Usage{
		Description:      "This authenticated command lists credential references for deployed products.",
		ShortDescription: "list credential references for a deployed product",
		Flags:            cr.Options,
	}
}
