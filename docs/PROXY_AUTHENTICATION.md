# Proxy Authentication in om-cli download-product

This document provides comprehensive documentation on the proxy authentication mechanism implemented in the `om-cli download-product` command, including both high-level architecture and low-level implementation details.

## Table of Contents

1. [High-Level Overview](#high-level-overview)
2. [Architecture](#architecture)
3. [Integration with om-cli](#integration-with-om-cli)
4. [Low-Level Implementation](#low-level-implementation)
5. [Authentication Mechanisms](#authentication-mechanisms)
6. [Flow Diagrams](#flow-diagrams)
7. [Usage Examples](#usage-examples)
8. [Troubleshooting](#troubleshooting)

---

## High-Level Overview

### What is Proxy Authentication?

Proxy authentication allows the `om-cli download-product` command to authenticate with HTTP/HTTPS proxies when downloading products from Pivotal Network. This is essential in enterprise environments where network traffic must pass through authenticated proxies.

### Supported Authentication Types

The implementation supports two authentication mechanisms:

1. **Basic Authentication** - Simple username/password authentication (RFC 7617)
2. **SPNEGO/Kerberos Authentication** - Enterprise-grade Kerberos-based authentication

### Key Features

- **Extensible Architecture**: Built on a pluggable interface, allowing easy addition of new authentication mechanisms
- **Automatic Configuration**: When proxy flags are not provided, the client operates normally without proxy authentication
- **Kerberos Support**: Full support for Kerberos/SPNEGO with optional custom krb5.conf configuration
- **Transparent Integration**: Works seamlessly with existing download functionality

---

## Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    om-cli download-product                       │
│                         Command Layer                            │
└────────────────────────────┬────────────────────────────────────┘
                              │
                              │ Command Flags:
                              │ --proxy-url
                              │ --proxy-username
                              │ --proxy-password
                              │ --proxy-auth-type
                              │ --proxy-krb5-config
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              download_clients.NewPivnetClient()                 │
│                    Client Creation Layer                         │
│  - Validates proxy flags                                        │
│  - Constructs ClientConfig                                      │
│  - Passes config to go-pivnet library                          │
└────────────────────────────┬────────────────────────────────────┘
                              │
                              │ ClientConfig
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  go-pivnet.NewClient()                          │
│                    Library Layer                                │
│  - Creates HTTP transport                                       │
│  - Configures proxy settings                                    │
│  - Creates ProxyAuthenticator                                   │
│  - Wraps transport with ProxyAuthTransport                     │
└────────────────────────────┬────────────────────────────────────┘
                              │
                              │ HTTP Requests
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              ProxyAuthTransport.RoundTrip()                    │
│                    Transport Layer                             │
│  - Intercepts HTTP requests                                    │
│  - Calls Authenticator.Authenticate()                          │
│  - Adds Proxy-Authorization header                            │
│  - Executes request through underlying transport               │
└────────────────────────────┬────────────────────────────────────┘
                              │
                              │ Authenticated Request
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    HTTP Proxy Server                            │
│              (with authentication enabled)                       │
└─────────────────────────────────────────────────────────────────┘
```

### Interface Design

The architecture is built around the `ProxyAuthenticator` interface:

```go
type ProxyAuthenticator interface {
    // Authenticate adds authentication headers to the HTTP request
    Authenticate(req *http.Request) error
    
    // Close cleans up any resources used by the authenticator
    Close() error
}
```

This interface allows for:
- **Pluggable authentication**: Different mechanisms can be swapped in/out
- **Resource management**: Proper cleanup of Kerberos clients, etc.
- **Testability**: Easy to mock for testing

---

## Integration with om-cli

### Command Flags

The `download-product` command accepts the following proxy-related flags:

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--proxy-url` | string | No | Proxy URL for downloading products from Pivnet (e.g., `http://proxy.example.com:8080`) |
| `--proxy-username` | string | No | Username for proxy authentication |
| `--proxy-password` | string | No | Password for proxy authentication |
| `--proxy-auth-type` | string | No | Type of proxy authentication: `basic` or `spnego` |
| `--proxy-krb5-config` | string | No | Path to Kerberos config file (krb5.conf) for SPNEGO authentication |

### Flag Validation Logic

The implementation uses explicit validation:

1. **No Proxy Configuration**: If `--proxy-url` is not provided, the client is created normally without any proxy settings
2. **Proxy Without Authentication**: If `--proxy-url` is provided but `--proxy-auth-type` is not, the proxy is used without authentication
3. **Proxy With Authentication**: If both `--proxy-url` and `--proxy-auth-type` are provided, authentication is configured

### Code Flow in om-cli

```go
// commands/download_product.go
type PivnetOptions struct {
    ProxyURL        string `long:"proxy-url"`
    ProxyUsername   string `long:"proxy-username"`
    ProxyPassword   string `long:"proxy-password"`
    ProxyAuthType   string `long:"proxy-auth-type"`
    ProxyKrb5Config string `long:"proxy-krb5-config"`
}

// When creating the client:
download_clients.NewPivnetClient(
    stdout, stderr,
    download_clients.DefaultPivnetFactory,
    c.PivnetToken,
    c.PivnetDisableSSL,
    c.PivnetHost,
    c.ProxyURL,           // Passed from command flags
    c.ProxyUsername,      // Passed from command flags
    c.ProxyPassword,      // Passed from command flags
    c.ProxyAuthType,      // Passed from command flags
    c.ProxyKrb5Config,    // Passed from command flags
)
```

---

## Low-Level Implementation

### Client Creation Flow

#### Step 1: Flag Processing (om-cli)

```go
// download_clients/pivnet_client.go
var NewPivnetClient = func(
    stdout *log.Logger,
    stderr *log.Logger,
    factory PivnetFactory,
    token string,
    skipSSL bool,
    pivnetHost string,
    proxyURL string,
    proxyUsername string,
    proxyPassword string,
    proxyAuthType string,
    proxyKrb5Config string,
) ProductDownloader {
    // ... logger setup ...
    
    // Create base config without proxy settings
    config := pivnet.ClientConfig{
        Host:              pivnetHost,
        UserAgent:         userAgent,
        SkipSSLValidation: skipSSL,
    }
    
    // Only configure proxy settings if proxy URL is provided
    if proxyURL != "" {
        config.ProxyURL = proxyURL
        
        // Set proxy authentication if auth type is provided
        if proxyAuthType != "" {
            config.ProxyAuthType = pivnet.ProxyAuthType(proxyAuthType)
            config.ProxyUsername = proxyUsername
            config.ProxyPassword = proxyPassword
            
            // Set Kerberos config file path if provided
            if proxyKrb5Config != "" {
                config.ProxyKrb5Config = proxyKrb5Config
            }
        }
    }
    
    // Create client with config
    client := pivnet.NewClient(tokenGenerator, config, logger)
    // ...
}
```

#### Step 2: Transport Creation (go-pivnet)

```go
// go-pivnet/pivnet.go
func NewClient(token AccessTokenService, config ClientConfig, logger logger.Logger) Client {
    // Create proxy function
    proxyFunc := http.ProxyFromEnvironment
    if config.ProxyURL != "" {
        proxyURL, err := url.Parse(config.ProxyURL)
        if err == nil {
            proxyFunc = http.ProxyURL(proxyURL)
        }
    }
    
    // Create base transport
    baseTransport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: config.SkipSSLValidation,
        },
        Proxy: proxyFunc,
    }
    
    // Wrap with proxy authentication if configured
    var httpTransport http.RoundTripper = baseTransport
    if config.ProxyAuthType != "" {
        authenticator, err := NewProxyAuthenticator(
            config.ProxyAuthType,
            config.ProxyUsername,
            config.ProxyPassword,
            config.ProxyURL,
            config.ProxyKrb5Config,
        )
        if err == nil {
            proxyAuthTransport, err := NewProxyAuthTransport(baseTransport, authenticator)
            if err == nil {
                httpTransport = proxyAuthTransport
            }
        }
    }
    
    httpClient := &http.Client{
        Transport: httpTransport,
    }
    // ...
}
```

#### Step 3: Authenticator Factory

```go
// go-pivnet/proxy_authenticator.go
func NewProxyAuthenticator(
    authType ProxyAuthType,
    username, password, proxyURL, krb5ConfigPath string,
) (ProxyAuthenticator, error) {
    switch authType {
    case ProxyAuthTypeBasic:
        if username == "" || password == "" {
            return nil, fmt.Errorf("username and password are required")
        }
        return NewBasicProxyAuth(username, password), nil
        
    case ProxyAuthTypeSPNEGO:
        return NewSPNEGOProxyAuth(username, password, proxyURL, krb5ConfigPath)
        
    default:
        return nil, fmt.Errorf("unsupported proxy authentication type: %s", authType)
    }
}
```

#### Step 4: Request Interception

```go
// go-pivnet/proxy_auth.go
func (t *ProxyAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    // Add authentication to the request
    if err := t.Authenticator.Authenticate(req); err != nil {
        return nil, fmt.Errorf("failed to authenticate proxy request: %w", err)
    }
    
    // Execute the request with the underlying transport
    return t.Transport.RoundTrip(req)
}
```

---

## Authentication Mechanisms

### Basic Authentication

#### High-Level Flow

1. User provides `--proxy-url`, `--proxy-username`, `--proxy-password`, and `--proxy-auth-type basic`
2. Client creates `BasicProxyAuth` authenticator
3. For each HTTP request:
   - Combines username and password as `username:password`
   - Base64-encodes the credentials
   - Adds header: `Proxy-Authorization: Basic <base64-credentials>`

#### Low-Level Implementation

```go
// go-pivnet/proxy_auth_basic.go
type BasicProxyAuth struct {
    username string
    password string
}

func (b *BasicProxyAuth) Authenticate(req *http.Request) error {
    auth := b.username + ":" + b.password
    encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
    req.Header.Set("Proxy-Authorization", "Basic "+encodedAuth)
    return nil
}
```

#### HTTP Request Example

```
GET /api/v2/products HTTP/1.1
Host: network.pivotal.io
Proxy-Authorization: Basic dXNlcm5hbWU6cGFzc3dvcmQ=
```

### SPNEGO/Kerberos Authentication

#### High-Level Flow

1. User provides `--proxy-url`, `--proxy-username`, `--proxy-password`, `--proxy-auth-type spnego`, and optionally `--proxy-krb5-config`
2. Client creates `SPNEGOProxyAuth` authenticator:
   - Parses proxy URL to extract hostname
   - Derives Kerberos realm from hostname (e.g., `proxy.example.com` → `EXAMPLE.COM`)
   - Loads Kerberos config (from file if provided, or generates default)
   - Authenticates with Kerberos KDC using credentials
3. For each HTTP request:
   - Generates SPNEGO token using Kerberos client
   - Base64-encodes the token
   - Adds header: `Proxy-Authorization: Negotiate <base64-token>`

#### Low-Level Implementation

```go
// go-pivnet/proxy_auth_spnego.go
type SPNEGOProxyAuth struct {
    username       string
    password       string
    proxyURL       string
    kerberosClient *client.Client
}

func NewSPNEGOProxyAuth(username, password, proxyURL, krb5ConfigPath string) (*SPNEGOProxyAuth, error) {
    // Parse proxy URL
    parsedURL, err := url.Parse(proxyURL)
    // Extract domain from hostname
    domain := extractDomain(parsedURL.Hostname())
    
    // Load or create Kerberos config
    var krb5conf *config.Config
    if krb5ConfigPath != "" {
        krb5conf, err = config.Load(krb5ConfigPath)
    } else {
        krb5conf = createDefaultKrb5Config(domain)
    }
    
    // Create Kerberos client
    kerberosClient := client.NewWithPassword(
        username,
        strings.ToUpper(domain),
        password,
        krb5conf,
        client.DisablePAFXFAST(true),
    )
    
    // Login to Kerberos
    err = kerberosClient.Login()
    
    return &SPNEGOProxyAuth{
        kerberosClient: kerberosClient,
        // ...
    }, nil
}

func (s *SPNEGOProxyAuth) Authenticate(req *http.Request) error {
    // Generate SPNEGO token
    spnegoClient := spnego.SPNEGOClient(s.kerberosClient, "HTTP")
    token, err := spnegoClient.InitSecContext()
    
    // Marshal and encode token
    tokenBytes, err := token.Marshal()
    encodedToken := base64.StdEncoding.EncodeToString(tokenBytes)
    
    // Add header
    req.Header.Set("Proxy-Authorization", "Negotiate "+encodedToken)
    return nil
}
```

#### HTTP Request Example

```
GET /api/v2/products HTTP/1.1
Host: network.pivotal.io
Proxy-Authorization: Negotiate YIIBhwYGKwYBBQUCoIIBezCCAXegGDAWBgkqhkiG9w0BBQ...
```

---

## Flow Diagrams

### Complete Request Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. User executes: om download-product --proxy-url ...          │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Command parses flags and creates PivnetOptions              │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. download_clients.NewPivnetClient() called                   │
│    - Validates proxy flags                                      │
│    - Constructs ClientConfig                                    │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. go-pivnet.NewClient() called                                 │
│    - Creates base HTTP transport                                │
│    - Configures proxy function                                  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. NewProxyAuthenticator() called                               │
│    - Creates BasicProxyAuth or SPNEGOProxyAuth                  │
│    - For SPNEGO: authenticates with Kerberos KDC                │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 6. NewProxyAuthTransport() called                               │
│    - Wraps base transport with ProxyAuthTransport              │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 7. HTTP request made                                            │
│    - ProxyAuthTransport.RoundTrip() intercepts                  │
│    - Authenticator.Authenticate() adds header                   │
│    - Request sent through proxy                                 │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 8. Proxy authenticates request and forwards to Pivnet           │
└─────────────────────────────────────────────────────────────────┘
```

### Basic Authentication Flow

```
Request → ProxyAuthTransport.RoundTrip()
              │
              ├─→ BasicProxyAuth.Authenticate()
              │       │
              │       ├─→ Combine username:password
              │       ├─→ Base64 encode
              │       └─→ Set Proxy-Authorization: Basic <encoded>
              │
              └─→ Transport.RoundTrip() → Proxy → Pivnet
```

### SPNEGO Authentication Flow

```
Request → ProxyAuthTransport.RoundTrip()
              │
              ├─→ SPNEGOProxyAuth.Authenticate()
              │       │
              │       ├─→ spnego.SPNEGOClient()
              │       ├─→ InitSecContext() → Generate token
              │       ├─→ Marshal token to bytes
              │       ├─→ Base64 encode
              │       └─→ Set Proxy-Authorization: Negotiate <token>
              │
              └─→ Transport.RoundTrip() → Proxy → Pivnet
```

---

## Usage Examples

### Example 1: Basic Authentication

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

**What happens:**
1. Client connects to `http://proxy.example.com:8080`
2. Each request includes: `Proxy-Authorization: Basic bXl1c2VyOm15cGFzcw==`
3. Proxy authenticates and forwards requests to Pivnet

### Example 2: SPNEGO Authentication (Auto-generated Config)

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

**What happens:**
1. Client extracts domain from proxy URL: `example.com` → `EXAMPLE.COM`
2. Creates default Kerberos config with realm `EXAMPLE.COM`
3. Authenticates with Kerberos KDC
4. Each request includes SPNEGO token in `Proxy-Authorization` header

### Example 3: SPNEGO Authentication (Custom krb5.conf)

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

**What happens:**
1. Client loads Kerberos configuration from `/etc/krb5.conf`
2. Uses realm and KDC settings from the config file
3. Authenticates with Kerberos KDC
4. Each request includes SPNEGO token

### Example 4: No Proxy Authentication

```bash
om download-product \
  --pivnet-product-slug p-healthwatch \
  --pivnet-api-token YOUR_TOKEN \
  --file-glob "*.pivotal" \
  --product-version "2.1.0" \
  --output-directory ./downloads
```

**What happens:**
1. No proxy flags provided
2. Client created normally without proxy configuration
3. Uses system environment variables (`HTTP_PROXY`, `HTTPS_PROXY`) if set
4. No authentication headers added

---

## Troubleshooting

### Common Issues

#### 1. Basic Authentication Fails

**Symptoms:**
- 407 Proxy Authentication Required
- Connection refused

**Solutions:**
- Verify username and password are correct
- Check that proxy supports Basic authentication
- Ensure proxy URL is correct and accessible
- Check network connectivity to proxy

#### 2. SPNEGO Authentication Fails

**Symptoms:**
- Kerberos login errors
- "Failed to login to Kerberos" messages
- 407 Proxy Authentication Required

**Solutions:**
- Verify Kerberos KDC is reachable (typically port 88)
- Check username format (may need `DOMAIN\\user` or `user@DOMAIN.COM`)
- Ensure domain/realm is correctly derived from proxy hostname
- If using custom krb5.conf, verify file path and format
- Check Kerberos configuration matches your environment

#### 3. Proxy URL Not Used

**Symptoms:**
- Requests bypass proxy
- Direct connection to Pivnet

**Solutions:**
- Verify `--proxy-url` flag is correctly formatted
- Check that proxy URL includes protocol (`http://` or `https://`)
- Ensure proxy is accessible from your network

#### 4. Kerberos Config File Not Found

**Symptoms:**
- "failed to load Kerberos config" error

**Solutions:**
- Verify file path is correct and absolute
- Check file permissions (must be readable)
- Ensure file is valid krb5.conf format

### Debugging

Enable verbose logging to see detailed proxy authentication flow:

```bash
om download-product \
  --pivnet-product-slug ... \
  --proxy-url ... \
  --proxy-auth-type spnego \
  # ... other flags
```

Check logs for:
- Proxy configuration details
- Authentication header generation
- Kerberos authentication steps
- HTTP request/response details

### Environment Variables

The implementation respects standard proxy environment variables when proxy flags are not provided:

- `HTTP_PROXY` / `http_proxy`
- `HTTPS_PROXY` / `https_proxy`
- `NO_PROXY` / `no_proxy`

---

## Security Considerations

### Basic Authentication

- **Credentials are base64-encoded, not encrypted**
- Always use HTTPS/TLS when possible
- Consider using SPNEGO in enterprise environments

### SPNEGO Authentication

- Requires proper Kerberos infrastructure
- KDC must be accessible and properly configured
- Credentials are used for Kerberos authentication, not sent in plain text

### General Recommendations

1. Use HTTPS for proxy connections when possible
2. Store credentials securely (environment variables, credential managers)
3. Use SPNEGO in enterprise environments with Active Directory
4. Regularly rotate proxy credentials
5. Monitor proxy authentication logs for suspicious activity

---

## References

- [go-pivnet Proxy Authentication Architecture](../go-pivnet/PROXY_AUTH_ARCHITECTURE.md)
- [go-pivnet Proxy Authentication Examples](../go-pivnet/PROXY_AUTH_EXAMPLE.md)
- [RFC 7617: HTTP Basic Authentication](https://tools.ietf.org/html/rfc7617)
- [RFC 4559: SPNEGO for HTTP](https://tools.ietf.org/html/rfc4559)

