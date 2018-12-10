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
	default:
		out, err := httputil.DumpRequest(req, true)
		Expect(err).NotTo(HaveOccurred())
		Fail(fmt.Sprintf("unexpected request: %s", out))
	}

	w.Write([]byte(responseString))
}

var _ = Describe("upload-stemcell command", func() {
	var (
		stemcellName  string
		content       *os.File
		server        *httptest.Server
		uploadHandler func(http.ResponseWriter, *http.Request)
		snip          chan struct{}
	)

	BeforeEach(func() {
		var err error
		content, err = ioutil.TempFile("", "cool_name.com")
		Expect(err).NotTo(HaveOccurred())

		_, err = content.WriteString("content so validation does not fail")
		Expect(err).NotTo(HaveOccurred())

		uploadHandler = func(w http.ResponseWriter, req *http.Request) {
			err := req.ParseMultipartForm(100)
			if err != nil {
				panic(err)
			}

			stemcellName = req.MultipartForm.File["stemcell[file]"][0].Filename
			w.Write([]byte("{}"))
		}
	})

	JustBeforeEach(func() {
		server = httptest.NewTLSServer(&UploadStemcellTestServer{UploadHandler: http.HandlerFunc(uploadHandler)})
	})

	AfterEach(func() {
		os.Remove(content.Name())
		server.Close()
	})

	It("successfully sends the stemcell to the Ops Manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "pass",
			"--skip-ssl-validation",
			"upload-stemcell",
			"--stemcell", content.Name(),
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say("processing stemcell"))
		Eventually(session.Out).Should(gbytes.Say("beginning stemcell upload to Ops Manager"))
		Eventually(session.Out).Should(gbytes.Say("finished upload"))

		Expect(stemcellName).To(Equal(filepath.Base(content.Name())))
	})

	Context("when the stemcell already exists", func() {
		It("exits early with no error", func() {
			var diagnosticReport []byte
			server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch req.URL.Path {
				case "/uaa/oauth/token":
					w.Write([]byte(`{
						"access_token": "some-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`))
				case "/api/v0/diagnostic_report":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}

					w.Write(diagnosticReport)
				default:
					out, err := httputil.DumpRequest(req, true)
					Expect(err).NotTo(HaveOccurred())
					Fail(fmt.Sprintf("unexpected request: %s", out))
				}
			}))

			diagnosticReport = []byte(fmt.Sprintf(`{
			 "stemcells": [
					"bosh-stemcell-3215-vsphere-esxi-ubuntu-trusty-go_agent.tgz",
					%q
				]
			}`, filepath.Base(content.Name())))

			server.StartTLS()
			defer server.Close()

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

			Eventually(session, 10*time.Second).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("stemcell has already been uploaded"))
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
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"upload-stemcell",
					"--stemcell", emptyContent.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 10*time.Second).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say("failed to load stemcell: file provided has no content"))
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
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"upload-stemcell",
					"--stemcell", content.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 10*time.Second).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say(`no such file or directory`))
			})
		})

		Context("when the server returns EOF during upload", func() {
			BeforeEach(func() {
				snip = make(chan struct{})
				uploadCallCount := 0
				uploadHandler = func(w http.ResponseWriter, req *http.Request) {
					uploadCallCount++

					if uploadCallCount == 1 {
						close(snip)
						return
					} else {
						err := req.ParseMultipartForm(100)
						if err != nil {
							panic(err)
						}

						stemcellName = req.MultipartForm.File["stemcell[file]"][0].Filename
						w.Write([]byte("{}"))
					}
				}
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
					"--stemcell", content.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(0))
			})
		})
	})
})
