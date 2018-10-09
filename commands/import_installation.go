package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type ImportInstallation struct {
	multipart  multipart
	logger     logger
	service    importInstallationService
	passphrase string
	Options    struct {
		ConfigFile      string `long:"config"                short:"c"                  description:"path to yml file for configuration (keys must match the following command line flags)"`
		Installation    string `long:"installation"          short:"i"  required:"true" description:"path to installation."`
		PollingInterval int    `long:"polling-interval"      short:"pi"                 description:"interval (in seconds) at which to print status" default:"1"`
	}
}

//go:generate counterfeiter -o ./fakes/import_installation_service.go --fake-name ImportInstallationService . importInstallationService
type importInstallationService interface {
	UploadInstallationAssetCollection(api.ImportInstallationInput) error
	EnsureAvailability(input api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error)
}

func NewImportInstallation(multipart multipart, service importInstallationService, passphrase string, logger logger) ImportInstallation {
	return ImportInstallation{
		multipart:  multipart,
		logger:     logger,
		service:    service,
		passphrase: passphrase,
	}
}

func (ii ImportInstallation) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This unauthenticated command attempts to import an installation to the Ops Manager targeted.",
		ShortDescription: "imports a given installation to the Ops Manager targeted",
		Flags:            ii.Options,
	}
}

func (ii ImportInstallation) Execute(args []string) error {
	if ii.passphrase == "" {
		return fmt.Errorf("the global decryption-passphrase argument is required for this command")
	}

	err := loadConfigFile(args, &ii.Options)
	if err != nil {
		return fmt.Errorf("could not parse import-installation flags: %s", err)
	}

	ensureAvailabilityOutput, err := ii.service.EnsureAvailability(api.EnsureAvailabilityInput{})
	if err != nil {
		return fmt.Errorf("could not check Ops Manager status: %s", err)
	}

	if ensureAvailabilityOutput.Status != api.EnsureAvailabilityStatusUnstarted {
		ii.logger.Printf("Ops Manager is already configured")
		return nil
	}

	ii.logger.Printf("processing installation")

	err = ii.multipart.AddFile("installation[file]", ii.Options.Installation)
	if err != nil {
		return fmt.Errorf("failed to load installation: %s", err)
	}

	err = ii.multipart.AddField("passphrase", ii.passphrase)
	if err != nil {
		return fmt.Errorf("failed to insert passphrase: %s", err)
	}

	submission := ii.multipart.Finalize()
	if err != nil {
		return fmt.Errorf("failed to create multipart form: %s", err)
	}

	ii.logger.Printf("beginning installation import to Ops Manager")

	err = ii.service.UploadInstallationAssetCollection(api.ImportInstallationInput{
		Installation:    submission.Content,
		ContentType:     submission.ContentType,
		ContentLength:   submission.ContentLength,
		PollingInterval: ii.Options.PollingInterval,
	})
	if err != nil {
		return fmt.Errorf("failed to import installation: %s", err)
	}

	ii.logger.Printf("waiting for import to complete, this should take only a couple minutes...")
	for ensureAvailabilityOutput.Status != api.EnsureAvailabilityStatusComplete {
		ensureAvailabilityOutput, err = ii.service.EnsureAvailability(api.EnsureAvailabilityInput{})
		if err != nil {
			return fmt.Errorf("could not check Ops Manager Status: %s", err)
		}
	}

	ii.logger.Printf("finished import")

	return nil
}
