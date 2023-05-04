package vmmanagers_test

import (
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/vmlifecycle/matchers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers/fakes"
)

var _ = Describe("AWS VMManager", func() {
	const vmID = "i-0016d0fe3a11c73c2"

	createCommand := func(config string) (*vmmanagers.AWSVMManager, *fakes.AwsRunner) {
		var err error
		runner := &fakes.AwsRunner{}
		var validConfig *vmmanagers.OpsmanConfigFilePayload
		err = yaml.Unmarshal([]byte(config), &validConfig)
		Expect(err).ToNot(HaveOccurred())

		testUriFile, err := ioutil.TempFile("", "some*.yaml")
		Expect(err).ToNot(HaveOccurred())
		_, _ = testUriFile.WriteString(`
---
us-east-1: ami-63b6961c
us-east-2: ami-11e1d974
us-west-1: ami-19a9497a
us-west-2: ami-789dc900
`)
		Expect(testUriFile.Close()).ToNot(HaveOccurred())

		command := vmmanagers.NewAWSVMManager(ioutil.Discard, ioutil.Discard, validConfig, testUriFile.Name(), vmmanagers.StateInfo{}, runner, time.Millisecond)

		return command, runner
	}

	CreateValidCommandWithSecrets := func(publicIP, privateIP, region string, assumeRole string, accessKey string, secretAccessKey string) (*vmmanagers.AWSVMManager, *fakes.AwsRunner) {
		configStrTemplate := `
opsman-configuration:
  aws:
    version: 1.2.3-build.4
    access_key_id: %s
    secret_access_key: %s
    assume_role: %s
    region: %s
    vm_name: awesome-vm
    vpc_subnet_id: awesome-subnet
    security_group_ids: [sg-awesome, sg-great]
    key_pair_name: superuser
    iam_instance_profile_name: awesome-profile
    boot_disk_size: 200
    public_ip: %s
    private_ip: %s
    instance_type: m3.large
    tags:
      Owner: DbAdmin
      Stack: Test
    
`

		command, runner := createCommand(fmt.Sprintf(configStrTemplate, accessKey, secretAccessKey, assumeRole, region, publicIP, privateIP))
		runner.ExecuteWithEnvVarsReturnsOnCall(0, bytes.NewBufferString(fmt.Sprintf("\"%s\"\r\n", vmID)), nil, nil)
		runner.ExecuteWithEnvVarsReturnsOnCall(1, bytes.NewBufferString("null\r\n"), bytes.NewBufferString("TestStatus: pending creation"), nil)
		runner.ExecuteWithEnvVarsReturnsOnCall(2, bytes.NewBufferString("\"vol-0cf5b911680a78bb9\"\r\n"), bytes.NewBufferString("TestStatus: volume created"), nil)
		runner.ExecuteWithEnvVarsReturnsOnCall(4, bytes.NewBufferString("\"eipalloc-18643c24\"\r\n"), nil, nil)
		runner.ExecuteWithEnvVarsReturnsOnCall(7, bytes.NewBufferString("\"stopping\"\r\n"), nil, nil)
		runner.ExecuteWithEnvVarsReturnsOnCall(8, bytes.NewBufferString("\"stopped\"\r\n"), nil, nil)
		return command, runner
	}

	createValidCommand := func(publicIP, privateIP, region string) (*vmmanagers.AWSVMManager, *fakes.AwsRunner) {
		return CreateValidCommandWithSecrets(publicIP, privateIP, region, "", "some-key-id", "some-key-secret")
	}

	Context("create vm", func() {
		Context("with a config with all values filled out", func() {
			It("calls aws with correct cli arguments and does not describe instances", func() {
				numCLICalls := 10
				commands := [][]string{
					{ //1
						`ec2`, `run-instances`,
						`--tag-specifications`, `ResourceType=instance,Tags=[{Key=Name,Value=awesome-vm},{Key=Owner,Value=DbAdmin},{Key=Stack,Value=Test}]`,
						`--image-id`, `ami-789dc900`,
						`--subnet-id`, `awesome-subnet`,
						`--security-group-ids`, `sg-awesome`, `sg-great`,
						`--count`, `1`,
						`--instance-type`, "m3.large",
						`--key-name`, "superuser",
						`--no-associate-public-ip-address`,
						`--iam-instance-profile`, `Name=awesome-profile`,
						`--query`, `Instances[0].InstanceId`,
						`--private-ip-address`, `10.10.10.10`,
					},
					{ //2
						`ec2`, `describe-volumes`,
						`--filters`, `Name=attachment.instance-id,Values=i-0016d0fe3a11c73c2`,
						`Name=attachment.status,Values=attached`,
						`Name=status,Values=in-use`,
						`--query`, `Volumes[0].VolumeId`,
					},
					{ //3
						`ec2`, `describe-volumes`,
						`--filters`, `Name=attachment.instance-id,Values=i-0016d0fe3a11c73c2`,
						`Name=attachment.status,Values=attached`,
						`Name=status,Values=in-use`,
						`--query`, `Volumes[0].VolumeId`,
					},
					{ //4
						`ec2`, `modify-volume`,
						`--volume-id`, `vol-0cf5b911680a78bb9`,
						`--size`, `200`,
					},
					{ //5
						`ec2`, `describe-addresses`,
						`--filters`, `Name=public-ip,Values=1.2.3.4`,
						`--query`, `Addresses[0].AllocationId`,
					},
					{ //6
						`ec2`, `associate-address`,
						`--allocation-id`, `eipalloc-18643c24`,
						`--instance-id`, `i-0016d0fe3a11c73c2`,
					},
					{ //7
						`ec2`, `stop-instances`,
						`--instance-ids`, `i-0016d0fe3a11c73c2`,
					},
					{ //8
						`ec2`, `describe-instances`,
						`--instance-ids`, `i-0016d0fe3a11c73c2`,
						`--query`, `Reservations[*].Instances[*].State.Name`,
					},
					{ //9
						`ec2`, `describe-instances`,
						`--instance-ids`, `i-0016d0fe3a11c73c2`,
						`--query`, `Reservations[*].Instances[*].State.Name`,
					},
					{ //10
						`ec2`, `start-instances`,
						`--instance-ids`, `i-0016d0fe3a11c73c2`,
					},
				}
				command, runner := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
				status, _, err := command.CreateVM()
				Expect(err).ToNot(HaveOccurred())
				Expect(status).To(Equal(vmmanagers.Success))

				Expect(runner.ExecuteWithEnvVarsCallCount()).To(BeEquivalentTo(numCLICalls))

				for i, expectedArgs := range commands {
					_, actualArgs := runner.ExecuteWithEnvVarsArgsForCall(i)
					Expect(actualArgs).To(matchers.OrderedConsistOf(expectedArgs))
				}
			})

			When("private IP is not provided", func() {
				It("does not attach it to the ec2 instance", func() {
					command, runner := createValidCommand("1.2.3.4", "", "us-west-2")
					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					invokes := runner.Invocations()["ExecuteWithEnvVars"]
					Expect(invokes).ToNot(HaveLen(0))
					for _, args := range invokes {
						Expect(args[1]).ToNot(ContainElement(MatchRegexp("private-ip")))
					}
				})
			})

			When("the public IP is not provided", func() {
				It("does not attach it to the ec2 instance", func() {
					command, runner := createValidCommand("", "1.2.3.5", "us-west-2")
					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					invokes := runner.Invocations()["ExecuteWithEnvVars"]
					Expect(invokes).ToNot(HaveLen(0))
					for _, args := range invokes {
						Expect(args[1]).ToNot(ContainElement(MatchRegexp("associate-address")))
						Expect(args[1]).ToNot(ContainElement(MatchRegexp("decsribe-address")))
					}
				})
			})

			It("calls aws with the correct environment variables", func() {
				command, runner := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
				_, _, err := command.CreateVM()
				Expect(err).ToNot(HaveOccurred())

				for i := 0; i < runner.ExecuteWithEnvVarsCallCount(); i++ {
					env, _ := runner.ExecuteWithEnvVarsArgsForCall(i)
					Expect(env).Should(ContainElement(`AWS_ACCESS_KEY_ID=some-key-id`))
					Expect(env).Should(ContainElement(`AWS_SECRET_ACCESS_KEY=some-key-secret`))
					Expect(env).Should(ContainElement(`AWS_DEFAULT_REGION=us-west-2`))
				}
			})

			When("using instance profiles", func() {
				It("calls aws with the correct environment variables", func() {
					command, runner := CreateValidCommandWithSecrets("1.2.3.4", "10.10.10.10", "us-west-2", "", "", "")
					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())

					for i := 0; i < runner.ExecuteWithEnvVarsCallCount(); i++ {
						env, _ := runner.ExecuteWithEnvVarsArgsForCall(i)
						Expect(env).ShouldNot(ContainElement(`AWS_ACCESS_KEY_ID=some-key-id`))
						Expect(env).ShouldNot(ContainElement(`AWS_SECRET_ACCESS_KEY=some-key-secret`))
						Expect(env).Should(ContainElement(`AWS_DEFAULT_REGION=us-west-2`))
					}
				})

				It("calls aws with the correct environment variables when assume_role is set", func() {
					config := `
opsman-configuration:
  aws:
    version: 1.2.3-build.4
    assume_role: "toad" # this field is the thing that makes the config file be written
    region: us-west-2
    vm_name: awesome-vm
    vpc_subnet_id: awesome-subnet
    security_group_ids: [sg-awesome, sg-great]
    key_pair_name: superuser
    iam_instance_profile_name: awesome-profile
    boot_disk_size: 200
    public_ip: 1.1.1.1
    private_ip: 10.0.0.1
    instance_type: m3.large
    tags:
      Owner: DbAdmin
      Stack: Test`
					command, runner := createCommand(config)

					var (
						// awsConfigFileContents
						awsConfigFileContents []string
					)
					happyPathStub := happyPathAWSRunnerStubFunc(vmID)
					runner.ExecuteWithEnvVarsCalls(func(env []string, _ []interface{}) (*bytes.Buffer, *bytes.Buffer, error) {
						p, found := makeEnvironmentMap(env)["AWS_CONFIG_FILE"]
						if found {
							buf, err := os.ReadFile(p)
							if err != nil {
								panic(err)
							}
							awsConfigFileContents = append(awsConfigFileContents, string(buf))
						}
						return happyPathStub()
					})
					_, _, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())

					for i := 0; i < runner.ExecuteWithEnvVarsCallCount(); i++ {
						env, _ := runner.ExecuteWithEnvVarsArgsForCall(i)
						comment := fmt.Sprintf("call %d", i)
						Expect(env).ToNot(ContainElement(`AWS_ACCESS_KEY_ID=some-key-id`), comment)
						Expect(env).ToNot(ContainElement(`AWS_SECRET_ACCESS_KEY=some-key-secret`), comment)
						Expect(env).To(ContainElement(`AWS_DEFAULT_REGION=us-west-2`), comment)
						envMap := makeEnvironmentMap(env)
						Expect(envMap).To(HaveKey("AWS_CONFIG_FILE"), comment)
						Expect(strings.TrimSpace(awsConfigFileContents[i])).To(Equal(strings.TrimSpace(`[profile p-automator-assume]
role_arn = toad
credential_source = Ec2InstanceMetadata`)), comment)
					}
				})

				It("requires assume role to be set without credentials", func() {
					command, _ := CreateValidCommandWithSecrets("1.2.3.4", "1.2.3.4", "us-west-2", "test_role", "test", "test")
					_, _, err := command.CreateVM()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Assume Role only works when using an instance profile for AWS authentication"))
				})
			})

			It("returns a stateFile with VM details upon success", func() {
				command, _ := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
				status, stateInfo, err := command.CreateVM()
				Expect(err).ToNot(HaveOccurred())
				Expect(status).To(Equal(vmmanagers.Success))
				Expect(stateInfo.IAAS).To(Equal("aws"))
				Expect(stateInfo.ID).To(Equal(vmID))
			})

			When("vm in stateFile is terminated", func() {
				It("execute as if no vm has been deployed", func() {
					command, runner := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
					runner.ExecuteWithEnvVarsReturnsOnCall(0, bytes.NewBufferString("terminated\r\n"), nil, nil)
					runner.ExecuteWithEnvVarsReturnsOnCall(1, bytes.NewBufferString(fmt.Sprintf("\"%s\"\r\n", vmID)), nil, nil)
					runner.ExecuteWithEnvVarsReturnsOnCall(2, bytes.NewBufferString("null\r\n"), bytes.NewBufferString("TestStatus: pending creation"), nil)
					runner.ExecuteWithEnvVarsReturnsOnCall(3, bytes.NewBufferString("\"vol-0cf5b911680a78bb9\"\r\n"), bytes.NewBufferString("TestStatus: volume created"), nil)
					runner.ExecuteWithEnvVarsReturnsOnCall(5, bytes.NewBufferString("\"eipalloc-18643c24\"\r\n"), nil, nil)
					runner.ExecuteWithEnvVarsReturnsOnCall(8, bytes.NewBufferString("\"stopping\"\r\n"), nil, nil)
					runner.ExecuteWithEnvVarsReturnsOnCall(9, bytes.NewBufferString("\"stopped\"\r\n"), nil, nil)

					command.State = vmmanagers.StateInfo{
						IAAS: "aws",
						ID:   vmID,
					}

					status, stateInfo, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal(vmmanagers.Success))
					Expect(stateInfo.IAAS).To(Equal("aws"))
					Expect(stateInfo.ID).To(Equal(vmID))
				})
			})

			When("vm in stateFile does not exist", func() {
				It("execute as if no vm has been deployed", func() {
					command, runner := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
					runner.ExecuteWithEnvVarsReturnsOnCall(0, bytes.NewBufferString("[]\n"), nil, nil)
					runner.ExecuteWithEnvVarsReturnsOnCall(1, bytes.NewBufferString(fmt.Sprintf("\"%s\"\r\n", vmID)), nil, nil)
					runner.ExecuteWithEnvVarsReturnsOnCall(2, bytes.NewBufferString("null\r\n"), bytes.NewBufferString("TestStatus: pending creation"), nil)
					runner.ExecuteWithEnvVarsReturnsOnCall(3, bytes.NewBufferString("\"vol-0cf5b911680a78bb9\"\r\n"), bytes.NewBufferString("TestStatus: volume created"), nil)
					runner.ExecuteWithEnvVarsReturnsOnCall(5, bytes.NewBufferString("\"eipalloc-18643c24\"\r\n"), nil, nil)
					runner.ExecuteWithEnvVarsReturnsOnCall(8, bytes.NewBufferString("\"stopping\"\r\n"), nil, nil)
					runner.ExecuteWithEnvVarsReturnsOnCall(9, bytes.NewBufferString("\"stopped\"\r\n"), nil, nil)

					command.State = vmmanagers.StateInfo{
						IAAS: "aws",
						ID:   vmID,
					}

					status, stateInfo, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal(vmmanagers.Success))
					Expect(stateInfo.IAAS).To(Equal("aws"))
					Expect(stateInfo.ID).To(Equal(vmID))
				})
			})

			DescribeTable("uses the correct AMI for the region", func(region, ami string) {
				command, runner := createValidCommand("1.2.3.4", "10.10.10.10", region)

				_, _, err := command.CreateVM()
				Expect(err).ToNot(HaveOccurred())

				_, args := runner.ExecuteWithEnvVarsArgsForCall(0)
				Expect(args).To(ContainElement(ami))
			},
				Entry("for us-east-1", "us-east-1", "ami-63b6961c"),
				Entry("for us-west-1", "us-west-1", "ami-19a9497a"),
			)

			When("vm already exists in the state file", func() {
				It("returns exist status, doesn't make additional CLI calls, and exits 0", func() {
					command, runner := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
					runner.ExecuteWithEnvVarsReturnsOnCall(0, bytes.NewBufferString(`running\r\n`), nil, nil)

					command.State = vmmanagers.StateInfo{
						IAAS: "aws",
						ID:   vmID,
					}

					status, state, err := command.CreateVM()
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal(vmmanagers.Exist))

					numCLICalls := 1
					Expect(runner.ExecuteWithEnvVarsCallCount()).To(BeEquivalentTo(numCLICalls))

					_, actualArgs := runner.ExecuteWithEnvVarsArgsForCall(0)
					Expect(actualArgs).To(matchers.OrderedConsistOf(
						"ec2",
						"describe-instances",
						"--instance-ids",
						vmID,
						"--query",
						"Reservations[*].Instances[*].State.Name",
					))

					Expect(state.IAAS).To(Equal("aws"))
					Expect(state.ID).To(Equal(vmID))

				})

				Describe("failure cases", func() {
					When("id is malformed", func() {
						It("returns a nice error message", func() {
							command, runner := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
							runner.ExecuteWithEnvVarsReturnsOnCall(0, bytes.NewBufferString(""), bytes.NewBufferString("InvalidInstanceID.Malformed"), errors.New(""))

							command.State = vmmanagers.StateInfo{
								IAAS: "aws",
								ID:   vmID,
							}

							status, _, err := command.CreateVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(MatchRegexp("instance ID .* is malformed. Please check your statefile and try again"))
							Expect(status).To(Equal(vmmanagers.Unknown))

							numCLICalls := 1
							Expect(runner.ExecuteWithEnvVarsCallCount()).To(BeEquivalentTo(numCLICalls))
						})
					})

					When("vm is not exist", func() {
						It("returns a nice error message", func() {
							command, runner := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
							runner.ExecuteWithEnvVarsReturnsOnCall(0, bytes.NewBufferString(""), bytes.NewBufferString("InvalidInstanceID.NotFound"), errors.New(""))

							command.State = vmmanagers.StateInfo{
								IAAS: "aws",
								ID:   vmID,
							}

							status, _, err := command.CreateVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Could not find VM with ID \"" + vmID + "\".\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile."))
							Expect(status).To(Equal(vmmanagers.Unknown))

							numCLICalls := 1
							Expect(runner.ExecuteWithEnvVarsCallCount()).To(BeEquivalentTo(numCLICalls))
						})
					})
				})
			})

			Describe("failure cases", func() {
				When("external tools fail", func() {
					DescribeTable("prints errors from aws", func(offset int, expectedStatus vmmanagers.Status) {
						command, runner := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
						runner.ExecuteWithEnvVarsReturns(bytes.NewBufferString("null\r\n"), nil, nil)

						runner.ExecuteWithEnvVarsReturnsOnCall(offset, nil, nil, errors.New("some error occurred"))
						status, _, err := command.CreateVM()
						Expect(status).To(Equal(expectedStatus))
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("aws error "))
					},
						Entry("create vm", 0, vmmanagers.Unknown),
						Entry("find root device", 1, vmmanagers.Incomplete),
						Entry("modify root device size", 2, vmmanagers.Incomplete),
						Entry("find public ip address", 3, vmmanagers.Incomplete),
						Entry("assign public ip address", 4, vmmanagers.Incomplete),
						Entry("reboot vm", 5, vmmanagers.Incomplete),
					)
				})

				When("region is not in uri map", func() {
					It("returns that the region could not be found", func() {
						command, _ := createValidCommand("1.2.3.4", "10.10.10.10", "lunar-region")
						_, _, err := command.CreateVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("could not find a image uri for region"))
					})
				})

				When("the image file is not valid YAML", func() {
					It("returns that the yaml is invalid", func() {
						invalidUriFile, err := ioutil.TempFile("", "some*.yaml")
						Expect(err).ToNot(HaveOccurred())
						_, _ = invalidUriFile.WriteString("not valid yaml")
						Expect(invalidUriFile.Close()).ToNot(HaveOccurred())

						command, _ := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
						command.ImageYaml = invalidUriFile.Name()

						_, _, err = command.CreateVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("cannot unmarshal"))
					})
				})

				When("image file is not a yaml file", func() {
					var command *vmmanagers.AWSVMManager
					BeforeEach(func() {
						command, _ = createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
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
						command, _ := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
						command.ImageYaml = "does-not-exist.yml"

						_, _, err := command.CreateVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("open does-not-exist.yml"))
					})
				})

				When("the state file has an invalid IAAS", func() {
					It("prints error", func() {
						command, runner := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
						runner.ExecuteWithEnvVarsReturns(nil, nil, errors.New("some error occurred"))

						command.State = vmmanagers.StateInfo{
							IAAS: "gcp",
						}
						status, _, err := command.CreateVM()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("authentication file provided is for aws, while the state file is for "))
						Expect(status).To(Equal(vmmanagers.Unknown))
					})
				})
			})
		})

		DescribeTable("errors when required params are missing", func(param string) {
			invalidConfig := &vmmanagers.OpsmanConfigFilePayload{
				OpsmanConfig: vmmanagers.OpsmanConfig{
					AWS: &vmmanagers.AWSConfig{},
				}}

			command, _ := createCommand("")
			command.Config = invalidConfig

			_, _, err := command.CreateVM()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Field validation for '%s' failed on the 'required' tag", param)))
		},
			Entry("requires Region", "Region"),
			Entry("requires VPCSubnetId", "VPCSubnetId"),
			Entry("requires KeyPairName", "KeyPairName"),
			Entry("requires IAMInstanceProfileName", "IAMInstanceProfileName"),
		)

		It("requires at least public IP or private IP to be set", func() {
			command, _ := createValidCommand("", "", "us-west-2")
			_, _, err := command.CreateVM()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("PublicIP and/or PrivateIP must be set"))
		})

		When("setting a subnet ID", func() {
			It("uses security groups IDs and deprecates on a security_group_id", func() {
				command, _ := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
				command.Config.OpsmanConfig.AWS.SecurityGroupIdDEPRECATED = ""
				command.Config.OpsmanConfig.AWS.SecurityGroupIds = nil
				_, _, err := command.CreateVM()
				Expect(err).To(HaveOccurred())

				command, _ = createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
				command.Config.OpsmanConfig.AWS.SecurityGroupIdDEPRECATED = "sg-123"
				command.Config.OpsmanConfig.AWS.SecurityGroupIds = nil
				_, _, err = command.CreateVM()
				Expect(err).ToNot(HaveOccurred())

				command, _ = createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
				command.Config.OpsmanConfig.AWS.SecurityGroupIdDEPRECATED = ""
				command.Config.OpsmanConfig.AWS.SecurityGroupIds = []string{"sg-123"}
				_, _, err = command.CreateVM()
				Expect(err).ToNot(HaveOccurred())

				command, _ = createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
				command.Config.OpsmanConfig.AWS.SecurityGroupIdDEPRECATED = "sg-123"
				command.Config.OpsmanConfig.AWS.SecurityGroupIds = []string{"sg-123"}
				_, _, err = command.CreateVM()
				Expect(err).To(HaveOccurred())
			})

			It("uses the security group id", func() {
				command, runner := createValidCommand("1.2.3.4", "10.10.10.10", "us-west-2")
				command.Config.OpsmanConfig.AWS.SecurityGroupIdDEPRECATED = "sg-123"
				command.Config.OpsmanConfig.AWS.SecurityGroupIds = nil
				_, _, err := command.CreateVM()
				Expect(err).ToNot(HaveOccurred())

				Expect(runner.ExecuteWithEnvVarsCallCount()).To(BeNumerically(">", 0))
				_, actualArgs := runner.ExecuteWithEnvVarsArgsForCall(0)
				Expect(actualArgs).To(ContainElement("sg-123"))
			})
		})

		It("defaulting any missing optional params", func() {
			command, runner := createCommand(`
opsman-configuration:
  aws:
    version: 1.2.3-build.4
    access_key_id: some-key-id
    secret_access_key: some-key-secret
    region: us-west-1
    vpc_subnet_id: awesome-subnet
    security_group_id: sg-awesome
    key_pair_name: superuser
    iam_instance_profile_name: awesome-profile
    public_ip: 1.2.3.4
    private_ip: 1.2.3.4
`)
			runner.ExecuteWithEnvVarsReturnsOnCall(4, bytes.NewBufferString("some-id\r\n"), nil, nil)
			runner.ExecuteWithEnvVarsReturnsOnCall(9, bytes.NewBufferString("stopped\r\n"), nil, nil)

			_, _, err := command.CreateVM()
			Expect(err).ToNot(HaveOccurred())

			_, args := runner.ExecuteWithEnvVarsArgsForCall(0)
			Expect(args).To(ContainElement(MatchRegexp("ops-manager-vm")))
			Expect(args).To(ContainElement(MatchRegexp("m5\\.large")))

			_, args = runner.ExecuteWithEnvVarsArgsForCall(2)
			Expect(args).To(ContainElement(MatchRegexp("200")))

		})
	})

	Context("delete vm", func() {
		var (
			state = vmmanagers.StateInfo{
				IAAS: "aws",
				ID:   "i-somevmid",
			}
		)

		Context("with a valid config", func() {
			Describe("DeleteVM", func() {
				It("calls aws with correct cli arguments", func() {
					expectedArgs := [][]string{
						{
							`ec2`, `describe-instances`,
							`--instance-ids`, `i-somevmid`,
							`--query`, `Reservations[*].Instances[*].State.Name`,
						},
						{
							`ec2`, `terminate-instances`,
							`--instance-ids`, `i-somevmid`,
						},
						{
							"ec2", "describe-instances",
							"--instance-ids", "i-somevmid",
							"--query", "Reservations[*].Instances[*].State.Name",
						},
						{
							"ec2", "describe-instances",
							"--instance-ids", "i-somevmid",
							"--query", "Reservations[*].Instances[*].State.Name",
						},
					}

					command, runner := createValidCommand("1.1.1.1", "1.1.1.1", "us-west-2")
					command.State = state

					runner.ExecuteWithEnvVarsReturnsOnCall(1, bytes.NewBufferString("terminating"), nil, nil)
					runner.ExecuteWithEnvVarsReturnsOnCall(2, bytes.NewBufferString("terminated"), nil, nil)

					err := command.DeleteVM()
					Expect(err).ToNot(HaveOccurred())

					Expect(runner.ExecuteWithEnvVarsCallCount()).To(Equal(3))
					for i, actualArgs := range runner.Invocations()["ExecuteWithEnvVars"] {
						Expect(actualArgs[1]).To(matchers.OrderedConsistOf(expectedArgs[i]))
					}

				})

				It("calls aws with the correct environment variables", func() {
					command, runner := createValidCommand("1.1.1.1", "1.1.1.1", "us-west-2")
					command.State = state
					err := command.DeleteVM()
					Expect(err).ToNot(HaveOccurred())

					Expect(runner.ExecuteWithEnvVarsCallCount()).To(Equal(202))
					env, _ := runner.ExecuteWithEnvVarsArgsForCall(0)
					Expect(env).Should(ContainElement(`AWS_ACCESS_KEY_ID=some-key-id`))
					Expect(env).Should(ContainElement(`AWS_SECRET_ACCESS_KEY=some-key-secret`))
					Expect(env).Should(ContainElement(`AWS_DEFAULT_REGION=us-west-2`))
				})

				Describe("failure cases", func() {
					When("external tools fail", func() {
						It("prints errors from aws", func() {
							command, runner := createValidCommand("1.1.1.1", "1.1.1.1", "us-west-2")
							command.State = state
							runner.ExecuteWithEnvVarsReturnsOnCall(0, nil, nil, errors.New("some error occurred"))
							err := command.DeleteVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("aws error "))
						})
					})

					When("vm specified in the state file does not exist", func() {
						It("returns an error", func() {
							command, runner := createValidCommand("1.1.1.1", "1.1.1.1", "us-west-2")
							runner.ExecuteWithEnvVarsReturnsOnCall(1, bytes.NewBufferString(""), bytes.NewBufferString("InvalidInstanceID.NotFound"), errors.New("vm does not exist"))

							command.State = vmmanagers.StateInfo{
								IAAS: "aws",
								ID:   "invalid-id",
							}

							err := command.DeleteVM()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("vm does not exist\n       Could not find VM with ID \"invalid-id\".\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile."))

							Expect(runner.ExecuteWithEnvVarsCallCount()).To(BeEquivalentTo(2))
						})
					})
				})
			})
		})

		DescribeTable("errors when required params are missing", func(param string) {
			invalidConfig := &vmmanagers.OpsmanConfigFilePayload{
				OpsmanConfig: vmmanagers.OpsmanConfig{
					AWS: &vmmanagers.AWSConfig{},
				}}

			command := vmmanagers.NewAWSVMManager(ioutil.Discard, ioutil.Discard, invalidConfig, "", state, nil, time.Millisecond)

			err := command.DeleteVM()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Field validation for '%s' failed on the 'required' tag", param)))
		},
			Entry("requires Region", "Region"),
		)

		When("using instance profiles", func() {
			It("requires assume role to be set without credentials", func() {
				command, _ := CreateValidCommandWithSecrets("1.2.3.4", "1.2.3.4", "us-west-2", "test_role", "test", "test")
				command.State = state
				err := command.DeleteVM()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Assume Role only works when using an instance profile for AWS authentication"))
			})
		})

		Context("with an invalid iaas", func() {
			It("prints error", func() {
				command, runner := createValidCommand("1.1.1.1", "1.1.1.1", "us-west-2")
				runner.ExecuteWithEnvVarsReturns(nil, nil, errors.New("some error occurred"))

				command.State = vmmanagers.StateInfo{
					IAAS: "gcp",
				}
				err := command.DeleteVM()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("authentication file provided is for aws, while the state file is for "))
			})
		})
	})

	testIAASForPropertiesInExampleFile("AWS")
})

