package commands

import (
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
)

//go:generate counterfeiter -o ./fakes/presenter.go --fake-name Presenter . Presenter

type Presenter interface {
	PresentAvailableProducts([]models.Product)
	PresentCertificateAuthorities([]api.CA)
	PresentCredentialReferences([]string)
	PresentCredentials(map[string]string)
	PresentDeployedProducts([]api.DiagnosticProduct)
	PresentErrands([]models.Errand)
	PresentGeneratedCertificateAuthority(api.CA)
	PresentInstallations([]models.Installation)
	PresentPendingChanges([]api.ProductChange)
	PresentStagedProducts([]api.DiagnosticProduct)
}
