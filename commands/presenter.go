package commands

import "github.com/pivotal-cf/om/models"

//go:generate counterfeiter -o ./fakes/presenter.go --fake-name Presenter . Presenter

type Presenter interface {
	PresentAvailableProducts([]models.Product)
	PresentCredentialReferences([]string)
	PresentErrands([]models.Errand)
	PresentInstallations([]models.Installation)
}
