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

var _ = Describe("global trace flag", func() {
	var server *httptest.Server

	const tableOutput = `+--------------+---------+
|     NAME     | VERSION |
+--------------+---------+
| some-product | 1.2.3   |
| p-redis      | 1.7.2   |
+--------------+---------+
`

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
			case "/api/v0/available_products":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				w.Write([]byte(`[{"name": "some-product", "product_version": "1.2.3"},{"name":"p-redis","product_version":"1.7.2"}]`))
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	It("prints helpful debug output for http request", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"--trace",
			"available-products")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "40s").Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(ContainSubstring(tableOutput))
		Expect(string(session.Err.Contents())).To(ContainSubstring("GET /api/v0"))
		Expect(string(session.Err.Contents())).To(ContainSubstring("200 OK"))
	})
})
