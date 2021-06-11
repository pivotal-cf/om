package vmmanagers_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"archive/tar"

	"bytes"
	"errors"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/vmlifecycle/matchers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers/fakes"
	"gopkg.in/yaml.v2"
)

var _ = Describe("vSphere VMManager", func() {
	createCommand := func(configStr string, opsmanVersion string) (*vmmanagers.VsphereVMManager, *fakes.GovcRunner) {

		var validConfig *vmmanagers.OpsmanConfigFilePayload
		err := yaml.UnmarshalStrict([]byte(configStr), &validConfig)
		Expect(err).ToNot(HaveOccurred())

		state := vmmanagers.StateInfo{
			IAAS: "vsphere",
		}

		testOVA := createOVA(opsmanVersion)
		runner := &fakes.GovcRunner{}
		command := vmmanagers.NewVsphereVMManager(validConfig, testOVA.Name(), state, runner)
		return command, runner

	}

	opsmanVersionBelow26 := "2.5.2"
	Context("create vm", func() {
		Context("with a valid config", func() {
			configStr := `
opsman-configuration:
  vsphere:
    vcenter:
      url: vcenter.nowhere.nonexist
      username: goodman
      password: badguy
      datastore: datastore
      datacenter: datacenter
      insecure: 1
      ca_cert: ca
      resource_pool: resource-pool
      host: host
      folder: /datacenter/vm/folder
    disk_type: thin
    private_ip: 1.2.3.4
    dns: 1.1.1.1
    ntp: ntp.server.xyz
    ssh_password: password
    hostname: full.domain.name
    network: some-edge
    netmask: 255.255.255.192
    gateway: 2.2.2.2
    vm_name: vm_name
    disk_size: 200
    memory: 25
    cpu: 10
`

			Describe("CreateVM", func() {
				It("calls without error", func() {
					command, _ := createCommand(configStr, opsmanVersionBelow26)
					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
				})

				It("calls govc with correct cli arguments, and does not duplicate /datacenter/vm path", func() {
					command, runner := createCommand(configStr, opsmanVersionBelow26)

					status, state, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal(vmmanagers.Success))
					Expect(state).To(Equal(vmmanagers.StateInfo{IAAS: "vsphere", ID: "/datacenter/vm/folder/vm_name"}))

					_, args := runner.ExecuteWithEnvVarsArgsForCall(0)
					Expect(args).To(matchers.OrderedConsistOf(
						"import.ova",
						MatchRegexp("-options=.*options.json.*"),
						MatchRegexp(".*ova"),
					))

					_, args = runner.ExecuteWithEnvVarsArgsForCall(1)
					Expect(args).To(matchers.OrderedConsistOf(
						"vm.power",
						"-off=true",
						"-vm.ipath=/datacenter/vm/folder/vm_name",
					))
					_, args = runner.ExecuteWithEnvVarsArgsForCall(2)
					Expect(args).To(matchers.OrderedConsistOf(
						"vm.change",
						"-vm.ipath=/datacenter/vm/folder/vm_name",
						"-m=25600",
						"-c=10",
					))
					_, args = runner.ExecuteWithEnvVarsArgsForCall(3)
					Expect(args).To(matchers.OrderedConsistOf(
						"vm.power",
						"-on=true",
						"-vm.ipath=/datacenter/vm/folder/vm_name",
					))
				})

				When("setting custom cpu and memory", func() {
					const configTemplate = `
opsman-configuration:
  vsphere:
    vcenter:
      url: vcenter.nowhere.nonexist
      username: goodman
      password: badguy
      datastore: datastore
      datacenter: datacenter
      insecure: 1
      ca_cert: ca
      resource_pool: resource-pool
      host: host
      folder: /datacenter/vm/folder
    disk_type: thin
    private_ip: 1.2.3.4
    dns: 1.1.1.1
    ntp: ntp.server.xyz
    ssh_password: password
    hostname: full.domain.name
    network: some-edge
    netmask: 255.255.255.192
    gateway: 2.2.2.2
    vm_name: vm_name
    memory: %s
    cpu: %s
`
					It("sets cpu or memory if only one is not default", func() {
						memDefault := fmt.Sprintf(configTemplate, vmmanagers.DefaultMemory, "12")
						command, runner := createCommand(memDefault, opsmanVersionBelow26)
						_, _, _ = command.CreateVM()
						Expect(runner.ExecuteWithEnvVarsCallCount()).To(Equal(4))

						cpuDefault := fmt.Sprintf(configTemplate, "25", vmmanagers.DefaultCPU)
						command, runner = createCommand(cpuDefault, opsmanVersionBelow26)
						_, _, _ = command.CreateVM()
						Expect(runner.ExecuteWithEnvVarsCallCount()).To(Equal(4))
					})

					It("does not make extra calls if memory and cpu are default", func() {
						defaultConfigStr := `
opsman-configuration:
  vsphere:
    vcenter:
      url: vcenter.nowhere.nonexist
      username: goodman
      password: badguy
      datastore: datastore
      datacenter: datacenter
      insecure: 1
      ca_cert: ca
      resource_pool: resource-pool
      host: host
      folder: /datacenter/vm/folder
    disk_type: thin
    private_ip: 1.2.3.4
    dns: 1.1.1.1
    ntp: ntp.server.xyz
    ssh_password: password
    hostname: full.domain.name
    network: some-edge
    netmask: 255.255.255.192
    gateway: 2.2.2.2
    vm_name: vm_name
`
						command, runner := createCommand(defaultConfigStr, opsmanVersionBelow26)

						status, state, err := command.CreateVM()
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal(vmmanagers.Success))
						Expect(state).To(Equal(vmmanagers.StateInfo{IAAS: "vsphere", ID: "/datacenter/vm/folder/vm_name"}))

						Expect(runner.ExecuteWithEnvVarsCallCount()).To(Equal(1))
						_, args := runner.ExecuteWithEnvVarsArgsForCall(0)
						Expect(args).To(matchers.OrderedConsistOf(
							"import.ova",
							MatchRegexp("-options=.*options.json.*"),
							MatchRegexp(".*ova"),
						))
					})

					It("returns an error when the cpu/memory fails", func() {
						memDefault := fmt.Sprintf(configTemplate, vmmanagers.DefaultMemory, "12")
						command, runner := createCommand(memDefault, opsmanVersionBelow26)
						runner.ExecuteWithEnvVarsReturnsOnCall(1, nil, nil, errors.New("some error occurred"))
						status, stateInfo, err := command.CreateVM()

						Expect(err).To(HaveOccurred())
						Expect(status).To(Equal(vmmanagers.Incomplete))
						Expect(stateInfo.ID).To(Equal("/datacenter/vm/folder/vm_name"))
					})
				})


				When("setting custom disk_size", func() {
					const configTemplate = `
opsman-configuration:
  vsphere:
    vcenter:
      url: vcenter.nowhere.nonexist
      username: goodman
      password: badguy
      datastore: datastore
      datacenter: datacenter
      insecure: 1
      ca_cert: ca
      resource_pool: resource-pool
      host: host
      folder: /datacenter/vm/folder
    disk_type: thin
    private_ip: 1.2.3.4
    dns: 1.1.1.1
    ntp: ntp.server.xyz
    ssh_password: password
    hostname: full.domain.name
    network: some-edge
    netmask: 255.255.255.192
    gateway: 2.2.2.2
    vm_name: vm_name
    disk_size: %s
`
					It("sets disk_size if not default", func() {
						diskSizeDefault := fmt.Sprintf(configTemplate, "200")
						command, runner := createCommand(diskSizeDefault, opsmanVersionBelow26)
						_, _, _ = command.CreateVM()
						Expect(runner.ExecuteWithEnvVarsCallCount()).To(Equal(4))
					})

					It("does not make extra calls if disk_size is default", func() {
						defaultConfigStr := `
opsman-configuration:
  vsphere:
    vcenter:
      url: vcenter.nowhere.nonexist
      username: goodman
      password: badguy
      datastore: datastore
      datacenter: datacenter
      insecure: 1
      ca_cert: ca
      resource_pool: resource-pool
      host: host
      folder: /datacenter/vm/folder
    disk_type: thin
    private_ip: 1.2.3.4
    dns: 1.1.1.1
    ntp: ntp.server.xyz
    ssh_password: password
    hostname: full.domain.name
    network: some-edge
    netmask: 255.255.255.192
    gateway: 2.2.2.2
    vm_name: vm_name
`
						command, runner := createCommand(defaultConfigStr, opsmanVersionBelow26)

						status, state, err := command.CreateVM()
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal(vmmanagers.Success))
						Expect(state).To(Equal(vmmanagers.StateInfo{IAAS: "vsphere", ID: "/datacenter/vm/folder/vm_name"}))

						Expect(runner.ExecuteWithEnvVarsCallCount()).To(Equal(1))
						_, args := runner.ExecuteWithEnvVarsArgsForCall(0)
						Expect(args).To(matchers.OrderedConsistOf(
							"import.ova",
							MatchRegexp("-options=.*options.json.*"),
							MatchRegexp(".*ova"),
						))
					})

					It("returns an error when the disk_size fails", func() {
						memDefault := fmt.Sprintf(configTemplate, "120")
						command, runner := createCommand(memDefault, opsmanVersionBelow26)
						runner.ExecuteWithEnvVarsReturnsOnCall(1, nil, nil, errors.New("some error occurred"))
						status, stateInfo, err := command.CreateVM()

						Expect(err).To(HaveOccurred())
						Expect(status).To(Equal(vmmanagers.Incomplete))
						Expect(stateInfo.ID).To(Equal("/datacenter/vm/folder/vm_name"))
					})
				})

				It("calls govc with the correct environment variables", func() {
					command, runner := createCommand(configStr, opsmanVersionBelow26)

					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())

					env, _ := runner.ExecuteWithEnvVarsArgsForCall(0)
					Eventually(env).Should(ContainElement(`GOVC_URL=vcenter.nowhere.nonexist`))
					Eventually(env).Should(ContainElement(`GOVC_USERNAME=goodman`))
					Eventually(env).Should(ContainElement(`GOVC_PASSWORD=badguy`))
					Eventually(env).Should(ContainElement(`GOVC_DATASTORE=datastore`))
					Eventually(env).Should(ContainElement(`GOVC_DATACENTER=datacenter`))
					Eventually(env).Should(ContainElement(`GOVC_INSECURE=1`))
					Eventually(env).Should(ContainElement(`GOVC_NETWORK=some-edge`))
					Eventually(env).Should(ContainElement(`GOVC_RESOURCE_POOL=resource-pool`))
					Eventually(env).Should(ContainElement(`GOVC_HOST=host`))
					Eventually(env).Should(ContainElement(`GOVC_FOLDER=/datacenter/vm/folder`))
					Eventually(env).Should(ContainElement(`GOMAXPROCS=1`))
				})

				When("creation fails because the vm already exists ", func() {
					It("returns that it exists", func() {
						command, runner := createCommand(configStr, opsmanVersionBelow26)
						runner.ExecuteWithEnvVarsReturnsOnCall(0, nil, bytes.NewBufferString("already exists"), errors.New(""))
						status, stateInfo, err := command.CreateVM()
						Expect(err).ToNot(HaveOccurred())

						Expect(status).To(Equal(vmmanagers.Exist))
						Expect(stateInfo).To(Equal(vmmanagers.StateInfo{IAAS: "vsphere", ID: "/datacenter/vm/folder/vm_name"}))
					})
				})

				When("the vm specified in the state file already exists", func() {
					It("returns exist status, doesn't make additional CLI calls, and exits 0", func() {
						command, runner := createCommand(configStr, opsmanVersionBelow26)
						runner.ExecuteWithEnvVarsReturnsOnCall(0, nil, nil, nil)

						command.State = vmmanagers.StateInfo{
							IAAS: "vsphere",
							ID:   "some-ipath",
						}

						status, state, err := command.CreateVM()
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal(vmmanagers.Exist))

						Expect(runner.ExecuteWithEnvVarsCallCount()).To(BeEquivalentTo(1))

						_, actualArgs := runner.ExecuteWithEnvVarsArgsForCall(0)
						Expect(actualArgs).To(Equal([]interface{}{
							`vm.info`, "-vm.ipath=some-ipath",
						}))

						Expect(state.IAAS).To(Equal("vsphere"))
						Expect(state.ID).To(Equal("some-ipath"))
					})

					Describe("failure cases", func() {
						When("vm specified in the state file does not exist", func() {
							It("returns an error", func() {
								command, runner := createCommand(configStr, opsmanVersionBelow26)
								runner.ExecuteWithEnvVarsReturnsOnCall(0, nil, nil, errors.New("vm does not exist"))

								command.State = vmmanagers.StateInfo{
									IAAS: "vsphere",
									ID:   "some-vm-id",
								}

								status, _, err := command.CreateVM()
								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(Equal("error: vm does not exist\n       Could not find VM with ID \"some-vm-id\".\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile."))
								Expect(status).To(Equal(vmmanagers.Unknown))

								Expect(runner.ExecuteWithEnvVarsCallCount()).To(BeEquivalentTo(1))
							})
						})
					})
				})

				Context("with SSL certs defined", func() {
					It("set the environment variable pointing to a file", func() {
						command, runner := createCommand(configStr, opsmanVersionBelow26)

						_, _, err := command.CreateVM()
						Expect(err).ToNot(HaveOccurred())

						env, _ := runner.ExecuteWithEnvVarsArgsForCall(0)
						Eventually(env).Should(ContainElement(MatchRegexp(`GOVC_TLS_CA_CERTS=.*ca.crt.*`)))
					})
				})

				Context("failure cases", func() {
					When("the state file has an invalid IAAS", func() {
						It("prints error", func() {
							command, runner := createCommand(configStr, opsmanVersionBelow26)
							runner.ExecuteWithEnvVarsReturns(nil, nil, errors.New("some error occurred"))

							command.State = vmmanagers.StateInfo{
								IAAS: "gcp",
							}
							status, _, err := command.CreateVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("authentication file provided is for vsphere, while the state file is for "))
							Expect(status).To(Equal(vmmanagers.Unknown))
						})
					})
					When("external tools fail for creating a vm", func() {
						It("prints errors from govc", func() {
							command, runner := createCommand(configStr, opsmanVersionBelow26)

							runner.ExecuteWithEnvVarsReturns(nil, nil, errors.New("some error occurred"))

							status, _, err := command.CreateVM()
							Expect(status).To(Equal(vmmanagers.Unknown))
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("govc error: "))
						})
					})

					When("the image file is not valid", func() {
						var (
							invalidOVA  *os.File
							validConfig *vmmanagers.OpsmanConfigFilePayload
						)

						BeforeEach(func() {
							var err error
							invalidOVA, err = ioutil.TempFile("", "test.ova")
							Expect(err).ToNot(HaveOccurred())
							_, _ = invalidOVA.WriteString("some-string-that-makes-the-ova-invalid")
							Expect(invalidOVA.Close()).ToNot(HaveOccurred())

							configStrTemplate := `
opsman-configuration:
  vsphere:
    vcenter:
      url: vcenter.nowhere.nonexist
      username: goodman
      password: badguy
      datastore: datastore
      datacenter: datacenter
      resource_pool: resource-pool
`
							err = yaml.Unmarshal([]byte(configStrTemplate), &validConfig)
							Expect(err).ToNot(HaveOccurred())
						})
						AfterEach(func() {
							os.Remove(invalidOVA.Name())
						})

						It("fails on creation with an OVA validation error", func() {
							command, _ := createCommand(configStr, opsmanVersionBelow26)
							command.ImageOVA = invalidOVA.Name()

							_, _, err := command.CreateVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("could not validate image-file format"))
							Expect(err.Error()).To(ContainSubstring("Is your image an OVA file?"))
							Expect(err.Error()).To(ContainSubstring(invalidOVA.Name()))
						})
					})

					When("the image file does not exist", func() {
						var validConfig *vmmanagers.OpsmanConfigFilePayload
						BeforeEach(func() {
							configStr := `
opsman-configuration:
  vsphere:
    vcenter:
      url: vcenter.nowhere.nonexist
      username: goodman
      password: badguy
      datastore: datastore
      datacenter: datacenter
      resource_pool: resource-pool
`
							err := yaml.Unmarshal([]byte(configStr), &validConfig)
							Expect(err).ToNot(HaveOccurred())
						})

						It("fails on creation with an image file validation error", func() {
							command, _ := createCommand(configStr, opsmanVersionBelow26)
							command.ImageOVA = "does-not-exist.ova"

							_, _, err := command.CreateVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("could not read image file"))
							Expect(err.Error()).To(ContainSubstring("does-not-exist.ova"))

						})
					})
				})
			})
		})

		Describe("configuration validation", func() {
			DescribeTable("errors when required params are missing", func(param string) {
				command, _ := createCommand("{opsman-configuration: {vsphere: {vcenter: {}, ssh_password: password}}}", opsmanVersionBelow26)

				_, _, err := command.CreateVM()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Field validation for '%s' failed on the 'required' tag", param)))
			},
				Entry("URL is required", "URL"),
				Entry("Username is required", "Username"),
				Entry("Password is required", "Password"),
				Entry("Datacenter is required", "Datacenter"),
				Entry("Datastore is required", "Datastore"),
				Entry("ResourcePool is required", "ResourcePool"),
				Entry("DiskType is required", "DiskType"),
				Entry("PrivateIP is required", "PrivateIP"),
				Entry("DNS is required", "DNS"),
				Entry("NTP is required", "NTP"),
				Entry("Hostname is required", "Hostname"),
				Entry("Network is required", "Network"),
				Entry("Netmask is required", "Netmask"),
				Entry("Gateway is required", "Gateway"),
			)

			When("providing ssh configuration", func() {
				When("OpsMan version is 2.3 to 2.6", func() {
					It("requires ssh_password, ssh_public_key, or both to be set", func() {
						configStrTemplate := `{
						   "opsman-configuration": {
						      "vsphere": {
						         "vcenter": {
						            "url": "vcenter.nowhere.nonexist",
						            "username": "goodman",
						            "password": "badguy",
						            "datastore": "datastore",
						            "datacenter": "datacenter",
						            "insecure": 1,
						            "ca_cert": "ca",
						            "resource_pool": "resource-pool",
						            "host": "host",
						            "folder": "/datacenter/vm/folder"
						         },
						         "disk_type": "thin",
						         "private_ip": "1.2.3.4",
						         "dns": "1.1.1.1",
						         "ntp": "ntp.server.xyz",
								 %s
						         "hostname": "full.domain.name",
						         "network": "some-edge",
						         "netmask": "255.255.255.192",
						         "gateway": "2.2.2.2",
						         "vm_name": "vm_name",
						         "memory": 25,
						         "cpu": 10
						      }
						   }
						}`
						command, _ := createCommand(fmt.Sprintf(configStrTemplate, ""), opsmanVersionBelow26)
						_, _, err := command.CreateVM()
						Expect(err.Error()).To(ContainSubstring("'ssh_password' or 'ssh_public_key' must be set"))

						command, _ = createCommand(fmt.Sprintf(configStrTemplate, "\"ssh_password\": \"password\","), opsmanVersionBelow26)
						_, _, err = command.CreateVM()
						Expect(err).ToNot(HaveOccurred())

						command, _ = createCommand(fmt.Sprintf(configStrTemplate, "\"ssh_public_key\": \"pubkey\","), opsmanVersionBelow26)
						_, _, err = command.CreateVM()
						Expect(err).ToNot(HaveOccurred())

						command, _ = createCommand(fmt.Sprintf(configStrTemplate, "\"ssh_password\": \"password\", \"ssh_public_key\": \"pubkey\","), opsmanVersionBelow26)
						_, _, err = command.CreateVM()
						Expect(err).ToNot(HaveOccurred())
					})
				})

				When("OpsMan version is 2.6 or higher", func() {
					It("requires ssh to be set", func() {
						command, _ := createCommand("{opsman-configuration: {vsphere: {vcenter: {}}}}", "2.6.1")
						_, _, err := command.CreateVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring(`'ssh_public_key' is required for OpsManager 2.6+`))
					})

					It("does not allow ssh_password to be set", func() {
						command, _ := createCommand("{opsman-configuration: {vsphere: {vcenter: {}, ssh_public_key: pubkey, ssh_password: password}}}", "2.6.1")
						_, _, err := command.CreateVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring(`'ssh_password' cannot be used with OpsManager 2.6+`))
					})
				})
			})

			When("insecure is set to 0 (secure)", func() {
				It("requires ca_cert", func() {
					command, _ := createCommand("{opsman-configuration: {vsphere: {vcenter: {insecure: 0}}}}", "2.6.1")
					_, _, err := command.CreateVM()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`'ca_cert' is required if 'insecure' is set to 0 (secure)`))
				})
			})

			Context("defaults", func() {
				var configStr string
				BeforeEach(func() {
					configStr = `{
   						"opsman-configuration": {
   						   "vsphere": {
   						      "vcenter": {
   						         "url": "vcenter.nowhere.nonexist",
   						         "username": "goodman",
   						         "password": "badguy",
   						         "datastore": "datastore",
   						         "datacenter": "datacenter",
   						         "resource_pool": "resource-pool",
								 "ca_cert": "cert"
   						      },
   						      "disk_type": "thin",
   						      "private_ip": "1.2.3.4",
   						      "dns": "8.8.8.8",
   						      "ntp": "com.google",
   						      "hostname": "example.com",
   						      "network": "example-network",
   						      "netmask": "255.255.255.255",
   						      "gateway": "1.2.3.1",
   						      "ssh_password": "password"
   						   }
   						}
					}`
				})

				It("if available, sets params to default", func() {
					command, _ := createCommand(configStr, opsmanVersionBelow26)

					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())

					Expect(command.Config.OpsmanConfig.Vsphere.Insecure).Should(Equal("0"))
					Expect(command.Config.OpsmanConfig.Vsphere.Memory).Should(Equal("8"))
					Expect(command.Config.OpsmanConfig.Vsphere.CPU).Should(Equal("1"))
					Expect(command.Config.OpsmanConfig.Vsphere.VMName).Should(Equal("ops-manager-vm"))
				})
			})
		})
	})

	Context("delete vm", func() {
		const configStr = `
opsman-configuration:
  vsphere:
    vcenter:
      url: vcenter.nowhere.nonexist
      username: goodman
      password: badguy
`

		Context("with a valid config", func() {
			Describe("DeleteVM", func() {
				It("calls govc with correct cli arguments", func() {
					command, runner := createCommand(configStr, opsmanVersionBelow26)
					command.State.ID = "/datacenter/vm/some-posi/testing-machine"

					err := command.DeleteVM()
					Expect(err).ToNot(HaveOccurred())
					_, args := runner.ExecuteWithEnvVarsArgsForCall(1)
					Expect(args).To(matchers.OrderedConsistOf(
						"vm.power",
						"-off=true",
						"-vm.ipath=/datacenter/vm/some-posi/testing-machine",
					))

					_, args = runner.ExecuteWithEnvVarsArgsForCall(2)
					Expect(args).To(matchers.OrderedConsistOf(
						"vm.destroy",
						"-vm.ipath=/datacenter/vm/some-posi/testing-machine",
					))
				})

				Describe("failure cases", func() {
					When("external tools fail", func() {
						It("prints errors from govc", func() {
							for i := 1; i < 3; i++ {
								command, runner := createCommand(configStr, opsmanVersionBelow26)
								command.State.ID = "/datacenter/vm/some-posi/testing-machine"

								runner.ExecuteWithEnvVarsReturnsOnCall(i, nil, nil, errors.New("some error occurred"))
								err := command.DeleteVM()
								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("govc error"))
							}
						})
					})

					When("vm specified in the state file does not exist", func() {
						It("returns an error", func() {
							command, runner := createCommand(configStr, opsmanVersionBelow26)
							runner.ExecuteWithEnvVarsReturnsOnCall(0, nil, nil, errors.New("vm does not exist"))

							command.State = vmmanagers.StateInfo{
								IAAS: "vsphere",
								ID:   "some-vm-id",
							}

							err := command.DeleteVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(Equal("error: vm does not exist\n       Could not find VM with ID \"some-vm-id\".\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile."))

							Expect(runner.ExecuteWithEnvVarsCallCount()).To(BeEquivalentTo(1))
						})
					})
				})
				Describe("configuration validation", func() {
					DescribeTable("errors when required params are missing", func(param string) {
						command, _ := createCommand("{opsman-configuration: {vsphere: {vcenter: {}}}}", opsmanVersionBelow26)
						err := command.DeleteVM()

						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Field validation for '%s' failed on the 'required' tag", param)))
					},
						Entry("URL is required", "URL"),
						Entry("Username is required", "Username"),
						Entry("Password is required", "Password"),
					)
				})

				When("the VM is already powered off", func() {
					deleteVMError := bytes.NewBufferString("The attempted operation cannot be performed in the current state (Powered off)")
					It("Deletes it without error", func() {
						command, runner := createCommand(configStr, opsmanVersionBelow26)

						runner.ExecuteWithEnvVarsReturnsOnCall(0, nil, deleteVMError, errors.New("exit status 1"))

						err := command.DeleteVM()
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("with an invalid iaas", func() {
			var state = vmmanagers.StateInfo{
				IAAS: "gcp",
			}

			It("prints error", func() {
				command, runner := createCommand(configStr, opsmanVersionBelow26)
				runner.ExecuteWithEnvVarsReturns(nil, nil, errors.New("some error occurred"))

				command.State = state
				err := command.DeleteVM()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("authentication file provided is for vsphere, while the state file is for "))
			})
		})
	})

	testIAASForPropertiesInExampleFile("Vsphere")
})

func createOVA(opsmanVersion string) *os.File {
	ovaFile, err := ioutil.TempFile("", fmt.Sprintf("opsman-%s-*-.ova", opsmanVersion))
	Expect(err).ToNot(HaveOccurred())
	defer ovaFile.Close()
	tarWriter := tar.NewWriter(ovaFile)
	defer tarWriter.Close()

	ovfFile, err := ioutil.TempFile("", "file*.ovf")
	Expect(err).ToNot(HaveOccurred())

	header := &tar.Header{
		Name: ovfFile.Name(),
	}

	err = tarWriter.WriteHeader(header)
	Expect(err).ToNot(HaveOccurred())

	_, err = io.Copy(tarWriter, ovfFile)
	Expect(err).ToNot(HaveOccurred())

	return ovaFile
}
