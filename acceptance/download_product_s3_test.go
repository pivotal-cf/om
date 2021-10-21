package acceptance

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var minioEndpoint = "http://minio:9001"
var minioUser = "minioadmin"
var minioPassword = "minioadmin"

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
			runCommand2(true, "mc", "rb", "--force", "testing/"+bucketName)
		})

		When("specifying the stemcell iaas to download", func() {
			It("downloads the product and correct stemcell", func() {
				pivotalFile := createPivotalFile("[pivnet-example-slug,1.10.1]example*pivotal", "./fixtures/example-product.yml")
				runCommand("mc", "cp", pivotalFile, "testing/"+bucketName+"/some/product/[pivnet-example-slug,1.10.1]example-product.pivotal")
				runCommand("mc", "cp", pivotalFile, "testing/"+bucketName+"/another/stemcell/[stemcells-ubuntu-xenial,97.57]light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz")
				runCommand("mc", "ls", "testing/"+bucketName+"/some/product/")
				runCommand("mc", "ls", "testing/"+bucketName+"/another/stemcell/")

				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--file-glob", "example-product.pivotal",
					"--pivnet-product-slug", "pivnet-example-slug",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", minioUser,
					"--s3-secret-access-key", minioPassword,
					"--s3-region-name", "unknown",
					"--s3-endpoint", minioEndpoint,
					"--stemcell-iaas", "google",
					"--s3-stemcell-path", "/another/stemcell",
					"--s3-product-path", "/some/product",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(session.Err).To(gbytes.Say(`attempting to download the file.*example-product.pivotal.*from source s3`))
				Expect(session.Err).To(gbytes.Say(`attempting to download the file.*light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz.*from source s3`))
				Expect(session.Err).To(gbytes.Say(`Writing a list of downloaded artifact to download-file.json`))

				fileInfo, err := os.Stat(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.1]example-product.pivotal"))
				Expect(err).ToNot(HaveOccurred())
				Expect(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.1]example-product.pivotal.partial")).ToNot(BeAnExistingFile())
				Expect(filepath.Join(tmpDir, "[stemcells-ubuntu-xenial,97.57]light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz.partial")).ToNot(BeAnExistingFile())

				By("ensuring an assign stemcell artifact is created")
				contents, err := ioutil.ReadFile(filepath.Join(tmpDir, "assign-stemcell.yml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(MatchYAML(`{product: example-product, stemcell: "97.57"}`))

				err = ioutil.WriteFile(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.0]example-product.pivotal"), nil, 0777)
				Expect(err).ToNot(HaveOccurred())
				err = ioutil.WriteFile(filepath.Join(tmpDir, "example-product.pivotal"), nil, 0777)
				Expect(err).ToNot(HaveOccurred())

				By("running the command again, it uses the cache")
				command = exec.Command(pathToMain, "download-product",
					"--file-glob", "example*.pivotal",
					"--pivnet-product-slug", "pivnet-example-slug",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", minioUser,
					"--s3-secret-access-key", minioPassword,
					"--s3-region-name", "unknown",
					"--s3-endpoint", minioEndpoint,
					"--stemcell-iaas", "google",
					"--s3-stemcell-path", "/another/stemcell",
					"--s3-product-path", "/some/product",
					"--cache-cleanup", "I acknowledge this will delete files in the output directories",
				)

				session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(string(session.Err.Contents())).To(ContainSubstring("[pivnet-example-slug,1.10.1]example-product.pivotal already exists, skip downloading"))
				Expect(string(session.Err.Contents())).To(ContainSubstring("[stemcells-ubuntu-xenial,97.57]light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz already exists, skip downloading"))

				cachedFileInfo, err := os.Stat(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.1]example-product.pivotal"))
				Expect(err).ToNot(HaveOccurred())
				Expect(cachedFileInfo.ModTime()).To(Equal(fileInfo.ModTime()))

				Expect(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.1]example-product.pivotal.partial")).ToNot(BeAnExistingFile())
				Expect(filepath.Join(tmpDir, "[stemcells-ubuntu-xenial,97.57]light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz.partial")).ToNot(BeAnExistingFile())

				Expect(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.0]example-product.pivotal")).ToNot(BeAnExistingFile())
				Expect(filepath.Join(tmpDir, "example-product.pivotal")).ToNot(BeAnExistingFile())
			})

			When("specifying a heavy stemcell", func() {
				It("downloads the heavy one", func() {
					pivotalFile := createPivotalFile("[pivnet-example-slug,1.10.1]example*pivotal", "./fixtures/example-product.yml")
					runCommand("mc", "cp", pivotalFile, "testing/"+bucketName+"/some/product/[pivnet-example-slug,1.10.1]example-product.pivotal")
					runCommand("mc", "cp", pivotalFile, "testing/"+bucketName+"/another/stemcell/[stemcells-ubuntu-xenial,97.57]bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz")

					tmpDir, err := ioutil.TempDir("", "")
					Expect(err).ToNot(HaveOccurred())
					command := exec.Command(pathToMain, "download-product",
						"--output-directory", tmpDir,
						"--s3-bucket", bucketName,
						"--config", writeFile(`---
file-glob: example-product.pivotal
pivnet-product-slug: pivnet-example-slug
product-version: 1.10.1
source: s3
s3-access-key-id: `+minioUser+`
s3-secret-access-key: `+minioPassword+`
s3-region-name: unknown
s3-endpoint: `+minioEndpoint+`
stemcell-iaas: google
s3-stemcell-path: /another/stemcell
s3-product-path: /some/product
stemcell-heavy: true`))

					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())
					Eventually(session, "10s").Should(gexec.Exit(0))

					_, err = os.Stat(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.1]example-product.pivotal"))
					Expect(err).ToNot(HaveOccurred())
					Expect(session.Err).To(gbytes.Say(`attempting to download the file.*example-product.pivotal.*from source s3`))
					Expect(session.Err).To(gbytes.Say(`attempting to download the file.*bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz.*from source s3`))
				})
			})
		})

		When("the bucket does not exist", func() {
			It("gives a helpful error message", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", "unknown",
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", minioEndpoint,
					"--s3-enable-v2-signing", "true",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`specified bucket does not exist`))
			})
		})

		When("specifying the version of the AWS signature", func() {
			It("supports v2 signing", func() {
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/[example-product,1.10.1]product.yml")
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", minioEndpoint,
					"--s3-enable-v2-signing", "true",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(session.Err).To(gbytes.Say(`Writing a list of downloaded artifact to download-file.json`))
			})

			It("supports v4 signing", func() {
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/[example-product,1.10.1]product.yml")
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", minioEndpoint,
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
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
					"--file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", minioEndpoint,
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(fmt.Sprintf("bucket '%s' contains no files", bucketName)))
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
					"--file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", minioEndpoint,
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`no valid versions found for product "example-product" and product version "1.10.1"`))
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
					"--file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", minioEndpoint,
					"--s3-product-path", "/some-path",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))

				Expect(fileContents(tmpDir, "download-file.json")).To(MatchJSON(fmt.Sprintf(`{
					"product_slug": "example-product",
					"product_path": "%s/[example-product,1.10.1]product.yml",
					"product_version": "1.10.1"
				}`, tmpDir)))
				Expect(fileContents(tmpDir, "[example-product,1.10.1]product.yml")).To(MatchYAML(`{
					"nothing": "to see here"
				}`))
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
					"--file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", minioEndpoint,
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
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
					"--file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version-regex", "1.*",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", minioEndpoint,
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(fileContents(tmpDir, "download-file.json")).To(MatchJSON(fmt.Sprintf(`{
					"product_slug": "example-product",
					"product_path": "%s/[example-product,1.10.2]product-456.yml",
					"product_version": "1.10.2"
				}`, tmpDir)))
				Expect(fileContents(tmpDir, "[example-product,1.10.2]product-456.yml")).To(MatchYAML(`{
					"nothing": "to see here"
				}`))
			})
		})
	})
})
