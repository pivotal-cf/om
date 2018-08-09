# Kiln

_Kiln bakes tiles_

Kiln helps tile developers build products for Pivotal Operations Manager. It provides
an opinionated folder structure and templating capabilities. It is designed to be used
both in CI environments and in command-line to produce a tile.

## Subcommands

### `bake`

It takes release and stemcell tarballs, metadata YAML, and JavaScript migrations
as inputs and produces an OpsMan-compatible tile as its output.

Here is an example command line:
```
$ kiln bake \
    --version 2.0.0 \
    --metadata /path/to/metadata.yml \
    --releases-directory /path/to/releases \
    --stemcell-tarball /path/to/stemcell.tgz \
    --migrations-directory /path/to/migrations \
    --output-file /path/to/cf-2.0.0-build.4.pivotal
```

Refer to the [example-tile](example-tile) for a complete example showing the
different features kiln supports.

#### Options

##### `--bosh-variables-directory`

The `--bosh-variables-directory` flag can be used to include CredHub variable
declarations. You should prefer the use of variables rather than Ops Manager
secrets. Each `.yml` file in the directory should define a top-level `variables`
key.

This flag can be specified multiple times if you have organized your
variables into subdirectories for development convenience.

Example [variables](example-tile/bosh-variables) directory.

##### `--embed`

The `--embed` flag is for embedding any extra files or directories into the
tile. There are very few reasons a tile developer should want to do this, but if
you do, you can include these extra files here. The flag can be specified
multiple times to embed multiple files or directories.

##### `--forms-directory`

The `--forms-directory` flag takes a path to a directory that contains one
or more forms. The flag can be specified more than once.

To reference a form file in the directory you can use the `form`
template helper:

```
$ cat /path/to/metadata
---
form_types:
- $( form "first" )
```

Example [forms](example-tile/forms) directory.

##### `--icon`

The `--icon` flag takes a path to an icon file.

To include the base64'd representation of the icon you can use the `icon`
template helper:

```
$ cat /path/to/metadata
---
icon_image: $( icon )
```

##### `--instance-groups-directory`

The `--instance-groups-directory` flag takes a path to a directory that contains one
or more instance groups. The flag can be specified more than once.

To reference an instance group in the directory you can use the `instance_group`
template helper:

```
$ cat /path/to/metadata
---
job_types:
- $( instance_group "my-instance-group" )
```

Example [instance-groups](example-tile/instance-groups) directory.

##### `--jobs-directory`

The `--jobs-directory` flag takes a path to a directory that contains one
or more jobs. The flag can be specified more than once.

To reference a job file in the directory you can use the `job`
template helper:

```
$ cat /path/to/instance-group
---
templates:
- $( job "my-job" )
- $( job "my-aliased-job" )
- $( job "my-errand" )
```

Example [jobs](example-tile/jobs) directory.

You may find that you want to define different job files for the same BOSH job
with different properties. To do this you add an `alias` key to the job which
will be used in preference to the job name when resolving job references:

```
$ cat /path/to/my-aliased-job
---
name: my-job
alias: my-aliased-job
```

##### `--metadata`

Specify a file path to a tile metadata file for the `--metadata` flag. This
metadata file will contain the contents of your tile configuration as specified
in the OpsManager tile development documentation.

##### `--migrations-directory`

If your tile has JavaScript migrations, then you will need to include the
`--migrations-directory` flag. This flag can be specified multiple times if you
have organized your migrations into subdirectories for development convenience.

##### `--output-file`

The `--output-file` flag takes a path to the location on the filesystem where
your tile will be created. The flag expects a full file name like
`tiles/my-tile-1.2.3-build.4.pivotal`.

##### `--properties-directory`

The `--properties-directory` flag takes a path to a directory that contains one
or more blueprint property files. The flag can be specified more than once.

To reference a properties file in the directory you can use the `property`
template helper:

```
$ cat /path/to/metadata
---
property_blueprints:
- $( property "rep_password" )
```

Example [properties](example-tile/properties) directory.

##### `--releases-directory`

The `--releases-directory` flag takes a path to a directory that contains one or
many release tarballs. The flag can be specified more than once. This is
useful if you consume bosh releases as Concourse resources. Each release will
likely show up in the task as a separate directory. For example:

```
$ tree /path/to/releases
|-- first
|   |-- cflinuxfs2-release-1.166.0.tgz
|   `-- consul-release-190.tgz
`-- second
    `-- nats-release-22.tgz
```

To reference a release you can use the `release` template helper:

```
$ cat /path/to/metadata
---
releases:
- $( release "cflinuxfs2" )
- $( release "consul" )
- $( release "nats" )
```

Example kiln command line:

```
$ kiln bake \
    --version 2.0.0 \
    --metadata /path/to/metadata.yml \
    --releases-directory /path/to/releases/first \
    --releases-directory /path/to/releases/second \
    --stemcell-tarball /path/to/stemcell.tgz \
    --output-file /path/to/cf-2.0.0-build.4.pivotal
```

##### `--runtime-configs-directory`

The `--runtime-configs-directory` flag takes a path to a directory that
contains one or more runtime config files. The flag can be specified
more than once.

To reference a runtime config in the directory you can use the `runtime_config`
template helper:

```
$ cat /path/to/metadata
---
runtime_configs:
- $( runtime_config "first-runtime-config" )
```

Example [runtime-configs](example-tile/runtime-configs) directory.

##### `--stemcell-tarball`

The `--stemcell-tarball` flag takes a path to a stemcell.

To include information about the stemcell in your metadata you can use the
`stemcell` template helper:

```
$ cat /path/to/metadata
---
stemcell_criteria: $( stemcell )
```

##### `--stub-releases`

For tile developers looking to get some quick feedback about their tile
metadata, the `--stub-releases` flag will skip including the release tarballs
into the built tile output. This should result in a much smaller file that
should upload much more quickly to OpsManager.

##### `--variable`

The `--variable` flag takes a `key=value` argument that allows you to specify
arbitrary variables for use in your metadata. The flag can be specified
more than once.

To reference a variable you can use the `variable` template helper:

```
$ cat /path/to/metadata
---
$( variable "some-variable" )
```

##### `--variables-file`

The `--variables-file` flag takes a path to a YAML file that contains arbitrary
variables for use in your metadata. The flag can be specified more than once.

To reference a variable you can use the `variable` template helper:

```
$ cat /path/to/metadata
---
$( variable "some-variable" )
```

Example [variables file](example-tile/variables.yml).

##### `--version`

The `--version` flag takes the version number you want your tile to become.

To reference the version you use the `version` template helper:

```
$ cat /path/to/metadata
---
product_version: $( version )
provides_product_versions:
- name: example
  version: $( version )
```

### Template functions

#### `select`

The `select` function allows you to pluck values for nested fields from a
template helper.

For instance, this section in our example tile:

```
my_release_version: $( release "my-release" | select "version" 
```

Results in:

```
my_release_version: 1.2.3
```
