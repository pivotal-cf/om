package commands_test

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/renderers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bosh-env", func() {
	Context("Creating Command", func() {
		var (
			fakeService         *fakes.BoshEnvironmentService
			fakeRendererFactory *fakes.RendererFactory
			stdout              *fakes.Logger
		)
		BeforeEach(func() {
			fakeService = &fakes.BoshEnvironmentService{}
			fakeRendererFactory = &fakes.RendererFactory{}
			stdout = &fakes.Logger{}
		})
		It("Should use the target as is", func() {
			command := commands.NewBoshEnvironment(fakeService, stdout, "opsman.pivotal.io", fakeRendererFactory)
			Expect(command.Target()).Should(Equal("opsman.pivotal.io"))
		})

		It("Should remove protocol", func() {
			command := commands.NewBoshEnvironment(fakeService, stdout, "https://opsman.pivotal.io", fakeRendererFactory)
			Expect(command.Target()).Should(Equal("opsman.pivotal.io"))
		})

		It("Should remove protocol", func() {
			command := commands.NewBoshEnvironment(fakeService, stdout, "http://opsman.pivotal.io", fakeRendererFactory)
			Expect(command.Target()).Should(Equal("opsman.pivotal.io"))
		})

		It("should remove trailing slash", func() {
			command := commands.NewBoshEnvironment(fakeService, stdout, "opsman.pivotal.io/", fakeRendererFactory)
			Expect(command.Target()).Should(Equal("opsman.pivotal.io"))
		})

		It("should remove trailing slash and protocol", func() {
			command := commands.NewBoshEnvironment(fakeService, stdout, "https://opsman.pivotal.io/", fakeRendererFactory)
			Expect(command.Target()).Should(Equal("opsman.pivotal.io"))
		})
	})
	Context("calling the api", func() {
		var (
			command             commands.BoshEnvironment
			fakeService         *fakes.BoshEnvironmentService
			fakeRendererFactory *fakes.RendererFactory
			stdout              *fakes.Logger
		)

		BeforeEach(func() {
			fakeService = &fakes.BoshEnvironmentService{}
			fakeRendererFactory = &fakes.RendererFactory{}
			stdout = &fakes.Logger{}
			command = commands.NewBoshEnvironment(fakeService, stdout, "opsman.pivotal.io", fakeRendererFactory)
			fakeService.GetBoshEnvironmentReturns(api.GetBoshEnvironmentOutput{
				Client:       "opsmanager_client",
				ClientSecret: "my-super-secret",
				Environment:  "10.0.0.10",
			}, nil)
			fakeService.ListCertificateAuthoritiesReturns(api.CertificateAuthoritiesOutput{
				CAs: []api.CA{
					api.CA{
						Active:  true,
						CertPEM: "-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI....",
					},
				},
			}, nil)
			fakeRendererFactory.CreateReturns(renderers.NewPosix(), nil)
		})

		Describe("Execute without ssh key", func() {
			It("executes the API call", func() {
				err := command.Execute([]string{"-i", "somepath.pem"})

				Expect(err).ShouldNot(HaveOccurred())
				Expect(stdout.PrintlnCallCount()).To(Equal(10))
			})
		})

		Describe("Execute when multiple Active CAs", func() {
			It("executes the API call", func() {
				fakeService.ListCertificateAuthoritiesReturns(api.CertificateAuthoritiesOutput{
					CAs: []api.CA{
						api.CA{
							Active:  true,
							CertPEM: "-----BEGIN CERTIFICATE-----\ncert1....",
						},
						api.CA{
							Active:  true,
							CertPEM: "-----BEGIN CERTIFICATE-----\ncert2....",
						},
					},
				}, nil)
				err := command.Execute([]string{})

				Expect(err).ShouldNot(HaveOccurred())
				Expect(stdout.PrintlnCallCount()).To(Equal(8))
				for i := 0; i <= 7; i++ {
					value := fmt.Sprintf("%v", stdout.PrintlnArgsForCall(i))
					if strings.Contains(value, "BOSH_CA_CERT") {
						Expect(value).To(ContainSubstring("-----BEGIN CERTIFICATE-----\ncert1....\n-----BEGIN CERTIFICATE-----\ncert2...."))
					}
				}
			})
		})

		Describe("Execute without ssh key", func() {
			It("executes the API call", func() {
				err := command.Execute([]string{})

				Expect(err).ShouldNot(HaveOccurred())
				Expect(stdout.PrintlnCallCount()).To(Equal(8))
			})
		})
	})

	Describe("Usage", func() {
		It("returns the usage information for the bosh-env command", func() {
			command := commands.NewBoshEnvironment(nil, nil, "", nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This prints bosh environment variables to target bosh director. You can invoke it directly to see its output, or use it directly with an evaluate-type command:\nOn posix system: eval \"$(om bosh-env)\"\nOn powershell: iex $(om bosh-env | Out-String)",
				ShortDescription: "prints bosh environment variables",
				Flags:            command.Options,
			}))
		})
	})
})
