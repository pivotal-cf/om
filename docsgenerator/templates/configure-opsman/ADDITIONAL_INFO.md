<!--- Anything in this file will be appended to the final docs/configure-opsman/README.md file --->
# Creating a Config File
Settings that can be set using this command include:

- SSL Certificate
- Pivotal Network Settings (pending)
- Custom Banner (pending)
- Syslog (pending)
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
opsman-configuration:
  aws:
    ...
```

Note that this config support the `opsman-configuration` top level key.
This allows for compatibility with the [Platform Automation Toolkit](https://docs.pivotal.io/platform-automation) product.
