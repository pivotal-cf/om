package commands_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("ActivateCertificateAuthority", func() {
	var (
		fakeService *fakes.ActivateCertificateAuthorityService
		fakeLogger  *fakes.Logger
		command     *commands.ActivateCertificateAuthority
	)

	BeforeEach(func() {
		fakeService = &fakes.ActivateCertificateAuthorityService{}
		fakeLogger = &fakes.Logger{}
		command = commands.NewActivateCertificateAuthority(fakeService, fakeLogger)
	})

	Describe("Execute", func() {
		It("activates the specified certificate authority", func() {
			err := executeCommand(command, []string{
				"--id", "some-certificate-authority-id",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.ActivateCertificateAuthorityCallCount()).To(Equal(1))
			Expect(fakeService.ActivateCertificateAuthorityArgsForCall(0)).To(Equal(api.ActivateCertificateAuthorityInput{
				GUID: "some-certificate-authority-id",
			}))

			Expect(fakeLogger.PrintfCallCount()).To(Equal(1))
			format, content := fakeLogger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("Certificate authority 'some-certificate-authority-id' activated\n"))
		})

		Context("with no guid specified", func() {
			args := []string{}
			Context("with an inactive CA newer than the active CA", func() {
				BeforeEach(func() {
					fakeService.ListCertificateAuthoritiesReturns(api.CertificateAuthoritiesOutput{CAs: []api.CA{
						{
							GUID:      "active-ca-guid",
							Active:    true,
							CreatedOn: "2015-06-16T05:17:43Z",
						},
						{
							GUID:      "inactive-ca-guid",
							Active:    false,
							CreatedOn: "2025-06-16T05:17:44Z",
						},
					}}, nil)
				})

				It("activates the inactive CA", func() {
					err := executeCommand(command, args)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeService.ListCertificateAuthoritiesCallCount()).To(Equal(1), "list certificates call count")
					Expect(fakeService.ActivateCertificateAuthorityCallCount()).To(Equal(1), "activate certificate call count")
					Expect(fakeService.ActivateCertificateAuthorityArgsForCall(0)).To(Equal(api.ActivateCertificateAuthorityInput{
						GUID: "inactive-ca-guid",
					}), "activate ca API args")
				})
			})

			Context("with an inactive CA older than the active CA", func() {
				BeforeEach(func() {
					fakeService.ListCertificateAuthoritiesReturns(api.CertificateAuthoritiesOutput{CAs: []api.CA{
						{
							GUID:      "active-ca-guid",
							Active:    true,
							CreatedOn: "1995-06-16T05:17:43Z",
						},
						{
							GUID:      "inactive-ca-guid",
							Active:    false,
							CreatedOn: "1895-06-16T05:17:44Z",
						},
					}}, nil)
				})

				It("makes no activate call", func() {
					err := executeCommand(command, args)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeService.ListCertificateAuthoritiesCallCount()).To(Equal(1), "list certificates call count")
					Expect(fakeService.ActivateCertificateAuthorityCallCount()).To(Equal(0), "activate certificate call count")

					Expect(fakeLogger.PrintfCallCount()).To(Equal(1))
					format, content := fakeLogger.PrintfArgsForCall(0)
					Expect(fmt.Sprintf(format, content...)).To(Equal("No newer certificate authority available to activate\n"))
				})
			})

			Context("with multiple inactive CAs that are newer than the active CA", func() {
				BeforeEach(func() {
					fakeService.ListCertificateAuthoritiesReturns(api.CertificateAuthoritiesOutput{CAs: []api.CA{
						{
							GUID:      "active-ca-guid",
							Active:    true,
							CreatedOn: "1950-06-16T05:17:43Z",
						},
						{
							GUID:      "inactive-ca-guid-2",
							Active:    false,
							CreatedOn: "1952-06-16T09:17:44Z",
						},
						{
							GUID:      "inactive-ca-guid-1",
							Active:    false,
							CreatedOn: "1951-06-16T05:17:44Z",
						},
					}}, nil)
				})

				It("makes activate call with the newest inactive CA", func() {
					err := executeCommand(command, args)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeService.ListCertificateAuthoritiesCallCount()).To(Equal(1), "list certificates call count")
					Expect(fakeService.ActivateCertificateAuthorityCallCount()).To(Equal(1), "activate certificate call count")
					Expect(fakeService.ActivateCertificateAuthorityArgsForCall(0)).To(Equal(api.ActivateCertificateAuthorityInput{
						GUID: "inactive-ca-guid-2",
					}), "activate ca API args")

				})
			})

			Context("with multiple inactive CAs that are older than the active CA", func() {
				BeforeEach(func() {
					fakeService.ListCertificateAuthoritiesReturns(api.CertificateAuthoritiesOutput{CAs: []api.CA{
						{
							GUID:      "active-ca-guid",
							Active:    true,
							CreatedOn: "1965-06-16T05:17:43Z",
						},
						{
							GUID:      "inactive-ca-guid-2",
							Active:    false,
							CreatedOn: "1964-06-16T09:17:44Z",
						},
						{
							GUID:      "inactive-ca-guid-1",
							Active:    false,
							CreatedOn: "1962-06-16T05:17:44Z",
						},
					}}, nil)
				})

				It("makes no activate call", func() {
					err := executeCommand(command, args)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeService.ListCertificateAuthoritiesCallCount()).To(Equal(1), "list certificates call count")
					Expect(fakeService.ActivateCertificateAuthorityCallCount()).To(Equal(0), "activate certificate call count")

					Expect(fakeLogger.PrintfCallCount()).To(Equal(1))
					format, content := fakeLogger.PrintfArgsForCall(0)
					Expect(fmt.Sprintf(format, content...)).To(Equal("No newer certificate authority available to activate\n"))
				})
			})

			Context("with multiple inactive CAs one older and one newer than the active CA", func() {
				BeforeEach(func() {
					fakeService.ListCertificateAuthoritiesReturns(api.CertificateAuthoritiesOutput{CAs: []api.CA{
						{
							GUID:      "active-ca-guid",
							Active:    true,
							CreatedOn: "1974-06-16T05:17:43Z",
						},
						{
							GUID:      "inactive-ca-guid-2",
							Active:    false,
							CreatedOn: "1975-06-16T09:17:44Z",
						},
						{
							GUID:      "inactive-ca-guid-1",
							Active:    false,
							CreatedOn: "1972-06-16T05:17:44Z",
						},
					}}, nil)
				})

				It("makes activate call with the newest inactive CA", func() {
					err := executeCommand(command, args)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeService.ListCertificateAuthoritiesCallCount()).To(Equal(1), "list certificates call count")
					Expect(fakeService.ActivateCertificateAuthorityCallCount()).To(Equal(1), "activate certificate call count")
					Expect(fakeService.ActivateCertificateAuthorityArgsForCall(0)).To(Equal(api.ActivateCertificateAuthorityInput{
						GUID: "inactive-ca-guid-2",
					}), "activate ca API args")

				})
			})

			Context("with no inactive CA", func() {
				BeforeEach(func() {
					fakeService.ListCertificateAuthoritiesReturns(api.CertificateAuthoritiesOutput{CAs: []api.CA{
						{
							GUID:      "active-ca-guid",
							Active:    true,
							CreatedOn: "2023-01-31T12:00:00Z",
						},
					}}, nil)
				})

				It("returns an error", func() {
					err := executeCommand(command, args)
					Expect(err).To(MatchError("no inactive certificate authorities to activate"))

					Expect(fakeService.ListCertificateAuthoritiesCallCount()).To(Equal(1), "list certificates call count")
					Expect(fakeService.ActivateCertificateAuthorityCallCount()).To(Equal(0), "activate certificate call count")
				})
			})
		})

		When("the service fails to activate a certificate", func() {
			It("returns an error", func() {
				fakeService.ActivateCertificateAuthorityReturns(errors.New("failed to activate certificate"))

				err := executeCommand(command, []string{
					"--id", "some-certificate-authority-id",
				})
				Expect(err).To(MatchError("failed to activate certificate"))
			})
		})
	})
})
