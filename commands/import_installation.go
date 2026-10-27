package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/validator"
)

const maxRetries = 3

type ImportInstallation struct {
	multipart  multipart
	logger     logger
	service    importInstallationService
	passphrase string
	Options    struct {
		InterpolateOptions interpolateConfigFileOptions `group:"config file interpolation"`

		Installation    string `long:"installation"          short:"i"  required:"true" description:"path to installation."`
		PollingInterval int    `long:"polling-interval"      short:"p"                 description:"interval (in seconds) to check OpsManager availability" default:"10"`
	}
}

//counterfeiter:generate -o ./fakes/import_installation_service.go --fake-name ImportInstallationService . importInstallationService
type importInstallationService interface {
	UploadInstallationAssetCollection(api.ImportInstallationInput) error
	EnsureAvailability(input api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error)
}

func NewImportInstallation(multipart multipart, service importInstallationService, passphrase string, logger logger) *ImportInstallation {
	return &ImportInstallation{
		multipart:  multipart,
		logger:     logger,
		service:    service,
		passphrase: passphrase,
	}
}

func (ii *ImportInstallation) Execute(args []string) error {
	err := ii.validate(args)
	if err != nil {
		return err
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
		Installation:  submission.Content,
		ContentType:   submission.ContentType,
		ContentLength: submission.ContentLength,
	})
	if err != nil {
		return fmt.Errorf("failed to import installation: %s", err)
	}

	ii.logger.Printf("waiting for import to complete, this should take only a couple minutes...")

	err = ii.ensureAvailability()
	if err != nil {
		return err
	}

	ii.logger.Printf("finished import")

	return nil
}

func (ii ImportInstallation) ensureAvailability() error {
	var tryCount int

	for {
		time.Sleep(time.Second * time.Duration(ii.Options.PollingInterval))
		ensureAvailabilityOutput, err := ii.service.EnsureAvailability(api.EnsureAvailabilityInput{})
		if err != nil {
			if isErrThatMightResolveOnRetry(err, tryCount) {
				ii.logger.Printf("waiting for ops manager web server boots up...")
				tryCount++
				continue
			}
			return fmt.Errorf("could not check Ops Manager Status: %s", err)
		}
		if ensureAvailabilityOutput.Status == api.EnsureAvailabilityStatusComplete {
			break
		}
	}
	return nil
}

func (ii *ImportInstallation) validate(args []string) error {
	if ii.passphrase == "" {
		return errors.New("the global decryption-passphrase argument is required for this command")
	}

	if _, err := os.Stat(ii.Options.Installation); err != nil {
		return fmt.Errorf("file: \"%s\" does not exist. Please check the name and try again.", ii.Options.Installation)
	}

	if err := validator.ValidateInstallationZip(ii.Options.Installation); err != nil {
		return err
	}

	return nil
}

func isErrThatMightResolveOnRetry(err error, tryCount int) bool {
	return (strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "bad gateway")) &&
		tryCount < maxRetries
}
