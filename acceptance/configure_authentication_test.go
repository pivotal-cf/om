package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
)

var _ = Describe("configure-authentication command", func() {
	It("configures the admin user account on OpsManager", func() {
		var auth struct {
			Setup api.SetupInput `json:"setup"`
		}
		var ensureAvailabilityCallCount int

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/api/v0/info":
				_, err := w.Write([]byte(`{
						"info": {
							"version": "2.6.0"
						}
					}`))

				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/setup":
				err := json.NewDecoder(req.Body).Decode(&auth)
				Expect(err).NotTo(HaveOccurred())
			case "/login/ensure_availability":
				ensureAvailabilityCallCount++

				if ensureAvailabilityCallCount == 1 {
					w.Header().Set("Location", "/setup")
					w.WriteHeader(http.StatusFound)
					return
				}

				if ensureAvailabilityCallCount < 3 {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte("Waiting for authentication system to start..."))
					Expect(err).ToNot(HaveOccurred())
					return
				}

				w.Header().Set("Location", "/auth/cloudfoundry")
				w.WriteHeader(http.StatusFound)
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
		defer server.Close()

		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--skip-ssl-validation",
			"configure-authentication",
			"--username", "username",
			"--password", "password",
			"--decryption-passphrase", "passphrase",
			"--http-proxy-url", "http://http-proxy.com",
			"--https-proxy-url", "http://https-proxy.com",
			"--no-proxy", "10.10.10.10,11.11.11.11",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "5s").Should(gexec.Exit(0))

		Expect(auth.Setup.IdentityProvider).To(Equal("internal"))
		Expect(auth.Setup.AdminUserName).To(Equal("username"))
		Expect(auth.Setup.AdminPassword).To(Equal("password"))
		Expect(auth.Setup.AdminPasswordConfirmation).To(Equal("password"))
		Expect(auth.Setup.DecryptionPassphrase).To(Equal("passphrase"))
		Expect(auth.Setup.DecryptionPassphraseConfirmation).To(Equal("passphrase"))
		Expect(auth.Setup.EULAAccepted).To(Equal("true"))
		Expect(auth.Setup.HTTPProxyURL).To(Equal("http://http-proxy.com"))
		Expect(auth.Setup.HTTPSProxyURL).To(Equal("http://https-proxy.com"))
		Expect(auth.Setup.NoProxy).To(Equal("10.10.10.10,11.11.11.11"))

		Expect(ensureAvailabilityCallCount).To(Equal(3))

		Expect(session.Out).To(gbytes.Say("configuring internal userstore..."))
		Expect(session.Out).To(gbytes.Say("waiting for configuration to complete..."))
		Expect(session.Out).To(gbytes.Say("configuration complete"))
	})
})
