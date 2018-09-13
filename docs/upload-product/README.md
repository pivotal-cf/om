&larr; [back to Commands](../README.md)

# `om upload-product`

The `upload-product` command will upload a product to the Ops Manager.
After uploading, you can then use the [`stage-product` command](../stage-product/README.md) to add the product to the installation dashboard.

## Command Usage
```
ॐ  upload-product
This command attempts to upload a product to the Ops Manager

Usage: om [options] upload-product [<args>]
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
  --polling-interval, -pi  int                interval (in seconds) at which to print status (default: 1)
  --product, -p            string (required)  path to product
  --shasum, -sha           string             shasum of the provided product file to be used for validation
```
