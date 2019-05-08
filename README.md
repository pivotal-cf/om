# Om

_is a mantra, or vibration, that is traditionally chanted_

![enhancing your calm](http://i.giphy.com/3o7qDQ5iw1oXyDeJAk.gif)

## What is it?

A tool that helps you configure and deploy tiles to Ops-Manager.
Currently being maintained by the PCF Platform Automation team,
with engineering support and review from PCF Release Engineering.
The (private) backlog for Platform Automation is [here](https://www.pivotaltracker.com/n/projects/1472134).

## Documentation

See [here](docs/README.md) for useful examples and documentation.

When working with `om`,
it can sometimes be useful to reference the Ops Manager API docs.
You can find them at
`https://pcf.your-ops-manager.example.com/docs`.

## Installation

To download `om` go to [Releases](https://github.com/pivotal-cf/om/releases).

Alternatively, you can install `om` via `apt-get` or `homebrew`
from package manager distributions maintained by [Stark and Wayne](https://www.starkandwayne.com/):
```sh
# apt-get:
wget -q -O - https://raw.githubusercontent.com/starkandwayne/homebrew-cf/master/public.key | apt-key add -
echo "deb http://apt.starkandwayne.com stable main" | tee /etc/apt/sources.list.d/starkandwayne.list
apt-get update

apt-get install om

# homebrew:
brew tap starkandwayne/cf
brew install om
```

You can also build from source.

### Building from Source
You'll need at least Go 1.11, as
`om` uses Go Modules to manage dependencies.

To build from source, after you've cloned the repo, run these commands from the top level of the repo:

```bash
GO111MODULE=on go mod download
GO111MODULE=on go build
```

Go 1.11 uses some heuristics to determine if Go Modules should be used.
The process above overrides those herusitics
to ensure that Go Modules are _always_ used.
If you have cloned this repo outside of your GOPATH,
`GO111MODULE=on` can be excluded from the above steps.

## Current Commands
```
ॐ
om helps you interact with an Ops Manager

Usage: om [options] <command> [<args>]
  --client-id, -c, OM_CLIENT_ID                          string  Client ID for the Ops Manager VM (not required for unauthenticated commands)
  --client-secret, -s, OM_CLIENT_SECRET                  string  Client Secret for the Ops Manager VM (not required for unauthenticated commands)
  --connect-timeout, -o, OM_CONNECT_TIMEOUT              int     timeout in seconds to make TCP connections (default: 10)
  --decryption-passphrase, -d, OM_DECRYPTION_PASSPHRASE  string  Passphrase to decrypt the installation if the Ops Manager VM has been rebooted (optional for most commands)
  --env, -e                                              string  env file with login credentials
  --help, -h                                             bool    prints this usage information (default: false)
  --password, -p, OM_PASSWORD                            string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  --request-timeout, -r, OM_REQUEST_TIMEOUT              int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)
  --skip-ssl-validation, -k, OM_SKIP_SSL_VALIDATION      bool    skip ssl certificate validation during http requests (default: false)
  --target, -t, OM_TARGET                                string  location of the Ops Manager VM
  --trace, -tr, OM_TRACE                                 bool    prints HTTP requests and response payloads
  --username, -u, OM_USERNAME                            string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  --version, -v                                          bool    prints the om release version (default: false)

Commands:
  activate-certificate-authority  activates a certificate authority on the Ops Manager
  apply-changes                   triggers an install on the Ops Manager targeted
  assign-multi-stemcell           assigns multiple uploaded stemcells to a product in the targeted Ops Manager 2.6+
  assign-stemcell                 assigns an uploaded stemcell to a product in the targeted Ops Manager
  available-products              list available products
  bosh-env                        prints bosh environment variables
  certificate-authorities         lists certificates managed by Ops Manager
  certificate-authority           prints requested certificate authority
  config-template                 **EXPERIMENTAL** generates a config template from a Pivnet product
  configure-authentication        configures Ops Manager with an internal userstore and admin user account
  configure-director              configures the director
  configure-ldap-authentication   configures Ops Manager with LDAP authentication
  configure-product               configures a staged product
  configure-saml-authentication   configures Ops Manager with SAML authentication
  create-certificate-authority    creates a certificate authority on the Ops Manager
  create-vm-extension             creates/updates a VM extension
  credential-references           list credential references for a deployed product
  credentials                     fetch credentials for a deployed product
  curl                            issues an authenticated API request
  delete-certificate-authority    deletes a certificate authority on the Ops Manager
  delete-installation             deletes all the products on the Ops Manager targeted
  delete-product                  deletes a product from the Ops Manager
  delete-ssl-certificate          deletes certificate applied to Ops Manager
  delete-unused-products          deletes unused products on the Ops Manager targeted
  deployed-manifest               prints the deployed manifest for a product
  deployed-products               lists deployed products
  diagnostic-report               reports current state of your Ops Manager
  download-product                downloads a specified product file from Pivotal Network
  errands                         list errands for a product
  export-installation             exports the installation of the target Ops Manager
  generate-certificate            generates a new certificate signed by Ops Manager's root CA
  generate-certificate-authority  generates a certificate authority on the Opsman
  help                            prints this usage information
  import-installation             imports a given installation to the Ops Manager targeted
  installation-log                output installation logs
  installations                   list recent installation events
  interpolate                     Interpolates variables into a manifest
  pending-changes                 lists pending changes
  pre-deploy-check                lists pending changes
  regenerate-certificates         deletes all non-configurable certificates in Ops Manager so they will automatically be regenerated on the next apply-changes
  revert-staged-changes           reverts staged changes on the Ops Manager targeted
  ssl-certificate                 gets certificate applied to Ops Manager
  stage-product                   stages a given product in the Ops Manager targeted
  staged-config                   **EXPERIMENTAL** generates a config from a staged product
  staged-director-config          **EXPERIMENTAL** generates a config from a staged director
  staged-manifest                 prints the staged manifest for a product
  staged-products                 lists staged products
  tile-metadata                   prints tile metadata
  unstage-product                 unstages a given product from the Ops Manager targeted
  update-ssl-certificate          updates the SSL Certificate on the Ops Manager
  upload-product                  uploads a given product to the Ops Manager targeted
  upload-stemcell                 uploads a given stemcell to the Ops Manager targeted
  version                         prints the om release version

```
