# Usage Example

An example of how you might use `om` to install a product can be found [here](EXAMPLE.md).
The example will show some configuration that could apply to installing the Elastic Runtime
on a vSphere Ops Manager.

# Commands
* [apply-changes](apply-changes/README.md)
* [available-products](available-products/README.md)
* [configure-authentication](configure-authentication/README.md)
* [configure-bosh](configure-bosh/README.md)
* [configure-product](configure-product/README.md)
* [curl](curl/README.md)
* [delete-installation](delete-installation/README.md)
* [delete-unused-products](delete-unused-products/README.md)
* [export-installation](export-installation/README.md)
* [help](help/README.md)
* [import-installation](import-installation/README.md)
* [stage-product](stage-product/README.md)
* [upload-product](upload-product/README.md)
* [upload-stemcell](upload-stemcell/README.md)
* [version](version/README.md)

# Authentication
OM will by preference use Client ID and Client Secret if provided. To create a Client ID and Client Secret

1. `uaac target https://YOUR_OPSMANAGER/uaa`
1. `uaac token sso get` if using SAML or `uaac token owner get` if using internal auth. Specify the Client ID as `opsman` and leave Client Secret blank.
1. Generate a client ID and secret

```
uaac client add -i
Client ID:  NEW_CLIENT_NAME
New client secret:  DESIRED_PASSWORD
Verify new client secret:  DESIRED_PASSWORD
scope (list):  opsman.admin
authorized grant types (list):  client_credentials
authorities (list):  opsman.admin
access token validity (seconds):  43200
refresh token validity (seconds):  43200
redirect uri (list):
autoapprove (list):
signup redirect url (url):
```
