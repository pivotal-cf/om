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
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Command Arguments:
  -i, --installation            string  path to installation.
  -dp, --decryption-passphrase  string  passphrase for Ops Manager to decrypt the installation
```
