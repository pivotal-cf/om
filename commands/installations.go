package commands

import (
	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
)

type Installations struct {
	service   installationsService
	presenter Presenter
}

func (i Installations) ListInstallations() (api.InstallationsServiceOutput, error) {
	return api.InstallationsServiceOutput{}, nil
}

func NewInstallations(incomingService installationsService, presenter Presenter) Installations {
	return Installations{
		service:   incomingService,
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

func (i Installations) Usage() commands.Usage {
	return commands.Usage{
		Description:      "This authenticated command lists all recent installation events.",
		ShortDescription: "list recent installation events",
	}
}
