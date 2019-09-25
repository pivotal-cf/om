package commands

import (
	"errors"
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type DisableProductVerifiers struct {
	service   disableProductVerifiersService
	presenter presenters.FormattedPresenter
	logger    logger
	Options   struct {
		ProductName   string   `long:"product-name"   short:"c" required:"true" description:"the name of the product"`
		VerifierTypes []string `long:"type" short:"t"  description:"verifier types to disable" required:"true"`
	}
}

//counterfeiter:generate -o ./fakes/disableProductVerifiersService.go --fake-name DisableProductVerifiersService . disableProductVerifiersService
type disableProductVerifiersService interface {
	ListProductVerifiers(productName string) ([]api.Verifier, string, error)
	DisableProductVerifiers(verifierTypes []string, productGUID string) error
}

func NewDisableProductVerifiers(presenter presenters.FormattedPresenter, service disableProductVerifiersService, logger logger) DisableProductVerifiers {
	return DisableProductVerifiers{
		service:   service,
		presenter: presenter,
		logger:    logger,
	}
}

func (dpv DisableProductVerifiers) Execute(args []string) error {
	_, err := jhanda.Parse(&dpv.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse disable-product-verifiers flags: %s", err)
	}

	productName := dpv.Options.ProductName
	productVerifiers, productGUID, err := dpv.service.ListProductVerifiers(productName)
	if err != nil {
		return fmt.Errorf("could not get available verifiers from Ops Manager: %s", err)
	}

	var missingVerifiers []string
	for _, verifier := range dpv.Options.VerifierTypes {
		found := false
		for _, pverifier := range productVerifiers {
			if verifier == pverifier.Type {
				found = true
				continue
			}
		}

		if !found {
			missingVerifiers = append(missingVerifiers, verifier)
		}
	}

	if len(missingVerifiers) > 0 {
		dpv.logger.Printf("The following verifiers do not exist for %s:\n", productName)
		for _, v := range missingVerifiers {
			dpv.logger.Printf("- %s\n", v)
		}

		dpv.logger.Println("\nNo changes were made.\n")

		return errors.New("verifier does not exist for product")
	}

	dpv.logger.Printf("Disabling Product Verifiers for %s...\n\n", productName)

	err = dpv.service.DisableProductVerifiers(dpv.Options.VerifierTypes, productGUID)
	if err != nil {
		return fmt.Errorf("could not disable verifiers in Ops Manager: %s", err)
	}

	dpv.logger.Println("The following verifiers were disabled:")
	for _, v := range dpv.Options.VerifierTypes {
		dpv.logger.Printf("- %s\n", v)
	}

	return nil
}

func (dpv DisableProductVerifiers) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command disables product verifiers",
		ShortDescription: "disables product verifiers",
		Flags:            dpv.Options,
	}
}
