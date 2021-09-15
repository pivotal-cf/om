package acceptance

import (
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/ghttp"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("configure-ldap-authentication command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	It("configures the admin user account on OpsManager with LDAP", func() {
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
						"identity_provider": "ldap",
						"decryption_passphrase": "some-passphrase",
						"decryption_passphrase_confirmation": "some-passphrase",
						"eula_accepted": "true",
						"ldap_settings": {
							"email_attribute": "mail",
							"group_search_base": "ou=groups,dc=opsmanager,dc=com",
							"group_search_filter": "member={0}",
							"ldap_password": "password",
							"ldap_rbac_admin_group_name": "cn=opsmgradmins,ou=groups,dc=opsmanager,dc=com",
							"ldap_referrals": "follow",
							"ldap_username": "cn=admin,dc=opsmanager,dc=com",
							"server_url": "ldap://YOUR-LDAP-SERVER",
							"user_search_base": "ou=users,dc=opsmanager,dc=com",
							"user_search_filter": "cn={0}"
						},
						"create_bosh_admin_client": true
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
			"configure-ldap-authentication",
			"--decryption-passphrase", "some-passphrase",
			"--server-url", "ldap://YOUR-LDAP-SERVER",
			"--ldap-username", "cn=admin,dc=opsmanager,dc=com",
			"--ldap-password", "password",
			"--user-search-base", "ou=users,dc=opsmanager,dc=com",
			"--user-search-filter", "cn={0}",
			"--group-search-base", "ou=groups,dc=opsmanager,dc=com",
			"--group-search-filter", "member={0}",
			"--ldap-rbac-admin-group-name", "cn=opsmgradmins,ou=groups,dc=opsmanager,dc=com",
			"--email-attribute", "mail",
			"--ldap-referrals", "follow",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, "5s").Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say("configuring LDAP authentication..."))
		Expect(session.Out).To(gbytes.Say("waiting for configuration to complete..."))
		Expect(session.Out).To(gbytes.Say("configuration complete"))
		Expect(session.Out).To(gbytes.Say(`
BOSH admin client will be created when the director is deployed.
The client secret can then be found in the Ops Manager UI:
director tile -> Credentials tab -> click on 'Link to Credential' for 'Uaa Bosh Client Credentials'
Note both the client ID and secret.
`))
	})
})
