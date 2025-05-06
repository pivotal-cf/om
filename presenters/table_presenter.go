package presenters

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
)

//counterfeiter:generate -o ./fakes/table_writer.go --fake-name TableWriter . tableWriter

type tableWriter interface {
	SetHeader([]string)
	Append([]string)
	SetAlignment(int)
	Render()
	SetAutoFormatHeaders(bool)
	SetAutoWrapText(bool)
}

type TablePresenter struct {
	tableWriter tableWriter
}

func NewTablePresenter(tableWriter tableWriter) TablePresenter {
	return TablePresenter{
		tableWriter: tableWriter,
	}
}

func (t TablePresenter) PresentAvailableProducts(products []models.Product) {
	t.tableWriter.SetAlignment(tablewriter.ALIGN_LEFT)
	t.tableWriter.SetHeader([]string{"Name", "Version"})

	for _, product := range products {
		t.tableWriter.Append([]string{product.Name, product.Version})
	}

	t.tableWriter.Render()
}

func (t TablePresenter) PresentCertificateAuthorities(certificateAuthorities []api.CA) {
	t.tableWriter.SetAutoWrapText(false)
	t.tableWriter.SetHeader([]string{"id", "issuer", "active", "created on", "expires on", "certicate pem"})

	for _, values := range certificateAuthorities {
		t.tableWriter.Append([]string{values.GUID, values.Issuer, strconv.FormatBool(values.Active), values.CreatedOn, values.ExpiresOn, values.CertPEM})
	}

	t.tableWriter.Render()
}

func (t TablePresenter) PresentCredentialReferences(credentialReferences []string) {
	t.tableWriter.SetAlignment(tablewriter.ALIGN_LEFT)
	t.tableWriter.SetHeader([]string{"Credentials"})

	for _, credential := range credentialReferences {
		t.tableWriter.Append([]string{credential})
	}

	t.tableWriter.Render()
}

func (t TablePresenter) PresentCredentials(credentials map[string]string) {
	t.tableWriter.SetAlignment(tablewriter.ALIGN_LEFT)

	header, credential := sortCredentialMap(credentials)

	t.tableWriter.SetAutoFormatHeaders(false)
	t.tableWriter.SetHeader(header)
	t.tableWriter.SetAutoWrapText(false)
	t.tableWriter.Append(credential)
	t.tableWriter.Render()
}

func (t TablePresenter) PresentDeployedProducts(deployedProducts []api.DiagnosticProduct) {
	t.tableWriter.SetHeader([]string{"Name", "Version"})

	for _, product := range deployedProducts {
		t.tableWriter.Append([]string{product.Name, product.Version})
	}

	t.tableWriter.Render()
}

func (t TablePresenter) PresentErrands(errands []models.Errand) {
	t.tableWriter.SetHeader([]string{"Name", "Post Deploy Enabled", "Pre Delete Enabled"})

	for _, errand := range errands {
		t.tableWriter.Append([]string{errand.Name, errand.PostDeployEnabled, errand.PreDeleteEnabled})
	}

	t.tableWriter.Render()
}

func (t TablePresenter) PresentCertificateAuthority(certificateAuthority api.CA) {
	t.tableWriter.SetAutoWrapText(false)
	t.tableWriter.SetHeader([]string{"id", "issuer", "active", "created on", "expires on", "certicate pem"})
	t.tableWriter.Append([]string{certificateAuthority.GUID,
		certificateAuthority.Issuer,
		strconv.FormatBool(certificateAuthority.Active),
		certificateAuthority.CreatedOn,
		certificateAuthority.ExpiresOn,
		certificateAuthority.CertPEM})
	t.tableWriter.Render()
}

func (t TablePresenter) PresentGenerateCAResponse(car api.GenerateCAResponse) {
	t.tableWriter.SetAutoWrapText(false)
	t.tableWriter.SetHeader([]string{"id", "issuer", "active", "created on", "expires on", "certicate pem", "warnings"})
	t.tableWriter.Append([]string{car.GUID,
		car.Issuer,
		strconv.FormatBool(car.Active),
		car.CreatedOn,
		car.ExpiresOn,
		car.CertPEM,
		strings.Join(car.Warnings, ";"),
	})
	t.tableWriter.Render()
}

func (t TablePresenter) PresentSSLCertificate(certificate api.SSLCertificate) {
	t.tableWriter.SetAutoWrapText(false)
	t.tableWriter.SetHeader([]string{"Certificate"})
	t.tableWriter.Append([]string{certificate.Certificate})
	t.tableWriter.Render()
}

