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
	var server *httptest.Server

	BeforeEach(func() {
		var callCount int
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/uaa/oauth/token":
				w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
			case "/api/v0/installations":
			case "/api/v0/stemcells":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				w.Write([]byte(`{ "install": { "id": 42 } }`))
			case "/api/v0/installations/42":
				if callCount != 3 {
					w.Write([]byte(`{"status": "running"}`))
					return
				}

				w.Write([]byte(`{"status": "succeeded"}`))
				callCount++
			case "/api/v0/installations/42/logs":
				w.Write([]byte(`{ "logs": "some logs returned" }`))
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	It("applies changes to an ops-manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"apply-changes",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out).To(gbytes.Say("attempting to apply changes to Ops Manager"))
		Expect(session.Out).To(gbytes.Say("some logs returned"))
		Expect(session.Out).To(gbytes.Say("some logs returned"))
		Expect(session.Out).To(gbytes.Say("some logs returned"))
	})
})
