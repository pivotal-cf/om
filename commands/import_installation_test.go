package commands_test

import (
	"archive/zip"
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
	"os"
)

var _ = Describe("ImportInstallation", func() {
	var (
		fakeService      *fakes.ImportInstallationService
		multipart        *fakes.Multipart
		logger           *fakes.Logger
		installationFile string
	)

	createZipFile := func(files []struct{ Name, Body string }) string {
		tmpFile, err := ioutil.TempFile("", "")
		w := zip.NewWriter(tmpFile)

		Expect(err).ToNot(HaveOccurred())
		for _, file := range files {
			f, err := w.Create(file.Name)
			if err != nil {
				Expect(err).ToNot(HaveOccurred())
			}
			_, err = f.Write([]byte(file.Body))
			if err != nil {
				Expect(err).ToNot(HaveOccurred())
			}
		}
		err = w.Close()
		Expect(err).ToNot(HaveOccurred())

		return tmpFile.Name()
	}

	BeforeEach(func() {
		multipart = &fakes.Multipart{}
		fakeService = &fakes.ImportInstallationService{}
		logger = &fakes.Logger{}
		installationFile = createZipFile([]struct{ Name, Body string }{
			{"installation.yml", ""},
		})
	})

	AfterEach(func() {
		os.Remove(installationFile)
	})

	It("imports an installation", func() {
		submission := formcontent.ContentSubmission{
			Content:       ioutil.NopCloser(strings.NewReader("")),
			ContentType:   "some content-type",
			ContentLength: 10,
		}
		multipart.FinalizeReturns(submission)

		eaOutputs := []api.EnsureAvailabilityOutput{
			{Status: api.EnsureAvailabilityStatusUnstarted},
			{Status: api.EnsureAvailabilityStatusPending},
			{Status: api.EnsureAvailabilityStatusPending},
			{Status: api.EnsureAvailabilityStatusPending},
			{Status: api.EnsureAvailabilityStatusComplete},
		}

		fakeService.EnsureAvailabilityStub = func(api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error) {
			return eaOutputs[fakeService.EnsureAvailabilityCallCount()-1], nil
		}

		command := commands.NewImportInstallation(multipart, fakeService, "some-passphrase", logger)

		err := executeCommand(command, []string{"--polling-interval", "0",
			"--installation", installationFile,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeService.EnsureAvailabilityCallCount()).To(Equal(5))

		key, file := multipart.AddFileArgsForCall(0)
		Expect(key).To(Equal("installation[file]"))
		Expect(file).To(Equal(installationFile))

		key, val := multipart.AddFieldArgsForCall(0)
		Expect(key).To(Equal("passphrase"))
		Expect(val).To(Equal("some-passphrase"))

		Expect(multipart.FinalizeCallCount()).To(Equal(1))

		Expect(fakeService.UploadInstallationAssetCollectionArgsForCall(0)).To(Equal(api.ImportInstallationInput{
			ContentLength: 10,
			Installation:  ioutil.NopCloser(strings.NewReader("")),
			ContentType:   "some content-type",
		}))

		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("processing installation"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("beginning installation import to Ops Manager"))

		format, v = logger.PrintfArgsForCall(2)
		Expect(fmt.Sprintf(format, v...)).To(Equal("waiting for import to complete, this should take only a couple minutes..."))

		format, v = logger.PrintfArgsForCall(3)
		Expect(fmt.Sprintf(format, v...)).To(Equal("finished import"))
	})

	When("the Ops Manager is already configured", func() {
		It("prints a helpful message", func() {
			fakeService.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
				Status: api.EnsureAvailabilityStatusComplete,
			}, nil)

			command := commands.NewImportInstallation(multipart, fakeService, "some-passphrase", logger)

			err := executeCommand(command, []string{"--polling-interval", "0",
				"--installation", installationFile,
			})
			Expect(err).ToNot(HaveOccurred())

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(Equal("Ops Manager is already configured"))
			Expect(fakeService.EnsureAvailabilityCallCount()).To(Equal(1))
		})
	})

	When("EnsureAvailability returns 'connection refused'", func() {
		var command *commands.ImportInstallation

		BeforeEach(func() {
			submission := formcontent.ContentSubmission{
				Content:       ioutil.NopCloser(strings.NewReader("")),
				ContentType:   "some content-type",
				ContentLength: 10,
			}
			multipart.FinalizeReturns(submission)

			command = commands.NewImportInstallation(multipart, fakeService, "some-passphrase", logger)
		})

		It("it retries on the specified polling interval to allow nginx time to boot up", func() {
			fakeService.EnsureAvailabilityStub = func(api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error) {
				if fakeService.EnsureAvailabilityCallCount() < 4 && fakeService.EnsureAvailabilityCallCount() > 2 {
					return api.EnsureAvailabilityOutput{}, fmt.Errorf("connection refused")
				}

				eaOutputs := []api.EnsureAvailabilityOutput{
					{Status: api.EnsureAvailabilityStatusUnstarted},
					{Status: api.EnsureAvailabilityStatusPending},
					{Status: api.EnsureAvailabilityStatusPending},
					{Status: api.EnsureAvailabilityStatusPending},
					{Status: api.EnsureAvailabilityStatusComplete},
				}
				return eaOutputs[fakeService.EnsureAvailabilityCallCount()-1], nil
			}

			err := executeCommand(command, []string{"--polling-interval", "0",
				"--installation", installationFile,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeService.EnsureAvailabilityCallCount()).To(Equal(5))

			Expect(logger.PrintfCallCount()).To(Equal(5))

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(Equal("processing installation"))

			format, v = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, v...)).To(Equal("beginning installation import to Ops Manager"))

			format, v = logger.PrintfArgsForCall(2)
			Expect(fmt.Sprintf(format, v...)).To(Equal("waiting for import to complete, this should take only a couple minutes..."))

			format, v = logger.PrintfArgsForCall(3)
			Expect(fmt.Sprintf(format, v...)).To(Equal("waiting for ops manager web server boots up..."))

			format, v = logger.PrintfArgsForCall(4)
			Expect(fmt.Sprintf(format, v...)).To(Equal("finished import"))
		})

		It("it only retries 3 times before giving up", func(done Done) {
			fakeService.EnsureAvailabilityStub = func(api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error) {
				if fakeService.EnsureAvailabilityCallCount() > 2 {
					return api.EnsureAvailabilityOutput{}, fmt.Errorf("connection refused")
				}

				return api.EnsureAvailabilityOutput{Status: api.EnsureAvailabilityStatusUnstarted}, nil
			}

			err := executeCommand(command, []string{"--polling-interval", "0",
				"--installation", installationFile,
			})
			Expect(err).To(MatchError(ContainSubstring("could not check Ops Manager Status:")))
			close(done)
		}, 1)
	})

	When("the global decryption-passphrase is not provided", func() {
		It("returns an error", func() {
			command := commands.NewImportInstallation(multipart, fakeService, "", logger)
			err := executeCommand(command, []string{"--polling-interval", "0", "--installation", "installation.zip"})
			Expect(err).To(MatchError("the global decryption-passphrase argument is required for this command"))
		})
	})

	When("the --installation provided is a file that does not exist", func() {
		It("returns an error", func() {
			command := commands.NewImportInstallation(multipart, fakeService, "passphrase", logger)
			err := executeCommand(command, []string{"--installation", "does-not-exist.zip"})
			Expect(err).To(MatchError("file: \"does-not-exist.zip\" does not exist. Please check the name and try again."))
		})
	})

	When("the --installation provided is not a valid zip file", func() {
		var notZipFile string
		BeforeEach(func() {
			tmpFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())
			notZipFile = tmpFile.Name()
		})

		AfterEach(func() {
			os.Remove(notZipFile)
		})

		It("returns an error", func() {
			command := commands.NewImportInstallation(multipart, fakeService, "passphrase", logger)
			err := executeCommand(command, []string{"--installation", notZipFile})
			Expect(err).To(MatchError(fmt.Sprintf("file: \"%s\" is not a valid zip file", notZipFile)))
		})
	})

	When("the --installation provided does not have required installation.yml", func() {
		var invalidInstallation string
		BeforeEach(func() {
			invalidInstallation = createZipFile([]struct{ Name, Body string }{})
		})

		AfterEach(func() {
			os.Remove(invalidInstallation)
		})

		It("returns an error", func() {
			command := commands.NewImportInstallation(multipart, fakeService, "passphrase", logger)
			err := executeCommand(command, []string{"--installation", invalidInstallation})
			expectedErrorTemplate := "file: \"%s\" is not a valid installation file. Validate that the provided installation file is correct, or run \"om export-installation\" and try again."
			Expect(err).To(MatchError(fmt.Sprintf(expectedErrorTemplate, invalidInstallation)))
		})
	})

	When("the ensure_availability endpoint returns an error", func() {
		It("returns an error", func() {
			fakeService.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{}, errors.New("some error"))
			command := commands.NewImportInstallation(multipart, fakeService, "some-passphrase", logger)
			err := executeCommand(command, []string{"--polling-interval", "0", "--installation", installationFile})
			Expect(err).To(MatchError("could not check Ops Manager status: some error"))
		})
	})

	When("the file cannot be opened", func() {
		It("returns an error", func() {
			fakeService.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
				Status: api.EnsureAvailabilityStatusUnstarted,
			}, nil)
			command := commands.NewImportInstallation(multipart, fakeService, "some-passphrase", logger)
			multipart.AddFileReturns(errors.New("bad file"))

			err := executeCommand(command, []string{"--polling-interval", "0", "--installation", installationFile})
			Expect(err).To(MatchError("failed to load installation: bad file"))
		})
	})

	When("the installation cannot be imported", func() {
		It("returns an error", func() {
			fakeService.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
				Status: api.EnsureAvailabilityStatusUnstarted,
			}, nil)
			command := commands.NewImportInstallation(multipart, fakeService, "some-passphrase", logger)
			fakeService.UploadInstallationAssetCollectionReturns(errors.New("some installation error"))

			err := executeCommand(command, []string{"--polling-interval", "0", "--installation", installationFile})
			Expect(err).To(MatchError("failed to import installation: some installation error"))
		})
	})
})
