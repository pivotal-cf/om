package acceptance

import (
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const USAGE = `om cli helps you interact with an OpsManager

Usage: om [options] <command> [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the OpsManager VM
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)

Commands:
  configure-authentication  configures OpsManager with an internal userstore and admin user account
  help                      prints this usage information
  version                   prints the om release version
`

var _ = Describe("Help", func() {
	Context("when given the -h short flag", func() {
		It("prints the global usage", func() {
			command := exec.Command(pathToMain, "-h")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(USAGE))
		})
	})

	Context("when given the --help long flag", func() {
		It("prints the global usage", func() {
			command := exec.Command(pathToMain, "--help")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(USAGE))
		})
	})

	Context("when given the help command", func() {
		It("prints the global usage", func() {
			command := exec.Command(pathToMain, "help")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(USAGE))
		})
	})
})
