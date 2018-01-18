package commands

import (
	"fmt"
	"path/filepath"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/formcontent"
)

type UploadStemcell struct {
	multipart         multipart
	logger            logger
	stemcellService   stemcellService
	diagnosticService diagnosticService
	Options           struct {
		Stemcell string `long:"stemcell" short:"s" required:"true" description:"path to stemcell"`
		Force    bool   `long:"force"    short:"f"                 description:"upload stemcell even if it already exists on the target Ops Manager"`
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

func (us UploadStemcell) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command will upload a stemcell to the target Ops Manager. Unless the force flag is used, if the stemcell already exists that upload will be skipped",
		ShortDescription: "uploads a given stemcell to the Ops Manager targeted",
		Flags:            us.Options,
	}
}

func (us UploadStemcell) Execute(args []string) error {
	if _, err := jhanda.Parse(&us.Options, args); err != nil {
		return fmt.Errorf("could not parse upload-stemcell flags: %s", err)
	}

	if !us.Options.Force {
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
	}

	err := us.multipart.AddFile("stemcell[file]", us.Options.Stemcell)
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
