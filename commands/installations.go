package commands

import (
	"strconv"
	"time"

	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
)

type Installations struct {
	service   installationsService
	presenter presenter
}

func (i Installations) ListInstallations() (api.InstallationsServiceOutput, error) {
	return api.InstallationsServiceOutput{}, nil
}

func NewInstallations(incomingService installationsService, presenter presenter) Installations {
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

	installations := []models.Installation{}
	for _, installation := range installationsOutput {
		finishedTime := ""

		if installation.FinishedAt != nil {
			finishedTime = installation.FinishedAt.Format(time.RFC3339Nano)
		}

		installations = append(installations, models.Installation{
			Id:         strconv.Itoa(installation.ID),
			User:       installation.UserName,
			Status:     installation.Status,
			StartedAt:  installation.StartedAt.Format(time.RFC3339Nano),
			FinishedAt: finishedTime,
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
