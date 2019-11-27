package commands

import (
	"fmt"
	"github.com/pivotal-cf/jhanda"
)

type RevertStagedChanges struct {
	service revertStagedChangesService
	logger  logger
}

type revertStagedChangesService interface {
	RevertStagedChanges() (bool, error)
}

func NewRevertStagedChanges(service revertStagedChangesService, logger logger) RevertStagedChanges {
	return RevertStagedChanges{service: service, logger: logger}
}

func (r RevertStagedChanges) Execute(_ []string) error {
	reverted, err := r.service.RevertStagedChanges()

	if err != nil {
		return fmt.Errorf("revert staged changes command failed: %s", err)
	}

	if reverted {
		r.logger.Printf("Changes reverted.\n")
	} else {
		r.logger.Printf("No changes to revert.\n")
	}

	return nil
}

func (r RevertStagedChanges) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command reverts the staged changes already on an Ops Manager. Useful for ensuring that unintended changes are not applied.",
		ShortDescription: "This command reverts the staged changes already on an Ops Manager.",
	}
}
