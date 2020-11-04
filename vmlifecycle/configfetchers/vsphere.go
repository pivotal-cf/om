package configfetchers

import (
	"context"
	"errors"
	"fmt"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"path"
	"strconv"
	"strings"
)

type VSphereConfigFetcher struct {
	state       *vmmanagers.StateInfo
	credentials *Credentials
	client      *vim25.Client
	collector   *property.Collector
	ctx         context.Context
}

func NewVSphereConfigFetcher(state *vmmanagers.StateInfo, creds *Credentials, client *govmomi.Client) *VSphereConfigFetcher {
	return &VSphereConfigFetcher{
		state:       state,
		credentials: creds,
		client:      client.Client,
		collector:   property.DefaultCollector(client.Client),
		ctx:         context.Background(),
	}
}

func (v *VSphereConfigFetcher) FetchConfig() (*vmmanagers.OpsmanConfigFilePayload, error) {
	datacenter, err := v.parseDatacenter()
	if err != nil {
		return nil, err
	}

	machine, err := v.lookupVM()
	if err != nil {
		return nil, fmt.Errorf("could not fetch vm id: %s", err)
	}

	if len(machine.Guest.Net) == 0 {
		return nil, fmt.Errorf("expected the VM to be assigned to a network")
	}

	appConfigProperties, err := v.lookupAppConfigProperties(machine)
	if err != nil {
		return nil, err
	}

	datastore, err := v.lookupDatastore(machine)
	if err != nil {
		return nil, err
	}

	resourcePool, err := v.lookupResourcePool(machine)
	if err != nil {
		return nil, err
	}

	diskType, err := v.lookupDiskType(machine)
	if err != nil {
		return nil, err
	}

	network, err := v.lookupNetwork(machine)
	if err != nil {
		return nil, err
	}

	insecure := "0"
	caCert := "((ca-cert))"
	if v.credentials.VSphere.Insecure == true {
		insecure = "1"
		caCert = ""
	}

	folder := strings.Split(v.state.ID, "/")

	return &vmmanagers.OpsmanConfigFilePayload{
		OpsmanConfig: vmmanagers.OpsmanConfig{
			Vsphere: &vmmanagers.VsphereConfig{
				Vcenter: vmmanagers.Vcenter{
					VcenterCredential: vmmanagers.VcenterCredential{
						URL:      v.credentials.VSphere.URL,
						Username: "((vcenter-username))",
						Password: "((vcenter-password))",
					},
					Datacenter:   datacenter,
					Datastore:    datastore,
					Insecure:     insecure,
					CACert:       caCert,
					ResourcePool: resourcePool,
					Folder:       strings.Join(folder[:len(folder)-1], "/"),
				},
				DiskType:     diskType,
				PrivateIP:    appConfigProperties["ip0"],
				DNS:          appConfigProperties["DNS"],
				NTP:          appConfigProperties["ntp_servers"],
				SSHPublicKey: "((ssh-public-key))",
				Hostname:     appConfigProperties["custom_hostname"],
				Network:      network,
				Netmask:      appConfigProperties["netmask0"],
				Gateway:      appConfigProperties["gateway"],
				VMName:       machine.Name,
				Memory:       strconv.Itoa(int(machine.Summary.Config.MemorySizeMB) / 1024),
				CPU:          strconv.Itoa(int(machine.Summary.Config.NumCpu)),
			},
		},
	}, nil
}

func (v *VSphereConfigFetcher) lookupVM() (mo.VirtualMachine, error) {
	si := object.NewSearchIndex(v.client)

	ref, err := si.FindByInventoryPath(v.ctx, v.state.ID)
	if err != nil {
		return mo.VirtualMachine{}, fmt.Errorf("find by inventory path error: %s", err)
	}

	if ref == nil {
		return mo.VirtualMachine{}, fmt.Errorf("no machine found using ID: %s. Please check your state file and try again", v.state.ID)
	}

	var machine mo.VirtualMachine
	err = v.collector.RetrieveOne(v.ctx, ref.Reference(), nil, &machine)
	if err != nil {
		return mo.VirtualMachine{}, fmt.Errorf("failed to lookup virtual machine: %s", err)
	}

	return machine, nil
}

func (v *VSphereConfigFetcher) lookupDatastore(machine mo.VirtualMachine) (string, error) {
	var datastore mo.Datastore

	err := v.collector.RetrieveOne(v.ctx, machine.Datastore[0], []string{"name"}, &datastore)
	if err != nil {
		return "", fmt.Errorf("failed to lookup datastore: %s", err)
	}
	return datastore.Name, nil
}

