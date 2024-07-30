package integration_test

import (
	"io/ioutil"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("CreateVm", func() {
	It("creates a VM on the targeted IAAS", func() {
		imageFile := writeFile(`---
us: ops-manager-us-uri.tar.gz
eu: ops-manager-eu-uri.tar.gz
asia: ops-manager-asia-uri.tar.gz`)
		configFile := writeFile(`
opsman-configuration:
  gcp:
    gcp_service_account: something
    project: dummy-project
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
		stateFile := writeFile("")
		command := exec.Command(pathToMain, "vm-lifecycle", "create-vm",
			"--image-file", imageFile,
			"--config", configFile,
			"--state-file", stateFile,
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 5).Should(gexec.Exit(0))

		Eventually(session.Err).Should(gbytes.Say("gcloud compute instances create"))

		contents, err := ioutil.ReadFile(stateFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(contents).To(MatchYAML(`{"iaas": "gcp", "vm_id": "opsman-vm"}`))
	})
})
