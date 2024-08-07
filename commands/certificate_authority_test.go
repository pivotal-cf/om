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

var _ = Describe("Certificate Authority", func() {
	var (
		command                           *commands.CertificateAuthority
		fakeCertificateAuthoritiesService *fakes.CertificateAuthoritiesService
		fakePresenter                     *presenterfakes.FormattedPresenter
		fakeLogger                        *fakes.Logger
	)

	BeforeEach(func() {
		fakeCertificateAuthoritiesService = &fakes.CertificateAuthoritiesService{}
		fakePresenter = &presenterfakes.FormattedPresenter{}
		fakeLogger = &fakes.Logger{}
		command = commands.NewCertificateAuthority(fakeCertificateAuthoritiesService, fakePresenter, fakeLogger)

		certificateAuthorities := []api.CA{
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
			api.CertificateAuthoritiesOutput{CAs: certificateAuthorities},
			nil,
		)
	})

	Describe("Execute", func() {
		It("requests CAs from the server and prints to a table", func() {
			err := executeCommand(command, []string{
				"--id", "other-guid",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeCertificateAuthoritiesService.ListCertificateAuthoritiesCallCount()).To(Equal(1))

			Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
			Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("table"))
			Expect(fakePresenter.PresentCertificateAuthorityCallCount()).To(Equal(1))
			Expect(fakePresenter.PresentCertificateAuthorityArgsForCall(0)).To(Equal(api.CA{
				GUID:      "other-guid",
				Issuer:    "Customer",
				CreatedOn: "2017-01-10",
				ExpiresOn: "2021-01-10",
				Active:    false,
				CertPEM:   "-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBBhI....",
			}))
		})

		When("the cert-pem flag is provided", func() {
			It("logs the cert pem to the stdout", func() {
				err := executeCommand(command, []string{
					"--id", "other-guid",
					"--cert-pem",
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(fakePresenter.PresentCertificateAuthorityCallCount()).To(Equal(0))
				Expect(fakeLogger.PrintlnCallCount()).To(Equal(1))
				output := fakeLogger.PrintlnArgsForCall(0)
				Expect(output).To(ConsistOf("-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBBhI...."))
			})
		})

		When("the format flag is provided", func() {
			It("calls the presenter to set the json format", func() {
				err := executeCommand(command, []string{
					"--id", "other-guid",
					"--format", "json",
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
				Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("json"))
				Expect(fakePresenter.PresentCertificateAuthorityCallCount()).To(Equal(1))
			})
		})

		When("there is only one CA and the id flag is not present", func() {
			It("requests CAs from the server and prints to a table", func() {
				certificateAuthorities := []api.CA{
					{
						GUID:      "some-guid",
						Issuer:    "Pivotal",
						CreatedOn: "2017-01-09",
						ExpiresOn: "2021-01-09",
						Active:    true,
						CertPEM:   "-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI....",
					},
				}

				fakeCertificateAuthoritiesService.ListCertificateAuthoritiesReturns(
					api.CertificateAuthoritiesOutput{CAs: certificateAuthorities},
					nil,
				)
				err := executeCommand(command, []string{})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeCertificateAuthoritiesService.ListCertificateAuthoritiesCallCount()).To(Equal(1))

				Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
				Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("table"))
				Expect(fakePresenter.PresentCertificateAuthorityCallCount()).To(Equal(1))
				Expect(fakePresenter.PresentCertificateAuthorityArgsForCall(0)).To(Equal(api.CA{
					GUID:      "some-guid",
					Issuer:    "Pivotal",
					CreatedOn: "2017-01-09",
					ExpiresOn: "2021-01-09",
					Active:    true,
					CertPEM:   "-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI....",
				}))
			})
		})

		When("the service fails to retrieve CAs", func() {
			BeforeEach(func() {
				fakeCertificateAuthoritiesService.ListCertificateAuthoritiesReturns(
					api.CertificateAuthoritiesOutput{},
					errors.New("service failed"),
				)
			})

			It("returns an error", func() {
				err := executeCommand(command, []string{
					"--id", "some-guid",
				})
				Expect(err).To(MatchError("service failed"))
			})
		})

		When("the --id flag is missing", func() {
			It("returns an error", func() {
				err := executeCommand(command, []string{})
				Expect(err).To(MatchError("More than one certificate authority found. Please use --id flag to specify. IDs can be found using the certificate-authorities command"))
			})
		})

		When("the request certificate authority is not found", func() {
			It("returns an error", func() {
				err := executeCommand(command, []string{
					"--id", "doesnt-exist",
				})
				Expect(err).To(MatchError(`could not find a certificate authority with ID: "doesnt-exist"`))
			})
		})
	})
})
