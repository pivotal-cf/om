package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
)

var _ = Describe("CreateCertificateAuthority", func() {
	var (
		fakePresenter                   *presenterfakes.Presenter
		fakeCertificateAuthorityService *fakes.CertificateAuthorityCreator
		command                         commands.CreateCertificateAuthority
	)

	BeforeEach(func() {
		fakePresenter = &presenterfakes.Presenter{}
		fakeCertificateAuthorityService = &fakes.CertificateAuthorityCreator{}
		command = commands.NewCreateCertificateAuthority(fakeCertificateAuthorityService, fakePresenter)
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
			ca := api.CA{
				GUID:      "some GUID",
				Issuer:    "some Issuer",
				CreatedOn: "2017-09-12",
				ExpiresOn: "2018-09-12",
				Active:    true,
				CertPEM:   "some CertPem",
			}

			fakeCertificateAuthorityService.CreateReturns(ca, nil)

			err := command.Execute([]string{
				"--certificate-pem", "some CertPem",
				"--private-key-pem", "some PrivateKey",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakePresenter.PresentCertificateAuthorityCallCount()).To(Equal(1))
			Expect(fakePresenter.PresentCertificateAuthorityArgsForCall(0)).To(Equal(ca))
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
					Expect(err).To(MatchError("could not parse create-certificate-authority flags: missing required flag \"--certificate-pem\""))
				})
			})

			Context("when the private key flag is not provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{
						"--certificate-pem", "some CertPem",
					})
					Expect(err).To(MatchError("could not parse create-certificate-authority flags: missing required flag \"--private-key-pem\""))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage info", func() {
			usage := command.Usage()
			Expect(usage).To(Equal(jhanda.Usage{
				Description:      "This authenticated command creates a certificate authority on the Ops Manager with the given cert and key",
				ShortDescription: "creates a certificate authority on the Ops Manager",
				Flags:            command.Options,
			}))
		})
	})
})
