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

var _ = Describe("assign-stemcell command", func() {
	var (
		server *httptest.Server
	)

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var responseString string
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				req.ParseForm()

				if req.PostForm.Get("password") == "" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				responseString = `{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`
			case "/api/v0/stemcell_assignments":
				if req.Method == "GET" {
					responseString = `{"products": [{"guid": "cf-guid", "identifier": "cf", "available_stemcell_versions": ["1234.5, 1234.9"]}]}`
				} else if req.Method == "PATCH" {
					responseString = `{}`
				}
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}

			w.Write([]byte(responseString))
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
			"assign-stemcell",
			"--product", "cf",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say("assigned stemcell successfully"))
	})
})
