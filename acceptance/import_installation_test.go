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

var _ = Describe("import-installation command", func() {
	var (
		installation                string
		passphrase                  string
		content                     *os.File
		server                      *httptest.Server
		ensureAvailabilityCallCount int
	)

	BeforeEach(func() {
		var err error
		content, err = ioutil.TempFile("", "cool_name.com")
		Expect(err).NotTo(HaveOccurred())

		_, err = content.WriteString("content so validation does not fail")
		Expect(err).NotTo(HaveOccurred())

		ensureAvailabilityCallCount = 0

		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var responseString string
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/login/ensure_availability":
				if ensureAvailabilityCallCount < 2 {
					w.Header().Set("Location", "/setup")
					w.WriteHeader(http.StatusFound)
				} else {
					w.Header().Set("Location", "/auth/cloudfoundry")
					w.WriteHeader(http.StatusFound)
				}
				ensureAvailabilityCallCount++
			case "/api/v0/installation_asset_collection":
				err := req.ParseMultipartForm(100)
				Expect(err).NotTo(HaveOccurred())

				installation = req.MultipartForm.File["installation[file]"][0].Filename
				passphrase = req.MultipartForm.Value["passphrase"][0]
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
		os.Remove(content.Name())
	})

	It("successfully uploads an installation to the Ops Manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--skip-ssl-validation",
			"import-installation",
			"--installation", content.Name(),
			"--decryption-passphrase", "fake-passphrase",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 5).Should(gexec.Exit(0))
		Eventually(session.Out, 5).Should(gbytes.Say("processing installation"))
		Eventually(session.Out, 5).Should(gbytes.Say("beginning installation import to Ops Manager"))
		Eventually(session.Out, 5).Should(gbytes.Say("finished import"))

		Expect(installation).To(Equal(filepath.Base(content.Name())))
		Expect(passphrase).To(Equal("fake-passphrase"))
	})

	Context("when the ops manager is already configured", func() {
		BeforeEach(func() {
			server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				var responseString string
				w.Header().Set("Content-Type", "application/json")

				switch req.URL.Path {
				case "/login/ensure_availability":
					w.Header().Set("Location", "/auth/cloudfoundry")
					w.WriteHeader(http.StatusFound)
				default:
					out, err := httputil.DumpRequest(req, true)
					Expect(err).NotTo(HaveOccurred())
					Fail(fmt.Sprintf("unexpected request: %s", out))
				}

				w.Write([]byte(responseString))
			}))
		})

		It("returns an error", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--skip-ssl-validation",
				"import-installation",
				"--installation", content.Name(),
				"--decryption-passphrase", "fake-passphrase",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 5).Should(gexec.Exit(1))
			Eventually(session.Err, 5).Should(gbytes.Say("cannot import installation to an Ops Manager that is already configured"))
		})
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
					"--skip-ssl-validation",
					"import-installation",
					"--installation", emptyContent.Name(),
					"--decryption-passphrase", "fake-passphrase",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(1))
				Eventually(session.Err, 5).Should(gbytes.Say("failed to load installation: file provided has no content"))
			})
		})

		Context("when the content cannot be read", func() {
			BeforeEach(func() {
				err := os.Remove(content.Name())
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL,
					"--skip-ssl-validation",
					"import-installation",
					"--installation", content.Name(),
					"--decryption-passphrase", "fake-passphrase",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(1))
				Eventually(session.Err, 5).Should(gbytes.Say(`no such file or directory`))
			})
		})
	})
})
