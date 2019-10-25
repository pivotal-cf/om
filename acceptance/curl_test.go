package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("curl command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v0/some-api-endpoint"),
				ghttp.RespondWith(http.StatusOK, `{"some-key": "some-value"}`, map[string][]string{"Content-Type": {"application/json"}}),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("issues an API with credentials", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"curl",
			"--path", "/api/v0/some-api-endpoint",
			"--request", "POST",
			"--data", `{"some-key": "some-value"}`,
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(session.Err.Contents()).To(ContainSubstring("Status: 200 OK"))
		Expect(session.Err.Contents()).To(ContainSubstring("Content-Type: application/json"))
		Expect(string(session.Out.Contents())).To(MatchJSON(`{"some-key": "some-value"}`))
	})
})
