package commands

import (
	"fmt"
)

type ExportInstallation struct {
	logger  logger
	service exportInstallationService
	Options struct {
		OutputFile string `long:"output-file"      short:"o"  required:"true" description:"output path to write installation to"`
	}
}

//counterfeiter:generate -o ./fakes/export_installation_service.go --fake-name ExportInstallationService . exportInstallationService
type exportInstallationService interface {
	DownloadInstallationAssetCollection(outputFile string) error
}

func NewExportInstallation(service exportInstallationService, logger logger) *ExportInstallation {
	return &ExportInstallation{
		logger:  logger,
		service: service,
	}
}

func (ei ExportInstallation) Execute(args []string) error {
	ei.logger.Printf("exporting installation")

	err := ei.service.DownloadInstallationAssetCollection(ei.Options.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to export installation: %s", err)
	}

	ei.logger.Printf("finished exporting installation")

	return nil
}
