&larr; [back to Commands](../README.md)

# `om curl`

The `curl` command will help you make arbitrary API calls against the Ops Manager VM.

## Command Usage
```
ॐ  curl
This command issues an authenticated API request as defined in the arguments

Usage: om [options] curl [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Command Arguments:
  -p, --path     string  path to api endpoint
  -x, --request  string  http verb (default: GET)
  -d, --data     string  api request payload
```
