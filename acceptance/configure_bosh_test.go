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
		server *httptest.Server
		Forms  []url.Values
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
				w.Write([]byte(`<html>
				<body>
					<form action="/some-form" method="post">
						<input name="_method" value="fakemethod" />
						<input name="authenticity_token" value="fake_authenticity" />
					</form>
					</body>
				</html>`))
			case "/some-form":
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

	It("configures the bosh tile with iaas configuration", func() {
		iaasConfiguration := `{
			"project": "my-project",
			"default_deployment_tag": "my-vms",
			"auth_json": "{\"service_account_key\": \"some-service-key\",\"private_key\": \"some-private-key\"}"
		}`

		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "fake-username",
			"--password", "fake-password",
			"--skip-ssl-validation",
			"configure-bosh",
			"--iaas-configuration", iaasConfiguration)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say("configuring iaas specific options for bosh tile"))
		Expect(session.Out).To(gbytes.Say("finished configuring bosh tile"))

		Expect(Forms[0].Get("iaas_configuration[project]")).To(Equal("my-project"))
		Expect(Forms[0].Get("iaas_configuration[default_deployment_tag]")).To(Equal("my-vms"))
		Expect(Forms[0].Get("iaas_configuration[auth_json]")).To(Equal(`{"service_account_key": "some-service-key","private_key": "some-private-key"}`))
		Expect(Forms[0].Get("authenticity_token")).To(Equal("fake_authenticity"))
		Expect(Forms[0].Get("_method")).To(Equal("fakemethod"))
	})
})
