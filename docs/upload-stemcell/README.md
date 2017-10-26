&larr; [back to Commands](../README.md)

# `om upload-stemcell`

The `upload-stemcell` command will upload a stemcell to the Ops Manager.
This stemcell will then be available for use by any product specifying that stemcell version.

## Command Usage
```
ॐ  upload-stemcell
This command will upload a stemcell to the target Ops Manager. Unless the force flag is used, if the stemcell already exists that upload will be skipped

Usage: om [options] upload-stemcell [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Command Arguments:
  -s, --stemcell  string  path to stemcell
  -f, --force     string  upload stemcell even if it already exists on the target Ops Manager
```
