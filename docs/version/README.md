&larr; [back to Commands](../README.md)

# `om version`

The `version` command prints the version of `om` that is installed.
You can find new versions of `om` on the [Releases](https://github.com/pivotal-cf/om/releases) page.

## Command Usage
```
ॐ  version
This command prints the om release version number.

Usage: om [options] version
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)
```
