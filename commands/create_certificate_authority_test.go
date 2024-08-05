package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"
)

var _ = Describe("CreateCertificateAuthority", func() {
	var (
		fakePresenter *presenterfakes.FormattedPresenter
		fakeService   *fakes.CreateCertificateAuthorityService
		command       *commands.CreateCertificateAuthority
	)

	BeforeEach(func() {
		fakePresenter = &presenterfakes.FormattedPresenter{}
		fakeService = &fakes.CreateCertificateAuthorityService{}
		command = commands.NewCreateCertificateAuthority(fakeService, fakePresenter)
	})

	Describe("Execute", func() {
		It("makes a request to the Opsman to create a certificate authority", func() {
			err := executeCommand(command, []string{
				"--certificate-pem", "some CertPem",
				"--private-key-pem", "some PrivateKey",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.CreateCertificateAuthorityCallCount()).To(Equal(1))
			Expect(fakeService.CreateCertificateAuthorityArgsForCall(0)).To(Equal(api.CertificateAuthorityInput{
				CertPem:       "some CertPem",
				PrivateKeyPem: "some PrivateKey",
			}))
		})

		It("prints a table containing the certificate authority that was created", func() {
			caResp := api.GenerateCAResponse{
				CA: api.CA{
					GUID:      "some GUID",
					Issuer:    "some Issuer",
					CreatedOn: "2017-09-12",
					ExpiresOn: "2018-09-12",
					Active:    true,
					CertPEM:   "some CertPem",
				},
			}

			fakeService.CreateCertificateAuthorityReturns(caResp, nil)

			err := executeCommand(command, []string{
				"--certificate-pem", "some CertPem",
				"--private-key-pem", "some PrivateKey",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakePresenter.PresentCertificateAuthorityCallCount()).To(Equal(1))
			Expect(fakePresenter.PresentCertificateAuthorityArgsForCall(0)).To(Equal(caResp.CA))
		})

		When("warnings are returned", func() {
			It("prints the response including warnings", func() {
				caResp := api.GenerateCAResponse{
					CA: api.CA{
						GUID:      "some GUID",
						Issuer:    "some Issuer",
						CreatedOn: "2017-09-12",
						ExpiresOn: "2018-09-12",
						Active:    false,
						CertPEM:   "some CertPem",
					},
					Warnings: []string{"something went wrong, but only kinda!"},
				}

				fakeService.CreateCertificateAuthorityReturns(caResp, nil)

				err := executeCommand(command, []string{
					"--certificate-pem", "some CertPem",
					"--private-key-pem", "some PrivateKey",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeService.CreateCertificateAuthorityCallCount()).To(Equal(1))

				Expect(fakePresenter.PresentGenerateCAResponseCallCount()).To(Equal(1))
				Expect(fakePresenter.PresentGenerateCAResponseArgsForCall(0)).To(Equal(caResp))
			})
		})

		When("the format flag is provided", func() {
			It("sets the format on the presenter", func() {
				err := executeCommand(command, []string{
					"--format", "json",
					"--certificate-pem", "some CertPem",
					"--private-key-pem", "some PrivateKey",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
				Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("json"))
			})
		})

		When("the service fails to create a certificate", func() {
			It("returns an error", func() {
				fakeService.CreateCertificateAuthorityReturns(api.GenerateCAResponse{}, errors.New("failed to create certificate"))

				err := executeCommand(command, []string{
					"--certificate-pem", "some CertPem",
					"--private-key-pem", "some PrivateKey",
				})
				Expect(err).To(MatchError("failed to create certificate"))
			})
		})
	})
})
