package commands_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pivotal-cf/jhanda"
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
		fakeService *fakes.ImportInstallationService
		multipart   *fakes.Multipart
		logger      *fakes.Logger
	)

	BeforeEach(func() {
		multipart = &fakes.Multipart{}
		fakeService = &fakes.ImportInstallationService{}
		logger = &fakes.Logger{}
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

		err := command.Execute([]string{"--polling-interval", "0",
			"--installation", "/path/to/some-installation",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeService.EnsureAvailabilityCallCount()).To(Equal(5))

		key, file := multipart.AddFileArgsForCall(0)
		Expect(key).To(Equal("installation[file]"))
		Expect(file).To(Equal("/path/to/some-installation"))

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

	Context("when the Ops Manager is already configured", func() {
		It("prints a helpful message", func() {
			fakeService.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
				Status: api.EnsureAvailabilityStatusComplete,
			}, nil)

			command := commands.NewImportInstallation(multipart, fakeService, "some-passphrase", logger)

			err := command.Execute([]string{"--polling-interval", "0",
				"--installation", "/path/to/some-installation",
			})
			Expect(err).NotTo(HaveOccurred())

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(Equal("Ops Manager is already configured"))
			Expect(fakeService.EnsureAvailabilityCallCount()).To(Equal(1))
		})
	})

	Context("when config file is provided", func() {
		var configFile *os.File

		BeforeEach(func() {
			var err error
			configContent := `
installation: /path/to/some-installation
`
			configFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = configFile.WriteString(configContent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("reads configuration from config file", func() {
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

			err := command.Execute([]string{"--polling-interval", "0",
				"--config", configFile.Name(),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeService.EnsureAvailabilityCallCount()).To(Equal(5))

			key, file := multipart.AddFileArgsForCall(0)
			Expect(key).To(Equal("installation[file]"))
			Expect(file).To(Equal("/path/to/some-installation"))

			key, val := multipart.AddFieldArgsForCall(0)
			Expect(key).To(Equal("passphrase"))
			Expect(val).To(Equal("some-passphrase"))
		})

		It("is overridden by commandline flags", func() {
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

			err := command.Execute([]string{"--polling-interval", "0",
				"--config", configFile.Name(),
				"--installation", "/path/to/some-installation1",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeService.EnsureAvailabilityCallCount()).To(Equal(5))

			key, file := multipart.AddFileArgsForCall(0)
			Expect(key).To(Equal("installation[file]"))
			Expect(file).To(Equal("/path/to/some-installation1"))

			key, val := multipart.AddFieldArgsForCall(0)
			Expect(key).To(Equal("passphrase"))
			Expect(val).To(Equal("some-passphrase"))
		})
	})

	Context("when EnsureAvailability returns 'connection refused'", func() {
		var command commands.ImportInstallation

		BeforeEach(func() {
			submission := formcontent.ContentSubmission{
				Content:       ioutil.NopCloser(strings.NewReader("")),
				ContentType:   "some content-type",
				ContentLength: 10,
			}
			multipart.FinalizeReturns(submission)

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

			command = commands.NewImportInstallation(multipart, fakeService, "some-passphrase", logger)
		})

		It("it retries on the specified polling interval to allow nginx time to boot up", func() {
			err := command.Execute([]string{"--polling-interval", "0",
				"--installation", "/path/to/some-installation",
			})
			Expect(err).NotTo(HaveOccurred())
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
	})

	Context("failure cases", func() {
		Context("when the global decryption-passphrase is not provided", func() {
			It("returns an error", func() {
				command := commands.NewImportInstallation(multipart, fakeService, "", logger)
				err := command.Execute([]string{"--polling-interval", "0"})
				Expect(err).To(MatchError("the global decryption-passphrase argument is required for this command"))
			})
		})

		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewImportInstallation(multipart, fakeService, "passphrase", logger)
				err := command.Execute([]string{"--polling-interval", "0", "--badflag"})
				Expect(err).To(MatchError("could not parse import-installation flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when config file cannot be opened", func() {
			It("returns an error", func() {
				command := commands.NewImportInstallation(multipart, fakeService, "passphrase", logger)
				err := command.Execute([]string{"--config", "something"})
				Expect(err).To(MatchError("could not parse import-installation flags: could not load the config file: open something: no such file or directory"))

			})
		})

		Context("when the --installation flag is missing", func() {
			It("returns an error", func() {
				command := commands.NewImportInstallation(multipart, fakeService, "passphrase", logger)
				err := command.Execute([]string{"--polling-interval", "0"})
				Expect(err).To(MatchError("could not parse import-installation flags: missing required flag \"--installation\""))
			})
		})

		Context("when the ensure_availability endpoint returns an error", func() {
			It("returns an error", func() {
				fakeService.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{}, errors.New("some error"))
				command := commands.NewImportInstallation(multipart, fakeService, "some-passphrase", logger)
				err := command.Execute([]string{"--polling-interval", "0", "--installation", "/some/path"})
				Expect(err).To(MatchError("could not check Ops Manager status: some error"))
			})
		})

		Context("when the file cannot be opened", func() {
			It("returns an error", func() {
				fakeService.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusUnstarted,
				}, nil)
				command := commands.NewImportInstallation(multipart, fakeService, "some-passphrase", logger)
				multipart.AddFileReturns(errors.New("bad file"))

				err := command.Execute([]string{"--polling-interval", "0", "--installation", "/some/path"})
				Expect(err).To(MatchError("failed to load installation: bad file"))
			})
		})

		Context("when the installation cannot be imported", func() {
			It("returns an error", func() {
				fakeService.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusUnstarted,
				}, nil)
				command := commands.NewImportInstallation(multipart, fakeService, "some-passphrase", logger)
				fakeService.UploadInstallationAssetCollectionReturns(errors.New("some installation error"))

				err := command.Execute([]string{"--polling-interval", "0", "--installation", "/some/path"})
				Expect(err).To(MatchError("failed to import installation: some installation error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewImportInstallation(nil, nil, "", nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This unauthenticated command attempts to import an installation to the Ops Manager targeted.",
				ShortDescription: "imports a given installation to the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