func (v *VSphereConfigFetcher) lookupResourcePool(machine mo.VirtualMachine) (string, error) {
	datacenter, err := v.parseDatacenter()
	if err != nil {
		return "", err
	}

	clusterChild, resourcePoolPath, err := v.navigatePathToCluster(machine)
	if err != nil {
		return "", err
	}

	cluster, err := v.lookupCluster(clusterChild)
	if err != nil {
		return "", err
	}

	return path.Clean(fmt.Sprintf("/%s/host/%s/%s", datacenter, cluster, resourcePoolPath)), nil
}

func (v *VSphereConfigFetcher) navigatePathToCluster(machine mo.VirtualMachine) (clusterChild mo.ResourcePool, resourcePoolPath string, err error) {
	err = v.collector.RetrieveOne(v.ctx, machine.ResourcePool.Reference(), []string{"name", "parent"}, &clusterChild)
	if err != nil {
		return mo.ResourcePool{}, "", fmt.Errorf("could not lookup resource pool/cluster: %s", err)
	}

	var resourcePoolSuffix string
	resourcePoolSuffix = fmt.Sprintf("%s/%s", clusterChild.Name, resourcePoolSuffix)

	for clusterChild.Parent.Type == "ResourcePool" {
		err = v.collector.RetrieveOne(v.ctx, clusterChild.Parent.Reference(), []string{"name", "parent"}, &clusterChild)
		if err != nil {
			return mo.ResourcePool{}, "", fmt.Errorf("resourcePool get error: %s", err)
		}
		resourcePoolSuffix = fmt.Sprintf("%s/%s", clusterChild.Name, resourcePoolSuffix)
	}

	return clusterChild, resourcePoolSuffix, nil
}

func (v *VSphereConfigFetcher) lookupCluster(clusterChild mo.ResourcePool) (string, error) {
	var cluster mo.ComputeResource
	err := v.collector.RetrieveOne(v.ctx, clusterChild.Parent.Reference(), []string{"name"}, &cluster)
	if err != nil {
		return "", fmt.Errorf("could not lookup resource pool/cluster: could not find cluster: %s", err)
	}
	return cluster.Name, nil
}

func (v *VSphereConfigFetcher) parseDatacenter() (datacenter string, err error) {
	splitID := strings.Split(v.state.ID, "/")
	if len(splitID) == 0 {
		return "", fmt.Errorf("could not parse datacenter from state vm_id %s: vm_id should start with '/datacenter/`", v.state.ID)
	}
	datacenter = splitID[1]
	return datacenter, nil
}

func (v *VSphereConfigFetcher) lookupDiskType(machine mo.VirtualMachine) (string, error) {
	diskType := "thin"
	var foundBackingDevice bool
	devices := object.VirtualDeviceList(machine.Config.Hardware.Device)
	for _, device := range devices {
		if devices.TypeName(device) == "VirtualDisk" {
			virtualDevice := device.GetVirtualDevice()
			if backing, ok := virtualDevice.Backing.(*types.VirtualDiskFlatVer2BackingInfo); ok {
				if !*backing.ThinProvisioned {
					diskType = "thick"
				}
				foundBackingDevice = true
			}
		}
	}

	if !foundBackingDevice {
		return "", fmt.Errorf("could not lookup disk type: no backing device found for VM: %s", machine.Name)
	}

	return diskType, nil
}

func (v *VSphereConfigFetcher) lookupNetwork(machine mo.VirtualMachine) (string, error) {
	var network mo.Network
	err := v.collector.RetrieveOne(v.ctx, machine.Network[0].Reference(), []string{"name"}, &network)
	if err != nil {
		return "", fmt.Errorf("could not lookup network: could not find network name: %s", err)
	}
	return network.Name, nil
}

func (v *VSphereConfigFetcher) lookupAppConfigProperties(machine mo.VirtualMachine) (map[string]string, error) {
	appConfigProperties := make(map[string]string)
	for _, prop := range machine.Config.VAppConfig.GetVmConfigInfo().Property {
		appConfigProperties[prop.Id] = prop.Value
	}

	requiredProperties := []string{"ip0", "DNS", "ntp_servers", "custom_hostname", "netmask0", "gateway"}
	var errs []string
	for _, property := range requiredProperties {
		if appConfigProperties[property] == "" {
			errs = append(errs, fmt.Sprintf("could not find '%s' for VM '%s'", property, machine.Name))
		}
	}

	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, "\n"))
	}
	return appConfigProperties, nil
}
