package acceptance

import (
	"archive/zip"
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

var _ = Describe("upload-product command", func() {
	var (
		product     string
		productFile *os.File
		server      *httptest.Server
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

					err := req.ParseMultipartForm(100)
					Expect(err).NotTo(HaveOccurred())

					product = req.MultipartForm.File["product[file]"][0].Filename
					responseString = "{}"
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
		os.Remove(productFile.Name())
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
					"upload-product",
					"--product", emptyContent.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(1))
				Eventually(session.Err, 5).Should(gbytes.Say("not a valid zip file"))
			})
		})

		Context("when the content cannot be read", func() {
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
	})
})
