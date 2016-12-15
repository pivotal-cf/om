package commands

import (
	"fmt"
	"path/filepath"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
	"github.com/pivotal-cf/om/formcontent"
)

type UploadStemcell struct {
	multipart         multipart
	logger            logger
	stemcellService   stemcellService
	diagnosticService diagnosticService
	Options           struct {
		Stemcell string `short:"s"  long:"stemcell"  description:"path to stemcell"`
	}
}

//go:generate counterfeiter -o ./fakes/multipart.go --fake-name Multipart . multipart
type multipart interface {
	Finalize() (formcontent.ContentSubmission, error)
	AddFile(key, path string) error
	AddField(key, value string) error
}

//go:generate counterfeiter -o ./fakes/stemcell_service.go --fake-name StemcellService . stemcellService
type stemcellService interface {
	Upload(api.StemcellUploadInput) (api.StemcellUploadOutput, error)
}

//go:generate counterfeiter -o ./fakes/diagnostic_service.go --fake-name DiagnosticService . diagnosticService
type diagnosticService interface {
	Report() (api.DiagnosticReport, error)
}

func NewUploadStemcell(multipart multipart, stemcellService stemcellService, diagnosticService diagnosticService, logger logger) UploadStemcell {
	return UploadStemcell{
		multipart:         multipart,
		logger:            logger,
		stemcellService:   stemcellService,
		diagnosticService: diagnosticService,
	}
}

func (us UploadStemcell) Usage() Usage {
	return Usage{
		Description:      "This command will upload a stemcell to the target Ops Manager. If your stemcell already exists that upload will be skipped",
		ShortDescription: "uploads a given stemcell to the Ops Manager targeted",
		Flags:            us.Options,
	}
}

func (us UploadStemcell) Execute(args []string) error {
	_, err := flags.Parse(&us.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse upload-stemcell flags: %s", err)
	}

	us.logger.Printf("processing stemcell")

	report, err := us.diagnosticService.Report()
	if err != nil {
		switch err.(type) {
		case api.DiagnosticReportUnavailable:
			us.logger.Printf("%s", err)
		default:
			return fmt.Errorf("failed to get diagnostic report: %s", err)
		}
	}

	for _, stemcell := range report.Stemcells {
		if stemcell == filepath.Base(us.Options.Stemcell) {
			us.logger.Printf("stemcell has already been uploaded")
			return nil
		}
	}

	err = us.multipart.AddFile("stemcell[file]", us.Options.Stemcell)
	if err != nil {
		return fmt.Errorf("failed to load stemcell: %s", err)
	}

	submission, err := us.multipart.Finalize()
	if err != nil {
		return fmt.Errorf("failed to create multipart form: %s", err)
	}

	us.logger.Printf("beginning stemcell upload to Ops Manager")

	_, err = us.stemcellService.Upload(api.StemcellUploadInput{
		ContentLength: submission.Length,
		Stemcell:      submission.Content,
		ContentType:   submission.ContentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload stemcell: %s", err)
	}

	us.logger.Printf("finished upload")

	return nil
}
