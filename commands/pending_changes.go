package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type PendingChanges struct {
	service   pendingChangesService
	presenter presenters.Presenter
}

//go:generate counterfeiter -o ./fakes/pending_changes_service.go --fake-name PendingChangesService . pendingChangesService
type pendingChangesService interface {
	List() (api.PendingChangesOutput, error)
}

func NewPendingChanges(presenter presenters.Presenter, service pendingChangesService) PendingChanges {
	return PendingChanges{
		service:   service,
		presenter: presenter,
	}
}

func (pc PendingChanges) Execute(args []string) error {
	output, err := pc.service.List()
	if err != nil {
		return fmt.Errorf("failed to retrieve pending changes %s", err)
	}

	pc.presenter.PresentPendingChanges(output.ChangeList)
	return nil
}

func (pc PendingChanges) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists all pending changes.",
		ShortDescription: "lists pending changes",
	}
}
