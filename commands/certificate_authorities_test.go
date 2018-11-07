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
		fakePresenter                     *presenterfakes.FormattedPresenter
	)

	BeforeEach(func() {
		fakeCertificateAuthoritiesService = &fakes.CertificateAuthoritiesService{}
		fakePresenter = &presenterfakes.FormattedPresenter{}
		certificateAuthorities = commands.NewCertificateAuthorities(fakeCertificateAuthoritiesService, fakePresenter)
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
			err := certificateAuthorities.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeCertificateAuthoritiesService.ListCertificateAuthoritiesCallCount()).To(Equal(1))

			Expect(fakePresenter.PresentCertificateAuthoritiesCallCount()).To(Equal(1))
			Expect(fakePresenter.PresentCertificateAuthoritiesArgsForCall(0)).To(Equal(certificateAuthoritiesOutput))
		})

		Context("when the format flag is provided", func() {
			It("calls the presenter to set the json format", func() {
				err := certificateAuthorities.Execute([]string{
					"--format", "json",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
				Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("json"))
			})
		})

		Context("when the flag cannot parsed", func() {
			It("returns an error", func() {
				err := certificateAuthorities.Execute([]string{"--bogus", "nothing"})
				Expect(err).To(MatchError(
					"could not parse certificate-authorities flags: flag provided but not defined: -bogus",
				))
			})
		})

		Context("when request for certificate authorities fails", func() {
			It("returns an error", func() {
				fakeCertificateAuthoritiesService.ListCertificateAuthoritiesReturns(
					api.CertificateAuthoritiesOutput{},
					fmt.Errorf("could not get certificate authorities"),
				)

				err := certificateAuthorities.Execute([]string{})
				Expect(err).To(MatchError("could not get certificate authorities"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage", func() {
			usage := certificateAuthorities.Usage()

			Expect(usage).To(Equal(jhanda.Usage{
				Description:      "lists certificates managed by Ops Manager",
				ShortDescription: "lists certificates managed by Ops Manager",
				Flags:            certificateAuthorities.Options,
			}))
		})
	})
})
