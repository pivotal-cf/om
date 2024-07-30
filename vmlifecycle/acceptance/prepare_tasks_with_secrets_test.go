package integration_test

import (
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PrepareTasksWithSecrets", func() {
	It("fails when tasksDir is not provided", func() {
		command := exec.Command(pathToMain, "vm-lifecycle", "prepare-tasks-with-secrets",
			"--config-dir",
			"some-dir",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 5).Should(gexec.Exit(1))
		Eventually(session.Err).Should(gbytes.Say("the required flag `--task-dir' was not specified"))
	})

	It("fails when configDir is not provided", func() {
		command := exec.Command(pathToMain, "vm-lifecycle", "prepare-tasks-with-secrets",
			"--task-dir",
			"some-dir",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 5).Should(gexec.Exit(1))
		Eventually(session.Err).Should(gbytes.Say("the required flag `--config-dir' was not specified"))
	})
})
