package acceptance

import (
	"fmt"
	"github.com/onsi/gomega/gbytes"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const validConfigFile = `
---
password: some-env-provided-password
username: some-env-provided-username
target: %s
skip-ssl-validation: true
connect-timeout: 10
`

var _ = Describe("global env file", func() {
	When("provided --env flag", func() {
		var (
			configFile *os.File
			command    *exec.Cmd
		)

		createConfigFile := func(configContent string) {
			var err error

			configFile, err = ioutil.TempFile("", "config.yml")
			Expect(err).NotTo(HaveOccurred())

			_, err = configFile.WriteString(configContent)
			Expect(err).NotTo(HaveOccurred())

			err = configFile.Close()
			Expect(err).NotTo(HaveOccurred())
		}

		It("authenticates with creds in config file", func() {
			server := testServer(true)
			createConfigFile(fmt.Sprintf(validConfigFile, server.URL))
			command := exec.Command(pathToMain,
				"--env", configFile.Name(),
				"curl",
				"-p", "/api/v0/available_products",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(MatchJSON(`[ { "name": "p-bosh", "product_version": "999.99" } ]`))
		})

		It("errors if given an unexpected key", func() {
			server := testServer(false)
			configContent := fmt.Sprintf(`
---
password: some-env-provided-password
username: some-env-provided-username
target: %s
skip-ssl-validation: true
connect-timeout: 10
bad-key: bad-value
`, server.URL)
			createConfigFile(configContent)

			command := exec.Command(pathToMain,
				"--env", configFile.Name(),
				"curl",
				"-p", "/api/v0/available_products",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(string(session.Err.Contents())).To(ContainSubstring("could not parse env file: "))
			Expect(string(session.Err.Contents())).To(ContainSubstring("field bad-key not found"))
		})

		When("the env file contains variables", func() {
			BeforeEach(func() {
				createConfigFile(fmt.Sprintf(validConfigFile,"((target_url))"))
				command = exec.Command(pathToMain,
					"--env", configFile.Name(),
					"curl",
					"-p", "/api/v0/available_products",
				)
			})

			When("the OM_VARS_ENV environment variable IS set", func() {
				BeforeEach(func() {
					command.Env = append(command.Env, "OM_VARS_ENV=OM_VAR")
				})

				It("uses variables with the specified prefix in the env file", func() {
					server := testServer(true)
					command.Env = append(command.Env, fmt.Sprintf("OM_VAR_target_url=%s", server.URL))
					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(session).Should(gexec.Exit(0))

					Expect(string(session.Out.Contents())).To(MatchJSON(`[ { "name": "p-bosh", "product_version": "999.99" } ]`))

				})

				When("the env file contains variables not found in the environment", func() {
					It("exits 1 and lists the missing variables", func() {
						session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Expect(err).NotTo(HaveOccurred())

						Eventually(session).Should(gexec.Exit(1))

						Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("found problem in --env file:")))
						Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("env file contains YAML placeholders.")))
						Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("Pleases provide them via interpolation or environment variables.")))
						Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("* use OM_TARGET environment variable for the target value")))
						Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("Or, to enable interpolation of env.yml with variables from env-vars,")))
						Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("set the OM_VARS_ENV env var and put export the needed vars.")))
					})
				})
			})

			Context("and the OM_VARS_ENV environment variable IS NOT set", func() {
				It("exits 1, lists the needed vars, and notes the experimental OM_VARS_ENV feature", func() {
					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(session).Should(gexec.Exit(1))

					Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("found problem in --env file:")))
					Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("env file contains YAML placeholders.")))
					Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("Pleases provide them via interpolation or environment variables.")))
					Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("* use OM_TARGET environment variable for the target value")))
					Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("Or, to enable interpolation of env.yml with variables from env-vars,")))
					Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("set the OM_VARS_ENV env var and put export the needed vars.")))
				})
			})
		})

		When("given an invalid env file", func() {
			BeforeEach(func() {
				createConfigFile("invalid yaml")

				command = exec.Command(pathToMain,
					"--env", configFile.Name(),
					"curl",
					"-p", "/api/v0/available_products",
				)
			})

			It("returns an error", func() {
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(string(session.Err.Contents())).To(ContainSubstring("could not parse env file: "))
			})
		})

		When("given an env file that does not exist", func() {
			BeforeEach(func() {
				command = exec.Command(pathToMain,
					"--env", "does-not-exist.yml",
					"curl",
					"-p", "/api/v0/available_products",
				)
			})
			It("returns an error", func() {
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(string(session.Err.Contents())).To(ContainSubstring("env file does not exist: "))
			})
		})
	})
})
