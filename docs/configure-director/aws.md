&larr; [back to `configure-director`](README.md)

# AWS configure-director

#### iaas-configuration:
**Note for `ssh_private_key`**: You will need to replace newlines with "\n".
This can be done with the simple `bash` chainsaw `cat key.pem | awk '{print $1 "\\n"}' | tr -d '\n'`.

##### Minimal example
```json
{
  "access_key_id": "my-access-key",
  "secret_access_key": "my-secret-key",
  "vpc_id": "vpc-123456",
  "security_group": "sg-123456",
  "key_pair_name": "some-key-pair-name",
  "ssh_private_key": "my-private-key",
  "region": "us-west-1",
  "encrypted": false
}

```

#### director-configuration:
**Note regarding `ntp_servers_string`**: We recommend using this NTP server to all PCF users on AWS

##### Minimal example
```json
{
  "ntp_servers_string": "0.amazon.pool.ntp.org"
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

##### Minimal example
```json
[
  {"name": "us-west-1b"},
  {"name": "us-west-1c"}
]
```

#### networks-configuration:
**Note:** Only one availability zone can be specified per network subnet.

##### Minimal example
```json
{
  "icmp_checks_enabled": false,
  "networks": [
    {
      "name": "opsman-network",
      "subnets": [
        {
          "iaas_identifier": "vpc-subnet-id-1",
          "cidr": "10.0.0.0/24",
          "reserved_ip_ranges": "10.0.0.0-10.0.0.4",
          "dns": "8.8.8.8",
          "gateway": "10.0.0.1",
          "availability_zone_names": [
            "us-west-1b",
          ]
        }
      ]
    },
    {
      "name": "ert-network",
      "subnets": [
        {
          "iaas_identifier": "vpc-subnet-id-2",
          "cidr": "10.0.4.0/22",
          "reserved_ip_ranges": "10.0.4.0-10.0.4.4",
          "dns": "8.8.8.8",
          "gateway": "10.0.4.1",
          "availability_zone_names": [
            "us-west-1b",
          ]
        }
      ]
    },
    {
      "name": "services-network",
      "subnets": [
        {
          "iaas_identifier": "vpc-subnet-id-3",
          "cidr": "10.0.8.0/22",
          "reserved_ip_ranges": "10.0.8.0-10.0.8.4",
          "dns": "8.8.8.8",
          "gateway": "10.0.8.1",
          "availability_zone_names": [
            "us-west-1b",
          ]
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
  "singleton_availability_zone": {
    "name": "us-west-1b"
  },
  "network": {
    "name": "opsman-network"
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
