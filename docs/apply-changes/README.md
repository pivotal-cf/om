&larr; [back to Commands](../README.md)

# `om apply-changes`

The `apply-changes` command will kick-off an installation on the Ops Manager VM.
It will then track the installation progress, printing logs as they become available.

## Command Usage
```
ॐ  apply-changes
This authenticated command kicks off an install of any staged changes on the Ops Manager.

Usage: om [options] apply-changes [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Command Arguments:
  -i, --ignore-warnings         bool  ignore issues reported by Ops Manager when applying changes
  -sdp, --skip-deploy-products  bool  skip deploying products when applying changes - just update the director
```
