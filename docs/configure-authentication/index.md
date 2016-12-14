# `configure-authentication`

```
ॐ  configure-authentication
This unauthenticated command helps setup the authentication mechanism for your Ops Manager.
The "internal" userstore mechanism is the only currently supported option.

Usage: om [options] configure-authentication [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Command Arguments:
  -u, --username                string  admin username
  -p, --password                string  admin password
  -dp, --decryption-passphrase  string  passphrase used to encrypt the installation
```
