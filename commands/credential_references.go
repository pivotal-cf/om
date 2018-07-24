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
	presenter presenters.FormattedPresenter
	logger    logger
	Options   struct {
		Product string `long:"product-name" short:"p" required:"true" description:"name of deployed product"`
		Format  string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

//go:generate counterfeiter -o ./fakes/credential_references_service.go --fake-name CredentialReferencesService . credentialReferencesService
type credentialReferencesService interface {
	ListDeployedProductCredentials(deployedProductGUID string) (api.CredentialReferencesOutput, error)
	ListDeployedProducts() ([]api.DeployedProductOutput, error)
}

func NewCredentialReferences(crService credentialReferencesService, presenter presenters.FormattedPresenter, logger logger) CredentialReferences {
	return CredentialReferences{service: crService, presenter: presenter, logger: logger}
}

func (cr CredentialReferences) Execute(args []string) error {
	if _, err := jhanda.Parse(&cr.Options, args); err != nil {
		return fmt.Errorf("could not parse credential-references flags: %s", err)
	}

	deployedProductGUID := ""
	deployedProducts, err := cr.service.ListDeployedProducts()
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

	output, err := cr.service.ListDeployedProductCredentials(deployedProductGUID)
	sort.Strings(output.Credentials)
	if err != nil {
		return fmt.Errorf("failed to list credential references: %s", err)
	}

	if len(output.Credentials) == 0 {
		cr.logger.Printf("no credential references found")
		return nil
	}

	cr.presenter.SetFormat(cr.Options.Format)
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
