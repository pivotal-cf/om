&larr; [back to Commands](../README.md)

# `om create-vm-extension`

The `create-vm-extension` command will create or update an existing vm extension.

## Command Usage
```
ॐ  create-vm-extension
This creates/updates a VM extension

Usage: om [options] configure-product [<args>]
  --client-id, -c, OM_CLIENT_ID          string  Client ID for the Ops Manager VM (not required for unauthenticated commands)
  --client-secret, -s, OM_CLIENT_SECRET  string  Client Secret for the Ops Manager VM (not required for unauthenticated commands)
  --connect-timeout, -o                  int     timeout in seconds to make TCP connections (default: 5)
  --format, -f                           string  Format to print as (options: table,json) (default: table)
  --help, -h                             bool    prints this usage information (default: false)
  --password, -p, OM_PASSWORD            string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  --request-timeout, -r                  int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)
  --skip-ssl-validation, -k              bool    skip ssl certificate validation during http requests (default: false)
  --target, -t, OM_TARGET                string  location of the Ops Manager VM
  --trace, -tr                           bool    prints HTTP requests and response payloads
  --username, -u, OM_USERNAME            string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  --version, -v                          bool    prints the om release version (default: false)

Command Arguments:
  --cloud-properties, -cp  string (required)  cloud properties in JSON format
  --config, -c             string             path to yml file containing all config fields (see docs/create-vm-extension/README.md for format)
  --name, -n               string (required)  VM extension name
  --ops-file, -o           string (variadic)  YAML operations file
  --vars-file, -l          string (variadic)  Load variables from a YAML file
```

### Configuring via file

#### Example YAML:
```yaml
vm-extension-config:
  name: some_vm_extension
  cloud_properties:
    source_dest_check: false
```
