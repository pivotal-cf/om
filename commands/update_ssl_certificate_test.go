package commands_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("UpdateSSLCertificate", func() {
	var (
		fakeLogger  *fakes.Logger
		fakeService *fakes.UpdateSSLCertificateService
		command     commands.UpdateSSLCertificate
	)

	BeforeEach(func() {
		fakeService = &fakes.UpdateSSLCertificateService{}
		fakeLogger = &fakes.Logger{}
		command = commands.NewUpdateSSLCertificate(os.Environ, fakeService, fakeLogger)
	})

	Describe("Execute", func() {
		It("makes a request to the Opsman to apply a custom certificate", func() {
			err := command.Execute([]string{
				"--certificate-pem", "some CertPem",
				"--private-key-pem", "some PrivateKey",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(1))
			Expect(fakeService.UpdateSSLCertificateArgsForCall(0)).To(Equal(api.SSLCertificateSettings{
				CertPem:       "some CertPem",
				PrivateKeyPem: "some PrivateKey",
			}))
		})

		It("prints a success message saying the custom cert was applied", func() {
			fakeService.UpdateSSLCertificateReturns(nil)

			err := command.Execute([]string{
				"--certificate-pem", "some CertPem",
				"--private-key-pem", "some PrivateKey",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeLogger.PrintfCallCount()).To(Equal(2))
			format, content := fakeLogger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("Successfully applied custom SSL Certificate.\n"))
			format, content = fakeLogger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("Please allow about 1 min for the new certificate to take effect.\n"))
		})

		Context("with a config file and no vars", func() {
			var (
				configFile *os.File
				err        error
			)

			const config = `---
certificate-pem: some CertPem
private-key-pem: some PrivateKey
`

			BeforeEach(func() {
				configFile, err = ioutil.TempFile("", "")
				Expect(err).ToNot(HaveOccurred())

				_, err = configFile.WriteString(config)
				Expect(err).ToNot(HaveOccurred())

				err = configFile.Close()
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				err = os.RemoveAll(configFile.Name())
				Expect(err).ToNot(HaveOccurred())
			})

			It("makes a request to the Opsman to apply a custom certificate", func() {
				err = command.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(1))
				Expect(fakeService.UpdateSSLCertificateArgsForCall(0)).To(Equal(api.SSLCertificateSettings{
					CertPem:       "some CertPem",
					PrivateKeyPem: "some PrivateKey",
				}))
			})

			It("prints a success message saying the custom cert was applied", func() {
				fakeService.UpdateSSLCertificateReturns(nil)

				err = command.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeLogger.PrintfCallCount()).To(Equal(2))
				format, content := fakeLogger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("Successfully applied custom SSL Certificate.\n"))
				format, content = fakeLogger.PrintfArgsForCall(1)

				Expect(fmt.Sprintf(format, content...)).To(Equal("Please allow about 1 min for the new certificate to take effect.\n"))
			})
		})

		Context("with a config file and vars", func() {
			var (
				configFile *os.File
				varsFile   *os.File
				err        error
			)

			const config = `---
certificate-pem: ((cert))
private-key-pem: ((pkey))
`

			const vars = `---
cert: some CertPem
pkey: some PrivateKey
`

			BeforeEach(func() {
				configFile, err = ioutil.TempFile("", "")
				Expect(err).ToNot(HaveOccurred())

				_, err = configFile.WriteString(config)
				Expect(err).ToNot(HaveOccurred())

				err = configFile.Close()
				Expect(err).ToNot(HaveOccurred())

				varsFile, err = ioutil.TempFile("", "")
				Expect(err).ToNot(HaveOccurred())

				_, err = varsFile.WriteString(vars)
				Expect(err).ToNot(HaveOccurred())

				err = varsFile.Close()
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				err = os.RemoveAll(configFile.Name())
				Expect(err).ToNot(HaveOccurred())

				err = os.RemoveAll(varsFile.Name())
				Expect(err).ToNot(HaveOccurred())
			})

			It("makes a request to the Opsman to apply a custom certificate", func() {
				err = command.Execute([]string{
					"--config", configFile.Name(),
					"--vars-file", varsFile.Name(),
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(1))
				Expect(fakeService.UpdateSSLCertificateArgsForCall(0)).To(Equal(api.SSLCertificateSettings{
					CertPem:       "some CertPem",
					PrivateKeyPem: "some PrivateKey",
				}))
			})

			It("prints a success message saying the custom cert was applied", func() {
				fakeService.UpdateSSLCertificateReturns(nil)

				err = command.Execute([]string{
					"--config", configFile.Name(),
					"--vars-file", varsFile.Name(),
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeLogger.PrintfCallCount()).To(Equal(2))
				format, content := fakeLogger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("Successfully applied custom SSL Certificate.\n"))
				format, content = fakeLogger.PrintfArgsForCall(1)

				Expect(fmt.Sprintf(format, content...)).To(Equal("Please allow about 1 min for the new certificate to take effect.\n"))
			})
		})

		Context("failure cases", func() {
			When("the service fails to apply a certificate", func() {
				It("returns an error", func() {
					fakeService.UpdateSSLCertificateReturns(errors.New("failed to apply certificate"))

					err := command.Execute([]string{
						"--certificate-pem", "some CertPem",
						"--private-key-pem", "some PrivateKey",
					})
					Expect(err).To(MatchError("failed to apply certificate"))
				})
			})

			When("an unknown flag is provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--badflag"})
					Expect(err).To(MatchError("could not parse update-ssl-certificate flags: flag provided but not defined: -badflag"))
				})
			})

			When("the certificate is not provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{
						"--private-key-pem", "some PrivateKey",
					})
					Expect(err).To(MatchError("could not parse update-ssl-certificate flags: missing required flag \"--certificate-pem\""))
				})
			})

			When("the private key flag is not provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{
						"--certificate-pem", "some CertPem",
					})
					Expect(err).To(MatchError("could not parse update-ssl-certificate flags: missing required flag \"--private-key-pem\""))
				})
			})
		})
	})
})