func makeEnvironmentMap(values []string) map[string]string {
	m := make(map[string]string, len(values))
	for _, val := range values {
		elements := strings.SplitN(val, "=", 2)
		if len(elements) <= 1 {
			panic("unexpected environment string without an equals sign")
		}
		m[elements[0]] = elements[1]
	}
	return m
}

func happyPathAWSRunnerStubFunc(vmID string) func() (*bytes.Buffer, *bytes.Buffer, error) {
	var executeCallCount int64
	return func() (*bytes.Buffer, *bytes.Buffer, error) {
		defer atomic.AddInt64(&executeCallCount, 1)
		switch executeCallCount {
		case 0:
			return bytes.NewBufferString(fmt.Sprintf("%q\r\n", vmID)), nil, nil
		case 1:
			return bytes.NewBufferString("null\r\n"), bytes.NewBufferString("TestStatus: pending creation"), nil
		case 2:
			return bytes.NewBufferString("\"vol-0cf5b911680a78bb9\"\r\n"), bytes.NewBufferString("TestStatus: volume created"), nil
		case 3:
			return nil, nil, nil
		case 4:
			return bytes.NewBufferString("\"eipalloc-18643c24\"\r\n"), nil, nil
		case 5, 6:
			return nil, nil, nil
		case 7:
			return bytes.NewBufferString("\"stopping\"\r\n"), nil, nil
		case 8:
			return bytes.NewBufferString("\"stopped\"\r\n"), nil, nil
		case 9, 10:
			return nil, nil, nil
		default:
			panic("stub for nth call not implemented")
		}
	}
}
