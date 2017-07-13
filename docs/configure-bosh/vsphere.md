&larr; [back to `configure-bosh`](README.md)

# vSphere-specific inputs for the `configure-bosh` command

:exclamation: The support for this feature is still **alpha**. Please open issues for any problems you experience.

#### --iaas-configuration
**Note for `disk_type`**: the only valid options are 'thick' or 'thin'.

##### Minimal example
```json
{
  "vcenter_host": "some-vcenter-host",
  "vcenter_username": "my-vcenter-username",
  "vcenter_password": "my-vcenter-password",
  "datacenter": "some-datacenter-name",
  "disk_type": "some-virtual-disk-type",
  "ephemeral_datastores_string": "some,ephemeral,datastores",
  "persistent_datastores_string": "some,persistent,datastores",
  "bosh_vm_folder": "some-vm-folder",
  "bosh_template_folder": "some-template-folder",
  "bosh_disk_path": "some-disk-path"
}
```

##### NSX Example
```json
{
  "vcenter_host": "some-vcenter-host",
  "vcenter_username": "my-vcenter-username",
  "vcenter_password": "my-vcenter-password",
  "datacenter": "some-datacenter-name",
  "disk_type": "some-virtual-disk-type",
  "ephemeral_datastores_string": "some,ephemeral,datastores",
  "persistent_datastores_string": "some,persistent,datastores",
  "bosh_vm_folder": "some-vm-folder",
  "bosh_template_folder": "some-template-folder",
  "bosh_disk_path": "some-disk-path",
  "nsx_networking_enabled": true,
  "nsx_address": "some-nsx-address",
  "nsx_password": "some-password",
  "nsx_username": "some-username",
  "nsx_ca_certificate": "some-ca-certificate"
}
```

#### --director-configuration
Change this to a valid internal NTP server address for your organization

##### Minimal example
```json
{
  "ntp_servers_string": "10.0.0.1"
}
```

#### --security-configuration
No additional security configuration is strictly required.

##### Minimal example
```json
{
  "trusted_certificates": "some-trusted-certificates",
  "vm_password_type": "generate"
}
```

#### --az-configuration

##### Minimal example
```json
{
  "availability_zones": [
    {
      "name": "az-1",
      "cluster": "cluster-1",
      "resource_pool": "pool-1"
    },
    {
      "name": "az-2",
      "cluster": "cluster-2",
      "resource_pool": "pool-2"
    },
    {
      "name": "az-3",
      "cluster": "cluster-3",
      "resource_pool": "pool-3"
    },
  ]
}
```

#### --networks-configuration

##### Minimal example
```json
{
  "icmp_checks_enabled": false,
  "networks": [
    {
      "name": "opsman-network",
      "service_network": false,
      "subnets": [
        {
          "iaas_identifier": "vsphere-network-name",
          "cidr": "10.0.0.0/24",
          "reserved_ip_ranges": "10.0.0.0-10.0.0.4",
          "dns": "8.8.8.8",
          "gateway": "10.0.0.1",
          "availability_zones": [
            "az-1",
            "az-2",
            "az-3"
          ]
        }
      ]
    },
    {
      "name": "ert-network",
      "service_network": false,
      "subnets": [
        {
          "iaas_identifier": "vsphere-network-name",
          "cidr": "10.0.4.0/24",
          "reserved_ip_ranges": "10.0.4.0-10.0.4.4",
          "dns": "8.8.8.8",
          "gateway": "10.0.4.1",
          "availability_zones": [
            "az-1",
            "az-2",
            "az-3"
          ]
        }
      ]
    },
    {
      "name": "services-network",
      "service_network": false,
      "subnets": [
        {
          "iaas_identifier": "vsphere-network-name",
          "cidr": "10.0.8.0/24",
          "reserved_ip_ranges": "10.0.8.0-10.0.8.4",
          "dns": "8.8.8.8",
          "gateway": "10.0.8.1",
          "availability_zones": [
            "az-1",
            "az-2",
            "az-3"
          ]
        }
      ]
    }
  ]
}
```

#### --network-assignment

##### Minimal example
```json
{
  "singleton_availability_zone": "az-1",
  "network": "opsman-network"
}
```
