package commands

import (
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type PreDeployCheck struct {
	service   preDeployCheckService
	presenter presenters.FormattedPresenter
	Options   struct {
		Check  bool   `long:"check" description:"Exit 1 if there are any pending changes. Useful for validating that Ops Manager is in a clean state."`
		Format string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

//go:generate counterfeiter -o ./fakes/pre_deploy_check_service.go --fake-name PreDeployCheckService . preDeployCheckService
type preDeployCheckService interface {
	ListPendingDirectorChanges() (api.PendingDirectorChangesOutput, error)
}

func NewPreDeployCheck(presenter presenters.FormattedPresenter, service preDeployCheckService) PreDeployCheck {
	return PreDeployCheck{
		service:   service,
		presenter: presenter,
	}
}

func (pc PreDeployCheck) Execute(args []string) error {
	if _, err := jhanda.Parse(&pc.Options, args); err != nil {
		return fmt.Errorf("could not parse pending-changes flags: %s", err)
	}

	pendingDirectorChanges, err := pc.service.ListPendingDirectorChanges()
	_ = err
	if !pendingDirectorChanges.EndpointResults.Complete {
		return fmt.Errorf("director configuration incomplete")
	}

	return nil
}

func (pc PreDeployCheck) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists all pending changes.",
		ShortDescription: "lists pending changes",
		Flags:            pc.Options,
	}
}
