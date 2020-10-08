package commands

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type PendingChanges struct {
	service   pendingChangesService
	presenter presenters.FormattedPresenter
	Options   struct {
		Check  bool   `long:"check" description:"Exit 1 if there are any pending changes. Useful for validating that Ops Manager is in a clean state."`
		Format string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
	logger logger
}

//counterfeiter:generate -o ./fakes/pending_changes_service.go --fake-name PendingChangesService . pendingChangesService
type pendingChangesService interface {
	ListStagedPendingChanges() (api.PendingChangesOutput, error)
}

func NewPendingChanges(presenter presenters.FormattedPresenter, service pendingChangesService, logger logger) PendingChanges {
	return PendingChanges{
		service:   service,
		presenter: presenter,
		logger:    logger,
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
	pc.presenter.PresentPendingChanges(output)

	var errs []string
	for _, change := range output.ChangeList {
		if change.CompletenessChecks != nil {
			if !change.CompletenessChecks.ConfigurationComplete {
				errs = append(errs, fmt.Sprintf("configuration is incomplete for guid %s", change.GUID))
			}

			if !change.CompletenessChecks.StemcellPresent {
				errs = append(errs, fmt.Sprintf("stemcell is missing for one or more products for guid %s", change.GUID))
			}

			if !change.CompletenessChecks.ConfigurablePropertiesValid {
				errs = append(errs, fmt.Sprintf("one or more properties are invalid for guid %s", change.GUID))
			}
		}
	}

	for _, ProductChange := range output.ChangeList {
		if ProductChange.Action != "unchanged" {
			errs = append(errs, fmt.Sprintf("there are pending changes.\nGo into the Ops Manager UI, unstage changes, and try again"))
			break
		}
	}

	if len(errs) > 0 {
		if pc.Options.Check {
			return fmt.Errorf("%s\nPlease validate your Ops Manager installation in the UI", strings.Join(errs, ",\n"))
		}

		pc.logger.Printf("Warnings:\n%s", strings.Join(errs, ",\n"))
	}
	return nil
}

func (pc PendingChanges) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists all products and will display whether they are unchanged (no pending changes) or changed (has pending changes).",
		ShortDescription: "checks for pending changes",
		Flags:            pc.Options,
	}
}
