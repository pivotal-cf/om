package acceptance

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("upload-stemcell command", func() {
	var (
		stemcellName string
		content      *os.File
		server       *httptest.Server
	)

	BeforeEach(func() {
		var err error
		content, err = ioutil.TempFile("", "cool_name.com")
		Expect(err).NotTo(HaveOccurred())

		_, err = content.WriteString("content so validation does not fail")
		Expect(err).NotTo(HaveOccurred())

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
			case "/api/v0/stemcells":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				err := req.ParseMultipartForm(100)
				if err != nil {
					panic(err)
				}

				stemcellName = req.MultipartForm.File["stemcell[file]"][0].Filename
				responseString = "{}"
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}

			w.Write([]byte(responseString))
		}))
	})

	AfterEach(func() {
		err := os.Remove(content.Name())
		Expect(err).NotTo(HaveOccurred())
	})

	It("successfully sends the stemcell to ops-manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"upload-stemcell",
			"--stemcell", content.Name(),
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say("processing stemcell"))
		Eventually(session.Out).Should(gbytes.Say("beginning stemcell upload to Ops Manager"))
		Eventually(session.Out).Should(gbytes.Say("finished upload"))

		Expect(stemcellName).To(Equal(filepath.Base(content.Name())))
	})

	Context("when an error occurs", func() {
		Context("when the content to upload is empty", func() {
			var emptyContent *os.File

			BeforeEach(func() {
				var err error
				emptyContent, err = ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := os.Remove(emptyContent.Name())
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL,
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"upload-stemcell",
					"--stemcell", emptyContent.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Eventually(session.Out).Should(gbytes.Say("failed to load stemcell: file provided has no content"))
			})
		})

		Context("when the content cannot be read", func() {
			BeforeEach(func() {
				err := content.Chmod(000)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL,
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"upload-stemcell",
					"--stemcell", content.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Eventually(session.Out).Should(gbytes.Say(`permission denied`))
			})
		})
	})
})
