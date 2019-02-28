package acceptance

import (
	"archive/zip"
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
			bucketName string
		)

		BeforeEach(func() {
			_, err := exec.LookPath("minio")
			if err != nil {
				Skip("minio not installed")
			}
			_, err = exec.LookPath("mc")
			if err != nil {
				Skip("mc not installed")
			}

			// upload artifact to it
			bucketName = fmt.Sprintf("bucket-%d", config.GinkgoConfig.ParallelNode)
			runCommand("mc", "mb", "--ignore-existing", "testing/"+bucketName)
		})

		AfterEach(func() {
			runCommand("mc", "rm", "--force", "--recursive", "testing/"+bucketName)
		})

		When("specifying the stemcell iaas to download", func() {
			It("downloads the product and correct stemcell", func() {
				pivotalFile := createPivotalFile("[example-product,1.10.1]example*pivotal", "./fixtures/example-product.yml")
				runCommand("mc", "cp", pivotalFile, "testing/"+bucketName+"/[example-product,1.10.1]product.pivotal")
				runCommand("mc", "cp", pivotalFile, "testing/"+bucketName+"/[ubuntu-xenial,97.57]light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz")

				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.pivotal",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--blobstore", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", "http://127.0.0.1:9001",
					"--stemcell-iaas", "google",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(session.Err).To(gbytes.Say(`Writing a list of downloaded artifact to download-file.json`))
			})
		})

		When("specifying the version of the AWS signature", func() {
			It("supports v2 signing", func() {
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/[example-product,1.10.1]product.yml")
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--blobstore", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", "http://127.0.0.1:9001",
					"--s3-enable-v2-signing", "true",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(session.Err).To(gbytes.Say(`Writing a list of downloaded artifact to download-file.json`))
			})

			It("supports v4 signing", func() {
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/[example-product,1.10.1]product.yml")
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--blobstore", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", "http://127.0.0.1:9001",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
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
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", "http://127.0.0.1:9001",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`could not download product: bucket contains no files`))
			})
		})

		When("a file with a prefix for the desired slug/version is not found", func() {
			BeforeEach(func() {
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/example-product-1.10.1_product.yml")
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/still-useless.yml")
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/[example-product,2.22.3]product-456.yml")
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/[example-product,2.22.2]product-123.yml")
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
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", "http://127.0.0.1:9001",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`no product files with expected prefix \[example-product,1.10.1\] found. Please ensure the file you're trying to download was initially persisted from Pivotal Network net using an appropriately configured download-product command`))
			})
		})

		When("one prefixed file matches the product slug and version", func() {
			BeforeEach(func() {
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/some-path/[example-product,1.10.1]product.yml")
				runCommand("mc", "ls", "testing/"+bucketName+"/some-path/")
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
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", "http://127.0.0.1:9001",
					"--s3-path", "/some-path",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))

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
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/[example-product,1.10.1]product-456.yml")
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/[example-product,1.10.1]product-123.yml")
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
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", "http://127.0.0.1:9001",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`could not download product: the glob '\*\.yml' matches multiple files`))
			})
		})

		When("using product-regex to find the latest version", func() {
			BeforeEach(func() {
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/[example-product,1.10.1]product-123.yml")
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/[example-product,1.10.2]product-456.yml")
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
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", "http://127.0.0.1:9001",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
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

func createPivotalFile(productFileName, metadataFilename string) string {
	tempfile, err := ioutil.TempFile("", productFileName)
	Expect(err).NotTo(HaveOccurred())

	zipper := zip.NewWriter(tempfile)
	file, err := zipper.Create("metadata/props.yml")
	Expect(err).NotTo(HaveOccurred())

	contents, err := ioutil.ReadFile(metadataFilename)
	Expect(err).NotTo(HaveOccurred())

	_, err = file.Write(contents)
	Expect(err).NotTo(HaveOccurred())

	Expect(zipper.Close()).NotTo(HaveOccurred())
	return tempfile.Name()
}
