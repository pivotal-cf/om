&larr; [back to Commands](../README.md)

# `om configure-product`

The `configure-product` command will configure your product properties, network, and resources on the Ops Manager.

## Command Usage
```
ॐ  configure-product
This authenticated command configures a staged product

Usage: om [options] configure-product [<args>]
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
  --config, -c              string             path to yml file containing all config fields (see docs/configure-product/README.md for format)
  --ops-file, -o            string (variadic)  YAML operations file
  --vars-env                string (variadic)  Load variables from environment variables (e.g.: 'MY' to load MY_var=value)
  --vars-file, -l           string (variadic)  Load variables from a YAML file
```

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

#### Configuring the `network-properties` on Azure

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
