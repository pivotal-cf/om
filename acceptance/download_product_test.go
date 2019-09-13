package acceptance

import (
	"archive/zip"
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
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
				pivotalFile := createPivotalFile("[pivnet-example-slug,1.10.1]example*pivotal", "./fixtures/example-product.yml")
				runCommand("mc", "cp", pivotalFile, "testing/"+bucketName+"/some/product/[pivnet-example-slug,1.10.1]example-product.pivotal")
				runCommand("mc", "cp", pivotalFile, "testing/"+bucketName+"/another/stemcell/[stemcells-ubuntu-xenial,97.57]light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz")

				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-file-glob", "example-product.pivotal",
					"--pivnet-product-slug", "pivnet-example-slug",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", "http://127.0.0.1:9001",
					"--stemcell-iaas", "google",
					"--s3-stemcell-path", "/another/stemcell",
					"--s3-product-path", "/some/product",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
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
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(MatchYAML(`{product: example-product, stemcell: "97.57"}`))

				By("running the command again, it uses the cache")
				command = exec.Command(pathToMain, "download-product",
					"--pivnet-file-glob", "*.pivotal",
					"--pivnet-product-slug", "pivnet-example-slug",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", "http://127.0.0.1:9001",
					"--stemcell-iaas", "google",
					"--s3-stemcell-path", "/another/stemcell",
					"--s3-product-path", "/some/product",
				)

				session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(string(session.Err.Contents())).To(ContainSubstring("[pivnet-example-slug,1.10.1]example-product.pivotal already exists, skip downloading"))
				Expect(string(session.Err.Contents())).To(ContainSubstring("[stemcells-ubuntu-xenial,97.57]light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz already exists, skip downloading"))

				cachedFileInfo, err := os.Stat(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.1]example-product.pivotal"))
				Expect(err).ToNot(HaveOccurred())
				Expect(cachedFileInfo.ModTime()).To(Equal(fileInfo.ModTime()))

				Expect(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.1]example-product.pivotal.partial")).ToNot(BeAnExistingFile())
				Expect(filepath.Join(tmpDir, "[stemcells-ubuntu-xenial,97.57]light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz.partial")).ToNot(BeAnExistingFile())
			})
		})

		When("the bucket does not exist", func() {
			It("gives a helpful error message", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", "unknown",
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", "http://127.0.0.1:9001",
					"--s3-enable-v2-signing", "true",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`could not reach provided endpoint and bucket `))
			})
		})

		When("specifying the version of the AWS signature", func() {
			It("supports v2 signing", func() {
				runCommand("mc", "cp", "fixtures/product.yml", "testing/"+bucketName+"/[example-product,1.10.1]product.yml")
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
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
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
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
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", "http://127.0.0.1:9001",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(fmt.Sprintf("could not download product: bucket '%s' contains no files", bucketName)))
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
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
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
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
					"--s3-bucket", bucketName,
					"--s3-access-key-id", "minio",
					"--s3-secret-access-key", "password",
					"--s3-region-name", "unknown",
					"--s3-endpoint", "http://127.0.0.1:9001",
					"--s3-product-path", "/some-path",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))

				Expect(fileContents(tmpDir, "download-file.json")).To(MatchJSON(fmt.Sprintf(`{
					"product_slug": "example-product",
					"product_path": "%s/[example-product,1.10.1]product.yml",
					"product_version": "1.10.1"
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
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "s3",
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
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version-regex", "1.*",
					"--output-directory", tmpDir,
					"--source", "s3",
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
					"product_path": "%s/[example-product,1.10.2]product-456.yml",
					"product_version": "1.10.2"
				}`, tmpDir)))
				Expect(fileContents(tmpDir, "[example-product,1.10.2]product-456.yml")).To(MatchYAML(fmt.Sprintf(`{
					"nothing": "to see here"
				}`)))
			})
		})
	})

	When("downloading from gcs", func() {
		var (
			bucketName string
			serviceAccountKey string
			projectID string
		)

		BeforeEach(func() {
			_, err := exec.LookPath("gsutil")
			if err != nil {
				Skip("gsutil not installed")
			}

			serviceAccountKey = os.Getenv("OM_GCP_SERVICE_ACCOUNT_KEY")
			if serviceAccountKey == "" {
				Skip("OM_GCP_SERVICE_ACCOUNT_KEY is not set")
			}

			projectID = os.Getenv("OM_GCP_PROJECT_ID")
			if projectID == "" {
				Skip("OM_GCP_PROJECT_ID is not set")
			}

			// upload artifact to it
			fmt.Println("**********************************", config.GinkgoConfig.ParallelNode)
			bucketName = fmt.Sprintf("om-acceptance-bucket-%d", config.GinkgoConfig.ParallelNode)
			runCommand("gsutil", "mb","gs://"+bucketName)
		})

		AfterEach(func() {
			runCommand("gsutil", "rm", "-r", "gs://"+bucketName)
		})

		When("specifying the stemcell iaas to download", func() {
			It("downloads the product and correct stemcell", func() {
				pivotalFile := createPivotalFile("[pivnet-example-slug,1.10.1]example*pivotal", "./fixtures/example-product.yml")
				uploadGCSFile(pivotalFile, serviceAccountKey, bucketName, "some/product/[pivnet-example-slug,1.10.1]example-product.pivotal")
				uploadGCSFile(pivotalFile, serviceAccountKey, bucketName, "another/stemcell/[stemcells-ubuntu-xenial,97.57]light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz")

				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-file-glob", "example-product.pivotal",
					"--pivnet-product-slug", "pivnet-example-slug",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "gcs",
					"--gcs-bucket", bucketName,
					"--gcs-service-account-json", serviceAccountKey,
					"--gcs-project-id", projectID,
					"--stemcell-iaas", "google",
					"--gcs-stemcell-path", "another/stemcell",
					"--gcs-product-path", "some/product",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(session.Err).To(gbytes.Say(`attempting to download the file.*example-product.pivotal.*from source gcs`))
				Expect(session.Err).To(gbytes.Say(`attempting to download the file.*light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz.*from source gcs`))
				Expect(session.Err).To(gbytes.Say(`Writing a list of downloaded artifact to download-file.json`))

				fileInfo, err := os.Stat(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.1]example-product.pivotal"))
				Expect(err).ToNot(HaveOccurred())

				By("ensuring an assign stemcell artifact is created")
				contents, err := ioutil.ReadFile(filepath.Join(tmpDir, "assign-stemcell.yml"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(MatchYAML(`{product: example-product, stemcell: "97.57"}`))

				By("running the command again, it uses the cache")
				command = exec.Command(pathToMain, "download-product",
					"--pivnet-file-glob", "example-product.pivotal",
					"--pivnet-product-slug", "pivnet-example-slug",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "gcs",
					"--gcs-bucket", bucketName,
					"--gcs-service-account-json", serviceAccountKey,
					"--gcs-project-id", projectID,
					"--stemcell-iaas", "google",
					"--gcs-stemcell-path", "another/stemcell",
					"--gcs-product-path", "some/product",
				)

				session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(string(session.Err.Contents())).To(ContainSubstring("[pivnet-example-slug,1.10.1]example-product.pivotal already exists, skip downloading"))
				Expect(string(session.Err.Contents())).To(ContainSubstring("[stemcells-ubuntu-xenial,97.57]light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz already exists, skip downloading"))

				cachedFileInfo, err := os.Stat(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.1]example-product.pivotal"))
				Expect(err).ToNot(HaveOccurred())
				Expect(cachedFileInfo.ModTime()).To(Equal(fileInfo.ModTime()))
			})
		})

		When("the bucket does not exist", func() {
			It("gives a helpful error message", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "gcs",
					"--gcs-bucket", "unknown",
					"--gcs-service-account-json", serviceAccountKey,
					"--gcs-project-id", projectID,
					"--stemcell-iaas", "google",
					"--gcs-stemcell-path", "another/stemcell",
					"--gcs-product-path", "some/product",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`could not reach provided bucket 'unknown'`))
			})
		})

		When("no files exist in the bucket", func() {
			// The bucket we get from the before each is already empty, so, no setup
			It("raises an error saying that no automation-downloaded files were found", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "gcs",
					"--gcs-bucket", bucketName,
					"--gcs-service-account-json", serviceAccountKey,
					"--gcs-project-id", projectID,
					"--stemcell-iaas", "google",
					"--gcs-stemcell-path", "another/stemcell",
					"--gcs-product-path", "some/product",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`could not download product: bucket contains no files`))
			})
		})

		When("a file with a prefix for the desired slug/version is not found", func() {
			BeforeEach(func() {
				uploadGCSFile("fixtures/product.yml", serviceAccountKey, bucketName, "example-product-1.10.1_product.yml")
				uploadGCSFile("fixtures/product.yml", serviceAccountKey, bucketName, "still-useless.yml")
				uploadGCSFile("fixtures/product.yml", serviceAccountKey, bucketName, "[example-product,2.22.3]product-456.yml")
				uploadGCSFile("fixtures/product.yml", serviceAccountKey, bucketName, "[example-product,2.22.2]product-123.yml")
			})

			It("raises an error that no files with a prefixed name matching the slug and version are available", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "gcs",
					"--gcs-bucket", bucketName,
					"--gcs-service-account-json", serviceAccountKey,
					"--gcs-project-id", projectID,
					"--stemcell-iaas", "google",
					"--gcs-stemcell-path", "another/stemcell",
					"--gcs-product-path", "some/product",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`no product files with expected prefix \[example-product,1.10.1\] found. Please ensure the file you're trying to download was initially persisted from Pivotal Network net using an appropriately configured download-product command`))
			})
		})

		When("one prefixed file matches the product slug and version", func() {
			BeforeEach(func() {
				uploadGCSFile("fixtures/product.yml", serviceAccountKey, bucketName, "some-path/[example-product,1.10.1]product.yml")
			})

			It("outputs the file and downloaded file metadata", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "gcs",
					"--gcs-bucket", bucketName,
					"--gcs-service-account-json", serviceAccountKey,
					"--gcs-project-id", projectID,
					"--gcs-product-path", "some-path",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))

				Expect(fileContents(tmpDir, "download-file.json")).To(MatchJSON(fmt.Sprintf(`{
					"product_slug": "example-product",
					"product_path": "%s/[example-product,1.10.1]product.yml",
					"product_version": "1.10.1"
				}`, tmpDir)))
				Expect(fileContents(tmpDir, "[example-product,1.10.1]product.yml")).To(MatchYAML(fmt.Sprintf(`{
					"nothing": "to see here"
				}`)))
			})
		})

		When("more than one prefixed file matches the product slug and version", func() {
			BeforeEach(func() {
				uploadGCSFile("fixtures/product.yml", serviceAccountKey, bucketName, "[example-product,1.10.1]product-456.yml")
				uploadGCSFile("fixtures/product.yml", serviceAccountKey, bucketName, "[example-product,1.10.1]product-123.yml")
			})

			It("raises an error that too many files match the glob", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "gcs",
					"--gcs-bucket", bucketName,
					"--gcs-service-account-json", serviceAccountKey,
					"--gcs-project-id", projectID,
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`could not download product: the glob '\*\.yml' matches multiple files`))
			})
		})

		When("using product-regex to find the latest version", func() {
			BeforeEach(func() {
				uploadGCSFile("fixtures/product.yml", serviceAccountKey, bucketName, "[example-product,1.10.2]product-456.yml")
				uploadGCSFile("fixtures/product.yml", serviceAccountKey, bucketName, "[example-product,1.10.1]product-123.yml")
			})

			It("raises an error that too many files match the glob", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--pivnet-file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version-regex", "1.*",
					"--output-directory", tmpDir,
					"--source", "gcs",
					"--gcs-bucket", bucketName,
					"--gcs-service-account-json", serviceAccountKey,
					"--gcs-project-id", projectID,
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(fileContents(tmpDir, "download-file.json")).To(MatchJSON(fmt.Sprintf(`{
					"product_slug": "example-product",
					"product_path": "%s/[example-product,1.10.2]product-456.yml",
					"product_version": "1.10.2"
				}`, tmpDir)))
				Expect(fileContents(tmpDir, "[example-product,1.10.2]product-456.yml")).To(MatchYAML(fmt.Sprintf(`{
					"nothing": "to see here"
				}`)))
			})
		})
	})

	When("downloading from Pivnet", func() {
		var server *ghttp.Server
		var pathToHTTPSPivnet string

		AfterEach(func() {
			server.Close()
		})

		BeforeEach(func() {
			pivotalFile := createPivotalFile("[example-product,1.10.1]example*pivotal", "./fixtures/example-product.yml")
			contents, err := ioutil.ReadFile(pivotalFile)
			Expect(err).NotTo(HaveOccurred())
			modTime := time.Now()

			var fakePivnetMetadataResponse []byte

			fixtureMetadata, err := os.Open("fixtures/example-product.yml")
			defer fixtureMetadata.Close()

			Expect(err).NotTo(HaveOccurred())

			_, err = fixtureMetadata.Read(fakePivnetMetadataResponse)
			Expect(err).NotTo(HaveOccurred())

			server = ghttp.NewTLSServer()
			pathToHTTPSPivnet, err = gexec.Build("github.com/pivotal-cf/om",
				"--ldflags", fmt.Sprintf("-X github.com/pivotal-cf/om/download_clients.pivnetHost=%s", server.URL()))
			Expect(err).NotTo(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases"),
					ghttp.RespondWith(http.StatusOK, `{
  "releases": [
    {
      "id": 24,
      "version": "1.10.1"
    }
  ]
}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases/24"),
					ghttp.RespondWith(http.StatusOK, `{"id":24}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases/24/product_files"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{
  "product_files": [
  {
    "id": 1,
    "aws_object_key": "example-product.pivotal",
    "_links": {
      "download": {
        "href": "%s/api/v2/products/example-product/releases/32/product_files/21/download"
      }
    }
  }
]
}`, server.URL())),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases/24/file_groups"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases/24/product_files/1"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{
"product_file": {
    "id": 1,
	"_links": {
		"download": {
			"href":"%s/api/v2/products/example-product/releases/24/product_files/1/download"
		}
	}
}
}`, server.URL())),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v2/products/example-product/releases/24/product_files/1/download"),
					ghttp.RespondWith(http.StatusFound, "{}", http.Header{"Location": {fmt.Sprintf("%s/api/v2/products/example-product/releases/24/product_files/1/download", server.URL())}}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("HEAD", "/api/v2/products/example-product/releases/24/product_files/1/download"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
			)
			server.RouteToHandler("GET", "/api/v2/products/example-product/releases/24/product_files/1/download",
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases/24/product_files/1/download"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
			)
		})

		It("downloads the product", func() {
			tmpDir, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			command := exec.Command(pathToHTTPSPivnet, "download-product",
				"--pivnet-api-token", "token",
				"--pivnet-file-glob", "example-product.pivotal",
				"--pivnet-product-slug", "example-product",
				"--pivnet-disable-ssl",
				"--product-version", "1.10.1",
				"--output-directory", tmpDir,
			)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(session.Err).To(gbytes.Say(`attempting to download the file.*example-product.pivotal.*from source pivnet`))
			Expect(session.Err).To(gbytes.Say(`Writing a list of downloaded artifact to download-file.json`))

			Expect(filepath.Join(tmpDir, "example-product.pivotal")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "example-product.pivotal.partial")).ToNot(BeAnExistingFile())
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


func uploadGCSFile(localFile, serviceAccountKey, bucketName, objectName string) {
	f, err := os.Open(localFile)
	Expect(err).ToNot(HaveOccurred())

	defer f.Close()

	cxt := context.Background()
	creds, err := google.CredentialsFromJSON(cxt, []byte(serviceAccountKey), storage.ScopeReadWrite)
	Expect(err).ToNot(HaveOccurred())

	client, err := storage.NewClient(cxt, option.WithCredentials(creds))
	Expect(err).ToNot(HaveOccurred())

	wc := client.Bucket(bucketName).Object(objectName).NewWriter(cxt)
	_, err = io.Copy(wc, f)
	Expect(err).ToNot(HaveOccurred())

	err = wc.Close()
	Expect(err).ToNot(HaveOccurred())
}
