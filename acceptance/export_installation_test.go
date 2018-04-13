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

			switch req.URL.Path {
			case "/uaa/oauth/token":
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"access_token": "some-opsman-token",
					"token_type": "bearer",
					"expires_in": 3600
				}`))

			case "/api/v0/installation_asset_collection":
				time.Sleep(2 * time.Second)
				w.Write([]byte("some-installation"))

			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}

		}))
	})

	AfterEach(func() {
		os.Remove(outputFileName)
	})

	FIt("successfully exports the installation of the ops-manager", func() {
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
		fmt.Printf("session.Out: %#v\n", string(session.Out.Contents()))
		fmt.Printf("session.Err: %#v\n", string(session.Err.Contents()))
		Expect(session.Out).To(gbytes.Say("exporting installation"))
		Expect(session.Err).To(gbytes.Say("waiting for response"))
		Expect(session.Out).To(gbytes.Say(`100(\.\d+)?%`))
		Expect(session.Out).To(gbytes.Say("finished exporting installation"))

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
				Eventually(session.Err, 5).Should(gbytes.Say("cannot create output file:"))
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
				Eventually(session.Err, 3).Should(gbytes.Say(`.*request canceled \(Client\.Timeout exceeded while awaiting headers\)`))
			})
		})
	})
})
