# Om

_is a mantra, or vibration, that is traditionally chanted_

![enhancing your calm](http://i.giphy.com/3o7qDQ5iw1oXyDeJAk.gif)

## What is it?

Magical tool that helps you configure and deploy tiles to an Ops-Manager 1.8+ . 
Currently being developed by RelEng, backlog link is [here](https://www.pivotaltracker.com/epic/show/2982497).

## Design Goals

- less flakey / faster replacement of [opsmgr](https://github.com/pivotal-cf/opsmgr)
- single binary that can be run on multiple platforms
- split environment creation from Ops Manager configuration (these are two tools)
- no longer rely on specific environment file format
- fully tested, not using tests to execute browser behavior
- no capybara
- [small sharp tool](https://brandur.org/small-sharp-tools)
- idempotency for all commands

## Documentation

See [here] (docs/README.md) for useful examples and documentation

## Current Commands
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
  apply-changes             triggers an install on the Ops Manager targeted
  available-products        list available products
  configure-authentication  configures Ops Manager with an internal userstore and admin user account
  configure-bosh            configures Ops Manager deployed bosh director
  configure-product         configures a staged product
  curl                      issues an authenticated API request
  delete-installation       deletes all the products on the Ops Manager targeted
  delete-unused-products    deletes unused products on the Ops Manager targeted
  errands                   list errands for a product
  export-installation       exports the installation of the target Ops Manager
  help                      prints this usage information
  import-installation       imports a given installation to the Ops Manager targeted
  revert-staged-changes     reverts staged changes on the Ops Manager targeted
  stage-product             stages a given product in the Ops Manager targeted
  upload-product            uploads a given product to the Ops Manager targeted
  upload-stemcell           uploads a given stemcell to the Ops Manager targeted
  version                   prints the om release version

```
## Authentication
OM will by preference use Client ID and Client Secret if provided. To create a Client ID and Client Secret           
                                                             
1. `uaac target https://YOUR_OPSMANAGER/uaa`
1. `uaac token sso get` if using SAML or `uaac token owner get` if using internal auth
1. Generate a client ID and secret
```
uaac client add -i
Client ID:  NEW_CLIENT_NAME
New client secret:  DESIRED_PASSWORD
Verify new client secret:  DESIRED_PASSWORD
scope (list):  opsman.admin
authorized grant types (list):  client_credentials
authorities (list):  opsman.admin
access token validity (seconds):  43200
refresh token validity (seconds):  43200
redirect uri (list):
autoapprove (list):
signup redirect url (url):```

