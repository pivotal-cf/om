# Basic Authentication Flow - Complete Sequence Diagram

This document provides a comprehensive sequence diagram demonstrating the complete Basic proxy authentication flow, touching all components from user input through om-cli, go-pivnet library, and external services.

## Architecture Overview

The proxy authentication solution involves changes in **both om-cli and go-pivnet library**:

### om-cli Changes
- **Command Layer** (`commands/download_product.go`): Added command-line flags for proxy authentication parameters
- **Adapter Layer** (`download_clients/pivnet_client.go`): Converts flags to `ClientConfig` and passes to go-pivnet

### go-pivnet Library Changes
- **Client Factory** (`pivnet.go`): Extended `ClientConfig` and conditionally wraps transport with authentication
- **Authentication Components**: `proxy_authenticator.go`, `proxy_auth_basic.go`, `proxy_auth.go` implement Basic authentication

### Interface Boundary
The interface between om-cli and go-pivnet is the `ClientConfig` structure, which is populated by om-cli and consumed by go-pivnet.

## Overview

The Basic authentication flow involves multiple layers:
1. **om-cli Layer**: Command parsing, flag validation, and client creation
2. **go-pivnet Library Layer**: Client initialization, authenticator creation, and transport setup
3. **External Services**: HTTP Proxy for request forwarding

## Complete Basic Authentication Flow

### Mermaid Sequence Diagram (Simplified)

```mermaid
sequenceDiagram
    participant User
    participant CMD as download_product.go
    participant ADAPTER as pivnet_client.go
    participant PIVNET as pivnet.go
    participant AUTH_FACTORY as proxy_authenticator.go
    participant BASIC_AUTH as BasicProxyAuth
    participant TRANSPORT as ProxyAuthTransport
    participant PROXY as HTTP Proxy
    participant API as Pivnet API

    rect rgb(225, 245, 255)
        Note over CMD,ADAPTER: om-cli Layer
    end
    
    rect rgb(255, 244, 225)
        Note over PIVNET,TRANSPORT: go-pivnet Library
    end

    Note over User,API: Phase 1: Client Initialization

    User->>CMD: om download-product<br/>--proxy-auth-type=basic<br/>--proxy-url=...<br/>--proxy-username=...<br/>--proxy-password=...
    CMD->>CMD: Parse & validate flags
    Note over CMD: Create PivnetOptions<br/>with proxy settings<br/>[om-cli]
    CMD->>ADAPTER: NewPivnetClient(flags)
    ADAPTER->>ADAPTER: Build ClientConfig<br/>with proxy settings
    Note over ADAPTER: Only set proxy if URL provided<br/>[om-cli]
    ADAPTER->>PIVNET: NewClient(config)
    PIVNET->>PIVNET: Setup HTTP transport<br/>with proxy configuration
    Note over PIVNET: [go-pivnet]
    PIVNET->>AUTH_FACTORY: NewProxyAuthenticator(basic, ...)
    Note over AUTH_FACTORY: [go-pivnet]
    AUTH_FACTORY->>BASIC_AUTH: NewBasicProxyAuth(username, password)
    Note over BASIC_AUTH: [go-pivnet]
    BASIC_AUTH->>BASIC_AUTH: Store username & password
    BASIC_AUTH-->>AUTH_FACTORY: Return BasicProxyAuth
    AUTH_FACTORY-->>PIVNET: Return authenticator
    PIVNET->>TRANSPORT: NewProxyAuthTransport<br/>(transport, authenticator)
    Note over TRANSPORT: [go-pivnet]
    TRANSPORT-->>PIVNET: Return ProxyAuthTransport
    PIVNET->>PIVNET: Create HTTP client<br/>with authenticated transport
    PIVNET-->>ADAPTER: Return Client
    ADAPTER-->>CMD: Return ProductDownloader
    CMD-->>User: Client ready

    Note over User,API: Phase 2: HTTP Request with Basic Authentication

    User->>PIVNET: HTTP Request
    PIVNET->>TRANSPORT: RoundTrip(request)
    TRANSPORT->>BASIC_AUTH: Authenticate(request)
    BASIC_AUTH->>BASIC_AUTH: Encode username:password<br/>to Base64
    BASIC_AUTH->>BASIC_AUTH: Add Proxy-Authorization header:<br/>Basic encoded-credentials
    BASIC_AUTH-->>TRANSPORT: Request authenticated
    TRANSPORT->>PROXY: HTTP Request with<br/>Proxy-Authorization header
    PROXY->>PROXY: Decode Base64 credentials<br/>Validate username & password
    alt Authentication successful
        PROXY->>API: Forward request to Pivnet API
        API->>API: Process request
        API->>PROXY: HTTP Response
        PROXY->>TRANSPORT: HTTP 200 OK
    else Authentication failed
        PROXY->>TRANSPORT: HTTP 407<br/>Proxy Auth Required
    end
    TRANSPORT-->>PIVNET: Return response
    PIVNET-->>User: Return response
```

