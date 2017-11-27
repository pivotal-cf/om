package presenters

import (
	"encoding/json"
	"io"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
)

type JSONPresenter struct {
	stdout io.Writer
}

func NewJSONPresenter(stdout io.Writer) JSONPresenter {
	return JSONPresenter{
		stdout: stdout,
	}
}

func (j JSONPresenter) PresentAvailableProducts(products []models.Product) {
	encoder := json.NewEncoder(j.stdout)
	encoder.Encode(&products)
}

func (j JSONPresenter) PresentCredentialReferences(credentials []string) {
	encoder := json.NewEncoder(j.stdout)
	encoder.Encode(&credentials)
}

func (j JSONPresenter) PresentCredentials(credentials map[string]string) {
	encoder := json.NewEncoder(j.stdout)
	encoder.Encode(&credentials)
}

func (j JSONPresenter) PresentDeployedProducts(deployedProducts []api.DiagnosticProduct) {
	encoder := json.NewEncoder(j.stdout)
	encoder.Encode(&deployedProducts)
}

func (j JSONPresenter) PresentErrands(errands []models.Errand) {
	encoder := json.NewEncoder(j.stdout)
	encoder.Encode(&errands)
}

func (j JSONPresenter) PresentInstallations(installations []models.Installation) {
	encoder := json.NewEncoder(j.stdout)
	encoder.Encode(&installations)
}

func (j JSONPresenter) PresentStagedProducts(stagedProducts []api.DiagnosticProduct) {
	encoder := json.NewEncoder(j.stdout)
	encoder.Encode(&stagedProducts)
}
