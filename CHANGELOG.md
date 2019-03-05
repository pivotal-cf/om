## 0.54.0 (unreleased)

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


### Bug Fixes
* when trying to delete a product on Ops Manager during a selective deploy (`apply-changes --product-name tile`),
  OpsManager would fail to `apply-changes` due to a change to the version string for 2.5 (would include the build
  number). A change was made to the info service to accept the new semver formatting as well as the old 
  versioning. 
* upload-product (among other things) is no longer sensitive to subdirectories in tile metadata directories
* to support 2.5, the Redis Team (jplebre and edwardecook) submitted a PR to support new semver versioning for 
  OpsManager in addition to supporting the current versioning format.


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
