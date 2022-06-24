package configfetchers_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/vmlifecycle/configfetchers"
	"github.com/pivotal-cf/om/vmlifecycle/configfetchers/fakes"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
)

var _ = Describe("aws", func() {
	var (
		state           *vmmanagers.StateInfo
		instancesOutput *ec2.DescribeInstancesOutput
		expectedOutput  *vmmanagers.OpsmanConfigFilePayload
		ec2Client       *fakes.Ec2Client
	)

	When("the api returns valid responses", func() {
		BeforeEach(func() {
			ec2Client = &fakes.Ec2Client{}

			state = &vmmanagers.StateInfo{
				IAAS: "aws",
				ID:   "some-vm-id",
			}

			expectedOutput = &vmmanagers.OpsmanConfigFilePayload{
				OpsmanConfig: vmmanagers.OpsmanConfig{
					AWS: &vmmanagers.AWSConfig{
						AWSCredential: vmmanagers.AWSCredential{
							Region: "some-region",
						},
						VPCSubnetId:            "some-subnet",
						SecurityGroupIds:       []string{"some-security-group", "another-security-group"},
						KeyPairName:            "some-key-pair",
						IAMInstanceProfileName: "some-instance-profile",
						PublicIP:               "1.2.3.4",
						PrivateIP:              "5.6.7.8",
						VMName:                 "opsman-vm",
						BootDiskSize:           "160",
						InstanceType:           "some-instance-type",
					},
				},
			}

			instancesOutput = &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{{
					Instances: []*ec2.Instance{{
						Placement: &ec2.Placement{
							AvailabilityZone: aws.String("current-availability-zone"),
						},
						Tags: []*ec2.Tag{{
							Key:   aws.String("Name"),
							Value: aws.String("opsman-vm"),
						}},
						SubnetId: aws.String("some-subnet"),
						SecurityGroups: []*ec2.GroupIdentifier{{
							GroupId: aws.String("some-security-group"),
						}, {
							GroupId: aws.String("another-security-group"),
						}},
						KeyName: aws.String("some-key-pair"),
						IamInstanceProfile: &ec2.IamInstanceProfile{
							Arn: aws.String("arn:aws:iam::473893145203:instance-profile/some-instance-profile"),
						},
						PublicIpAddress:  aws.String("1.2.3.4"),
						PrivateIpAddress: aws.String("5.6.7.8"),
						InstanceType:     aws.String("some-instance-type"),
						BlockDeviceMappings: []*ec2.InstanceBlockDeviceMapping{{
							Ebs: &ec2.EbsInstanceBlockDevice{
								VolumeId: aws.String("some-volume-id"),
							},
						}},
					}},
				}},
			}

			ec2Client.DescribeInstancesStub = func(input *ec2.DescribeInstancesInput) (output *ec2.DescribeInstancesOutput, e error) {
				Expect(input).To(Equal(&ec2.DescribeInstancesInput{
					InstanceIds: []*string{aws.String("some-vm-id")},
				}))

				return instancesOutput, nil
			}

			ec2Client.DescribeVolumesStub = func(input *ec2.DescribeVolumesInput) (output *ec2.DescribeVolumesOutput, e error) {
				Expect(input).To(Equal(&ec2.DescribeVolumesInput{
					VolumeIds: []*string{aws.String("some-volume-id")},
				}))

				return &ec2.DescribeVolumesOutput{
					Volumes: []*ec2.Volume{{
						Size: aws.Int64(160),
					}},
				}, nil
			}
		})

		When("aws credentials are not passed in and an instance profile is available", func() {
			It("creates an opsman.yml that does't include aws credentials", func() {
				creds := &configfetchers.Credentials{
					AWS: &vmmanagers.AWSCredential{
						Region: "some-region",
					},
				}

				fetcher := configfetchers.NewAWSConfigFetcher(state, creds, ec2Client)

				output, err := fetcher.FetchConfig()
				Expect(err).ToNot(HaveOccurred())

				Expect(output).To(Equal(expectedOutput))
			})
		})

		When("aws credentials are passed in", func() {
			It("creates an opsman.yml that includes placeholders for aws credentials", func() {
				extendedExpectedOutput := expectedOutput

				extendedExpectedOutput.OpsmanConfig.AWS.SecretAccessKey = "((secret_access_key))"
				extendedExpectedOutput.OpsmanConfig.AWS.AccessKeyId = "((access_key_id))"

				creds := &configfetchers.Credentials{
					AWS: &vmmanagers.AWSCredential{
						Region:          "some-region",
						AccessKeyId:     "some-access-key",
						SecretAccessKey: "some-secret-access-key",
					},
				}

				fetcher := configfetchers.NewAWSConfigFetcher(state, creds, ec2Client)

				output, err := fetcher.FetchConfig()
				Expect(err).ToNot(HaveOccurred())

				Expect(output).To(Equal(extendedExpectedOutput))
			})
		})

		When("the ops man VM does not have a public IP", func() {
			It("creates an opsman.yml that does't include a public IP", func() {
				creds := &configfetchers.Credentials{
					AWS: &vmmanagers.AWSCredential{
						Region: "some-region",
					},
				}

				instancesOutput.Reservations[0].Instances[0].PublicIpAddress = nil

				fetcher := configfetchers.NewAWSConfigFetcher(state, creds, ec2Client)

				output, err := fetcher.FetchConfig()
				Expect(err).ToNot(HaveOccurred())

				Expect(output.OpsmanConfig.AWS.PublicIP).To(Equal(""))
			})
		})
	})

	When("the api returns partial responses", func() {
		BeforeEach(func() {
			ec2Client = &fakes.Ec2Client{}

			state = &vmmanagers.StateInfo{
				IAAS: "aws",
				ID:   "some-vm-id",
			}

			ec2Client.DescribeInstancesReturns(&ec2.DescribeInstancesOutput{Reservations: nil}, nil)
		})

		It("returns an error", func() {
			creds := &configfetchers.Credentials{
				AWS: &vmmanagers.AWSCredential{
					Region:          "some-region",
					AccessKeyId:     "some-access-key",
					SecretAccessKey: "some-secret-access-key",
				},
			}

			fetcher := configfetchers.NewAWSConfigFetcher(state, creds, ec2Client)

			_, err := fetcher.FetchConfig()
			Expect(err).To(HaveOccurred())
		})
	})
})
