package commands_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/formcontent"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UploadStemcell", func() {
	var (
		fakeService *fakes.UploadStemcellService
		multipart   *fakes.Multipart
		logger      *fakes.Logger
	)

	BeforeEach(func() {
		multipart = &fakes.Multipart{}
		fakeService = &fakes.UploadStemcellService{}
		logger = &fakes.Logger{}
	})

	Context("uploads the stemcell", func() {
		It("to all compatible products", func() {
			submission := formcontent.ContentSubmission{
				Content:       ioutil.NopCloser(strings.NewReader("")),
				ContentType:   "some content-type",
				ContentLength: 10,
			}
			multipart.FinalizeReturns(submission)

			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{Stemcells: []string{}}, nil)

			command := commands.NewUploadStemcell(multipart, fakeService, logger)

			err := command.Execute([]string{
				"--stemcell", "/path/to/stemcell.tgz",
			})
			Expect(err).NotTo(HaveOccurred())

			key, file := multipart.AddFileArgsForCall(0)
			Expect(key).To(Equal("stemcell[file]"))
			Expect(file).To(Equal("/path/to/stemcell.tgz"))

			key, value := multipart.AddFieldArgsForCall(0)
			Expect(key).To(Equal("stemcell[floating]"))
			Expect(value).To(Equal("true"))

			Expect(fakeService.UploadStemcellArgsForCall(0)).To(Equal(api.StemcellUploadInput{
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

		It("disables floating", func() {
			submission := formcontent.ContentSubmission{
				ContentLength: 10,
				Content:       ioutil.NopCloser(strings.NewReader("")),
				ContentType:   "some content-type",
			}
			multipart.FinalizeReturns(submission)

			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{Stemcells: []string{}}, nil)

			command := commands.NewUploadStemcell(multipart, fakeService, logger)

			err := command.Execute([]string{
				"--stemcell", "/path/to/stemcell.tgz",
				"--floating=false",
			})
			Expect(err).NotTo(HaveOccurred())

			key, file := multipart.AddFileArgsForCall(0)
			Expect(key).To(Equal("stemcell[file]"))
			Expect(file).To(Equal("/path/to/stemcell.tgz"))

			key, value := multipart.AddFieldArgsForCall(0)
			Expect(key).To(Equal("stemcell[floating]"))
			Expect(value).To(Equal("false"))

			Expect(fakeService.UploadStemcellArgsForCall(0)).To(Equal(api.StemcellUploadInput{
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
	})

	Context("when the stemcell already exists", func() {
		Context("and force is not specified", func() {
			It("exits successfully without uploading", func() {
				submission := formcontent.ContentSubmission{
					ContentLength: 10,
					Content:       ioutil.NopCloser(strings.NewReader("")),
					ContentType:   "some content-type",
				}
				multipart.FinalizeReturns(submission)

				fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{
					Stemcells: []string{"stemcell.tgz"},
				}, nil)

				command := commands.NewUploadStemcell(multipart, fakeService, logger)

				err := command.Execute([]string{
					"--stemcell", "/path/to/stemcell.tgz",
				})
				Expect(err).NotTo(HaveOccurred())

				format, v := logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, v...)).To(Equal("stemcell has already been uploaded"))
			})
		})

		Context("and force is specified", func() {
			It("uploads the stemcell", func() {
				submission := formcontent.ContentSubmission{
					Content:       ioutil.NopCloser(strings.NewReader("")),
					ContentType:   "some content-type",
					ContentLength: 10,
				}
				multipart.FinalizeReturns(submission)

				fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{
					Stemcells: []string{"stemcell.tgz"},
				}, nil)

				command := commands.NewUploadStemcell(multipart, fakeService, logger)

				err := command.Execute([]string{
					"--stemcell", "/path/to/stemcell.tgz",
					"--force",
				})
				Expect(err).NotTo(HaveOccurred())

				key, file := multipart.AddFileArgsForCall(0)
				Expect(key).To(Equal("stemcell[file]"))
				Expect(file).To(Equal("/path/to/stemcell.tgz"))
				Expect(fakeService.UploadStemcellArgsForCall(0)).To(Equal(api.StemcellUploadInput{
					ContentLength: 10,
					Stemcell:      ioutil.NopCloser(strings.NewReader("")),
					ContentType:   "some content-type",
				}))

				Expect(multipart.FinalizeCallCount()).To(Equal(1))

				format, v := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, v...)).To(Equal("beginning stemcell upload to Ops Manager"))

				format, v = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, v...)).To(Equal("finished upload"))
			})
		})
	})

	Context("when the --shasum flag is defined", func() {
		It("proceeds normally when the sha sums match", func() {
			file, err := ioutil.TempFile("", "test-file.tgz")
			Expect(err).ToNot(HaveOccurred())

			file.Close()
			defer os.Remove(file.Name())

			file.WriteString("testing-shasum")

			submission := formcontent.ContentSubmission{
				ContentLength: 10,
				Content:       ioutil.NopCloser(strings.NewReader("")),
				ContentType:   "some content-type",
			}
			multipart.FinalizeReturns(submission)

			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{Stemcells: []string{}}, nil)

			command := commands.NewUploadStemcell(multipart, fakeService, logger)
			err = command.Execute([]string{
				"--stemcell", file.Name(),
				"--shasum", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			})
			Expect(err).NotTo(HaveOccurred())
			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(ContainSubstring("expected shasum matches stemcell shasum."))
		})

		It("returns an error when the sha sums don't match", func() {
			file, err := ioutil.TempFile("", "test-file.tgz")
			Expect(err).ToNot(HaveOccurred())

			file.Close()
			defer os.Remove(file.Name())

			file.WriteString("testing-shasum")

			command := commands.NewUploadStemcell(multipart, fakeService, logger)
			err = command.Execute([]string{
				"--stemcell", file.Name(),
				"--shasum", "not-the-correct-shasum",
			})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("expected shasum not-the-correct-shasum does not match file shasum e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"))
		})
		It("fails when the file can not calculate a shasum", func() {
			command := commands.NewUploadStemcell(multipart, fakeService, logger)
			err := command.Execute([]string{
				"--stemcell", "/path/to/testing.tgz",
				"--shasum", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("open /path/to/testing.tgz: no such file or directory"))
		})
	})

	Context("when the diagnostic report is unavailable", func() {
		It("uploads the stemcell", func() {
			submission := formcontent.ContentSubmission{
				ContentLength: 10,
				Content:       ioutil.NopCloser(strings.NewReader("")),
				ContentType:   "some content-type",
			}
			multipart.FinalizeReturns(submission)

			fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{}, api.DiagnosticReportUnavailable{})

			command := commands.NewUploadStemcell(multipart, fakeService, logger)

			err := command.Execute([]string{
				"--stemcell", "/path/to/stemcell.tgz",
			})
			Expect(err).NotTo(HaveOccurred())

			key, file := multipart.AddFileArgsForCall(0)
			Expect(key).To(Equal("stemcell[file]"))
			Expect(file).To(Equal("/path/to/stemcell.tgz"))
			Expect(fakeService.UploadStemcellArgsForCall(0)).To(Equal(api.StemcellUploadInput{
				ContentLength: 10,
				Stemcell:      ioutil.NopCloser(strings.NewReader("")),
				ContentType:   "some content-type",
			}))

			Expect(multipart.FinalizeCallCount()).To(Equal(1))

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(Equal("processing stemcell"))

			format, v = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, v...)).To(Equal("diagnostic report is currently unavailable"))

			format, v = logger.PrintfArgsForCall(2)
			Expect(fmt.Sprintf(format, v...)).To(Equal("beginning stemcell upload to Ops Manager"))

			format, v = logger.PrintfArgsForCall(3)
			Expect(fmt.Sprintf(format, v...)).To(Equal("finished upload"))
		})
	})

	Context("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewUploadStemcell(multipart, fakeService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse upload-stemcell flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when the --stemcell flag is missing", func() {
			It("returns an error", func() {
				command := commands.NewUploadStemcell(multipart, fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not parse upload-stemcell flags: missing required flag \"--stemcell\""))
			})
		})

		Context("when the file cannot be opened", func() {
			It("returns an error", func() {
				command := commands.NewUploadStemcell(multipart, fakeService, logger)
				multipart.AddFileReturns(errors.New("bad file"))

				err := command.Execute([]string{"--stemcell", "/some/path"})
				Expect(err).To(MatchError("failed to load stemcell: bad file"))
			})
		})

		Context("when the stemcell cannot be uploaded", func() {
			It("returns an error", func() {
				command := commands.NewUploadStemcell(multipart, fakeService, logger)
				fakeService.UploadStemcellReturns(api.StemcellUploadOutput{}, errors.New("some stemcell error"))

				err := command.Execute([]string{"--stemcell", "/some/path"})
				Expect(err).To(MatchError("failed to upload stemcell: some stemcell error"))
			})
		})

		Context("when the diagnostic report cannot be fetched", func() {
			It("returns an error", func() {
				command := commands.NewUploadStemcell(multipart, fakeService, logger)
				fakeService.GetDiagnosticReportReturns(api.DiagnosticReport{}, errors.New("some diagnostic error"))

				err := command.Execute([]string{"--stemcell", "/some/path"})
				Expect(err).To(MatchError("failed to get diagnostic report: some diagnostic error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewUploadStemcell(nil, nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command will upload a stemcell to the target Ops Manager. Unless the force flag is used, if the stemcell already exists that upload will be skipped",
				ShortDescription: "uploads a given stemcell to the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
