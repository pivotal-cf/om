package acceptance

import (
	"archive/zip"
	"github.com/onsi/gomega/ghttp"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("global trace flag", func() {
	const tableOutput = `+--------------+---------+
|     NAME     | VERSION |
+--------------+---------+
| some-product | 1.2.3   |
| p-redis      | 1.7.2   |
+--------------+---------+
`

	var (
		productFile *os.File
		server      *ghttp.Server
	)

	BeforeEach(func() {
		var err error
		productFile, err = ioutil.TempFile("", "cool_name.com")
		Expect(err).ToNot(HaveOccurred())

		stat, err := productFile.Stat()
		Expect(err).ToNot(HaveOccurred())

		zipper := zip.NewWriter(productFile)

		productWriter, err := zipper.CreateHeader(&zip.FileHeader{
			Name:               "./metadata/some-product.yml",
			UncompressedSize64: uint64(stat.Size()),
			ModifiedTime:       uint16(stat.ModTime().Unix()),
		})
		Expect(err).ToNot(HaveOccurred())

		_, err = io.WriteString(productWriter, `
---
product_version: 1.8.14
name: some-product`)
		Expect(err).ToNot(HaveOccurred())

		err = zipper.Close()
		Expect(err).ToNot(HaveOccurred())

		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/available_products"),
				ghttp.RespondWith(http.StatusOK, `[{
					"name": "some-product",
					"product_version": "1.2.3"
				}, {
					"name": "p-redis",
					"product_version": "1.7.2"
				}]`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v0/available_products"),
				ghttp.RespondWith(http.StatusOK, `{}`),
			),
		)
	})

	AfterEach(func() {
		os.Remove(productFile.Name())
		server.Close()
	})

	It("prints helpful debug output for http request", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"--trace",
			"available-products")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, "40s").Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(ContainSubstring(tableOutput))
		Expect(string(session.Err.Contents())).To(ContainSubstring("GET /api/v0"))
		Expect(string(session.Err.Contents())).To(ContainSubstring("200 OK"))
	})

	It("prints helpful debug output for upload requests", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"--trace",
			"upload-product",
			"--product", productFile.Name(),
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, "40s").Should(gexec.Exit(0))

		Expect(string(session.Err.Contents())).To(ContainSubstring("POST /api/v0/available_products"))
		Expect(string(session.Err.Contents())).To(ContainSubstring("200 OK"))
	})
})
