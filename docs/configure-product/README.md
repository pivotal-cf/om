&larr; [back to Commands](../README.md)

# `om configure-product`

The `configure-product` command will configure your product properties, network, and resources on the Ops Manager.

## Command Usage
```
ॐ  configure-product
This authenticated command configures a staged product

Usage: om [options] configure-product [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Command Arguments:
  -n, --product-name        string  name of the product being configured
  -p, --product-properties  string  properties to be configured in JSON format (default: )
  -pn, --product-network    string  network properties in JSON format (default: )
  -pr, --product-resources  string  resource configurations in JSON format (default: {})
```
