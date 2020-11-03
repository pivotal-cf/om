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

var _ = Describe("configure-authentication command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()

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
						"identity_provider": "internal",
						"admin_user_name": "username",
						"admin_password": "password",
						"admin_password_confirmation": "password",
						"decryption_passphrase": "passphrase",
						"decryption_passphrase_confirmation": "passphrase",
						"eula_accepted": "true",
						"http_proxy": "http://http-proxy.com",
						"https_proxy": "http://https-proxy.com",
						"no_proxy": "10.10.10.10,11.11.11.11"
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
	})

	AfterEach(func() {
		server.Close()
	})

	It("configures the admin user account on OpsManager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
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
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, "5s").Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say("configuring internal userstore..."))
		Expect(session.Out).To(gbytes.Say("waiting for configuration to complete..."))
		Expect(session.Out).To(gbytes.Say("configuration complete"))
	})

	It("supports environment variables for username, password, and decryption-passphrase", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--skip-ssl-validation",
			"configure-authentication",
			"--http-proxy-url", "http://http-proxy.com",
			"--https-proxy-url", "http://https-proxy.com",
			"--no-proxy", "10.10.10.10,11.11.11.11",
		)

		command.Env = append(command.Env, []string{
			"OM_USERNAME=username",
			"OM_PASSWORD=password",
			"OM_DECRYPTION_PASSPHRASE=passphrase",
		}...)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, "5s").Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say("configuring internal userstore..."))
		Expect(session.Out).To(gbytes.Say("waiting for configuration to complete..."))
		Expect(session.Out).To(gbytes.Say("configuration complete"))
	})

	When("the config file contains variables", func() {
		var configFile string

		When("required command line arguments are missing", func() {
			It("returns an error", func() {
				configContent := `username: some-username`
				configFile = writeFile(configContent)

				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--skip-ssl-validation",
					"configure-authentication",
					"--config", configFile,
				)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session, "5s").Should(gexec.Exit(1))
				Expect(session.Err.Contents()).To(ContainSubstring("the required flags"))
				Expect(session.Err.Contents()).To(ContainSubstring("were not specified"))
			})
		})

		When("variables are not provided", func() {
			It("returns an error", func() {
				configContent := `
username: some-username
password: ((vars-password))
decryption-passphrase: ((vars-passphrase))
`

				configFile = writeFile(configContent)

				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--skip-ssl-validation",
					"configure-authentication",
					"--config", configFile,
				)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session, "5s").Should(gexec.Exit(1))
				Expect(session.Err.Contents()).To(ContainSubstring("Expected to find variables"))
			})
		})
	})
})
