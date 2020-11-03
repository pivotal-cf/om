package commands_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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
			command             *commands.BoshEnvironment
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

		Describe("Execute with a nonexistent ssh key", func() {
			It("executes the API call", func() {
				err := executeCommand(command,[]string{"-i", "somepath.pem"})
				Expect(err).To(MatchError(ContainSubstring("ssh key file 'somepath.pem' does not exist")))
			})
		})

		Describe("Execute with a real ssh key", func() {
			var keyFile string
			var f *os.File
			var err error

			BeforeEach(func() {
				err = os.Mkdir("./tmp", os.ModePerm)
				Expect(err).ToNot(HaveOccurred())
				f, err = ioutil.TempFile("./tmp", "opsmankey-*.pem")
				Expect(err).ToNot(HaveOccurred())

				keyFile, err = filepath.Abs(f.Name())
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				os.RemoveAll("./tmp")
			})

			It("Resolves to the absolute path", func() {
				wd, err := os.Getwd()
				Expect(err).ToNot(HaveOccurred())
				defer func() {
					err = os.Chdir(wd)
					Expect(err).ToNot(HaveOccurred())
				}()

				err = os.Chdir("./tmp")
				Expect(err).ToNot(HaveOccurred())

				err = executeCommand(command,[]string{"-i", filepath.Base(keyFile)})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(stdout.PrintlnCallCount()).To(Equal(10))
				for i := 0; i < 10; i++ {
					value := fmt.Sprintf("%v", stdout.PrintlnArgsForCall(i))
					if strings.Contains(value, "BOSH_ALL_PROXY") {
						Expect(value).To(Equal(fmt.Sprintf("[export BOSH_ALL_PROXY=ssh+socks5://ubuntu@opsman.pivotal.io:22?private-key=%s]", keyFile)))
					}
				}
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
				err := executeCommand(command,[]string{})

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
				err := executeCommand(command,[]string{})

				Expect(err).ShouldNot(HaveOccurred())
				Expect(stdout.PrintlnCallCount()).To(Equal(8))
			})
		})
	})

	Context("printing environment variables", func() {
		var (
			command             *commands.BoshEnvironment
			fakeService         *fakes.BoshEnvironmentService
			fakeRendererFactory *fakes.RendererFactory
			stdout              *fakes.Logger
			err                 error
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

		It("prints all of the environment variables when neither the bosh or credhub flags are passed", func() {
			err = os.Mkdir("./tmp-all-env", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			defer func() {
				err = os.RemoveAll("./tmp-all-env")
				Expect(err).ToNot(HaveOccurred())
			}()
			f, err := ioutil.TempFile("./tmp-all-env", "opsmankey-*.pem")
			Expect(err).ToNot(HaveOccurred())

			keyFile, err := filepath.Abs(f.Name())
			Expect(err).ToNot(HaveOccurred())

			wd, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				err = os.Chdir(wd)
				Expect(err).ToNot(HaveOccurred())
			}()

			err = os.Chdir("./tmp-all-env")
			Expect(err).ToNot(HaveOccurred())

			err = executeCommand(command,[]string{"-i", filepath.Base(keyFile)})
			Expect(err).ToNot(HaveOccurred())

			var lines []string
			for _, outer := range stdout.Invocations()["Println"] {
				for _, middle := range outer {
					for _, line := range middle.([]interface{}) {
						lines = append(lines, fmt.Sprintf("%v", line))
					}
				}
			}

			Expect(lines).To(ContainElements(
				"export CREDHUB_SERVER=https://10.0.0.10:8844",
				"export CREDHUB_CA_CERT='-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI....\n'",
				fmt.Sprintf("export CREDHUB_PROXY=ssh+socks5://ubuntu@opsman.pivotal.io:22?private-key=%s", keyFile),
				"export BOSH_CLIENT=opsmanager_client",
				"export BOSH_CLIENT_SECRET=my-super-secret",
				"export BOSH_CA_CERT='-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI....\n'",
				fmt.Sprintf("export BOSH_ALL_PROXY=ssh+socks5://ubuntu@opsman.pivotal.io:22?private-key=%s", keyFile),
				"export BOSH_ENVIRONMENT=10.0.0.10",
				"export CREDHUB_CLIENT=opsmanager_client",
				"export CREDHUB_SECRET=my-super-secret",
			))
		})

		It("prints only BOSH environment variables when the bosh flag is passed", func() {
			err = os.Mkdir("./tmp-bosh-env", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			defer func() {
				err = os.RemoveAll("./tmp-bosh-env")
				Expect(err).ToNot(HaveOccurred())
			}()
			f, err := ioutil.TempFile("./tmp-bosh-env", "opsmankey-*.pem")
			Expect(err).ToNot(HaveOccurred())

			keyFile, err := filepath.Abs(f.Name())
			Expect(err).ToNot(HaveOccurred())

			wd, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				err = os.Chdir(wd)
				Expect(err).ToNot(HaveOccurred())
			}()

			err = os.Chdir("./tmp-bosh-env")
			Expect(err).ToNot(HaveOccurred())

			err = executeCommand(command,[]string{"-i", filepath.Base(keyFile), "-b"})
			Expect(err).ToNot(HaveOccurred())

			var lines []string
			for _, outer := range stdout.Invocations()["Println"] {
				for _, middle := range outer {
					for _, line := range middle.([]interface{}) {
						lines = append(lines, fmt.Sprintf("%v", line))
					}
				}
			}

			Expect(lines).To(ContainElements(
				"export BOSH_CLIENT=opsmanager_client",
				"export BOSH_CLIENT_SECRET=my-super-secret",
				"export BOSH_CA_CERT='-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI....\n'",
				fmt.Sprintf("export BOSH_ALL_PROXY=ssh+socks5://ubuntu@opsman.pivotal.io:22?private-key=%s", keyFile),
				"export BOSH_ENVIRONMENT=10.0.0.10",
			))
		})

		It("prints only the Credhub environment variables when the credhub flag is passed", func() {
			err = os.Mkdir("./tmp-credhub-env", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			defer func() {
				err = os.RemoveAll("./tmp-credhub-env")
				Expect(err).ToNot(HaveOccurred())
			}()
			f, err := ioutil.TempFile("./tmp-credhub-env", "opsmankey-*.pem")
			Expect(err).ToNot(HaveOccurred())

			keyFile, err := filepath.Abs(f.Name())
			Expect(err).ToNot(HaveOccurred())

			wd, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				err = os.Chdir(wd)
				Expect(err).ToNot(HaveOccurred())
			}()

			err = os.Chdir("./tmp-credhub-env")
			Expect(err).ToNot(HaveOccurred())

			err = executeCommand(command,[]string{"-i", filepath.Base(keyFile), "-c"})
			Expect(err).ToNot(HaveOccurred())

			var lines []string
			for _, outer := range stdout.Invocations()["Println"] {
				for _, middle := range outer {
					for _, line := range middle.([]interface{}) {
						lines = append(lines, fmt.Sprintf("%v", line))
					}
				}
			}

			Expect(lines).To(ContainElements(
				"export CREDHUB_SERVER=https://10.0.0.10:8844",
				"export CREDHUB_CA_CERT='-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI....\n'",
				fmt.Sprintf("export CREDHUB_PROXY=ssh+socks5://ubuntu@opsman.pivotal.io:22?private-key=%s", keyFile),
				"export CREDHUB_CLIENT=opsmanager_client",
				"export CREDHUB_SECRET=my-super-secret",
			))
		})
	})

	Context("printing unset commands for the environment variables", func() {
		var (
			command             *commands.BoshEnvironment
			fakeService         *fakes.BoshEnvironmentService
			fakeRendererFactory *fakes.RendererFactory
			stdout              *fakes.Logger
			err                 error
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

		It("prints all of the unset commands when neither the bosh or credhub flags are passed", func() {
			err = executeCommand(command,[]string{"--unset"})
			Expect(err).ToNot(HaveOccurred())

			var lines []string
			for _, outer := range stdout.Invocations()["Println"] {
				for _, middle := range outer {
					for _, line := range middle.([]interface{}) {
						lines = append(lines, fmt.Sprintf("%v", line))
					}
				}
			}

			Expect(lines).To(ContainElements(
				"unset CREDHUB_SERVER",
				"unset CREDHUB_CA_CERT",
				"unset CREDHUB_PROXY",
				"unset BOSH_CLIENT",
				"unset BOSH_CLIENT_SECRET",
				"unset BOSH_CA_CERT",
				"unset BOSH_ALL_PROXY",
				"unset BOSH_ENVIRONMENT",
				"unset CREDHUB_CLIENT",
				"unset CREDHUB_SECRET",
			))
		})

		It("prints only unset commands for BOSH environment variables when the bosh flag is passed", func() {
			err = executeCommand(command,[]string{"--unset", "-b"})
			Expect(err).ToNot(HaveOccurred())

			var lines []string
			for _, outer := range stdout.Invocations()["Println"] {
				for _, middle := range outer {
					for _, line := range middle.([]interface{}) {
						lines = append(lines, fmt.Sprintf("%v", line))
					}
				}
			}

			Expect(lines).To(ContainElements(
				"unset BOSH_CLIENT",
				"unset BOSH_CLIENT_SECRET",
				"unset BOSH_CA_CERT",
				"unset BOSH_ALL_PROXY",
				"unset BOSH_ENVIRONMENT",
			))
		})

		It("prints only inset commands for the Credhub environment variables when the credhub flag is passed", func() {
			err = executeCommand(command,[]string{"--unset", "-c"})
			Expect(err).ToNot(HaveOccurred())

			var lines []string
			for _, outer := range stdout.Invocations()["Println"] {
				for _, middle := range outer {
					for _, line := range middle.([]interface{}) {
						lines = append(lines, fmt.Sprintf("%v", line))
					}
				}
			}

			Expect(lines).To(ContainElements(
				"unset CREDHUB_SERVER",
				"unset CREDHUB_CA_CERT",
				"unset CREDHUB_PROXY",
				"unset CREDHUB_CLIENT",
				"unset CREDHUB_SECRET",
			))
		})
	})
})
