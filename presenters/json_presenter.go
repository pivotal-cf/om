package presenters

import (
	"encoding/json"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
	"io"
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
	j.encodeJSON(products)
}

func (j JSONPresenter) PresentCertificateAuthorities(certificateAuthorities []api.CA) {
	j.encodeJSON(certificateAuthorities)
}

func (j JSONPresenter) PresentCredentialReferences(credentialReferences []string) {
	j.encodeJSON(credentialReferences)
}

func (j JSONPresenter) PresentCredentials(credentials map[string]string) {
	j.encodeJSON(credentials)
}

func (j JSONPresenter) PresentDeployedProducts(deployedProducts []api.DiagnosticProduct) {
	j.encodeJSON(deployedProducts)
}

func (j JSONPresenter) PresentErrands(errands []models.Errand) {
	j.encodeJSON(errands)
}

func (j JSONPresenter) PresentCertificateAuthority(certificateAuthority api.CA) {
	j.encodeJSON(certificateAuthority)
}

func (j JSONPresenter) PresentGenerateCAResponse(gcar api.GenerateCAResponse) {
	j.encodeJSON(gcar)
}

func (j JSONPresenter) PresentSSLCertificate(certificate api.SSLCertificate) {
	j.encodeJSON(certificate)
}

func (j JSONPresenter) PresentInstallations(installations []models.Installation) {
	j.encodeJSON(installations)
}

func (j JSONPresenter) PresentStagedProducts(stagedProducts []api.DiagnosticProduct) {
	j.encodeJSON(stagedProducts)
}

func (j JSONPresenter) PresentPendingChanges(pendingChangesOutput api.PendingChangesOutput) {
	_, _ = j.stdout.Write([]byte(pendingChangesOutput.FullReport))
}

func (j JSONPresenter) PresentProducts(products models.ProductsVersionsDisplay) {
	var output []models.ProductVersions
	for _, product := range products.ProductVersions {
		hasData := false

		productVersions := models.ProductVersions{
			Name: product.Name,
		}

		if products.Available {
			if len(product.Available) > 0 {
				hasData = true
				productVersions.Available = product.Available
			}
		}

		if products.Staged {
			if product.Staged != "" {
				hasData = true
				productVersions.Staged = product.Staged
			}
		}

		if products.Deployed {
			if product.Deployed != "" {
				hasData = true
				productVersions.Deployed = product.Deployed
			}
		}

		if hasData {
			output = append(output, productVersions)
		}
	}

	j.encodeJSON(output)
}

func (j JSONPresenter) PresentDiagnosticReport(report api.DiagnosticReport) {
	_, _ = j.stdout.Write([]byte(report.FullReport))
}

func (j JSONPresenter) PresentLicensedProducts(products []api.ExpiringLicenseOutput) {
	type licenseOutput struct {
		ProductName    string `json:"product_name"`
		GUID           string `json:"guid"`
		ExpiresAt      string `json:"expires_at"`
		ProductState   string `json:"product_state"`
		ProductVersion string `json:"product_version"`
		LicenseVersion string `json:"licensed_version"`
	}
	var output []licenseOutput
	var expiresAt string
	for _, product := range products {

		if product.ExpiresAt.IsZero() {
			expiresAt = ""
		} else {
			expiresAt = product.ExpiresAt.Format("2006-01-02")
		}
		output = append(output, licenseOutput{ProductName: product.ProductName,
			GUID: product.GUID, ExpiresAt: expiresAt,
			ProductState: product.ProductState, ProductVersion: product.ProductVersion,
			LicenseVersion: product.LicenseVersion})
	}
	j.encodeJSON(output)
}

func (j JSONPresenter) encodeJSON(v interface{}) {
	b, _ := json.MarshalIndent(&v, "", "  ")

	_, _ = j.stdout.Write(b)
	_, _ = j.stdout.Write([]byte("\n"))
}
