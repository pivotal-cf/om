package configfetchers

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-09-01/network"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"net/url"
	"strings"
)

//go:generate counterfeiter -o ./fakes/azureVMClient.go --fake-name AzureVMClient . AzureVMClient
type AzureVMClient interface {
	Get(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error)
}

//go:generate counterfeiter -o ./fakes/azureNetworkClient.go --fake-name AzureNetworkClient . AzureNetworkClient
type AzureNetworkClient interface {
	Get(ctx context.Context, resourceGroupName string, networkInterfaceName string, expand string) (result network.Interface, err error)
}

//go:generate counterfeiter -o ./fakes/azureIPClient.go --fake-name AzureIPClient . AzureIPClient
type AzureIPClient interface {
	Get(ctx context.Context, resourceGroupName string, publicIPAddressName string, expand string) (result network.PublicIPAddress, err error)
}

//go:generate counterfeiter -o ./fakes/azureImageClient.go --fake-name AzureImageClient . AzureImageClient
type AzureImageClient interface {
	Get(ctx context.Context, resourceGroupName string, imageName string, expand string) (result compute.Image, err error)
}

type AzureConfigFetcher struct {
	state         *vmmanagers.StateInfo
	credentials   *Credentials
	vmClient      AzureVMClient
	networkClient AzureNetworkClient
	ipClient      AzureIPClient
	imageClient   AzureImageClient
}

func NewAzureConfigFetcher(
	state *vmmanagers.StateInfo,
	credentials *Credentials,
	vmClient AzureVMClient,
	networkClient AzureNetworkClient,
	ipClient AzureIPClient,
	imageClient AzureImageClient) *AzureConfigFetcher {

	return &AzureConfigFetcher{
		state:         state,
		credentials:   credentials,
		vmClient:      vmClient,
		networkClient: networkClient,
		ipClient:      ipClient,
		imageClient:   imageClient,
	}
}

func getLastSection(id string) string {
	parts := strings.Split(id, "/")
	return parts[len(parts)-1]
}

func (a *AzureConfigFetcher) FetchConfig() (*vmmanagers.OpsmanConfigFilePayload, error) {
	resourceGroup := a.credentials.Azure.ResourceGroup

	vm, err := a.vmClient.Get(context.Background(), resourceGroup, a.state.ID, compute.InstanceView)
	if err != nil {
		return nil, fmt.Errorf("could not get the vm object from Azure: %s", err)
	}

	networkInterface, err := a.parseNetworkInterface(vm)
	if err != nil {
		return nil, err
	}

	ipConfigurations := *networkInterface.IPConfigurations
	if len(ipConfigurations) == 0 {
		return nil, fmt.Errorf("no ip configurations found for vm '%s'", a.state.ID)
	}

	ipConfiguration := ipConfigurations[0]
	publicIPAddress, err := a.parsePublicIP(ipConfiguration)
	if err != nil {
		return nil, err
	}

	storageAccount, container, err := a.parsePropertiesFromImage(vm)
	if err != nil {
		return nil, err
	}

	useManagedDisk := "false"
	storageSKU := ""
	if vm.VirtualMachineProperties.StorageProfile.OsDisk.ManagedDisk != nil {
		useManagedDisk = "true"
		storageSKU = string(vm.VirtualMachineProperties.StorageProfile.OsDisk.ManagedDisk.StorageAccountType)
	}

	return &vmmanagers.OpsmanConfigFilePayload{
		OpsmanConfig: vmmanagers.OpsmanConfig{
			Azure: &vmmanagers.AzureConfig{
				AzureCredential: vmmanagers.AzureCredential{
					TenantID:       a.credentials.Azure.TenantID,
					SubscriptionID: a.credentials.Azure.SubscriptionID,
					ClientID:       a.credentials.Azure.ClientID,
					ClientSecret:   "((client-secret))",
					Location:       *vm.Location,
					ResourceGroup:  resourceGroup,
					StorageAccount: storageAccount,
					StorageKey:     "((storage-account-key))",
				},
				CloudName:      a.credentials.Azure.CloudName,
				Container:      container,
				SubnetID:       *ipConfiguration.Subnet.ID,
				NSG:            getLastSection(*networkInterface.NetworkSecurityGroup.ID),
				VMName:         *vm.Name,
				SSHPublicKey:   "((ssh-public-key))",
				BootDiskSize:   fmt.Sprintf("%d", *vm.StorageProfile.OsDisk.DiskSizeGB),
				PrivateIP:      *ipConfiguration.PrivateIPAddress,
				PublicIP:       publicIPAddress,
				UseManagedDisk: useManagedDisk,
				StorageSKU:     storageSKU,
				VMSize:         string(vm.VirtualMachineProperties.HardwareProfile.VMSize),
			},
		},
	}, nil
}

func (a *AzureConfigFetcher) parsePropertiesFromImage(vm compute.VirtualMachine) (string, string, error) {
	var blobURLToParse string
	if vm.VirtualMachineProperties.StorageProfile.OsDisk.Image != nil {
		blobURLToParse = *vm.VirtualMachineProperties.StorageProfile.OsDisk.Image.URI
	} else {
		image, err := a.imageClient.Get(context.Background(), a.credentials.Azure.ResourceGroup, getLastSection(*vm.VirtualMachineProperties.StorageProfile.ImageReference.ID), "")
		if err != nil {
			return "", "", fmt.Errorf("could not get the image object from Azure: %s", err)
		}

		blobURLToParse = *image.ImageProperties.StorageProfile.OsDisk.BlobURI
	}

	blobURL, err := url.Parse(blobURLToParse)
	if err != nil {
		return "", "", fmt.Errorf("invalid blob URL returned from Azure: '%s'", blobURLToParse)
	}

	splitBlobPath := strings.Split(blobURL.Path, "/")
	splitBlobHost := strings.Split(blobURL.Host, ".")
	if len(splitBlobPath) < 2 || len(splitBlobHost) < 1 {
		return "", "", fmt.Errorf("could not parse storage account and container from the Azure blob URL: '%s'", blobURLToParse)
	}

	storageAccount := splitBlobHost[0]
	container := splitBlobPath[1]

	return storageAccount, container, nil
}

func (a *AzureConfigFetcher) parsePublicIP(ipConfiguration network.InterfaceIPConfiguration) (string, error) {
	var publicIPAddress string
	if ipConfiguration.PublicIPAddress != nil {
		address, err := a.ipClient.Get(context.Background(), a.credentials.Azure.ResourceGroup, getLastSection(*ipConfiguration.InterfaceIPConfigurationPropertiesFormat.PublicIPAddress.ID), "")
		if err != nil {
			return "", fmt.Errorf("could not get the public IP publicIPAddress object from Azure: %s", err)
		}

		if address.IPAddress != nil {
			publicIPAddress = *address.IPAddress
		}
	}

	return publicIPAddress, nil
}

func (a *AzureConfigFetcher) parseNetworkInterface(vm compute.VirtualMachine) (network.Interface, error) {
	networkInterfaces := *vm.NetworkProfile.NetworkInterfaces
	if len(networkInterfaces) == 0 {
		return network.Interface{}, fmt.Errorf("no network interface found for vm '%s'", a.state.ID)
	}

	networkInterfaceName := getLastSection(*(networkInterfaces[0].ID))
	networkInterface, err := a.networkClient.Get(context.Background(), a.credentials.Azure.ResourceGroup, networkInterfaceName, "")
	if err != nil {
		return network.Interface{}, fmt.Errorf("could not get the network interface object from Azure: %s", err)
	}
	return networkInterface, nil
}
