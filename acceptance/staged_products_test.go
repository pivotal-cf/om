package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("staged-products command", func() {
	var server *httptest.Server

	const tableOutput = `+----------------+------------------+
|      NAME      |     VERSION      |
+----------------+------------------+
| acme-product-1 | 1.13.0-build.100 |
| acme-product-2 | 1.8.9-build.1    |
+----------------+------------------+
`

	const jsonOutput = `[
		{"name":"acme-product-1","version":"1.13.0-build.100"},
		{"name":"acme-product-2","version":"1.8.9-build.1"}
	]`

	BeforeEach(func() {
		diagnosticReport := []byte(`{
			"added_products": {
				"staged": [
					{"name":"acme-product-1","version":"1.13.0-build.100"},
					{"name":"acme-product-2","version":"1.8.9-build.1"}
				]
			}
		}`)

		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
			case "/api/v0/diagnostic_report":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				w.Write(diagnosticReport)
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	It("lists the staged products on Ops Manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"staged-products",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	Context("when json format is requested", func() {
		It("lists the staged products on Ops Manager", func() {
			command := exec.Command(pathToMain,
				"--format", "json",
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"staged-products",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})
})
