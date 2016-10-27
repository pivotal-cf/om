package acceptance

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const propertiesJSON = `{
	".properties.something": {"value": "configure-me"},
	".a-job.job-property": {"value": {"identity": "username", "password": "example-new-password"} },
	".top-level-property": { "value": [ { "guid": "some-guid", "name": "max", "my-secret": {"secret": "headroom"} } ] }
}`

var _ = Describe("configure-product command", func() {
	var (
		server           *httptest.Server
		propertiesMethod string
		propertiesBody   []byte
	)

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
					{
						"installation_name": "some-product-guid",
						"guid": "some-product-guid",
						"type": "cf"
					},
					{
						"installation_name": "p-bosh-installation-name",
						"guid": "p-bosh-guid",
						"type": "p-bosh"
					}
				]`))
			case "/api/v0/staged/products/some-product-guid/properties":
				var err error
				propertiesMethod = req.Method
				propertiesBody, err = ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				w.Write([]byte(`{}`))
			default:
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	It("successfully configures any product properties", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"configure-product",
			"--product-name", "cf",
			"--product-properties", propertiesJSON,
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say("i know this, setting properties"))
		Expect(session.Out).To(gbytes.Say("finished setting properties, hold on to your butts"))

		Expect(propertiesMethod).To(Equal("PUT"))
		Expect(propertiesBody).To(MatchJSON(fmt.Sprintf(`{"properties": %s}`, propertiesJSON)))
	})
})
