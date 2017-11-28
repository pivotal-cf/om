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
	j.encodeJSON(&products)
}

func (j JSONPresenter) PresentCertificateAuthorities(certificateAuthorities []api.CA) {
	j.encodeJSON(&map[string][]api.CA{
		"certificate_authorities": certificateAuthorities,
	})
}

func (j JSONPresenter) PresentCredentialReferences(credentialReferences []string) {
	j.encodeJSON(&credentialReferences)
}

func (j JSONPresenter) PresentCredentials(credentials map[string]string) {
	j.encodeJSON(&credentials)
}

func (j JSONPresenter) PresentDeployedProducts(deployedProducts []api.DiagnosticProduct) {
	j.encodeJSON(&deployedProducts)
}

func (j JSONPresenter) PresentErrands(errands []models.Errand) {
	j.encodeJSON(&errands)
}

func (j JSONPresenter) PresentCertificateAuthority(certificateAuthority api.CA) {
	j.encodeJSON(&map[string]api.CA{
		"certificate_authority": certificateAuthority,
	})
}

func (j JSONPresenter) PresentInstallations(installations []models.Installation) {
	j.encodeJSON(&installations)
}

func (j JSONPresenter) PresentPendingChanges(pendingChanges []api.ProductChange) {
	j.encodeJSON(&pendingChanges)
}

func (j JSONPresenter) PresentStagedProducts(stagedProducts []api.DiagnosticProduct) {
	j.encodeJSON(&stagedProducts)
}

func (j JSONPresenter) encodeJSON(v interface{}) {
	encoder := json.NewEncoder(j.stdout)
	encoder.Encode(&v)
}
