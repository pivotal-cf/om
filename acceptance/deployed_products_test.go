package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("deployed-products command", func() {
	const tableOutput = `+----------------+------------------+
|      NAME      |     VERSION      |
+----------------+------------------+
| acme-product-1 | 1.13.0-build.100 |
| acme-product-2 | 1.8.0            |
+----------------+------------------+
`

	const jsonOutput = `[
		{"name":"acme-product-1","version":"1.13.0-build.100"},
		{"name":"acme-product-2","version":"1.8.0"}
	]`

	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
				ghttp.RespondWith(http.StatusOK, `{
					"added_products": {
						"deployed": [
							{"name":"acme-product-1","version":"1.13.0-build.100"},
							{"name":"acme-product-2","version":"1.8.0"}
						]
					}
				}`),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("lists the deployed products on Ops Manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"deployed-products",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	When("json format is requested", func() {
		It("lists the deployed products in JSON format", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"deployed-products",
				"--format", "json",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})
})
