package acceptance

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("download-product command", func() {
	When("downloading from azure", func() {
		var (
			bucketName     string
			storageAccount string
			key            string
		)

		BeforeEach(func() {
			_, err := exec.LookPath("az")
			if err != nil {
				Skip("az not installed")
			}

			storageAccount = os.Getenv("TEST_AZURE_STORAGE_ACCOUNT")
			if storageAccount == "" {
				Skip("TEST_AZURE_STORAGE_ACCOUNT is not set")
			}

			key = os.Getenv("TEST_AZURE_STORAGE_KEY")
			if key == "" {
				Skip("TEST_AZURE_STORAGE_KEY is not set")
			}

			bucketName = os.Getenv("TEST_AZURE_CONTAINER_NAME")
			if bucketName == "" {
				Skip("TEST_AZURE_CONTAINER_NAME is not set")
			}
		})

		az := func(args ...string) {
			args = append([]string{
				"az",
			}, args...)
			args = append(args, []string{
				"--account-name", storageAccount,
				"--account-key", key,
				"--container-name", bucketName,
			}...)

			runCommand(args...)
		}

		When("specifying the stemcell iaas to download", func() {
			It("downloads the product and correct stemcell", func() {
				pivotalFile := createPivotalFile("[pivnet-example-slug,1.10.1]example*pivotal", "./fixtures/example-product.yml")
				az("storage", "blob", "upload", "--overwrite", "-f", pivotalFile, "-n", "/some/product/[pivnet-example-slug,1.10.1]example-product.pivotal")
				az("storage", "blob", "upload", "--overwrite", "-f", pivotalFile, "-n", "/another/stemcell/[stemcells-ubuntu-xenial,97.57]light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz")

				tmpDir, err := os.MkdirTemp("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain, "download-product",
					"--file-glob", "example-product.pivotal",
					"--pivnet-product-slug", "pivnet-example-slug",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "azure",
					"--stemcell-iaas", "google",
					"--azure-container", bucketName,
					"--azure-storage-account", storageAccount,
					"--azure-storage-key", key,
					"--blobstore-product-path", "/some/product",
					"--blobstore-stemcell-path", "/another/stemcell",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(session.Err).To(gbytes.Say(`attempting to download the file.*example-product.pivotal.*from source azure`))
				Expect(session.Err).To(gbytes.Say(`attempting to download the file.*light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz.*from source azure`))
				Expect(session.Err).To(gbytes.Say(`Writing a list of downloaded artifact to download-file.json`))

				fileInfo, err := os.Stat(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.1]example-product.pivotal"))
				Expect(err).ToNot(HaveOccurred())
				Expect(filepath.Join(tmpDir, "[pivnet-example-slug,1.10.1]example-product.pivotal.partial")).ToNot(BeAnExistingFile())
				Expect(filepath.Join(tmpDir, "[stemcells-ubuntu-xenial,97.57]light-bosh-stemcell-97.57-google-kvm-ubuntu-xenial-go_agent.tgz.partial")).ToNot(BeAnExistingFile())

				By("ensuring an assign stemcell artifact is created")
				contents, err := os.ReadFile(filepath.Join(tmpDir, "assign-stemcell.yml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(MatchYAML(`{product: example-product, stemcell: "97.57"}`))

				By("running the command again, it uses the cache")
				command = exec.Command(pathToMain, "download-product",
					"--file-glob", "*.pivotal",
					"--pivnet-product-slug", "pivnet-example-slug",
					"--product-version", "1.10.1",
					"--output-directory", tmpDir,
					"--source", "azure",
					"--stemcell-iaas", "google",
					"--azure-container", bucketName,
					"--azure-storage-account", storageAccount,
					"--azure-storage-key", key,
					"--blobstore-product-path", "/some/product",
					"--blobstore-stemcell-path", "/another/stemcell",
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
			})
		})
	})
})
