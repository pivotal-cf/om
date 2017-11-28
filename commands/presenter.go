package commands

import (
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
)

//go:generate counterfeiter -o ./fakes/presenter.go --fake-name Presenter . Presenter

type Presenter interface {
	PresentAvailableProducts([]models.Product)
	PresentCredentialReferences([]string)
	PresentCredentials(map[string]string)
	PresentDeployedProducts([]api.DiagnosticProduct)
	PresentErrands([]models.Errand)
	PresentInstallations([]models.Installation)
	PresentPendingChanges([]api.ProductChange)
	PresentStagedProducts([]api.DiagnosticProduct)
}
