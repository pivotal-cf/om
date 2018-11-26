&larr; [back to `configure-director`](README.md)

# GCP-specific inputs for the `configure-director` command

#### iaas-configuration:
**Note regarding `auth_json`**: To easily format this JSON for use as a string here, use `cat service_account_key.json | jq 'tostring'`.
`jq` can be downloaded [here](https://stedolan.github.io/jq/).
If you have configured your OpsMan VM to have an attached Service Account, then this field is not required.

##### Minimal example
```json
{
  "project": "my-foo-project",
  "default_deployment_tag": "foo-vms",
  "associated_service_account": "some-service-account",
  "auth_json": "{\"some-key\":\"some-value\"}"
}
```

#### director-configuration:
**Note regarding `ntp_servers_string`**: We recommend using the metadata server (169.254.169.254) as the NTP server to all PCF users on GCP.

##### Minimal example
```json
{
  "ntp_servers_string": "169.254.169.254"
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
We tend to use the "us-central1" region because it has 3 zones to balance across for high-availability deployments.

##### Minimal example
```json
[
  {"name": "us-central1-a"},
  {"name": "us-central1-b"},
  {"name": "us-central1-c"}
]
```

#### networks-configuration:
In the example configuration below, we have taken a single GCP network and described it as 3 Ops Manager networks along subnet boundaries.

##### Minimal example
```json
{
  "icmp_checks_enabled": false,
  "networks": [
    {
      "name": "opsman-network",
      "subnets": [
        {
          "iaas_identifier": "some-network/opsman-subnet/us-central1",
          "cidr": "10.0.0.0/24",
          "reserved_ip_ranges": "10.0.0.0-10.0.0.4",
          "dns": "8.8.8.8",
          "gateway": "10.0.0.1",
          "availability_zone_names": [
            "us-central1-a",
            "us-central1-b",
            "us-central1-c"
          ]
        }
      ]
    },
    {
      "name": "ert-network",
      "subnets": [
        {
          "iaas_identifier": "some-network/ert-subnet/us-central1",
          "cidr": "10.0.4.0/22",
          "reserved_ip_ranges": "10.0.4.0-10.0.4.4",
          "dns": "8.8.8.8",
          "gateway": "10.0.4.1",
          "availability_zone_names": [
            "us-central1-a",
            "us-central1-b",
            "us-central1-c"
          ]
        }
      ]
    },
    {
      "name": "services-network",
      "service_network": true,
      "subnets": [
        {
          "iaas_identifier": "some-network/services-subnet/us-central1",
          "cidr": "10.0.8.0/22",
          "reserved_ip_ranges": "10.0.8.0-10.0.8.4",
          "dns": "8.8.8.8",
          "gateway": "10.0.8.1",
          "availability_zone_names": [
            "us-central1-a",
            "us-central1-b",
            "us-central1-c"
          ]
        }
      ]
    }
  ]
}
```

#### network-assignment:
This flag will set the network assignment for the BOSH Director tile itself.

##### Minimal example
```json
{
  "singleton_availability_zone": {
    "name": "us-central1-a"
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
