package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("configure-saml-authentication command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	It("configures the admin user account on OpsManager with SAML", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/login/ensure_availability"),
				ghttp.RespondWith(http.StatusFound, "", map[string][]string{"Location": {"/setup"}}),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/info"),
				ghttp.RespondWith(http.StatusOK, `{
					"info": {
						"version": "2.6.0"
					}
				}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v0/setup"),
				ghttp.VerifyJSON(`{
      			  	"setup": {
				  		"identity_provider": "saml",
				  		"decryption_passphrase": "passphrase",
				  		"decryption_passphrase_confirmation": "passphrase",
				  		"eula_accepted": "true",
				  		"http_proxy": "http://http-proxy.com",
				  		"https_proxy": "http://https-proxy.com",
				  		"no_proxy": "10.10.10.10,11.11.11.11",
				  		"idp_metadata": "https://saml.example.com:8080",
				  		"bosh_idp_metadata": "https://bosh-saml.example.com:8080",
				  		"rbac_saml_admin_group": "opsman.full_control",
				  		"rbac_saml_groups_attribute": "myenterprise",
				  		"create_bosh_admin_client": true,
				  		"precreated_client_secret": "test-client-secret"
					}
      			}`),
				ghttp.RespondWith(http.StatusOK, ""),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/login/ensure_availability"),
				ghttp.RespondWith(http.StatusOK, "Waiting for authentication system to start..."),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/login/ensure_availability"),
				ghttp.RespondWith(http.StatusFound, "", map[string][]string{"Location": {"/auth/cloudfoundry"}}),
			),
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
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
			"--precreated-client-secret", "test-client-secret",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "5s").Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say("configuring SAML authentication..."))
		Expect(session.Out).To(gbytes.Say("waiting for configuration to complete..."))
		Expect(session.Out).To(gbytes.Say("configuration complete"))
		Expect(session.Out).To(gbytes.Say(`
BOSH admin client will be created when the director is deployed.
The client secret can then be found in the Ops Manager UI:
director tile -> Credentials tab -> click on 'Link to Credential' for 'Uaa Bosh Client Credentials'
Note both the client ID and secret.
`))
		Expect(session.Out).To(gbytes.Say(`
Ops Manager UAA client will be created when authentication system starts.
It will have the username 'precreated-client' and the client secret you provided.
`))
	})
})
