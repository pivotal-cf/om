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
	j.encodeJSON(&map[string][]models.Product{
		"available_products": products,
	})
}

func (j JSONPresenter) PresentCertificateAuthorities(certificateAuthorities []api.CA) {
	j.encodeJSON(&map[string][]api.CA{
		"certificate_authorities": certificateAuthorities,
	})
}

func (j JSONPresenter) PresentCredentialReferences(credentialReferences []string) {
	j.encodeJSON(&map[string][]string{
		"credential_references": credentialReferences,
	})
}

func (j JSONPresenter) PresentCredentials(credentials map[string]string) {
	j.encodeJSON(&map[string]map[string]string{
		"credential": credentials,
	})
}

func (j JSONPresenter) PresentDeployedProducts(deployedProducts []api.DiagnosticProduct) {
	j.encodeJSON(&map[string][]api.DiagnosticProduct{
		"deployed_products": deployedProducts,
	})
}

func (j JSONPresenter) PresentErrands(errands []models.Errand) {
	j.encodeJSON(&map[string][]models.Errand{
		"errands": errands,
	})
}

func (j JSONPresenter) PresentCertificateAuthority(certificateAuthority api.CA) {
	j.encodeJSON(&map[string]api.CA{
		"certificate_authority": certificateAuthority,
	})
}

func (j JSONPresenter) PresentInstallations(installations []models.Installation) {
	j.encodeJSON(&map[string][]models.Installation{
		"installations": installations,
	})
}

func (j JSONPresenter) PresentPendingChanges(pendingChanges []api.ProductChange) {
	j.encodeJSON(&map[string][]api.ProductChange{
		"pending_changes": pendingChanges,
	})
}

func (j JSONPresenter) PresentStagedProducts(stagedProducts []api.DiagnosticProduct) {
	j.encodeJSON(&map[string][]api.DiagnosticProduct{
		"staged_products": stagedProducts,
	})
}

func (j JSONPresenter) encodeJSON(v interface{}) {
	b, _ := json.MarshalIndent(&v, "", "  ")
	j.stdout.Write(b)
}
