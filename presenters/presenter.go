package presenters

import (
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
)

//go:generate counterfeiter -o ../commands/fakes/presenter.go --fake-name Presenter . Presenter

type Presenter interface {
	PresentAvailableProducts([]models.Product)
	PresentCertificateAuthorities([]api.CA)
	PresentCredentialReferences([]string)
	PresentCredentials(map[string]string)
	PresentDeployedProducts([]api.DiagnosticProduct)
	PresentErrands([]models.Errand)
	PresentCertificateAuthority(api.CA)
	PresentInstallations([]models.Installation)
	PresentPendingChanges([]api.ProductChange)
	PresentStagedProducts([]api.DiagnosticProduct)
}
