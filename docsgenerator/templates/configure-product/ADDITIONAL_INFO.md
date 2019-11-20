### Configuring via YAML config file

The preferred approach is to include all configuration in a single YAML
configuration file.

#### Example YAML

A real product may have many more product properties to configure but this gives
you the general structure of the file:

```yaml
product-name: sample-product
product-properties:
  .cloud_controller.apps_domain:
    value: apps.example.com
network-properties:
  network:
    name: some-network
  other_availability_zones:
  - name: us-west-2a
  - name: us-west-2b
  - name: us-west-2c
  singleton_availability_zone:
    name: us-west-2a
resource-config:
  diego_cell:
    instances: 3
  diego_brain:
    elb_names:
    - some-elb
errand-config:
  smoke_tests:
    post-deploy-state: true
    pre-delete-state: default
  push-usage-service:
    post-deploy-state: false
    pre-delete-state: default
```

To retrieve the current configuration of your product you can use the `om
staged-config` command.

#### Variables

The `configure-product` command now supports variable substitution inside the config template:

```yaml
# config.yml
product-properties:
  .some.property:
    value:
      password: ((password))
```

Values can be provided from a separate variables yaml file (`--vars-file`) or from environment variables (`--vars-env`).

To load variables from a file use the `--vars-file` flag.

```yaml
# vars.yml
password: something-secure
```

```
om configure-product \
  --config config.yml \
  --vars-file vars.yml
```

To load variables from a set of environment variables, specify the common
environment variable prefix with the `--vars-env` flag.

```
OM_VAR_password=something-secure OM_VAR_another_key=another_value om configure-product \
  --config config.yml \
  --vars-env OM_VAR
```

The interpolation support is inspired by similar features in BOSH. You can
[refer to the BOSH documentation](https://bosh.io/docs/cli-int/) for details on how interpolation
is performed.

#### Configuring the `network-properties` on Azure prior to Ops Manager 2.5

The product network on Azure does not include Availability Zones, but the API will still expect them to be provided.
To satisfy the API, you can submit "null" AZs for the API as is shown here:

```yaml
network-properties:
  network:
    name: some-network
  other_availability_zones:
  - name: "null"
  singleton_availability_zone:
    name: "null"
```

**Note:** you will need to remove this null
for use with Ops Manager 2.5 and after, or you will see this error:
```json
{"errors":["Availability zones cannot find availability zone with name null"]}
```