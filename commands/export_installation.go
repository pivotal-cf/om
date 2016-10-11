package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/flags"
)

type ExportInstallation struct {
	logger              logger
	installationService installationService
	Options             struct {
		OutputFile string `short:"o"  long:"output-file"  description:"output path to write installation to"`
	}
}

//go:generate counterfeiter -o ./fakes/installation_service.go --fake-name InstallationService . installationService
type installationService interface {
	Export(string) error
}

func NewExportInstallation(installationService installationService, logger logger) ExportInstallation {
	return ExportInstallation{
		logger:              logger,
		installationService: installationService,
	}
}

func (ei ExportInstallation) Usage() Usage {
	return Usage{
		Description:      "This command will export the current installation of the target Ops Manager.",
		ShortDescription: "exports the installation of the target ops manager",
		Flags:            ei.Options,
	}
}

func (ei ExportInstallation) Execute(args []string) error {
	_, err := flags.Parse(&ei.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse export-installation flags: %s", err)
	}

	if ei.Options.OutputFile == "" {
		return errors.New("expected flag --output-file. Run 'om help export-installation' for more information.")
	}

	ei.logger.Printf("exporting installation")
	err = ei.installationService.Export(ei.Options.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to export installation: %s", err)
	}
	ei.logger.Printf("finished exporting installation")

	return nil
}
