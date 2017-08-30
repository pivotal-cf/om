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
)

var _ = Describe("configure-authentication command", func() {
	It("configures the admin user account on OpsManager", func() {
		var auth struct {
			Setup struct {
				IdentityProvider       string `json:"identity_provider"`
				Username               string `json:"admin_user_name"`
				Password               string `json:"admin_password"`
				PasswordConfirmation   string `json:"admin_password_confirmation"`
				Passphrase             string `json:"decryption_passphrase"`
				PassphraseConfirmation string `json:"decryption_passphrase_confirmation"`
				EULAAccepted           string `json:"eula_accepted"`
				HTTPProxy              string `json:"http_proxy"`
				HTTPSProxy             string `json:"https_proxy"`
				NoProxy                string `json:"no_proxy"`
			} `json:"setup"`
		}
		var ensureAvailabilityCallCount int

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
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
					w.Write([]byte("Waiting for authentication system to start..."))
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
		Expect(auth.Setup.Username).To(Equal("username"))
		Expect(auth.Setup.Password).To(Equal("password"))
		Expect(auth.Setup.PasswordConfirmation).To(Equal("password"))
		Expect(auth.Setup.Passphrase).To(Equal("passphrase"))
		Expect(auth.Setup.PassphraseConfirmation).To(Equal("passphrase"))
		Expect(auth.Setup.EULAAccepted).To(Equal("true"))
		Expect(auth.Setup.HTTPProxy).To(Equal("http://http-proxy.com"))
		Expect(auth.Setup.HTTPSProxy).To(Equal("http://https-proxy.com"))
		Expect(auth.Setup.NoProxy).To(Equal("10.10.10.10,11.11.11.11"))

		Expect(ensureAvailabilityCallCount).To(Equal(3))

		Expect(session.Out).To(gbytes.Say("configuring internal userstore..."))
		Expect(session.Out).To(gbytes.Say("waiting for configuration to complete..."))
		Expect(session.Out).To(gbytes.Say("configuration complete"))
	})
})
