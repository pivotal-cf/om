package commands

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type PreDeployCheck struct {
	service   preDeployCheckService
	presenter presenters.FormattedPresenter
	logger    logger
	Options   struct {
		Format string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

//go:generate counterfeiter -o ./fakes/pre_deploy_check_service.go --fake-name PreDeployCheckService . preDeployCheckService
type preDeployCheckService interface {
	Info() (api.Info, error)
	ListPendingDirectorChanges() (api.PendingDirectorChangesOutput, error)
	ListAllPendingProductChanges() ([]api.PendingProductChangesOutput, error)
}

func NewPreDeployCheck(presenter presenters.FormattedPresenter, service preDeployCheckService, logger logger) PreDeployCheck {
	return PreDeployCheck{
		service:   service,
		presenter: presenter,
		logger:    logger,
	}
}

func (pc PreDeployCheck) Execute(args []string) error {
	var isError = false

	if _, err := jhanda.Parse(&pc.Options, args); err != nil {
		return fmt.Errorf("could not parse pending-changes flags: %s", err)
	}

	pc.logger.Println("Scanning OpsManager now ...")

	info, err := pc.service.Info()
	if err != nil {
		return err
	}
	if ok, _ := info.VersionAtLeast(2, 6); !ok {
		return fmt.Errorf("pre deploy checks are only supported in OpsManager 2.6")
	}

	pendingDirectorChanges, err := pc.service.ListPendingDirectorChanges()
	if err != nil {
		return fmt.Errorf("while getting director: %s", err)
	}

	directorOk := pendingDirectorChanges.EndpointResults.Complete
	if !directorOk {
		pc.logger.Printf(color.RedString("[X] director: %s", pendingDirectorChanges.EndpointResults.Identifier))
		isError = true
	} else {
		pc.logger.Printf(color.GreenString("[✓] director: %s", pendingDirectorChanges.EndpointResults.Identifier))
	}

	pendingProductChanges, err := pc.service.ListAllPendingProductChanges()
	if err != nil {
		return fmt.Errorf("while getting products: %s", err)
	}

	for _, change := range pendingProductChanges {
		if change.EndpointResults.Identifier == pendingDirectorChanges.EndpointResults.Identifier {
			continue
		}

		if !change.EndpointResults.Complete {
			pc.logger.Printf(color.RedString("[X] product: %s", change.EndpointResults.Identifier))
			isError = true
		} else {
			pc.logger.Printf(color.GreenString("[✓] product: %s", change.EndpointResults.Identifier))
		}
	}

	if isError {
		return fmt.Errorf("OpsManager is not fully configured")
	}

	pc.logger.Println("The director and products are configured correctly.")
	return nil
}

func (pc PreDeployCheck) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "**EXPERIMENTAL** This authenticated command lists all pending changes.",
		ShortDescription: "**EXPERIMENTAL** lists pending changes",
		Flags:            pc.Options,
	}
}
