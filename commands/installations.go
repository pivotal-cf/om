package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
	"github.com/pivotal-cf/om/presenters"
)

type Installations struct {
	service   installationsService
	presenter presenters.FormattedPresenter
	Options   struct {
		Format string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

//go:generate counterfeiter -o ./fakes/installations_service.go --fake-name InstallationsService . installationsService
type installationsService interface {
	ListInstallations() ([]api.InstallationsServiceOutput, error)
}

func NewInstallations(service installationsService, presenter presenters.FormattedPresenter) Installations {
	return Installations{
		service:   service,
		presenter: presenter,
	}
}

func (i Installations) Execute(args []string) error {
	if _, err := jhanda.Parse(&i.Options, args); err != nil {
		return fmt.Errorf("could not parse installations flags: %s", err)
	}

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

	i.presenter.SetFormat(i.Options.Format)
	i.presenter.PresentInstallations(installations)

	return nil
}

func (i Installations) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists all recent installation events.",
		ShortDescription: "list recent installation events",
		Flags:            i.Options,
	}
}
