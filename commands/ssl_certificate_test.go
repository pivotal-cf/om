package commands_test

import (
	"errors"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SslCertificate", func() {
	var (
		command                   *commands.SSLCertificate
		fakeSSLCertificateService *fakes.SSLCertificateService
		fakePresenter             *presenterfakes.FormattedPresenter
	)

	BeforeEach(func() {
		fakeSSLCertificateService = &fakes.SSLCertificateService{}
		fakePresenter = &presenterfakes.FormattedPresenter{}
		command = commands.NewSSLCertificate(fakeSSLCertificateService, fakePresenter)
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
			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeSSLCertificateService.GetSSLCertificateCallCount()).To(Equal(1))

			Expect(fakePresenter.PresentSSLCertificateCallCount()).To(Equal(1))
			Expect(fakePresenter.PresentSSLCertificateArgsForCall(0)).To(Equal(sslCertificateOutput))
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
				fakeSSLCertificateService.GetSSLCertificateReturns(
					api.SSLCertificateOutput{},
					errors.New("could not get custom certificate"),
				)

				err := executeCommand(command, []string{})
				Expect(err).To(MatchError("could not get custom certificate"))
			})
		})
	})
})