## Component Summary

This sequence diagram demonstrates the interaction of the following components:

### om-cli Components
1. **commands/download_product.go**: Command-line interface, flag parsing, validation
2. **download_clients/pivnet_client.go**: Adapter layer, converts flags to ClientConfig

### go-pivnet Library Components
1. **pivnet.go (NewClient)**: Client factory, transport setup
2. **proxy_authenticator.go (NewProxyAuthenticator)**: Factory for creating authenticators
3. **proxy_auth_basic.go (BasicProxyAuth)**: Basic authentication implementation
4. **proxy_auth.go (ProxyAuthTransport)**: Transport wrapper that adds authentication

### External Services
1. **HTTP Proxy**: Enterprise proxy server requiring authentication
2. **Pivnet API**: Target API for product downloads

## Key Phases

### Phase 1: Client Initialization (Steps 1-18)
- User provides Basic authentication parameters via CLI flags
- om-cli parses flags and creates ClientConfig
- go-pivnet creates Basic authenticator (no external service needed)
- Transport is wrapped with authenticator
- Client is ready for use

### Phase 2: HTTP Request (Steps 19-32)
- HTTP request is initiated
- Basic authentication credentials are Base64 encoded
- Request is sent to proxy with Proxy-Authorization header
- Proxy validates credentials and forwards to Pivnet API (if successful)
- Response is returned to user

## Authentication Flow Details

### Basic Authentication Protocol
1. **Credential Encoding**: Username and password are combined as `username:password`
2. **Base64 Encoding**: The credentials string is Base64 encoded
3. **Header Addition**: The encoded credentials are added to the `Proxy-Authorization` header with the `Basic` scheme
4. **Format**: `Proxy-Authorization: Basic <base64-encoded-credentials>`

### Basic Authentication Header Format
```
Proxy-Authorization: Basic dXNlcm5hbWU6cGFzc3dvcmQ=
```

Where `dXNlcm5hbWU6cGFzc3dvcmQ=` is the Base64 encoding of `username:password`.

## Comparison with SPNEGO

### Key Differences

| Aspect | Basic Authentication | SPNEGO Authentication |
|--------|---------------------|----------------------|
| **Complexity** | Simple | Complex |
| **External Services** | None required | Requires Kerberos KDC |
| **Token Generation** | Base64 encoding only | Requires TGT and service tickets |
| **Security** | Credentials sent with each request | Tokens are time-limited |
| **Performance** | Fast (no network calls) | Slower (requires KDC communication) |
| **Setup** | Username/password only | Requires Kerberos configuration |

### When to Use

- **Basic Authentication**: Simple environments, quick setup, no Kerberos infrastructure
- **SPNEGO Authentication**: Enterprise environments with Kerberos, higher security requirements

## Error Handling

The sequence diagram shows the happy path. In practice, errors can occur at various points:

- **Invalid Credentials**: Proxy returns HTTP 407 if username/password are incorrect
- **Missing Credentials**: If username or password is empty, authentication may fail
- **Network Errors**: Connection failures to proxy or Pivnet API
- **Base64 Encoding Errors**: Rare, but possible if credentials contain invalid characters

## Related Documentation

- [SPNEGO_AUTHENTICATION_FLOW.md](./SPNEGO_AUTHENTICATION_FLOW.md) - SPNEGO authentication flow
- [PROXY_AUTHENTICATION_LOW_LEVEL_DESIGN.md](./PROXY_AUTHENTICATION_LOW_LEVEL_DESIGN.md) - Detailed low-level design
- [PROXY_AUTHENTICATION.md](./PROXY_AUTHENTICATION.md) - High-level overview
- [PROXY_AUTHENTICATION_QUICK_REFERENCE.md](./PROXY_AUTHENTICATION_QUICK_REFERENCE.md) - Quick reference guide

