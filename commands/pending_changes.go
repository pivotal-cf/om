package commands

import (
	"fmt"

	"github.com/pivotal-cf/om/api"
)

type PendingChanges struct {
	service     pendingChangesService
	tableWriter tableWriter
}

//go:generate counterfeiter -o ./fakes/pending_changes_service.go --fake-name PendingChangesService . pendingChangesService
type pendingChangesService interface {
	List() (api.PendingChangesOutput, error)
}

func NewPendingChanges(tableWriter tableWriter, service pendingChangesService) PendingChanges {
	return PendingChanges{
		service:     service,
		tableWriter: tableWriter,
	}
}

func (pc PendingChanges) Execute(args []string) error {
	output, err := pc.service.List()
	if err != nil {
		return fmt.Errorf("failed to retrieve pending changes %s", err)
	}

	pc.tableWriter.SetHeader([]string{"PRODUCT", "ACTION", "ERRANDS"})

	for _, change := range output.ChangeList {
		if len(change.Errands) == 0 {
			pc.tableWriter.Append([]string{change.Product, change.Action, ""})
		}
		for i, errand := range change.Errands {
			if i == 0 {
				pc.tableWriter.Append([]string{change.Product, change.Action, errand.Name})
			} else {
				pc.tableWriter.Append([]string{"", "", errand.Name})
			}
		}
	}

	pc.tableWriter.Render()
	return nil
}

func (pc PendingChanges) Usage() Usage {
	return Usage{
		Description:      "This authenticated command lists all pending changes.",
		ShortDescription: "lists pending changes",
	}
}
