package commands

import (
	"errors"
	"fmt"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type DisableDirectorVerifiers struct {
	service   disableDirectorVerifiersService
	presenter presenters.FormattedPresenter
	logger    logger
	Options   struct {
		interpolateConfigFileOptions

		VerifierTypes []string `long:"type" short:"t"  description:"verifier types to disable" required:"true"`
	}
}

//counterfeiter:generate -o ./fakes/disableDirectorVerifiersService.go --fake-name DisableDirectorVerifiersService . disableDirectorVerifiersService
type disableDirectorVerifiersService interface {
	ListDirectorVerifiers() ([]api.Verifier, error)
	DisableDirectorVerifiers(verifierTypes []string) error
}

func NewDisableDirectorVerifiers(presenter presenters.FormattedPresenter, service disableDirectorVerifiersService, logger logger) *DisableDirectorVerifiers {
	return &DisableDirectorVerifiers{
		service:   service,
		presenter: presenter,
		logger:    logger,
	}
}

func (dv DisableDirectorVerifiers) Execute(args []string) error {
	directorVerifiers, err := dv.service.ListDirectorVerifiers()
	if err != nil {
		return fmt.Errorf("could not get available verifiers from Ops Manager: %s", err)
	}

	var missingVerifiers []string
	for _, verifier := range dv.Options.VerifierTypes {
		found := false
		for _, dverifier := range directorVerifiers {
			if verifier == dverifier.Type {
				found = true
				continue
			}
		}

		if !found {
			missingVerifiers = append(missingVerifiers, verifier)
		}
	}

	if len(missingVerifiers) > 0 {
		dv.logger.Println("The following verifiers do not exist:")
		for _, v := range missingVerifiers {
			dv.logger.Printf("- %s\n", v)
		}

		dv.logger.Println("\nNo changes were made.\n")

		return errors.New("verifier does not exist for director")
	}

	dv.logger.Println("Disabling Director Verifiers...\n")

	err = dv.service.DisableDirectorVerifiers(dv.Options.VerifierTypes)
	if err != nil {
		return fmt.Errorf("could not disable verifiers in Ops Manager: %s", err)
	}

	dv.logger.Println("The following verifiers were disabled:")
	for _, v := range dv.Options.VerifierTypes {
		dv.logger.Printf("- %s\n", v)
	}

	return nil
}
