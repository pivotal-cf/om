package commands_test

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Certificate Authorities", func() {
	var (
		certificateAuthorities            commands.CertificateAuthorities
		fakeCertificateAuthoritiesService *fakes.CertificateAuthoritiesService
		fakePresenter                     *presenterfakes.Presenter
	)

	BeforeEach(func() {
		fakeCertificateAuthoritiesService = &fakes.CertificateAuthoritiesService{}
		fakePresenter = &presenterfakes.Presenter{}
		certificateAuthorities = commands.NewCertificateAuthorities(fakeCertificateAuthoritiesService, fakePresenter)
	})

	Describe("Execute", func() {
		It("requests certificate authorities from the server", func() {
			err := certificateAuthorities.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeCertificateAuthoritiesService.ListCallCount()).To(Equal(1))
		})

		Context("when request for certificate authorities fails", func() {
			It("returns an error", func() {
				fakeCertificateAuthoritiesService.ListReturns(
					api.CertificateAuthoritiesOutput{},
					fmt.Errorf("could not get certificate authorities"),
				)

				err := certificateAuthorities.Execute([]string{})
				Expect(err).To(MatchError("could not get certificate authorities"))
			})
		})

		It("prints the certificate authorities to a table", func() {
			certificateAuthoritiesOutput := []api.CA{
				{
					GUID:      "some-guid",
					Issuer:    "Pivotal",
					CreatedOn: "2017-01-09",
					ExpiresOn: "2021-01-09",
					Active:    true,
					CertPEM:   "-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI....",
				},
				{
					GUID:      "other-guid",
					Issuer:    "Customer",
					CreatedOn: "2017-01-10",
					ExpiresOn: "2021-01-10",
					Active:    false,
					CertPEM:   "-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBBhI....",
				},
			}

			fakeCertificateAuthoritiesService.ListReturns(
				api.CertificateAuthoritiesOutput{certificateAuthoritiesOutput},
				nil,
			)

			err := certificateAuthorities.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakePresenter.PresentCertificateAuthoritiesCallCount()).To(Equal(1))
			Expect(fakePresenter.PresentCertificateAuthoritiesArgsForCall(0)).To(Equal(certificateAuthoritiesOutput))
		})
	})

	Describe("Usage", func() {
		It("returns usage", func() {
			usage := certificateAuthorities.Usage()

			Expect(usage).To(Equal(jhanda.Usage{
				Description:      "lists certificates managed by Ops Manager",
				ShortDescription: "lists certificates managed by Ops Manager",
			}))
		})
	})
})
