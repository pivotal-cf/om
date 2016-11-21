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

var _ = Describe("ImportInstallation", func() {
	var (
		installationService *fakes.InstallationAssetImporterService
		setupService        *fakes.SetupService
		multipart           *fakes.Multipart
		logger              *fakes.Logger
	)

	BeforeEach(func() {
		multipart = &fakes.Multipart{}
		installationService = &fakes.InstallationAssetImporterService{}
		logger = &fakes.Logger{}
	})

	It("imports an installation", func() {
		submission := formcontent.ContentSubmission{
			Length:      10,
			Content:     ioutil.NopCloser(strings.NewReader("")),
			ContentType: "some content-type",
		}
		multipart.FinalizeReturns(submission, nil)
		setupService = &fakes.SetupService{}
		setupService.EnsureAvailabilityCall.Returns.Outputs = []api.EnsureAvailabilityOutput{
			{Status: api.EnsureAvailabilityStatusUnstarted},
			{Status: api.EnsureAvailabilityStatusPending},
			{Status: api.EnsureAvailabilityStatusPending},
			{Status: api.EnsureAvailabilityStatusPending},
			{Status: api.EnsureAvailabilityStatusComplete},
		}

		command := commands.NewImportInstallation(multipart, installationService, setupService, logger)

		err := command.Execute([]string{
			"--installation", "/path/to/some-installation",
			"--decryption-passphrase", "some-passphrase",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(setupService.EnsureAvailabilityCall.CallCount).To(Equal(5))

		key, file := multipart.AddFileArgsForCall(0)
		Expect(key).To(Equal("installation[file]"))
		Expect(file).To(Equal("/path/to/some-installation"))

		key, val := multipart.AddFieldArgsForCall(0)
		Expect(key).To(Equal("passphrase"))
		Expect(val).To(Equal("some-passphrase"))

		Expect(multipart.FinalizeCallCount()).To(Equal(1))

		Expect(installationService.ImportArgsForCall(0)).To(Equal(api.ImportInstallationInput{
			ContentLength: 10,
			Installation:  ioutil.NopCloser(strings.NewReader("")),
			ContentType:   "some content-type",
		}))

		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("processing installation"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("beginning installation import to Ops Manager"))

		format, v = logger.PrintfArgsForCall(2)
		Expect(fmt.Sprintf(format, v...)).To(Equal("finished import"))
	})

	Context("when the Ops Manager is already configured", func() {
		It("returns an error", func() {
			setupService = &fakes.SetupService{}
			setupService.EnsureAvailabilityCall.Returns.Outputs = []api.EnsureAvailabilityOutput{
				{Status: api.EnsureAvailabilityStatusComplete},
			}
			command := commands.NewImportInstallation(multipart, installationService, setupService, logger)

			err := command.Execute([]string{
				"--installation", "/path/to/some-installation",
				"--decryption-passphrase", "some-passphrase",
			})
			Expect(err).To(MatchError(ContainSubstring("cannot import installation to an Ops Manager that is already configured")))
			Expect(setupService.EnsureAvailabilityCall.CallCount).To(Equal(1))
		})
	})

	Context("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewImportInstallation(multipart, installationService, setupService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse import-installation flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when the passphrase is not provided", func() {
			It("returns an error", func() {
				command := commands.NewImportInstallation(multipart, installationService, setupService, logger)
				err := command.Execute([]string{"--installation", "/some/path"})
				Expect(err).To(MatchError("could not parse import-installation flags: decryption passphrase not provided"))
			})
		})

		Context("when the ensure_availability endpoint returns an error", func() {
			It("returns an error", func() {
				setupService = &fakes.SetupService{}
				setupService.EnsureAvailabilityCall.Returns.Outputs = []api.EnsureAvailabilityOutput{}
				setupService.EnsureAvailabilityCall.Returns.Errors = []error{errors.New("some error")}
				command := commands.NewImportInstallation(multipart, installationService, setupService, logger)
				err := command.Execute([]string{"--installation", "/some/path", "--decryption-passphrase", "some-passphrase"})
				Expect(err).To(MatchError("could not check Ops Manager status: some error"))
			})
		})

		Context("when the file cannot be opened", func() {
			It("returns an error", func() {
				setupService = &fakes.SetupService{}
				setupService.EnsureAvailabilityCall.Returns.Outputs = []api.EnsureAvailabilityOutput{
					{Status: api.EnsureAvailabilityStatusUnstarted},
				}
				command := commands.NewImportInstallation(multipart, installationService, setupService, logger)
				multipart.AddFileReturns(errors.New("bad file"))

				err := command.Execute([]string{"--installation", "/some/path", "--decryption-passphrase", "some-passphrase"})
				Expect(err).To(MatchError("failed to load installation: bad file"))
			})
		})

		Context("when the installation cannot be imported", func() {
			It("returns and error", func() {
				setupService = &fakes.SetupService{}
				setupService.EnsureAvailabilityCall.Returns.Outputs = []api.EnsureAvailabilityOutput{
					{Status: api.EnsureAvailabilityStatusUnstarted},
				}
				command := commands.NewImportInstallation(multipart, installationService, setupService, logger)
				installationService.ImportReturns(errors.New("some installation error"))

				err := command.Execute([]string{"--installation", "/some/path", "--decryption-passphrase", "some-passphrase"})
				Expect(err).To(MatchError("failed to import installation: some installation error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewImportInstallation(nil, nil, nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This unauthenticated command attempts to import an installation to the Ops Manager targeted.",
				ShortDescription: "imports a given installation to the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
