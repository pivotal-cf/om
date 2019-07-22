package commands

import (
	"errors"
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/formcontent"
	"github.com/pivotal-cf/om/network"
	"github.com/pivotal-cf/om/validator"
	"os"
	"path/filepath"
	"regexp"
)

const maxStemcellUploadRetries = 2

type UploadStemcell struct {
	multipart multipart
	logger    logger
	service   uploadStemcellService
	Options   struct {
		ConfigFile string `long:"config"   short:"c"                 description:"path to yml file for configuration (keys must match the following command line flags)"`
		Stemcell   string `long:"stemcell" short:"s" required:"true" description:"path to stemcell"`
		Force      bool   `long:"force"    short:"f"                 description:"upload stemcell even if it already exists on the target Ops Manager"`
		Floating   string `long:"floating" default:"true"            description:"assigns the stemcell to all compatible products "`
		Shasum     string `long:"shasum"                             description:"shasum of the provided product file to be used for validation"`
	}
}

//go:generate counterfeiter -o ./fakes/multipart.go --fake-name Multipart . multipart
type multipart interface {
	Finalize() formcontent.ContentSubmission
	Reset()
	AddFile(key, path string) error
	AddField(key, value string) error
}

//go:generate counterfeiter -o ./fakes/upload_stemcell_service.go --fake-name UploadStemcellService . uploadStemcellService
type uploadStemcellService interface {
	UploadStemcell(api.StemcellUploadInput) (api.StemcellUploadOutput, error)
	GetDiagnosticReport() (api.DiagnosticReport, error)
	Info() (api.Info, error)
}

func NewUploadStemcell(multipart multipart, service uploadStemcellService, logger logger) UploadStemcell {
	return UploadStemcell{
		multipart: multipart,
		logger:    logger,
		service:   service,
	}
}

func (us UploadStemcell) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command will upload a stemcell to the target Ops Manager. Unless the force flag is used, if the stemcell already exists that upload will be skipped",
		ShortDescription: "uploads a given stemcell to the Ops Manager targeted",
		Flags:            us.Options,
	}
}

func (us UploadStemcell) Execute(args []string) error {
	err := loadConfigFile(args, &us.Options, nil)
	if err != nil {
		return fmt.Errorf("could not parse upload-stemcell flags: %s", err)
	}

	err = us.validate()
	if err != nil {
		return err
	}

	stemcellFilename := us.Options.Stemcell
	if !us.Options.Force {
		exists, err := us.checkStemcellUploaded()
		if err != nil {
			return err
		}

		if exists {
			return nil
		}
	}

	prefixRegex := regexp.MustCompile(`^\[.*?,.*?\](.+)$`)
	if prefixRegex.MatchString(filepath.Base(stemcellFilename)) {
		matches := prefixRegex.FindStringSubmatch(filepath.Base(stemcellFilename))

		symlinkedStemcell := filepath.Join(filepath.Dir(stemcellFilename), matches[1])
		err = os.Symlink(stemcellFilename, symlinkedStemcell)
		if err != nil {
			return err
		}
		stemcellFilename = symlinkedStemcell

		defer os.Remove(symlinkedStemcell)
	}

	err = us.uploadStemcell(stemcellFilename)
	if err != nil {
		return fmt.Errorf("failed to upload stemcell: %s", err)
	}

	us.logger.Printf("finished upload")

	return nil
}

func (us UploadStemcell) uploadStemcell(stemcellFilename string) (err error) {
	for i := 0; i <= maxStemcellUploadRetries; i++ {
		err = us.multipart.AddFile("stemcell[file]", stemcellFilename)
		if err != nil {
			return err
		}

		err = us.multipart.AddField("stemcell[floating]", us.Options.Floating)
		if err != nil {
			return err
		}

		submission := us.multipart.Finalize()
		if err != nil {
			return fmt.Errorf("failed to create multipart form: %s", err)
		}

		us.logger.Printf("beginning stemcell upload to Ops Manager")

		_, err = us.service.UploadStemcell(api.StemcellUploadInput{
			Stemcell:      submission.Content,
			ContentType:   submission.ContentType,
			ContentLength: submission.ContentLength,
		})
		if network.CanRetry(err) && i < maxStemcellUploadRetries {
			us.logger.Printf("retrying stemcell upload after error: %s\n", err)
			us.multipart.Reset()
		} else {
			break
		}
	}

	return err
}

func (us UploadStemcell) validate() error {
	if us.Options.Floating != "true" && us.Options.Floating != "false" {
		return errors.New("--floating must be \"true\" or \"false\". Default: true")
	}

	if us.Options.Shasum != "" {
		shaValidator := validator.NewSHA256Calculator()
		shasum, err := shaValidator.Checksum(us.Options.Stemcell)

		if err != nil {
			return err
		}

		if shasum != us.Options.Shasum {
			return fmt.Errorf("expected shasum %s does not match file shasum %s", us.Options.Shasum, shasum)
		}

		us.logger.Printf("expected shasum matches stemcell shasum.")
	}

	return nil
}

func (us UploadStemcell) checkStemcellUploaded() (exists bool, err error) {
	us.logger.Printf("processing stemcell")

	stemcellFilename := us.Options.Stemcell
	exists = true

	report, err := us.service.GetDiagnosticReport()
	if err != nil {
		switch err.(type) {
		case api.DiagnosticReportUnavailable:
			us.logger.Printf("%s", err)
		default:
			return !exists, fmt.Errorf("failed to get diagnostic report: %s", err)
		}
	}

	info, err := us.service.Info()
	if err != nil {
		return !exists, fmt.Errorf("cannot retrieve version of Ops Manager")
	}

	validVersion, err := info.VersionAtLeast(2, 6)
	if err != nil {
		return !exists, fmt.Errorf("could not determine version was 2.6+ compatible: %s", err)
	}

	if validVersion {
		for _, stemcell := range report.AvailableStemcells {
			if stemcell.Filename == filepath.Base(stemcellFilename) {
				us.logger.Printf("stemcell has already been uploaded")
				return exists, nil
			}
		}
	}

	for _, stemcell := range report.Stemcells {
		if stemcell == filepath.Base(stemcellFilename) {
			us.logger.Printf("stemcell has already been uploaded")
			return exists, nil
		}
	}

	return !exists, nil
}
