package acceptance

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os/exec"
)

var _ = Describe("download-product command", func() {
	When("downloading from s3", func() {
		runCommand := func(args ...string) {
			fmt.Fprintf(GinkgoWriter, "cmd: %s", args)
			command := exec.Command(args[0], args[1:]...)
			configure, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(configure, "10s").Should(gexec.Exit(0))
		}

		It("supports v2 signing", func() {
			_, err := exec.LookPath("minio")
			if err != nil {
				Skip("minio not installed")
			}
			_, err = exec.LookPath("mc")
			if err != nil {
				Skip("mc not installed")
			}

			// start minio
			dataDir, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			command := exec.Command("minio", "server", "--config-dir", dataDir, "--address", ":9001", dataDir)
			command.Env = []string{
				"MINIO_ACCESS_KEY=minio",
				"MINIO_SECRET_KEY=password",
				"MINIO_BROWSER=off",
				"TERM=xterm-256color",
			}
			minio, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			defer minio.Kill()

			Eventually(minio.Out).Should(gbytes.Say("Endpoint:"))

			// upload artifact to it
			runCommand("mc", "--debug", "config", "host", "add", "testing", "http://127.0.0.1:9001", "minio", "password")
			runCommand("mc", "--debug", "mb", "testing/bucket")
			runCommand("mc", "--debug", "cp", "fixtures/example-product-1.10.1_something-1.10-beta.1.yml", "testing/bucket")

			// try to download it
			tmpDir, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			command = exec.Command(pathToMain, "download-product",
				"--pivnet-api-token", "token",
				"--pivnet-file-glob", "*.yml",
				"--pivnet-product-slug", "example-product",
				"--product-version", "1.10.1",
				"--output-directory", tmpDir,
				"--blobstore", "s3",
				"--s3-config", `{enable-v2-signing: true, region-name: unknown, bucket: bucket, access-key-id: minio, secret-access-key: password, endpoint: http://127.0.0.1:9001}`,
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "5s").Should(gexec.Exit(0))
			Expect(session.Err).To(gbytes.Say(`Writing a list of downloaded artifact to download-file.json`))
		})
	})
})
