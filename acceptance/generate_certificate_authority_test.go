package acceptance

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"
)

var _ = Describe("generate certificate authority", func() {
	const tableOutput = `+----------------------+-------------+--------+------------+------------+-----------------------------+
|          ID          |   ISSUER    | ACTIVE | CREATED ON | EXPIRES ON |        CERTICATE PEM        |
+----------------------+-------------+--------+------------+------------+-----------------------------+
| f7bc18f34f2a7a9403c3 | some-issuer | true   | 2017-01-19 | 2021-01-19 | -----BEGIN CERTIFICATE----- |
|                      |             |        |            |            | fake-cert                   |
|                      |             |        |            |            | -----END CERTIFICATE-----   |
|                      |             |        |            |            |                             |
+----------------------+-------------+--------+------------+------------+-----------------------------+
`

	const jsonOutput = `{
		"guid": "f7bc18f34f2a7a9403c3",
		"issuer": "some-issuer",
		"created_on": "2017-01-19",
		"expires_on": "2021-01-19",
		"active": true,
		"cert_pem": "-----BEGIN CERTIFICATE-----\nfake-cert\n-----END CERTIFICATE-----\n"
	}`

	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities/generate"),
				ghttp.RespondWith(http.StatusOK, `{
					"guid": "f7bc18f34f2a7a9403c3",
					"issuer": "some-issuer",
					"created_on": "2017-01-19",
					"expires_on": "2021-01-19",
					"active": true,
					"cert_pem": "-----BEGIN CERTIFICATE-----\nfake-cert\n-----END CERTIFICATE-----\n"
				}`),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("generates a certificate authority on the OpsMan", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"generate-certificate-authority",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	When("the requested format is JSON", func() {
		It("generates a certificate authority", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"generate-certificate-authority",
				"--format", "json",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})
})
