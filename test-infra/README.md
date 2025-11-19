# OM Proxy Authentication Test Infrastructure

1. **Run the setup script**:
   ```bash
   cd test-infra
   chmod +x setup.sh
   ./setup.sh
   ```

   This will:
   - Start Kerberos KDC container
   - Start Squid proxy with Kerberos authentication
   - Create test user principal (`testuser@EXAMPLE.COM`)
   - Create proxy service principal
   - Generate keytab for proxy

2. **Verify services are running**:
   ```bash
   docker ps
   ```

   You should see:
   - `om-test-kdc` (Kerberos KDC)
   - `om-test-proxy` (Squid proxy)

## Test Credentials

- **Username**: `testuser@EXAMPLE.COM`
- **Password**: `testpass123`
- **Realm**: `EXAMPLE.COM`
- **Proxy URL**: `http://localhost:3128`

## Testing the Proxy

```bash
# Set krb5.conf location (use .local version for local testing)
export KRB5_CONFIG=$(pwd)/test-infra/krb5/krb5.conf.local

# Run om download-product with proxy authentication
om download-product \
  --pivnet-api-token "your-pivnet-token" \
  --pivnet-product-slug "elastic-runtime" \
  --file-glob "cf-*.pivotal" \
  --product-version "10.3.1" \
  --output-directory /tmp/downloads \
  --proxy-url "http://localhost:3128" \
  --proxy-username "testuser@EXAMPLE.COM" \
  --proxy-password "testpass123" \
  --proxy-domain "EXAMPLE.COM" \
  --proxy-auth-type spnego
```
