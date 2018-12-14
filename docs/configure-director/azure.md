&larr; [back to `configure-director`](README.md)

# Azure-specific inputs for the `configure-director` command

#### iaas-configuration:

##### Minimal example
```json
{
  "subscription_id": "some-subscription-id",
  "tenant_id": "some-tenant-id",
  "client_id": "some-client-id",
  "client_secret": "some-client-id",
  "resource_group_name": "some-resource-group-name",
  "bosh_storage_account_name": "some-bosh-storage-account-name",
  "deployments_storage_account_name": "some-deployments-storage-account-name",
  "default_security_group": "some-default-security-group",
  "ssh_public_key": "some-ssh-public-key",
  "ssh_private_key": "some-ssh-private-key"
}
```

#### director-configuration:

##### Minimal example
```json
{
  "ntp_servers_string": "us.pool.ntp.org"
}
```

#### security-configuration:
No additional security configuration is strictly required.

##### Minimal example
```json
{
  "trusted_certificates": "some-trusted-certificates"
}
```

#### az-configuration:
Azure does not configure or manage AZs and so this configuration is not required.

#### networks-configuration:

##### Minimal example
```json
{
  "icmp_checks_enabled": false,
  "networks": [
    {
      "name": "opsman-network",
      "subnets": [
        {
          "iaas_identifier": "some-network/some-opsman-subnet",
          "cidr": "10.0.8.0/26",
          "reserved_ip_ranges": "10.0.8.0-10.0.8.4",
          "dns": "8.8.8.8",
          "gateway": "10.0.8.1"
        }
      ]
    },
    {
      "name": "ert-network",
      "subnets": [
        {
          "iaas_identifier": "some-network/some-ert-subnet",
          "cidr": "10.0.0.0/22",
          "reserved_ip_ranges": "10.0.0.0-10.0.0.4",
          "dns": "8.8.8.8",
          "gateway": "10.0.0.1"
        }
      ]
    },
    {
      "name": "services-network",
      "subnets": [
        {
          "iaas_identifier": "some-network/some-services-subnet",
          "cidr": "10.0.4.0/22",
          "reserved_ip_ranges": "10.0.4.0-10.0.4.4",
          "dns": "8.8.8.8",
          "gateway": "10.0.4.1"
        }
      ]
    }
  ]
}
```

#### network-assignment:

##### Minimal example
```json
{
  "network": {
    "name" : "opsman-network"
  }
}
```


#### vmextensions-configuration:

##### Minimal example
```json
{
  "some_vm_extension": {
    "cloud_properties": {
      "source_dest_check": false
    }
  }
}
```
