&larr; [back to Commands](../README.md)

# `om apply-changes`

The `apply-changes` command will kick-off an installation on the Ops Manager VM.
It will then track the installation progress, printing logs as they become available.

## Command Usage
```
ॐ  apply-changes
This authenticated command kicks off an install of any staged changes on the Ops Manager.

Usage: om [options] apply-changes [<args>]
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
  --ignore-warnings, -i         bool               ignore issues reported by Ops Manager when applying changes
  --product-name, -n            string (variadic)  name of the product(s) to deploy, cannot be used in conjunction with --skip-deploy-products (OM 2.2+)
  --skip-deploy-products, -sdp  bool               skip deploying products when applying changes - just update the director
```
