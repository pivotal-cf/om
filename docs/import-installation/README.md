&larr; [back to Commands](../README.md)

# `om import-installation`

The `import-installation` command will upload an existing installation archive to the Ops Manager.
This is helpful when upgrading the Ops Manager itself.
You can download an archive from the Ops Manager by using the [`export-installation` command](../export-installation/README.md).

## Command Usage
```
ॐ  import-installation
This unauthenticated command attempts to import an installation to the Ops Manager targeted.

Usage: om [options] import-installation [<args>]
  --client-id, -c, OM_CLIENT_ID          string  Client ID for the Ops Manager VM (not required for unauthenticated commands)
  --client-secret, -s, OM_CLIENT_SECRET  string  Client Secret for the Ops Manager VM (not required for unauthenticated commands)
  --connect-timeout, -o                  int     timeout in seconds to make TCP connections (default: 5)
  --env, -e                              string  env file with login credentials
  --help, -h                             bool    prints this usage information (default: false)
  --password, -p, OM_PASSWORD            string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  --request-timeout, -r                  int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)
  --skip-ssl-validation, -k              bool    skip ssl certificate validation during http requests (default: false)
  --target, -t, OM_TARGET                string  location of the Ops Manager VM
  --trace, -tr                           bool    prints HTTP requests and response payloads
  --username, -u, OM_USERNAME            string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  --version, -v                          bool    prints the om release version (default: false)

Command Arguments:
  --decryption-passphrase, -dp  string (required)  passphrase for Ops Manager to decrypt the installation
  --installation, -i            string (required)  path to installation.
  --polling-interval, -pi       int                interval (in seconds) at which to print status (default: 1)
```
