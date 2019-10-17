package acceptance

import (
	"archive/zip"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var _ = Describe("upload-product command", func() {
	var (
		productFile *os.File
		server      *ghttp.Server
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
						Expect(err).NotTo(HaveOccurred())

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
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 5).Should(gexec.Exit(0))
			Eventually(session.Out, 5).Should(gbytes.Say("processing product"))
			Eventually(session.Out, 5).Should(gbytes.Say("beginning product upload to Ops Manager"))
			Eventually(session.Out, 5).Should(gbytes.Say("finished upload"))
		})
	})

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
				"--target", server.URL(),
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
				"--target", server.URL(),
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
		var uploadCallCount int

		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/available_products"),
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						uploadCallCount++

						server.CloseClientConnections()
						time.Sleep(1 * time.Second)
						return
					}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/available_products"),
					ghttp.RespondWith(http.StatusOK, `[]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/available_products"),
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						uploadCallCount++

						err := req.ParseMultipartForm(100)
						Expect(err).ToNot(HaveOccurred())

						requestFileName := req.MultipartForm.File["product[file]"][0].Filename
						Expect(requestFileName).To(Equal(filepath.Base(productFile.Name())))

						_, err = w.Write([]byte(`{}`))
						Expect(err).ToNot(HaveOccurred())
					}),
				),
			)
		})

		It("retries the upload", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-product",
				"--product", productFile.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 5).Should(gexec.Exit(0))

			Expect(uploadCallCount).To(Equal(2))
		})
	})
})
