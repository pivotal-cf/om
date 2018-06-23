package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("apply-changes command", func() {
	var (
		server                       *httptest.Server

	)

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
			case "/api/v0/installations":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" && auth != "Bearer some-running-install-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				if auth == "Bearer some-running-install-opsman-token" {
					w.Write([]byte(`{ "installations": [ { "id": 42, "status": "running", "started_at": "2017-03-02T06:50:32.370Z" } ] }`))
				} else {
					w.Write([]byte(`{ "install": { "id": 42 } }`))
				}
			case "/api/v0/installations/current_log":
				w.Write([]byte(`
event:step_info
data:[{"id":"bosh_product.deploying","description":"Installing BOSH"}]

event:step_state_changed
data:{"type":"step_started","id":"bosh_product.deploying"}

data:This is some data; I do not know what it is; But I do not care.

event:exit
data:{"type":"exit","code":0}

`))
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	It("successfully applies the changes to the Ops Manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"apply-changes")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "40s").Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say(`attempting to apply changes to the targeted Ops Manager`))
		Expect(session.Out).To(gbytes.Say("This is some data; I do not know what it is; But I do not care."))
	})

	It("successfully re-attaches to an existing deployment", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--username", "some-running-install-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"apply-changes")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "40s").Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say(`found already running installation...re-attaching \(Installation ID: 42, Started: Thu Mar  2 06:50:32 UTC 2017\)`))
		Expect(session.Out).To(gbytes.Say("This is some data; I do not know what it is; But I do not care."))
	})

})
