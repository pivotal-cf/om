package commands

import (
	"errors"
	"fmt"
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
		interpolateConfigFileOptions

		Stemcell string `long:"stemcell" short:"s" required:"true" description:"path to stemcell"`
		Force    bool   `long:"force"    short:"f"                 description:"upload stemcell even if it already exists on the target Ops Manager"`
		Floating string `long:"floating" default:"true"            description:"assigns the stemcell to all compatible products "`
		Shasum   string `long:"shasum"                             description:"shasum of the provided product file to be used for validation"`
	}
}

//counterfeiter:generate -o ./fakes/multipart.go --fake-name Multipart . multipart
type multipart interface {
	Finalize() formcontent.ContentSubmission
	Reset()
	AddFile(key, path string) error
	AddField(key, value string) error
}

//counterfeiter:generate -o ./fakes/upload_stemcell_service.go --fake-name UploadStemcellService . uploadStemcellService
type uploadStemcellService interface {
	UploadStemcell(api.StemcellUploadInput) (api.StemcellUploadOutput, error)
	CheckStemcellAvailability(string) (bool, error)
	GetDiagnosticReport() (api.DiagnosticReport, error)
	Info() (api.Info, error)
}

func NewUploadStemcell(multipart multipart, service uploadStemcellService, logger logger) *UploadStemcell {
	return &UploadStemcell{
		multipart: multipart,
		logger:    logger,
		service:   service,
	}
}

func (us UploadStemcell) Execute(args []string) error {
	//err := cmd.loadConfigFile(args, &us.Options, os.Environ)
	//if err != nil {
	//	return fmt.Errorf("could not parse upload-stemcell flags: %s", err)
	//}

	err := us.validate()
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

		stemcellAbsPath, err := filepath.Abs(stemcellFilename)
		if err != nil {
			return err
		}

		symlinkedStemcell := filepath.Join(filepath.Dir(stemcellAbsPath), matches[1])
		err = os.Symlink(stemcellAbsPath, symlinkedStemcell)
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

	found, err := us.service.CheckStemcellAvailability(us.Options.Stemcell)
	if err != nil {
		return false, err
	}

	if found {
		us.logger.Printf("stemcell has already been uploaded")
	}

	return found, nil
}
