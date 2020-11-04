package acceptance

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"
)

var _ = Describe("When passing a config file to a command", func() {
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

	It("uses the keys in the file as command line args", func() {
		configFile := writeFile(`
username: username
password: password
decryption-passphrase: passphrase
http-proxy-url: http://http-proxy.com
https-proxy-url: http://https-proxy.com
no-proxy: 10.10.10.10,11.11.11.11
`)
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--skip-ssl-validation",
			"configure-authentication",
			"--config", configFile,
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, "5s").Should(gexec.Exit(0))
	})

	When("the keys are invalid command line args", func() {
		It("returns an error", func() {
			configFile := writeFile(`
username123: username
`)
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--skip-ssl-validation",
				"configure-authentication",
				"--config", configFile,
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, "5s").Should(gexec.Exit(1))
			Expect(session.Err.Contents()).To(ContainSubstring("unknown flag `username123'"))
		})
	})

	When("providing command line args and a config file", func() {
		It("makes the config file have lower precedence", func() {
			configFile := writeFile(`
username: invalid-username-in-JSON-verifier
password: password
decryption-passphrase: passphrase
http-proxy-url: http://http-proxy.com
https-proxy-url: http://https-proxy.com
no-proxy: 10.10.10.10,11.11.11.11
`)
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--skip-ssl-validation",
				"configure-authentication",
				"--config", configFile,
				"--username", "username",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, "5s").Should(gexec.Exit(0))
		})
	})

	When("the parameters are set as values in the config file", func() {
		It("can be evaluated by vars file", func() {
			configFile := writeFile(`
username: ((username))
password: password
decryption-passphrase: passphrase
http-proxy-url: http://http-proxy.com
https-proxy-url: http://https-proxy.com
no-proxy: 10.10.10.10,11.11.11.11
`)
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--skip-ssl-validation",
				"configure-authentication",
				"--config", configFile,
				"--vars-file", writeFile(`username: username`),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, "5s").Should(gexec.Exit(0))
		})

		It("can be evaluated by vars environment variables by setting OM_VARS_ENV", func() {
			configFile := writeFile(`
username: ((username))
password: password
decryption-passphrase: passphrase
http-proxy-url: http://http-proxy.com
https-proxy-url: http://https-proxy.com
no-proxy: 10.10.10.10,11.11.11.11
`)
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--skip-ssl-validation",
				"configure-authentication",
				"--config", configFile,
			)
			command.Env = append(command.Env,
				"OM_VARS_ENV=SOME_VAR",
				"SOME_VAR_username=username",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, "5s").Should(gexec.Exit(0))
		})

		It("can be evaluated by vars environment variables", func() {
			configFile := writeFile(`
username: ((username))
password: password
decryption-passphrase: passphrase
http-proxy-url: http://http-proxy.com
https-proxy-url: http://https-proxy.com
no-proxy: 10.10.10.10,11.11.11.11
`)
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--skip-ssl-validation",
				"configure-authentication",
				"--config", configFile,
				"--vars-env", "SOME_VAR",
			)
			command.Env = append(command.Env, "SOME_VAR_username=username")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, "5s").Should(gexec.Exit(0))
		})

		It("can be evaluated by vars command line option", func() {
			configFile := writeFile(`
username: ((username))
password: password
decryption-passphrase: passphrase
http-proxy-url: http://http-proxy.com
https-proxy-url: http://https-proxy.com
no-proxy: 10.10.10.10,11.11.11.11
`)
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--skip-ssl-validation",
				"configure-authentication",
				"--config", configFile,
				"--var", "username=username",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, "5s").Should(gexec.Exit(0))
		})

		When("all the above are provided", func() {
			var configFile string

			BeforeEach(func() {
				configFile = writeFile(`
username: ((username))
password: password
decryption-passphrase: passphrase
http-proxy-url: http://http-proxy.com
https-proxy-url: http://https-proxy.com
no-proxy: 10.10.10.10,11.11.11.11
`)
			})

			It("evaluates --var, --var-file, and env var in interpolation precedence", func() {
				By("having --var win")
				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--skip-ssl-validation",
					"configure-authentication",
					"--config", configFile,
					"--vars-env", "SOME_VAR",
					"--vars-file", writeFile("username: username-from-file"),
					"--var", "username=username",
				)
				command.Env = append(command.Env, "SOME_VAR_username=username-from-env")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session, "5s").Should(gexec.Exit(0))
			})

			It("evaluates --var-file and env var in interpolation precedence", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--skip-ssl-validation",
					"configure-authentication",
					"--config", configFile,
					"--vars-env", "SOME_VAR",
					"--vars-file", writeFile("username: username"),
				)
				command.Env = append(command.Env, "SOME_VAR_username=username-from-env")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session, "5s").Should(gexec.Exit(0))
			})
		})

		When("the variables cannot be evaluated", func() {
			It("returns an error", func() {
				configFile := writeFile(`
username: ((username))
password: password
decryption-passphrase: passphrase
http-proxy-url: http://http-proxy.com
https-proxy-url: http://https-proxy.com
no-proxy: 10.10.10.10,11.11.11.11
`)
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
