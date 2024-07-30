package acceptance

import (
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("generate certificate", func() {
	const output = `{
		"certificate": "-----BEGIN CERTIFICATE-----\nfake-cert\n-----END CERTIFICATE-----\n",
		"key": "-----BEGIN RSA PRIVATE KEY-----\nfake-key\n-----END RSA PRIVATE KEY-----\n"
	}`

	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v0/certificates/generate"),
				ghttp.RespondWith(http.StatusOK, `{
					"certificate": "-----BEGIN CERTIFICATE-----\nfake-cert\n-----END CERTIFICATE-----\n",
					"key": "-----BEGIN RSA PRIVATE KEY-----\nfake-key\n-----END RSA PRIVATE KEY-----\n"
				}`),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("generates a certificate authority on the OpsMan", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"generate-certificate",
			"--domains", "example.com,*.example.org",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(MatchJSON(output))
	})
})
