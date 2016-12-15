&larr; [back to Commands](../README.md)

# `om stage-product`

The `stage-product` command will add or update a product to the Ops Manager installation dashboard.

## Command Usage
```
ॐ  stage-product
This command attempts to stage a product in the Ops Manager

Usage: om [options] stage-product [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Command Arguments:
  -p, --product-name     string  name of product
  -v, --product-version  string  version of product
```
