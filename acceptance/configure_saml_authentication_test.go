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

var _ = Describe("configure-saml-authentication command", func() {
	It("configures the admin user account on OpsManager with SAML", func() {
		var auth struct {
			Setup api.SetupInput `json:"setup"`
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

		defer server.Close()

		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--skip-ssl-validation",
			"configure-saml-authentication",
			"--saml-idp-metadata", "https://saml.example.com:8080",
			"--saml-bosh-idp-metadata", "https://bosh-saml.example.com:8080",
			"--saml-rbac-admin-group", "opsman.full_control",
			"--saml-rbac-groups-attribute", "myenterprise",
			"--decryption-passphrase", "passphrase",
			"--http-proxy-url", "http://http-proxy.com",
			"--https-proxy-url", "http://https-proxy.com",
			"--no-proxy", "10.10.10.10,11.11.11.11",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "5s").Should(gexec.Exit(0))

		Expect(auth.Setup.IdentityProvider).To(Equal("saml"))
		Expect(auth.Setup.DecryptionPassphrase).To(Equal("passphrase"))
		Expect(auth.Setup.DecryptionPassphraseConfirmation).To(Equal("passphrase"))
		Expect(auth.Setup.EULAAccepted).To(Equal("true"))
		Expect(auth.Setup.HTTPProxyURL).To(Equal("http://http-proxy.com"))
		Expect(auth.Setup.HTTPSProxyURL).To(Equal("http://https-proxy.com"))
		Expect(auth.Setup.NoProxy).To(Equal("10.10.10.10,11.11.11.11"))
		Expect(auth.Setup.IDPMetadata).To(Equal("https://saml.example.com:8080"))
		Expect(auth.Setup.BoshIDPMetadata).To(Equal("https://bosh-saml.example.com:8080"))
		Expect(auth.Setup.RBACAdminGroup).To(Equal("opsman.full_control"))
		Expect(auth.Setup.RBACGroupsAttribute).To(Equal("myenterprise"))

		Expect(ensureAvailabilityCallCount).To(Equal(3))

		Expect(session.Out).To(gbytes.Say("configuring SAML authentication..."))
		Expect(session.Out).To(gbytes.Say("waiting for configuration to complete..."))
		Expect(session.Out).To(gbytes.Say("configuration complete"))
	})
})
