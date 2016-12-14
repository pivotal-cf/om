```
ॐ
om helps you interact with an Ops Manager

Usage: om [options] <command> [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Commands:
  [apply-changes](docs/apply-changes/index.md)             triggers an install on the Ops Manager targeted
  configure-authentication  configures Ops Manager with an internal userstore and admin user account
  configure-bosh            configures Ops Manager deployed bosh director
  configure-product         configures a staged product
  curl                      issues an authenticated API request
  delete-installation       deletes all the products on the Ops Manager targeted
  delete-unused-products    deletes unused products on the Ops Manager targeted
  export-installation       exports the installation of the target Ops Manager
  help                      prints this usage information
  import-installation       imports a given installation to the Ops Manager targeted
  stage-product             stages a given product in the Ops Manager targeted
  upload-product            uploads a given product to the Ops Manager targeted
  upload-stemcell           uploads a given stemcell to the Ops Manager targeted
  version                   prints the om release version

```
