package acceptance

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

var _ = Describe("download-product command", func() {
	When("downloading from s3", func() {
		var (
			minio      *gexec.Session
			targetName string
			port       int
		)

		runCommand := func(args ...string) {
			fmt.Fprintf(GinkgoWriter, "cmd: %s", args)
			command := exec.Command(args[0], args[1:]...)
			configure, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(configure, "10s").Should(gexec.Exit(0))
		}

		BeforeEach(func() {
			_, err := exec.LookPath("minio")
			if err != nil {
				Skip("minio not installed")
			}
			fmt.Println("\nRunning minio v2 signing test")
			_, err = exec.LookPath("mc")
			if err != nil {
				Skip("mc not installed")
			}

			// start minio
			dataDir, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			port = 9000 + config.GinkgoConfig.ParallelNode
			command := exec.Command("minio", "server", "--config-dir", dataDir, "--address", fmt.Sprintf(":%d", port), dataDir)
			command.Env = []string{
				"MINIO_ACCESS_KEY=minio",
				"MINIO_SECRET_KEY=password",
				"MINIO_BROWSER=off",
				"TERM=xterm-256color",
			}
			minio, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(minio.Out).Should(gbytes.Say("Endpoint:"))

			// upload artifact to it
			targetName = fmt.Sprintf("testing-%d", port)
			runCommand("mc", "--debug", "config", "host", "add", targetName, fmt.Sprintf("http://127.0.0.1:%d", port), "minio", "password")
			runCommand("mc", "--debug", "mb", targetName+"/bucket")
		})

		AfterEach(func() {
			minio.Kill()
		})

		When("specifying the version of the AWS signature", func() {
			It("supports v2 signing", func() {
				runCommand("mc", "--debug", "cp", "fixtures/product.yml", targetName+"/bucket/[example-product,1.10.1]product.yml")
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--blobstore", "s3",
					"--s3-bucket", "bucket",
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", fmt.Sprintf("http://127.0.0.1:%d", port),
					"--s3-enable-v2-signing", "true",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "5s").Should(gexec.Exit(0))
				Expect(session.Err).To(gbytes.Say(`Writing a list of downloaded artifact to download-file.json`))
			})

			It("supports v4 signing", func() {
				runCommand("mc", "--debug", "cp", "fixtures/product.yml", targetName+"/bucket/[example-product,1.10.1]product.yml")
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--blobstore", "s3",
					"--s3-bucket", "bucket",
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", fmt.Sprintf("http://127.0.0.1:%d", port),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "5s").Should(gexec.Exit(0))
				Expect(session.Err).To(gbytes.Say(`Writing a list of downloaded artifact to download-file.json`))
			})
		})

		When("no files exist in the bucket", func() {
			// The bucket we get from the before each is already empty, so, no setup
			It("raises an error saying that no automation-downloaded files were found", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--blobstore", "s3",
					"--s3-bucket", "bucket",
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", fmt.Sprintf("http://127.0.0.1:%d", port),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "5s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`could not download product: bucket contains no files`))
			})
		})

		When("a file with a prefix for the desired slug/version is not found", func() {
			BeforeEach(func() {
				runCommand("mc", "--debug", "cp", "fixtures/product.yml", targetName+"/bucket/example-product-1.10.1_product.yml")
				runCommand("mc", "--debug", "cp", "fixtures/product.yml", targetName+"/bucket/still-useless.yml")
				runCommand("mc", "--debug", "cp", "fixtures/product.yml", targetName+"/bucket/[example-product,2.22.3]product-456.yml")
				runCommand("mc", "--debug", "cp", "fixtures/product.yml", targetName+"/bucket/[example-product,2.22.2]product-123.yml")
			})

			It("raises an error that no files with a prefixed name matching the slug and version are available", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--blobstore", "s3",
					"--s3-bucket", "bucket",
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", fmt.Sprintf("http://127.0.0.1:%d", port),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "5s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`no product files with expected prefix \[example-product,1.10.1\] found. Please ensure the file you're trying to download was initially persisted from Pivotal Network net using an appropriately configured download-product command`))
			})
		})

		When("one prefixed file matches the product slug and version", func() {
			BeforeEach(func() {
				runCommand("mc", "--debug", "cp", "fixtures/product.yml", targetName+"/bucket/[example-product,1.10.1]product.yml")
			})

			It("outputs the file and downloaded file metadata", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--blobstore", "s3",
					"--s3-bucket", "bucket",
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", fmt.Sprintf("http://127.0.0.1:%d", port),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "5s").Should(gexec.Exit(0))

				Expect(fileContents(tmpDir, "download-file.json")).To(MatchJSON(fmt.Sprintf(`{
					"product_slug": "example-product",
					"product_path": "%s/[example-product,1.10.1]product.yml"
				}`, tmpDir)))
				Expect(fileContents(tmpDir, "[example-product,1.10.1]product.yml")).To(MatchYAML(fmt.Sprintf(`{
					"nothing": "to see here"
				}`)))
			})
		})

		When("more than one prefixed file matches the product slug and version", func() {
			BeforeEach(func() {
				runCommand("mc", "--debug", "cp", "fixtures/product.yml", targetName+"/bucket/[example-product,1.10.1]product-456.yml")
				runCommand("mc", "--debug", "cp", "fixtures/product.yml", targetName+"/bucket/[example-product,1.10.1]product-123.yml")
			})

			It("raises an error that too many files match the glob", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--blobstore", "s3",
					"--s3-bucket", "bucket",
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", fmt.Sprintf("http://127.0.0.1:%d", port),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "5s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`could not download product: the glob '\*\.yml' matches multiple files`))
			})
		})

		When("using product-regex to find the latest version", func() {
			BeforeEach(func() {
				runCommand("mc", "--debug", "cp", "fixtures/product.yml", targetName+"/bucket/[example-product,1.10.1]product-123.yml")
				runCommand("mc", "--debug", "cp", "fixtures/product.yml", targetName+"/bucket/[example-product,1.10.2]product-456.yml")
			})

			It("raises an error that too many files match the glob", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version-regex", "1.*",
					"--output-directory", tmpDir,
					"--blobstore", "s3",
					"--s3-bucket", "bucket",
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", fmt.Sprintf("http://127.0.0.1:%d", port),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "5s").Should(gexec.Exit(0))
				Expect(fileContents(tmpDir, "download-file.json")).To(MatchJSON(fmt.Sprintf(`{
					"product_slug": "example-product",
					"product_path": "%s/[example-product,1.10.2]product-456.yml"
				}`, tmpDir)))
				Expect(fileContents(tmpDir, "[example-product,1.10.2]product-456.yml")).To(MatchYAML(fmt.Sprintf(`{
					"nothing": "to see here"
				}`)))
			})
		})
	})
})

func fileContents(paths ...string) []byte {
	path := filepath.Join(paths...)
	contents, err := ioutil.ReadFile(filepath.Join(path))
	Expect(err).ToNot(HaveOccurred())
	return contents
}
