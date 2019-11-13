package acceptance

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"

	. "github.com/onsi/gomega"
)

var _ = Describe("revert-staged-changes command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
	})

	It("reverts the staged changes on the Ops Manager", func() {
		ensureHandler := &ensureHandler{}
		server.AppendHandlers(
			ensureHandler.Ensure(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/staged"),
					ghttp.RespondWith(http.StatusOK, ""),
				),
			)...,
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"revert-staged-changes",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(ensureHandler.Handlers()).To(HaveLen(0))
		Eventually(session.Out).Should(gbytes.Say("Changes Reverted."))
	})

	When("there are no changes to revert", func() {
		It("does nothing", func() {
			ensureHandler := &ensureHandler{}
			server.AppendHandlers(
				ensureHandler.Ensure(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/v0/staged"),
						ghttp.RespondWith(http.StatusNotModified, ""),
					),
				)...,
			)

			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"revert-staged-changes",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(ensureHandler.Handlers()).To(HaveLen(0))
			Eventually(session.Out).Should(gbytes.Say("No changes to revert."))
		})
	})

	When("the revert is forbidden", func() {
		It("errors", func() {
			ensureHandler := &ensureHandler{}
			server.AppendHandlers(
				ensureHandler.Ensure(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/v0/staged"),
						ghttp.RespondWith(http.StatusForbidden, ""),
					),
				)...,
			)

			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"revert-staged-changes",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(ensureHandler.Handlers()).To(HaveLen(0))
			Eventually(session.Err).Should(gbytes.Say("revert staged changes command failed: request failed: unexpected response from /api/v0/staged"))
		})
	})
})
