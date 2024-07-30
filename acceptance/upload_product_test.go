package acceptance

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("upload-product command", func() {
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
				ghttp.RespondWith(http.StatusOK, `[]`),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	When("the upload is successful", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/available_products"),
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						err := req.ParseMultipartForm(100)
						Expect(err).ToNot(HaveOccurred())

						requestFileName := req.MultipartForm.File["product[file]"][0].Filename
						Expect(requestFileName).To(Equal(filepath.Base(productFile.Name())))

						w.WriteHeader(http.StatusOK)
						_, err = w.Write([]byte(`{}`))
						Expect(err).ToNot(HaveOccurred())
					}),
				),
			)
		})

		It("successfully uploads a product to the Ops Manager", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-product",
				"--product", productFile.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, 5).Should(gexec.Exit(0))
			Eventually(session.Out, 5).Should(gbytes.Say("processing product"))
			Eventually(session.Out, 5).Should(gbytes.Say("beginning product upload to Ops Manager"))
			Eventually(session.Out, 5).Should(gbytes.Say("finished upload"))
		})

		When("a config file is provided with incorrect verison info", func() {
			It("prints a helpful error message ", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"upload-product",
					"--product", productFile.Name(),
					"--config", writeFile("product-version: 1.8.15"),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(1))
				Eventually(session.Err, 5).Should(gbytes.Say("expected version 1.8.15 does not match product version 1.8.14"))
			})
		})
	})

	When("the content to upload is empty", func() {
		var emptyContent *os.File

		BeforeEach(func() {
			var err error
			emptyContent, err = ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := os.Remove(emptyContent.Name())
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-product",
				"--product", emptyContent.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, 5).Should(gexec.Exit(1))
			Eventually(session.Err, 5).Should(gbytes.Say("not a valid zip file"))
		})
	})

	When("the content cannot be read", func() {
		BeforeEach(func() {
			err := os.Remove(productFile.Name())
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-product",
				"--product", productFile.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, 5).Should(gexec.Exit(1))
			Eventually(session.Err, 5).Should(gbytes.Say(`no such file or directory`))
		})
	})

})
