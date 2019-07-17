package acceptance

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type UploadStemcellTestServer struct {
	UploadHandler http.Handler
}

func (t *UploadStemcellTestServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var responseString string
	w.Header().Set("Content-Type", "application/json")

	switch req.URL.Path {
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
	case "/api/v0/diagnostic_report":
		responseString = "{}"
	case "/api/v0/stemcells":
		auth := req.Header.Get("Authorization")

		if auth != "Bearer some-opsman-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		t.UploadHandler.ServeHTTP(w, req)
		return
	case "/api/v0/info":
		responseString = `{
			"info": {
				"version": "2.1-build.79"
			}
		}`
	default:
		out, err := httputil.DumpRequest(req, true)
		Expect(err).NotTo(HaveOccurred())
		Fail(fmt.Sprintf("unexpected request: %s", out))
	}

	_, err := w.Write([]byte(responseString))
	Expect(err).ToNot(HaveOccurred())
}

var _ = Describe("upload-stemcell command", func() {
	var stemcellName string

	createStemcell := func(filename string) string {
		dir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		path := filepath.Join(dir, filename)
		err = ioutil.WriteFile(path, []byte("content so validation does not fail"), 0777)
		Expect(err).NotTo(HaveOccurred())
		return path
	}

	createSuccessfulUploadHandler := func() func(w http.ResponseWriter, req *http.Request) {
		return func(w http.ResponseWriter, req *http.Request) {
			err := req.ParseMultipartForm(100)
			if err != nil {
				panic(err)
			}

			stemcellName = req.MultipartForm.File["stemcell[file]"][0].Filename
			_, err = w.Write([]byte("{}"))
			Expect(err).ToNot(HaveOccurred())
		}
	}
	setupServerWithUploadHandler := func(f func(w http.ResponseWriter, req *http.Request)) *httptest.Server {
		return httptest.NewTLSServer(&UploadStemcellTestServer{UploadHandler: http.HandlerFunc(f)})
	}

	It("successfully sends the stemcell to the Ops Manager", func() {
		server := setupServerWithUploadHandler(createSuccessfulUploadHandler())
		defer server.Close()

		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "pass",
			"--skip-ssl-validation",
			"upload-stemcell",
			"--stemcell", createStemcell("stemcell.tgz"),
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say("processing stemcell"))
		Eventually(session.Out).Should(gbytes.Say("beginning stemcell upload to Ops Manager"))
		Eventually(session.Out).Should(gbytes.Say("finished upload"))

		Expect(stemcellName).To(Equal("stemcell.tgz"))
	})

	When("the stemcell name has the `download-product` prefix", func() {
		It("successfully sends the stemcell to the Ops Manager", func() {
			server := setupServerWithUploadHandler(createSuccessfulUploadHandler())
			defer server.Close()

			filename := createStemcell("[ubuntu-xenial,97.88]stemcell.tgz")
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "pass",
				"--skip-ssl-validation",
				"upload-stemcell",
				"--stemcell", filename,
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 10*time.Second).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("processing stemcell"))
			Eventually(session.Out).Should(gbytes.Say("beginning stemcell upload to Ops Manager"))
			Eventually(session.Out).Should(gbytes.Say("finished upload"))

			Expect(stemcellName).To(Equal("stemcell.tgz"))
			Expect(filename).To(BeAnExistingFile())
		})
	})

	Context("when the stemcell already exists", func() {
		It("exits early with no error", func() {
			var diagnosticReport []byte
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch req.URL.Path {
				case "/uaa/oauth/token":
					_, err := w.Write([]byte(`{
						"access_token": "some-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`))
					Expect(err).ToNot(HaveOccurred())
				case "/api/v0/diagnostic_report":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}

					_, err := w.Write(diagnosticReport)
					Expect(err).ToNot(HaveOccurred())
				case "/api/v0/info":
					_, err := w.Write([]byte(`{
						"info": {
							"version": "2.1-build.79"
						}
					}`))

					Expect(err).ToNot(HaveOccurred())
				default:
					out, err := httputil.DumpRequest(req, true)
					Expect(err).NotTo(HaveOccurred())
					Fail(fmt.Sprintf("unexpected request: %s", out))
				}
			}))

			diagnosticReport = []byte(`{
			 "stemcells": [
					"bosh-stemcell-3215-vsphere-esxi-ubuntu-trusty-go_agent.tgz",
					"stemcell.tgz"
				]
			}`)

			server.StartTLS()
			defer server.Close()

			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-stemcell",
				"--stemcell", createStemcell("stemcell.tgz"),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 10*time.Second).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("stemcell has already been uploaded"))
		})
	})

	Context("when an error occurs", func() {
		Context("when the content to upload is empty", func() {
			It("returns an error", func() {
				emptyContent, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())
				server := setupServerWithUploadHandler(func(w http.ResponseWriter, req *http.Request) {})
				defer server.Close()

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

				Eventually(session, 10*time.Second).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say("failed to upload stemcell: file provided has no content"))
			})
		})

		Context("when the content cannot be read", func() {
			It("returns an error", func() {
				server := setupServerWithUploadHandler(func(w http.ResponseWriter, req *http.Request) {})
				defer server.Close()

				command := exec.Command(pathToMain,
					"--target", server.URL,
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"upload-stemcell",
					"--stemcell", "/unknown/path/whatever",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 10*time.Second).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say(`no such file or directory`))
			})
		})

		Context("when the server returns EOF during upload", func() {
			var (
				snip   chan struct{}
				server *httptest.Server
			)

			BeforeEach(func() {
				snip = make(chan struct{})
				uploadCallCount := 0
				server = setupServerWithUploadHandler(func(w http.ResponseWriter, req *http.Request) {
					uploadCallCount++

					if uploadCallCount == 1 {
						close(snip)
						return
					} else {
						err := req.ParseMultipartForm(100)
						if err != nil {
							panic(err)
						}

						_, err = w.Write([]byte("{}"))
						Expect(err).ToNot(HaveOccurred())
					}
				})
			})

			JustBeforeEach(func() {
				go func() {
					<-snip

					server.CloseClientConnections()
				}()
			})

			It("retries the upload", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL,
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"upload-stemcell",
					"--stemcell", createStemcell("stemcell.tgz"),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(0))
			})
		})
	})
})
