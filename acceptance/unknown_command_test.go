package acceptance

import (
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("unknown command", func() {
	It("prints the usage", func() {
		cmd := exec.Command(pathToMain, "-t", "pcf.foo.cf-app.com", "banana")

		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err).Should(gbytes.Say("Unknown command"))
	})
})
