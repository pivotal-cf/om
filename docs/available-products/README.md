&larr; [back to Commands](../README.md)

# `om available-products`

The `available-products` command will list all available product names and versions on the Ops Manager.

## Command Usage
```
ॐ  available-products
This authenticated command lists all available products.

Usage: om [options] available-products
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)
```
