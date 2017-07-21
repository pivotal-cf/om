package commands

import (
	"errors"
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

type CredentialReferences struct {
	service     credentialReferencesService
	lister      deployedProductsLister
	tableWriter tableWriter
	logger      logger
	Options     struct {
		Product string `short:"p"  long:"product-name"  description:"name of deployed product"`
	}
}

//go:generate counterfeiter -o ./fakes/credential_references_service.go --fake-name CredentialReferencesService . credentialReferencesService
type credentialReferencesService interface {
	List(deployedProductGUID string) (api.CredentialReferencesOutput, error)
}

func NewCredentialReferences(crService credentialReferencesService, dpLister deployedProductsLister, tableWriter tableWriter, logger logger) CredentialReferences {
	return CredentialReferences{service: crService, lister: dpLister, tableWriter: tableWriter, logger: logger}
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
	deployedProducts, _ := cr.lister.DeployedProducts()
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
	if err != nil {
		return err
	}

	if len(output.Credentials) == 0 {
		cr.logger.Printf("no credential references found")
		return nil
	}

	cr.tableWriter.SetAlignment(tablewriter.ALIGN_LEFT)
	cr.tableWriter.SetHeader([]string{"Credentials"})

	for _, credential := range output.Credentials {
		cr.tableWriter.Append([]string{credential})
	}

	cr.tableWriter.Render()

	return nil
}

func (cr CredentialReferences) Usage() Usage {
	return Usage{
		Description:      "This authenticated command lists credential references for deployed products.",
		ShortDescription: "list credential references for a deployed product",
		Flags:            cr.Options,
	}
}
