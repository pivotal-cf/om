package commands

import (
	"fmt"

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
	Create(path string) (formcontent.ContentSubmission, error)
}

//go:generate counterfeiter -o ./fakes/stemcell_service.go --fake-name StemcellService . stemcellService
type stemcellService interface {
	Upload(api.StemcellUploadInput) (api.StemcellUploadOutput, error)
}

// TODO: check stemcell availability first
type diagnosticService interface {
}

func NewUploadStemcell(multipart multipart, stemcellService stemcellService, logger logger) UploadStemcell {
	return UploadStemcell{
		multipart:       multipart,
		logger:          logger,
		stemcellService: stemcellService,
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

	submission, err := us.multipart.Create(us.Options.Stemcell)
	if err != nil {
		return fmt.Errorf("failed to load stemcell: %s", err)
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
