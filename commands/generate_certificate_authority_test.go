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
		var caResp api.GenerateCAResponse

		When("no warnings are returned", func() {
			It("makes a request to the Opsman to generate a certificate authority and prints to a table", func() {
				caResp = api.GenerateCAResponse{
					CA: api.CA{
						GUID:      "some GUID",
						Issuer:    "some Issuer",
						CreatedOn: "2017-09-12",
						ExpiresOn: "2018-09-12",
						Active:    true,
						CertPEM:   "some CertPem",
					},
				}

				fakeService.GenerateCertificateAuthorityReturns(caResp, nil)

				err := executeCommand(command, []string{})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeService.GenerateCertificateAuthorityCallCount()).To(Equal(1))

				Expect(fakePresenter.PresentCertificateAuthorityCallCount()).To(Equal(1))
				Expect(fakePresenter.PresentCertificateAuthorityArgsForCall(0)).To(Equal(caResp.CA))
			})
		})

		When("warnings are returned", func() {
			It("prints the response including warnings", func() {
				caResp = api.GenerateCAResponse{
					CA: api.CA{
						GUID:      "some GUID",
						Issuer:    "some Issuer",
						CreatedOn: "2017-09-12",
						ExpiresOn: "2018-09-12",
						Active:    true,
						CertPEM:   "some CertPem",
					},
					Warnings: []string{"something went wrong, but only kinda!"},
				}

				fakeService.GenerateCertificateAuthorityReturns(caResp, nil)

				err := executeCommand(command, []string{})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeService.GenerateCertificateAuthorityCallCount()).To(Equal(1))

				Expect(fakePresenter.PresentGenerateCAResponseCallCount()).To(Equal(1))
				Expect(fakePresenter.PresentGenerateCAResponseArgsForCall(0)).To(Equal(caResp))
			})
		})

		When("the format flag is provided", func() {
			It("sets the format on the presenter", func() {
				err := executeCommand(command, []string{
					"--format", "json",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
				Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("json"))
			})
		})

		It("returns an error when the service fails to generate a certificate", func() {
			fakeService.GenerateCertificateAuthorityReturns(api.GenerateCAResponse{}, errors.New("failed to generate certificate"))

			err := executeCommand(command, []string{})
			Expect(err).To(MatchError("failed to generate certificate"))
		})
	})
})
