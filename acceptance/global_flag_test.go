package acceptance

import (
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("global flags", func() {
	Context("when provided an unknown global flag", func() {
		It("prints the usage", func() {
			cmd := exec.Command(pathToMain, "-?")

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Out).Should(gbytes.Say("flag provided but not defined: -?"))
		})
	})

	Context("when not provided a target flag", func() {
		It("returns an error if the command is not help or version", func() {
			cmd := exec.Command(pathToMain, "-u", "foo", "-p", "bar", "apply-changes")

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Out).Should(gbytes.Say("error: target flag is required. Run `om help` for more info."))
		})

		It("does not return an error if the command is help or version", func() {
			cmd := exec.Command(pathToMain, "help")

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
		})
	})
})
