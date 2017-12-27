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

var _ = Describe("GenerateCertificateAuthority", func() {
	var (
		fakePresenter                   *presenterfakes.Presenter
		fakeCertificateAuthorityService *fakes.CertificateAuthorityGenerator
		command                         commands.GenerateCertificateAuthority
	)

	BeforeEach(func() {
		fakePresenter = &presenterfakes.Presenter{}
		fakeCertificateAuthorityService = &fakes.CertificateAuthorityGenerator{}
		command = commands.NewGenerateCertificateAuthority(fakeCertificateAuthorityService, fakePresenter)
	})

	Describe("Execute", func() {
		It("makes a request to the Opsman to generate a certificate authority", func() {
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeCertificateAuthorityService.GenerateCallCount()).To(Equal(1))
		})

		It("prints a table containing the certificate authority that was generated", func() {
			certificateAuthority := api.CA{
				GUID:      "some GUID",
				Issuer:    "some Issuer",
				CreatedOn: "2017-09-12",
				ExpiresOn: "2018-09-12",
				Active:    true,
				CertPEM:   "some CertPem",
			}

			fakeCertificateAuthorityService.GenerateReturns(certificateAuthority, nil)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakePresenter.PresentCertificateAuthorityCallCount()).To(Equal(1))
			Expect(fakePresenter.PresentCertificateAuthorityArgsForCall(0)).To(Equal(certificateAuthority))

		})

		Context("failure cases", func() {
			It("returns an error when the service fails to generate a certificate", func() {
				fakeCertificateAuthorityService.GenerateReturns(api.CA{}, errors.New("failed to generate certificate"))

				err := command.Execute([]string{})
				Expect(err).To(MatchError("failed to generate certificate"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage info", func() {
			usage := command.Usage()
			Expect(usage).To(Equal(jhanda.Usage{
				Description:      "This authenticated command generates a certificate authority on the Ops Manager",
				ShortDescription: "generates a certificate authority on the Opsman",
			}))
		})
	})
})
