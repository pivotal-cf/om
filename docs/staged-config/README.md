&larr; [back to Commands](../README.md)

# `om staged-product`

The `staged-config` command will export a YAML config file that can be used with `configure-product`.

## Command Usage
```
ॐ  staged-config
This command generates a config from a staged product that can be passed in to om configure-product (Note: credentials are not available and will appear as '***')

Usage: om [options] staged-config [<args>]
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
  --include-credentials, -c  bool               include credentials. note: requires product to have been deployed
  --product-name, -p         string (required)  name of product
```
