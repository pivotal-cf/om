package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
)

type ExportInstallation struct {
	logger                           logger
	installationAssetExporterService installationAssetExporterService
	Options                          struct {
		OutputFile      string `long:"output-file"      short:"o"  required:"true" description:"output path to write installation to"`
		PollingInterval int    `long:"polling-interval" short:"pi"                 description:"interval (in seconds) at which to print status" default:"1"`
	}
}

//go:generate counterfeiter -o ./fakes/installation_asset_exporter_service.go --fake-name InstallationAssetExporterService . installationAssetExporterService
type installationAssetExporterService interface {
	Export(outputFile string, pollingInterval int) error
}

func NewExportInstallation(installationAssetExporterService installationAssetExporterService, logger logger) ExportInstallation {
	return ExportInstallation{
		logger: logger,
		installationAssetExporterService: installationAssetExporterService,
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

	err := ei.installationAssetExporterService.Export(ei.Options.OutputFile, ei.Options.PollingInterval)
	if err != nil {
		return fmt.Errorf("failed to export installation: %s", err)
	}

	ei.logger.Printf("finished exporting installation")

	return nil
}
