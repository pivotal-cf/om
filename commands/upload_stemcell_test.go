package commands_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/formcontent"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UploadStemcell", func() {
	var (
		service   *fakes.StemcellService
		multipart *fakes.Multipart
		logger    *fakes.OtherLogger
		tmpfile   *os.File
	)

	BeforeEach(func() {
		var err error
		tmpfile, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

		multipart = &fakes.Multipart{}
		service = &fakes.StemcellService{}
		logger = &fakes.OtherLogger{}
	})

	It("uploads the stemcell", func() {
		submission := formcontent.ContentSubmission{
			Length:      10,
			Content:     tmpfile,
			ContentType: "some content-type",
		}
		multipart.CreateReturns(submission, nil)

		command := commands.NewUploadStemcell(multipart, service, logger)

		err := command.Execute([]string{
			"--stemcell", "/path/to/stemcell.tgz",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(multipart.CreateArgsForCall(0)).To(Equal("/path/to/stemcell.tgz"))
		Expect(service.UploadArgsForCall(0)).To(Equal(api.StemcellUploadInput{
			ContentLength: 10,
			Stemcell:      tmpfile,
			ContentType:   "some content-type",
		}))

		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("processing stemcell"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("beginning stemcell upload to Ops Manager"))

		format, v = logger.PrintfArgsForCall(2)
		Expect(fmt.Sprintf(format, v...)).To(Equal("finished upload"))
	})

	Context("failure cases", func() {
		Context("when an unkwown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewUploadStemcell(multipart, service, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse upload-stemcell flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when the file cannot be opened", func() {
			It("returns an error", func() {
				command := commands.NewUploadStemcell(multipart, service, logger)
				multipart.CreateReturns(formcontent.ContentSubmission{}, errors.New("bad file"))

				err := command.Execute([]string{"--stemcell", "/some/path"})
				Expect(err).To(MatchError("failed to load stemcell: bad file"))
			})
		})

		Context("when the stemcell cannot be uploaded", func() {
			It("returns and error", func() {
				command := commands.NewUploadStemcell(multipart, service, logger)
				service.UploadReturns(api.StemcellUploadOutput{}, errors.New("some stemcell error"))

				err := command.Execute([]string{"--stemcell", "/some/path"})
				Expect(err).To(MatchError("failed to upload stemcell: some stemcell error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewUploadStemcell(nil, nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This command will upload a stemcell to the target Ops Manager. If your stemcell already exists that upload will be skipped",
				ShortDescription: "uploads a given stemcell to the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
