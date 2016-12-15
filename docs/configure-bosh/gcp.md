&larr; [back to `configure-bosh`](README.md)

# GCP-specific inputs for the `configure-bosh` command

#### --iaas-configuration
**Note regarding `auth_json`**: You will have likely generated this in order to use our [terraforming-gcp](https://github.com/pivotal-cf/terraforming-gcp/) tooling.
To easily format this JSON for use as a string here, use `cat service_account_key.json | jq 'tostring'`.
`jq` can be downloaded [here](https://stedolan.github.io/jq/).

##### Minimal example
```json
{
  "project": "my-foo-project",
  "default_deployment_tag": "foo-vms",
  "auth_json": "{\"some-key\":\"some-value\"}"
}
```

#### --director-configuration
**Note regarding `ntp_servers_string`**: We recommend using this NTP server to all PCF users on GCP.

##### Minimal example
```json
{
  "ntp_servers_string": "169.254.169.254"
}
```

#### --az-configuration
We tend to use the "us-central1" region because it has 3 zones to balance across for high-availability deployments.

##### Minimal example
```json
{
  "availability_zones": ["us-central1-a","us-central1-b","us-central1-c"]
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
