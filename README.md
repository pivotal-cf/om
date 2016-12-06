# Om

_is a mantra, or vibration, that is traditionally chanted_

![enhancing your calm](http://i.giphy.com/3o7qDQ5iw1oXyDeJAk.gif)

## What it is

Magical tool that helps you configure and deploy tiles to an Ops-Manager. 
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

## Original design spec

please note, this can and will change from the final readme below

https://gist.github.com/ryanmoran/6cc4cfd33d7830753819139acf0e7c8d

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
  configure-authentication  configures Ops Manager with an internal userstore and admin user account
  configure-bosh            configures Ops Manager deployed bosh director
  configure-product         configures a staged product
  curl                      issues an authenticated API request
  delete-installation       deletes all the products on the Ops Manager targeted
  export-installation       exports the installation of the target Ops Manager
  help                      prints this usage information
  import-installation       imports a given installation to the Ops Manager targeted
  stage-product             stages a given product in the Ops Manager targeted
  upload-product            uploads a given product to the Ops Manager targeted
  upload-stemcell           uploads a given stemcell to the Ops Manager targeted
  version                   prints the om release version

```

## configure-bosh command
The `configure-bosh` command has several flags that take user-generated json. Here're some examples organized by various IaaS:

### GCP

#### --iaas-configuration
Note regarding `auth_json`: You will have likely generated this in order to use our [terraforming-gcp](https://github.com/pivotal-cf/terraforming-gcp/) tooling. To easily format this JSON for use as a string here, use `cat service_account_key.json | jq 'tostring'`.

Minimal example:

```json
{
  "project": "my-foo-project",
  "default_deployment_tag": "foo-vms",
  "auth_json": "{\"some-key\":\"some-value\"}"
}
```

#### --director-configuration
Note regarding `ntp_servers_string`: We recommend using this NTP server to all PCF users on GCP

Minimal example:
```json
{
  "ntp_servers_string": "169.254.169.254"
}
```

#### --az-configuration
We tend to use us-central1 because it has these 3 zones to balance across for HA deployments

Minimal example:
```json
{
  "availability_zones": ["us-central1-a","us-central1-b","us-central1-c"]
}
```

#### --security-configuration
No additional security configuration is strictly required.

Minimal example:
```json
{
  "trusted_certificates": "some-trusted-certificates",
  "vm_password_type": "generate"
}
```
