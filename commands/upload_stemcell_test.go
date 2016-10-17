package commands_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/formcontent"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UploadStemcell", func() {
	var (
		stemcellService   *fakes.StemcellService
		diagnosticService *fakes.DiagnosticService
		multipart         *fakes.Multipart
		logger            *fakes.OtherLogger
	)

	BeforeEach(func() {
		multipart = &fakes.Multipart{}
		stemcellService = &fakes.StemcellService{}
		diagnosticService = &fakes.DiagnosticService{}
		logger = &fakes.OtherLogger{}
	})

	It("uploads the stemcell", func() {
		submission := formcontent.ContentSubmission{
			Length:      10,
			Content:     ioutil.NopCloser(strings.NewReader("")),
			ContentType: "some content-type",
		}
		multipart.FinalizeReturns(submission, nil)

		diagnosticService.ReportReturns(api.DiagnosticReport{Stemcells: []string{}}, nil)

		command := commands.NewUploadStemcell(multipart, stemcellService, diagnosticService, logger)

		err := command.Execute([]string{
			"--stemcell", "/path/to/stemcell.tgz",
		})
		Expect(err).NotTo(HaveOccurred())

		key, file := multipart.AddFileArgsForCall(0)
		Expect(key).To(Equal("stemcell[file]"))
		Expect(file).To(Equal("/path/to/stemcell.tgz"))
		Expect(stemcellService.UploadArgsForCall(0)).To(Equal(api.StemcellUploadInput{
			ContentLength: 10,
			Stemcell:      ioutil.NopCloser(strings.NewReader("")),
			ContentType:   "some content-type",
		}))

		Expect(multipart.FinalizeCallCount()).To(Equal(1))

		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("processing stemcell"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("beginning stemcell upload to Ops Manager"))

		format, v = logger.PrintfArgsForCall(2)
		Expect(fmt.Sprintf(format, v...)).To(Equal("finished upload"))
	})

	Context("when the stemcell already exists", func() {
		It("exists successfully without uploading", func() {
			submission := formcontent.ContentSubmission{
				Length:      10,
				Content:     ioutil.NopCloser(strings.NewReader("")),
				ContentType: "some content-type",
			}
			multipart.FinalizeReturns(submission, nil)

			diagnosticService.ReportReturns(api.DiagnosticReport{
				Stemcells: []string{"stemcell.tgz"},
			}, nil)

			command := commands.NewUploadStemcell(multipart, stemcellService, diagnosticService, logger)

			err := command.Execute([]string{
				"--stemcell", "/path/to/stemcell.tgz",
			})
			Expect(err).NotTo(HaveOccurred())

			format, v := logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, v...)).To(Equal("stemcell has already been uploaded"))
		})
	})

	Context("failure cases", func() {
		Context("when an unkwown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewUploadStemcell(multipart, stemcellService, diagnosticService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse upload-stemcell flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when the file cannot be opened", func() {
			It("returns an error", func() {
				command := commands.NewUploadStemcell(multipart, stemcellService, diagnosticService, logger)
				multipart.AddFileReturns(errors.New("bad file"))

				err := command.Execute([]string{"--stemcell", "/some/path"})
				Expect(err).To(MatchError("failed to load stemcell: bad file"))
			})
		})

		Context("when the stemcell cannot be uploaded", func() {
			It("returns and error", func() {
				command := commands.NewUploadStemcell(multipart, stemcellService, diagnosticService, logger)
				stemcellService.UploadReturns(api.StemcellUploadOutput{}, errors.New("some stemcell error"))

				err := command.Execute([]string{"--stemcell", "/some/path"})
				Expect(err).To(MatchError("failed to upload stemcell: some stemcell error"))
			})
		})

		Context("when the diagnostic report cannot be fetched", func() {
			It("returns an error", func() {
				command := commands.NewUploadStemcell(multipart, stemcellService, diagnosticService, logger)
				diagnosticService.ReportReturns(api.DiagnosticReport{}, errors.New("some diagnostic error"))

				err := command.Execute([]string{"--stemcell", "/some/path"})
				Expect(err).To(MatchError("failed to get diagnostic report: some diagnostic error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewUploadStemcell(nil, nil, nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This command will upload a stemcell to the target Ops Manager. If your stemcell already exists that upload will be skipped",
				ShortDescription: "uploads a given stemcell to the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
