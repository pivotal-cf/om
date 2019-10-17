package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("credentials command", func() {
	const tableOutput = `+-----------------------+
|  some-credential-key  |
+-----------------------+
| some-credential-value |
| newline               |
| another-line          |
| another-line          |
+-----------------------+
`

	const jsonOutput = `{
		"some-credential-key": "some-credential-value\nnewline\nanother-line\nanother-line"
	}`

	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
				ghttp.RespondWith(http.StatusOK, `[
					{"installation_name":"p-bosh","guid":"p-bosh-guid","type":"p-bosh","product_version":"1.10.0.0"},
					{"installation_name":"cf","guid":"cf-guid","type":"cf","product_version":"1.10.0-build.177"},
					{"installation_name":"some-product","guid":"some-product-guid","type":"some-product","product_version":"1.0.0"},
					{"installation_name":"p-isolation-segment","guid":"p-isolation-segment-guid","type":"p-isolation-segment","product_version":"1.10.0-build.31"}
				]`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/deployed/products/some-product-guid/credentials/some-credential"),
				ghttp.RespondWith(http.StatusOK, `{
					"credential": {
						"type": "some-credential-type",
						"value": {
							"some-credential-key": "some-credential-value\nnewline\nanother-line\nanother-line"
						}
					}
				}`),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("fetches a credential of a deployed product", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"credentials",
			"--product-name", "some-product",
			"--credential-reference", "some-credential")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	When("json formatting is requested", func() {
		It("lists credentials of a deployed product", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"credentials",
				"--format", "json",
				"--product-name", "some-product",
				"--credential-reference", "some-credential")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})

	It("outputs a specific credential value of a deployed product", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"credentials",
			"--product-name", "some-product",
			"--credential-reference", "some-credential",
			"--credential-field", "some-credential-key")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(Equal("some-credential-value\nnewline\nanother-line\nanother-line\n"))
	})
})
