&larr; [back to Commands](../README.md)

# `om configure-saml-authentication`

The `configure-saml-authentication` command will allow you to setup your user account on the Ops Manager with SAML authentication.

To set up your Ops Manager with internal authentication instead, use `configure-authentication`.

## Command Usage
```
ॐ  configure-saml-authentication
This unauthenticated command helps setup the authentication mechanism for your Ops Manager with SAML.

Usage: om [options] configure-saml-authentication [<args>]
  --client-id, -c, OM_CLIENT_ID          string  Client ID for the Ops Manager VM (not required for unauthenticated commands)
  --client-secret, -s, OM_CLIENT_SECRET  string  Client Secret for the Ops Manager VM (not required for unauthenticated commands)
  --connect-timeout, -o                  int     timeout in seconds to make TCP connections (default: 5)
  --format, -f                           string  Format to print as (options: table,json) (default: table)
  --help, -h                             bool    prints this usage information (default: false)
  --password, -p, OM_PASSWORD            string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  --request-timeout, -r                  int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)
  --skip-ssl-validation, -k              bool    skip ssl certificate validation during http requests (default: false)
  --target, -t, OM_TARGET                string  location of the Ops Manager VM
  --trace, -tr                           bool    prints HTTP requests and response payloads
  --username, -u, OM_USERNAME            string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  --version, -v                          bool    prints the om release version (default: false)

Command Arguments:
  --decryption-passphrase, -dp  string (required)  passphrase used to encrypt the installation
  --http-proxy-url              string             proxy for outbound HTTP network traffic
  --https-proxy-url             string             proxy for outbound HTTPS network traffic
  --no-proxy                    string             comma-separated list of hosts that do not go through the proxy
  --saml-bosh-idp-metadata      string (required)  XML, or URL to XML, for the IDP that BOSH should use
  --saml-idp-metadata           string (required)  XML, or URL to XML, for the IDP that Ops Manager should use
  --saml-rbac-admin-group       string (required)  If SAML is specified, please provide the admin group for your SAML
  --saml-rbac-groups-attribute  string (required)  If SAML is specified, please provide the groups attribute for your SAML
```

The `--saml-idp-metadata` and `--saml-bosh-idp-metadata` can be the same.
