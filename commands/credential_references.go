package commands

import (
	"fmt"
	"sort"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type CredentialReferences struct {
	service   credentialReferencesService
	lister    deployedProductsLister
	presenter presenters.Presenter
	logger    logger
	Options   struct {
		Product string `long:"product-name" short:"p" required:"true" description:"name of deployed product"`
	}
}

//go:generate counterfeiter -o ./fakes/credential_references_service.go --fake-name CredentialReferencesService . credentialReferencesService
type credentialReferencesService interface {
	List(deployedProductGUID string) (api.CredentialReferencesOutput, error)
}

func NewCredentialReferences(crService credentialReferencesService, dpLister deployedProductsLister, presenter presenters.Presenter, logger logger) CredentialReferences {
	return CredentialReferences{service: crService, lister: dpLister, presenter: presenter, logger: logger}
}

func (cr CredentialReferences) Execute(args []string) error {
	if _, err := jhanda.Parse(&cr.Options, args); err != nil {
		return fmt.Errorf("could not parse credential-references flags: %s", err)
	}

	deployedProductGUID := ""
	deployedProducts, err := cr.lister.List()
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

	cr.presenter.PresentCredentialReferences(output.Credentials)

	return nil
}

func (cr CredentialReferences) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists credential references for deployed products.",
		ShortDescription: "list credential references for a deployed product",
		Flags:            cr.Options,
	}
}
