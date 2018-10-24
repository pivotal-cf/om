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
  --client-id, -c, OM_CLIENT_ID          string  Client ID for the Ops Manager VM (not required for unauthenticated commands)
  --client-secret, -s, OM_CLIENT_SECRET  string  Client Secret for the Ops Manager VM (not required for unauthenticated commands)
  --connect-timeout, -o                  int     timeout in seconds to make TCP connections (default: 5)
  --env, -e                              string  env file with login credentials
  --help, -h                             bool    prints this usage information (default: false)
  --password, -p, OM_PASSWORD            string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  --request-timeout, -r                  int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)
  --skip-ssl-validation, -k              bool    skip ssl certificate validation during http requests (default: false)
  --target, -t, OM_TARGET                string  location of the Ops Manager VM
  --trace, -tr                           bool    prints HTTP requests and response payloads
  --username, -u, OM_USERNAME            string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  --version, -v                          bool    prints the om release version (default: false)

Command Arguments:
  --az-configuration, -a        string             configures network availability zones
  --config, -c                  string             path to yml file containing all config fields (see docs/configure-director/README.md for format)
  --director-configuration, -d  string             properties for director configuration
  --iaas-configuration, -i      string             iaas specific JSON configuration for the bosh director
  --network-assignment, -na     string             assigns networks and AZs
  --networks-configuration, -n  string             configures networks for the bosh director
  --ops-file                    string (variadic)  YAML operations file
  --resource-configuration, -r  string
  --security-configuration, -s  string
  --syslog-configuration, -l    string
  --vars-env                    string (variadic)  Load variables from environment variables (e.g.: 'MY' to load MY_var=value)
  --vars-file                   string (variadic)  Load variables from a YAML file
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
- name: us-east-1a
  iaas_configuration_guid: ae31e369c92b6a647da0
director-configuration:
  allow_legacy_agents: true
  blobstore_type: local
  bosh_recreate_on_next_deploy: false
  bosh_recreate_persistent_disks_on_next_deploy: false
  database_type: internal
  director_worker_count: 5
  encryption:
    keys: []
    providers: []
  excluded_recursors: []
  hm_emailer_options:
    enabled: false
  hm_pager_duty_options:
    enabled: false
  identification_tags:
    iaas: aws
    proof: pudding
  keep_unreachable_vms: false
  local_blobstore_options:
    tls_enabled: false
  max_threads: 30
  ntp_servers_string: 0.amazon.pool.ntp.org, 1.amazon.pool.ntp.org, 2.amazon.pool.ntp.org,
    3.amazon.pool.ntp.org
  post_deploy_enabled: false
  resurrector_enabled: true
  retry_bosh_deploys: true
iaas-configuration:
  encrypted: false
  guid: ((iaas-configuration_guid))
  iam_instance_profile: ((iaas-configuration_iam_instance_profile))
  key_pair_name: ((iaas-configuration_key_pair_name))
  kms_key_arn: ((iaas-configuration_kms_key_arn))
  name: ((iaas-configuration_name))
  region: ((iaas-configuration_region))
  security_group: ((iaas-configuration_security_group))
network-assignment:
  network:
    name: pcf-management-network
  other_availability_zones: []
  singleton_availability_zone:
    name: us-east-1a
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: pcf-management-network
    subnets:
    - iaas_identifier: subnet-05c15ea14334dfdf4 # You will need to change this to your subnet
      cidr: 10.0.16.0/28
      dns: 10.0.0.2
      gateway: 10.0.16.1
      reserved_ip_ranges: 10.0.16.0-10.0.16.4
      availability_zone_names:
      - us-east-1a
  - name: pcf-pks-network
    subnets:
    - iaas_identifier: subnet-029592de0b99d3a18 # You will need to change this to your subnet
      cidr: 10.0.4.0/24
      dns: 10.0.0.2
      gateway: 10.0.4.1
      reserved_ip_ranges: 10.0.4.0-10.0.4.4
      availability_zone_names:
      - us-east-1a
  - name: pcf-services-network
    subnets:
    - iaas_identifier: subnet-01f95dd2873535b85 # You will need to change this to your subnet
      cidr: 10.0.8.0/24
      dns: 10.0.0.2
      gateway: 10.0.8.1
      reserved_ip_ranges: 10.0.8.0-10.0.8.3
      availability_zone_names:
      - us-east-1a
resource-configuration:
  compilation:
    instances: automatic
    instance_type:
      id: t2.small
    internet_connected: false
  director:
    instances: automatic
    persistent_disk:
      size_mb: automatic
    instance_type:
      id: m4.large
    internet_connected: false
security-configuration:
  generate_vm_passwords: true
syslog-configuration:
  enabled: false
vmextensions-configuration: []
```

#### Variables

The `configure-director` command now supports variable substitution inside the config template:

```yaml
# config.yml
network-assignment:
  network:
    name: ((network_name))
```

Values can be provided from a separate variables yaml file (`--vars-file`) or from environment variables (`--vars-env`).

To load variables from a file use the `--vars-file` flag.

```yaml
# vars.yml
network_name: some-network
```

```
om configure-director \
  --config config.yml \
  --vars-file vars.yml
```

To load variables from a set of environment variables, specify the common
environment variable prefix with the `--vars-env` flag.

```
OM_VAR_network_name=some-network om configure-director \
  --config config.yml \
  --vars-env OM_VAR
```

The interpolation support is inspired by similar features in BOSH. You can
[refer to the BOSH documentation](https://bosh.io/docs/cli-int/) for details on how interpolation
is performed.
