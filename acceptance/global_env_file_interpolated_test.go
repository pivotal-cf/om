package acceptance

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("global env file", func() {
	When("provided config file with interpolatable variables and a vars file", func() {
		var (
			configContent    string
			varsContent      string
			configFile       *os.File
			varsFile         *os.File
			createConfigFile func(string)
			createVarsFile   func()
		)

		BeforeEach(func() {
			configContent = `
---
password: ((some-password))
username: some-env-provided-username
target: %s
skip-ssl-validation: true
connect-timeout: 10
`

			varsContent = `
---
some-password: some-env-provided-password
`

			createConfigFile = func(target string) {
				var err error

				configFile, err = ioutil.TempFile("", "config.yml")
				Expect(err).NotTo(HaveOccurred())

				_, err = configFile.WriteString(fmt.Sprintf(configContent, target))
				Expect(err).NotTo(HaveOccurred())

				err = configFile.Close()
				Expect(err).NotTo(HaveOccurred())
			}

			createVarsFile = func() {
				var err error

				varsFile, err = ioutil.TempFile("", "vars.yml")
				Expect(err).NotTo(HaveOccurred())

				_, err = varsFile.WriteString(varsContent)
				Expect(err).NotTo(HaveOccurred())

				err = varsFile.Close()
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("authenticates with creds in vars file", func() {
			server := testServer(true)

			createConfigFile(server.URL)
			createVarsFile()

			command := exec.Command(pathToMain,
				"--env", configFile.Name(),
				"--vars-file", varsFile.Name(),
				"curl",
				"-p", "/api/v0/available_products",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(MatchJSON(`[ { "name": "p-bosh", "product_version": "999.99" } ]`))
		})

		It("still errors if given an unexpected key", func() {
			server := testServer(true)

			configContent = `
---
password: ((some-password))
username: some-env-provided-username
target: %s
skip-ssl-validation: true
connect-timeout: 10
bad-key: bad-value
`

			varsContent = `
---
some-password: some-env-provided-password
`

			createConfigFile(server.URL)
			createVarsFile()

			command := exec.Command(pathToMain,
				"--env", configFile.Name(),
				"--vars-file", varsFile.Name(),
				"curl",
				"-p", "/api/v0/available_products",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(string(session.Err.Contents())).To(ContainSubstring("could not parse env file: "))
			Expect(string(session.Err.Contents())).To(ContainSubstring("field bad-key not found"))
		})

		When("given a vars file that does not exist", func() {
			It("returns an error", func() {
				command := exec.Command(pathToMain,
					"--env", "does-not-exist.yml",
					"curl",
					"-p", "/api/v0/available_products",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(string(session.Err.Contents())).To(ContainSubstring("env file does not exist: "))
			})
		})

		When("a parameter is missing", func() {
			It("", func() {
				server := testServer(true)

				createConfigFile(server.URL)

				command := exec.Command(pathToMain,
					"--env", configFile.Name(),
					"curl",
					"-p", "/api/v0/available_products",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(string(session.Err.Contents())).To(ContainSubstring(`Expected to find variables: some-password`))
			})
		})
	})

	When("provided config file with interpolatable variables and ENV variables", func() {
		var (
			configContent    string
			configFile       *os.File
			createConfigFile func(string)
		)

		BeforeEach(func() {
			configContent = `
---
password: ((some-password))
username: some-env-provided-username
target: %s
skip-ssl-validation: true
connect-timeout: 10
`

			createConfigFile = func(target string) {
				var err error

				configFile, err = ioutil.TempFile("", "config.yml")
				Expect(err).NotTo(HaveOccurred())

				_, err = configFile.WriteString(fmt.Sprintf(configContent, target))
				Expect(err).NotTo(HaveOccurred())

				err = configFile.Close()
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("authenticates with creds in config file", func() {
			server := testServer(true)

			createConfigFile(server.URL)

			command := exec.Command(pathToMain,
				"--env", configFile.Name(),
				"--vars-env", "PASSWORDS",
				"curl",
				"-p", "/api/v0/available_products",
			)

			command.Env = append(command.Env, "PASSWORDS_some-password=some-env-provided-password")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(MatchJSON(`[ { "name": "p-bosh", "product_version": "999.99" } ]`))
		})

		It("returns error if env value not found", func() {
			server := testServer(true)

			createConfigFile(server.URL)

			command := exec.Command(pathToMain,
				"--env", configFile.Name(),
				"--vars-env", "NOT_THERE",
				"curl",
				"-p", "/api/v0/available_products",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(string(session.Err.Contents())).To(ContainSubstring(`Expected to find variables: some-password`))
		})
	})

	When("both vars files and env variables are provided", func() {
		var (
			configContent    string
			varsContent      string
			configFile       *os.File
			varsFile         *os.File
			createConfigFile func(string)
			createVarsFile   func()
		)

		BeforeEach(func() {
			configContent = `
---
password: ((some-password))
username: ((some-username))
target: %s
skip-ssl-validation: true
connect-timeout: 10
`

			varsContent = `
---
some-password: some-env-provided-password
`

			createConfigFile = func(target string) {
				var err error

				configFile, err = ioutil.TempFile("", "config.yml")
				Expect(err).NotTo(HaveOccurred())

				_, err = configFile.WriteString(fmt.Sprintf(configContent, target))
				Expect(err).NotTo(HaveOccurred())

				err = configFile.Close()
				Expect(err).NotTo(HaveOccurred())
			}

			createVarsFile = func() {
				var err error

				varsFile, err = ioutil.TempFile("", "vars.yml")
				Expect(err).NotTo(HaveOccurred())

				_, err = varsFile.WriteString(varsContent)
				Expect(err).NotTo(HaveOccurred())

				err = varsFile.Close()
				Expect(err).NotTo(HaveOccurred())
			}
			It("interpolates both values", func() {
				It("authenticates with creds in config file", func() {
					server := testServer(true)

					createConfigFile(server.URL)
					createVarsFile()

					command := exec.Command(pathToMain,
						"--env", configFile.Name(),
						"--vars-file", varsFile.Name(),
						"--vars-env", "USERNAMES",
						"curl",
						"-p", "/api/v0/available_products",
					)

					command.Env = append(command.Env, "USERNAMES_some-username=some-env-provided-username")

					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(session).Should(gexec.Exit(0))
					Expect(string(session.Out.Contents())).To(MatchJSON(`[ { "name": "p-bosh", "product_version": "999.99" } ]`))
				})
			})
		})
	})
})
