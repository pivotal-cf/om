<!--- This file is autogenerated from the files in docsgenerator/templates/config-template --->
&larr; [back to Commands](../README.md)

# `om config-template`

this command generates a product configuration template from a .pivotal file on
Pivnet

## Command Usage
```
Usage:
  om [OPTIONS] config-template [config-template-OPTIONS]

this command generates a product configuration template from a .pivotal file on
Pivnet

Application Options:
      --ca-cert=                 OpsManager CA certificate path or value
                                 [$OM_CA_CERT]
  -c, --client-id=               Client ID for the Ops Manager VM (not required
                                 for unauthenticated commands) [$OM_CLIENT_ID]
  -s, --client-secret=           Client Secret for the Ops Manager VM (not
                                 required for unauthenticated commands)
                                 [$OM_CLIENT_SECRET]
  -o, --connect-timeout=         timeout in seconds to make TCP connections
                                 (default: 10) [$OM_CONNECT_TIMEOUT]
  -d, --decryption-passphrase=   Passphrase to decrypt the installation if the
                                 Ops Manager VM has been rebooted (optional for
                                 most commands) [$OM_DECRYPTION_PASSPHRASE]
  -e, --env=                     env file with login credentials
  -p, --password=                admin password for the Ops Manager VM (not
                                 required for unauthenticated commands)
                                 [$OM_PASSWORD]
  -r, --request-timeout=         timeout in seconds for HTTP requests to Ops
                                 Manager (default: 1800) [$OM_REQUEST_TIMEOUT]
  -k, --skip-ssl-validation      skip ssl certificate validation during http
                                 requests [$OM_SKIP_SSL_VALIDATION]
  -t, --target=                  location of the Ops Manager VM [$OM_TARGET]
      --trace                    prints HTTP requests and response payloads
                                 [$OM_TRACE]
  -u, --username=                admin username for the Ops Manager VM (not
                                 required for unauthenticated commands)
                                 [$OM_USERNAME]
      --vars-env=                load vars from environment variables by
                                 specifying a prefix (e.g.: 'MY' to load
                                 MY_var=value) [$OM_VARS_ENV]
  -v, --version                  prints the om release version

Help Options:
  -h, --help                     Show this help message

[config-template command options]
          --pivnet-api-token=
          --pivnet-product-slug= the product name in pivnet
          --product-version=     the version of the product from which to
                                 generate a template
          --pivnet-host=         the API endpoint for Pivotal Network (default:
                                 https://network.pivotal.io)
      -f, --file-glob=           a glob to match exactly one file in the pivnet
                                 product slug (default: *.pivotal)
          --pivnet-disable-ssl   whether to disable ssl validation when
                                 contacting the Pivotal Network
          --product-path=        path to product file
          --output-directory=    a directory to create templates under. must
                                 already exist.
          --exclude-version      if set, will not output a version-specific
                                 directory
          --size-of-collections=

    config file interpolation:
      -c, --config=              path to yml file for configuration (keys must
                                 match the following command line flags)
          --vars-env=            load variables from environment variables
                                 matching the provided prefix (e.g.: 'MY' to
                                 load MY_var=value) [$OM_VARS_ENV]
      -l, --vars-file=           load variables from a YAML file
      -v, --var=                 load variable from the command line. Format:
                                 VAR=VAL
```

