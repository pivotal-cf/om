package vmmanagers_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/om/vmlifecycle/matchers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers/fakes"
)

var _ = Describe("Azure VMManager", func() {

	createCommand := func(region string, configStrTemplate string, state vmmanagers.StateInfo) (*vmmanagers.AzureVMManager, *fakes.AzureRunner, *bytes.Buffer) {
		var err error
		runner := &fakes.AzureRunner{}
		testUriFile, err := os.CreateTemp("", "some*.yaml")
		Expect(err).ToNot(HaveOccurred())
		_, _ = testUriFile.WriteString(`
---
west_us: https://opsmanagerwestus.blob.core.windows.net/images/ops-manager-2.2-build.292.vhd
east_us: https://opsmanagereastus.blob.core.windows.net/images/ops-manager-2.2-build.292.vhd
west_europe: https://opsmanagerwesteurope.blob.core.windows.net/images/ops-manager-2.2-build.292.vhd
southeast_asia: https://opsmanagersoutheastasia.blob.core.windows.net/images/ops-manager-2.2-build.292.vhd
`)
		Expect(testUriFile.Close()).ToNot(HaveOccurred())

		var validConfig *vmmanagers.OpsmanConfigFilePayload
		err = yaml.UnmarshalStrict([]byte(fmt.Sprintf(configStrTemplate, region)), &validConfig)
		Expect(err).ToNot(HaveOccurred())

		stderr := &bytes.Buffer{}
		command := vmmanagers.NewAzureVMManager(io.Discard, stderr, validConfig, testUriFile.Name(), state, runner, time.Millisecond)
		return command, runner, stderr
	}

	Context("create vm", func() {
		const configStrTemplate = `
opsman-configuration:
  azure:
    vm_name: vm-name
    subscription_id: subscription
    resource_group: rg
    tenant_id: tenant
    client_id: client
    client_secret: secret
    location: %s
    public_ip: 1.2.3.4
    private_ip: 10.0.0.3
    container: opsman_container
    boot_disk_size: 200
    network_security_group: nsg
    subnet_id: /subscriptions/sub-guid/resourceGroups/network-resource-group/providers/Microsoft.Network/virtualNetworks/vnet1/subnets/sub1
    storage_account: account
    storage_key: key
    ssh_public_key: asdfasdfs
    cloud_name: AzureGermanCloud
    tags: Project=ECommerce CostCenter=00123 Team=Web
`
		const configWithManagedDiskDisabledStrTemplate = `
opsman-configuration:
  azure:
    vm_name: vm-name
    subscription_id: subscription
    resource_group: rg
    tenant_id: tenant
    client_id: client
    client_secret: secret
    location: %s
    public_ip: 1.2.3.4
    private_ip: 10.0.0.3
    container: opsman_container
    boot_disk_size: 200
    network_security_group: nsg
    subnet_id: /subscriptions/sub-guid/resourceGroups/network-resource-group/providers/Microsoft.Network/virtualNetworks/vnet1/subnets/sub1
    storage_account: account
    storage_key: key
    ssh_public_key: asdfasdfs
    use_managed_disk: false
`
		const configStrWithoutIPsTemplate = `
opsman-configuration:
  azure:
    vm_name: vm-name
    subscription_id: subscription
    resource_group: rg
    tenant_id: tenant
    client_id: client
    client_secret: secret
    location: %s
    container: opsman_container
    boot_disk_size: 200
    network_security_group: nsg
    subnet_id: /subscriptions/sub-guid/resourceGroups/network-resource-group/providers/Microsoft.Network/virtualNetworks/vnet1/subnets/sub1
    storage_account: account
    storage_key: key
    ssh_public_key: asdfasdfs
`
		const configStrWithoutStorageKeyTemplate = `
opsman-configuration:
  azure:
    vm_name: vm-name
    subscription_id: subscription
    resource_group: rg
    tenant_id: tenant
    client_id: client
    client_secret: secret
    location: %s
    public_ip: 1.2.3.4
    private_ip: 10.0.0.3
    container: opsman_container
    boot_disk_size: 200
    network_security_group: nsg
    subnet_id: /subscriptions/sub-guid/resourceGroups/network-resource-group/providers/Microsoft.Network/virtualNetworks/vnet1/subnets/sub1
    storage_account: account
    ssh_public_key: asdfasdfs
    cloud_name: AzureGermanCloud
`
		Context("with a valid config", func() {
			var (
				command *vmmanagers.AzureVMManager
				runner  *fakes.AzureRunner
			)
			setupCommand := func(config string) {
				//NOTE: these outputs are in strings because `--query` returns results as JSON formatted strings
				command, runner, _ = createCommand("westus", config, vmmanagers.StateInfo{})
				runner.ExecuteWithEnvVarsReturnsOnCall(4, bytes.NewBufferString("\"pending\"\r\n"), nil, nil)
				runner.ExecuteWithEnvVarsReturnsOnCall(5, bytes.NewBufferString("\"success\"\r\n"), nil, nil)
				runner.ExecuteWithEnvVarsReturnsOnCall(6, bytes.NewBufferString("\"https://example.com/asdf\"\r\n"), nil, nil)
				runner.ExecuteWithEnvVarsReturnsOnCall(7, bytes.NewBufferString("\"some-ip\"\r\n"), nil, nil)
				runner.ExecuteWithEnvVarsReturnsOnCall(8, bytes.NewBufferString("\"/some/resource/id/image\"\r\n"), nil, nil)
			}

			It("calls azure with correct cli arguments", func() {
				setupCommand(configStrTemplate)
				status, stateInfo, err := command.CreateVM()
				Expect(err).ToNot(HaveOccurred())
				Expect(status).To(Equal(vmmanagers.Success))
				Expect(stateInfo).To(Equal(vmmanagers.StateInfo{IAAS: "azure", ID: "vm-name"}))

				commands := [][]interface{}{
					{
						"cloud", "set",
						"--name", "AzureGermanCloud",
					},
					{
						`login`, `--service-principal`,
						`-u`, gstruct.Ignore(),
						`-p`, gstruct.Ignore(),
						`--tenant`, `tenant`,
					},
					{
						"account", "set", "--subscription", "subscription",
					},
					{
						"storage", "blob", "copy", "start",
						"--source-uri", "https://opsmanagerwestus.blob.core.windows.net/images/ops-manager-2.2-build.292.vhd",
						"--destination-container", "opsman_container",
						"--destination-blob", "opsman-image-3cc17a6d403e0f6560f601ff3b22bd4c.vhd",
					},
					{
						"storage", "blob", "show",
						"--name", "opsman-image-3cc17a6d403e0f6560f601ff3b22bd4c.vhd",
						"--container-name", "opsman_container",
						"--query", "properties.copy.status",
					},
					{
						"storage", "blob", "show",
						"--name", "opsman-image-3cc17a6d403e0f6560f601ff3b22bd4c.vhd",
						"--container-name", "opsman_container",
						"--query", "properties.copy.status",
					},
					{
						"storage", "blob", "url",
						"--name", "opsman-image-3cc17a6d403e0f6560f601ff3b22bd4c.vhd",
						"--container-name", "opsman_container",
					},
					{
						"network", "public-ip", "list",
						"--query", "[?ipAddress == '1.2.3.4'].name | [0]",
					},
					{
						"image", "create",
						"--resource-group", "rg",
						"--name", "opsman-image-3cc17a6d403e0f6560f601ff3b22bd4c",
						"--source", "https://example.com/asdf",
						"--location", "westus", "--os-type", "Linux",
						"--query", "id",
					},
					{
						"vm", "create", "--name", "vm-name", "--resource-group", "rg",
						"--location", "westus",
						"--os-disk-size-gb", "200",
						"--admin-username", "ubuntu",
						"--size", "Standard_DS2_v2",
						"--ssh-key-value", MatchRegexp(".*"),
						"--subnet", "/subscriptions/sub-guid/resourceGroups/network-resource-group/providers/Microsoft.Network/virtualNetworks/vnet1/subnets/sub1",
						"--nsg", "nsg",
						"--public-ip-address", "some-ip",
						"--private-ip-address", "10.0.0.3",
						"--storage-sku", "Standard_LRS",
						"--image", "/some/resource/id/image",
						"--tags", "Project=ECommerce CostCenter=00123 Team=Web",
					},
					{
						"storage", "blob", "delete-batch",
						"--source", "opsman_container",
						"--pattern", "opsman-image-*.vhd",
					},
				}

				for i, expectedArgs := range commands {
					Expect(i).To(BeNumerically("<", runner.ExecuteWithEnvVarsCallCount()))
					_, actualArgs := runner.ExecuteWithEnvVarsArgsForCall(i)
					Expect(actualArgs).To(matchers.OrderedConsistOf(expectedArgs))
				}
			})

			Context("DEPRECATED tests", func() {
				const configWithManagedDiskAndUnmanagedDiskStrTemplateDEPRECATED = `
opsman-configuration:
  azure:
    vm_name: vm-name
    subscription_id: subscription
    resource_group: rg
    tenant_id: tenant
    client_id: client
    client_secret: secret
    location: %s
    public_ip: 1.2.3.4
    private_ip: 10.0.0.3
    container: opsman_container
    boot_disk_size: 200
    network_security_group: nsg
    subnet_id: /subscriptions/sub-guid/resourceGroups/network-resource-group/providers/Microsoft.Network/virtualNetworks/vnet1/subnets/sub1
    storage_account: account
    storage_key: key
    ssh_public_key: asdfasdfs
    use_managed_disk: false
    use_unmanaged_disk: true
`
				const configWithVPCSubnetAndSubnetIDStrTemplateDEPRECATED = `
opsman-configuration:
  azure:
    vm_name: vm-name
    subscription_id: subscription
    resource_group: rg
    tenant_id: tenant
    client_id: client
    client_secret: secret
    location: %s
    public_ip: 1.2.3.4
    private_ip: 10.0.0.3
    container: opsman_container
    boot_disk_size: 200
    network_security_group: nsg
    subnet_id: /subscriptions/sub-guid/resourceGroups/network-resource-group/providers/Microsoft.Network/virtualNetworks/vnet1/subnets/sub1
    vpc_subnet: /subscriptions/sub-guid/resourceGroups/network-resource-group/providers/Microsoft.Network/virtualNetworks/vnet1/subnets/sub1
    storage_account: account
    storage_key: key
    ssh_public_key: asdfasdfs
    use_managed_disk: false
`

				const configWithUnmanagedDiskStrTemplateDEPRECATED = `
opsman-configuration:
  azure:
    vm_name: vm-name
    subscription_id: subscription
    resource_group: rg
    tenant_id: tenant
    client_id: client
    client_secret: secret
    location: %s
    public_ip: 1.2.3.4
    private_ip: 10.0.0.3
    container: opsman_container
    boot_disk_size: 200
    network_security_group: nsg
    subnet_id: /subscriptions/sub-guid/resourceGroups/network-resource-group/providers/Microsoft.Network/virtualNetworks/vnet1/subnets/sub1
    storage_account: account
    storage_key: key
    ssh_public_key: asdfasdfs
    use_unmanaged_disk: true
`
				const DEPRECATEDconfigStrVPCSubnetTemplate = `
opsman-configuration:
  azure:
    vm_name: vm-name
    subscription_id: subscription
    resource_group: rg
    tenant_id: tenant
    client_id: client
    client_secret: secret
    location: %s
    public_ip: 1.2.3.4
    private_ip: 10.0.0.3
    container: opsman_container
    boot_disk_size: 200
    network_security_group: nsg
    vpc_subnet: /subscriptions/sub-guid/resourceGroups/network-resource-group/providers/Microsoft.Network/virtualNetworks/vnet1/subnets/sub1
    storage_account: account
    storage_key: key
    ssh_public_key: asdfasdfs
    cloud_name: AzureGermanCloud
`
				It("DEPRECATED: allows azure to set vpc_subnet instead of subnet_id (backwards compatability)", func() {
					setupCommand(DEPRECATEDconfigStrVPCSubnetTemplate)
					status, stateInfo, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal(vmmanagers.Success))
					Expect(stateInfo).To(Equal(vmmanagers.StateInfo{IAAS: "azure", ID: "vm-name"}))

					commands := [][]interface{}{
						{
							"cloud", "set",
							"--name", "AzureGermanCloud",
						},
						{
							`login`, `--service-principal`,
							`-u`, gstruct.Ignore(),
							`-p`, gstruct.Ignore(),
							`--tenant`, `tenant`,
						},
						{
							"account", "set", "--subscription", "subscription",
						},
						{
							"storage", "blob", "copy", "start",
							"--source-uri", "https://opsmanagerwestus.blob.core.windows.net/images/ops-manager-2.2-build.292.vhd",
							"--destination-container", "opsman_container",
							"--destination-blob", "opsman-image-3cc17a6d403e0f6560f601ff3b22bd4c.vhd",
						},
						{
							"storage", "blob", "show",
							"--name", "opsman-image-3cc17a6d403e0f6560f601ff3b22bd4c.vhd",
							"--container-name", "opsman_container",
							"--query", "properties.copy.status",
						},
						{
							"storage", "blob", "show",
							"--name", "opsman-image-3cc17a6d403e0f6560f601ff3b22bd4c.vhd",
							"--container-name", "opsman_container",
							"--query", "properties.copy.status",
						},
						{
							"storage", "blob", "url",
							"--name", "opsman-image-3cc17a6d403e0f6560f601ff3b22bd4c.vhd",
							"--container-name", "opsman_container",
						},
						{
							"network", "public-ip", "list",
							"--query", "[?ipAddress == '1.2.3.4'].name | [0]",
						},
						{
							"image", "create",
							"--resource-group", "rg",
							"--name", "opsman-image-3cc17a6d403e0f6560f601ff3b22bd4c",
							"--source", "https://example.com/asdf",
							"--location", "westus", "--os-type", "Linux",
							"--query", "id",
						},
						{
							"vm", "create", "--name", "vm-name", "--resource-group", "rg",
							"--location", "westus",
							"--os-disk-size-gb", "200",
							"--admin-username", "ubuntu",
							"--size", "Standard_DS2_v2",
							"--ssh-key-value", MatchRegexp(".*"),
							"--subnet", "/subscriptions/sub-guid/resourceGroups/network-resource-group/providers/Microsoft.Network/virtualNetworks/vnet1/subnets/sub1",
							"--nsg", "nsg",
							"--public-ip-address", "some-ip",
							"--private-ip-address", "10.0.0.3",
							"--storage-sku", "Standard_LRS",
							"--image", "/some/resource/id/image",
						},
						{
							"storage", "blob", "delete-batch",
							"--source", "opsman_container",
							"--pattern", "opsman-image-*.vhd",
						},
					}

					for i, expectedArgs := range commands {
						Expect(i).To(BeNumerically("<", runner.ExecuteWithEnvVarsCallCount()))
						_, actualArgs := runner.ExecuteWithEnvVarsArgsForCall(i)
						Expect(actualArgs).To(matchers.OrderedConsistOf(expectedArgs))
					}
				})

				It("DEPRECATED: should error if both vpc_subnet and subnet_id are defined", func() {
					setupCommand(configWithVPCSubnetAndSubnetIDStrTemplateDEPRECATED)
					_, _, err := command.CreateVM()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("\"vpc_subnet\" is DEPRECATED. Cannot use \"vpc_subnet\" and \"subnet_id\" together. Use \"subnet_id\" instead"))
				})

				It("DEPRECATED: should error if both use_unmanaged_disk and use_managed_disk are defined", func() {
					setupCommand(configWithManagedDiskAndUnmanagedDiskStrTemplateDEPRECATED)
					_, _, err := command.CreateVM()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("\"use_unmanaged_disk\" is DEPRECATED. Cannot use \"use_unmanaged_disk\" and \"use_managed_disk\" together. Use \"use_managed_disk\" instead"))
				})

				When("DEPRECATED unmanaged disk is enabled(backwards compatability)", func() {
					It("creates the VM with an unmanaged disk", func() {
						setupCommand(configWithUnmanagedDiskStrTemplateDEPRECATED)
						_, _, err := command.CreateVM()
						Expect(err).ToNot(HaveOccurred())

						_, actualArgs := runner.ExecuteWithEnvVarsArgsForCall(8)
						Expect(actualArgs).To(matchers.OrderedConsistOf(
							"vm",
							"create",
							"--name",
							"vm-name",
							"--resource-group",
							"rg",
							"--location",
							"westus",
							"--os-disk-size-gb",
							"200",
							"--admin-username",
							"ubuntu",
							"--size",
							"Standard_DS2_v2",
							"--ssh-key-value",
							MatchRegexp("ssh-key"),
							"--subnet",
							"/subscriptions/sub-guid/resourceGroups/network-resource-group/providers/Microsoft.Network/virtualNetworks/vnet1/subnets/sub1",
							"--nsg",
							"nsg",
							"--public-ip-address",
							"some-ip",
							"--private-ip-address",
							"10.0.0.3",
							"--use-unmanaged-disk",
							"--os-disk-name",
							MatchRegexp("opsman-disk-[a-f0-9]{32}"),
							"--os-type",
							"Linux",
							"--storage-container-name",
							"opsman_container",
							"--storage-account",
							"account",
							"--image",
							"https://example.com/asdf",
						))
					})
				})
			})

			It("calls azure with the correct environment variables", func() {
				setupCommand(configStrTemplate)
				_, _, err := command.CreateVM()
				Expect(err).ToNot(HaveOccurred())

				for i := 0; i < runner.ExecuteWithEnvVarsCallCount(); i++ {
					env, _ := runner.ExecuteWithEnvVarsArgsForCall(i)
					Expect(env).Should(ContainElement(`AZURE_STORAGE_KEY=key`))
					Expect(env).Should(ContainElement(`AZURE_STORAGE_ACCOUNT=account`))
				}
			})

			It("does not set storage key if not provided", func() {
				setupCommand(configStrWithoutStorageKeyTemplate)
				_, _, err := command.CreateVM()
				Expect(err).ToNot(HaveOccurred())

				for i := 0; i < runner.ExecuteWithEnvVarsCallCount(); i++ {
					env, _ := runner.ExecuteWithEnvVarsArgsForCall(i)
					Expect(env).Should(Equal([]string{`AZURE_STORAGE_ACCOUNT=account`}))
				}
			})

			When("the disk storage sku is provided", func() {
				It("uses it when creating the VM", func() {
					setupCommand(configStrTemplate + "    storage_sku: CustomSKU")
					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())

					invocations := getInvocations(runner, "ExecuteWithEnvVars")
					Expect(invocations).To(ContainElement(MatchRegexp("vm create.*--storage-sku CustomSKU")))
					Expect(invocations).ToNot(ContainElement(MatchRegexp(`--storage-sku.*--storage-sku`)))
				})
			})

			When("the vm size is provided", func() {
				It("uses it when creating the VM", func() {
					setupCommand(configStrTemplate + "    vm_size: CustomVM")
					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())

					invocations := getInvocations(runner, "ExecuteWithEnvVars")
					Expect(invocations).To(ContainElement(MatchRegexp("vm create.*--size CustomVM")))
					Expect(invocations).ToNot(ContainElement(MatchRegexp(`--size.*--size`)))
				})
			})

			When("use_managed_disk is disabled", func() {
				It("creates the VM with an unmanaged disk", func() {
					setupCommand(configWithManagedDiskDisabledStrTemplate)
					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())

					invocations := getInvocations(runner, "ExecuteWithEnvVars")

					expectedArgs := strings.Join([]string{"--use-unmanaged-disk",
						"--os-disk-name",
						"opsman-disk-[a-f0-9]{32}",
						"--os-type",
						"Linux",
						"--storage-container-name",
						"opsman_container",
						"--storage-account",
						"account",
						"--image",
						"https://example.com/asdf"}, " ")

					Expect(invocations).To(ContainElement(MatchRegexp(fmt.Sprintf("vm create.*%s", expectedArgs))))
				})
			})

			When("private IP is not provided", func() {
				It("does not attach it to the azure vm instance", func() {
					setupCommand(configStrTemplate)
					command.Config.OpsmanConfig.Azure.PublicIP = "1.2.3.4"
					command.Config.OpsmanConfig.Azure.PrivateIP = ""
					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())

					invocations := getInvocations(runner, "ExecuteWithEnvVars")
					Expect(invocations).ToNot(HaveLen(0))

					for _, line := range invocations {
						Expect(line).ToNot(ContainSubstring("private-ip"))
					}
				})
			})

			When("public IP is not provided", func() {
				It("does not attach it to the azure vm instance", func() {
					setupCommand(configStrTemplate)
					command.Config.OpsmanConfig.Azure.PublicIP = ""
					command.Config.OpsmanConfig.Azure.PrivateIP = "10.0.0.3"
					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())

					invocations := getInvocations(runner, "ExecuteWithEnvVars")
					Expect(invocations).ToNot(HaveLen(0))

					By("setting the --public-ip-address to '' so azure does not automatically create one")
					Expect(invocations).To(ContainElement(MatchRegexp(`--public-ip-address ""`)))
				})
			})

			DescribeTable("uses the correct image for the location", func(region string, matcher types.GomegaMatcher) {
				command, runner, _ := createCommand(region, configStrTemplate, vmmanagers.StateInfo{})

				runner.ExecuteWithEnvVarsReturns(bytes.NewBufferString("\"success\"\r\n"), nil, nil)
				_, _, err := command.CreateVM()
				Expect(err).ToNot(HaveOccurred())

				_, args := runner.ExecuteWithEnvVarsArgsForCall(3)
				Expect(args).To(ContainElement(matcher))
			},
				Entry("for westus", "westus", Equal("https://opsmanagerwestus.blob.core.windows.net/images/ops-manager-2.2-build.292.vhd")),
				Entry("for eastus", "eastus", Equal("https://opsmanagereastus.blob.core.windows.net/images/ops-manager-2.2-build.292.vhd")),
				Entry("for australiacentral", "australiacentral", MatchRegexp("https://.*.blob.core.windows.net/images/ops-manager-2.2-build.292.vhd")),
				Entry("for any-region", "any-region", MatchRegexp("https://.*.blob.core.windows.net/images/ops-manager-2.2-build.292.vhd")),
			)

			When("vm already exists in the state file", func() {
				It("returns exist status, doesn't make additional CLI calls, and exits 0", func() {
					setupCommand(configStrTemplate)
					runner.ExecuteWithEnvVarsReturnsOnCall(1, bytes.NewBufferString(`"/subscriptions/some-id/some-vm-id"`), nil, nil)

					command.State = vmmanagers.StateInfo{
						IAAS: "azure",
						ID:   "some-vm-id",
					}

					status, state, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal(vmmanagers.Exist))

					Expect(runner.ExecuteWithEnvVarsCallCount()).To(BeEquivalentTo(4))

					_, actualArgs := runner.ExecuteWithEnvVarsArgsForCall(3)
					Expect(actualArgs).To(Equal([]interface{}{
						`vm`, `show`,
						`--name`, "some-vm-id",
						`--resource-group`, `rg`,
						`--query`, `id`,
					}))

					Expect(state.IAAS).To(Equal("azure"))
					Expect(state.ID).To(Equal("some-vm-id"))
				})

				Describe("failure cases", func() {
					When("vm is not exist", func() {
						It("returns an error", func() {
							setupCommand(configStrTemplate)
							runner.ExecuteWithEnvVarsReturnsOnCall(3, bytes.NewBufferString("\r\n"), nil, nil)

							command.State = vmmanagers.StateInfo{
								IAAS: "azure",
								ID:   "invalid-id",
							}

							status, _, err := command.CreateVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Could not find VM with ID \"invalid-id\".\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile."))
							Expect(status).To(Equal(vmmanagers.Unknown))

							Expect(runner.ExecuteWithEnvVarsCallCount()).To(BeEquivalentTo(4))
						})
					})
				})
			})

			Describe("failure cases", func() {
				When("external tools fail", func() {
					It("prints errors from azure", func() {
						for i := 0; i <= 6; i++ {
							command, runner, _ := createCommand("westus", configStrTemplate, vmmanagers.StateInfo{})
							runner.ExecuteWithEnvVarsReturns(bytes.NewBufferString("\"success\"\r\n"), nil, nil)
							runner.ExecuteWithEnvVarsReturnsOnCall(i, nil, nil, errors.New("some error occurred"))
							status, _, err := command.CreateVM()
							Expect(status).To(Equal(vmmanagers.Unknown))
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("azure error: "))
						}
					})
				})
				When("the state file has an invalid IAAS", func() {
					badState := vmmanagers.StateInfo{
						IAAS: "gcp",
					}

					It("prints error", func() {
						setupCommand(configStrTemplate)
						runner.ExecuteWithEnvVarsReturns(nil, nil, errors.New("some error occurred"))

						command.State = badState
						status, _, err := command.CreateVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("authentication file provided is for azure, while the state file is for "))
						Expect(status).To(Equal(vmmanagers.Unknown))
					})
				})
				When("the image file is not valid YAML", func() {
					It("returns that the yaml is invalid", func() {
						var validConfig *vmmanagers.OpsmanConfigFilePayload
						runner := &fakes.AzureRunner{}
						runner.ExecuteWithEnvVarsReturns(bytes.NewBufferString("\"success\"\r\n"), nil, nil)

						invalidUriFile, err := os.CreateTemp("", "some*.yaml")
						Expect(err).ToNot(HaveOccurred())
						_, _ = invalidUriFile.WriteString("not valid yaml")
						Expect(invalidUriFile.Close()).ToNot(HaveOccurred())
						err = yaml.UnmarshalStrict([]byte(fmt.Sprintf(configStrTemplate, "us-west-1")), &validConfig)
						Expect(err).ToNot(HaveOccurred())

						command := vmmanagers.NewAzureVMManager(io.Discard, io.Discard, validConfig, invalidUriFile.Name(), vmmanagers.StateInfo{}, runner, time.Millisecond)
						_, _, err = command.CreateVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("cannot unmarshal"))

					})
				})

				When("image file is not a yaml file", func() {
					BeforeEach(func() {
						setupCommand(configStrTemplate)
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
						var validConfig *vmmanagers.OpsmanConfigFilePayload
						runner := &fakes.AzureRunner{}
						runner.ExecuteWithEnvVarsReturns(bytes.NewBufferString("\"success\"\r\n"), nil, nil)

						err := yaml.UnmarshalStrict([]byte(fmt.Sprintf(configStrTemplate, "us-west-1")), &validConfig)
						Expect(err).ToNot(HaveOccurred())
						command := vmmanagers.NewAzureVMManager(io.Discard, io.Discard, validConfig, "does-not-exist.yml", vmmanagers.StateInfo{}, runner, time.Millisecond)
						_, _, err = command.CreateVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("open does-not-exist.yml"))
					})
				})

				When("Provided publicIP address does not exists", func() {
					It("returns an error specifying the wrong IP address", func() {
						setupCommand(configStrTemplate)
						runner.ExecuteWithEnvVarsReturnsOnCall(7, bytes.NewBufferString(""), nil, nil)
						command.Config.OpsmanConfig.Azure.PublicIP = "1.2.3.4"
						command.Config.OpsmanConfig.Azure.PrivateIP = "10.0.0.3"
						_, _, err := command.CreateVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("could not find resource assignment for public ip address " + command.Config.OpsmanConfig.Azure.PublicIP))
					})
				})
			})
		})

		DescribeTable("errors when required params are missing", func(param string) {
			invalidConfig := &vmmanagers.OpsmanConfigFilePayload{OpsmanConfig: vmmanagers.OpsmanConfig{
				Azure: &vmmanagers.AzureConfig{},
			}}
			runner := &fakes.AzureRunner{}

			command := vmmanagers.NewAzureVMManager(io.Discard, io.Discard, invalidConfig, "", vmmanagers.StateInfo{}, runner, time.Millisecond)
			_, _, err := command.CreateVM()
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Field validation for '%s' failed on the 'required' tag", param)))
		},
			Entry("requires StorageAccount", "StorageAccount"),
			Entry("requires ClientID", "ClientID"),
			Entry("requires ClientSecret", "ClientSecret"),
			Entry("requires TenantID", "TenantID"),
			Entry("requires SubscriptionID", "SubscriptionID"),
			Entry("requires Location", "Location"),
			Entry("requires ResourceGroup", "ResourceGroup"),
			Entry("requires SubnetID", "SubnetID"),
			Entry("requires SSHPublicKey", "SSHPublicKey"),
		)

		It("will not error when optional NSG is missing", func() {
			configStr := `
opsman-configuration:
  azure:
    subscription_id: subscription
    resource_group: rg
    tenant_id: tenant
    client_id: client
    client_secret: secret
    location: westus
    public_ip: 1.2.3.4
    container: opsman_container
    subnet_id: subnet
    storage_account: account
    storage_key: key
    ssh_public_key: asdfasdfs
`
			var validConfig *vmmanagers.OpsmanConfigFilePayload
			err := yaml.UnmarshalStrict([]byte(configStr), &validConfig)
			Expect(err).ToNot(HaveOccurred())

			runner := &fakes.AzureRunner{}

			file, err := os.CreateTemp("", "some*.yml")
			Expect(err).ToNot(HaveOccurred())

			command := vmmanagers.NewAzureVMManager(io.Discard, io.Discard, validConfig, file.Name(), vmmanagers.StateInfo{}, runner, time.Millisecond)
			runner.ExecuteWithEnvVarsReturns(bytes.NewBufferString("\"success\"\r\n"), nil, nil)
			_, _, err = command.CreateVM()
			Expect(err).ToNot(HaveOccurred())
		})

		It("requires at least public IP or private IP to be set", func() {
			command, _, _ := createCommand("westus", configStrWithoutIPsTemplate, vmmanagers.StateInfo{})
			_, _, err := command.CreateVM()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("PublicIP and/or PrivateIP must be set"))
		})

		It("defaulting any missing optional params", func() {
			runner := &fakes.AzureRunner{}
			configStr := `
opsman-configuration:
  azure:
    subscription_id: subscription
    resource_group: rg
    tenant_id: tenant
    client_id: client
    client_secret: secret
    location: westus
    public_ip: 1.2.3.4
    network_security_group: nsg
    subnet_id: subnet
    storage_account: account
    storage_key: key
    ssh_public_key: asdfasdfs
`
			var validConfig *vmmanagers.OpsmanConfigFilePayload
			err := yaml.UnmarshalStrict([]byte(configStr), &validConfig)
			Expect(err).ToNot(HaveOccurred())

			file, err := os.CreateTemp("", "some*.yml")
			Expect(err).ToNot(HaveOccurred())
			command := vmmanagers.NewAzureVMManager(io.Discard, io.Discard, validConfig, file.Name(), vmmanagers.StateInfo{}, runner, time.Millisecond)

			runner.ExecuteWithEnvVarsReturns(bytes.NewBufferString("\"success\"\r\n"), nil, nil)
			_, _, err = command.CreateVM()
			Expect(err).ToNot(HaveOccurred())

			_, args := runner.ExecuteWithEnvVarsArgsForCall(0)
			Expect(args).To(ContainElement(MatchRegexp("AzureCloud")))

			_, args = runner.ExecuteWithEnvVarsArgsForCall(8)
			Expect(args).To(ContainElement(MatchRegexp("ops-manager-vm")))
			Expect(args).To(ContainElement(MatchRegexp("200")))

			_, args = runner.ExecuteWithEnvVarsArgsForCall(3)
			Expect(args).To(ContainElement(MatchRegexp("opsmanagerimage")))

			_, args = runner.ExecuteWithEnvVarsArgsForCall(7)
			Expect(args).To(ContainElement(MatchRegexp("create")))
			Expect(args).ToNot(ContainElement("--use-unmanaged-disk"))
		})

		When("use_managed_disk is not a valid bool string", func() {
			It("returns an error", func() {
				configStrTemplate := `
opsman-configuration:
  azure:
    use_managed_disk: notABoolean   
`
				var validConfig *vmmanagers.OpsmanConfigFilePayload
				err := yaml.UnmarshalStrict([]byte(configStrTemplate), &validConfig)
				Expect(err).ToNot(HaveOccurred())

				runner := &fakes.AzureRunner{}

				file, err := os.CreateTemp("", "some*.yml")
				Expect(err).ToNot(HaveOccurred())

				command := vmmanagers.NewAzureVMManager(io.Discard, io.Discard, validConfig, file.Name(), vmmanagers.StateInfo{}, runner, time.Millisecond)
				_, _, err = command.CreateVM()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("expected \"use_managed_disk\" to be a boolean. Got: notABoolean."))
			})
		})
	})

	Context("delete vm", func() {
		const configStrTemplate = `
opsman-configuration:
  azure:
    location: %s
    client_id: client
    client_secret: secret
    subscription_id: subscription
    resource_group: rg
    tenant_id: tenant
    cloud_name: AzureGermanCloud
    container: container
    storage_account: account
    storage_key: some-key
`
		const configWithUnmanagedDiskTemplate = `
opsman-configuration:
  azure:
    location: %s
    client_id: client
    client_secret: secret
    subscription_id: subscription
    resource_group: rg
    tenant_id: tenant
    cloud_name: AzureGermanCloud
    container: container
    storage_account: account
    storage_key: some-key
`
		const vmInfo = `
{
	"networkProfile": {
		"networkInterfaces": [
			{
				"id": "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/nicName"
			},
			{
				"id": "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/nicName2"
			}
		]
	},
	"storageProfile": {
		"imageReference": {
			"id": "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Compute/images/imageName"
		},
		"osDisk": {
			"managedDisk": {
				"id": "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Compute/disks/diskName"
			}
		}
	}
}
`

		const vmInfoUnmanagedDisk = `
{
	"networkProfile": {
		"networkInterfaces": [
			{
				"id": "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/nicName"
			},
			{
				"id": "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/nicName2"
			}
		]
	},
	"storageProfile": {
		"imageReference": {
			"id": "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Compute/images/imageName"
		},
		"osDisk": {
			"vhd": {"uri": "https://ciopsmanager.blob.core.windows.net/os-disk-container/diskName"}
		}
	}
}
`

		Context("with a valid config", func() {
			It("calls azure with correct cli arguments", func() {
				command, runner, _ := createCommand("us-west-1", configStrTemplate, vmmanagers.StateInfo{IAAS: "azure", ID: "vm-name"})

				runner.ExecuteWithEnvVarsReturnsOnCall(4, bytes.NewBufferString(vmInfo), nil, nil)
				runner.ExecuteWithEnvVarsReturnsOnCall(5, bytes.NewBufferString(vmInfo), nil, nil)

				err := command.DeleteVM()
				Expect(err).ToNot(HaveOccurred())

				commands := [][]interface{}{
					{
						"cloud", "set", "--name", "AzureGermanCloud",
					},
					{
						"login", "--service-principal",
						"-u", "client",
						"-p", gstruct.Ignore(),
						"--tenant", "tenant",
					},
					{
						"account", "set", "--subscription", "subscription",
					},
					{
						"vm", "show",
						"--name", "vm-name",
						"--resource-group", "rg",
						"--query", "id",
					},
					{
						"vm", "show",
						"--name", "vm-name",
						"--resource-group", "rg",
					},
					{
						"vm", "delete",
						"--yes",
						"--name", "vm-name",
						"--resource-group", "rg",
					},
					{
						"disk", "delete",
						"--yes",
						"--ids", "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Compute/disks/diskName",
					},
					{
						"network", "nic", "delete",
						"--ids", "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/nicName",
						"/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/nicName2",
					},
					{
						"image", "delete",
						"--ids", "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Compute/images/imageName",
					},
				}

				for i, expectedArgs := range commands {
					Expect(i).To(BeNumerically("<", runner.ExecuteWithEnvVarsCallCount()))
					_, actualArgs := runner.ExecuteWithEnvVarsArgsForCall(i)
					Expect(actualArgs).To(matchers.OrderedConsistOf(expectedArgs))
				}
			})

			Describe("failure cases", func() {
				When("external tools fail", func() {
					It("prints errors from azure", func() {
						for i := 0; i <= 2; i++ {
							command, runner, _ := createCommand("westus", configStrTemplate, vmmanagers.StateInfo{IAAS: "azure"})
							runner.ExecuteWithEnvVarsReturns(nil, nil, nil)
							runner.ExecuteWithEnvVarsReturnsOnCall(i, nil, nil, errors.New("some error occurred"))
							err := command.DeleteVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("azure error"))
						}
					})

					It("append sub-command errors together", func() {
						command, runner, _ := createCommand("westus", configStrTemplate, vmmanagers.StateInfo{IAAS: "azure"})
						runner.ExecuteWithEnvVarsReturnsOnCall(3, bytes.NewBufferString(vmInfo), nil, nil)

						runner.ExecuteWithEnvVarsReturnsOnCall(4, nil, nil, errors.New("delete vm error"))
						runner.ExecuteWithEnvVarsReturnsOnCall(6, nil, nil, errors.New("delete nic error"))

						err := command.DeleteVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(MatchRegexp("azure error deleting vm: delete vm error\n" +
							"azure error deleting nic: delete nic error"))
					})
				})

				When("vm specified in the state file does not exist", func() {
					It("returns an error", func() {
						command, runner, _ := createCommand("westus", configStrTemplate, vmmanagers.StateInfo{IAAS: "azure", ID: "invalid-id"})
						runner.ExecuteWithEnvVarsReturnsOnCall(3, bytes.NewBufferString("\r\n"), bytes.NewBufferString("The Resource 'Microsoft.Compute/virtualMachines/invalid-id' under resource group 'rg' was not found."), errors.New("vm does not exist"))

						command.State = vmmanagers.StateInfo{
							IAAS: "azure",
							ID:   "invalid-id",
						}

						err := command.DeleteVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Could not find VM with ID \"invalid-id\".\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile."))

						Expect(runner.ExecuteWithEnvVarsCallCount()).To(BeEquivalentTo(4))
					})
				})
			})
		})

		DescribeTable("errors when required params are missing", func(param string) {
			invalidConfig := &vmmanagers.OpsmanConfigFilePayload{OpsmanConfig: vmmanagers.OpsmanConfig{
				Azure: &vmmanagers.AzureConfig{},
			}}
			runner := &fakes.AzureRunner{}

			azureRunner := vmmanagers.NewAzureVMManager(io.Discard, io.Discard, invalidConfig, "", vmmanagers.StateInfo{}, runner, time.Millisecond)
			err := azureRunner.DeleteVM()
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Field validation for '%s' failed on the 'required' tag", param)))
		},
			Entry("requires ClientID", "ClientID"),
			Entry("requires ClientSecret", "ClientSecret"),
			Entry("requires TenantID", "TenantID"),
			Entry("requires SubscriptionID", "SubscriptionID"),
			Entry("requires Location", "Location"),
			Entry("requires ResourceGroup", "ResourceGroup"),
		)

		Context("with an invalid iaas", func() {
			state := vmmanagers.StateInfo{
				IAAS: "gcp",
			}

			It("prints error", func() {
				command, _, _ := createCommand("us-west-1", configStrTemplate, vmmanagers.StateInfo{})

				command.State = state
				err := command.DeleteVM()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("authentication file provided is for azure, while the state file is for "))
			})
		})

		When("neither managed or unmanaged disk is found", func() {
			It("gives a warning but does not error", func() {
				unmanagedDisk := `
{
	"networkProfile": {
		"networkInterfaces": [
			{
				"id": "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/nicName"
			},
			{
				"id": "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/nicName2"
			}
		]
	},
	"storageProfile": {
		"imageReference": {
			"id": "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Compute/images/imageName"
		}
	}
}
`
				command, runner, stderr := createCommand("us-west-1", configStrTemplate, vmmanagers.StateInfo{IAAS: "azure", ID: "vm-name"})

				runner.ExecuteWithEnvVarsReturnsOnCall(4, bytes.NewBufferString(unmanagedDisk), nil, nil)
				runner.ExecuteWithEnvVarsReturnsOnCall(5, bytes.NewBufferString(unmanagedDisk), nil, nil)

				err := command.DeleteVM()
				Expect(err).ToNot(HaveOccurred())

				Expect(stderr.String()).To(MatchRegexp("could not find disk to cleanup. Doing Nothing."))
			})
		})

		When("there is no managed disk associated with the VM", func() {
			It("deletes the VM and cleans up the unmanaged disk", func() {
				command, runner, _ := createCommand("us-west-1", configWithUnmanagedDiskTemplate, vmmanagers.StateInfo{IAAS: "azure", ID: "vm-name"})

				runner.ExecuteWithEnvVarsReturnsOnCall(4, bytes.NewBufferString(vmInfoUnmanagedDisk), nil, nil)
				runner.ExecuteWithEnvVarsReturnsOnCall(5, bytes.NewBufferString(vmInfoUnmanagedDisk), nil, nil)

				err := command.DeleteVM()
				Expect(err).ToNot(HaveOccurred())

				commands := [][]interface{}{
					{
						"cloud", "set", "--name", "AzureGermanCloud",
					},
					{
						"login", "--service-principal",
						"-u", "client",
						"-p", gstruct.Ignore(),
						"--tenant", "tenant",
					},
					{
						"account", "set", "--subscription", "subscription",
					},
					{
						"vm", "show",
						"--name", "vm-name",
						"--resource-group", "rg",
						"--query", "id",
					},
					{
						"vm", "show",
						"--name", "vm-name",
						"--resource-group", "rg",
					},
					{
						"vm", "delete",
						"--yes",
						"--name", "vm-name",
						"--resource-group", "rg",
					},
					{
						"storage", "blob", "delete",
						"--container-name", "os-disk-container",
						"--name", "diskName",
					},
					{
						"network", "nic", "delete",
						"--ids", "/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/nicName",
						"/subscriptions/sub-id/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/nicName2",
					},
				}

				for i, expectedArgs := range commands {
					Expect(i).To(BeNumerically("<", runner.ExecuteWithEnvVarsCallCount()))
					_, actualArgs := runner.ExecuteWithEnvVarsArgsForCall(i)
					Expect(actualArgs).To(matchers.OrderedConsistOf(expectedArgs))
				}
			})
		})
	})

	testIAASForPropertiesInExampleFile("Azure")
})

func getInvocations(r *fakes.AzureRunner, method string) []string {
	invocations := []string{}
	for _, i := range r.Invocations()[method] {
		args := []string{}
		for _, arg := range i[1].([]interface{}) {
			if s, ok := arg.(string); ok && s == "" {
				args = append(args, `""`)
			} else {
				args = append(args, fmt.Sprintf("%s", arg))
			}

		}
		invoke := strings.Join(args, " ")
		invocations = append(invocations, invoke)
	}
	return invocations
}
