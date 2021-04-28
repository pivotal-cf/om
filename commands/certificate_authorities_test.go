package commands_test

import (
	"errors"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Certificate Authorities", func() {
	var (
		command                           *commands.CertificateAuthorities
		fakeCertificateAuthoritiesService *fakes.CertificateAuthoritiesService
		fakePresenter                     *presenterfakes.FormattedPresenter
	)

	BeforeEach(func() {
		fakeCertificateAuthoritiesService = &fakes.CertificateAuthoritiesService{}
		fakePresenter = &presenterfakes.FormattedPresenter{}
		command = commands.NewCertificateAuthorities(fakeCertificateAuthoritiesService, fakePresenter)
	})

	Describe("Execute", func() {
		var certificateAuthoritiesOutput []api.CA

		BeforeEach(func() {
			certificateAuthoritiesOutput = []api.CA{
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

			fakeCertificateAuthoritiesService.ListCertificateAuthoritiesReturns(
				api.CertificateAuthoritiesOutput{CAs: certificateAuthoritiesOutput},
				nil,
			)
		})

		It("prints the certificate authorities to a table", func() {
			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeCertificateAuthoritiesService.ListCertificateAuthoritiesCallCount()).To(Equal(1))

			Expect(fakePresenter.PresentCertificateAuthoritiesCallCount()).To(Equal(1))
			Expect(fakePresenter.PresentCertificateAuthoritiesArgsForCall(0)).To(Equal(certificateAuthoritiesOutput))
		})

		When("the format flag is provided", func() {
			It("calls the presenter to set the json format", func() {
				err := executeCommand(command, []string{
					"--format", "json",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
				Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("json"))
			})
		})

		When("request for certificate authorities fails", func() {
			It("returns an error", func() {
				fakeCertificateAuthoritiesService.ListCertificateAuthoritiesReturns(
					api.CertificateAuthoritiesOutput{},
					errors.New("could not get certificate authorities"),
				)

				err := executeCommand(command, []string{})
				Expect(err).To(MatchError("could not get certificate authorities"))
			})
		})
	})
})
