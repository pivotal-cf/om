&larr; [back to Commands](../README.md)

# `om export-installation`

The `export-installation` command will trigger an archive of the existing installation to be downloaded from the Ops Manager.

## Command Usage
```
ॐ  export-installation
This command will export the current installation of the target Ops Manager.

Usage: om [options] export-installation [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Command Arguments:
  -o, --output-file  string  output path to write installation to
```
