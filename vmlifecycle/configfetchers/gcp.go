package configfetchers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"google.golang.org/api/compute/v1"
)

type GCPConfigFetcher struct {
	state       *vmmanagers.StateInfo
	credentials *Credentials
	service     *compute.Service
}

func NewGCPConfigFetcher(state *vmmanagers.StateInfo, creds *Credentials, service *compute.Service) *GCPConfigFetcher {
	return &GCPConfigFetcher{
		state:       state,
		credentials: creds,
		service:     service,
	}
}

func (g *GCPConfigFetcher) FetchConfig() (*vmmanagers.OpsmanConfigFilePayload, error) {
	instance, err := g.service.Instances.Get(
		g.credentials.GCP.Project,
		g.credentials.GCP.Zone,
		g.state.ID,
	).Do()
	if err != nil {
		return nil, fmt.Errorf("could not fetch instance data: %s", err)
	}

	if len(instance.Disks) == 0 {
		return nil, errors.New("expected a boot disk to be attached to the VM")
	}

	if len(instance.NetworkInterfaces) == 0 {
		return nil, errors.New("expected a network interface to be attached to the VM")
	}

	var scopes []string
	if len(instance.ServiceAccounts) > 0 {
		scopes = instance.ServiceAccounts[0].Scopes
	}

	tags := ""
	if instance.Tags != nil {
		tags = strings.Join(instance.Tags.Items, ",")
	}

	splitMachineTypeURL := strings.Split(instance.MachineType, "/")
	machineType, err := g.service.MachineTypes.Get(
		g.credentials.GCP.Project,
		g.credentials.GCP.Zone,
		splitMachineTypeURL[len(splitMachineTypeURL)-1],
	).Do()
	if err != nil {
		return nil, fmt.Errorf("could not fetch machine type data: %s", err)
	}

	splitDiskURL := strings.Split(instance.Disks[0].Source, "/")
	disk, err := g.service.Disks.Get(
		g.credentials.GCP.Project,
		g.credentials.GCP.Zone,
		splitDiskURL[len(splitDiskURL)-1],
	).Do()
	if err != nil {
		return nil, fmt.Errorf("could not fetch disk data: %s", err)
	}

	publicIP := ""
	if len(instance.NetworkInterfaces[0].AccessConfigs) > 0 {
		publicIP = instance.NetworkInterfaces[0].AccessConfigs[0].NatIP
	}

	serviceAccount := ""
	serviceAccountName := "((gcp-service-account-name))"
	if g.credentials.GCP.ServiceAccount != "" {
		serviceAccount = "((gcp-service-account-json))"
		serviceAccountName = ""
	}

	return &vmmanagers.OpsmanConfigFilePayload{
		OpsmanConfig: vmmanagers.OpsmanConfig{
			GCP: &vmmanagers.GCPConfig{
				GCPCredential: vmmanagers.GCPCredential{
					ServiceAccount:     serviceAccount,
					ServiceAccountName: serviceAccountName,
					Project:            g.credentials.GCP.Project,
					Region:             strings.Join(strings.Split(g.credentials.GCP.Zone, "-")[0:2], "-"),
					Zone:               g.credentials.GCP.Zone,
				},
				VpcSubnet:    strings.TrimPrefix(instance.NetworkInterfaces[0].Subnetwork, "https://www.googleapis.com/compute/v1/"),
				PublicIP:     publicIP,
				PrivateIP:    instance.NetworkInterfaces[0].NetworkIP,
				VMName:       instance.Name,
				Tags:         tags,
				CPU:          fmt.Sprintf("%d", machineType.GuestCpus),
				Memory:       fmt.Sprintf("%dMB", machineType.MemoryMb),
				BootDiskSize: fmt.Sprintf("%dGB", disk.SizeGb),
				Scopes:       scopes,
				SSHPublicKey: "((ssh-public-key))",
			},
		},
	}, nil
}
