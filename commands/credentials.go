package commands

import (
	"errors"
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

type Credentials struct {
	service     credentialsService
	lister      deployedProductsLister
	tableWriter tableWriter
	logger      logger
	Options     struct {
		Product             string `short:"p"  long:"product-name"  description:"name of deployed product"`
		CredentialReference string `short:"c"  long:"credential-reference"  description:"name of credential reference"`
	}
}

//go:generate counterfeiter -o ./fakes/credentials_service.go --fake-name CredentialsService . credentialsService
type credentialsService interface {
	Fetch(deployedProductGUID, credentialReference string) (api.CredentialOutput, error)
}

func NewCredentials(csService credentialsService, dpLister deployedProductsLister, tableWriter tableWriter, logger logger) Credentials {
	return Credentials{service: csService, lister: dpLister, tableWriter: tableWriter, logger: logger}
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
	deployedProducts, _ := cs.lister.DeployedProducts()
	for _, deployedProduct := range deployedProducts {
		if deployedProduct.Type == cs.Options.Product {
			deployedProductGUID = deployedProduct.GUID
			break
		}
	}

	if deployedProductGUID == "" {
		return fmt.Errorf("failed to list credential references: %s is not deployed", cs.Options.Product)
	}

	output, err := cs.service.Fetch(deployedProductGUID, cs.Options.CredentialReference)
	if err != nil {
		return err
	}

	if len(output.Credential.Value) == 0 {
		return fmt.Errorf("failed to fetch credential for: %s", cs.Options.CredentialReference)
	}

	cs.tableWriter.SetAlignment(tablewriter.ALIGN_LEFT)
	var header []string
	var credential []string

	for k, v := range output.Credential.Value {
		header = append(header, k)
		credential = append(credential, v)
	}
	cs.tableWriter.SetHeader(header)
	cs.tableWriter.Append(credential)
	cs.tableWriter.Render()

	return nil
}

func (cs Credentials) Usage() Usage {
	return Usage{
		Description:      "This authenticated command fetches credentials for deployed products.",
		ShortDescription: "fetch credentials for a deployed product",
		Flags:            cs.Options,
	}
}
