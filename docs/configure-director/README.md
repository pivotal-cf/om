&larr; [back to Commands](../README.md)

# `om configure-director`
The `configure-director` command will allow you to setup your BOSH tile on the OpsManager.

## Supported Infrastructures
* [AWS](aws.md)
* [GCP](gcp.md)
* [Azure](azure.md)
* [vSphere](vsphere.md)

## Command Usage
```
ॐ  configure-director
This authenticated command configures the director.

Usage: om [options] configure-director [<args>]
  --client-id, -c            string  Client ID for the Ops Manager VM (not required for unauthenticated commands, $OM_CLIENT_ID)
  --client-secret, -s        string  Client Secret for the Ops Manager VM (not required for unauthenticated commands, $OM_CLIENT_SECRET)
  --format, -f               string  Format to print as (options: table,json) (default: table)
  --help, -h                 bool    prints this usage information (default: false)
  --password, -p             string  admin password for the Ops Manager VM (not required for unauthenticated commands, $OM_PASSWORD)
  --request-timeout, -r      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)
  --skip-ssl-validation, -k  bool    skip ssl certificate validation during http requests (default: false)
  --target, -t               string  location of the Ops Manager VM
  --trace, -tr               bool    prints HTTP requests and response payloads
  --username, -u             string  admin username for the Ops Manager VM (not required for unauthenticated commands, $OM_USERNAME)
  --version, -v              bool    prints the om release version (default: false)

Command Arguments:
  --az-configuration, -a        string  configures network availability zones
  --config, -c                  string  path to yml file containing all config fields (see docs/configure-director/README.md for format)
  --director-configuration, -d  string  properties for director configuration
  --iaas-configuration, -i      string  iaas specific JSON configuration for the bosh director
  --network-assignment, -na     string  assigns networks and AZs
  --networks-configuration, -n  string  configures networks for the bosh director
  --resource-configuration, -r  string
  --security-configuration, -s  string
  --syslog-configuration, -l    string
```

### Configuring via file

The `--config` flag is available for convenience to allow you to pass a single
file with all the configuration required to configure your director.

When providing a single config file each of the other individual flags maps to a
top-level element in the YAML file.

#### Example YAML:
```yaml
---
az-configuration:
- name: some-az
director-configuration:
  max_threads: 5
iaas-configuration:
  iaas_specific_key: some-value
network-assignment:
  network:
    name: some-network
networks-configuration:
  networks:
  - network: network-1
resource-configuration:
  compilation:
    instance_type:
      id: m4.xlarge
security-configuration:
  trusted_certificates: some-certificate
syslog-configuration:
  syslogconfig: awesome
```
