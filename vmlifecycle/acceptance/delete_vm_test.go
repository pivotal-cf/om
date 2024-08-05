package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os/exec"
)

var _ = Describe("DeleteVm", func() {
	It("deletes a VM on the targeted IAAS", func() {
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
    public_ip: 1.2.3.4`)
		stateFile := writeFile(`{"iaas": "gcp", "vm_id": "opsman-vm"}`)
		command := exec.Command(pathToMain, "vm-lifecycle", "delete-vm",
			"--config", configFile,
			"--state-file", stateFile,
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 5).Should(gexec.Exit(0))

		Eventually(session.Err).Should(gbytes.Say("gcloud compute instances delete"))

		contents, err := ioutil.ReadFile(stateFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(contents).To(MatchYAML(`{"iaas": "gcp"}`))
	})

	Context("unrecognized flags", func() {
		It("throws an error", func() {
			command := exec.Command(pathToMain, "vm-lifecycle", "delete-vm",
				"--config", "",
				"--state-file", "state.yml",
				"--random",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 5).Should(gexec.Exit(1))

			Eventually(session.Err).Should(gbytes.Say("unknown flag `random'"))
		})
	})
})
