package configfetchers

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"strconv"
	"strings"
)

//go:generate counterfeiter -o ./fakes/ec2Client.go --fake-name Ec2Client . Ec2Client
type Ec2Client interface {
	DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error)
}

type AWSConfigFetcher struct {
	state       *vmmanagers.StateInfo
	credentials *Credentials
	ec2Client   Ec2Client
}

func NewAWSConfigFetcher(state *vmmanagers.StateInfo, credentials *Credentials, client Ec2Client) *AWSConfigFetcher {
	return &AWSConfigFetcher{
		state:       state,
		credentials: credentials,
		ec2Client:   client,
	}
}

func (a *AWSConfigFetcher) FetchConfig() (*vmmanagers.OpsmanConfigFilePayload, error) {
	instancesOutput, volumeOutput, err := a.fetchDataFromAWS()
	if err != nil {
		return &vmmanagers.OpsmanConfigFilePayload{}, err
	}

	return a.createConfig(instancesOutput, volumeOutput), nil
}

func (a *AWSConfigFetcher) fetchDataFromAWS() (*ec2.DescribeInstancesOutput, *ec2.DescribeVolumesOutput, error) {
	instancesOutput, err := a.ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{&a.state.ID},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not get required data from the instance %s", err)
	}

	if len(instancesOutput.Reservations) == 0 {
		return nil, nil, fmt.Errorf("no instances could be found with instance ID %s", a.state.ID)
	}

	volumeID := instancesOutput.Reservations[0].Instances[0].BlockDeviceMappings[0].Ebs.VolumeId
	volumeOutput, err := a.ec2Client.DescribeVolumes(&ec2.DescribeVolumesInput{
		VolumeIds: []*string{volumeID},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not get volume data %s", err)
	}

	return instancesOutput, volumeOutput, nil
}

func (a *AWSConfigFetcher) createConfig(output *ec2.DescribeInstancesOutput, volumesOutput *ec2.DescribeVolumesOutput) *vmmanagers.OpsmanConfigFilePayload {
	instance := output.Reservations[0].Instances[0]

	var vmName string
	for _, tag := range instance.Tags {
		if *tag.Key == "Name" {
			vmName = *tag.Value
		}
	}

	var securityGroups []string
	for _, securityGroup := range instance.SecurityGroups {
		securityGroups = append(securityGroups, *securityGroup.GroupId)
	}

	splitARN := strings.Split(*instance.IamInstanceProfile.Arn, "/")

	opsmanConfig := &vmmanagers.OpsmanConfigFilePayload{
		OpsmanConfig: vmmanagers.OpsmanConfig{
			AWS: &vmmanagers.AWSConfig{
				AWSCredential: vmmanagers.AWSCredential{
					Region: a.credentials.AWS.Region,
				},
				VMName:                 vmName,
				VPCSubnetId:            *instance.SubnetId,
				SecurityGroupIds:       securityGroups,
				KeyPairName:            *instance.KeyName,
				IAMInstanceProfileName: splitARN[1],
				PublicIP:               aws.StringValue(instance.PublicIpAddress),
				PrivateIP:              *instance.PrivateIpAddress,
				InstanceType:           *instance.InstanceType,
				BootDiskSize:           strconv.Itoa(int(*volumesOutput.Volumes[0].Size)),
				BootDiskType:           *volumesOutput.Volumes[0].VolumeType,
			},
		},
	}

	if a.credentials.AWS.AccessKeyId != "" && a.credentials.AWS.SecretAccessKey != "" {
		opsmanConfig.OpsmanConfig.AWS.AWSCredential.AccessKeyId = "((access_key_id))"
		opsmanConfig.OpsmanConfig.AWS.AWSCredential.SecretAccessKey = "((secret_access_key))"
	}

	return opsmanConfig
}