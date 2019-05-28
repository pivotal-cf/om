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
  
Changes internal to `om` will _**NOT**_ be included as a part of the om API.
The `om` team reserves the right to change any internal structs or structures
as long as the outputs and behavior of the commands remain the same.

**NOTE**: Additional documentation for om commands 
leveraged by Pivotal Platform Automation 
can be found in [Pivotal Documentation](docs.pivotal.io/platform-automation.)
`om` is versioned independently from platform-automation. 


## 1.1.1 (unreleased)

(( Placeholder ))

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
  Thanks to @vchrisb for the PR on issue [#360](https://github.com/pivotal-cf/om/issues/360)

### Bug Fixes
* Extra values passed in the env file 
  will now fail if they are not recognized properties.
  This closes issue [#258](https://github.com/pivotal-cf/om/issues/258)
* `om` will now allow non-string entities
  to be passed as strings to Ops Manager.
  This closes issue [#352](https://github.com/pivotal-cf/om/issues/352)
* the file outputted by `download-product`
  will now use the `product-name` as defined 
  in the downloaded-product, 
  _not_ from the Pivnet slug.
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
  ([PR #338](https://github.com/pivotal-cf/om/pull/338) Thanks @Logiraptor!)
  This is an _advanced_ feature
  that should only be used if the user knows how to disable verifiers in OpsManager.
  This flag will only disable verifiers for configure-director,
  and will not disable the warnings for apply-changes.
* There's now a shell-completion script;
  see the readme for details.
* We have totally replaced the code and behavior
  of the _experimental_ `config-template` command.
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
* to support 2.5, @jplebre and @edwardecook submitted a PR to support new semver versioning for 
  OpsManager in addition to supporting the current versioning format.
  
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
* `config-template` & `staged-config` & `staged-director-config`: pluralize `--include-placeholders` flag
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
