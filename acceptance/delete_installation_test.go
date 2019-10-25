package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete-installation command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.VerifyRequest("GET", "/api/v0/installations"),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("DELETE", "/api/v0/installation_asset_collection"),
				ghttp.RespondWith(http.StatusOK, `{"install": {"id": 42}}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations/42"),
				ghttp.RespondWith(http.StatusOK, `{ "status": "running" }`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations/42/logs"),
				ghttp.RespondWith(http.StatusOK, `{ "logs": "call #0\n"}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations/42"),
				ghttp.RespondWith(http.StatusOK, `{ "status": "succeeded" }`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations/42/logs"),
				ghttp.RespondWith(http.StatusOK, `{ "logs": "call #0\ncall #1"}`),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("successfully deletes the installation on the Ops Manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"delete-installation",
			"--force")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, "5s").Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say("attempting to delete the installation on the targeted Ops Manager"))
		Expect(session.Out).To(gbytes.Say("call #0"))
		Expect(session.Out).To(gbytes.Say("call #1"))
	})
})
