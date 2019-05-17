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
	var errorBuffer []string

	if _, err := jhanda.Parse(&pc.Options, args); err != nil {
		return fmt.Errorf("could not parse pending-changes flags: %s", err)
	}

	pc.logger.Println("Scanning OpsManager now ...\n")

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

		errs := pc.determineDirectorErrors(pendingDirectorChanges)
		errorBuffer = append(errorBuffer, errs...)
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
			errs := pc.determineProductErrors(change)
			errorBuffer = append(errorBuffer, errs...)
		} else {
			pc.logger.Printf(color.GreenString("[✓] product: %s", change.EndpointResults.Identifier))
		}
	}

	pc.logger.Println()

	if len(errorBuffer) > 0 {
		for _, err := range errorBuffer {
			pc.logger.Printf("%s\n", err)
		}

		return fmt.Errorf("OpsManager is not fully configured")
	}

	pc.logger.Println("The director and products are configured correctly.")
	return nil
}

func (pc PreDeployCheck) determineDirectorErrors(directorOutput api.PendingDirectorChangesOutput) []string {
	var errBuffer []string

	errBuffer = append(errBuffer, fmt.Sprintf(color.RedString("[X] %s"), directorOutput.EndpointResults.Identifier))
	if !directorOutput.EndpointResults.Network.Assigned {
		errBuffer = append(errBuffer, "    Error: Network is not assigned\n")
	}

	if !directorOutput.EndpointResults.AvailabilityZone.Assigned {
		errBuffer = append(errBuffer, "    Error: Availability Zone is not assigned\n")
	}

	for _, stemcell := range directorOutput.EndpointResults.Stemcells {
		if !stemcell.Assigned {
			errBuffer = append(errBuffer, "    Error: missing stemcell")
			errBuffer = append(errBuffer, fmt.Sprintf("    Why: Required stemcell OS - %s version %s", stemcell.RequiredStemcellOS, stemcell.RequiredStemcellVersion))
			errBuffer = append(errBuffer, fmt.Sprintf("    Fix: Download %s version %s from Pivnet and upload to OpsManager\n", stemcell.RequiredStemcellOS, stemcell.RequiredStemcellVersion))
		}
	}

	for _, property := range directorOutput.EndpointResults.Properties {
		errBuffer = append(errBuffer, fmt.Sprintf("    Error: property - %s", property.Name))
		for _, err := range property.Errors {
			errBuffer = append(errBuffer, fmt.Sprintf("    Why: %s\n", err))
		}
	}

	for _, job := range directorOutput.EndpointResults.Resources.Jobs {
		errBuffer = append(errBuffer, fmt.Sprintf("    Error: resource - %s", job.Identifier))
		for _, err := range job.Errors {
			errBuffer = append(errBuffer, fmt.Sprintf("    Why: %s\n", err))
		}
	}

	for _, verifier := range directorOutput.EndpointResults.Verifiers {
		errBuffer = append(errBuffer, fmt.Sprintf("    Error: verifier - %s", verifier.Type))
		for _, err := range verifier.Errors {
			errBuffer = append(errBuffer, fmt.Sprintf("    Why: %s\n", err))
		}
	}

	return errBuffer
}

func (pc PreDeployCheck) determineProductErrors(productOutput api.PendingProductChangesOutput) []string {
	var errBuffer []string

	errBuffer = append(errBuffer, fmt.Sprintf(color.RedString("[X] %s"), productOutput.EndpointResults.Identifier))
	if !productOutput.EndpointResults.Network.Assigned {
		errBuffer = append(errBuffer, "    Error: Network is not assigned\n")
	}

	if !productOutput.EndpointResults.AvailabilityZone.Assigned {
		errBuffer = append(errBuffer, "    Error: Availability Zone is not assigned\n")
	}

	for _, stemcell := range productOutput.EndpointResults.Stemcells {
		if !stemcell.Assigned {
			errBuffer = append(errBuffer, "    Error: missing stemcell")
			errBuffer = append(errBuffer, fmt.Sprintf("    Why: Required stemcell OS - %s version %s", stemcell.RequiredStemcellOS, stemcell.RequiredStemcellVersion))
			errBuffer = append(errBuffer, fmt.Sprintf("    Fix: Download %s version %s from Pivnet and upload to OpsManager\n", stemcell.RequiredStemcellOS, stemcell.RequiredStemcellVersion))
		}
	}

	for _, property := range productOutput.EndpointResults.Properties {
		errBuffer = append(errBuffer, fmt.Sprintf("    Error: property - %s", property.Name))
		for _, err := range property.Errors {
			errBuffer = append(errBuffer, fmt.Sprintf("    Why: %s\n", err))
		}
	}

	for _, job := range productOutput.EndpointResults.Resources.Jobs {
		errBuffer = append(errBuffer, fmt.Sprintf("    Error: resource - %s", job.Identifier))
		for _, err := range job.Errors {
			errBuffer = append(errBuffer, fmt.Sprintf("    Why: %s\n", err))
		}
	}

	for _, verifier := range productOutput.EndpointResults.Verifiers {
		errBuffer = append(errBuffer, fmt.Sprintf("    Error: verifier - %s", verifier.Type))
		for _, err := range verifier.Errors {
			errBuffer = append(errBuffer, fmt.Sprintf("    Why: %s\n", err))
		}
	}

	return errBuffer
}

func (pc PreDeployCheck) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "**EXPERIMENTAL** This authenticated command lists all pending changes.",
		ShortDescription: "**EXPERIMENTAL** lists pending changes",
		Flags:            pc.Options,
	}
}
