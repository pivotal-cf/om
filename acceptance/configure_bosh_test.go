package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("configure-bosh command", func() {
	var (
		server          *httptest.Server
		receivedCookies []*http.Cookie
		Forms           []url.Values
	)

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			switch req.URL.Path {
			case "/uaa/oauth/token":
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
			case "/infrastructure/iaas_configuration/edit":
				http.SetCookie(w, &http.Cookie{
					Name:  "somecookie",
					Value: "somevalue",
					Path:  "/",
				})

				w.Write([]byte(`<html>
				<body>
					<form action="/some-form" method="post">
						<input name="_method" value="fakemethod" />
						<input name="authenticity_token" value="fake_authenticity" />
					</form>
					</body>
				</html>`))
			case "/infrastructure/director_configuration/edit":
				http.SetCookie(w, &http.Cookie{
					Name:  "somecookie",
					Value: "somevalue",
					Path:  "/",
				})

				w.Write([]byte(`<html>
				<body>
					<form action="/some-form" method="post">
						<input name="_method" value="fakemethod" />
						<input name="authenticity_token" value="fake_authenticity" />
					</form>
					</body>
				</html>`))
			case "/infrastructure/security_tokens/edit":
				http.SetCookie(w, &http.Cookie{
					Name:  "somecookie",
					Value: "somevalue",
					Path:  "/",
				})

				w.Write([]byte(`<html>
				<body>
					<form action="/some-form" method="post">
						<input name="_method" value="fakemethod" />
						<input name="authenticity_token" value="fake_authenticity" />
					</form>
					</body>
				</html>`))
			case "/some-form":
				receivedCookies = req.Cookies()
				req.ParseForm()
				Forms = append(Forms, req.Form)
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	AfterEach(func() {
		Forms = []url.Values{}
	})

	Context("GCP", func() {
		var (
			command *exec.Cmd
		)

		BeforeEach(func() {
			iaasConfiguration := `{
				"project": "my-project",
				"default_deployment_tag": "my-vms",
				"auth_json": "{\"service_account_key\": \"some-service-key\",\"private_key\": \"some-private-key\"}"
			}`

			directorConfiguration := `{
				"ntp_servers_string": "some-ntp-servers-string",
				"metrics_ip": "some-metrics-ip",
				"hm_pager_duty_options": {
					"enabled": true
				}
			}`

			securityConfiguration := `{
				"trusted_certificates": "some-trusted-certificates",
				"vm_password_type": "some-vm-password-type"
			}`

			command = exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "fake-username",
				"--password", "fake-password",
				"--skip-ssl-validation",
				"configure-bosh",
				"--iaas-configuration", iaasConfiguration,
				"--director-configuration", directorConfiguration,
				"--security-configuration", securityConfiguration)
		})

		It("configures the bosh tile with the provided bosh configuration", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(session.Out).To(gbytes.Say("configuring iaas specific options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("configuring director options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("configuring security options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("finished configuring bosh tile"))

			Expect(receivedCookies).To(HaveLen(1))
			Expect(receivedCookies[0].Name).To(Equal("somecookie"))

			Expect(Forms[0].Get("iaas_configuration[project]")).To(Equal("my-project"))
			Expect(Forms[0].Get("iaas_configuration[default_deployment_tag]")).To(Equal("my-vms"))
			Expect(Forms[0].Get("iaas_configuration[auth_json]")).To(Equal(`{"service_account_key": "some-service-key","private_key": "some-private-key"}`))
			Expect(Forms[0].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[0].Get("_method")).To(Equal("fakemethod"))

			Expect(Forms[1].Get("director_configuration[ntp_servers_string]")).To(Equal("some-ntp-servers-string"))
			Expect(Forms[1].Get("director_configuration[metrics_ip]")).To(Equal("some-metrics-ip"))
			Expect(Forms[1].Get("director_configuration[hm_pager_duty_options][enabled]")).To(Equal("true"))
			Expect(Forms[1].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[1].Get("_method")).To(Equal("fakemethod"))

			Expect(Forms[2].Get("security_tokens[trusted_certificates]")).To(Equal("some-trusted-certificates"))
			Expect(Forms[2].Get("security_tokens[vm_password_type]")).To(Equal("some-vm-password-type"))
			Expect(Forms[2].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[2].Get("_method")).To(Equal("fakemethod"))
		})

		It("does not configure keys that are not part of input", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			_, ok := Forms[0]["iaas_configuration[subscription_id]"]
			Expect(ok).To(BeFalse())
		})
	})

	Context("Azure", func() {
		var command *exec.Cmd

		BeforeEach(func() {
			iaasConfiguration := `{
				"subscription_id": "my-subscription",
				"tenant_id": "my-tenant",
				"client_id": "my-client",
				"client_secret": "my-client-secret",
				"resource_group_name": "my-group",
				"bosh_storage_account_name": "my-storage-account",
				"deployments_storage_account_name": "my-deployments-storage-account",
				"default_security_group": "my-security-group",
				"ssh_public_key": "my-public-key",
				"ssh_private_key": "my-private-key"
			}`

			command = exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "fake-username",
				"--password", "fake-password",
				"--skip-ssl-validation",
				"configure-bosh",
				"--iaas-configuration", iaasConfiguration)
		})

		It("configures the bosh tile with the provided iaas configuration", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(session.Out).To(gbytes.Say("configuring iaas specific options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("finished configuring bosh tile"))

			Expect(Forms[0].Get("iaas_configuration[subscription_id]")).To(Equal("my-subscription"))
			Expect(Forms[0].Get("iaas_configuration[tenant_id]")).To(Equal("my-tenant"))
			Expect(Forms[0].Get("iaas_configuration[client_id]")).To(Equal("my-client"))
			Expect(Forms[0].Get("iaas_configuration[client_secret]")).To(Equal("my-client-secret"))
			Expect(Forms[0].Get("iaas_configuration[resource_group_name]")).To(Equal("my-group"))
			Expect(Forms[0].Get("iaas_configuration[bosh_storage_account_name]")).To(Equal("my-storage-account"))
			Expect(Forms[0].Get("iaas_configuration[deployments_storage_account_name]")).To(Equal("my-deployments-storage-account"))
			Expect(Forms[0].Get("iaas_configuration[default_security_group]")).To(Equal("my-security-group"))
			Expect(Forms[0].Get("iaas_configuration[ssh_public_key]")).To(Equal("my-public-key"))
			Expect(Forms[0].Get("iaas_configuration[ssh_private_key]")).To(Equal("my-private-key"))

			Expect(Forms[0].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[0].Get("_method")).To(Equal("fakemethod"))
		})

		It("does not configure keys that are not part of input", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			_, ok := Forms[0]["iaas_configuration[project]"]
			Expect(ok).To(BeFalse())
		})
	})

	Context("AWS", func() {
		var command *exec.Cmd

		BeforeEach(func() {
			iaasConfiguration := `{
				"access_key_id": "my-access-key",
				"secret_access_key": "my-secret-key",
				"vpc_id": "my-vpc",
				"security_group": "my-security-group",
				"key_pair_name": "my-key-pair",
				"ssh_private_key": "my-private-ssh-key",
				"region": "some-region",
				"encrypted": true
			}`

			command = exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "fake-username",
				"--password", "fake-password",
				"--skip-ssl-validation",
				"configure-bosh",
				"--iaas-configuration", iaasConfiguration)
		})

		It("configures the bosh tile with the provided iaas configuration", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(session.Out).To(gbytes.Say("configuring iaas specific options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("finished configuring bosh tile"))

			Expect(Forms[0].Get("iaas_configuration[access_key_id]")).To(Equal("my-access-key"))
			Expect(Forms[0].Get("iaas_configuration[secret_access_key]")).To(Equal("my-secret-key"))
			Expect(Forms[0].Get("iaas_configuration[vpc_id]")).To(Equal("my-vpc"))
			Expect(Forms[0].Get("iaas_configuration[security_group]")).To(Equal("my-security-group"))
			Expect(Forms[0].Get("iaas_configuration[key_pair_name]")).To(Equal("my-key-pair"))
			Expect(Forms[0].Get("iaas_configuration[ssh_private_key]")).To(Equal("my-private-ssh-key"))
			Expect(Forms[0].Get("iaas_configuration[region]")).To(Equal("some-region"))
			Expect(Forms[0].Get("iaas_configuration[encrypted]")).To(Equal("true"))

			Expect(Forms[0].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[0].Get("_method")).To(Equal("fakemethod"))
		})

		It("does not configure keys that are not part of input", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			_, ok := Forms[0]["iaas_configuration[subscription_id]"]
			Expect(ok).To(BeFalse())
		})
	})
})
