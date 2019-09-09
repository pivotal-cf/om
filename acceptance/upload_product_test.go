package acceptance

import (
	"archive/zip"
	"crypto/tls"
	"fmt"
	"io"
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

type UploadProductTestServer struct {
	UploadHandler http.Handler
}

func (t *UploadProductTestServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var responseString string
	w.Header().Set("Content-Type", "application/json")

	switch req.URL.Path {
	case "/uaa/oauth/token":
		responseString = `{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`
	case "/api/v0/diagnostic_report":
		responseString = "{}"
	case "/api/v0/available_products":
		if req.Method == "GET" {
			responseString = "[]"
		} else if req.Method == "POST" {
			auth := req.Header.Get("Authorization")
			if auth != "Bearer some-opsman-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			t.UploadHandler.ServeHTTP(w, req)
			return
		}
	default:
		out, err := httputil.DumpRequest(req, true)
		Expect(err).NotTo(HaveOccurred())
		Fail(fmt.Sprintf("unexpected request: %s", out))
	}

	_, err := w.Write([]byte(responseString))
	Expect(err).ToNot(HaveOccurred())
}

var _ = Describe("upload-product command", func() {
	var (
		product       string
		productFile   *os.File
		server        *httptest.Server
		uploadHandler func(http.ResponseWriter, *http.Request)
		snip          chan struct{}
	)

	BeforeEach(func() {
		var err error
		productFile, err = ioutil.TempFile("", "cool_name.com")
		Expect(err).NotTo(HaveOccurred())

		stat, err := productFile.Stat()
		Expect(err).NotTo(HaveOccurred())

		zipper := zip.NewWriter(productFile)

		productWriter, err := zipper.CreateHeader(&zip.FileHeader{
			Name:               "./metadata/some-product.yml",
			UncompressedSize64: uint64(stat.Size()),
			ModifiedTime:       uint16(stat.ModTime().Unix()),
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = io.WriteString(productWriter, `
---
product_version: 1.8.14
name: some-product`)
		Expect(err).NotTo(HaveOccurred())

		err = zipper.Close()
		Expect(err).NotTo(HaveOccurred())

		uploadHandler = func(w http.ResponseWriter, req *http.Request) {
			err := req.ParseMultipartForm(100)
			Expect(err).NotTo(HaveOccurred())

			product = req.MultipartForm.File["product[file]"][0].Filename
			_, err = w.Write([]byte("{}"))
			Expect(err).ToNot(HaveOccurred())
		}

	})

	JustBeforeEach(func() {
		server = httptest.NewUnstartedServer(&UploadProductTestServer{UploadHandler: http.HandlerFunc(uploadHandler)})
		server.TLS = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		server.StartTLS()
	})

	AfterEach(func() {
		server.Close()
	})

	It("successfully uploads a product to the Ops Manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"upload-product",
			"--product", productFile.Name(),
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 5).Should(gexec.Exit(0))
		Eventually(session.Out, 5).Should(gbytes.Say("processing product"))
		Eventually(session.Out, 5).Should(gbytes.Say("beginning product upload to Ops Manager"))
		Eventually(session.Out, 5).Should(gbytes.Say("finished upload"))

		Expect(product).To(Equal(filepath.Base(productFile.Name())))
	})

	When("an error occurs", func() {
		When("the content to upload is empty", func() {
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
					"upload-product",
					"--product", emptyContent.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(1))
				Eventually(session.Err, 5).Should(gbytes.Say("not a valid zip file"))
			})
		})

		When("the content cannot be read", func() {
			BeforeEach(func() {
				err := os.Remove(productFile.Name())
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL,
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"upload-product",
					"--product", productFile.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(1))
				Eventually(session.Err, 5).Should(gbytes.Say(`no such file or directory`))
			})
		})

		When("the server returns EOF during upload", func() {
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
							http.Error(w, fmt.Sprintf("failed to parse request body: %s", err), http.StatusInternalServerError)
							return
						}

						product = req.MultipartForm.File["product[file]"][0].Filename
						_, err = w.Write([]byte("{}"))
						Expect(err).ToNot(HaveOccurred())
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
					"upload-product",
					"--product", productFile.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(0))
			})
		})
	})
})
