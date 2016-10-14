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
		installationService *fakes.InstallationService
		multipart           *fakes.Multipart
		logger              *fakes.OtherLogger
	)

	BeforeEach(func() {
		multipart = &fakes.Multipart{}
		installationService = &fakes.InstallationService{}
		logger = &fakes.OtherLogger{}
	})

	It("imports an installation", func() {
		submission := formcontent.ContentSubmission{
			Length:      10,
			Content:     ioutil.NopCloser(strings.NewReader("")),
			ContentType: "some content-type",
		}
		multipart.CreateReturns(submission, nil)

		command := commands.NewImportInstallation(multipart, installationService, logger)

		err := command.Execute([]string{
			"--installation", "/path/to/some-installation",
			"--decryption-passphrase", "some-passphrase",
		})
		Expect(err).NotTo(HaveOccurred())

		key, file := multipart.AddFileArgsForCall(0)
		Expect(key).To(Equal("installation[file]"))
		Expect(file).To(Equal("/path/to/some-installation"))

		key, val := multipart.AddFieldArgsForCall(0)
		Expect(key).To(Equal("passphrase"))
		Expect(val).To(Equal("some-passphrase"))

		Expect(multipart.CreateCallCount()).To(Equal(1))

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

	Context("failure cases", func() {
		Context("when an unkwown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewImportInstallation(multipart, installationService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse import-installation flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when the file cannot be opened", func() {
			It("returns an error", func() {
				command := commands.NewImportInstallation(multipart, installationService, logger)
				multipart.AddFileReturns(errors.New("bad file"))

				err := command.Execute([]string{"--installation", "/some/path"})
				Expect(err).To(MatchError("failed to load installation: bad file"))
			})
		})

		Context("when the passphrase is not provided", func() {
		})

		Context("when the installation cannot be imported", func() {
			It("returns and error", func() {
				command := commands.NewImportInstallation(multipart, installationService, logger)
				installationService.ImportReturns(errors.New("some installation error"))

				err := command.Execute([]string{"--installation", "/some/path"})
				Expect(err).To(MatchError("failed to import installation: some installation error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewImportInstallation(nil, nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This unauthenticated command attempts to import an installation to the Ops Manager targeted.",
				ShortDescription: "imports a given installation to the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
