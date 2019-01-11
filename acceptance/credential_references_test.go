package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("credential references command", func() {
	var (
		server *httptest.Server
	)

	const tableOutput = `+------------------------------+
|         CREDENTIALS          |
+------------------------------+
| .my-job.some-credentials     |
| .properties.some-credentials |
+------------------------------+
`

	const jsonOutput = `[
    ".my-job.some-credentials",
    ".properties.some-credentials"
	]`

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				_, err := w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/deployed/products":
				_, err := w.Write([]byte(`[
					{"installation_name":"p-bosh","guid":"p-bosh-guid","type":"p-bosh","product_version":"1.10.0.0"},
					{"installation_name":"cf","guid":"cf-guid","type":"cf","product_version":"1.10.0-build.177"},
					{"installation_name":"some-product","guid":"some-product-guid","type":"some-product","product_version":"1.0.0"},
					{"installation_name":"p-isolation-segment","guid":"p-isolation-segment-guid","type":"p-isolation-segment","product_version":"1.10.0-build.31"}
				]`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/deployed/products/some-product-guid/credentials":
				_, err := w.Write([]byte(`{
					"credentials": [
						".properties.some-credentials",
						".my-job.some-credentials"
					]
				}`))
				Expect(err).ToNot(HaveOccurred())
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	AfterEach(func() {
		server.Close()
	})

	It("lists the credential references belonging to the deployed product", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"credential-references",
			"--product-name", "some-product")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	Context("when json formatting is requested", func() {
		It("lists the credential references belonging to the deployed product in json", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"credential-references",
				"--format", "json",
				"--product-name", "some-product")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})
})
