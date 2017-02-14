&larr; [back to `configure-bosh`](README.md)

# OpenStack configure-bosh

#### --iaas-configuration
**Note for `ssh_private_key`**: You will need to replace newlines with "\n".
This can be done with the simple `bash` chainsaw `cat key.pem | awk '{print $1 "\\n"}' | tr -d '\n'`.

##### Minimal example
```json
{ 
  "openstack_authentication_url": "http://openstack.example.com:5000/v2",
  "openstack_username": "admin",
  "openstack_password": "s3cret",
  "openstack_tenant": "TENANT",
  "openstack_region": "RegionOne",
  "openstack_security_group": "opsmanager",
  "keystone_version": "v2.0",
  "ignore_server_availability_zone": true,
  "ssh_private_key": "my-ssh-key",
  "key_pair_name": "pcf"
  }"
```

#### --director-configuration

##### Minimal example
```json
{
  "ntp_servers_string": "169.254.169.254"
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
    {"name": "nova"}
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
      "iaas_identifier": "openstack-network-guid",
      "service_network": false,
      "subnets": [
        {
          "cidr": "10.0.0.0/24",
          "reserved_ip_ranges": "10.0.0.0-10.0.0.4",
          "dns": "8.8.8.8",
          "gateway": "10.0.0.1",
          "availability_zones": [
            "nova"
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
  "singleton_availability_zone": "nova",
  "network": "services"
}
```
