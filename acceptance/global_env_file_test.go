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
	Context("when provided config file flag", func() {
		var configFile *os.File
		configContent := `
---
password: some-env-provided-password
username: some-env-provided-username
target: %s
skip-ssl-validation: true
connect-timeout: 10
`

		createConfigFile := func(target string) {
			var err error

			configFile, err = ioutil.TempFile("", "config.yml")
			Expect(err).NotTo(HaveOccurred())

			_, err = configFile.WriteString(fmt.Sprintf(configContent, target))
			Expect(err).NotTo(HaveOccurred())

			err = configFile.Close()
			Expect(err).NotTo(HaveOccurred())
		}

		It("authenticates with creds in config file", func() {
			server := testServer(true)

			createConfigFile(server.URL)

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

		Context("when given an invalid env file", func() {
			It("returns an error", func() {
				var err error

				configFile, err = ioutil.TempFile("", "config.yml")
				Expect(err).NotTo(HaveOccurred())

				_, err = configFile.WriteString(`this is invalid yaml`)
				Expect(err).NotTo(HaveOccurred())

				err = configFile.Close()
				Expect(err).NotTo(HaveOccurred())

				command := exec.Command(pathToMain,
					"--env", configFile.Name(),
					"curl",
					"-p", "/api/v0/available_products",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(string(session.Err.Contents())).To(ContainSubstring("could not parse env file: "))
			})
		})

		Context("when given an env file that does not exist", func() {
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
	})
})
