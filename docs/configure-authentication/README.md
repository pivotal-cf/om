&larr; [back to Commands](../README.md)

# `om configure-authentication`

The `configure-authentication` command will allow you to setup your user account on the Ops Manager.

## Supported Authentication Methods
##### Using Internal Authentication)
* Internal
* SAML

## Command Usage
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
  --http-proxy-url              string  proxy for outbound HTTP network traffic
  --https-proxy-url             string  proxy for outbound HTTPS network traffic
  --no-proxy                    string  comma-separated list of hosts that do not go through the proxy
  
Command Arguments:
  --username, -u                string             Internal Authentication: admin username
  --password, -p                string             Internal Authentication: admin password
  --decryption-passphrase, -dp  string (required)  passphrase used to encrypt the installation
  --saml-idp-metadata           string             SAML Authentication: XML, or URL to XML, for the IDP that Ops Manager should use
  --saml-bosh-idp-metadata      string             SAML Authentication: XML, or URL to XML, for the IDP that BOSH should use
  --saml-rbac-admin-group       string             SAML Authentication: If SAML is specified, please provide the admin group for your SAML
  --saml-rbac-groups-attribute  string             SAML Authentication: If SAML is specified, please provide the groups attribute for your SAML
  --http-proxy-url              string             proxy for outbound HTTP network traffic
  --https-proxy-url             string             proxy for outbound HTTPS network traffic
  --no-proxy                    string             comma-separated list of hosts that do not go through the proxy
```

## Using Internal Authentication
This method requires `--username`, `--password` and `--decryption-passphrase` to be set.

## Using SAML Authentication
This method requires `--decryption-passphrase`, `--saml-idp-metadata`, `--saml-bosh-idp-metadata`,
 `--saml-rbac-admin-group`, and `--saml-rbac-groups-attribute` to be set.

The `--saml-idp-metadata` and `--saml-bosh-idp-metadata` can be the same.
