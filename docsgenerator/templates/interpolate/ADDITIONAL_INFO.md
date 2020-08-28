## Interpolation
Given a template file with a variable reference:

```yaml
# config.yml
key: ((variable_name))
```

Values can be provided from a separate variables yaml file (`--vars-file`)
or from environment variables (`--vars-env`).

To load variables from a file use the `--vars-file` flag.

```yaml
# vars.yml
variable_name: some_value
```

```
om interpolate \
  --config config.yml \
  --vars-file vars.yml
```

To load variables from a set of environment variables, specify the common
environment variable prefix with the `--vars-env` flag.

```
OM_VAR_variable_name=some_value om interpolate \
  --config config.yml \
  --vars-env OM_VAR
```

The interpolation support is inspired by similar features in BOSH. You can
[refer to the BOSH documentation](https://bosh.io/docs/cli-int/) for details on how interpolation
is performed.

One significant difference from similar features in BOSH
and Concourse `fly` is that `om` performs _dual-pass_ interpolation.
That is, the template is interpolated (including applying ops files) once,
and then the output of that interpolation is interpolated again,
using the same arguments (except for Ops Files,
which are not necessarily idempotent).

This allows the use of _mapping variables_,
variables that contain a value that is a variable in turn.
Mapping variables are useful for mapping
between programmatically-generated variable names
such as those created by `om config-template` and `om staged-director-config`,
and credentials that may be used in multiple places,
such as database passwords.

Such config might look something like this:

```
cc:
  database-password: ((cc_database_password))
uaa:
  database-password: ((uaa_database_password))
```

In such cases, a vars-file like this can encode the relationship:

```
---
cc_database_password: ((sql_password))
uaa_database_password: ((sql_password))
```

Assuming the value of `sql_password` is available to the interpolation,
it will be present in the final output,
like so:

```
cc:
  database-password: actualsqlpasswordverysecure
uaa:
  database-password: actualsqlpasswordverysecure
```

This feature generally shouldn't interfere with interpolation
as it would normally work in BOSH.
If you encounter any situation where this difference is an unwelcome surprise,
please open an issue; we were unable to think of any.
