# Proxy Authentication Architecture

## Overview

This document describes the extensible proxy authentication system designed to support multiple authentication mechanisms for Pivnet client requests.

## Architecture

### Core Components

1. **ProxyAuth Interface** (`proxy_auth.go`)
   - Defines the contract for all proxy authentication mechanisms
   - Methods: `ConfigureTransport()`, `Type()`, `Validate()`

2. **ProxyAuthConfig** (`proxy_auth.go`)
   - Centralized configuration structure for all proxy parameters
   - Supports: URL, Username, Password, Domain, and explicit Type

3. **ProxyAuthRegistry** (`proxy_auth.go`)
   - Manages available authentication mechanisms
   - Allows registration of new mechanisms at runtime
   - Provides lookup by type

4. **Authentication Mechanisms**
   - **BasicProxyAuth** (`proxy_auth_basic.go`) - HTTP Basic authentication
   - **SPNEGOProxyAuth** (`proxy_auth_spnego.go`) - SPNEGO/Kerberos authentication
   - Future: NTLM, OAuth, etc.

### Design Principles

1. **Extensibility**: New authentication mechanisms can be added by:
   - Implementing the `ProxyAuth` interface
   - Registering with the `ProxyAuthRegistry`

2. **Auto-detection**: Authentication type is automatically determined based on:
   - Explicit `Type` parameter (if provided)
   - Presence of `Domain` → SPNEGO
   - Presence of `Username/Password` → Basic
   - Otherwise → No authentication

3. **Fallback**: If a mechanism fails to initialize, the system falls back to basic auth

4. **Backward Compatibility**: Original `NewPivnetClient` function signature is preserved

## Usage

### Command Line Flags

```bash
# Basic authentication (auto-detected)
om download-product \
  --pivnet-proxy-url http://proxy.example.com:8080 \
  --pivnet-proxy-username user \
  --pivnet-proxy-password pass

# SPNEGO authentication (auto-detected when domain provided)
om download-product \
  --pivnet-proxy-url http://proxy.example.com:8080 \
  --pivnet-proxy-username user \
  --pivnet-proxy-password pass \
  --pivnet-proxy-domain EXAMPLE.COM

# Explicit authentication type
om download-product \
  --pivnet-proxy-url http://proxy.example.com:8080 \
  --pivnet-proxy-username user \
  --pivnet-proxy-password pass \
  --pivnet-proxy-auth-type basic
```

### Programmatic Usage

```go
// Using the new extensible API
proxyConfig := ProxyAuthConfig{
    URL:      "http://proxy.example.com:8080",
    Username: "user",
    Password: "pass",
    Domain:   "EXAMPLE.COM", // Optional, triggers SPNEGO
    Type:     "",            // Optional, auto-detected if empty
}

client := NewPivnetClientWithProxyConfig(
    stdout, stderr, factory, token, skipSSL, host, proxyConfig,
)

// Or using backward-compatible API
client := NewPivnetClient(
    stdout, stderr, factory, token, skipSSL, host,
    proxyURL, proxyUsername, proxyPassword, proxyDomain,
)
```

## Adding New Authentication Mechanisms

To add a new authentication mechanism (e.g., NTLM):

1. **Create the implementation file** (`proxy_auth_ntlm.go`):
```go
type NTLMProxyAuth struct {
    // ... fields
}

func (n *NTLMProxyAuth) Type() ProxyAuthType {
    return ProxyAuthTypeNTLM
}

func (n *NTLMProxyAuth) Validate(config ProxyAuthConfig) error {
    // Validate NTLM-specific requirements
}

func (n *NTLMProxyAuth) ConfigureTransport(transport http.RoundTripper, proxyURL *url.URL) (http.RoundTripper, error) {
    // Configure transport for NTLM
}
```

2. **Add the type constant** to `proxy_auth.go`:
```go
const (
    // ... existing types
    ProxyAuthTypeNTLM ProxyAuthType = "ntlm"
)
```

3. **Register in the registry** (in `NewProxyAuthRegistry()`):
```go
registry.Register(NewNTLMProxyAuth())
```

4. **Update auto-detection logic** in `DetermineAuthType()` if needed

That's it! The new mechanism will be automatically available.

## File Structure

```
download_clients/
├── proxy_auth.go              # Core interface, registry, and configuration
├── proxy_auth_basic.go        # Basic HTTP authentication implementation
├── proxy_auth_spnego.go       # SPNEGO/Kerberos implementation
└── pivnet_client.go           # Pivnet client using proxy auth system
```

## Benefits

1. **Separation of Concerns**: Each mechanism is self-contained
2. **Easy Testing**: Each mechanism can be tested independently
3. **Future-Proof**: New mechanisms don't require changes to existing code
4. **Flexible Configuration**: Supports both explicit and auto-detected auth types
5. **Backward Compatible**: Existing code continues to work


