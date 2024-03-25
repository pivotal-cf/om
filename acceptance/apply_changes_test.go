package acceptance

import (
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/ghttp"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("apply-changes command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	It("successfully applies the changes to the Ops Manager", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations"),
				ghttp.RespondWith(http.StatusOK, `{"install": {"id": 42}}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
				ghttp.RespondWith(http.StatusOK, `[{
					"guid": "guid1",
					"type": "product1"
				}, {
					"guid": "guid2",
					"type": "product2"
				}]`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
				ghttp.RespondWith(http.StatusOK, `[]`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v0/installations"),
				ghttp.VerifyJSON(`{
					"ignore_warnings": "false",
					"force_latest_variables": false,
					"deploy_products": "all"
				}`),
				ghttp.RespondWith(http.StatusOK, `{"install": {"id": 42}}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations/42"),
				ghttp.RespondWith(http.StatusOK, `{"status": "running"}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations/42/logs"),
				ghttp.RespondWith(http.StatusOK, `{ "logs": "call #0\n"}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations/42"),
				ghttp.RespondWith(http.StatusOK, `{"status": "succeeded"}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations/42/logs"),
				ghttp.RespondWith(http.StatusOK, `{ "logs": "call #0\ncall #1\n"}`),
			),
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"apply-changes")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, "5s").Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say("attempting to apply changes to the targeted Ops Manager"))
		Expect(session.Out).To(gbytes.Say("call #0"))
		Expect(session.Out).To(gbytes.Say("call #1"))
	})

	When("--recreate is enabled", func() {
		It("configures the director to recreate all VMs", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/installations"),
					ghttp.RespondWith(http.StatusOK, `{"install": {"id": 42}}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/info"),
					ghttp.RespondWith(http.StatusOK, `{
						"info": {
							"version": "2.1-build.79"
						}
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/director/properties"),
					ghttp.VerifyJSON(`{
						"director_configuration": {
							"bosh_recreate_on_next_deploy": true
						}
	  				}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[{
					"guid": "guid1",
					"type": "product1"
				}, {
					"guid": "guid2",
					"type": "product2"
				}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
					ghttp.RespondWith(http.StatusOK, `[]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/installations"),
					ghttp.VerifyJSON(`{
					"ignore_warnings": "false",
					"force_latest_variables": false,
					"deploy_products": "all"
				}`),
					ghttp.RespondWith(http.StatusOK, `{"install": {"id": 42}}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/installations/42"),
					ghttp.RespondWith(http.StatusOK, `{"status": "running"}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/installations/42/logs"),
					ghttp.RespondWith(http.StatusOK, `{ "logs": "call #0\n"}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/installations/42"),
					ghttp.RespondWith(http.StatusOK, `{"status": "succeeded"}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/installations/42/logs"),
					ghttp.RespondWith(http.StatusOK, `{ "logs": "call #0\ncall #1\n"}`),
				),
			)

			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"apply-changes",
				"--recreate-vms",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, "5s").Should(gexec.Exit(0))

			Expect(session.Out).To(gbytes.Say("attempting to apply changes to the targeted Ops Manager"))
			Expect(session.Out).To(gbytes.Say("call #0"))
			Expect(session.Out).To(gbytes.Say("call #1"))
		})
	})

	It("successfully re-attaches to an existing deployment", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations"),
				ghttp.RespondWith(http.StatusOK, `{
					"installations": [{
						"id": 42,
						"status": "running",
						"started_at": "2017-03-02T06:50:32.370Z"
					}]
				}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations/42"),
				ghttp.RespondWith(http.StatusOK, `{"status": "running"}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations/42/logs"),
				ghttp.RespondWith(http.StatusOK, `{ "logs": "call #0\n"}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations/42"),
				ghttp.RespondWith(http.StatusOK, `{"status": "succeeded"}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations/42/logs"),
				ghttp.RespondWith(http.StatusOK, `{ "logs": "call #0\ncall #1\n"}`),
			),
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--username", "some-running-install-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"apply-changes",
			"--reattach")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, "5s").Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say(`found already running installation... re-attaching \(Installation ID: 42, Started: Thu Mar  2 06:50:32 UTC 2017\)`))
		Expect(session.Out).To(gbytes.Say("call #0"))
		Expect(session.Out).To(gbytes.Say("call #1"))
	})
})
