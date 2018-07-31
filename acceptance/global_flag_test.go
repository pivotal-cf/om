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

	It("overwrites the env vars", func() {
		server := testServer(true)

		command := exec.Command(pathToMain,
			"--username", "some-env-provided-username",
			"--password", "some-env-provided-password",
			"--skip-ssl-validation",
			"curl",
			"-p", "/api/v0/available_products",
		)
		command.Env = append(command.Env, "OM_TARGET="+server.URL)
		command.Env = append(command.Env, "PASSWORD=bogus")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(MatchJSON(`[ { "name": "p-bosh", "product_version": "999.99" } ]`))
	})
})
