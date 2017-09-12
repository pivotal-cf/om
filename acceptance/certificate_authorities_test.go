package acceptance

import (
	"os/exec"

	"net/http"
	"net/http/httptest"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("certificate-authorities", func() {
	const expectedOutput = `+------------+----------+--------+------------+------------+------------------------------------------------------+
| ID         | ISSUER   | ACTIVE | CREATED ON | EXPIRES ON | CERTIFICATE PEM                                         |
+------------+----------+--------+------------+------------+------------------------------------------------------+
| some-guid  | Pivotal  | true   | 2017-01-09 | 2021-01-09 | -----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI.... |
| other-guid | Customer | false  | 2017-01-10 | 2021-01-10 | -----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI.... |
+------------+----------+--------+------------+------------+------------------------------------------------------+
`
    var server *httptest.Server

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				username := req.FormValue("username")

				if username == "some-username" {
					w.Write([]byte(`{
						"access_token": "some-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`))
				} else {
					w.Write([]byte(`{
						"access_token": "some-running-install-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`))
				}
			case "/api/v0/certificate_authorities":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				w.Write([]byte(`{
  "certificate_authorities": [
    {
      "guid": "f7bc18f34f2a7a9403c3",
      "issuer": "Pivotal",
      "created_on": "2017-01-09",
      "expires_on": "2021-01-09",
      "active": true,
      "cert_pem": "-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI....etc"
    }
  ]
}`))
			}
		}))
	})

	It("prints a table containing a list of certificate authorities", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"certificate-authorities")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "40s").Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say(expectedOutput))
	})
})