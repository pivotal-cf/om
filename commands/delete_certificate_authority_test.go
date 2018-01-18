package commands_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("DeleteCertificateAuthority", func() {
	var (
		fakeCertificateAuthorityService *fakes.CertificateAuthorityDeleter
		fakeLogger                      *fakes.Logger
		command                         commands.DeleteCertificateAuthority
	)

	BeforeEach(func() {
		fakeCertificateAuthorityService = &fakes.CertificateAuthorityDeleter{}
		fakeLogger = &fakes.Logger{}
		command = commands.NewDeleteCertificateAuthority(fakeCertificateAuthorityService, fakeLogger)
	})

	Describe("Execute", func() {
		It("deletes the specified certificate authority", func() {
			err := command.Execute([]string{
				"--id", "some-certificate-authority-id",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeCertificateAuthorityService.DeleteCallCount()).To(Equal(1))
			Expect(fakeCertificateAuthorityService.DeleteArgsForCall(0)).To(Equal(api.DeleteCertificateAuthorityInput{
				GUID: "some-certificate-authority-id",
			}))

			Expect(fakeLogger.PrintfCallCount()).To(Equal(1))
			format, content := fakeLogger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("Certificate authority 'some-certificate-authority-id' deleted\n"))
		})

		Context("failure cases", func() {
			Context("when the service fails to delete a certificate", func() {
				It("returns an error", func() {
					fakeCertificateAuthorityService.DeleteReturns(errors.New("failed to delete certificate"))

					err := command.Execute([]string{
						"--id", "some-certificate-authority-id",
					})
					Expect(err).To(MatchError("failed to delete certificate"))
				})
			})
			Context("when an unknown flag is provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--badflag"})
					Expect(err).To(MatchError("could not parse delete-certificate-authority flags: flag provided but not defined: -badflag"))
				})
			})
			Context("when the id flag is not provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{})
					Expect(err).To(MatchError("could not parse delete-certificate-authority flags: missing required flag \"--id\""))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage info", func() {
			usage := command.Usage()
			Expect(usage).To(Equal(jhanda.Usage{
				Description:      "This authenticated command deletes an existing certificate authority on the Ops Manager",
				ShortDescription: "deletes a certificate authority on the Ops Manager",
				Flags:            command.Options,
			}))
		})
	})
})
