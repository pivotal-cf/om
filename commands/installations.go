package commands

import (
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
	"github.com/pivotal-cf/om/presenters"
)

type Installations struct {
	service   installationsService
	presenter presenters.Presenter
}

//go:generate counterfeiter -o ./fakes/installations_service.go --fake-name InstallationsService . installationsService
type installationsService interface {
	ListInstallations() ([]api.InstallationsServiceOutput, error)
}

func NewInstallations(service installationsService, presenter presenters.Presenter) Installations {
	return Installations{
		service:   service,
		presenter: presenter,
	}
}

func (i Installations) Execute(args []string) error {
	installationsOutput, err := i.service.ListInstallations()
	if err != nil {
		return err
	}

	var installations []models.Installation
	for _, installation := range installationsOutput {
		installations = append(installations, models.Installation{
			Id:         installation.ID,
			User:       installation.UserName,
			Status:     installation.Status,
			StartedAt:  installation.StartedAt,
			FinishedAt: installation.FinishedAt,
		})
	}

	i.presenter.PresentInstallations(installations)

	return nil
}

func (i Installations) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists all recent installation events.",
		ShortDescription: "list recent installation events",
	}
}
