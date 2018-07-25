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

var _ = Describe("create VM extension", func() {
	var server *httptest.Server
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
			case "/api/v0/staged/vm_extensions/some-vm-extension":
				Expect(req.Method).To(Equal(http.MethodPut))

				body, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(body).To(MatchJSON(`
{
  "name": "some-vm-extension",
  "cloud_properties": {
    "iam_instance_profile": "some-iam-profile",
    "elbs": ["some-elb"]
  }
}
`))

				responseJSON, err := json.Marshal([]byte("{}"))
				Expect(err).NotTo(HaveOccurred())

				w.Write([]byte(responseJSON))
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	It("creates a VM extension in OpsMan", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"create-vm-extension",
			"--name", "some-vm-extension",
			"--cloud-properties", "{ \"iam_instance_profile\": \"some-iam-profile\", \"elbs\": [\"some-elb\"] }",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal("VM Extension 'some-vm-extension' created/updated\n"))
	})
})
