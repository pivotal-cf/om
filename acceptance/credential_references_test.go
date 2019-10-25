package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("credential references command", func() {
	const tableOutput = `+------------------------------+
|         CREDENTIALS          |
+------------------------------+
| .my-job.some-credentials     |
| .properties.some-credentials |
+------------------------------+
`

	const jsonOutput = `[
		".my-job.some-credentials",
		".properties.some-credentials"
	]`

	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
				ghttp.RespondWith(http.StatusOK, `[
					{"installation_name": "p-bosh", "guid":"p-bosh-guid", "type":"p-bosh", "product_version":"1.10.0.0"},
					{"installation_name": "cf", "guid":"cf-guid", "type":"cf", "product_version":"1.10.0-build.177"},
					{"installation_name": "some-product", "guid":"some-product-guid", "type":"some-product", "product_version":"1.0.0"},
					{"installation_name": "p-isolation-segment", "guid":"p-isolation-segment-guid", "type":"p-isolation-segment", "product_version":"1.10.0-build.31"}
				]`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/deployed/products/some-product-guid/credentials"),
				ghttp.RespondWith(http.StatusOK, `{
					"credentials": [
						".properties.some-credentials",
						".my-job.some-credentials"
					]
				}`),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("lists the credential references belonging to the deployed product", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"credential-references",
			"--product-name", "some-product")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	When("json formatting is requested", func() {
		It("lists the credential references belonging to the deployed product in json", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"credential-references",
				"--format", "json",
				"--product-name", "some-product")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})
})
