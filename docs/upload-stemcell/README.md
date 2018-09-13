&larr; [back to Commands](../README.md)

# `om upload-stemcell`

The `upload-stemcell` command will upload a stemcell to the Ops Manager.
This stemcell will then be available for use by any product specifying that stemcell version.

## Command Usage
```
ॐ  upload-stemcell
This command will upload a stemcell to the target Ops Manager. Unless the force flag is used, if the stemcell already exists that upload will be skipped

Usage: om [options] upload-stemcell [<args>]
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
  --floating      bool               assigns the stemcell to all compatible products  (default: true)
  --force, -f     bool               upload stemcell even if it already exists on the target Ops Manager
  --shasum, -sha  string             shasum of the provided stemcell file to be used for validation
  --stemcell, -s  string (required)  path to stemcell
```
