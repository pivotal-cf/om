package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("GenerateCertificateAuthority", func() {
	var (
		fakeTableWriter                 *fakes.TableWriter
		fakeCertificateAuthorityService *fakes.CertificateAuthoritiesService
		command                         commands.GenerateCertificateAuthority
	)

	BeforeEach(func() {
		fakeTableWriter = &fakes.TableWriter{}
		fakeCertificateAuthorityService = &fakes.CertificateAuthoritiesService{}
		command = commands.NewGenerateCertificateAuthority(fakeCertificateAuthorityService, fakeTableWriter)
	})

	Describe("Execute", func() {
		It("makes a request to the Opsman to generate a certificate authority", func() {
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeCertificateAuthorityService.GenerateCallCount()).To(Equal(1))
		})

		It("prints a table containing the certificate authority that was generated", func() {
			fakeCertificateAuthorityService.GenerateReturns(api.CA{GUID: "some GUID", Issuer: "some Issuer",
				CreatedOn: "2017-09-12", ExpiresOn: "2018-09-12", Active: true, CertPEM: "some CertPem"}, nil)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeTableWriter.SetAutoWrapTextCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetAutoWrapTextArgsForCall(0)).To(BeFalse())

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"id", "issuer", "active", "created on", "expires on", "certicate pem"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(1))
			Expect(fakeTableWriter.AppendArgsForCall(0)).To(Equal([]string{"some GUID", "some Issuer",
				"true", "2017-09-12", "2018-09-12", "some CertPem"}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))

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
			Expect(usage).To(Equal(commands.Usage{
				Description:      "This authenticated command generates a certificate authority on the Ops Manager",
				ShortDescription: "generates a certificate authority on the Opsman",
			}))
		})
	})
})
