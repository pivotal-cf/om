# `om` Changelog

Nominally kept up-to-date as we work,
sometimes pushed post-release.

### Versioning

`om` went 1.0.0 on May 7, 2019

As of that release, `om` is [semantically versioned](https://semver.org/).
When consuming `om` in your CI system,
it is now safe to pin to a particular minor version line (major.minor.patch)
without fear of breaking changes.

#### API Declaration for Semver

Any changes to the `om` commands are considered a part of the `om` API.
Any changes to `om` commands will be released according to the semver versioning scheme defined above.
The exceptions to this rule are any commands marked as "**EXPERIMENTAL**"
- "**EXPERIMENTAL**" commands work, and pull information from the API
  same as any other. The format in which the information is returned, however,
  is subject to change without announcing a breaking change
  by creating a major or minor bump of the semver version. 
  When the `om` team is comfortable enough with the command output,
  the "**EXPERIMENTAL**" mark will be removed.
  
Any changes to the `om` filename as presented in the Github Release page.
  
Changes internal to `om` will _**NOT**_ be included as a part of the om API.
The `om` team reserves the right to change any internal structs or structures
as long as the outputs and behavior of the commands remain the same.

**NOTE**: Additional documentation for om commands 
leveraged by Pivotal Platform Automation 
can be found in [Pivotal Documentation](docs.pivotal.io/platform-automation).

`om` is versioned independently from platform-automation. 

### Tips
* Use environment variables
  to set what Ops Manager `om` is targeting.
  For example:
  ```bash
  $  export OM_PASSWORD=example-password om -e env.yml deployed-products
  ```
  Note the additional space before the `export` command.
  This ensures that commands are not kept in `bash` history.
  The environment variable `OM_PASSWORD` will overwrite the password value in `env.yml`.

## 6.2.0

### Features
- `configure-product`'s _decorating collection with guid_ logic has been extended to associate existing collection item guids based on (in order)
  - equivalent item values
  - equal logical keys (in order; ie. 'name' will be used over 'Filename' if both exist)
    - `name`
    - `key`
    - fields ending in `name` (eg: `sqlServerName`)

  This addresses [#207](https://github.com/pivotal-cf/om/issues/207); improving GitOps style workflows

## 6.1.0

### Features

## 6.1.1

### Bug Fixes
- When using `cache-cleanup`, globbing was not correctly done for files that contain the metadata prefix.
  This meant that files with `[pivnet-slug,pivnet-version]` will still laying around.

## 6.1.0

### Features
- [`--source pivnet` only]`download-product` now supports the `--check-already-uploaded` flag.
  If a valid env file is provided with the flag,
  `download-product` will attempt to check
  if the product is already present on the Ops Manager.
  If the product is already present,
  `download-product` will not attempt to download from Tanzu Network.
  This task is compatible with the `--stemcell-ias` flag.
  If provided, the task will also check if the stemcell is already uploaded
  before attempting to download from Tanzu Network

## 6.0.1

### Bug Fixes
- `download-product` will now correctly cache if downloading from a blobstore
  when `CACHE_CLEANUP='I acknowledge this will delete files in the output directories'`
  is set.
  
## 6.0.0

### Features
- With some code refactoring,
  we've introduced support for `--vars`, `--vars-file`, and `--vars-env`
  into places it was missing before.
- `download-product` can now provide a separate `--stemcell-output-directory`
  for the downloaded stemcell to exist after downloading.
  This was added to take advantage of Concourse 5.0+'s ability
  to overlap the cache in the output.
- [`download-product`] To allow `pas-windows` to count as a cache hit
  even after winfs has been injected,
  the shasum check on the cache has been removed.
  `download-product` will still check the shasum
  after the product has been downloaded from Tanzu Network
 - [`download-product`] A new env var, `CACHE_CLEANUP` has been added.
   When `CACHE_CLEANUP='I acknowledge this will delete files in the output directories'`
   it will delete all products that do not match the slug and version
   in the output directory so the local (or Concourse) cache can remain clean.
   This env var will also clean up all old stemcells
   from the `output-directory` (or `stemcell-directory` if defined)
   if `--stemcell-iaas` is provided.
- `certificate-authority` no longer requires `--id`
  if there is only one certificate authority on the targeted Ops Manager.
  This resolves PR [#501](https://github.com/pivotal-cf/om/pull/501/)

### Breaking Changes
- With some code refactoring,
  we removed the short form of `-v` for `--product-version`
  found in `download-product` and `stage-product`.

## 5.0.0

### Breaking Changes
- Removed deprecated `tile-metadata` command.
  Please use `product-metadata` command.
- Removed deprecated `update-ssl-certificate` command.
  Please use `configure-opsman` command.
- Removed depreated `--download-stemcell` flag from `download-product`.
  If the `--stemcell-iaas` is defined, it will always download the stemcell, and has done so, for a long time.

### Features
- Everything marked as `**EXPERIMENTAL**` has been promoted to officially supported.
  - `bosh-diff` command
  - `config-template` command
  - `OM_VARS_ENV` global flag
  - `OM_VARS_ENV` flag under `configure-*-authentication` commands
- The `config-template` command
  will now generate one ops file for each collection
  when the `--size-of-collections` flag is provided.
  The number of elements in each of those ops files
  is now based on that flag rather than having an ops file
  for each number up to the `--size-of-collections` value.
  The default behaviour of `config-template`
  without the `--size-of-collections` flag remains unchanged.

### Bug Fixes
- `apply-changes --product-name <product> --config config.yml` with errands defined in `config.yml` 
  that were not in the `product-name` list would fail.
  An explicit breakdown of how these flags interact:
  
  - `apply-changes` with the `product-name` flag(s) defined
    - `--config config.yml` with different products defined than provided in the `product-name` list:
      - Succeeds with a warning message, but does not apply the errand, if a product exists in the `config.yml` file, but was not passed in the `product-name` list.
      - Fails if any products in the `product-name` list also exist in the `config.yml` but do not exist on Ops Manager as staged/installed.
    - `--config config.yml` has no product defined: succeeds.
    - `--config config.yml` with same products defined as provided in the `product-name` list: succeeds.
  - `apply-chages` with NO `product-name` flag(s) defined
    - `--config config.yml` with different products defined than what exists in Ops Manager: failure.
      - If applying changes to all products, the products in `config.yml` _must be_ staged/installed.
    - `--config config.yml` has no product defined: succeeds.
    - `--config config.yml` with same products defined as what exists in Ops Manager (does not need to include all staged/installed products): succeeds.
- `interpolate` will no longer append a newline to end of the output

## 4.8.0

### Features
- `download-product` now supports defining a custom `stemcell-version` in the config file.
  This flag is `--stemcell-version`, and requires `--stemcell-iaas` to be set.
  If `--stemcell-version` is not set, but `stemcell-iaas` is set,
  the command will download the latest stemcell for the product.
- `bosh-diff` now supports the `--check` flag. 
  If set, the command will fail if there are differences returned.
  This resolves issue [#488](https://github.com/pivotal-cf/om/issues/488)]
- `stage-product` now accepts a config file to define command line args.
  This matches patterns for automation available in other commands.

## 4.7.0

### Features
- `configure-opsman` command has been added.
  This allows configuration of several Ops Manager settings.
  Most were previously not available to configure through an `om` command,
  though `ssl-certificate` was also configurable with `update-ssl-certificate`.
  For config examples,
  see the [docs](https://github.com/pivotal-cf/om/tree/master/docs/configure-opsman)
  for the command.
  Supported top-level-keys:
    - `ssl-certificate`
    - `pivotal-network-settings`
    - `rbac-settings`
    - `banner-settings`
    - `syslog-settings`
- **EXPERIMENTAL** `config-template` now supports ops manager syslog in tiles.
  In the tile metadata, this property is turned on with the `opsmanager_syslog: true` field.
  Tiles with this property enabled will now add the section to `product.yml` 
  and create defaults in `default-vars.yml`.
- Added shorthand flag consistency to multiple commands.
  `--vars-file` shorthand is `-l` and `--var` shorthand is `-v`
- **EXPERIMENTAL** `config-template` can specify the number of collection ops files using `--size-of-collections`.
  Some use cases required that collections generate more ops-file for usage.
  The default value is still `10`.
- `config-template` has been updated to include placeholders for
  `network_name`, `singleton_availability_zone`, and `service_network_name`
  in `required-vars.yml` when appropriate.
- When using `apply-changes --recreate`, Ops Manager will recreate director VM on OM 2.9+
  If a product name is passed (`apply-changes --product-name <product> --recreate`),
  only the product VMs will be recreated. 
  When using `apply-changes --recreate --skip-deploy-products`,
  only the director VM will be recreated.
  This resolves issue [#468](https://github.com/pivotal-cf/om/issues/468)

### Bug Fixes
- Cleaned up all the interpolation to be more consistent with the behaviour of the `bosh` CLI.

  For example,
  
  ```bash
  # with a variable
  $ om interpolate -c <(echo "person: ((person))") -v person="{foo: bar}"
  person:
    foo: bar
  # with an env var
  $ PREFIX_person="{foo: bar}" om interpolate -c <(echo "person: ((person))") --vars-env PREFIX
  person:
    foo: bar
  ```
  
  We did maintain,
  when using environment variables or var (`-v`),
  a multiline string needs to be maintained.
  The `bosh` does not support this.

### Breaking Changes for Experimental Features
- `config-template` Bug Fix: Required collections now parametrize correctly in `product.yml`.
  In the [om issue](https://github.com/pivotal-cf/om/issues/483)
  for `p-dataflow`, the following was _incorrectly_ returned:
  ```
  .properties.maven_repositories:
    value:
    - key: spring
      password: ((password))
      url: https://repo.spring.io/libs-release
      username: username
  ```

  `config-template` now returns the following correct subsection in `product.yml`:
  ```
  .properties.maven_repositories:
    value:
    - key: spring
      password:
        secret: ((password))
      url: https://repo.spring.io/libs-release
      username: username
  ```

  **if you have used the workaround described in the issue**
  (storing the value as a JSON object)
  you will need to update the credential in Credhub
  to not be a JSON object.
- `config-template` generated `resource-vars.yml`
  that had the potential to conflict with property names
  (spring cloud dataflow had a configurable property called `max_in_flight`
  which is also a resource config property).
  `config-template` now prepends **all** resource-vars with `resource-var-`.
  This prevents this entire class of conflicts.
  If using `config-template` to update vars/ops-files/etc,
  check your resource var names in any files vars may be drawn from.
  This resolves om issue [#484](https://github.com/pivotal-cf/om/issues/484).

### Deprecation Notices
- `update-ssl-certificate` has been deprecated in favor of `configure-opsman`.
  This was done to allow extensibility for other endpoints on the Settings page.
  Please note that `configure-opsman` requires a config file, and does not accept
  `certificate-pem` or `private-key-pem` as command line arguments. 
  For config example, see the [docs](https://github.com/pivotal-cf/om/tree/master/docs/configure-opsman) for the command.

## 4.6.0

### Features
- `configure-authentication` now supports
  the `OM_DECRYPTION_PASSPHRASE` environment variable.

### Bug Fixes
- `configure-director` now correctly handles when you don't name your iaas_configuration `default` on vSphere.
  Previously, naming a configuration anything other than `default` would result in an extra, empty `default` configuration.
  This closes issue [#469](https://github.com/pivotal-cf/om/issues/469).
- Downloading a stemcell associated with a product will try to download the light or heavy stemcell.
  If anyone has experienced the recent issue with `download-product`
  and the AWS heavy stemcell,
  this will resolve your issue.
  Please remove any custom globbing that might've been added to circumvent this issue.
  For example, `stemcall-iaas: light*aws` should just be `stemcell-iaas: aws` now. 
- Heavy stemcells could not be downloaded. 
  Support has now been added.
  Force download of the heavy stemcell (if available) with the `--stemcell-heavy` flag.

## 4.5.0

### Features
- `interpolate` now supports the dot notation to reference array values.
  For example,

  ```bash
  $ echo "person: ((people.1))" | om interpolate -c - -l <(echo "people: [Bob, Susie, Diane]")
  person: Susie
  ```
  
## 4.4.2

### Features
- To mitigate confusion, the `pivnet-file-glob` param for `download-product` now has an alias of `file-glob`.
- `update-ssl-certificate` now supports passing `certificate-pem` and `private-key-pem`
  as `--config` arguments. The command also supports the `--vars-file` flag for interpolation.
  This closes issue [#463](https://github.com/pivotal-cf/om/issues/463).

## 4.4.1

### Features
- The experimental command `product-diff` had been renamed `bosh-diff`
  and now includes the director diff.
  This includes property, runtime config, cloud config, and CPI config differences.
  When the command is used, it will display the director and all products by default.
  The `--director` flag can be used to show only the director diff.
  The `--product-name` flag can be used to show one or more specific products.
  
  For example, `om bosh-diff --director --product-name cf --product-name p-healthwatch`
  will show the director, Pivotal Application Service, and Pivotal Healthwatch differences.

## 4.4.0

### Features
- The experimental command `product-diff` has been added.
  It gets the manifest diff for a specified set of products.
  This might be useful as a sanity review before apply-changes;
  see the detailed documentation for details/provisos.
- **EXPERIMENTAL** `config-template` now includes the option to use a local product file with `--product-path`.
  This resolves issue [#413](https://github.com/pivotal-cf/om/issues/413).
- `apply-changes` can for recreate the VMs that will apply with `--recreate-vms`.
  This is useful for the [three-Rs of security](https://devopedia.org/three-rs-of-security),
  to ensure the _repaving_.

### Bug Fixes
- The environment variable `OM_VARS_ENV` was not enabled on all commands that allows `--vars-env`.

## 4.3.0

### Features
- We'd like to welcome back the `revert-staged-changes` command.
  It requires an API endpoint released in Ops Manager versions 2.5.21+, 2.6.13+, or 2.7.2+.
  This now reverts changes like the the equivalant "Revert" button in the UI.
  Appropriate messages and warnings will appear from the command of what action has been taken.

  In v3.0.0, we removed `revert-staged-changes` because it had stopped working.
  (The necessary Ops Manager API wasn't present, so it was trying to automate
  through the UI - unsuccessfully).

### Bug fixes
- Maybe not technically a bug, but: 
  some commands you love (`pre-deploy-check`, `staged-config`, and `staged-director-config`)
  no longer have the EXPERIMENTAL tag.
  Nothing has changed with them, we literally just forgot to remove these ages ago.

## 4.2.1

### Bug fixes
* `interpolate` command now has order precedence when a file or stdin is provided.
  - `--config` with a file always takes precedence
  - `--config -` will read directly from STDIN
  - STDIN provided with no `--config` will use STDIN
* when using `--ignore-verifier-warnings` with `configure-director` the HTTP Status 207
  will be ignored when interacting with IAAS endpoints.

## 4.2.0

### Features
- **EXPERIMENTAL** `config-template` now includes `max-in-flight` for all resources. (PR: @jghiloni)
- When using `configure-product` and `configure-director`,
  the `additional_vm_extensions` for a resource will have the following behaviour:
  * If not set in config file, the value from Ops Manager will be persisted.
  * If defined in the config file and an emtpy array (`[]`), the values on Ops Manager will be removed.
  * If defined in the file with a value (`["web_lb"]`), these values will be set on Ops Manager.
- `configure-authentication`, `configure-ldap-authentication`, and `configure-saml-authentication`
  now support the `--var`, `--vars-file`, and `--vars-env` flags. 
- **EXPERIMENTAL** `config-template` now supports the `--config`, `--var`, `--vars-file`, and `--vars-env` flags.
  (PR: @jghiloni)

## 4.1.0

### Features
- `download-product` supports GCS (Google Cloud Storage)
  for Tanzu Network download artifacts.
  
  An example config,
  
  ```yaml
  pivnet-file-glob: "*.tgz"
  pivnet-product-slug: pivotal-telemetry-collector
  product-version: "1.0.1"
  output-directory: /tmp
  source: gcs
  gcs-bucket: some-bucket
  gcs-service-account-json: |
    {account-JSON}
  gcs-project-id: project-id
  ```
  
  This will download the `[pivotal-telemetry-collector,1.0.1]telemetry-collector-1.0.1.tgz`
  from the `some-bucket` bucket from the GCS account.
  
- `download-product` supports Azure Storage.
  for Tanzu Network download artifacts.
  
  ```yaml
  pivnet-file-glob: "*.tgz"
  pivnet-product-slug: pivotal-telemetry-collector
  product-version: "1.0.1"
  output-directory: /tmp
  source: azure
  azure-container: pivnet-blobs
  azure-storage-account: some-storage-account
  azure-key: "storage-account-key"
  ```

  This will download the `[[pivotal-telemetry-collector,1.0.1]telemetry-collector-1.0.1.tgz`
  from the `pivnet-blobs` container
  from the `some-storage-account` storage account from Azure Storage.

- The commands `disable-director-verifiers`
  and `disable-product-verifiers` have been added.
  They allow verifiers that are preventing Apply Changes from succeeding to be disabled.
  This feature should be used with caution,
  as the verifiers can provide useful feedback on mis-configuration. 

- When using `staged-director-config` and `configure-director`,
  the `iaas_configuration_name` will be used to assign an IAAS to an availability zone.
  This provides support for multiple iaas configurations on vSphere and Openstack.
  Prior to this, the `iass_configuration_guid` had to be discovered prior to assigning an availability zone;
  now the name can be used in one step.
  
- We've also made miscellanious improvements
  to warning and error messages,
  and to documentation.

## 4.0.1

### Bug Fixes
- The `ca-cert` option works in the `env.yml`.
  A filename or string value can be used.

## 4.0.0

### Breaking Changes
- `apply-changes` will no longer reattach when it finds an already running installation.
  to re-enable this feature, provide the `--reattach` flag.
  This makes the behavior of `apply-changes` easier to anticipate
  and specify whether applying all changes or applying changes to a particular product.  
  
### Features
- **EXPERIMENTAL** `config-template` now accepts `--pivnet-file-glob` instead of `--product-file-glob`.
  This is to create consistency with the `download-product` command's naming conventions.
  (PR: @poligraph)
## 3.2.2

### Bug Fixes
- `staged-config` will now work again for Ops Manager versions <= 2.3.
  This solves issue [#419](https://github.com/pivotal-cf/om/issues/419)

## 3.2.1

### Bug Fixes
- `configure-director` now will configure VM Extensions before setting Resource Config.
  This fixes issue [#411](https://github.com/pivotal-cf/om/issues/411)   
  
## 3.2.0

### Features
- `expiring-certificates` command was added.
  This command returns a list of certificates
  from an Ops Manager
  expiring within a specified (`--expires-within/-e`) time frame. 
  Default: "3m" (3 months)
  Root CAs cannot be included in this list until Ops Manager 2.7.
- `configure-product` and `staged-config` now have support for the `/syslog_configurations` endpoint. 
  This affects tiles, such as the Metrics tile,
  that do not return these properties nested in the `product-properties` section. 
  This provides a solution for issue [331](https://github.com/pivotal-cf/om/issues/331).
  An example of this inside of your product config:
  
    ```yaml
    syslog-properties:
      address: example.com
      custom_rsyslog_configuration: null
      enabled: true
      forward_debug_logs: false
      permitted_peer: null
      port: "4444"
      queue_size: null
      ssl_ca_certificate: null
      tls_enabled: false
      transport_protocol: tcp
    ```
- `generate-certificate` can now accept multiple `--domains | -d` flags.
  Comma separated values can be passed with a single `--domains | -d` flag,
  or using a `--domains | -d` flag for each value. (PR: @jghiloni)
  Example:
    ```bash
      om -e env.yml generate-certificate -d "example1.com" --domains "example2.com" \
         -d "example3.com,*.example4.com" --domains "example5.com,*.example6.com"
    ```
- `product-metadata` has been added to replace `tile-metadata`.
  This was done to increase naming consistency.
  Both commands currently exist and do exactly the same thing.
  (PR: @jghiloni)
- **EXPERIMENTAL** `config-template` now supports the `--exclude-version` flag.
  If provided, the command will exclude the version directory in the `--output-directory` tree.
  The contents will with or without the flag will remain the same.
  Please note including the `--exclude-version` flag
  will make it more difficult to track changes between versions
  unless using a version control system (such as git).
  (PR: @jghiloni)
- **EXPERIMENTAL** `config-template` supports `--pivnet-disable-ssl` to skip SSL validation.
- When interacting with an OpsManager, that OpsManager may have a custom CA cert.
  In the global options `--ca-cert` has been added to allow the usage of that custom CA cert.
  The value of `--ca-cert` can be a file or command line string.
  
### Bug Fix
- When using `config-template` (**EXPERIMENTAL**) or `download-product`,
  the `--pivnet-skip-ssl` is honored when capturing the token. 

### Deprecation Notices
- `tile-metadata` has been deprecated in favor of `product-metadata`.
  This was done to increase naming consistency.
  Both commands currently exist and do exactly the same thing.
  The `tile-metadata` command will be removed in a future release.
  
## 3.1.0

### Features

- TLS v1.2 is the minimum version supported when connecting to an Ops Manager
- **EXPERIMENTAL** `config-template` now will provide required-vars in addition to default-vars.
- **EXPERIMENTAL** `config-template` will define vars with an `_` instead of a `/`.
  This is an aesthetically motivated change.
  Ops files are denoted with `/`,
  so changing the vars separators to `_` makes this easier to differentiate.
- **EXPERIMENTAL** `config-template` output `product-default-vars.yml` has been changed to `default-vars.yml`
- `staged-config` includes the property `max_in_flight` will be included
  in the `resource-config` section of a job.
- `configure-product` can set the property `max_in_flight`
  in the `resource-config` section of a job.

  The legal values are:
  * an integer for the number of VMs (ie `2`)
  * a percentage of 1-100 (ie `20%`)
  * the default value specified in tile (`default`)
  For example,

  ```yaml
  resource-config:
    diego_cells:
      instances: 10
      max_in_flight: 10
  ```

## 3.0.0

### Features

- `pivnet-api-token` is now optional in `download-product`
  if a source is defined. (PR: @vchrisb)
- `configure-authentication`, `configure-ldap-authentication`, and `configure-saml-authentication`
  can create a UAA client on the Ops Manager vm.
  The client_secret will be the value provided to this option `precreated-client-secret`.
- add support for NSX and NSXT in Ops Manager 2.7+
  
### Breaking Changes

- remove `--skip-unchanged-products` from `apply-changes`
  This option has had issues with consistent successful behaviour.
  For example, if the apply changes fails for any reason, the subsequent apply changes cannot pick where it left off.
  This usually happens in the case of errands that are used for services.
  
  We are working on scoping a selective deploy feature that makes sense for users.
  We would love to have feedback from users about this.
  
- remove revert-staged-changes
  unstage-product will revert the changes if the tile has not been installed.
  There is currently no replacement for this command,
  however, it was not working for newer versions of Ops Manager, and did nothing. 
  This resolves issue [#399](https://github.com/pivotal-cf/om/issues/399)
  
### Bug Fix
- `apply-changes` will error with _product not found_ if that product has not been staged.
- `upload-stemcell` now accepts `--floating false` in addition to `floating=false`.
  This was done to offer consistency between all of the flags on the command.
- `configure-director` had a bug in which `iaas_configurations` could not be set
  on AWS/GCP/Azure because "POST" was unsupported for these IAASes
  (Multiple IAAS Configurations only work for vSphere and Openstack).
  `configure-director` will now check if the endpoint is supported.
  If it is not supported, it will construct a payload, and selectively configure
  iaas_configuration as if it were nested under `properties-configuration`. 
  _The behavior of this command remains the same._ 
  IAAS Configuration may still be set via `iaas_configurations` OR `properties.iaas_configuration`  
  

## 2.0.1

Was a release to make sure that `brew upgrade` works.

## 2.0.0

### Features
- `configure-ldap-authentication` and `configure-saml-authentication` can create a UAA client on the Ops Manager vm.
  The client_secret will be the value provided to this option `precreated-client-secret`.
  This is supported in OpsManager 2.5+.
- A homebrew formula has been added!
  It should support both linux and mac brew.
  Since, we don't have our own `tap`, we've used a simpler method:
  ```bash
  brew tap pivotal-cf/om https://github.com/pivotal-cf/om
  brew install om
  ```
  
### Bug Fixes
- The order of vm types and resources was being applied in the correct order.
  Now vm types will be applied then resources, so that resource can use the vm type.
- When using `bosh-env`, a check is done to ensure the SSH private key exists.
  If does not the command will exit 1.
- **EXPERIMENTAL** `config-template` will enforce the default value for a property to always be `configurable: false`.
  This is inline with the OpsManager behaviour.
  
### Breaking Change
- The artifacts on the Github Release include `.tar.gz` (for mac and linux) and `.zip` (windows) for compression.
  It also allows support for using `goreleaser` (in CI) to create other package manager artifacts -- `brew`.
  This will break globs that were permissive. For example `*linux*`, will download the binary and the `.tar.gz`, use `*linux*[^.gz]` to just download the binary.
  Our semver API declaration has been updated to reflect this.

## 1.2.0

### Features 
* Both `om configure-ldap-authentication` 
  and `om configure-saml-authentication`
  will now automatically
  create a BOSH UAA admin client as documented [here](https://docs.pivotal.io/pivotalcf/2-5/customizing/opsmanager-create-bosh-client.html#saml).
  This is only supported in OpsManager 2.4 and greater.
  You may specify the flag `skip-create-bosh-admin-client`
  to skip creating this client.
  If the command is run for an OpsManager less than 2.4,
  the client will not be created and a warning will be printed.
  However, it is recommended that you create this client.
  For example, your SAML or LDAP may become unavailable,
  you may need to sideload patches to the BOSH director, etc.
  Further, in order to perform automated operations on the BOSH director,
  you will need this BOSH UAA client.
  After the client has been created,
  you can find the client ID and secret
  by following [steps three and four found here](https://docs.pivotal.io/pivotalcf/2-5/customizing/opsmanager-create-bosh-client.html#-provision-admin-client).
* `om interpolate` now allows for the `-v` flag
  to allow variables to be passed via command line. 
  Command line args > file args > env vars.
  If a user passes a var multiple times via command line,
  the right-most version of that var will
  be the one that takes priority,
  and will be interpolated.
* `om configure-director` now supports custom VM types.
  (PR: @jghiloni)
  Refer to the [VM Types Bosh documentation](https://bosh.io/docs/cloud-config/#vm-types) for IaaS specific use cases.
  For further info: [`configure-director` readme](https://github.com/pivotal-cf/om/tree/master/docs/configure-director#vmtypes-configuration). 
  Please note this is an advanced feature, and should be used at your own discretion.
* `download-product` will now return a `download-file.json` 
  if `stemcell-iaas` is defined but the product has no stemcell.
  Previously, this would exit gracefully, but not return a file.
  
## 1.1.0

### Features
* (**EXPERIMENTAL**) `pre-deploy-check` has been added as a new command.
  This command can be run at any time. 
  It will scan the director and any staged tiles
  in an Ops Manager environment for invalid or missing properties.
  It displays these errors in a list format 
  for the user to manually (or automatedly) update the configuration.
  This command will also return an `exit status 1`;
  this command can be a gatekeeper in CI 
  before running an `apply-changes`
* `download-product` will now include the `product-version` in `download-file.json`
  (PR: @vchrisb)

### Bug Fixes
* Extra values passed in the env file 
  will now fail if they are not recognized properties.
  This closes issue [#258](https://github.com/pivotal-cf/om/issues/258)
* Non-string environment variables can now be read and passed as strings to Ops Manager.
  For example, if your environment variable (`OM_NAME`) is set to `"123"` (with quotes escaped),
  it will be evaluated in your config file with the quotes.
  
    Given `config.yml`
    ```yaml
    value: ((NAME))
    ```
    
    `om interpolate -c config.yml --vars-env OM`
    
    Will evaluate to:
    ```yaml
      value: "123"
    ```
  This closes issue [#352](https://github.com/pivotal-cf/om/issues/352)
* the file outputted by `download-product`
  will now use the `product-name` as defined 
  in the downloaded-product, 
  _not_ from the Tanzu Network slug.
  This fixes a mismatch between the two
  as documented in issue [#351](https://github.com/pivotal-cf/om/issues/351)
* `bosh-env` will now set `BOSH_ALL_PROXY` without a trailing slash
  if one is provided.
  This closes issue [#350](https://github.com/pivotal-cf/om/issues/350) 

## 1.0.0

### Breaking Changes
* `om` will now follow conventional Semantic Versioning,
  with breaking changes in major bumps,
  non-breaking changes for minor bumps,
  and bug fixes for patches.
* `delete-installation` now has a force flag. 
  The flag is required to run this command quietly, as it was working before.
  The reason behind this is
  it was easy to delete your installation without any confirmation. 
* `staged-director-config` no longer supports `--include-credentials`
  this functionality has been replaced by `--no-redact`.
  This can be paired with `--include-placeholders`
  to return a interpolate-able config
  with all the available secrets from a running OpsMan.
  This closes issue #356. 
  The OpsMan API changed so that IAAS Configurations
  were redacted at the API level. 

### Features
* new command `diagnostic-report`
  returns the full OpsMan diagnostic report
  which holds general information about the
  targeted OpsMan's state.
  Documentation on the report's payload
  can be found [here.](https://docs.pivotal.io/pivotalcf/2-2/opsman-api/#diagnostic-report)
* `om interpolate` now can take input from stdin.
  This can be used in conjunction with the new
  `diagnostic-report` command to extract
  a specific section or value
  from the report, simply by using the pipe operator. For example,
  ```bash
  om -e env.yml diagnostic-report | om interpolate --path /versions
  ```
  This will return the `versions` block of the json payload:
  ```yaml
  installation_schema_version: "2.6"
  javascript_migrations_version: v1
  metadata_version: "2.6"
  release_version: 2.6.0-build.77
  ```
* `staged-director-config` now checks
  `int`s and `bool`s when filtering secrets
* `configure-director` and `staged-director` now support `iaas-configurations`.
  This allows OpsManager 2.2+ to have multiple IAASes configured.
  Please see the API documentation for your version of OpsMan for what IAASes are supported.
  
  If you are using `iaas_configuration` in your `properties-configuration` and use `iaas-configurations`
  you'll receive an error message that only one method of configuration can be used. 

## 0.57.0

### Features
* new command `assign-multi-stemcell` supports the OpsMan 2.6+.
  This allows multiple stemcells to be assgined to a single product.
  For example, for product `foo`,
  you could assign Ubuntu Trusty 3586.96 and Windows 2019 2019.2,
  using the command, `om assign-multi-stemcell --product foo --stemcell ubuntu-trusty:3586.96 --stemcell windows2019:2019.2`.
* `upload-stemcell` will not upload the same stemcell (unless using `--force`) for OpsMan 2.6+.
  The API has changed that list the stemcells associated with a product.
  This command is still backwards compatible with OpsMan 2.1+, just has logic specific for 2.6+.

## NOTES
* https://github.com/graymeta/stow/issues/197 has been merged! This should make `om` `go get`-able again.

## 0.56.0

### Breaking Changes
* the `upload-product` flag `--sha256` has been changed to `--shasum`. `upload-stemcell`
  used the `--shasum` flag, and this change adds parity between the two. Using 
  `--shasum` instead of `--sha256` also future-proofs the flag when sha256 is no longer the
  de facto way of defining shasums.

### Features
* `download-product` now supports skipping ssl validation when specifying `--pivnet-disable-ssl`
* `download-product` ensures sha sum checking when downloading the file from Pivotal Network
* `upload-stemcell` now supports a `--config`(`-c`) flag to define all command line arguments
   in a config file. This gives `upload-stemcell` feature parity with `upload-product`
* Improved info messaging for `download-product` to explicitly state whether downloading
  from pivnet or S3
 

## 0.55.0

### Features
* configure-director now has the option to `ignore-verifier-warnings`.
  (PR: @Logiraptor)
  This is an _advanced_ feature
  that should only be used if the user knows how to disable verifiers in OpsManager.
  This flag will only disable verifiers for configure-director,
  and will not disable the warnings for apply-changes.
* There's now a shell-completion script;
  see the readme for details.
* We have totally replaced the code and behavior
  of the **EXPERIMENTAL** `config-template` command.
  It now contains the bones of the [tile-config-generator](https://github.com/pivotalservices/tile-config-generator).
  We expect to further refine
  (and make breaking changes to) this command in future releases.

## 0.54.0

### Breaking Changes
* download-product's prefix format and behavior has changed.
  - the prefix format is now `[example-product,1.2.3]original-filename.pivotal`.
  - the prefix is added to _all_ product files if `s3-bucket` is set in the config when downloading from Pivnet.

### Features
* download-product now supports downloading stemcells from S3, too.
* download-product allows use of an instance iam account when `s3-auth-method: iam` is set.
* apply-changes now has the ability to define errands via a config file when running (as a one-off errand run).
  The [apply-changes readme](https://github.com/pivotal-cf/om/docs/apply-changes/README.md) details how this 
  config file should look.
* pending-changes now supports a `--check` flag, that will return an exit code 0(pass) or 1(fail) when running the command, 
  to allow you to fail in CI if there are pending changes in the deployment. 
* download-product will now create a config file (`assign-stemcell.yml`) that can be fed into `assign-stemcell`. It will have the appropriate
  format with the information it received from download-product


### Bug Fixes
* when trying to delete a product on Ops Manager during a selective deploy (`apply-changes --product-name tile`),
  OpsManager would fail to `apply-changes` due to a change to the version string for 2.5 (would include the build
  number). A change was made to the info service to accept the new semver formatting as well as the old 
  versioning. 
* upload-product (among other things) is no longer sensitive to subdirectories in tile metadata directories
* to support 2.5, new semver versioning for OpsManager was added in addition to supporting the current versioning format.
  (PR: @jplebre & @edwardecook)
  
### WARNING

To anyone who is having go install fail, it will fail until graymeta/stow#199 is merged.

Here is the error you are probably seeing.

```
$ go install
# github.com/pivotal-cf/om/commands
commands/s3_client.go:62:3: undefined: s3.ConfigV2Signing
```
to work around, you can include `om` in your project without using `go get` or `go install`. you will need to add the following to your `go.mod`:
```
replace github.com/graymeta/stow => github.com/jtarchie/stow v0.0.0-20190209005554-0bff39424d5b
```

## 0.53.0 

### Bug Fixes

* `download-product` would panic if the product was already downloaded and you asked for a stemcell. This has been fixed to behave appropriately

### WARNING

The behavior of `download-product` in this release is not final. Please hold off on using this feature until a release without this warning.

## 0.52.0
### Breaking changes
* `download-product` will now enforce a prefix of `{product-slug}-{semver-version}` when downloading from pivnet. The original
  filename is preserved after the prefix. If the original filename already matches the intended format, there will be no
  change. Any regexes that strictly enforce the original filename at the _beginning_ of the regex will be broken. Please
  update accordingly. This change was done in order to encourage tile teams to change their file names to be more consistent. 
  Ops Manager itself has already agreed to implement this change in newer versions. 

### Features
* add support for the `selected_option` field when calling `staged-config` to have better support for selectors.
  * this support also extends to `configure-product`, which will accept both `selected_option` and `option_value` as
  the machine readable value. 
* `download-product` now has support for downloading from an external s3 compatible blobstore using the `--blobstore s3`
  flag. 
* `staged-director-config` now supports a `no-redact` flag that will return all of the credentials from an Ops Manager
  director, if the user has proper permissions to do so. It is recommended to use the admin user. 
  
### WARNING

The behavior of `download-product` in this release is not final. Please hold off on using this feature until a release without this warning.

## 0.51.0 

### Features

* `import-installation` provides validation on the installation file to ensure
  * it exists
  * it is a valid zip file
  * it contains the `installation.yml` artifact indicative of an exported installation
  
### Bug Fixes

* Fixed typo in `configure-director` vmextensions

## 0.50.0

### Breaking changes

`configure-director` and `staged-director-config` now include a `properties-configuration`.

  The following keys have recently been removed from the top level configuration: director-configuration, iaas-configuration, security-configuration, syslog-configuration.
  
  To fix this error, move the above keys under 'properties-configuration' and change their dashes to underscores.
  
  The old configuration file would contain the keys at the top level.

```yaml
director-configuration: {}
iaas-configuration: {}
network-assignment: {}
networks-configuration: {}
resource-configuration: {}
security-configuration: {}
syslog-configuration: {}
vmextensions-configuration: []
```

  They'll need to be moved to the new 'properties-configuration', with their dashes turn to underscore.
  For example, 'director-configuration' becomes 'director_configuration'.
  The new configration file will look like.

```yaml
az-configuration: {}
network-assignment: {}
networks-configuration: {}
properties-configuration:
  director_configuration: {}
  security_configuration: {}
  syslog_configuration: {}
  iaas_configuration: {}
  dns_configuration: {}
resource-configuration: {}
vmextensions-configuration: []
```
### Features

* The package manager has been migrated from `dep` to `go mod`. It now requires golang 1.11.4+. For information on go modules usage, see the [golang wiki](https://github.com/golang/go/wiki/Modules).

### Bug Fixes

* `import-installation` will now retry 3 times (it uses the polling interval configuration) if it suspects that nginx has not come up yet. This fixes an issue with opsman if you tried to import an installation with a custom SSL Cert for opsman.
* When using `configure-product` on opsman 2.1, it would fail because the completeness check does not work. To disable add the field `validate-config-complete: false` to your config file.
* fix the nil pointer dereference issue in `staged-products` when `om` cannot reach OpsManager

## 0.49.0

### Features

* `download-product` supports grabbing for a version via a regular expression.
  Using `--product-version-regex` sorts the versions returned by semver and
  returns the highest matching version to the regex. The sort ignores non-semver
  version numbers -- similar to the pivnet resource behaviour.
* `download-product` no longer requires `download-stemcell` to be set when specifying `stemcell-iaas`. It is there for backwards compatibility, but it is a no-op.
* added more copy for the help message of `bosh-env`
* fix documentation for `vm-extensions` usage

## 0.48.0

### Features

* Increased the default connect-timeout from `5` seconds to `10`. This should alleviate reliability issues some have seen in CI.

* Adds several commands (`delete-ssl-certificate`, `ssl-certificate`, `update-ssl-certificate`) around managing the Ops Manager VM's TLS cert. These new commands are courtesy of a PR, and we're still tinkering a bit (especially in terms of how they communicate in the case of a default cert, given that the Ops Manager API doesn't even bother returning a cert in that case).
  There should be a fast-to-follow release with these commands more polished; if we'd planned better we might have marked these as experimental until we were done tinkering with them, but we don't see any reason to delay releasing this time.

## 0.47.0

### Breaking changes

* `stage-product` & `configure-product` & `configure-director`: Now errors if `apply-changes` is running. [a3ebd5241d2aba3b93ec642255e0b9c11686d996]

### Features

* `configure-ldap-authentication`: add the command to configure ldap auth during initial setup

### Bug Fixes

* `assign-stemcell`: fix a message format

## 0.46.0

### Breaking changes

* download-product now outputs `product_path`, `product_slug`, `stemcell_path`, and `stemcell_version` instead
  of just `product` and `stemcell`. This will help compatability with `assign-stemcell`.

## 0.45.0

### Breaking changes

* removed individual configuration flags for configure-director \[[commit](https://github.com/pivotal-cf/om/commit/669eca466ca364e4d7597330e5600a013ab9ffe3)\]
* removed individual configuration flags for configure-product \[[commit](https://github.com/pivotal-cf/om/commit/040651b211b8985879337c86357727546099c46e)\]

### Features

* add more intelligent timeouts for commands
* fail fast if a key is not defined in configuration files for configure-product and configure-director
* add `assign-stemcell` command to associate a specified stemcell to the product

### Bug Fixes

* fix stemcell version check logic in `download-product` command -- stemcells can now be downloaded even if they
don't have a minor version (e.g. version 97)

## 0.44.0

### Bug fixes

* The decryption passphrase check was returning dial timeout errors more frequently. Three HTTP retries were added if dial timeout occurs. [Fixes #283]

## 0.43.0

### Breaking changes

* removed command `configure-bosh`, use command `configure-director` for configuring the bosh directory on OpsMan
* removed command `set-errand-state`, use the `errand-config` with your config with the command `configure-product`

### Features

* add command `download-product`, it can download product and associated stemcell from Pivnet
* add `--path` to command `interpolate` so individual values can be extracted out

### Bug Fixes

* automatic decryption passphrase unlock will only attempt doing so once on the first HTTP call #283
* when using command `configure-product`, collections won't fail when `guid` cannot be associated #274

## 0.42.0

### Breaking changes:
* `config-template` (**EXPERIMENTAL**) & `staged-config` & `staged-director-config`: pluralize `--include-placeholders` flag
* `import-installation`: removed `decryption-passphrase` from the arguments. Global `decryption-passphrase` flag is required when using this command

### Bug Fixes
* update command documentation to reflect various command flags change.
* `configure-product`: handles collection types correctly by decorate collection with guid
* `staged-director-config`: fix failed api request against azure
* `curl`: close http response body to avoid potential resource leaks

### Features
* `configure-product`: allow `product-name` be read from config file
* `interpolation`: added `--vars-env` support to `interpolation`
* `configure-authentication` & `configure-saml-authentication` & `import-installation`: allow the commandline flag been passed through config file
* `configure-director`: able to add/modify/remove vm extensions
*  `staged-config`: able to get errand state for the product
* `apply-changes`: added `skip-unchanged-products`
* `staged-config`: add `product-name` top-level-key in the returned payload to work better with `configure-product`
* `upload-product`: able to validate `sha256` and `product-version` before uploading
* global: added a `decryption-passphrase` to unlock the opsman vm if it is rebooted (if provided)

## 0.40.0

### Bug Fixes

Fix `tile-metadata` command for some tiles that were failing due to it attempting to parse the metadata directory itself as a file - via @chendrix and @aegershman

## 0.39.0

BACKWARDS INCOMPATIBILITIES:
- `om interpolate` no longer takes `--output-file` flag.
- `om interpolate` takes ops-files with `-o` instead of `--ops`.
- `om --format=json COMMAND` is no longer supported. This flag should not have
  been global as it is only supported on some commands. The flag is now
  supported on certain commands and needs to be called: `om COMMAND
  --format=json`. The commands that output in `table` format will continue to do
  so by default.

FEATURES:
- `om configure-product` accepts ops-files.
