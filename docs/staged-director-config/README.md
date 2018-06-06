&larr; [back to Commands](../README.md)

# `om staged-director-config`

The `staged-director-config` command will export a YAML config file that can be used with `configure-director`.

## Command Usage
```
ॐ  staged-director-config
This command generates a config from a staged director that can be passed in to om configure-director

Usage: om [options] staged-director-config [<args>]
  --client-id, -c            string  Client ID for the Ops Manager VM (not required for unauthenticated commands, $OM_CLIENT_ID)
  --client-secret, -s        string  Client Secret for the Ops Manager VM (not required for unauthenticated commands, $OM_CLIENT_SECRET)
  --connect-timeout, -o      int     timeout in seconds to make TCP connections (default: 5)
  --format, -f               string  Format to print as (options: table,json) (default: table)
  --help, -h                 bool    prints this usage information (default: false)
  --password, -p             string  admin password for the Ops Manager VM (not required for unauthenticated commands, $OM_PASSWORD)
  --request-timeout, -r      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)
  --skip-ssl-validation, -k  bool    skip ssl certificate validation during http requests (default: false)
  --target, -t               string  location of the Ops Manager VM
  --trace, -tr               bool    prints HTTP requests and response payloads
  --username, -u             string  admin username for the Ops Manager VM (not required for unauthenticated commands, $OM_USERNAME)
  --version, -v              bool    prints the om release version (default: false)

Command Arguments:
  --output-file, -o  string  output path to write config to
```
