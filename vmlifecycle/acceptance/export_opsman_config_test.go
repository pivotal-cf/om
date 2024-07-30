package integration_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("ExportOpsmanConfig", func() {
	Context("requires the state-file flag", func() {
		It("throws an error", func() {
			command := exec.Command(pathToMain, "vm-lifecycle", "export-opsman-config")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 5).Should(gexec.Exit(1))

			Eventually(session.Err).Should(gbytes.Say("the required flag `--state-file' was not specified"))
		})
	})
})
