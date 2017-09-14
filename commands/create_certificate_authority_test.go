package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("CreateCertificateAuthority", func() {
	var (
		fakeTableWriter                 *fakes.TableWriter
		fakeCertificateAuthorityService *fakes.CertificateAuthoritiesService
		command                         commands.CreateCertificateAuthority
	)

	BeforeEach(func() {
		fakeTableWriter = &fakes.TableWriter{}
		fakeCertificateAuthorityService = &fakes.CertificateAuthoritiesService{}
		command = commands.NewCreateCertificateAuthority(fakeCertificateAuthorityService, fakeTableWriter)
	})

	Describe("Execute", func() {
		It("makes a request to the Opsman to create a certificate authority", func() {
			err := command.Execute([]string{
				"--certificate-pem", "some CertPem",
				"--private-key-pem", "some PrivateKey",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeCertificateAuthorityService.CreateCallCount()).To(Equal(1))
			Expect(fakeCertificateAuthorityService.CreateArgsForCall(0)).To(Equal(api.CertificateAuthorityInput{
				CertPem:       "some CertPem",
				PrivateKeyPem: "some PrivateKey",
			}))
		})

		It("prints a table containing the certificate authority that was created", func() {
			fakeCertificateAuthorityService.CreateReturns(api.CA{GUID: "some GUID", Issuer: "some Issuer",
				CreatedOn: "2017-09-12", ExpiresOn: "2018-09-12", Active: true, CertPEM: "some CertPem"}, nil)

			err := command.Execute([]string{
				"--certificate-pem", "some CertPem",
				"--private-key-pem", "some PrivateKey",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeTableWriter.SetAutoWrapTextCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetAutoWrapTextArgsForCall(0)).To(BeFalse())

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"id", "issuer", "active", "created on", "expires on", "certicate pem"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(1))
			Expect(fakeTableWriter.AppendArgsForCall(0)).To(Equal([]string{"some GUID", "some Issuer",
				"true", "2017-09-12", "2018-09-12", "some CertPem"}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))

		})

		Context("failure cases", func() {
			Context("when the service fails to create a certificate", func() {
				It("returns an error", func() {
					fakeCertificateAuthorityService.CreateReturns(api.CA{}, errors.New("failed to create certificate"))

					err := command.Execute([]string{
						"--certificate-pem", "some CertPem",
						"--private-key-pem", "some PrivateKey",
					})
					Expect(err).To(MatchError("failed to create certificate"))
				})
			})
			Context("when an unknown flag is provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--badflag"})
					Expect(err).To(MatchError("could not parse create-certificate-authority flags: flag provided but not defined: -badflag"))
				})
			})
			Context("when the certificate flag is not provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{
						"--private-key-pem", "some PrivateKey",
					})
					Expect(err).To(MatchError("error: certificate-pem is missing. Please see usage for more information."))
				})
			})
			Context("when the private key flag is not provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{
						"--certificate-pem", "some CertPem",
					})
					Expect(err).To(MatchError("error: private-key-pem is missing. Please see usage for more information."))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage info", func() {
			usage := command.Usage()
			Expect(usage).To(Equal(commands.Usage{
				Description:      "This authenticated command creates a certificate authority on the Ops Manager with the given cert and key",
				ShortDescription: "creates a certificate authority on the Ops Manager",
				Flags:            command.Options,
			}))
		})
	})
})
