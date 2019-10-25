package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("create certificate authority", func() {
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
	const certificatePEM = `-----BEGIN CERTIFICATE-----
fake-cert
-----END CERTIFICATE-----
`
	const privateKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
fake-key
-----END RSA PRIVATE KEY-----
`

	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	BeforeEach(func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities"),
				ghttp.VerifyJSON(` {
        			"cert_pem": "-----BEGIN CERTIFICATE-----\nfake-cert\n-----END CERTIFICATE-----\n",
        			"private_key_pem": "-----BEGIN RSA PRIVATE KEY-----\nfake-key\n-----END RSA PRIVATE KEY-----\n"
      			}`),
				ghttp.RespondWith(http.StatusOK, `{
					"guid":       "f7bc18f34f2a7a9403c3",
					"issuer":     "some-issuer",
					"created_on": "2017-01-19",
					"expires_on": "2021-01-19",
					"active":     true,
					"cert_pem":   "-----BEGIN CERTIFICATE-----\nfake-cert\n-----END CERTIFICATE-----\n"
				}`),
			),
		)
	})

	It("creates a certificate authority in OpsMan", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"create-certificate-authority",
			"--certificate-pem", certificatePEM,
			"--private-key-pem", privateKeyPEM,
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	When("json format is requested", func() {
		It("creates a certificate authority in OpsMan", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"create-certificate-authority",
				"--format", "json",
				"--certificate-pem", certificatePEM,
				"--private-key-pem", privateKeyPEM,
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})
})
