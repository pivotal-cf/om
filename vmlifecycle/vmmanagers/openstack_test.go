package vmmanagers_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/om/vmlifecycle/matchers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers/fakes"
)

var _ = Describe("Openstack VMManager", func() {
	createCommand := func(configStrTemplate string) (*vmmanagers.OpenstackVMManager, *fakes.OpenstackRunner) {
		var err error
		runner := &fakes.OpenstackRunner{}
		testUriFile, err := ioutil.TempFile("", "")
		Expect(err).ToNot(HaveOccurred())

		state := vmmanagers.StateInfo{
			IAAS: "openstack",
		}

		var validConfig *vmmanagers.OpsmanConfigFilePayload
		err = yaml.UnmarshalStrict([]byte(configStrTemplate), &validConfig)
		Expect(err).ToNot(HaveOccurred())

		command := vmmanagers.NewOpenstackVMManager(validConfig, testUriFile.Name(), state, runner)
		return command, runner
	}

	BeforeEach(func() {
		var err error
		Expect(err).ToNot(HaveOccurred())
	})

	Context("create vm", func() {
		Context("with a valid config", func() {
			const configStrTemplate = `
opsman-configuration:
  openstack:
    auth_url: https://example.com:5000/v2.0
    project_name: marker
    project_domain_name: default
    user_domain_name: default
    net_id: 590790ef-f90f-4bfd-884f-cb6ece199a82
    username: admin
    password: password
    key_pair_name: marker-keypair
    security_group_name: marker-sec-group
    vm_name: awesome-vm
    public_ip: 10.10.10.9
    private_ip: 10.0.0.3
    flavor: m1.large
    insecure: true
    availability_zone: zone-01
`

			When("private IP is not provided", func() {
				It("does not attach it to the server instance", func() {
					command, runner := createCommand(configStrTemplate)
					command.Config.PrivateIP = ""

					runner.ExecuteReturnsOnCall(0, bytes.NewBufferString("\r\n"), nil, nil)
					runner.ExecuteReturnsOnCall(1, bytes.NewBufferString("custom-image-id\r\n"), nil, nil)
					runner.ExecuteReturnsOnCall(2, bytes.NewBufferString("custom-server-id\r\n"), bytes.NewBufferString("TestStatus: pending creation"), nil)

					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					invokes := runner.Invocations()["Execute"]
					Expect(invokes).ToNot(HaveLen(0))
					for _, args := range invokes {
						Expect(args[0]).ToNot(ContainElement(ContainSubstring("fixed-ip")))
					}
				})
			})

			When("public IP is not provided", func() {
				It("does not attach it to the server instance", func() {
					command, runner := createCommand(configStrTemplate)
					command.Config.PublicIP = ""

					runner.ExecuteReturnsOnCall(0, bytes.NewBufferString("\r\n"), nil, nil)
					runner.ExecuteReturnsOnCall(1, bytes.NewBufferString("custom-image-id\r\n"), nil, nil)
					runner.ExecuteReturnsOnCall(2, bytes.NewBufferString("custom-server-id\r\n"), bytes.NewBufferString("TestStatus: pending creation"), nil)

					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					invokes := runner.Invocations()["Execute"]
					Expect(invokes).ToNot(HaveLen(0))
					for _, args := range invokes {
						Expect(args[0]).ToNot(ContainElement(ContainSubstring("floating")))
					}
				})
			})

			When("vm already exists in the state file", func() {
				It("returns exist status, doesn't make additional CLI calls, and exits 0", func() {
					command, runner := createCommand(configStrTemplate)
					command.State = vmmanagers.StateInfo{
						IAAS: "openstack",
						ID:   "vm-name",
					}
					runner.ExecuteReturnsOnCall(0, bytes.NewBufferString("ACTIVE\r\n"), nil, nil)

					status, state, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal(vmmanagers.Exist))

					Expect(runner.ExecuteCallCount()).To(BeEquivalentTo(1))

					actualArgs := runner.ExecuteArgsForCall(0)
					Expect(actualArgs).To(matchers.OrderedConsistOf([]interface{}{
						"--os-username", gstruct.Ignore(),
						"--os-password", gstruct.Ignore(),
						"--os-auth-url", "https://example.com:5000/v2.0",
						"--os-project-name", "marker",
						"--insecure",
						"--os-project-domain-name", "default",
						"--os-user-domain-name", "default",
						"--os-identity-api-version", "3",
						"server", "show", "vm-name",
						"--column", "status",
						"--format", "value",
					}))

					Expect(state.IAAS).To(Equal("openstack"))
					Expect(state.ID).To(Equal("vm-name"))
				})

				Describe("failure cases", func() {
					When("vm is not exist", func() {
						It("returns an error", func() {
							command, runner := createCommand(configStrTemplate)
							command.State.ID = "vm-name"
							runner.ExecuteReturnsOnCall(0, nil, bytes.NewBufferString(""), errors.New("some error"))

							status, _, err := command.CreateVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(MatchRegexp("VM ID in statefile does not exist. Please check your statefile and try again"))
							Expect(status).To(Equal(vmmanagers.Unknown))

							Expect(runner.ExecuteCallCount()).To(BeEquivalentTo(1))
						})
					})
				})
			})

			When("the image does not exist", func() {
				It("calls openstack with correct cli arguments", func() {
					command, runner := createCommand(configStrTemplate)

					runner.ExecuteReturnsOnCall(0, bytes.NewBufferString("\r\n"), nil, nil)
					runner.ExecuteReturnsOnCall(1, bytes.NewBufferString("custom-image-id\r\n"), nil, nil)
					runner.ExecuteReturnsOnCall(2, bytes.NewBufferString("custom-server-id\r\n"), bytes.NewBufferString("TestStatus: pending creation"), nil)

					commands := [][]interface{}{
						{ //1
							"--os-username", gstruct.Ignore(),
							"--os-password", gstruct.Ignore(),
							"--os-auth-url", "https://example.com:5000/v2.0",
							"--os-project-name", "marker",
							"--insecure",
							"--os-project-domain-name", "default",
							"--os-user-domain-name", "default",
							"--os-identity-api-version", "3",
							"image", "list",
							"--name", "awesome-vm-image",
							"--format", "value",
							"--column", "ID",
						},
						{ //2
							"--os-username", gstruct.Ignore(),
							"--os-password", gstruct.Ignore(),
							"--os-auth-url", "https://example.com:5000/v2.0",
							"--os-project-name", "marker",
							"--insecure",
							"--os-project-domain-name", "default",
							"--os-user-domain-name", "default",
							"--os-identity-api-version", "3",
							"image", "create",
							"--format", "value", "--column", "id",
							"--file", MatchRegexp(".*"), "awesome-vm-image",
						},
						{ //3
							"--os-username", gstruct.Ignore(),
							"--os-password", gstruct.Ignore(),
							"--os-auth-url", "https://example.com:5000/v2.0",
							"--os-project-name", "marker",
							"--insecure",
							"--os-project-domain-name", "default",
							"--os-user-domain-name", "default",
							"--os-identity-api-version", "3",
							"server", "create",
							"--flavor", "m1.large", "--image", "custom-image-id",
							"--nic", "net-id=590790ef-f90f-4bfd-884f-cb6ece199a82,v4-fixed-ip=10.0.0.3",
							"--security-group", "marker-sec-group",
							"--key-name", "marker-keypair",
							"--format", "value", "--column", "id",
							"--wait",
							"--availability-zone", "zone-01",
							"awesome-vm",
						},
						{ //4
							"--os-username", gstruct.Ignore(),
							"--os-password", gstruct.Ignore(),
							"--os-auth-url", "https://example.com:5000/v2.0",
							"--os-project-name", "marker",
							"--insecure",
							"--os-project-domain-name", "default",
							"--os-user-domain-name", "default",
							"--os-identity-api-version", "3",
							"server", "add", "floating", "ip",
							"custom-server-id", "10.10.10.9",
						},
					}
					status, stateInfo, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal(vmmanagers.Success))
					Expect(stateInfo).To(Equal(vmmanagers.StateInfo{IAAS: "openstack", ID: "custom-server-id"}))

					for i, expectedArgs := range commands {
						actualArgs := runner.ExecuteArgsForCall(i)
						Expect(actualArgs).To(matchers.OrderedConsistOf(expectedArgs))
					}
				})

				Describe("failure cases", func() {
					When("the state file has an invalid IAAS", func() {
						badState := vmmanagers.StateInfo{
							IAAS: "azure",
						}

						It("prints error", func() {
							command, _ := createCommand(configStrTemplate)
							command.State = badState
							status, _, err := command.CreateVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("authentication file provided is for openstack, while the state file is for "))
							Expect(status).To(Equal(vmmanagers.Unknown))
						})
					})

					When("external tools fail", func() {
						DescribeTable("prints errors from openstack", func(callNumber int, expectedStatus vmmanagers.Status) {
							command, runner := createCommand(configStrTemplate)

							runner.ExecuteReturns(bytes.NewBufferString("null\r\n"), nil, nil)

							runner.ExecuteReturnsOnCall(callNumber, nil, nil, errors.New("some error occurred"))
							status, _, err := command.CreateVM()
							Expect(status).To(Equal(expectedStatus))
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("openstack error "))
						},
							Entry("listing previous images", 1, vmmanagers.Unknown),
							Entry("create image", 2, vmmanagers.Unknown),
							Entry("create vm", 3, vmmanagers.Unknown),
							Entry("attach IP address", 4, vmmanagers.Incomplete),
						)
					})

					When("the image file does not exist", func() {
						It("fails when the image file does not exist", func() {
							command, _ := createCommand(configStrTemplate)
							command.Image = "does-not-exist.raw"

							_, _, err := command.CreateVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("does-not-exist.raw: no such file or directory"))
						})
					})
				})
			})

			When("the image already exists", func() {
				It("deletes the image", func() {
					command, runner := createCommand(configStrTemplate)

					runner.ExecuteReturnsOnCall(0, bytes.NewBufferString("image-id\r\n"), nil, nil)
					runner.ExecuteReturnsOnCall(2, bytes.NewBufferString("custom-image-id\r\n"), nil, nil)
					runner.ExecuteReturnsOnCall(3, bytes.NewBufferString("custom-server-id\r\n"), bytes.NewBufferString("TestStatus: pending creation"), nil)

					status, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal(vmmanagers.Success))

					Expect(runner.ExecuteArgsForCall(1)).To(matchers.OrderedConsistOf(
						"--os-username", gstruct.Ignore(),
						"--os-password", gstruct.Ignore(),
						"--os-auth-url", "https://example.com:5000/v2.0",
						"--os-project-name", "marker",
						"--insecure",
						"--os-project-domain-name", "default",
						"--os-user-domain-name", "default",
						"--os-identity-api-version", "3",
						"image", "delete", "awesome-vm-image",
					))
				})

				It("displays an error the delete fails", func() {
					command, runner := createCommand(configStrTemplate)

					runner.ExecuteReturnsOnCall(0, bytes.NewBufferString("DOWN\r\n"), nil, nil)
					runner.ExecuteReturnsOnCall(1, bytes.NewBufferString("image-id\r\n"), nil, nil)
					runner.ExecuteReturnsOnCall(2, nil, nil, errors.New("error occurred"))

					status, _, err := command.CreateVM()
					Expect(err).To(HaveOccurred())
					Expect(status).To(Equal(vmmanagers.Unknown))
				})
			})

		})

		DescribeTable("errors when required params are missing", func(param string) {
			command, _ := createCommand("{ opsman-configuration: { openstack: {} }}")
			_, _, err := command.CreateVM()
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Field validation for '%s' failed on the 'required' tag", param)))
		},
			Entry("requires Username", "Username"),
			Entry("requires Password", "Password"),
			Entry("requires AuthUrl", "AuthUrl"),
			Entry("requires Project", "Project"),
			Entry("requires NetID", "NetID"),
			Entry("requires SecurityGroup", "SecurityGroup"),
			Entry("requires KeyName", "KeyName"),
		)

		It("requires at least public IP or private IP to be set", func() {
			configWithoutIPs := `
opsman-configuration:
  openstack:
    auth_url: https://example.com:5000/v2.0
    project_name: marker
    net_id: 590790ef-f90f-4bfd-884f-cb6ece199a82
    username: admin
    password: password
    key_pair_name: marker-keypair
    security_group_name: marker-sec-group
    vm_name: awesome-vm
    flavor: m1.large
`
			command, _ := createCommand(configWithoutIPs)
			_, _, err := command.CreateVM()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("PublicIP and/or PrivateIP must be set"))
		})

		It("defaulting any missing optional params", func() {
			configStr := `
opsman-configuration:
  openstack:
    auth_url: https://example.com:5000/v2.0
    project_name: marker
    net_id: 590790ef-f90f-4bfd-884f-cb6ece199a82
    username: admin
    password: password
    key_pair_name: marker-keypair
    security_group_name: marker-sec-group
    public_ip: 10.10.10.9
`

			command, runner := createCommand(configStr)

			runner.ExecuteReturnsOnCall(0, bytes.NewBufferString("custom-image-id\r\n"), nil, nil)
			runner.ExecuteReturnsOnCall(1, bytes.NewBufferString("custom-server-id\r\n"), bytes.NewBufferString("TestStatus: pending creation"), nil)

			_, _, err := command.CreateVM()
			Expect(err).ToNot(HaveOccurred())

			args := runner.ExecuteArgsForCall(3)
			Expect(args).To(ContainElement("ops-manager-vm"))
			Expect(args).To(ContainElement("m1.xlarge"))
			Expect(args).To(ContainElement("3"))
		})
	})

	Context("delete vm", func() {
		const configStrTemplate = `
opsman-configuration:
  openstack:
    auth_url: https://example.com:5000/v2.0
    project_name: marker
    username: admin
    password: password
`

		Context("with a valid config", func() {
			Describe("DeleteVM", func() {
				It("calls openstack with correct cli arguments", func() {
					command, runner := createCommand(configStrTemplate)
					command.State.ID = "some-random-id"

					runner.ExecuteReturnsOnCall(0, bytes.NewBufferString(`{"image":"testing-opsman-image (7bf1ac30-290f-42ae-bc5e-b240ef051fbf)"}`), nil, nil)

					err := command.DeleteVM()
					Expect(err).ToNot(HaveOccurred())

					Expect(runner.ExecuteArgsForCall(0)).To(matchers.OrderedConsistOf(
						"--os-username",
						gstruct.Ignore(),
						"--os-password",
						gstruct.Ignore(),
						"--os-auth-url",
						"https://example.com:5000/v2.0",
						"--os-project-name",
						"marker",
						"server",
						"show",
						"some-random-id",
						"--column",
						"image",
						"--format",
						"value",
					))

					Expect(runner.ExecuteArgsForCall(1)).To(matchers.OrderedConsistOf(
						"--os-username",
						gstruct.Ignore(),
						"--os-password",
						gstruct.Ignore(),
						"--os-auth-url",
						"https://example.com:5000/v2.0",
						"--os-project-name",
						"marker",
						"server",
						"delete",
						"some-random-id",
						"--wait",
					))

					Expect(runner.ExecuteArgsForCall(2)).To(matchers.OrderedConsistOf(
						"--os-username",
						gstruct.Ignore(),
						"--os-password",
						gstruct.Ignore(),
						"--os-auth-url",
						"https://example.com:5000/v2.0",
						"--os-project-name",
						"marker",
						"image",
						"delete",
						"7bf1ac30-290f-42ae-bc5e-b240ef051fbf",
					))
				})

				Describe("failure cases", func() {
					When("external tools fail", func() {
						It("prints errors from openstack", func() {
							command, runner := createCommand(configStrTemplate)

							runner.ExecuteReturns(nil, nil, errors.New("some error occurred"))
							err := command.DeleteVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("openstack error "))
						})
					})

					When("vm specified in the state file does not exist", func() {
						It("returns an error", func() {
							command, runner := createCommand(configStrTemplate)
							command.State.ID = "invalid-id"
							runner.ExecuteReturns(nil, nil, errors.New("vm does not exist"))

							command.State = vmmanagers.StateInfo{
								IAAS: "openstack",
								ID:   "invalid-id",
							}

							err := command.DeleteVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(Equal("openstack error deleting the vm: vm does not exist\n       Could not find VM with ID \"invalid-id\".\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile."))
						})
					})
				})
			})
		})

		DescribeTable("errors when required params are missing", func(param string) {
			invalidConfig := &vmmanagers.OpsmanConfigFilePayload{OpsmanConfig: vmmanagers.OpsmanConfig{Openstack: &vmmanagers.OpenstackConfig{}}}

			runner := vmmanagers.NewOpenstackVMManager(invalidConfig, "", vmmanagers.StateInfo{}, nil)

			err := runner.DeleteVM()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Field validation for '%s' failed on the 'required' tag", param)))
		},
			Entry("requires Username", "Username"),
			Entry("requires Password", "Password"),
			Entry("requires AuthUrl", "AuthUrl"),
			Entry("requires Project", "Project"),
		)

		Context("with an invalid iaas", func() {
			var state = vmmanagers.StateInfo{
				IAAS: "gcp",
			}

			It("prints error", func() {
				command, runner := createCommand(configStrTemplate)

				runner.ExecuteReturns(nil, nil, errors.New("some error occurred"))

				command.State = state
				err := command.DeleteVM()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("authentication file provided is for openstack, while the state file is for "))
			})
		})
	})

	testIAASForPropertiesInExampleFile("Openstack")
})
