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

var _ = Describe("errands command", func() {
	var (
		server *httptest.Server
	)

	const tableOutput = `+---------------+---------------------+--------------------+
|     NAME      | POST DEPLOY ENABLED | PRE DELETE ENABLED |
+---------------+---------------------+--------------------+
| some-errand-1 | true                | true               |
| some-errand-2 | false               | false              |
| some-errand-3 |                     | true               |
| some-errand-4 | when-changed        |                    |
+---------------+---------------------+--------------------+
`

	const jsonOutput = `[
		{"name": "some-errand-1", "post_deploy_enabled": "true", "pre_delete_enabled": "true"},
		{"name": "some-errand-2", "post_deploy_enabled": "false", "pre_delete_enabled": "false"},
		{"name": "some-errand-3", "pre_delete_enabled": "true"},
		{"name": "some-errand-4", "post_deploy_enabled": "when-changed"}
	]`

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
			case "/api/v0/staged/products":
				w.Write([]byte(`[
					{"installation_name":"p-bosh","guid":"p-bosh-guid","type":"p-bosh","product_version":"1.10.0.0"},
					{"installation_name":"cf","guid":"cf-guid","type":"cf","product_version":"1.10.0-build.177"},
					{"installation_name":"some-product","guid":"some-product-guid","type":"some-product","product_version":"1.0.0"},
					{"installation_name":"p-isolation-segment","guid":"p-isolation-segment-guid","type":"p-isolation-segment","product_version":"1.10.0-build.31"}
				]`))
			case "/api/v0/staged/products/some-product-guid/errands":
				w.Write([]byte(`{
					"errands": [
						{"post_deploy": "true", "pre_delete": true, "name": "some-errand-1"},
						{"post_deploy": "false", "pre_delete": false, "name": "some-errand-2"},
						{"pre_delete": true, "name": "some-errand-3"},
						{"post_deploy": "when-changed", "name": "some-errand-4"}
					]
				}`))
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	It("lists the errands belonging to the product", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"errands",
			"--product-name", "some-product")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	Context("when json format is requested", func() {
		It("lists the errands belonging to the product in json", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"--format", "json",
				"errands",
				"--product-name", "some-product")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})
})
