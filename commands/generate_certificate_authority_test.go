package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
)

var _ = Describe("GenerateCertificateAuthority", func() {
	var (
		fakePresenter *presenterfakes.FormattedPresenter
		fakeService   *fakes.GenerateCertificateAuthorityService
		command       *commands.GenerateCertificateAuthority
	)

	BeforeEach(func() {
		fakePresenter = &presenterfakes.FormattedPresenter{}
		fakeService = &fakes.GenerateCertificateAuthorityService{}
		command = commands.NewGenerateCertificateAuthority(fakeService, fakePresenter)
	})

	Describe("Execute", func() {
		var certificateAuthority api.CA

		BeforeEach(func() {
			certificateAuthority = api.CA{
				GUID:      "some GUID",
				Issuer:    "some Issuer",
				CreatedOn: "2017-09-12",
				ExpiresOn: "2018-09-12",
				Active:    true,
				CertPEM:   "some CertPem",
			}

			fakeService.GenerateCertificateAuthorityReturns(certificateAuthority, nil)
		})

		It("makes a request to the Opsman to generate a certificate authority and prints to a table", func() {
			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.GenerateCertificateAuthorityCallCount()).To(Equal(1))

			Expect(fakePresenter.PresentCertificateAuthorityCallCount()).To(Equal(1))
			Expect(fakePresenter.PresentCertificateAuthorityArgsForCall(0)).To(Equal(certificateAuthority))
		})

		When("the format flag is provided", func() {
			It("sets the format on the presenter", func() {
				err := command.Execute([]string{
					"--format", "json",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
				Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("json"))
			})
		})

		Context("failure cases", func() {
			When("an unknown flag is passed", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--unknown-flag"})
					Expect(err).To(MatchError("could not parse generate-certificate-authority flags: flag provided but not defined: -unknown-flag"))
				})
			})

			It("returns an error when the service fails to generate a certificate", func() {
				fakeService.GenerateCertificateAuthorityReturns(api.CA{}, errors.New("failed to generate certificate"))

				err := command.Execute([]string{})
				Expect(err).To(MatchError("failed to generate certificate"))
			})
		})
	})
})
