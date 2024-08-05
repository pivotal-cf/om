package vmlifecyclecommands_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/vmlifecycle/vmlifecyclecommands"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers/fakes"
	"io"
	"io/ioutil"
	"os"
	"time"
)

var _ = Describe("Delete VM", func() {
	var (
		outWriter, errWriter *gbytes.Buffer
		configStr            = `---
opsman-configuration:
  gcp:
    vm_name: some-name
`
	)

	BeforeEach(func() {
		outWriter = gbytes.NewBuffer()
		errWriter = gbytes.NewBuffer()
	})

	When("a vm exists", func() {
		gcpConfig := `
opsman-configuration:
  gcp:
    gcp_service_account: something
`
		vsphereConfig := `
opsman-configuration:
  vsphere:
    vcenter:
      url: vcenter.nowhere.nonexist
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
		DescribeTable("attempts to delete vm on the config-specified IaaS",
			func(configStr string, iaasIdentifier string) {
				command := deleteCommand(outWriter, errWriter, nil, configStr, "", fmt.Sprintf(`{"iaas": "%s", "vm_id": "id"}`, iaasIdentifier))

				_ = command.Execute([]string{})
				Expect(outWriter).To(gbytes.Say(iaasIdentifier))
			},
			Entry("works on gcp", gcpConfig, "gcp"),
			Entry("works on Vsphere", vsphereConfig, "vSphere"),
			Entry("works on aws", awsConfig, "aws"),
			Entry("works on azure", azureConfig, "azure"),
			Entry("works on openstack", openstackConfig, "openstack"),
		)

		It("modifies state file to reflect deletion of the VM", func() {
			fakeService := &fakes.DeleteVMService{}
			fakeService.DeleteVMReturns(nil)

			uuid := time.Now().UnixNano()
			command := deleteCommand(outWriter, errWriter, fakeService, configStr, "", fmt.Sprintf(`{"iaas": "iaas-%d", "vm_id": "some_id"}`, uuid))

			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(outWriter).To(gbytes.Say("VM deleted successfully\n"))

			finalState, err := ioutil.ReadFile(command.StateFile)
			Expect(err).ToNot(HaveOccurred())

			Expect(finalState).To(MatchYAML(fmt.Sprintf(`{"iaas": "iaas-%d"}`, uuid)))
		})
	})

	When("vm does not exist", func() {
		It("no op", func() {
			fakeService := &fakes.DeleteVMService{}
			fakeService.DeleteVMReturns(nil)

			command := deleteCommand(outWriter, errWriter, fakeService, configStr, "", `{"iaas": "gcp"}`)

			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(outWriter).To(gbytes.Say("Nothing to do\n"))

			finalState, err := ioutil.ReadFile(command.StateFile)
			Expect(err).ToNot(HaveOccurred())

			Expect(finalState).To(MatchYAML(`{"iaas": "gcp"}`))
		})
	})

	When("using interpolation features", func() {

		config := `---
opsman-configuration:
  gcp:
    gcp_service_account: something
    project: ((project_name))
`

		It("can interpolate variables into the configuration", func() {
			validVars := `---
project_name: awesome-project
`
			fakeService := &fakes.DeleteVMService{}

			command := deleteCommand(outWriter, errWriter, fakeService, config, validVars, "")

			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error of missing variables", func() {
			fakeService := &fakes.DeleteVMService{}

			command := deleteCommand(outWriter, errWriter, fakeService, config, "", "")

			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected to find variables"))
		})

		It("can interpolate variables from environment variables", func() {
			Expect(os.Setenv("OM_VAR_project_name", "awesome-project")).ToNot(HaveOccurred())
			defer func() {
				os.Unsetenv("OM_VAR_project_name")
			}()
			fakeService := &fakes.DeleteVMService{}
			command := deleteCommand(outWriter, errWriter, fakeService, config, "", "")

			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("top level keys are not recognized", func() {
		var command vmlifecyclecommands.DeleteVM
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
			fakeService := &fakes.DeleteVMService{}
			command = deleteCommand(outWriter, errWriter, fakeService, configStr, "", "")
		})

		It("does not return an error", func() {
			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("opsman-configuration is missing from the config", func() {
		var command vmlifecyclecommands.DeleteVM
		BeforeEach(func() {
			configStr := `---
unused-top-level-key-1:
  unused-nested-key: some-value
unused-top-level-key-2:
  unused-nested-key: some-value
`
			fakeService := &fakes.DeleteVMService{}
			command = deleteCommand(outWriter, errWriter, fakeService, configStr, "", "")
		})

		It("returns an error highlighting the missing key", func() {
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("top-level-key 'opsman-configuration' is a required key."))
			Expect(err.Error()).To(ContainSubstring("Ensure the correct file is passed, the 'opsman-configuration' key is present, and the key is spelled correctly with a dash(-)."))
			Expect(err.Error()).To(ContainSubstring("Found keys:\n  'unused-top-level-key-1'\n  'unused-top-level-key-2'"))
		})
	})

	When("unrecognized key is provided in opsman-configuration", func() {
		It("returns error message", func() {
			configStr := `---
opsman-configuration:
  vsphere:
    unknown-key: something
`
			command := deleteCommand(outWriter, errWriter, nil, configStr, "", "")
			err := command.Execute([]string{})
			Expect(err.Error()).To(ContainSubstring(`field unknown-key not found `))
		})
	})

	When("no IaaS config is provided", func() {
		It("returns error message", func() {
			configStr := `---
opsman-configuration:
`
			command := deleteCommand(outWriter, errWriter, nil, configStr, "", "")
			err := command.Execute([]string{})
			Expect(err.Error()).To(ContainSubstring(`no iaas configuration provided, please refer to documentation`))
		})
	})

	When("no iaas matches", func() {
		It("returns error message", func() {
			configStr := `---
opsman-configuration:
  non-existent-iaas:
    some-property: some-val
`
			command := deleteCommand(outWriter, errWriter, nil, configStr, "", "")
			err := command.Execute([]string{})
			Expect(err.Error()).To(ContainSubstring(`unknown iaas: non-existent-iaas, please refer to documentation`))
		})

	})

	When("more than one iaas matches", func() {
		It("returns error message", func() {
			configStr := `---
opsman-configuration:
  gcp:
    project: some-name
  vsphere:
    ssh_password: some-name
`
			command := deleteCommand(outWriter, errWriter, nil, configStr, "", "")
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`more than one iaas matched, only one in config allowed`))
		})
	})

	When("config file is not valid", func() {
		It("fails with an invalid YAML file", func() {
			invalidConfig := `not valid yaml`
			command := deleteCommand(outWriter, errWriter, nil, invalidConfig, "", "")
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("yaml: unmarshal errors"))
		})
	})

	When("config file does not exist", func() {
		It("returns an error saying it cannot read the file", func() {
			command := deleteCommand(outWriter, errWriter, nil, "", "", "")
			command.Config = "never-gonna-give-you-up.txt"
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not interpolate config file"))
		})
	})

	When("state file does not exist", func() {
		var command vmlifecyclecommands.DeleteVM
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
			command = deleteCommand(outWriter, errWriter, nil, gcpConfig, "", "")
			command.StateFile = "never-gonna-give-you-up.txt"
		})
		It("returns an error saying it cannot read the file", func() {
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("open never-gonna-give-you-up.txt"))
		})
	})
})

func deleteCommand(
	stdWriter, errWriter io.Writer,
	fakeService *fakes.DeleteVMService,
	configuration string,
	variables string,
	state string,
) vmlifecyclecommands.DeleteVM {
	initFunc := vmmanagers.NewDeleteVMManager
	if fakeService != nil {
		initFunc = func(_ *vmmanagers.OpsmanConfigFilePayload, _ string, _ vmmanagers.StateInfo, _, _ io.Writer) (vmmanagers.DeleteVMService, error) {
			return fakeService, nil
		}
	}

	command := vmlifecyclecommands.NewDeleteVMCommand(
		stdWriter,
		errWriter,
		initFunc,
	)
	command.Config = writeFile(configuration)
	command.VarsFile = []string{writeFile(variables)}
	command.VarsEnv = []string{"OM_VAR"}
	command.StateFile = writeFile(state)
	return command
}
