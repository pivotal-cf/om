package commands

import (
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

type ImportInstallation struct {
	multipart           multipart
	logger              logger
	installationService installationService
	Options             struct {
		Installation string `short:"i"  long:"installation"  description:"path to installation"`
		Passphrase   string `short:"p" long:"passphrase" description:"passphrase for Ops Manager to decrypt the installation"`
	}
}

func NewImportInstallation(multipart multipart, installationService installationService, logger logger) ImportInstallation {
	return ImportInstallation{
		multipart:           multipart,
		logger:              logger,
		installationService: installationService,
	}
}

func (ii ImportInstallation) Usage() Usage {
	return Usage{
		Description:      "This command attempts to import an installation to the Ops Manager",
		ShortDescription: "imports a given installation to the Ops Manager targeted",
		Flags:            ii.Options,
	}
}

func (ii ImportInstallation) Execute(args []string) error {
	_, err := flags.Parse(&ii.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse import-installation flags: %s", err)
	}

	ii.logger.Printf("processing installation")

	err = ii.multipart.AddFile("installation[file]", ii.Options.Installation)
	if err != nil {
		return fmt.Errorf("failed to load installation: %s", err)
	}

	err = ii.multipart.AddField("passphrase", ii.Options.Passphrase)
	if err != nil {
		return fmt.Errorf("failed to insert passphrase: %s", err)
	}

	submission, err := ii.multipart.Create()
	if err != nil {
		return fmt.Errorf("failed to create multipart form: %s", err)
	}

	ii.logger.Printf("beginning installation import to Ops Manager")

	err = ii.installationService.Import(api.ImportInstallationInput{
		ContentLength: submission.Length,
		Installation:  submission.Content,
		ContentType:   submission.ContentType,
	})
	if err != nil {
		return fmt.Errorf("failed to import installation: %s", err)
	}

	ii.logger.Printf("finished import")

	return nil
}
