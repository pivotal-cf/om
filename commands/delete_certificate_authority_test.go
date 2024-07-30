package commands_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("DeleteCertificateAuthority", func() {
	var (
		fakeService *fakes.DeleteCertificateAuthorityService
		fakeLogger  *fakes.Logger
		command     *commands.DeleteCertificateAuthority
	)

	BeforeEach(func() {
		fakeService = &fakes.DeleteCertificateAuthorityService{}
		fakeLogger = &fakes.Logger{}
		command = commands.NewDeleteCertificateAuthority(fakeService, fakeLogger)
	})

	Describe("Execute", func() {
		It("deletes the specified certificate authority", func() {
			err := executeCommand(command, []string{
				"--id", "some-certificate-authority-id",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.DeleteCertificateAuthorityCallCount()).To(Equal(1))
			Expect(fakeService.DeleteCertificateAuthorityArgsForCall(0)).To(Equal(api.DeleteCertificateAuthorityInput{
				GUID: "some-certificate-authority-id",
			}))

			Expect(fakeLogger.PrintfCallCount()).To(Equal(1))
			format, content := fakeLogger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("Certificate authority 'some-certificate-authority-id' deleted\n"))
		})

		When("using the --all-inactive flag", func() {
			args := []string{"--all-inactive"}

			Context("with one inactive CA", func() {
				BeforeEach(func() {
					fakeService.ListCertificateAuthoritiesReturns(api.CertificateAuthoritiesOutput{CAs: []api.CA{
						{
							GUID:   "active-ca-guid",
							Active: true,
						},
						{
							GUID:   "inactive-ca-guid",
							Active: false,
						},
						{
							GUID:   "another-inactive-ca-guid",
							Active: false,
						},
					}}, nil)
				})

				It("deletes the inactive certificate authorities", func() {
					err := executeCommand(command, args)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeService.ListCertificateAuthoritiesCallCount()).To(Equal(1))
					Expect(fakeService.DeleteCertificateAuthorityCallCount()).To(Equal(2))
					Expect(fakeService.DeleteCertificateAuthorityArgsForCall(0)).To(Equal(api.DeleteCertificateAuthorityInput{
						GUID: "inactive-ca-guid",
					}))
					Expect(fakeService.DeleteCertificateAuthorityArgsForCall(1)).To(Equal(api.DeleteCertificateAuthorityInput{
						GUID: "another-inactive-ca-guid",
					}))

					Expect(fakeLogger.PrintfCallCount()).To(Equal(2))
					format, content := fakeLogger.PrintfArgsForCall(0)
					Expect(fmt.Sprintf(format, content...)).To(Equal("Certificate authority 'inactive-ca-guid' deleted\n"))
					format, content = fakeLogger.PrintfArgsForCall(1)
					Expect(fmt.Sprintf(format, content...)).To(Equal("Certificate authority 'another-inactive-ca-guid' deleted\n"))
				})
			})

			Context("with no inactive CAs", func() {
				BeforeEach(func() {
					fakeService.ListCertificateAuthoritiesReturns(api.CertificateAuthoritiesOutput{CAs: []api.CA{
						{
							GUID:   "active-ca-guid",
							Active: true,
						},
					}}, nil)
				})

				It("does not delete anything", func() {
					err := executeCommand(command, args)
					Expect(err).To(HaveOccurred())

					Expect(fakeService.ListCertificateAuthoritiesCallCount()).To(Equal(1))
					Expect(fakeService.DeleteCertificateAuthorityCallCount()).To(Equal(0))
				})
			})
		})

		When("called with no args", func() {
			args := []string{}
			It("returns an error", func() {
				err := executeCommand(command, args)
				Expect(err).To(HaveOccurred())
			})
		})

		When("the service fails to delete a certificate", func() {
			It("returns an error", func() {
				fakeService.DeleteCertificateAuthorityReturns(errors.New("failed to delete certificate"))

				err := executeCommand(command, []string{
					"--id", "some-certificate-authority-id",
				})
				Expect(err).To(MatchError("failed to delete certificate"))
			})
		})
	})
})
