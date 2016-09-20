package acceptance

import (
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const USAGE = `om cli helps you interact with an OpsManager

Usage: om [options] <command> [<args>]
  -v, --version  prints the om release version
  -h, --help     prints this usage information

Commands:
  version  prints the om release version
  help     prints this usage information
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
