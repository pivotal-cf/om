&larr; [back to Commands](../README.md)

# `om delete-installation`

The `delete-installation` command will delete the existing installation from the Ops Manager, including any VMs deployed by BOSH and the BOSH director.

## Command Usage
```
ॐ  delete-installation
This authenticated command deletes all the products installed on the targeted Ops Manager.

Usage: om [options] delete-installation
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)
```
