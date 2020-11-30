<!--- Anything in this file will be appended to the final docs/configure-opsman/README.md file --->
# Creating a Config File
Settings that can be set using this command include:

- SSL Certificate
- Pivotal Network Settings (pending)
- Custom Banner (pending)
- Syslog (pending)
- UAA tokens expiration
- Role Based Access Control (if enabled) (pending)

An example config file for updating settings 
on the Ops Manager Settings page (will update as more functionality is added):

```yaml
ssl-certificate:
  certificate: |
    -----BEGIN CERTIFICATE-----
    certificate
    -----END CERTIFICATE-----
  private_key:
    ----BEGIN RSA PRIVATE KEY-----
    private-key
    -----END RSA PRIVATE KEY-----
pivotal-network-settings:
  api_token: your-pivnet-token
banner-settings:
  ui_banner_contents: UI Banner Contents
  ssh_banner_contents: SSH Banner Contents
syslog-settings:
  enabled: true
  address: 1.2.3.4
  port: 999
  transport_protocol: tcp
  tls_enabled: true
  permitted_peer: "*.example.com"
  ssl_ca_certificate: |
    -----BEGIN CERTIFICATE-----
    certificate
    -----END CERTIFICATE-----
  queue_size: 100000
  forward_debug_logs: false
  custom_rsyslog_configuration: if $message contains 'test' then stop
tokens-expiration:
  access_token_expiration: 100
  refresh_token_expiration: 1200
  session_idle_timeout: 50
rbac-settings: # if your RBAC is SAML, use these settings
  rbac_saml_admin_group: example_group_name
  rbac_saml_groups_attribute: example_attribute_name
#rbac-settings: # if your RBAC is LDAP, replace the above
#  ldap_rbac_admin_group_name: cn=opsmgradmins,ou=groups,dc=mycompany,dc=com
opsman-configuration:
  aws:
    ...
```

Note that this config support the `opsman-configuration` top level key.
This allows for compatibility with the [Platform Automation Toolkit](https://docs.pivotal.io/platform-automation) product.
