package vmlifecyclecommands_test

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-cf/om/vmlifecycle/vmlifecyclecommands"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers/fakes"
)

var _ = Describe("Create VM", func() {
	var (
		command              vmlifecyclecommands.CreateVM
		outWriter, errWriter *gbytes.Buffer
		fakeService          = &fakes.CreateVMService{}

		configStr = `---
opsman-configuration:
  gcp:
    vm_name: some-name
`
	)

	BeforeEach(func() {
		outWriter = gbytes.NewBuffer()
		errWriter = gbytes.NewBuffer()
	})

	Context("opsman does not exist", func() {
		gcpConfig := `
opsman-configuration:
  gcp:
    gcp_service_account: something
    project: someproject
    region: us-west1
    zone: us-west1-c
    vpc_subnet: infra
    tags: good
    public_ip: 1.2.3.4
`
		vsphereConfig := `
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

		awsConfig := `
opsman-configuration:
  aws:
    public_ip: 1.2.3.4
`

		azureConfig := `
opsman-configuration:
  azure:
    public_ip: 1.2.3.4
`

		openstackConfig := `
opsman-configuration:
  openstack:
    public_ip: 1.2.3.4
`
		DescribeTable("attempts to install Ops Manager on the config-specified IaaS",
			func(configStr string, iaasIdentifier string) {
				command = createCommand(outWriter, errWriter, nil, configStr, "", "", "")

				_ = command.Execute([]string{})
				Expect(outWriter).To(gbytes.Say(iaasIdentifier))
			},
			Entry("works on gcp", gcpConfig, "gcp"),
			Entry("works on vSphere", vsphereConfig, "vSphere"),
			Entry("works on aws", awsConfig, "aws"),
			Entry("works on azure", azureConfig, "azure"),
			Entry("works on openstack", openstackConfig, "openstack"),
		)

		It("modifies provided state file to reflect successful creation of the VM", func() {
			fakeService := &fakes.CreateVMService{}
			fakeService.CreateVMReturns(vmmanagers.Success, vmmanagers.StateInfo{IAAS: "gcp", ID: "vm-id"}, nil)

			command = createCommand(outWriter, errWriter, fakeService, configStr, "", "", "")

			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(outWriter).To(gbytes.Say("OpsMan VM created successfully"))

			finalState, err := os.ReadFile(command.StateFile)
			Expect(err).ToNot(HaveOccurred())

			Expect(finalState).To(MatchYAML(`{"iaas": "gcp", "vm_id": vm-id}`))
		})

		When("no file path is provided for the state file", func() {
			BeforeEach(func() {
				fakeService := &fakes.CreateVMService{}
				fakeService.CreateVMReturns(vmmanagers.Success, vmmanagers.StateInfo{IAAS: "gcp", ID: "vm-id"}, nil)

				initFunc := vmmanagers.NewCreateVMManager
				if fakeService != nil {
					initFunc = func(_ *vmmanagers.OpsmanConfigFilePayload, _ string, _ vmmanagers.StateInfo, _, _ io.Writer) (vmmanagers.CreateVMService, error) {
						return fakeService, nil
					}
				}

				command = vmlifecyclecommands.NewCreateVMCommand(
					outWriter,
					errWriter,
					initFunc,
				)

				_, err := flags.ParseArgs(&command, []string{
					`--image-file`, writeFile("image"),
					`--config`, writeFile(configStr),
				})
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				os.Remove("state.yml")
			})

			It("creates a new state file if none is provided", func() {
				err := command.Execute([]string{})
				Expect(err).ToNot(HaveOccurred())
				Expect(outWriter).To(gbytes.Say("OpsMan VM created successfully"))

				Expect(command.StateFile).To(Equal("state.yml"))
				finalState, err := os.ReadFile(command.StateFile)
				Expect(err).ToNot(HaveOccurred())

				Expect(finalState).To(MatchYAML(`{"iaas": "gcp", "vm_id": vm-id}`))
			})
		})
	})

	When("opsman already exists", func() {
		When("the VM already exists", func() {
			It("returns a warning and rewrites the existing state", func() {
				fakeService.CreateVMReturns(vmmanagers.Exist, vmmanagers.StateInfo{IAAS: "gcp", ID: "vm-id"}, nil)
				command = createCommand(outWriter, errWriter, fakeService, configStr, "", "", "")

				err := command.Execute([]string{})
				Expect(err).ToNot(HaveOccurred())
				Expect(outWriter).To(gbytes.Say("VM already exists, not attempting to create it"))

				finalState, err := os.ReadFile(command.StateFile)
				Expect(err).ToNot(HaveOccurred())

				Expect(finalState).To(MatchYAML(`{"iaas": "gcp", "vm_id": vm-id}`))
			})
		})
	})

	When("using interpolation features", func() {
		validConfig := `---
opsman-configuration:
  gcp:
    gcp_service_account: something
    project: ((project_name))
    region: us-west1
    zone: us-west1-c
    vpc_subnet: infra
    tags: good
`

		It("can interpolate variables into the configuration", func() {
			validVars := `---
project_name: awesome-project
`
			fakeService := &fakes.CreateVMService{}
			command = createCommand(outWriter, errWriter, fakeService, validConfig, validVars, "", "")

			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error of missing variables", func() {
			command = createCommand(outWriter, errWriter, nil, validConfig, "", "", "")

			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected to find variables"))
		})

		It("can interpolate variables from environment variables", func() {
			Expect(os.Setenv("OM_VAR_project_name", "awesome-project")).ToNot(HaveOccurred())
			defer func() {
				os.Unsetenv("OM_VAR_project_name")
			}()
			fakeService := &fakes.CreateVMService{}
			command = createCommand(outWriter, errWriter, fakeService, validConfig, "", "", "")

			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("top level keys are not recognized", func() {
		BeforeEach(func() {
			configStr := `---
opsman-configuration:
  gcp:
    gcp_service_account: something
    project: some-project-name
    region: us-west1
    zone: us-west1-c
    vpc_subnet: infra
    tags: good
unused-top-level-key:
  unused-nested-key: some-value
`
			fakeService := &fakes.CreateVMService{}
			command = createCommand(outWriter, errWriter, fakeService, configStr, "", "", "")
		})

		It("does not return an error", func() {
			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("opsman-configuration is missing from the config", func() {
		BeforeEach(func() {
			configStr := `---
unused-top-level-key-1:
  unused-nested-key: some-value
unused-top-level-key-2:
  unused-nested-key: some-value
`
			fakeService := &fakes.CreateVMService{}
			command = createCommand(outWriter, errWriter, fakeService, configStr, "", "", "")
		})

		It("returns an error highlighting the missing key", func() {
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("top-level-key 'opsman-configuration' is a required key."))
			Expect(err.Error()).To(ContainSubstring("Ensure the correct file is passed, the 'opsman-configuration' key is present, and the key is spelled correctly with a dash(-)."))
			Expect(err.Error()).To(ContainSubstring("Found keys:\n  'unused-top-level-key-1'\n  'unused-top-level-key-2'"))
		})
	})

	When("unrecognized key is provided within opsman configuration", func() {
		BeforeEach(func() {
			configStr := `---
opsman-configuration:
  vsphere:
    something-unknown: totally-good-stuff
`
			command = createCommand(outWriter, errWriter, nil, configStr, "", "", "")
		})

		It("returns error message", func() {
			err := command.Execute([]string{})
			Expect(err.Error()).To(ContainSubstring(`field something-unknown not found`))
		})
	})

	When("no IaaS config is provided", func() {
		BeforeEach(func() {
			configStr := `---
opsman-configuration:
`
			command = createCommand(outWriter, errWriter, nil, configStr, "", "", "")
		})

		It("returns error message", func() {
			err := command.Execute([]string{})
			Expect(err.Error()).To(ContainSubstring(`no iaas configuration provided, please refer to documentation`))
		})
	})

	When("no iaas matches", func() {
		BeforeEach(func() {
			configStr := `---
opsman-configuration:
  non-existent-iaas:
    some-property: some-val
`
			command = createCommand(outWriter, errWriter, nil, configStr, "", "", "")
		})

		It("returns error message", func() {
			err := command.Execute([]string{})
			Expect(err.Error()).To(ContainSubstring(`unknown iaas: non-existent-iaas, please refer to documentation`))
		})
	})

	When("more than one iaas matches", func() {
		BeforeEach(func() {
			configStr := `---
opsman-configuration:
  gcp:
    vm_name: some-name
  vsphere:
    vm_name: some-name
`
			command = createCommand(outWriter, errWriter, nil, configStr, "", "", "")
		})

		It("returns error message", func() {
			err := command.Execute([]string{})
			Expect(err.Error()).To(ContainSubstring(`more than one iaas matched, only one in config allowed`))
		})
	})

	When("config file is not valid", func() {
		It("fails with an invalid YAML file", func() {
			command = createCommand(outWriter, errWriter, nil, "invalid YAML", "", "", "")
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("could not load Opsman Config file (%s): yaml: unmarshal errors", command.Config)))
		})

		When("There is a vars file", func() {
			It("it includes the vars file path in error messages", func() {
				command = createCommand(outWriter, errWriter, nil, configStr, "invalid YAML", "", "")
				err := command.Execute([]string{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Deserializing variables file '%s'", command.VarsFile[0])))
			})
		})
	})

	When("config file does not exist", func() {
		BeforeEach(func() {
			command = createCommand(outWriter, errWriter, nil, "", "", "", "")
			command.Config = "never-gonna-give-you-up.txt"
		})
		It("returns an error saying it cannot read the file", func() {
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("open never-gonna-give-you-up.txt: no such file or directory"))
		})
	})

	When("state file is not valid", func() {
		It("fails with an invalid YAML file", func() {
			command = createCommand(outWriter, errWriter, nil, configStr, "", "", "invalid YAML")
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("could not load state file (%s): yaml: unmarshal errors", command.StateFile)))
		})
	})

	When("image file does not exist", func() {
		var command vmlifecyclecommands.CreateVM
		BeforeEach(func() {
			gcpConfig := `
opsman-configuration:
  gcp:
    gcp_service_account: something
    project: someproject
    region: us-west1
    zone: us-west1-c
    vpc_subnet: infra
    tags: good
`
			command = createCommand(outWriter, errWriter, nil, gcpConfig, "", "", "")
			command.ImageFile = "never-gonna-give-you-up.txt"
		})
		It("returns an error saying it cannot read the file", func() {
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not read image file"))
		})
	})

	When("the stateFile vm_id does not map to an Ops Manager", func() {
		It("returns an error and gives a relevant error message", func() {
			fakeService.CreateVMReturns(vmmanagers.StateMismatch, vmmanagers.StateInfo{}, nil)
			command = createCommand(outWriter, errWriter, fakeService, configStr, "", "", "")

			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("VM specified in the statefile does not exist in your IAAS"))
		})
	})

	When("the vm is not created and there is an error", func() {
		It("does not write or change the state file", func() {
			fakeService.CreateVMReturns(vmmanagers.Unknown, vmmanagers.StateInfo{}, errors.New("unknown error"))
			command = createCommand(outWriter, errWriter, fakeService, configStr, "", "", "iaas: gcp\nvm_id: 123")

			err := command.Execute([]string{})
			Expect(err.Error()).To(Equal("unexpected error: unknown error"))

			Expect(readFile(command.StateFile)).To(MatchYAML("iaas: gcp\nvm_id: 123"))
		})
	})

	When("the vm is created, but some some subsequent step fails", func() {
		BeforeEach(func() {
			fakeService.CreateVMReturns(vmmanagers.Incomplete, vmmanagers.StateInfo{IAAS: "gcp", ID: "vm-id"}, errors.New("unknown error"))
			command = createCommand(outWriter, errWriter, fakeService, configStr, "", "", "")
		})
		It("still writes the VM ID to the statefile", func() {
			err := command.Execute([]string{})
			Expect(err.Error()).To(Equal("the VM was created, but subsequent configuration failed: unknown error"))

			finalState, err := os.ReadFile(command.StateFile)
			Expect(err).ToNot(HaveOccurred())

			Expect(finalState).To(MatchYAML(`{"iaas": "gcp", "vm_id": vm-id}`))
		})
	})
})

func createCommand(
	stdWriter,
	errWriter io.Writer,
	fakeService *fakes.CreateVMService,
	configuration string,
	variables string,
	image string,
	state string,
) vmlifecyclecommands.CreateVM {
	initFunc := vmmanagers.NewCreateVMManager
	if fakeService != nil {
		initFunc = func(_ *vmmanagers.OpsmanConfigFilePayload, _ string, _ vmmanagers.StateInfo, _, _ io.Writer) (vmmanagers.CreateVMService, error) {
			return fakeService, nil
		}
	}

	command := vmlifecyclecommands.NewCreateVMCommand(
		stdWriter,
		errWriter,
		initFunc,
	)

	command.Config = writeFile(configuration)
	command.VarsFile = []string{writeFile(variables)}
	command.VarsEnv = []string{"OM_VAR"}
	command.ImageFile = writeFile(image)
	command.StateFile = writeFile(state)
	return command
}