func (t TablePresenter) PresentInstallations(installations []models.Installation) {
	t.tableWriter.SetHeader([]string{"ID", "User", "Status", "Started At", "Finished At"})

	for _, installation := range installations {
		var startedAt, finishedAt string
		if installation.StartedAt != nil {
			startedAt = installation.StartedAt.Format(time.RFC3339Nano)
		}

		if installation.FinishedAt != nil {
			finishedAt = installation.FinishedAt.Format(time.RFC3339Nano)
		}

		t.tableWriter.Append([]string{
			strconv.Itoa(installation.Id),
			installation.User,
			installation.Status,
			startedAt,
			finishedAt,
		})
	}

	t.tableWriter.Render()
}

func (t TablePresenter) PresentPendingChanges(output api.PendingChangesOutput) {
	pendingChanges := output.ChangeList

	t.tableWriter.SetHeader([]string{"PRODUCT", "ACTION", "ERRANDS"})

	for _, change := range pendingChanges {
		if len(change.Errands) == 0 {
			t.tableWriter.Append([]string{change.GUID, change.Action, ""})
		}
		for i, errand := range change.Errands {
			if i == 0 {
				t.tableWriter.Append([]string{change.GUID, change.Action, errand.Name})
			} else {
				t.tableWriter.Append([]string{"", "", errand.Name})
			}
		}
	}

	t.tableWriter.Render()
}

func (t TablePresenter) PresentProducts(products models.ProductsVersionsDisplay) {
	var columns []string
	if products.Available {
		columns = append(columns, "Available")
	}

	if products.Staged {
		columns = append(columns, "Staged")
	}

	if products.Deployed {
		columns = append(columns, "Deployed")
	}

	t.tableWriter.SetAlignment(tablewriter.ALIGN_LEFT)
	t.tableWriter.SetHeader(append([]string{"Name"}, columns...))

	for _, product := range products.ProductVersions {
		moreAvailableProducts, firstProductLine := true, true
		hasData := false

		for moreAvailableProducts {
			var outputData []string

			moreAvailableProducts = false
			if products.Available {
				if len(product.Available) > 0 {
					hasData = true
					outputData = append(outputData, product.Available[0])
					moreAvailableProducts = true
					product.Available = product.Available[1:]
				} else if firstProductLine {
					outputData = append(outputData, "")
				} else {
					break
				}
			}

			if products.Staged {
				if product.Staged != "" && firstProductLine {
					hasData = true
					outputData = append(outputData, product.Staged)
				} else {
					outputData = append(outputData, "")
				}
			}

			if products.Deployed {
				if product.Deployed != "" && firstProductLine {
					hasData = true
					outputData = append(outputData, product.Deployed)
				} else {
					outputData = append(outputData, "")
				}
			}

			if len(outputData) != 0 && hasData {

				productName := product.Name
				if !firstProductLine {
					productName = ""
				}

				t.tableWriter.Append(append([]string{productName}, outputData...))

				firstProductLine = false
			}
		}
	}

	t.tableWriter.Render()
}

func (t TablePresenter) PresentStagedProducts(stagedProducts []api.DiagnosticProduct) {
	t.tableWriter.SetHeader([]string{"Name", "Version"})

	for _, product := range stagedProducts {
		t.tableWriter.Append([]string{product.Name, product.Version})
	}

	t.tableWriter.Render()
}

func (t TablePresenter) PresentLicensedProducts(products []api.ExpiringLicenseOutput) {
	t.tableWriter.SetHeader([]string{"Name", "GUID", "Product Version", "State", "Licensed Version", "Expiry"})
	var expiresAt string
	for _, product := range products {
		if product.ExpiresAt.IsZero() {
			expiresAt = ""
		} else {
			expiresAt = product.ExpiresAt.Format("2006-01-02")
		}
		states := strings.Join(product.ProductState, ", ")
		t.tableWriter.Append([]string{product.ProductName, product.GUID, product.ProductVersion, states, product.LicenseVersion, expiresAt})
	}

	t.tableWriter.Render()
}

func sortCredentialMap(cm map[string]string) ([]string, []string) {
	var header []string
	var credential []string

	key := make([]string, len(cm))
	i := 0

	for k := range cm {
		key[i] = k
		i++
	}
	sort.Strings(key)
	for i := 0; i < len(key); i++ {
		header = append(header, key[i])
		credential = append(credential, cm[key[i]])
	}

	return header, credential
}

func (t TablePresenter) PresentDiagnosticReport(api.DiagnosticReport) {

}
