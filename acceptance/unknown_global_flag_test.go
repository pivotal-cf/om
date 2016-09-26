package acceptance

import (
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("unknown global flag", func() {
	It("prints the usage", func() {
		cmd := exec.Command(pathToMain, "-?")

		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Out).Should(gbytes.Say("flag provided but not defined: -?"))
	})
})
