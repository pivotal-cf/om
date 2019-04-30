package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("assign-multi-stemcell command", func() {
	var (
		server *httptest.Server
	)

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var responseString string
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/api/v0/info":
				responseString = `{
						"info": {
							"version": "2.6.0"
						}
					}`

			case "/uaa/oauth/token":
				_ = req.ParseForm()

				if req.PostForm.Get("password") == "" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				responseString = `{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`

			case "/api/v0/stemcells_assignments":
				if req.Method == "GET" {
					responseString = `{"products": 
					[
						{
							"guid": "cf-guid", 
							"identifier": "cf", 
							"available_stemcells": [
							{
								"os": "ubuntu-trusty",
								"version": "1234.5"
							}, {
								"os": "ubuntu-trusty",
								"version": "1234.57"
							}]
						}
					]
				}`
				} else if req.Method == "PATCH" {
					responseString = `{}`
				}
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}

			_, err := w.Write([]byte(responseString))
			Expect(err).ToNot(HaveOccurred())
		}))
	})

	AfterEach(func() {
		server.Close()
	})

	It("successfully sends the stemcell to the Ops Manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "pass",
			"--skip-ssl-validation",
			"assign-multi-stemcell",
			"--product", "cf",
			"--stemcell", "ubuntu-trusty:latest",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say(`assigning stemcells: "ubuntu-trusty 1234.57" to product "cf"`))
		Eventually(session.Out).Should(gbytes.Say("assigned stemcells successfully"))
	})
})
