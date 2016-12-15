&larr; [back to Commands](../README.md)

# `om upload-product`

The `upload-product` command will upload a product to the Ops Manager.
After uploading, you can then use the [`stage-product` command](../stage-product/README.md) to add the product to the installation dashboard.

## Command Usage
```
ॐ  upload-product
This command attempts to upload a product to the Ops Manager

Usage: om [options] upload-product [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Command Arguments:
  -p, --product  string  path to product
```
