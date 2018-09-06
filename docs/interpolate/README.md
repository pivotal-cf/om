&larr; [back to Commands](../README.md)

# `om interpolate`

The `interpolate` command allows you to test template interpolation in
isolation.

For example if you are modifying a product configuration template you can use
`om interpolate` to verify that the generated config looks correct before
running `om configure-product`.

## Command Usage
```
ॐ  interpolate
Interpolates variables into a manifest

Usage: om [options] interpolate [<args>]
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
  --config, -c     string (required)  path for file to be interpolated
  --ops-file, -o   string (variadic)  YAML operations files
  --vars-env       string (variadic)  Load variables from environment variables (e.g.: 'MY' to load MY_var=value)
  --vars-file, -l  string (variadic)  Load variables from a YAML file
```

## Interpolation

Given a template file with a variable reference:

```yaml
# config.yml
key: ((variable_name))
```

Values can be provided from a separate variables yaml file (`--vars-file`) or from environment variables (`--vars-env`).

To load variables from a file use the `--vars-file` flag.

```yaml
# vars.yml
variable_name: some_value
```

```
om interpolate \
  --config config.yml \
  --vars-file vars.yml
```

To load variables from a set of environment variables, specify the common
environment variable prefix with the `--vars-env` flag.

```
OM_VAR_variable_name=some_value om interpolate \
  --config config.yml \
  --vars-env OM_VAR
```

The interpolation support is inspired by similar features in BOSH. You can
[refer to the BOSH documentation](https://bosh.io/docs/cli-int/) for details on how interpolation
is performed.
