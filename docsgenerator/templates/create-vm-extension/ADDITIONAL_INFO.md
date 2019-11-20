### Configuring via file

#### Example YAML:
```yaml
vm-extension-config:
  name: some_vm_extension
  cloud_properties:
    source_dest_check: false
```

#### Variables

The `create-vm-extension` command now supports variable substitution inside the config template:

```yaml
# config.yml
vm-extension-config:
  name: some_vm_extension
  cloud_properties:
    source_dest_check: ((enable_source_dest_check))
```

Values can be provided from a separate variables yaml file (`--vars-file`) or from environment variables (`--vars-env`).

To load variables from a file use the `--vars-file` flag.

```yaml
# vars.yml
enable_source_dest_check: false
```

```
om create-vm-extension \
  --config config.yml \
  --vars-file vars.yml
```

To load variables from a set of environment variables, specify the common
environment variable prefix with the `--vars-env` flag.

```
OM_VAR_enable_source_dest_check=false om create-vm-extension \
  --config config.yml \
  --vars-env OM_VAR
```

The interpolation support is inspired by similar features in BOSH. You can
[refer to the BOSH documentation](https://bosh.io/docs/cli-int/) for details on how interpolation
is performed.