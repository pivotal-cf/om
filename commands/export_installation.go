package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
)

type ExportInstallation struct {
	logger  logger
	service exportInstallationService
	Options struct {
		OutputFile      string `long:"output-file"      short:"o"  required:"true" description:"output path to write installation to"`
		PollingInterval int    `long:"polling-interval" short:"pi"                 description:"interval (in seconds) at which to print status" default:"1"`
	}
}

//go:generate counterfeiter -o ./fakes/export_installation_service.go --fake-name ExportInstallationService . exportInstallationService
type exportInstallationService interface {
	DownloadInstallationAssetCollection(outputFile string, pollingInterval int) error
}

func NewExportInstallation(service exportInstallationService, logger logger) ExportInstallation {
	return ExportInstallation{
		logger:  logger,
		service: service,
	}
}

func (ei ExportInstallation) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command will export the current installation of the target Ops Manager.",
		ShortDescription: "exports the installation of the target Ops Manager",
		Flags:            ei.Options,
	}
}

func (ei ExportInstallation) Execute(args []string) error {
	if _, err := jhanda.Parse(&ei.Options, args); err != nil {
		return fmt.Errorf("could not parse export-installation flags: %s", err)
	}

	ei.logger.Printf("exporting installation")

	err := ei.service.DownloadInstallationAssetCollection(ei.Options.OutputFile, ei.Options.PollingInterval)
	if err != nil {
		return fmt.Errorf("failed to export installation: %s", err)
	}

	ei.logger.Printf("finished exporting installation")

	return nil
}
