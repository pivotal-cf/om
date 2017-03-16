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

	Context("when the target url does not have a provided scheme", func() {
		It("assumes https", func() {
			cmd := exec.Command(pathToMain, "-t", "pcf.test-foo.cf-app.com", "apply-changes")

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Out).ShouldNot(gbytes.Say(`unsupported protocol scheme "" `))
			Expect(session.Out).Should(gbytes.Say(`Provided target has no scheme, assuming https`))
		})
	})

	Context("when not provided a target flag", func() {
		It("does not return an error if the command is help", func() {
			cmd := exec.Command(pathToMain, "help")

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
		})

		It("does not return an error if the command is version", func() {
			cmd := exec.Command(pathToMain, "version")

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
		})
	})
})
