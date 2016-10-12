# Om

_is a mantra, or vibration, that is traditionally chanted_

## What it is

Magical tool that helps you configure and deploy tiles to an Ops-Manager. 
Currently being developed by RelEng, backlog link is [here](https://www.pivotaltracker.com/epic/show/2982497).

## Design Goals

- less flakey / faster replacement of [opsmgr](https://github.com/pivotal-cf/opsmgr)
- single binary that can be run on multiple platforms
- fully tested, not using tests to to execute browser behavior
- no capybara
- [small sharp tool](https://brandur.org/small-sharp-tools)
- idempotency for all commands

## Original design spec

please note, this can and will change from the final readme below

https://gist.github.com/ryanmoran/6cc4cfd33d7830753819139acf0e7c8d

## Current Commands

```
ॐ
om helps you interact with Ops Manager

Usage: om [options] <command> [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM
  -p, --password             string  admin password for the Ops Manager VM
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)

Commands:
  configure-authentication  configures Ops Manager with an internal userstore and admin user account
  export-installation       exports the installation of the target Ops Manager
  help                      prints this usage information
  upload-product            uploads a given product to the Ops Manager targeted
  upload-stemcell           uploads a given stemcell to the Ops Manager targeted
  version                   prints the om release version
```
