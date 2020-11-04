package configfetchers_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-09-01/network"
	"github.com/pivotal-cf/om/vmlifecycle/configfetchers"
	"github.com/pivotal-cf/om/vmlifecycle/configfetchers/fakes"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("azure", func() {
	var (
		state                *vmmanagers.StateInfo
		expectedOutput       *vmmanagers.OpsmanConfigFilePayload
		azureVMClient        *fakes.AzureVMClient
		azureNetworkClient   *fakes.AzureNetworkClient
		azureIPClient        *fakes.AzureIPClient
		azureImageClient     *fakes.AzureImageClient
		fetcher              *configfetchers.AzureConfigFetcher
		location             = "some-location"
		diskSize             = int32(8)
		container            = "some-container"
		storageAccount       = "some-storage-account"
		validBlobURL         = fmt.Sprintf("https://%s.blob.core.windows.net/%s", storageAccount, container)
		nsg                  = "/some/path/to/some-nsg"
		privateIP            = "1.2.3.4"
		publicIP             = "5.6.7.8"
		vmName               = "some-vm-id"
		networkInterfaceName = "some-network-interface"
		imageReferenceID     = "some-image-reference-id"
		publicIPID           = "some-ip-address-id"
		subnet               = "some-subnet-id"
		rg                   = "some-resource-group"
	)

	setValidVMClientStub := func(blobURL string) {
		azureVMClient.GetStub = func(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
			Expect(resourceGroupName).To(Equal(rg))
			Expect(VMName).To(Equal("some-vm-id"))

			return compute.VirtualMachine{
				Name: &vmName,
				VirtualMachineProperties: &compute.VirtualMachineProperties{
					HardwareProfile: &compute.HardwareProfile{
						VMSize: "some-vm-size",
					},
					StorageProfile: &compute.StorageProfile{
						OsDisk: &compute.OSDisk{
							Image: &compute.VirtualHardDisk{
								URI: &blobURL,
							},
							DiskSizeGB: &diskSize,
							ManagedDisk: &compute.ManagedDiskParameters{
								StorageAccountType: "some-storage-sku",
							},
						},
					},
					NetworkProfile: &compute.NetworkProfile{
						NetworkInterfaces: &[]compute.NetworkInterfaceReference{{
							ID: &networkInterfaceName,
						}},
					},
				},
				Location: &location,
			}, nil
		}
	}
	setValidNetworkClientStub := func() {
		azureNetworkClient.GetStub = func(ctx context.Context, resourceGroupName string, networkInterfaceName string, expand string) (result network.Interface, err error) {
			Expect(networkInterfaceName).To(Equal(networkInterfaceName))
			Expect(resourceGroupName).To(Equal(rg))

			return network.Interface{
				InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
					NetworkSecurityGroup: &network.SecurityGroup{
						ID: &nsg,
					},
					IPConfigurations: &[]network.InterfaceIPConfiguration{{
						InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAddress: &privateIP,
							PublicIPAddress: &network.PublicIPAddress{
								ID: &publicIPID,
							},
							Subnet: &network.Subnet{
								ID: &subnet,
							},
						},
					}},
				},
			}, nil
		}
	}
	setValidIPClientStub := func() {
		azureIPClient.GetStub = func(ctx context.Context, resourceGroupName string, publicIPAddressName string, expand string) (result network.PublicIPAddress, err error) {
			Expect(resourceGroupName).To(Equal(rg))
			Expect(publicIPAddressName).To(Equal(publicIPID))
			return network.PublicIPAddress{
				PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
					IPAddress: &publicIP,
				},
			}, nil
		}
	}

	setValidImageClientStub := func(blobURL string) {
		azureImageClient.GetStub = func(ctx context.Context, resourceGroupName string, imageName string, expand string) (result compute.Image, err error) {
			Expect(resourceGroupName).To(Equal(rg))
			Expect(imageName).To(Equal(imageReferenceID))
			return compute.Image{
				ImageProperties: &compute.ImageProperties{
					StorageProfile: &compute.ImageStorageProfile{
						OsDisk: &compute.ImageOSDisk{
							BlobURI: &blobURL,
						},
					},
				},
			}, nil
		}
	}

	BeforeEach(func() {
		azureVMClient = &fakes.AzureVMClient{}
		azureNetworkClient = &fakes.AzureNetworkClient{}
		azureIPClient = &fakes.AzureIPClient{}
		azureImageClient = &fakes.AzureImageClient{}

		state = &vmmanagers.StateInfo{
			IAAS: "azure",
			ID:   "some-vm-id",
		}
	})

	When("the api returns full responses", func() {
		BeforeEach(func() {
			setValidVMClientStub(validBlobURL)
			setValidNetworkClientStub()
			setValidIPClientStub()

			expectedOutput = &vmmanagers.OpsmanConfigFilePayload{
				OpsmanConfig: vmmanagers.OpsmanConfig{
					Azure: &vmmanagers.AzureConfig{
						AzureCredential: vmmanagers.AzureCredential{
							TenantID:       "some-tenant-id",
							SubscriptionID: "some-subscription-id",
							ClientID:       "some-client-id",
							ClientSecret:   "((client-secret))",
							Location:       location,
							ResourceGroup:  "some-resource-group",
							StorageAccount: "some-storage-account",
							StorageKey:     "((storage-account-key))",
						},
						CloudName:      "AzureCloud",
						Container:      container,
						SubnetID:       subnet,
						NSG:            "some-nsg",
						VMName:         vmName,
						SSHPublicKey:   "((ssh-public-key))",
						BootDiskSize:   strconv.Itoa(int(diskSize)),
						PrivateIP:      privateIP,
						PublicIP:       publicIP,
						UseManagedDisk: "true",
						StorageSKU:     "some-storage-sku",
						VMSize:         "some-vm-size",
					},
				},
			}

			fetcher = configfetchers.NewAzureConfigFetcher(
				state,
				&configfetchers.Credentials{
					Azure: &configfetchers.AzureCredentialsWrapper{
						AzureCredential: vmmanagers.AzureCredential{
							TenantID:       "some-tenant-id",
							SubscriptionID: "some-subscription-id",
							ClientID:       "some-client-id",
							ClientSecret:   "some-client-secret",
							ResourceGroup:  "some-resource-group",
						},
						CloudName: "AzureCloud",
					},
				},
				azureVMClient,
				azureNetworkClient,
				azureIPClient,
				azureImageClient,
			)
		})

		It("creates an opsman.yml that does't include azure credentials", func() {
			payload, err := fetcher.FetchConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(payload).To(Equal(expectedOutput))
		})

		When("the vm client returns an error", func() {
			It("returns an error", func() {
				azureVMClient.GetReturns(compute.VirtualMachine{}, errors.New("vm client error"))
				_, err := fetcher.FetchConfig()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("vm client error"))
			})
		})

		When("the network client returns an error", func() {
			It("returns an error", func() {
				azureNetworkClient.GetReturns(network.Interface{}, errors.New("network interface client error"))
				_, err := fetcher.FetchConfig()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("network interface client error"))
			})
		})

		When("the ip client returns an error", func() {
			It("returns an error", func() {
				azureIPClient.GetReturns(network.PublicIPAddress{}, errors.New("public IP client error"))
				_, err := fetcher.FetchConfig()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("public IP client error"))
			})
		})
	})

	When("the api returns incomplete responses", func() {
		BeforeEach(func() {
			fetcher = configfetchers.NewAzureConfigFetcher(
				state,
				&configfetchers.Credentials{
					Azure: &configfetchers.AzureCredentialsWrapper{
						AzureCredential: vmmanagers.AzureCredential{
							TenantID:       "some-tenant-id",
							SubscriptionID: "some-subscription-id",
							ClientID:       "some-client-id",
							ClientSecret:   "some-client-secret",
							ResourceGroup:  "some-resource-group",
						},
						CloudName: "AzureCloud",
					},
				},
				azureVMClient,
				azureNetworkClient,
				azureIPClient,
				azureImageClient,
			)
		})

		It("errors when vm clientVirtualMachineProperties.NetworkProfile.NetworkInterfaces is empty", func() {
			setValidIPClientStub()
			setValidNetworkClientStub()

			azureVMClient.GetStub = func(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
				Expect(resourceGroupName).To(Equal(rg))
				Expect(VMName).To(Equal("some-vm-id"))

				return compute.VirtualMachine{
					Name: &vmName,
					VirtualMachineProperties: &compute.VirtualMachineProperties{
						HardwareProfile: &compute.HardwareProfile{
							VMSize: "some-vm-size",
						},
						StorageProfile: &compute.StorageProfile{
							ImageReference: &compute.ImageReference{
								ID: &imageReferenceID,
							},
							OsDisk: &compute.OSDisk{
								DiskSizeGB: &diskSize,
								ManagedDisk: &compute.ManagedDiskParameters{
									StorageAccountType: "some-storage-sku",
								},
							},
						},
						NetworkProfile: &compute.NetworkProfile{
							NetworkInterfaces: &[]compute.NetworkInterfaceReference{},
						},
					},
					Location: &location,
				}, nil
			}

			_, err := fetcher.FetchConfig()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no network interface found for vm 'some-vm-id'"))
		})

		It("does not return public IP when networkInterface.InterfacePropertiesFormat.IPConfigurations is empty", func() {
			setValidIPClientStub()
			setValidVMClientStub(validBlobURL)
			azureNetworkClient.GetStub = func(ctx context.Context, resourceGroupName string, networkInterfaceName string, expand string) (result network.Interface, err error) {
				Expect(networkInterfaceName).To(Equal(networkInterfaceName))
				Expect(resourceGroupName).To(Equal(rg))

				return network.Interface{
					InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
						NetworkSecurityGroup: &network.SecurityGroup{
							ID: &nsg,
						},
						IPConfigurations: &[]network.InterfaceIPConfiguration{{
							InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
								PrivateIPAddress: &privateIP,
								PublicIPAddress:  nil,
								Subnet: &network.Subnet{
									ID: &subnet,
								},
							},
						}},
					},
				}, nil
			}

			expectedOutput = &vmmanagers.OpsmanConfigFilePayload{
				OpsmanConfig: vmmanagers.OpsmanConfig{
					Azure: &vmmanagers.AzureConfig{
						AzureCredential: vmmanagers.AzureCredential{
							TenantID:       "some-tenant-id",
							SubscriptionID: "some-subscription-id",
							ClientID:       "some-client-id",
							ClientSecret:   "((client-secret))",
							Location:       location,
							ResourceGroup:  "some-resource-group",
							StorageAccount: "some-storage-account",
							StorageKey:     "((storage-account-key))",
						},
						CloudName:      "AzureCloud",
						Container:      container,
						SubnetID:       subnet,
						NSG:            "some-nsg",
						VMName:         vmName,
						SSHPublicKey:   "((ssh-public-key))",
						BootDiskSize:   strconv.Itoa(int(diskSize)),
						PrivateIP:      privateIP,
						PublicIP:       "",
						UseManagedDisk: "true",
						StorageSKU:     "some-storage-sku",
						VMSize:         "some-vm-size",
					},
				},
			}

			payload, err := fetcher.FetchConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(payload).To(Equal(expectedOutput))
		})

		It("errors when ip networkInterface.InterfacePropertiesFormat.IPConfigurations is empty", func() {
			setValidIPClientStub()
			setValidVMClientStub(validBlobURL)
			azureNetworkClient.GetStub = func(ctx context.Context, resourceGroupName string, networkInterfaceName string, expand string) (result network.Interface, err error) {
				Expect(networkInterfaceName).To(Equal(networkInterfaceName))
				Expect(resourceGroupName).To(Equal(rg))

				return network.Interface{
					InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
						NetworkSecurityGroup: &network.SecurityGroup{
							ID: &nsg,
						},
						IPConfigurations: &[]network.InterfaceIPConfiguration{},
					},
				}, nil
			}

			_, err := fetcher.FetchConfig()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no ip configurations found for vm 'some-vm-id'"))
		})

		When("virtualMachineProperties.StorageProfile.OsDisk.Image exists", func() {
			It("errors when blobURL cannot be parsed", func() {
				invalidBlobURL := "%%%"

				setValidNetworkClientStub()
				setValidIPClientStub()
				azureVMClient.GetStub = func(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
					Expect(resourceGroupName).To(Equal(rg))
					Expect(VMName).To(Equal("some-vm-id"))

					return compute.VirtualMachine{
						Name: &vmName,
						VirtualMachineProperties: &compute.VirtualMachineProperties{
							HardwareProfile: &compute.HardwareProfile{
								VMSize: "some-vm-size",
							},
							StorageProfile: &compute.StorageProfile{
								OsDisk: &compute.OSDisk{
									Image: &compute.VirtualHardDisk{
										URI: &invalidBlobURL,
									},
									DiskSizeGB: &diskSize,
									ManagedDisk: &compute.ManagedDiskParameters{
										StorageAccountType: "some-storage-sku",
									},
								},
							},
							NetworkProfile: &compute.NetworkProfile{
								NetworkInterfaces: &[]compute.NetworkInterfaceReference{{
									ID: &networkInterfaceName,
								}},
							},
						},
						Location: &location,
					}, nil
				}

				fetcher = configfetchers.NewAzureConfigFetcher(
					state,
					&configfetchers.Credentials{
						Azure: &configfetchers.AzureCredentialsWrapper{
							AzureCredential: vmmanagers.AzureCredential{
								TenantID:       "some-tenant-id",
								SubscriptionID: "some-subscription-id",
								ClientID:       "some-client-id",
								ClientSecret:   "some-client-secret",
								ResourceGroup:  "some-resource-group",
							},
							CloudName: "AzureCloud",
						},
					},
					azureVMClient,
					azureNetworkClient,
					azureIPClient,
					azureImageClient,
				)

				_, err := fetcher.FetchConfig()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid blob URL returned from Azure: '%%%'"))
			})

			It("errors when blobURL is not in the expected format", func() {
				invalidBlobURL := "https://example.com"

				setValidNetworkClientStub()
				setValidIPClientStub()
				azureVMClient.GetStub = func(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
					Expect(resourceGroupName).To(Equal(rg))
					Expect(VMName).To(Equal("some-vm-id"))

					return compute.VirtualMachine{
						Name: &vmName,
						VirtualMachineProperties: &compute.VirtualMachineProperties{
							HardwareProfile: &compute.HardwareProfile{
								VMSize: "some-vm-size",
							},
							StorageProfile: &compute.StorageProfile{
								OsDisk: &compute.OSDisk{
									Image: &compute.VirtualHardDisk{
										URI: &invalidBlobURL,
									},
									DiskSizeGB: &diskSize,
									ManagedDisk: &compute.ManagedDiskParameters{
										StorageAccountType: "some-storage-sku",
									},
								},
							},
							NetworkProfile: &compute.NetworkProfile{
								NetworkInterfaces: &[]compute.NetworkInterfaceReference{{
									ID: &networkInterfaceName,
								}},
							},
						},
						Location: &location,
					}, nil
				}

				fetcher = configfetchers.NewAzureConfigFetcher(
					state,
					&configfetchers.Credentials{
						Azure: &configfetchers.AzureCredentialsWrapper{
							AzureCredential: vmmanagers.AzureCredential{
								TenantID:       "some-tenant-id",
								SubscriptionID: "some-subscription-id",
								ClientID:       "some-client-id",
								ClientSecret:   "some-client-secret",
								ResourceGroup:  "some-resource-group",
							},
							CloudName: "AzureCloud",
						},
					},
					azureVMClient,
					azureNetworkClient,
					azureIPClient,
					azureImageClient,
				)

				_, err := fetcher.FetchConfig()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not parse storage account and container from the Azure blob URL: 'https://example.com'"))
			})
		})

		When("virtualMachineProperties.StorageProfile.OsDisk.Image is empty", func() {
			It("uses vm.VirtualMachineProperties.StorageProfile.ImageReference.ID with the image client", func() {
				setValidNetworkClientStub()
				setValidIPClientStub()
				setValidImageClientStub(validBlobURL)
				azureVMClient.GetStub = func(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
					Expect(resourceGroupName).To(Equal(rg))
					Expect(VMName).To(Equal("some-vm-id"))

					return compute.VirtualMachine{
						Name: &vmName,
						VirtualMachineProperties: &compute.VirtualMachineProperties{
							HardwareProfile: &compute.HardwareProfile{
								VMSize: "some-vm-size",
							},
							StorageProfile: &compute.StorageProfile{
								ImageReference: &compute.ImageReference{
									ID: &imageReferenceID,
								},
								OsDisk: &compute.OSDisk{
									DiskSizeGB: &diskSize,
									Image:      nil,
									ManagedDisk: &compute.ManagedDiskParameters{
										StorageAccountType: "some-storage-sku",
									},
								},
							},
							NetworkProfile: &compute.NetworkProfile{
								NetworkInterfaces: &[]compute.NetworkInterfaceReference{{
									ID: &networkInterfaceName,
								}},
							},
						},
						Location: &location,
					}, nil
				}

				_, err := fetcher.FetchConfig()
				Expect(err).ToNot(HaveOccurred())
			})

			It("errors when blobURL cannot be parsed", func() {
				invalidBlobURL := "%%%"

				setValidNetworkClientStub()
				setValidIPClientStub()
				setValidImageClientStub(invalidBlobURL)
				azureVMClient.GetStub = func(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
					Expect(resourceGroupName).To(Equal(rg))
					Expect(VMName).To(Equal("some-vm-id"))

					return compute.VirtualMachine{
						Name: &vmName,
						VirtualMachineProperties: &compute.VirtualMachineProperties{
							HardwareProfile: &compute.HardwareProfile{
								VMSize: "some-vm-size",
							},
							StorageProfile: &compute.StorageProfile{
								ImageReference: &compute.ImageReference{
									ID: &imageReferenceID,
								},
								OsDisk: &compute.OSDisk{
									DiskSizeGB: &diskSize,
									ManagedDisk: &compute.ManagedDiskParameters{
										StorageAccountType: "some-storage-sku",
									},
								},
							},
							NetworkProfile: &compute.NetworkProfile{
								NetworkInterfaces: &[]compute.NetworkInterfaceReference{{
									ID: &networkInterfaceName,
								}},
							},
						},
						Location: &location,
					}, nil
				}

				fetcher = configfetchers.NewAzureConfigFetcher(
					state,
					&configfetchers.Credentials{
						Azure: &configfetchers.AzureCredentialsWrapper{
							AzureCredential: vmmanagers.AzureCredential{
								TenantID:       "some-tenant-id",
								SubscriptionID: "some-subscription-id",
								ClientID:       "some-client-id",
								ClientSecret:   "some-client-secret",
								ResourceGroup:  "some-resource-group",
							},
							CloudName: "AzureCloud",
						},
					},
					azureVMClient,
					azureNetworkClient,
					azureIPClient,
					azureImageClient,
				)

				_, err := fetcher.FetchConfig()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid blob URL returned from Azure: '%%%'"))
			})

			It("errors when blobURL is not in the expected format", func() {
				invalidBlobURL := "https://example.com"

				setValidNetworkClientStub()
				setValidIPClientStub()
				setValidImageClientStub(invalidBlobURL)
				azureVMClient.GetStub = func(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
					Expect(resourceGroupName).To(Equal(rg))
					Expect(VMName).To(Equal("some-vm-id"))

					return compute.VirtualMachine{
						Name: &vmName,
						VirtualMachineProperties: &compute.VirtualMachineProperties{
							HardwareProfile: &compute.HardwareProfile{
								VMSize: "some-vm-size",
							},
							StorageProfile: &compute.StorageProfile{
								ImageReference: &compute.ImageReference{
									ID: &imageReferenceID,
								},
								OsDisk: &compute.OSDisk{
									DiskSizeGB: &diskSize,
									ManagedDisk: &compute.ManagedDiskParameters{
										StorageAccountType: "some-storage-sku",
									},
								},
							},
							NetworkProfile: &compute.NetworkProfile{
								NetworkInterfaces: &[]compute.NetworkInterfaceReference{{
									ID: &networkInterfaceName,
								}},
							},
						},
						Location: &location,
					}, nil
				}

				fetcher = configfetchers.NewAzureConfigFetcher(
					state,
					&configfetchers.Credentials{
						Azure: &configfetchers.AzureCredentialsWrapper{
							AzureCredential: vmmanagers.AzureCredential{
								TenantID:       "some-tenant-id",
								SubscriptionID: "some-subscription-id",
								ClientID:       "some-client-id",
								ClientSecret:   "some-client-secret",
								ResourceGroup:  "some-resource-group",
							},
							CloudName: "AzureCloud",
						},
					},
					azureVMClient,
					azureNetworkClient,
					azureIPClient,
					azureImageClient,
				)

				_, err := fetcher.FetchConfig()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not parse storage account and container from the Azure blob URL: 'https://example.com'"))
			})
		})
	})
})
