&larr; [back to Commands](../README.md)

# `om configure-bosh`
The `configure-bosh` command will allow you to setup your BOSH tile on the OpsManager.

## Supported Infrastructures
* [AWS](aws.md)
* [GCP](gcp.md)
* [Azure](azure.md)
* [vSphere](vsphere.md)
* [OpenStack](openstack.md)

## Command Usage
```
ॐ  configure-bosh
configures the bosh director that is deployed by the Ops Manager

Usage: om [options] configure-bosh [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Command Arguments:
  -i, --iaas-configuration      string  iaas specific JSON configuration for the bosh director
  -d, --director-configuration  string  director-specific JSON configuration for the bosh director
  -s, --security-configuration  string  security-specific JSON configuration for the bosh director
  -a, --az-configuration        string  availability zones JSON configuration for the bosh director
  -n, --networks-configuration  string  complete network configuration for the bosh director
  -na, --network-assignment     string  choose existing network and availability zone to deploy bosh director into
```
