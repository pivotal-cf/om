# Proxy Authentication Quick Reference

Quick reference guide for using proxy authentication with `om download-product`.

## Command Flags

| Flag | Required | Description | Example |
|------|----------|-------------|---------|
| `--proxy-url` | No | Proxy URL | `http://proxy.example.com:8080` |
| `--proxy-username` | No* | Username for authentication | `myuser` |
| `--proxy-password` | No* | Password for authentication | `mypass` |
| `--proxy-auth-type` | No* | Authentication type: `basic` or `spnego` | `basic` |
| `--proxy-krb5-config` | No | Path to Kerberos config file (for SPNEGO) | `/etc/krb5.conf` |

*Required when `--proxy-url` is provided and authentication is needed

## Quick Examples

### Basic Authentication

```bash
om download-product \
  --pivnet-product-slug p-healthwatch \
  --pivnet-api-token YOUR_TOKEN \
  --file-glob "*.pivotal" \
  --product-version "2.1.0" \
  --output-directory ./downloads \
  --proxy-url http://proxy.example.com:8080 \
  --proxy-username myuser \
  --proxy-password mypass \
  --proxy-auth-type basic
```

### SPNEGO Authentication (Auto Config)

```bash
om download-product \
  --pivnet-product-slug p-healthwatch \
  --pivnet-api-token YOUR_TOKEN \
  --file-glob "*.pivotal" \
  --product-version "2.1.0" \
  --output-directory ./downloads \
  --proxy-url http://proxy.example.com:8080 \
  --proxy-username DOMAIN\\user \
  --proxy-password password \
  --proxy-auth-type spnego
```

### SPNEGO Authentication (Custom krb5.conf)

```bash
om download-product \
  --pivnet-product-slug p-healthwatch \
  --pivnet-api-token YOUR_TOKEN \
  --file-glob "*.pivotal" \
  --product-version "2.1.0" \
  --output-directory ./downloads \
  --proxy-url http://proxy.example.com:8080 \
  --proxy-username user@EXAMPLE.COM \
  --proxy-password password \
  --proxy-auth-type spnego \
  --proxy-krb5-config /etc/krb5.conf
```

### No Proxy (Normal Operation)

```bash
om download-product \
  --pivnet-product-slug p-healthwatch \
  --pivnet-api-token YOUR_TOKEN \
  --file-glob "*.pivotal" \
  --product-version "2.1.0" \
  --output-directory ./downloads
```

## Behavior Matrix

| `--proxy-url` | `--proxy-auth-type` | Behavior |
|---------------|---------------------|----------|
| Not provided | - | Normal operation, uses environment proxy if set |
| Provided | Not provided | Uses proxy without authentication |
| Provided | `basic` | Uses proxy with Basic authentication |
| Provided | `spnego` | Uses proxy with SPNEGO/Kerberos authentication |

## Troubleshooting Quick Tips

### Basic Auth Issues
- Verify username/password
- Check proxy supports Basic auth
- Ensure proxy URL is correct

### SPNEGO Issues
- Verify Kerberos KDC is reachable (port 88)
- Check username format (`DOMAIN\\user` or `user@DOMAIN.COM`)
- Ensure krb5.conf path is correct (if using custom config)

### General
- Check network connectivity to proxy
- Verify proxy URL format includes protocol
- Review logs for detailed error messages

## See Also

For detailed documentation, see [PROXY_AUTHENTICATION.md](./PROXY_AUTHENTICATION.md)

