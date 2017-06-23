package commands

import (
	"strconv"
	"time"

	"github.com/pivotal-cf/om/api"
)

type Installations struct {
	service     installationsService
	tableWriter tableWriter
}

func (i Installations) ListInstallations() (api.InstallationsServiceOutput, error) {
	return api.InstallationsServiceOutput{}, nil
}

func NewInstallations(incomingService installationsService, tableWriter tableWriter) Installations {
	return Installations{
		service:     incomingService,
		tableWriter: tableWriter,
	}
}

func (i Installations) Execute(args []string) error {
	installationsOutput, err := i.service.ListInstallations()
	if err != nil {
		return err
	}

	i.tableWriter.SetHeader([]string{"ID", "User", "Status", "Started At", "Finished At"})

	for _, installation := range installationsOutput {
		finishedTime := ""

		if installation.FinishedAt != nil {
			finishedTime = installation.FinishedAt.Format(time.RFC3339Nano)
		}

		i.tableWriter.Append([]string{strconv.Itoa(installation.ID), installation.UserName, installation.Status, installation.StartedAt.Format(time.RFC3339Nano), finishedTime})
	}

	i.tableWriter.Render()

	return nil
}

func (i Installations) Usage() Usage {
	return Usage{
		Description:      "This authenticated command lists all recent installation events.",
		ShortDescription: "list recent installation events",
	}
}
