&larr; [back to `configure-bosh`](README.md)

# vSphere-specific inputs for the `configure-bosh` command

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
}
```

#### --networks-configuration

##### Minimal example
```json
{
}
```

#### --network-assignment

##### Minimal example
```json
{
}
```
