package acceptance

import (
	"archive/zip"
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

	createZipFile := func(files []struct{ Name, Body string }) *os.File {
		tmpFile, err := ioutil.TempFile("", "")
		w := zip.NewWriter(tmpFile)

		Expect(err).ToNot(HaveOccurred())
		for _, file := range files {
			f, err := w.Create(file.Name)
			if err != nil {
				Expect(err).ToNot(HaveOccurred())
			}
			_, err = f.Write([]byte(file.Body))
			if err != nil {
				Expect(err).ToNot(HaveOccurred())
			}
		}
		err = w.Close()
		Expect(err).ToNot(HaveOccurred())

		return tmpFile
	}

	BeforeEach(func() {
		content = createZipFile([]struct{ Name, Body string }{
			{"installation.yml", ""},
		})

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

			_, err := w.Write([]byte(responseString))
			Expect(err).ToNot(HaveOccurred())
		}))
	})

	AfterEach(func() {
		server.Close()
	})

	It("successfully uploads an installation to the Ops Manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--decryption-passphrase", "fake-passphrase",
			"--skip-ssl-validation",
			"import-installation",
			"--polling-interval", "0",
			"--installation", content.Name(),
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

				_, err := w.Write([]byte(responseString))
				Expect(err).ToNot(HaveOccurred())
			}))
		})

		It("returns an error", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--decryption-passphrase", "fake-passphrase",
				"--skip-ssl-validation",
				"import-installation",
				"--installation", content.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 5).Should(gexec.Exit(0))
			Eventually(session.Out, 5).Should(gbytes.Say("Ops Manager is already configured"))
		})
	})

	Context("when an error occurs", func() {
		Context("when the content cannot be read", func() {
			BeforeEach(func() {
				err := os.Remove(content.Name())
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL,
					"--decryption-passphrase", "fake-passphrase",
					"--skip-ssl-validation",
					"import-installation",
					"--installation", content.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(1))
				Eventually(session.Err, 5).Should(gbytes.Say(`does not exist. Please check the name and try again.`))
			})
		})
	})
})
