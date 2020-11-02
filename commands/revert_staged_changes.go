package commands

import (
	"fmt"
)

type RevertStagedChanges struct {
	service revertStagedChangesService
	logger  logger
}

type revertStagedChangesService interface {
	RevertStagedChanges() (bool, error)
}

func NewRevertStagedChanges(service revertStagedChangesService, logger logger) *RevertStagedChanges {
	return &RevertStagedChanges{service: service, logger: logger}
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
