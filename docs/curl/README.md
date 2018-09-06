&larr; [back to Commands](../README.md)

# `om curl`

The `curl` command will help you make arbitrary API calls against the Ops Manager VM.

## Command Usage
```
ॐ  curl
This command issues an authenticated API request as defined in the arguments

Usage: om [options] curl [<args>]
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
  --data, -d     string             api request payload
  --header, -H   string (variadic)  used to specify custom headers with your command (default: Content-Type: application/json)
  --path, -p     string (required)  path to api endpoint
  --request, -x  string             http verb (default: GET)
  --silent, -s   bool               only write response headers to stderr if response status is 4XX or 5XX
```
