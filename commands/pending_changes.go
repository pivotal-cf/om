package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type PendingChanges struct {
	service   pendingChangesService
	presenter presenters.FormattedPresenter
	Options   struct {
		Format string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

//go:generate counterfeiter -o ./fakes/pending_changes_service.go --fake-name PendingChangesService . pendingChangesService
type pendingChangesService interface {
	ListStagedPendingChanges() (api.PendingChangesOutput, error)
}

func NewPendingChanges(presenter presenters.FormattedPresenter, service pendingChangesService) PendingChanges {
	return PendingChanges{
		service:   service,
		presenter: presenter,
	}
}

func (pc PendingChanges) Execute(args []string) error {
	if _, err := jhanda.Parse(&pc.Options, args); err != nil {
		return fmt.Errorf("could not parse pending-changes flags: %s", err)
	}

	output, err := pc.service.ListStagedPendingChanges()
	if err != nil {
		return fmt.Errorf("failed to retrieve pending changes %s", err)
	}

	pc.presenter.SetFormat(pc.Options.Format)
	pc.presenter.PresentPendingChanges(output.ChangeList)

	return nil
}

func (pc PendingChanges) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists all pending changes.",
		ShortDescription: "lists pending changes",
	}
}
