package commands

import (
	"fmt"

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

//counterfeiter:generate -o ./fakes/credentials_service.go --fake-name CredentialsService . credentialsService
type credentialsService interface {
	GetDeployedProductCredential(api.GetDeployedProductCredentialInput) (api.GetDeployedProductCredentialOutput, error)
	ListDeployedProducts() ([]api.DeployedProductOutput, error)
}

func NewCredentials(csService credentialsService, presenter presenters.FormattedPresenter, logger logger) *Credentials {
	return &Credentials{service: csService, presenter: presenter, logger: logger}
}

func (cs Credentials) Execute(args []string) error {
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

	if output.Credential.Value == nil {
		return fmt.Errorf("failed to fetch credential for %q", cs.Options.CredentialReference)
	}
	if valMap, ok := output.Credential.Value.(map[string]interface{}); ok && len(valMap) == 0 {
		return fmt.Errorf("failed to fetch credential for %q", cs.Options.CredentialReference)
	}

	if cs.Options.CredentialField == "" {
		cs.presenter.SetFormat(cs.Options.Format)
		cs.presenter.PresentCredentials(output.Credential.Value)
		return nil
	}

	valMap, ok := output.Credential.Value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("credential is not a map, cannot lookup field %q", cs.Options.CredentialField)
	}

	value, ok := valMap[cs.Options.CredentialField]
	if !ok {
		return fmt.Errorf("credential field %q not found", cs.Options.CredentialField)
	}

	if vStr, ok := value.(string); ok {
		cs.logger.Println(vStr)
	} else {
		cs.presenter.SetFormat(cs.Options.Format)
		cs.presenter.PresentCredentials(value)
	}

	return nil
}
