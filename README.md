# Om

_is a mantra, or vibration, that is traditionally chanted_

![enhancing your calm](http://i.giphy.com/3o7qDQ5iw1oXyDeJAk.gif)

## What is it?

Magical tool that helps you configure and deploy tiles to an Ops-Manager 1.8+ . 
Currently being developed by RelEng, backlog link is [here](https://www.pivotaltracker.com/n/projects/2145041).

## Documentation

See [here](docs/README.md) for useful examples and documentation.

Some commands such as `configure-product` accept arguments as JSON. Please refer
to the Ops Manager API documentation for examples of the fields that can be
provided. The Ops Manager API documentation is available at
`https://pcf.your-ops-manager.example.com/docs`

## Installation

To install `om` go to [Releases](https://github.com/pivotal-cf/om/releases)

Additionally, you can install `om` via `apt-get` from [Stark and Wayne's](https://www.starkandwayne.com/) apt-get repo:
```
wget -q -O - https://raw.githubusercontent.com/starkandwayne/homebrew-cf/master/public.key | apt-key add -
echo "deb http://apt.starkandwayne.com stable main" | tee /etc/apt/sources.list.d/starkandwayne.list
apt-get update

apt-get install om
```

## Current Commands
```
ॐ
om helps you interact with an Ops Manager

Usage: om [options] <command> [<args>]
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

Commands:
  activate-certificate-authority  activates a certificate authority on the Ops Manager
  apply-changes                   triggers an install on the Ops Manager targeted
  available-products              list available products
  certificate-authorities         lists certificates managed by Ops Manager
  certificate-authority           prints requested certificate authority
  config-template                 **EXPERIMENTAL** generates a config template for the product
  configure-authentication        configures Ops Manager with an internal userstore and admin user account
  configure-bosh                  configures Ops Manager deployed bosh director
  configure-director              configures the director
  configure-product               configures a staged product
  create-certificate-authority    creates a certificate authority on the Ops Manager
  create-vm-extension             creates a VM extension
  credential-references           list credential references for a deployed product
  credentials                     fetch credentials for a deployed product
  curl                            issues an authenticated API request
  delete-certificate-authority    deletes a certificate authority on the Ops Manager
  delete-installation             deletes all the products on the Ops Manager targeted
  delete-product                  deletes a product from the Ops Manager
  delete-unused-products          deletes unused products on the Ops Manager targeted
  deployed-manifest               prints the deployed manifest for a product
  deployed-products               lists deployed products
  errands                         list errands for a product
  export-installation             exports the installation of the target Ops Manager
  generate-certificate            generates a new certificate signed by Ops Manager's root CA
  generate-certificate-authority  generates a certificate authority on the Opsman
  help                            prints this usage information
  import-installation             imports a given installation to the Ops Manager targeted
  installation-log                output installation logs
  installations                   list recent installation events
  pending-changes                 lists pending changes
  regenerate-certificates         regenerates a certificate authority on the Opsman
  revert-staged-changes           reverts staged changes on the Ops Manager targeted
  set-errand-state                sets state for a product's errand
  stage-product                   stages a given product in the Ops Manager targeted
  staged-manifest                 prints the staged manifest for a product
  staged-products                 lists staged products
  unstage-product                 unstages a given product from the Ops Manager targeted
  upload-product                  uploads a given product to the Ops Manager targeted
  upload-stemcell                 uploads a given stemcell to the Ops Manager targeted
  version                         prints the om release version

```
