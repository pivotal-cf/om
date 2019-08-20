package commands_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
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
		command = commands.NewUpdateSSLCertificate(fakeService, fakeLogger)
	})

	Describe("Execute", func() {
		It("makes a request to the Opsman to apply a custom certificate", func() {
			err := command.Execute([]string{
				"--certificate-pem", "some CertPem",
				"--private-key-pem", "some PrivateKey",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(1))
			Expect(fakeService.UpdateSSLCertificateArgsForCall(0)).To(Equal(api.SSLCertificateInput{
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
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLogger.PrintfCallCount()).To(Equal(2))
			format, content := fakeLogger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("Successfully applied custom SSL Certificate.\n"))
			format, content = fakeLogger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("Please allow about 1 min for the new certificate to take effect.\n"))
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

			When("the certificate flag is not provided", func() {
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

	Describe("Usage", func() {
		It("returns usage info", func() {
			usage := command.Usage()
			Expect(usage).To(Equal(jhanda.Usage{
				Description:      "This authenticated command updates the SSL Certificate on the Ops Manager with the given cert and key",
				ShortDescription: "updates the SSL Certificate on the Ops Manager",
				Flags:            command.Options,
			}))
		})
	})
})
