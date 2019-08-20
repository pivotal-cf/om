package commands_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

const ymlVMExtensionFile = `---
vm-extension-config:
  cloud_properties:
    elbs:
      - some-elb
    iam_instance_profile: some-iam-profile
  name: ((vm_extension_name))`

const ymlVMExtensionNoNameFile = `---
vm-extension-config:
  cloud_properties:
    elbs:
      - some-elb
    iam_instance_profile: some-iam-profile`

var _ = Describe("CreateVMExtension", func() {
	var (
		fakeService *fakes.CreateVMExtensionService
		fakeLogger  *fakes.Logger
		command     commands.CreateVMExtension
		configFile  *os.File
		err         error
		varsFile    *os.File
	)

	BeforeEach(func() {
		fakeService = &fakes.CreateVMExtensionService{}
		fakeLogger = &fakes.Logger{}
		command = commands.NewCreateVMExtension(func() []string { return nil }, fakeService, fakeLogger)
	})

	AfterEach(func() {
		if configFile != nil {
			os.RemoveAll(configFile.Name())
		}
		if varsFile != nil {
			os.RemoveAll(varsFile.Name())
		}
	})

	Describe("Execute", func() {
		It("makes a request to the OpsMan to create a VM extension", func() {
			err := command.Execute([]string{
				"--name", "some-vm-extension",
				"--cloud-properties", "{ \"iam_instance_profile\": \"some-iam-profile\", \"elbs\": [\"some-elb\"] }",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeService.CreateStagedVMExtensionArgsForCall(0)).To(Equal(api.CreateVMExtension{
				Name:            "some-vm-extension",
				CloudProperties: json.RawMessage(`{ "iam_instance_profile": "some-iam-profile", "elbs": ["some-elb"] }`),
			}))

			Expect(fakeLogger.PrintfCallCount()).To(Equal(1))
			format, content := fakeLogger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("VM Extension 'some-vm-extension' created/updated\n"))
		})

		When("using a config file", func() {
			Context("with a vars file", func() {
				It("makes a request to the OpsMan to create a VM extension", func() {
					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					varsFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(ymlVMExtensionFile)
					Expect(err).NotTo(HaveOccurred())

					_, err = varsFile.WriteString(`vm_extension_name: some-vm-extension`)
					Expect(err).NotTo(HaveOccurred())

					err := command.Execute([]string{
						"--config", configFile.Name(),
						"--vars-file", varsFile.Name(),
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(fakeService.CreateStagedVMExtensionArgsForCall(0)).To(Equal(api.CreateVMExtension{
						Name:            "some-vm-extension",
						CloudProperties: json.RawMessage("{\"elbs\":[\"some-elb\"],\"iam_instance_profile\":\"some-iam-profile\"}"),
					}))

					Expect(fakeLogger.PrintfCallCount()).To(Equal(1))
					format, content := fakeLogger.PrintfArgsForCall(0)
					Expect(fmt.Sprintf(format, content...)).To(Equal("VM Extension 'some-vm-extension' created/updated\n"))
				})

			})

			Context("with a var defined", func() {
				It("makes a request to the OpsMan to create a VM extension", func() {
					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					varsFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(ymlVMExtensionFile)
					Expect(err).NotTo(HaveOccurred())

					err := command.Execute([]string{
						"--config", configFile.Name(),
						"--var", "vm_extension_name=some-vm-extension",
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(fakeService.CreateStagedVMExtensionArgsForCall(0)).To(Equal(api.CreateVMExtension{
						Name:            "some-vm-extension",
						CloudProperties: json.RawMessage("{\"elbs\":[\"some-elb\"],\"iam_instance_profile\":\"some-iam-profile\"}"),
					}))

					Expect(fakeLogger.PrintfCallCount()).To(Equal(1))
					format, content := fakeLogger.PrintfArgsForCall(0)
					Expect(fmt.Sprintf(format, content...)).To(Equal("VM Extension 'some-vm-extension' created/updated\n"))
				})

			})

			Context("with environment variables", func() {
				It("makes a request to the OpsMan to create a VM extension", func() {
					command = commands.NewCreateVMExtension(
						func() []string { return []string{"OM_VAR_vm_extension_name=some-vm-extension"} },
						fakeService,
						fakeLogger)
					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(ymlVMExtensionFile)
					Expect(err).NotTo(HaveOccurred())

					err := command.Execute([]string{
						"--config", configFile.Name(),
						"--vars-env", "OM_VAR",
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(fakeService.CreateStagedVMExtensionArgsForCall(0)).To(Equal(api.CreateVMExtension{
						Name:            "some-vm-extension",
						CloudProperties: json.RawMessage("{\"elbs\":[\"some-elb\"],\"iam_instance_profile\":\"some-iam-profile\"}"),
					}))

					Expect(fakeLogger.PrintfCallCount()).To(Equal(1))
					format, content := fakeLogger.PrintfArgsForCall(0)
					Expect(fmt.Sprintf(format, content...)).To(Equal("VM Extension 'some-vm-extension' created/updated\n"))
				})

			})

		})

		Context("failure cases", func() {
			When("the service fails to create a VM extension", func() {
				It("returns an error", func() {
					fakeService.CreateStagedVMExtensionReturns(errors.New("failed to create VM extension"))

					err := command.Execute([]string{
						"--name", "some-vm-extension",
						"--cloud-properties", "{ \"iam_instance_profile\": \"some-iam-profile\", \"elbs\": [\"some-elb\"] }",
					})

					Expect(err).To(MatchError("failed to create VM extension"))
				})
			})

			When("an unknown flag is provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--badflag"})
					Expect(err).To(MatchError("could not parse create-vm-extension flags: flag provided but not defined: -badflag"))
				})
			})

			Context("error when name is not provided", func() {
				It("returns an error when flag is missing", func() {
					err := command.Execute([]string{"--cloud-properties", "{ \"iam_instance_profile\": \"some-iam-profile\", \"elbs\": [\"some-elb\"] }"})
					Expect(err).To(MatchError("VM Extension name must provide name via --name flag"))
					Expect(fakeService.CreateStagedVMExtensionCallCount()).Should(Equal(0))
				})
				It("returns an error when name not in file", func() {
					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(ymlVMExtensionNoNameFile)
					Expect(err).NotTo(HaveOccurred())

					err := command.Execute([]string{
						"--config", configFile.Name(),
					})

					Expect(err).To(MatchError("Config file must contain name element"))
					Expect(fakeService.CreateStagedVMExtensionCallCount()).Should(Equal(0))

				})
			})

			Context("fails to interpolate config file", func() {
				It("returns an error", func() {
					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(ymlVMExtensionFile)
					Expect(err).NotTo(HaveOccurred())

					err := command.Execute([]string{
						"--config", configFile.Name(),
					})

					Expect(err.Error()).Should(ContainSubstring("Expected to find variables"))

				})
			})

			Context("bad yaml in config file", func() {
				It("returns an error", func() {
					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString("asdfasdf")
					Expect(err).NotTo(HaveOccurred())

					err := command.Execute([]string{
						"--config", configFile.Name(),
					})

					Expect(err.Error()).Should(ContainSubstring("could not be parsed as valid configuration: yaml"))

				})
			})

		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewCreateVMExtension(nil, nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This creates/updates a VM extension",
				ShortDescription: "creates/updates a VM extension",
				Flags:            command.Options,
			}))
		})
	})
})
