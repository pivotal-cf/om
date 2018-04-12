&larr; [back to Commands](../README.md)

# `om configure-director`
The `configure-director` command will allow you to setup your BOSH tile on the OpsManager.

## Supported Infrastructures
* [AWS](aws.md)
* [GCP](gcp.md)
* [Azure](azure.md)
* [vSphere](vsphere.md)

## Command Usage
```
ॐ  configure-director
This authenticated command configures the director.

Usage: om [options] configure-director [<args>]
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
  --az-configuration, -a        string  configures network availability zones
  --director-configuration, -d  string  properties for director configuration
  --iaas-configuration, -i      string  iaas specific JSON configuration for the bosh director
  --network-assignment, -na     string  assigns networks and AZs
  --networks-configuration, -n  string  configures networks for the bosh director
  --resource-configuration, -r  string
  --security-configuration, -s  string
  --syslog-configuration, -l    string
```
