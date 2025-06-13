package acceptance

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/extractor"
	"github.com/pivotal-cf/om/validator"
)

var _ = Describe("upload-product command", func() {
	var (
		productFile *os.File
		server      *ghttp.Server
	)

	BeforeEach(func() {
		var err error
		productFile, err = os.CreateTemp("", "cool_name.com")
		Expect(err).ToNot(HaveOccurred())

		stat, err := productFile.Stat()
		Expect(err).ToNot(HaveOccurred())

		zipper := zip.NewWriter(productFile)

		productWriter, err := zipper.CreateHeader(&zip.FileHeader{
			Name:               "./metadata/some-product.yml",
			UncompressedSize64: uint64(stat.Size()),
			Modified:           stat.ModTime(),
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
			emptyContent, err = os.CreateTemp("", "")
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

	When("validating command flags", func() {
		var actualShasum string
		var actualVersion string
		var configFile *os.File
		var varsFile1 *os.File
		var varsFile2 *os.File

		BeforeEach(func() {
			// Calculate the actual shasum of the test file
			shaValidator := validator.NewSHA256Calculator()
			var err error
			actualShasum, err = shaValidator.Checksum(productFile.Name())
			Expect(err).ToNot(HaveOccurred())

			// Get the actual version from the test file
			metadataExtractor := extractor.NewMetadataExtractor()
			metadata, err := metadataExtractor.ExtractFromFile(productFile.Name())
			Expect(err).ToNot(HaveOccurred())
			actualVersion = metadata.Version

			// Create temporary config file
			configFile, err = os.CreateTemp("", "config.yml")
			Expect(err).ToNot(HaveOccurred())
			_, err = configFile.WriteString("product: ((product_path))\npolling-interval: ((interval))")
			Expect(err).ToNot(HaveOccurred())
			Expect(configFile.Close()).To(Succeed())

			// Create temporary vars files
			varsFile1, err = os.CreateTemp("", "vars1.yml")
			Expect(err).ToNot(HaveOccurred())
			_, err = varsFile1.WriteString(fmt.Sprintf("product_path: %s\ninterval: \"5\"", productFile.Name()))
			Expect(err).ToNot(HaveOccurred())
			Expect(varsFile1.Close()).To(Succeed())

			varsFile2, err = os.CreateTemp("", "vars2.yml")
			Expect(err).ToNot(HaveOccurred())
			_, err = varsFile2.WriteString("additional_var: value")
			Expect(err).ToNot(HaveOccurred())
			Expect(varsFile2.Close()).To(Succeed())

			// Add handler for product upload
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

		AfterEach(func() {
			os.Remove(configFile.Name())
			os.Remove(varsFile1.Name())
			os.Remove(varsFile2.Name())
		})

		It("accepts valid flags", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-product",
				"--product", productFile.Name(),
				"--polling-interval", "5",
				"--shasum", actualShasum,
				"--product-version", actualVersion,
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 5).Should(gexec.Exit(0))
		})

		It("accepts valid short flags", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-product",
				"-p", productFile.Name(),
				"-i", "5",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 5).Should(gexec.Exit(0))
		})

		It("accepts all config interpolation flags", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-product",
				"-c", configFile.Name(),
				"--vars-env", "MY",
				"-l", varsFile1.Name(),
				"-l", varsFile2.Name(),
				"-v", "FOO=bar",
				"-v", "BAZ=qux",
				"-p", productFile.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 5).Should(gexec.Exit(0))
		})

		It("accepts mix of config and main flags", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-product",
				"-p", productFile.Name(),
				"-c", configFile.Name(),
				"--vars-env", "MY",
				"--shasum", actualShasum,
				"-l", varsFile1.Name(),
				"-v", "FOO=bar",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 5).Should(gexec.Exit(0))
		})

		It("rejects unknown flags", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-product",
				"-p", productFile.Name(),
				"--notaflag",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 5).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("Error: unknown flag\\(s\\) \\[\"--notaflag\"\\] for command 'upload-product'"))
		})

		It("rejects multiple unknown flags", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-product",
				"--foo",
				"--bar",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 5).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("Error: unknown flag\\(s\\) \\[\"--foo\" \"--bar\"\\] for command 'upload-product'"))
		})

		It("rejects unknown short flags", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-product",
				"-p", productFile.Name(),
				"-x", "18000",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 5).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("Error: unknown flag\\(s\\) \\[\"-x\"\\] for command 'upload-product'"))
		})

		It("rejects another unknown short flag", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-product",
				"-p", productFile.Name(),
				"-z", "18000",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 5).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("Error: unknown flag\\(s\\) \\[\"-z\"\\] for command 'upload-product'"))
		})
	})
})
