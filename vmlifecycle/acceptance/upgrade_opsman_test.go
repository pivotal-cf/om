package integration_test

import (
	"archive/zip"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("UpgradeOpsman", func() {
	createZipFile := func(files []struct{ Name, Body string }) string {
		tmpFile, err := os.CreateTemp("", "")
		w := zip.NewWriter(tmpFile)

		Expect(err).ToNot(HaveOccurred())
		for _, file := range files {
			f, err := w.Create(file.Name)
			if err != nil {
				Expect(err).ToNot(HaveOccurred())
			}
			_, err = f.Write([]byte(file.Body))
			if err != nil {
				Expect(err).ToNot(HaveOccurred())
			}
		}
		err = w.Close()
		Expect(err).ToNot(HaveOccurred())

		return tmpFile.Name()
	}

	It("creates a VM on the targeted IAAS", func() {
		server := ghttp.NewTLSServer()
		server.RouteToHandler("POST", "/uaa/oauth/token",
			ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, `{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`, map[string][]string{
					"Content-Type": {"application/json"},
				}),
			),
		)
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v0/unlock"),
				ghttp.RespondWith(200, "{}"),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/login/ensure_availability"),
				ghttp.RespondWith(302, "", map[string][]string{
					"Location": []string{
						"https://example.com/auth/cloudfoundry",
					},
				}),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/info"),
				ghttp.RespondWith(http.StatusOK, `{
				  "info": {
					"version": "1.1-build.123"
				  }
				}`),
			),
		)

		os.Setenv("OM_VAR_project_name", "dummy-project")
		defer os.Unsetenv("OM_VAR_project_name")

		configFile := writeFile(`
opsman-configuration:
  gcp:
    gcp_service_account: something
    project: ((project_name))
    region: us-west1
    zone: us-west1-c
    vm_name: opsman-vm
    vpc_subnet: dummy-subnet
    tags: good
    custom_cpu: 8
    custom_memory: 16
    boot_disk_size: 400
    public_ip: 1.2.3.4
    private_ip: 10.0.0.2`)
		stateFile := writeFile(`{"iaas": "gcp", "vm_id": "opsman-vm"}`)

		installation := createZipFile([]struct{ Name, Body string }{
			{"installation.yml", ""}})

		fh, err := os.CreateTemp("", "Ops1.1-build.123.yml")
		Expect(err).ToNot(HaveOccurred())
		Expect(fh.Close()).ToNot(HaveOccurred())

		envFile := writeFile(fmt.Sprintf(`
target: %s
username: username
password: password
decryption-passphrase: password
`, server.URL()))

		command := exec.Command(pathToMain, "vm-lifecycle", "upgrade-opsman",
			"--state-file", stateFile,
			"--installation", installation,
			"--env-file", envFile,
			"--image-file", fh.Name(),
			"--vars-env", "OM_VAR",
			"--config", configFile,
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 5).Should(gexec.Exit(0))

		Eventually(session.Err).ShouldNot(gbytes.Say("gcloud"))

		contents, err := os.ReadFile(stateFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(contents).To(MatchYAML(`{"iaas": "gcp", "vm_id": "opsman-vm"}`))
	})
})
