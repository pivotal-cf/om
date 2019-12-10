# Usage Example

An example of how you might use `om` to install a product can be found [here](EXAMPLE.md).
The example will show some configuration that could apply to installing the Elastic Runtime
on a vSphere Ops Manager.

# Commands

| Command | Description |
| ------------- | ------------- |
| activate-certificate-authority |  activates a certificate authority on the Ops Manager
| [apply-changes](apply-changes/README.md) |  triggers an install on the Ops Manager targeted
| assign-multi-stemcell |  assigns multiple uploaded stemcells to a product in the targeted Ops Manager 2.6+
| assign-stemcell |  assigns an uploaded stemcell to a product in the targeted Ops Manager
| [available-products](available-products/README.md) |  list available products
| [bosh-env](bosh-env/README.md) |  prints bosh environment variables
| certificate-authorities |  lists certificates managed by Ops Manager
| certificate-authority |  prints requested certificate authority
| config-template | **EXPERIMENTAL** generates a config template for the product
| [configure-authentication](configure-authentication/README.md) |  configures Ops Manager with an internal userstore and admin user account
| [configure-director](configure-director/README.md) |  configures the director
| configure-ldap-authentication |  configures Ops Manager with LDAP authentication
| [configure-product](configure-product/README.md) |  configures a staged product
| [configure-saml-authentication](configure-saml-authentication/README.md) |  configures Ops Manager with SAML authentication
| create-certificate-authority |  creates a certificate authority on the Ops Manager
| [create-vm-extension](create-vm-extension/README.md) |  creates a VM extension
| credential-references |  list credential references for a deployed product
| credentials |  fetch credentials for a deployed product
| [curl](curl/README.md) |  issues an authenticated API request
| delete-certificate-authority |  deletes a certificate authority on the Ops Manager
| [delete-installation](delete-installation/README.md) |  deletes all the products on the Ops Manager targeted
| delete-product |  deletes a product from the Ops Manager
| delete-ssl-certificate |  deletes certificate applied to Ops Manager
| [delete-unused-products](delete-unused-products/README.md) |  deletes unused products on the Ops Manager targeted
| [deployed-manifest](deployed-manifest/README.md) |  prints the deployed manifest for a product
| deployed-products |  lists deployed products
| diagnostic-report |  reports current state of your Ops Manager
| disable-director-verifiers |  disables director verifiers
| disable-product-verifiers |  disables product verifiers
| download-product |  downloads a specified product file from Pivotal Network
| errands |  list errands for a product
| expiring-certificates |  lists expiring certificates from the Ops Manager targeted
| [export-installation](export-installation/README.md) |  exports the installation of the target Ops Manager
| generate-certificate |  generates a new certificate signed by Ops Manager's root CA
| generate-certificate-authority |  generates a certificate authority on the Opsman
| [help](help/README.md) |  prints this usage information
| [import-installation](import-installation/README.md) |  imports a given installation to the Ops Manager targeted
| installation-log |  output installation logs
| installations |  list recent installation events
| interpolate |  interpolates variables into a manifest
| pending-changes |  lists pending changes
| pre-deploy-check |  **EXPERIMENTAL** lists pending changes
| product-diff |  **EXPERIMENTAL** displays BOSH manifest diff for a product
| product-metadata |  prints product metadata
| regenerate-certificates |  deletes all non-configurable certificates in Ops Manager so they will automatically be regenerated on the next apply-changes
| revert-staged-changes |  reverts staged changes on the Ops Manager targeted
| [stage-product](stage-product/README.md) |  stages a given product in the Ops Manager targeted
| [staged-config](staged-config/README.md) |  **EXPERIMENTAL** generates a config from a staged product
| [staged-director-config](staged-director-config/README.md) |  **EXPERIMENTAL** generates a config from a staged director
| [staged-manifest](staged-manifest/README.md) |  prints the staged manifest for a product
| staged-products |  lists staged products
| tile-metadata |  **DEPRECATED** prints product metadata. Use product-metadata instead
| unstage-product |  unstages a given product from the Ops Manager targeted
| update-ssl-certificate |  updates the SSL Certificate on the Ops Manager
| [upload-product](upload-product/README.md) |  uploads a given product to the Ops Manager targeted
| [upload-stemcell](upload-stemcell/README.md) |  uploads a given stemcell to the Ops Manager targeted
| [version](version/README.md) |  prints the om release version

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
