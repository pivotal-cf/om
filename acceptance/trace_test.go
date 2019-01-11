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

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("global trace flag", func() {
	var (
		productFile *os.File
		server      *httptest.Server
	)

	const tableOutput = `+--------------+---------+
|     NAME     | VERSION |
+--------------+---------+
| some-product | 1.2.3   |
| p-redis      | 1.7.2   |
+--------------+---------+
`

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
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				_, err := w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/available_products":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				switch req.Method {
				case "GET":
					_, err := w.Write([]byte(`[{"name": "some-product", "product_version": "1.2.3"},{"name":"p-redis","product_version":"1.7.2"}]`))
					Expect(err).ToNot(HaveOccurred())
				case "POST":
					w.WriteHeader(http.StatusOK)
				default:
					Fail(fmt.Sprintf("unexpected method: %s", req.Method))
				}
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	AfterEach(func() {
		os.Remove(productFile.Name())
		server.Close()
	})

	It("prints helpful debug output for http request", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"--trace",
			"available-products")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "40s").Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(ContainSubstring(tableOutput))
		Expect(string(session.Err.Contents())).To(ContainSubstring("GET /api/v0"))
		Expect(string(session.Err.Contents())).To(ContainSubstring("200 OK"))
	})

	It("prints helpful debug output for upload requests", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"--trace",
			"upload-product",
			"--product", productFile.Name(),
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "40s").Should(gexec.Exit(0))

		Expect(string(session.Err.Contents())).To(ContainSubstring("POST /api/v0/available_products"))
		Expect(string(session.Err.Contents())).To(ContainSubstring("200 OK"))
	})
})
