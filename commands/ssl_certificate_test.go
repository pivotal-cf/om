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

var _ = Describe("SslCertificate", func() {
	var (
		sslCertificate            commands.SSLCertificate
		fakeSSLCertificateService *fakes.SSLCertificateService
		fakePresenter             *presenterfakes.FormattedPresenter
	)

	BeforeEach(func() {
		fakeSSLCertificateService = &fakes.SSLCertificateService{}
		fakePresenter = &presenterfakes.FormattedPresenter{}
		sslCertificate = commands.NewSSLCertificate(fakeSSLCertificateService, fakePresenter)
	})

	Describe("Execute", func() {
		var sslCertificateOutput api.SSLCertificate

		BeforeEach(func() {
			sslCertificateOutput = api.SSLCertificate{
				Certificate: "-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI....",
			}

			fakeSSLCertificateService.GetSSLCertificateReturns(
				api.SSLCertificateOutput{Certificate: sslCertificateOutput},
				nil,
			)
		})

		It("prints the certificate to a table", func() {
			err := sslCertificate.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeSSLCertificateService.GetSSLCertificateCallCount()).To(Equal(1))

			Expect(fakePresenter.PresentSSLCertificateCallCount()).To(Equal(1))
			Expect(fakePresenter.PresentSSLCertificateArgsForCall(0)).To(Equal(sslCertificateOutput))
		})

		When("the format flag is provided", func() {
			It("calls the presenter to set the json format", func() {
				err := sslCertificate.Execute([]string{
					"--format", "json",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
				Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("json"))
			})
		})

		When("the flag cannot parsed", func() {
			It("returns an error", func() {
				err := sslCertificate.Execute([]string{"--bogus", "nothing"})
				Expect(err).To(MatchError(
					"could not parse ssl-certificate flags: flag provided but not defined: -bogus",
				))
			})
		})

		When("request for certificate authorities fails", func() {
			It("returns an error", func() {
				fakeSSLCertificateService.GetSSLCertificateReturns(
					api.SSLCertificateOutput{},
					fmt.Errorf("could not get custom certificate"),
				)

				err := sslCertificate.Execute([]string{})
				Expect(err).To(MatchError("could not get custom certificate"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage", func() {
			usage := sslCertificate.Usage()

			Expect(usage).To(Equal(jhanda.Usage{
				Description:      "This authenticated command gets certificate applied to Ops Manager",
				ShortDescription: "gets certificate applied to Ops Manager",
				Flags:            usage.Flags,
			}))
		})
	})
})
