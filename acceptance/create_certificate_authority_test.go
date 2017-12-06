package acceptance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("create certificate authority", func() {
	var server *httptest.Server
	const tableOutput = `+----------------------+-------------+--------+------------+------------+------------------------------------------------------------------+
|          ID          |   ISSUER    | ACTIVE | CREATED ON | EXPIRES ON |                          CERTICATE PEM                           |
+----------------------+-------------+--------+------------+------------+------------------------------------------------------------------+
| f7bc18f34f2a7a9403c3 | some-issuer | true   | 2017-01-19 | 2021-01-19 | -----BEGIN CERTIFICATE-----                                      |
|                      |             |        |            |            | MIIC+zCCAeOgAwIBAgIBADANBgkqhkiG9w0BAQsFADAfMQswCQYDVQQGEwJVUzEQ |
|                      |             |        |            |            | MA4GA1UECgwHUGl2b3RhbDAeFw0xNzAxMTgyMTQyMjVaFw0yMTAxMTkyMTQyMjVa |
|                      |             |        |            |            | MB8xCzAJBgNVBAYTAlVTMRAwDgYDVQQKDAdQaXZvdGFsMIIBIjANBgkqhkiG9w0B |
|                      |             |        |            |            | AQEFAAOCAQ8AMIIBCgKCAQEAyV4OhPIIZTEym9OcdcNVip9Ev0ijPPLo9WPLUMzT |
|                      |             |        |            |            | IrpDx3nG/TgD+DP09mwVXfqwBlJmoj9DqRED1x/6bc0Ki/BAFo/P4MmOKm3QnDCt |
|                      |             |        |            |            | o+4RUvLkQqgA++2HYrNTKWJ5fsXmERs8lK9AXXT7RKXhktyWWU3oNGf7zo0e3YKp |
|                      |             |        |            |            | l07DdIW7h1NwIbNcGT1AurIDsxyOZy1HVzLDPtUR2MxhJmSCLsOw3qUDQjatjXKw |
|                      |             |        |            |            | 82RjcrswjG3nv2hvD4/aTOiHuKM3+AGbnmS2MdIOvFOh/7Y79tUp89csK0gs6uOd |
|                      |             |        |            |            | myfdxzDihe4DcKw5CzUTfHKNXgHyeoVOBPcVQTp4lJp1iQIDAQABo0IwQDAdBgNV |
|                      |             |        |            |            | HQ4EFgQUyH4y7VEuImLStXM0CKR8uVqxX/gwDwYDVR0TAQH/BAUwAwEB/zAOBgNV |
|                      |             |        |            |            | HQ8BAf8EBAMCAQYwDQYJKoZIhvcNAQELBQADggEBALmHOPxdyBGnuR0HgR9V4TwJ |
|                      |             |        |            |            | tnKFdFQJGLKVT7am5z6G2Oq5cwACFHWAFfrPG4W9Jm577QtewiY/Rad/PbkY0YSY |
|                      |             |        |            |            | rehLThKdkrfNjxjxI0H2sr7qLBFjJ0wBZHhVmDsO6A9PkfAPu4eJvqRMuL/xGmSQ |
|                      |             |        |            |            | tVkzgYmnCynMNz7FgHyFbd9D9X5YW8fWGSeVBPPikcONdRvjw9aEeAtbGEh8eZCP |
|                      |             |        |            |            | aBQOgsx7b33RuR+CTNqThXY9k8d7/7ba4KVdd4gP8ynFgwvnDQOjcJZ6Go5QY5HA |
|                      |             |        |            |            | R+OgIzs3PFW8pAYcvWrXKR0rE8fL5o9qgTyjmO+5yyyvWIYrKPqqIUIvMCdNr84= |
|                      |             |        |            |            | -----END CERTIFICATE-----                                        |
|                      |             |        |            |            |                                                                  |
+----------------------+-------------+--------+------------+------------+------------------------------------------------------------------+
`
	const jsonOutput = `{
		"guid": "f7bc18f34f2a7a9403c3",
		"issuer": "some-issuer",
		"created_on": "2017-01-19",
		"expires_on": "2021-01-19",
		"active": true,
		"cert_pem": "-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgIBADANBgkqhkiG9w0BAQsFADAfMQswCQYDVQQGEwJVUzEQ\nMA4GA1UECgwHUGl2b3RhbDAeFw0xNzAxMTgyMTQyMjVaFw0yMTAxMTkyMTQyMjVa\nMB8xCzAJBgNVBAYTAlVTMRAwDgYDVQQKDAdQaXZvdGFsMIIBIjANBgkqhkiG9w0B\nAQEFAAOCAQ8AMIIBCgKCAQEAyV4OhPIIZTEym9OcdcNVip9Ev0ijPPLo9WPLUMzT\nIrpDx3nG/TgD+DP09mwVXfqwBlJmoj9DqRED1x/6bc0Ki/BAFo/P4MmOKm3QnDCt\no+4RUvLkQqgA++2HYrNTKWJ5fsXmERs8lK9AXXT7RKXhktyWWU3oNGf7zo0e3YKp\nl07DdIW7h1NwIbNcGT1AurIDsxyOZy1HVzLDPtUR2MxhJmSCLsOw3qUDQjatjXKw\n82RjcrswjG3nv2hvD4/aTOiHuKM3+AGbnmS2MdIOvFOh/7Y79tUp89csK0gs6uOd\nmyfdxzDihe4DcKw5CzUTfHKNXgHyeoVOBPcVQTp4lJp1iQIDAQABo0IwQDAdBgNV\nHQ4EFgQUyH4y7VEuImLStXM0CKR8uVqxX/gwDwYDVR0TAQH/BAUwAwEB/zAOBgNV\nHQ8BAf8EBAMCAQYwDQYJKoZIhvcNAQELBQADggEBALmHOPxdyBGnuR0HgR9V4TwJ\ntnKFdFQJGLKVT7am5z6G2Oq5cwACFHWAFfrPG4W9Jm577QtewiY/Rad/PbkY0YSY\nrehLThKdkrfNjxjxI0H2sr7qLBFjJ0wBZHhVmDsO6A9PkfAPu4eJvqRMuL/xGmSQ\ntVkzgYmnCynMNz7FgHyFbd9D9X5YW8fWGSeVBPPikcONdRvjw9aEeAtbGEh8eZCP\naBQOgsx7b33RuR+CTNqThXY9k8d7/7ba4KVdd4gP8ynFgwvnDQOjcJZ6Go5QY5HA\nR+OgIzs3PFW8pAYcvWrXKR0rE8fL5o9qgTyjmO+5yyyvWIYrKPqqIUIvMCdNr84=\n-----END CERTIFICATE-----\n"
	}`

	const certificatePEM = `-----BEGIN CERTIFICATE-----
MIIC+zCCAeOgAwIBAgIBADANBgkqhkiG9w0BAQsFADAfMQswCQYDVQQGEwJVUzEQ
MA4GA1UECgwHUGl2b3RhbDAeFw0xNzAxMTgyMTQyMjVaFw0yMTAxMTkyMTQyMjVa
MB8xCzAJBgNVBAYTAlVTMRAwDgYDVQQKDAdQaXZvdGFsMIIBIjANBgkqhkiG9w0B
AQEFAAOCAQ8AMIIBCgKCAQEAyV4OhPIIZTEym9OcdcNVip9Ev0ijPPLo9WPLUMzT
IrpDx3nG/TgD+DP09mwVXfqwBlJmoj9DqRED1x/6bc0Ki/BAFo/P4MmOKm3QnDCt
o+4RUvLkQqgA++2HYrNTKWJ5fsXmERs8lK9AXXT7RKXhktyWWU3oNGf7zo0e3YKp
l07DdIW7h1NwIbNcGT1AurIDsxyOZy1HVzLDPtUR2MxhJmSCLsOw3qUDQjatjXKw
82RjcrswjG3nv2hvD4/aTOiHuKM3+AGbnmS2MdIOvFOh/7Y79tUp89csK0gs6uOd
myfdxzDihe4DcKw5CzUTfHKNXgHyeoVOBPcVQTp4lJp1iQIDAQABo0IwQDAdBgNV
HQ4EFgQUyH4y7VEuImLStXM0CKR8uVqxX/gwDwYDVR0TAQH/BAUwAwEB/zAOBgNV
HQ8BAf8EBAMCAQYwDQYJKoZIhvcNAQELBQADggEBALmHOPxdyBGnuR0HgR9V4TwJ
tnKFdFQJGLKVT7am5z6G2Oq5cwACFHWAFfrPG4W9Jm577QtewiY/Rad/PbkY0YSY
rehLThKdkrfNjxjxI0H2sr7qLBFjJ0wBZHhVmDsO6A9PkfAPu4eJvqRMuL/xGmSQ
tVkzgYmnCynMNz7FgHyFbd9D9X5YW8fWGSeVBPPikcONdRvjw9aEeAtbGEh8eZCP
aBQOgsx7b33RuR+CTNqThXY9k8d7/7ba4KVdd4gP8ynFgwvnDQOjcJZ6Go5QY5HA
R+OgIzs3PFW8pAYcvWrXKR0rE8fL5o9qgTyjmO+5yyyvWIYrKPqqIUIvMCdNr84=
-----END CERTIFICATE-----
`
	const privateKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAuiJ0XLnRhBT1IjyU8oaVT72LO2vnb7Witv27gSXkoLjDxmNP
VhvEeTo0XkKgF4QnKcVsN46IIrdsvw6AzJ/LneVwgrr6n0vqA6WUIW2WQTOBDL3Y
IpARxmOEgb9mKhi60RneAx6YHDOglD2Eg0cLkwI+gNFfxQ8LZuyRIUXI8oyyufr9
W8NkJ7GpIkzZrVO0t2TGzyJ1TK+pMX76JxcRTYbOIrbRfznQhIKbPyg+xTkzriS4
vQ1u2YoAVngnIrgwjI19OYdSfJIJaelKnXZXFjT9tSi6L24/Ybps5A/KoyRvAE+b
RIq1Tt0UY8ccDyQn+xf3K1M/eyGU63m6mthMCwIDAQABAoIBAQCrf9N3HD7PVAAI
64jRbO9l6V7AAUvcwZ6KvH5nIGLnM1YvFJGk5TDCAb7+mqSnBjyPYDe1eL42PosT
/mjuIM2bTiu8SEtjOq8DbSxvIGmw6aOd+c2LCvNVt5v/cDrRzrdSsmK8vROp6Ges
LoJJ8svXR9oPFtsG1jXLP2z5GzNrRmT9PXPdYTtHv5sWI2ZEqXsrNQpo5PtxhjGm
9wQHet7+7AtusRLpZt7cckEOHQrclaE8aSebyocjj42yVtVbqRL4N7jeas8hqYPo
Ap+th256O/6urJFs0CazK/GFDUNJylfGKeZYeMAqPEe+tFMxqp7X66278IGgiiAU
V7Qo31cBAoGBAOcXGp/q9BMuu29aDykRf7TWbnAcC4Ru8gOpabJFn0LXEgHVlhFG
TwbXZuvscLBwTUb3MZc+tJqXWp+nV1GyjLi1LKZszK7knmD8+7Vvoc5q9bvxsHtW
jmX0wuSyiyGHLW52uGai8n//nyIDT1y5VEGf1h+HTbH4eaFLW2uJ0XRtAoGBAM4y
zoU2YmiKR+JSWrzIzVpHl42EWwa/ncB5H4tyxRuutd+WiZz5C5VuW7q3qO8Wg3jH
umfh++LZwy3v9Ps11yh4xj59WXIF3yYc26Rw0ZDv/CH5AnzfyWHBRYUARuXKGdyE
KC2zxws9nt7/e+CMORsekf8WQldOyDzBZCPsLcdXAoGAX3HPcVVdUb7vc2JC+Ldd
g5c9LdineR9JnfGO0i6nRLgHm/JXdPMRGMZGoBKbyIPZpwHZ3Znshh0VNPOswPV5
4aASvPoa3/FU6MIURC/DKLpMnD+KoKZzUfDxvftwM3zdas5mAx4yAmPVmfq8AJQb
FK+rhIIhuOvjcJbrP1NAy1ECgYB8xb+0WkFYMvzmlaD0hanFjHbHmqSeQ8sIkgKl
llBxvNmvL1+cThNVXA9DwCkIbB4oMuu4OsX58n2pyX77mAvXIKYNYDqExcrPPD0o
l2AojR+LyytXNu+cKKCRp6Y/HHljt9C8PwId6i69j+l86j0QDQKZUfXY8QI3yWp4
Vk0pRwKBgENWD6Svbk9eL5OJx3sCxAIsLENkVFxg7xDZXj92kFT4E3aI9qR2YFYt
fN5Sp0u06UFdFFwBq3zUo7g85YpLoE8eMZfbUO9aruJsQy3IUsEKrUmH2uqVvcgm
c8Ltdl0ms92X6z4Qh2GiA/URKQLC7yV/kSQfgPEwyITXv4cCqm3o
-----END RSA PRIVATE KEY-----
`

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
			case "/api/v0/certificate_authorities":
				Expect(req.Method).To(Equal(http.MethodPost))

				body, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				expectedJSON, err := json.Marshal(map[string]string{
					"cert_pem":        certificatePEM,
					"private_key_pem": privateKeyPEM,
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(body).To(MatchJSON(expectedJSON))

				responseJSON, err := json.Marshal(map[string]interface{}{
					"guid":       "f7bc18f34f2a7a9403c3",
					"issuer":     "some-issuer",
					"created_on": "2017-01-19",
					"expires_on": "2021-01-19",
					"active":     true,
					"cert_pem":   certificatePEM,
				})
				Expect(err).NotTo(HaveOccurred())

				w.Write([]byte(responseJSON))
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	It("creates a certificate authority on the OpsMan", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"create-certificate-authority",
			"--certificate-pem", certificatePEM,
			"--private-key-pem", privateKeyPEM,
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	Context("when json format is requested", func() {
		It("creates a certificate authority on Ops Man", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"--format", "json",
				"create-certificate-authority",
				"--certificate-pem", certificatePEM,
				"--private-key-pem", privateKeyPEM,
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})
})
