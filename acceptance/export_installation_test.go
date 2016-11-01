package acceptance

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"os/exec"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("export-installation command", func() {
	var (
		outputFileName string
		server         *httptest.Server
	)

	BeforeEach(func() {
		var err error

		tempFile, err := ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())
		outputFileName = tempFile.Name()

		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var responseString string
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				responseString = `{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`
			case "/api/v0/installation_asset_collection":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				responseString = "some-installation"
				time.Sleep(2 * time.Second)
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}

			w.Write([]byte(responseString))
		}))
	})

	AfterEach(func() {
		os.Remove(outputFileName)
	})

	It("successfully exports the installation of the ops-manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"export-installation",
			"--output-file", outputFileName,
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 5).Should(gexec.Exit(0))
		Eventually(session.Out, 5).Should(gbytes.Say("exporting installation"))
		Eventually(session.Out, 5).Should(gbytes.Say("2s elapsed, waiting for response"))
		Eventually(session.Out, 5).Should(gbytes.Say("finished exporting installation"))

		content, err := ioutil.ReadFile(outputFileName)
		Expect(err).NotTo(HaveOccurred())
		Expect(content).To(Equal([]byte("some-installation")))
	})

	Context("when an error occurs", func() {
		Context("when the output file cannot be written to", func() {
			It("returns an error", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL,
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"export-installation",
					"--output-file", "fake-dir/fake-file",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(1))
				Eventually(session.Out, 5).Should(gbytes.Say("request failed: cannot write to output file: open"))
			})
		})
		Context("when the request takes longer than specified timeout", func() {
			It("returns an error", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL,
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"--request-timeout", "1",
					"export-installation",
					"--output-file", outputFileName,
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 3).Should(gexec.Exit(1))
				Eventually(session.Out, 3).Should(gbytes.Say(`.*request canceled \(Client\.Timeout exceeded while awaiting headers\)`))
			})
		})
	})
})
