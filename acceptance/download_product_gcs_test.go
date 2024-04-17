package acceptance

import (
	"encoding/json"
	"fmt"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("download-product command", func() {
	When("downloading from gcs", func() {
		var (
			bucketName        string
			serviceAccountKey string
			projectID         string
		)

		BeforeEach(func() {
			_, err := exec.LookPath("gsutil")
			if err != nil {
				Skip("gsutil not installed")
			}

			serviceAccountKey = os.Getenv("TEST_GCP_SERVICE_ACCOUNT_KEY")
			if serviceAccountKey == "" {
				Skip("TEST_GCP_SERVICE_ACCOUNT_KEY is not set")
			}

			projectID = os.Getenv("TEST_GCP_PROJECT_ID")
			if projectID == "" {
				Skip("TEST_GCP_PROJECT_ID is not set")
			}

			// upload artifact to it
			bucketName = fmt.Sprintf("om-acceptance-bucket-%d", time.Now().UnixNano())

			//log into gcloud
			clientEmail := struct {
				Email string `json:"client_email"`
			}{}
			err = json.Unmarshal([]byte(serviceAccountKey), &clientEmail)
			Expect(err).ToNot(HaveOccurred())
			authFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(authFile.Name(), []byte(serviceAccountKey), 0600)
			Expect(err).ToNot(HaveOccurred())
			Expect(authFile.Close()).ToNot(HaveOccurred())
			runCommand("gcloud", "auth", "activate-service-account", clientEmail.Email, "--key-file", authFile.Name())

			runCommand("gsutil", "mb", "gs://"+bucketName)
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
					"--file-glob", "example-product.pivotal",
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
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "2m").Should(gexec.Exit(0))
				Expect(session.Err).To(gbytes.Say(`attempting to download the file.*example-product.pivotal.*from source google`))
				Expect(session.Err).To(gbytes.Say(`attempting to download the file.*light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz.*from source google`))
				Expect(session.Err).To(gbytes.Say(`Writing a list of downloaded artifact to download-file.json`))

				fileInfo, err := os.Stat(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.1]example-product.pivotal"))
				Expect(err).ToNot(HaveOccurred())

				By("ensuring an assign stemcell artifact is created")
				contents, err := ioutil.ReadFile(filepath.Join(tmpDir, "assign-stemcell.yml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(MatchYAML(`{product: example-product, stemcell: "97.57"}`))

				By("running the command again, it uses the cache")
				command = exec.Command(pathToMain, "download-product",
					"--file-glob", "example-product.pivotal",
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
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "2m").Should(gexec.Exit(0))
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
					"--file-glob", "*.yml",
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
				Expect(err).ToNot(HaveOccurred())
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
					"--file-glob", "*.yml",
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
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`bucket '.*' contains no files`))
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
					"--file-glob", "*.yml",
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
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`no valid versions found for product "example-product" and product version "1.10.1"`))
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
					"--file-glob", "*.yml",
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
				uploadGCSFile("fixtures/product.yml", serviceAccountKey, bucketName, "[example-product,1.10.1]product-456.yml")
				uploadGCSFile("fixtures/product.yml", serviceAccountKey, bucketName, "[example-product,1.10.1]product-123.yml")
			})

			It("raises an error that too many files match the glob", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "gcs",
					"--gcs-bucket", bucketName,
					"--gcs-service-account-json", serviceAccountKey,
					"--gcs-project-id", projectID,
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
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
					"--file-glob", "*.yml",
					"--pivnet-product-slug", "example-product",
					"--product-version-regex", "1.*",
					"--output-directory", tmpDir,
					"--source", "gcs",
					"--gcs-bucket", bucketName,
					"--gcs-service-account-json", serviceAccountKey,
					"--gcs-project-id", projectID,
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
