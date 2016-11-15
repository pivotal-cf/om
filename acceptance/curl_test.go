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

var _ = Describe("curl command", func() {
	var server *httptest.Server

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var responseString string
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				responseString = `{
					"access_token": "some-opsman-token",
					"token_type": "bearer",
					"expires_in": 3600
				}`
			case "/api/v0/some-api-endpoint":
				auth := req.Header.Get("Authorization")

				if auth != "Bearer some-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				responseString = `{"some-key": "some-value"}`
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}

			w.Write([]byte(responseString))
		}))
	})

	It("issues an API with credentials", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"curl",
			"--path", "/api/v0/some-api-endpoint",
			"--request", "POST",
			"--data", `{"some-key": "some-value"}`,
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(session.Err.Contents()).To(ContainSubstring("Status: 200 OK"))
		Expect(session.Err.Contents()).To(ContainSubstring("Content-Type: application/json"))
		Expect(string(session.Out.Contents())).To(MatchJSON(`{"some-key": "some-value"}`))
	})
})
