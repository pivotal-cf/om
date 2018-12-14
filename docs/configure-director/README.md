&larr; [back to Commands](../README.md)

# `om configure-director`
The `configure-director` command will allow you to setup your BOSH tile on the OpsManager.

Once configured, changes will not take affect until the next [`apply-changes`](../apply-changes/README.md).

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
  --config, -c                  string             path to yml file containing all config fields (see docs/configure-director/README.md for format)
  --ops-file                    string (variadic)  YAML operations file
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
vmextensions-configuration:
  some_vm_extension:
    cloud_properties:
      source_dest_check: false
  some_other_vm_extension:
    cloud_properties:
      foo: bar
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
