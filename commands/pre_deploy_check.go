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
		Check  bool   `long:"check" description:"Exit 1 if there are any pending changes. Useful for validating that Ops Manager is in a clean state."`
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
	failedCompleteness := fmt.Errorf("OpsManager is not fully configured")
	if _, err := jhanda.Parse(&pc.Options, args); err != nil {
		return fmt.Errorf("could not parse pending-changes flags: %s", err)
	}

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
		pc.logger.Println("The director is not configured correctly.")
		return failedCompleteness
	}

	pendingProductChanges, err := pc.service.ListAllPendingProductChanges()
	if err != nil {
		return fmt.Errorf("while getting products: %s", err)
	}

	var productsIncomplete []string
	for _, change := range pendingProductChanges {
		if !change.EndpointResults.Complete {
			productsIncomplete = append(productsIncomplete, change.EndpointResults.Identifier)
		}
	}

	if len(productsIncomplete) > 0 {
		pc.logger.Println("The director is configured correctly, but the following product(s) are not.")
		for _, incomplete := range productsIncomplete {
			pc.logger.Printf(color.RedString("[X] %s", incomplete))
		}
		return failedCompleteness
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
