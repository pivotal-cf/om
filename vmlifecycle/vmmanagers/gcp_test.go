package vmmanagers_test

import (
	"errors"
	"fmt"
	"github.com/pivotal-cf/om/vmlifecycle/matchers"
	"io/ioutil"

	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers/fakes"
	"gopkg.in/yaml.v2"
)

var _ = Describe("GCP VMManager", func() {
	createCommand := func(region string, configStrTemplate string) (*vmmanagers.GCPVMManager, *fakes.GCloudRunner) {
		var err error
		runner := &fakes.GCloudRunner{}
		testUriFile, err := ioutil.TempFile("", "some*.yml")
		Expect(err).ToNot(HaveOccurred())
		_, _ = testUriFile.WriteString(`
---
us: ops-manager-us-uri.tar.gz
eu: ops-manager-eu-uri.tar.gz
asia: ops-manager-asia-uri.tar.gz
image:
 name: ops-manager-2-9-10-build-177
 project: pivotal-ops-manager-images
`)
		Expect(testUriFile.Close()).ToNot(HaveOccurred())

		state := vmmanagers.StateInfo{
			IAAS: "gcp",
		}

		var validConfig *vmmanagers.OpsmanConfigFilePayload
		err = yaml.Unmarshal([]byte(fmt.Sprintf(configStrTemplate, region)), &validConfig)
		Expect(err).ToNot(HaveOccurred())

		command := vmmanagers.NewGcpVMManager(validConfig, testUriFile.Name(), state, runner)
		return command, runner
	}

	Describe("create vm", func() {
		It("requires private ip and/or public ip to be defined", func() {
			const configStrTemplate = `
opsman-configuration:
  gcp:
    version: 1.2.3
    gcp_service_account: something
    project: dummy-project
    region: %s
    zone: us-west1-c
    vm_name: opsman-vm
    vpc_network: dummy-network
    vpc_subnet: dummy-subnet
    tags: good
    custom_cpu: 8
    custom_memory: 16
    boot_disk_size: 400
`
			command, _ := createCommand("us-west1", configStrTemplate)
			_, _, err := command.CreateVM()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("PublicIP and/or PrivateIP must be set"))
		})

		It("requires gcp_service_account or gcp_service_account_name to be defined", func() {
			const configStrTemplate = `
opsman-configuration:
  gcp:
    version: 1.2.3
    project: dummy-project
    region: %s
    zone: us-west1-c
    vm_name: opsman-vm
    vpc_network: dummy-network
    vpc_subnet: dummy-subnet
    tags: good
    custom_cpu: 8
    custom_memory: 16
    boot_disk_size: 400
    public_ip: 1.2.3.4
`
			command, _ := createCommand("us-west1", configStrTemplate)
			_, _, err := command.CreateVM()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("gcp_service_account or gcp_service_account_name must be set"))
		})

		It("does not set private IP address when not defined", func() {
			const configStrTemplate = `
opsman-configuration:
  gcp:
    version: 1.2.3
    gcp_service_account: something
    project: dummy-project
    region: %s
    zone: us-west1-c
    vm_name: opsman-vm
    vpc_network: dummy-network
    vpc_subnet: dummy-subnet
    tags: good
    custom_cpu: 8
    custom_memory: 16
    boot_disk_size: 400
    public_ip: 1.2.3.4
`
			command, runner := createCommand("us-west1", configStrTemplate)
			_, _, err := command.CreateVM()
			Expect(err).ToNot(HaveOccurred())

			invokes := runner.Invocations()["Execute"]
			Expect(invokes).ToNot(HaveLen(0))
			for _, args := range invokes {
				Expect(args[0]).ToNot(ContainElement(ContainSubstring("private-network-ip=")))
			}
		})

		It("does not set public IP address when not defined", func() {
			const configStrTemplate = `
opsman-configuration:
  gcp:
    version: 1.2.3
    gcp_service_account: something
    project: dummy-project
    region: %s
    zone: us-west1-c
    vm_name: opsman-vm
    vpc_network: dummy-network
    vpc_subnet: dummy-subnet
    tags: good
    custom_cpu: 8
    custom_memory: 16
    boot_disk_size: 400
    private_ip: 1.2.3.4
`
			command, runner := createCommand("us-west1", configStrTemplate)
			_, _, err := command.CreateVM()
			Expect(err).ToNot(HaveOccurred())

			invokes := runner.Invocations()["Execute"]
			Expect(invokes).ToNot(HaveLen(0))
			for _, args := range invokes {
				Expect(args[0]).ToNot(ContainElement(ContainSubstring("address=")))
			}

			By("setting the no-address so gcloud does not automatically create one")
			Expect(invokes[5][0]).To(ContainElement(ContainSubstring("no-address")))
		})

		It("does not set tags when not defined", func() {
			const configStrTemplate = `
opsman-configuration:
  gcp:
    version: 1.2.3
    gcp_service_account: something
    project: dummy-project
    region: %s
    zone: us-west1-c
    vm_name: opsman-vm
    vpc_network: dummy-network
    vpc_subnet: dummy-subnet
    custom_cpu: 8
    custom_memory: 16
    boot_disk_size: 400
    private_ip: 1.2.3.4
    public_ip: 1.2.3.4
`
			command, runner := createCommand("us-west1", configStrTemplate)
			_, _, err := command.CreateVM()
			Expect(err).ToNot(HaveOccurred())

			invokes := runner.Invocations()["Execute"]
			Expect(invokes).ToNot(HaveLen(0))
			for _, args := range invokes {
				Expect(args[0]).ToNot(ContainElement(ContainSubstring("--tags")))
			}
		})

		Context("with a valid config", func() {
			const configStrTemplate = `
opsman-configuration:
 gcp:
   version: 1.2.3
   gcp_service_account: something
   project: dummy-project
   region: %s
   zone: us-west1-c
   vm_name: opsman-vm
   vpc_network: dummy-network
   vpc_subnet: dummy-subnet
   tags: good
   custom_cpu: 8
   custom_memory: 16
   boot_disk_size: 400
   public_ip: 1.2.3.4
   private_ip: 10.0.0.2
   scopes: ["my-custom-scope1", "my-custom-scope-2"]
`
			When("specifying gcp_service_account_name instead of gcp_service_account", func() {
				It("it does not authenticate", func() {
					gcpServiceAccountConfigStrTemplate := `
opsman-configuration:
  gcp:
    version: 1.2.3
    gcp_service_account_name: something@something.com
    project: dummy-project
    region: %s
    zone: us-west1-c
    vm_name: opsman-vm
    vpc_network: dummy-network
    vpc_subnet: dummy-subnet
    tags: good
    custom_cpu: 8
    custom_memory: 16
    boot_disk_size: 400
    public_ip: 1.2.3.4
    private_ip: 10.0.0.2
`
					command, runner := createCommand("us-west1", gcpServiceAccountConfigStrTemplate)
					status, stateInfo, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					Expect(runner.ExecuteArgsForCall(0)).To(Equal([]interface{}{"iam", "service-accounts", "describe", "something@something.com"}))
					Expect(runner.ExecuteArgsForCall(1)).To(Equal([]interface{}{"config", "set", "project", "dummy-project"}))
					Expect(runner.ExecuteArgsForCall(2)).To(Equal([]interface{}{"config", "set", "compute/region", "us-west1"}))
					Expect(runner.ExecuteArgsForCall(3)).To(Equal([]interface{}{"compute", "images", "delete", "opsman-vm-image", "--quiet"}))
					Expect(runner.ExecuteArgsForCall(4)).To(Equal([]interface{}{"compute", "images", "create", "opsman-vm-image",
						"--source-uri=https://storage.googleapis.com/ops-manager-us-uri.tar.gz"}))
					Expect(runner.ExecuteArgsForCall(5)).To(Equal([]interface{}{"compute", "instances", "create", "opsman-vm",
						"--zone", "us-west1-c",
						"--image", "opsman-vm-image",
						"--custom-cpu", "8",
						"--custom-memory", "16",
						"--boot-disk-size", "400",
						"--network-interface", "subnet=dummy-subnet,address=1.2.3.4,private-network-ip=10.0.0.2",
						"--tags", "good",
						"--service-account", "something@something.com"}))

					Expect(status).To(Equal(vmmanagers.Success))
					Expect(stateInfo).To(Equal(vmmanagers.StateInfo{IAAS: "gcp", ID: "opsman-vm"}))
				})
			})

			When("specifying gcp_service_account_name does not exist", func() {
				It("it fails fast before creating the vm", func() {
					gcpServiceAccountConfigStrTemplate := `
opsman-configuration:
  gcp:
    version: 1.2.3
    gcp_service_account_name: bad@account.sad
    project: dummy-project
    region: %s
    zone: us-west1-c
    vm_name: opsman-vm
    vpc_network: dummy-network
    vpc_subnet: dummy-subnet
    tags: good
    custom_cpu: 8
    custom_memory: 16
    boot_disk_size: 400
    public_ip: 1.2.3.4
    private_ip: 10.0.0.2
`
					command, runner := createCommand("us-west1", gcpServiceAccountConfigStrTemplate)
					runner.ExecuteReturnsOnCall(0, nil, bytes.NewBufferString(""), errors.New("service account not there"))

					status, _, err := command.CreateVM()
					Expect(err).To(HaveOccurred())
					Expect(runner.ExecuteArgsForCall(0)).To(Equal([]interface{}{"iam", "service-accounts", "describe", "bad@account.sad"}))
					Expect(status).To(Equal(vmmanagers.Unknown))
					Expect(err.Error()).To(ContainSubstring("service_account_name error"))
				})
			})

			When("specifying ssh_public_key", func() {
				It("calls gcloud with the metadata flag", func() {
					sshKeyConfigStrTemplate := `
opsman-configuration:
  gcp:
    version: 1.2.3
    gcp_service_account: something
    project: dummy-project
    region: %s
    zone: us-west1-c
    vm_name: opsman-vm
    vpc_network: dummy-network
    vpc_subnet: dummy-subnet
    tags: good
    custom_cpu: 8
    custom_memory: 16
    boot_disk_size: 400
    public_ip: 1.2.3.4
    private_ip: 10.0.0.2
    scopes: ["my-custom-scope1", "my-custom-scope-2"]
    ssh_public_key: ssh-rsa abcd
    hostname: custom.domain.name
`
					command, runner := createCommand("us-west1", sshKeyConfigStrTemplate)
					status, stateInfo, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					Expect(runner.ExecuteArgsForCall(0)).To(matchers.OrderedConsistOf("auth", "activate-service-account", "--key-file", MatchRegexp(".*key.yaml.*")))
					Expect(runner.ExecuteArgsForCall(1)).To(Equal([]interface{}{"config", "set", "project", "dummy-project"}))
					Expect(runner.ExecuteArgsForCall(2)).To(Equal([]interface{}{"config", "set", "compute/region", "us-west1"}))
					Expect(runner.ExecuteArgsForCall(3)).To(Equal([]interface{}{"compute", "images", "delete", "opsman-vm-image", "--quiet"}))
					Expect(runner.ExecuteArgsForCall(4)).To(Equal([]interface{}{"compute", "images", "create", "opsman-vm-image",
						"--source-uri=https://storage.googleapis.com/ops-manager-us-uri.tar.gz"}))
					Expect(runner.ExecuteArgsForCall(5)).To(Equal([]interface{}{"compute", "instances", "create", "opsman-vm",
						"--zone", "us-west1-c",
						"--image", "opsman-vm-image",
						"--custom-cpu", "8",
						"--custom-memory", "16",
						"--boot-disk-size", "400",
						"--network-interface", "subnet=dummy-subnet,address=1.2.3.4,private-network-ip=10.0.0.2",
						"--tags", "good",
						"--scopes", "my-custom-scope1,my-custom-scope-2",
						"--metadata", "ssh-keys=ubuntu:ssh-rsa abcd,block-project-ssh-keys=TRUE",
						"--hostname", "custom.domain.name",
					}))

					Expect(status).To(Equal(vmmanagers.Success))
					Expect(stateInfo).To(Equal(vmmanagers.StateInfo{IAAS: "gcp", ID: "opsman-vm"}))
				})
			})

			When("the gcloud reports the vm already exists", func() {
				It("returns that it exists", func() {
					command, runner := createCommand("us-west1", configStrTemplate)
					runner.ExecuteReturnsOnCall(5, nil, bytes.NewBufferString("already exists"), errors.New(""))
					status, stateInfo, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())

					Expect(status).To(Equal(vmmanagers.Exist))
					Expect(stateInfo).To(Equal(vmmanagers.StateInfo{IAAS: "gcp", ID: "opsman-vm"}))
				})
			})

			DescribeTable("always uses the globally-accessible image listed for the us region",
				func(configRegion string, uriFromProductFile string) {
					command, runner := createCommand(configRegion, configStrTemplate)

					status, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal(vmmanagers.Success))

					Expect(runner.ExecuteArgsForCall(4)).To(Equal([]interface{}{"compute", "images", "create", "opsman-vm-image",
						"--source-uri=https://storage.googleapis.com/" + uriFromProductFile}))
				},
				Entry("parses asia-east1", "asia-east1", "ops-manager-us-uri.tar.gz"),
				Entry("parses asia-northeast1", "asia-northeast1", "ops-manager-us-uri.tar.gz"),
				Entry("parses asia-south1", "asia-south1", "ops-manager-us-uri.tar.gz"),
				Entry("parses asia-southeast1", "asia-southeast1", "ops-manager-us-uri.tar.gz"),
				Entry("parses australia-southeast1", "australia-southeast1", "ops-manager-us-uri.tar.gz"),
				Entry("parses europe-north1", "europe-north1", "ops-manager-us-uri.tar.gz"),
				Entry("parses europe-west1", "europe-west1", "ops-manager-us-uri.tar.gz"),
				Entry("parses europe-west2", "europe-west2", "ops-manager-us-uri.tar.gz"),
				Entry("parses europe-west3", "europe-west3", "ops-manager-us-uri.tar.gz"),
				Entry("parses europe-west4", "europe-west4", "ops-manager-us-uri.tar.gz"),
				Entry("parses northamerica-northeast1", "northamerica-northeast1", "ops-manager-us-uri.tar.gz"),
				Entry("parses southamerica-east1", "southamerica-east1", "ops-manager-us-uri.tar.gz"),
				Entry("parses us-central1", "us-central1", "ops-manager-us-uri.tar.gz"),
				Entry("parses us-east1", "us-east1", "ops-manager-us-uri.tar.gz"),
				Entry("parses us-east4", "us-east4", "ops-manager-us-uri.tar.gz"),
				Entry("parses us-west1", "us-west1", "ops-manager-us-uri.tar.gz"),
			)

			When("vm already exists in the state file", func() {
				It("returns exist status, doesn't make additional CLI calls, and exits 0", func() {
					command, runner := createCommand("us-west1", configStrTemplate)
					command.State.ID = "vm-name"
					runner.ExecuteReturnsOnCall(0, bytes.NewBufferString(`TERMINATE`), nil, nil)

					status, state, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal(vmmanagers.Exist))

					Expect(runner.ExecuteCallCount()).To(BeEquivalentTo(4))

					actualArgs := runner.ExecuteArgsForCall(3)
					Expect(actualArgs).To(Equal([]interface{}{
						`compute`, `instances`, `describe`, `vm-name`,
						`--zone`, `us-west1-c`,
						`--format`, `value(status)`,
					}))

					Expect(state.IAAS).To(Equal("gcp"))
					Expect(state.ID).To(Equal("vm-name"))
				})

				Describe("failure cases", func() {
					When("vm is not exist", func() {
						It("returns an error", func() {
							command, runner := createCommand("us-west1", configStrTemplate)
							command.State.ID = "invalid-id"
							runner.ExecuteReturnsOnCall(3, nil, bytes.NewBufferString(""), errors.New("vm does not exist"))

							status, _, err := command.CreateVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("error: vm does not exist\n       Could not find VM with ID \"invalid-id\".\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile."))
							Expect(status).To(Equal(vmmanagers.Unknown))

							Expect(runner.ExecuteCallCount()).To(BeEquivalentTo(4))
						})
					})
				})
			})

			Describe("failure cases", func() {
				When("external tools fail", func() {
					It("prints errors from gcloud", func() {
						command, runner := createCommand("us-west1", configStrTemplate)
						runner.ExecuteReturns(nil, nil, errors.New("some error occurred"))

						status, _, err := command.CreateVM()
						Expect(status).To(Equal(vmmanagers.Unknown))
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("gcloud error "))
					})
				})
				When("the state file has an invalid IAAS", func() {
					badState := vmmanagers.StateInfo{
						IAAS: "azure",
					}

					It("prints error", func() {
						command, runner := createCommand("us-west1", configStrTemplate)

						runner.ExecuteReturns(nil, nil, errors.New("some error occurred"))

						command.State = badState
						status, _, err := command.CreateVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("authentication file provided is for gcp, while the state file is for "))
						Expect(status).To(Equal(vmmanagers.Unknown))
					})
				})

				When("the image file is not valid YAML", func() {
					It("returns that the yaml is invalid", func() {
						command, _ := createCommand("us-west1", configStrTemplate)

						invalidUriFile, err := ioutil.TempFile("", "some*.yaml")
						Expect(err).ToNot(HaveOccurred())
						_, _ = invalidUriFile.WriteString("not valid yaml")
						Expect(invalidUriFile.Close()).ToNot(HaveOccurred())

						command.ImageYaml = invalidUriFile.Name()

						_, _, err = command.CreateVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("could not marshal image file"))
					})
				})

				When("image file is not a yaml file", func() {
					var command *vmmanagers.GCPVMManager
					BeforeEach(func() {
						command, _ = createCommand("us-west1", configStrTemplate)
						pdfFile := writePDFFile("never-gonna-give-you-up")
						command.ImageYaml = pdfFile
					})

					It("returns an error saying it cannot read the file", func() {
						_, _, err := command.CreateVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("ensure provided file " + command.ImageYaml + " is a .yml file"))
					})
				})

				When("the image file does not exist", func() {
					It("fails when the image file does not exist", func() {
						command, _ := createCommand("us-west1", configStrTemplate)

						command.ImageYaml = "does-not-exist.yml"
						_, _, err := command.CreateVM()
						Expect(err).To(HaveOccurred())
					})
				})
			})
		})

		DescribeTable("errors when required params are missing", func(param string) {
			invalidConfig := &vmmanagers.OpsmanConfigFilePayload{OpsmanConfig: vmmanagers.OpsmanConfig{
				GCP: &vmmanagers.GCPConfig{},
			}}

			gcpRunner := vmmanagers.NewGcpVMManager(invalidConfig, "", vmmanagers.StateInfo{}, nil)
			_, _, err := gcpRunner.CreateVM()
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Field validation for '%s' failed on the 'required' tag", param)))
		},
			Entry("requires Project", "Project"),
			Entry("requires Region", "Region"),
			Entry("requires Zone", "Zone"),
			Entry("requires VpcSubnet", "VpcSubnet"),
		)

		DescribeTable("defaulting any missing optional params", func(defaultParam string) {
			configStr := `
opsman-configuration:
  gcp:
    version: 1.2.3
    gcp_service_account: something
    project: someproject
    region: %s
    zone: us-west1-c
    vpc_network: opman-net
    vpc_subnet: infra
    tags: good
    public_ip: 1.2.3.4
    private_ip: 10.0.0.2
`
			command, runner := createCommand("us-west1", configStr)

			_, _, err := command.CreateVM()
			Expect(err).ToNot(HaveOccurred())

			Expect(runner.ExecuteArgsForCall(5)).To(ContainElement(defaultParam))

		},
			Entry("defaults the vm name", "ops-manager-vm"),
			Entry("defaults the cpu", "2"),
			Entry("defaults the memory", "8"),
			Entry("defaults the boot disk size", "100"))
	})

	Context("delete vm", func() {
		const configStrTemplate = `
opsman-configuration:
 gcp:
   project: dummy-project
   gcp_service_account: some-key-id
   region: %s
   zone: us-west1-b
`
		Context("with a valid config", func() {
			It("calls gcloud with correct cli arguments", func() {
				command, runner := createCommand("us-west1", configStrTemplate)
				command.State.ID = "vm-name"

				err := command.DeleteVM()
				Expect(err).ToNot(HaveOccurred())

				Expect(runner.ExecuteArgsForCall(0)).To(matchers.OrderedConsistOf("auth", "activate-service-account", "--key-file", MatchRegexp(".*key.yaml.*")))
				Expect(runner.ExecuteArgsForCall(1)).To(Equal([]interface{}{"config", "set", "project", "dummy-project"}))
				Expect(runner.ExecuteArgsForCall(2)).To(Equal([]interface{}{"config", "set", "compute/region", "us-west1"}))
				Expect(runner.ExecuteArgsForCall(3)).To(Equal([]interface{}{"compute", "instances", "describe", "vm-name", "--zone", "us-west1-b", "--format", "value(status)"}))
				Expect(runner.ExecuteArgsForCall(4)).To(Equal([]interface{}{"compute", "instances", "delete", "vm-name", "--zone", "us-west1-b", "--quiet"}))
				Expect(runner.ExecuteArgsForCall(5)).To(Equal([]interface{}{"compute", "images", "delete", "vm-name-image", "--quiet"}))

			})

			When("the image does not exist", func() {
				It("does not return an error", func() {
					command, runner := createCommand("us-west1", configStrTemplate)
					runner.ExecuteReturnsOnCall(5, nil, bytes.NewBufferString("The resource 'some-image' was not found"), errors.New("exit status 1"))
					command.State.ID = "vm-name"

					err := command.DeleteVM()
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Describe("failure cases", func() {
				When("external tools fail", func() {
					It("prints errors from gcloud", func() {
						command, runner := createCommand("us-west1", configStrTemplate)

						for i := 1; i <= 4; i++ {
							runner.ExecuteReturnsOnCall(i, nil, nil, errors.New("some error occurred"))
							err := command.DeleteVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("gcloud error "))
						}
					})
				})

				When("vm specified in the state file does not exist", func() {
					It("returns an error", func() {
						command, runner := createCommand("us-west1", configStrTemplate)
						runner.ExecuteReturnsOnCall(3, nil, nil, errors.New("vm does not exist"))

						command.State = vmmanagers.StateInfo{
							IAAS: "gcp",
							ID:   "invalid-id",
						}

						err := command.DeleteVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("error: vm does not exist\n       Could not find VM with ID \"invalid-id\".\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile."))
					})
				})
			})
		})

		DescribeTable("errors when required params are missing", func(param string) {
			invalidConfig := &vmmanagers.OpsmanConfigFilePayload{OpsmanConfig: vmmanagers.OpsmanConfig{
				GCP: &vmmanagers.GCPConfig{},
			}}

			command := vmmanagers.NewGcpVMManager(invalidConfig, "", vmmanagers.StateInfo{}, nil)

			err := command.DeleteVM()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Field validation for '%s' failed on the 'required' tag", param)))
		},
			Entry("requires Project", "Project"),
			Entry("requires Region", "Region"),
		)

		Context("with an invalid iaas", func() {
			state := vmmanagers.StateInfo{
				IAAS: "aws",
			}

			It("prints error", func() {
				command, _ := createCommand("us-west1", configStrTemplate)

				command.State = state
				err := command.DeleteVM()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("authentication file provided is for gcp, while the state file is for "))
			})
		})
	})

	testIAASForPropertiesInExampleFile("GCP")
})
