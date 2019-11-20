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
- name: a_vm_extension
  cloud_properties:
    source_dest_check: false
- name: another_vm_extension
  cloud_properties:
    foo: bar
vmtypes-configuration:
  custom_only: false
  vm_types:
  - name: a1.large
    cpu: 4
    ram: 8192
    ephemeral_disk: 10240
  - name: t2.small
    cpu: 1
    ram: 512
    ephemeral_disk: 1024
```

#### vmtypes-configuration:

Will set or update custom VM types on the director. If `custom_only` is `true`, 
the VM types specified in your configuration will be the **entire** list of
available VM types in the Ops Manager. If `false` or omitted, it will add the 
listed VM types to the list of default VM types for your IaaS. If a specified
VM type is named the same as a predefined VM type, it will overwrite the predefined
type. If multiple specified VM types have the same name, the one specified last
will be created. In either case, existing custom VM types do not persist across
`configure-director` calls, and it should be expected that the entire list of custom
VM types is specified in the director configuration.

##### Minimal example
```yaml
vmtypes-configuration:
  custom_only: false
  vm_types:
  - name: x1.large
    cpu: 8
    ram: 8192
    ephemeral_disk: 10240
  - name: mycustomvmtype
    cpu: 4
    ram: 16384
    ephemeral_disk: 4096
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