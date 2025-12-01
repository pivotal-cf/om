# Low-Level Design: Proxy Authentication Mechanism in go-pivnet

This document provides a detailed low-level design specification for the proxy authentication mechanism implemented in the go-pivnet library, including data structures, algorithms, state management, and implementation details.

## Solution Architecture: om-cli and go-pivnet Collaboration

The proxy authentication solution involves **changes in both om-cli and go-pivnet library**, with a clear separation of concerns and well-defined interface boundaries. This collaborative approach allows each codebase to focus on its core responsibilities while providing a seamless user experience.

### om-cli Responsibilities

The om-cli application handles all user-facing concerns and acts as the entry point for proxy authentication configuration:

#### Command Layer (`commands/download_product.go`)
- **Flag Definition**: Defines command-line flags for proxy authentication:
  - `--proxy-url`: Proxy server URL
  - `--proxy-username`: Username for proxy authentication
  - `--proxy-password`: Password for proxy authentication
  - `--proxy-auth-type`: Type of authentication (basic, spnego)
  - `--proxy-krb5-config`: Path to Kerberos config file (for SPNEGO authentication)
- **Flag Parsing**: Uses `go-flags` library to parse command-line arguments
- **Validation**: Validates user input and provides helpful error messages
- **Data Structure**: Populates `PivnetOptions` struct with user-provided values

#### Adapter Layer (`download_clients/pivnet_client.go`)
- **Configuration Building**: Converts om-cli flags to go-pivnet's `ClientConfig` structure
- **Explicit Validation Logic**: Implements critical validation to ensure normal operation when flags are not provided:
  ```go
  // Only configure proxy settings if proxy URL is provided
  if proxyURL != "" {
      config.ProxyURL = proxyURL
      
      // Set proxy authentication if auth type is provided
      if proxyAuthType != "" {
          config.ProxyAuthType = pivnet.ProxyAuthType(proxyAuthType)
          // ... set other auth fields
      }
  }
  ```
- **Type Conversion**: Converts string flags to appropriate types (e.g., `ProxyAuthType`)
- **Interface Boundary**: Acts as the bridge between om-cli's command layer and go-pivnet's library layer

### go-pivnet Library Responsibilities

The go-pivnet library handles all authentication logic and transport management:

#### Client Factory (`pivnet.go`)
- **Configuration Consumption**: Receives `ClientConfig` from om-cli with proxy settings
- **Transport Setup**: Creates HTTP transport with proxy configuration:
  - Parses proxy URL and creates proxy function
  - Sets up base transport with TLS configuration
  - Conditionally wraps transport with authentication based on `ProxyAuthType`
- **Graceful Degradation**: Continues operation without authentication if authenticator creation fails (logs warning)
- **Client Creation**: Creates HTTP client with appropriate transport (authenticated or standard)

#### Authentication Factory (`proxy_authenticator.go`)
- **Factory Pattern**: Implements factory pattern to create appropriate authenticator based on `ProxyAuthType`
- **Type Validation**: Validates authentication type and parameters
- **Extensibility**: Easy to add new authentication types in the future

#### Authentication Implementations
- **Basic Authentication** (`proxy_auth_basic.go`):
  - Simple Base64 encoding of username:password
  - Adds `Proxy-Authorization: Basic <encoded>` header
  - No external dependencies
  
- **SPNEGO Authentication** (`proxy_auth_spnego.go`):
  - Complex Kerberos protocol implementation
  - Integrates with `gokrb5` library
  - Handles TGT acquisition, service ticket requests, and SPNEGO token generation
  - Supports custom Kerberos configuration files

#### Transport Wrapper (`proxy_auth.go`)
- **RoundTripper Implementation**: Implements `http.RoundTripper` interface
- **Request Interception**: Intercepts HTTP requests before they reach the proxy
- **Header Injection**: Calls authenticator to add `Proxy-Authorization` header
- **Transparent Operation**: Wraps existing transport without changing its behavior

### Interface Boundary

The interface between om-cli and go-pivnet is the **`ClientConfig` structure**:

```go
// Defined in go-pivnet library
type ClientConfig struct {
    Host              string
    UserAgent         string
    SkipSSLValidation bool
    ProxyURL          string        // Set by om-cli
    ProxyAuthType     ProxyAuthType // Set by om-cli
    ProxyUsername     string        // Set by om-cli
    ProxyPassword     string        // Set by om-cli
    ProxyKrb5Config   string        // Set by om-cli
}
```

**Data Flow**:
1. om-cli receives user input via command-line flags
2. om-cli validates and converts flags to `ClientConfig`
3. om-cli passes `ClientConfig` to `pivnet.NewClient()`
4. go-pivnet uses `ClientConfig` to set up authentication

### Benefits of This Architecture

1. **Separation of Concerns**: 
   - om-cli focuses on user experience (flag parsing, validation, error messages)
   - go-pivnet focuses on authentication logic (token generation, protocol handling)

2. **Reusability**: 
   - go-pivnet authentication can be used by other applications
   - om-cli can be extended with new commands without changing go-pivnet

3. **Testability**: 
   - Each layer can be tested independently
   - Clear interfaces make mocking easier

4. **Maintainability**: 
   - Changes to authentication logic don't affect om-cli command structure
   - Changes to command flags don't affect authentication implementation

5. **Backward Compatibility**: 
   - Explicit validation ensures normal operation when flags are not provided
   - Existing code continues to work without modification

### Code Locations

**om-cli Repository** (`/Users/hyayi/work/om/`):
- `commands/download_product.go`: Command-line flag definitions
- `download_clients/pivnet_client.go`: Adapter layer and configuration building

**go-pivnet Library** (`/Users/hyayi/work/go-pivnet/`):
- `pivnet.go`: Client factory and transport setup
- `proxy_authenticator.go`: Authenticator factory
- `proxy_auth_basic.go`: Basic authentication implementation
- `proxy_auth_spnego.go`: SPNEGO authentication implementation
- `proxy_auth.go`: Transport wrapper

### MVP Scope and Limitations

**Note: Caching is out of scope for this MVP**

This MVP implementation focuses on providing functional proxy authentication support. The following caching optimizations are explicitly **out of scope** for this release:

1. **Authentication Token Caching**: 
   - No explicit caching of SPNEGO tokens or Basic auth credentials
   - Each request generates authentication headers as needed
   - Note: The underlying `gokrb5` library for SPNEGO may cache Kerberos tickets internally, but this is library-managed, not application-managed

2. **Proxy Authentication State Caching**:
   - No caching of proxy authentication state across requests
   - No persistent storage of authentication tokens
   - No cross-session token reuse

3. **Configuration Caching**:
   - No caching of proxy configuration or credentials
   - Configuration is read from command-line flags on each execution

**Future Enhancements (Post-MVP)**:
- Token caching strategies to reduce KDC round-trips for SPNEGO
- Credential caching with secure storage
- Configuration persistence for frequently used proxy settings
- Performance optimizations through intelligent token reuse

The current implementation prioritizes correctness and security over performance optimizations that would require additional caching infrastructure.

## Table of Contents

1. [Component Ownership](#component-ownership)
   - [High-Level Component Architecture](#high-level-component-architecture)
2. [Data Structures](#data-structures)
3. [Type System](#type-system)
4. [Component Architecture](#component-architecture)
5. [Initialization Sequence](#initialization-sequence)
6. [Request Flow Algorithm](#request-flow-algorithm)
7. [Authentication Mechanisms](#authentication-mechanisms)
8. [Sequence Diagrams](#sequence-diagrams)
9. [State Management](#state-management)
10. [Error Handling](#error-handling)
11. [Resource Lifecycle](#resource-lifecycle)
12. [Thread Safety](#thread-safety)
13. [Performance Considerations](#performance-considerations)

---

## Component Ownership

This section clearly delineates which components belong to om-cli and which belong to the go-pivnet library. Understanding this separation is crucial for maintenance, debugging, and extending the functionality.

### Code Location Reference

**om-cli Repository:**
- Base path: `/Users/hyayi/work/om/`
- Command layer: `commands/download_product.go`
- Client adapter: `download_clients/pivnet_client.go`

**go-pivnet Library:**
- Base path: `/Users/hyayi/work/go-pivnet/`
- Client creation: `pivnet.go`
- Authentication: `proxy_auth*.go`, `proxy_authenticator.go`

### High-Level Component Architecture

The following diagram illustrates the high-level architecture and component relationships:

#### Mermaid Diagram: Component Architecture (Flowchart)

```mermaid
graph TB
    subgraph CLI["om-cli Application<br/>(/Users/hyayi/work/om/)"]
        subgraph UI["User Interface Layer"]
            CLI_CMD["Command Line Interface<br/>om download-product<br/>--proxy-url=...<br/>Flag parsing & validation"]
        end
        
        subgraph CMD["Command Layer"]
            DOWNLOAD_PRODUCT["commands/download_product.go"]
            PIVNET_OPTS["PivnetOptions struct<br/>- ProxyURL flag<br/>- ProxyUsername flag<br/>- ProxyPassword flag<br/>- ProxyAuthType flag<br/>- ProxyKrb5Config flag"]
            NEW_CLIENT_SRC["newDownloadClientFromSource()"]
        end
        
        subgraph ADAPTER["Download Clients Adapter Layer"]
            PIVNET_CLIENT["download_clients/pivnet_client.go"]
            NEW_PIVNET_CLIENT["NewPivnetClient()<br/>- Receives proxy flags<br/>- Validates & converts<br/>- Creates ClientConfig"]
        end
    end
    
    subgraph LIB["go-pivnet Library<br/>(/Users/hyayi/work/go-pivnet/)"]
        subgraph FACTORY["Client Factory Layer"]
            PIVNET_GO["pivnet.go"]
            NEW_CLIENT["NewClient()<br/>- Receives ClientConfig<br/>- Creates proxy function<br/>- Sets up HTTP transport<br/>- Conditionally wraps with auth"]
        end
        
        subgraph AUTH_FACTORY["Authentication Factory"]
            PROXY_AUTH_FACTORY["proxy_authenticator.go<br/>NewProxyAuthenticator()<br/>Factory Pattern"]
        end
        
        subgraph AUTH_IMPL["Authentication Implementations"]
            BASIC_AUTH["proxy_auth_basic.go<br/>BasicProxyAuth<br/>- username<br/>- password<br/>Authenticate(): Base64 encode"]
            SPNEGO_AUTH["proxy_auth_spnego.go<br/>SPNEGOProxyAuth<br/>- username<br/>- password<br/>- kerberosClient<br/>- proxyURL<br/>Authenticate(): Generate SPNEGO"]
            PROXY_AUTH_IFACE["ProxyAuthenticator Interface<br/>- Authenticate()<br/>- Close()"]
        end
        
        subgraph TRANSPORT["Transport Layer"]
            PROXY_AUTH_GO["proxy_auth.go<br/>ProxyAuthTransport<br/>- Transport: http.RoundTripper<br/>- Authenticator: ProxyAuthenticator<br/>RoundTrip():<br/>1. Authenticate()<br/>2. Transport.RoundTrip()"]
            HTTP_CLIENT["http.Client<br/>- Transport: ProxyAuthTransport<br/>  (or base transport)<br/>- Used for API & downloads"]
        end
    end
    
    subgraph EXT["External Components"]
        HTTP_PROXY["HTTP Proxy<br/>(Enterprise)<br/>- Authenticates requests<br/>- Forwards to Pivnet"]
        PIVNET_API["Pivnet API<br/>(network.pivotal.io)<br/>- Product API<br/>- File downloads"]
        KERBEROS_KDC["Kerberos KDC<br/>(For SPNEGO)<br/>- Issues TGT<br/>- Validates credentials"]
    end
    
    CLI_CMD -->|Command Options| DOWNLOAD_PRODUCT
    DOWNLOAD_PRODUCT --> PIVNET_OPTS
    PIVNET_OPTS --> NEW_CLIENT_SRC
    NEW_CLIENT_SRC -->|Function Call| NEW_PIVNET_CLIENT
    NEW_PIVNET_CLIENT -->|Creates ClientConfig| NEW_CLIENT
    NEW_CLIENT -->|If ProxyAuthType != ""| PROXY_AUTH_FACTORY
    PROXY_AUTH_FACTORY -->|Factory Pattern| BASIC_AUTH
    PROXY_AUTH_FACTORY -->|Factory Pattern| SPNEGO_AUTH
    BASIC_AUTH -->|implements| PROXY_AUTH_IFACE
    SPNEGO_AUTH -->|implements| PROXY_AUTH_IFACE
    PROXY_AUTH_IFACE -->|Wrapped by| PROXY_AUTH_GO
    PROXY_AUTH_GO --> HTTP_CLIENT
    HTTP_CLIENT -->|HTTP Requests<br/>with Proxy-Authorization| HTTP_PROXY
    HTTP_PROXY -->|Forwards| PIVNET_API
    SPNEGO_AUTH -.->|Kerberos Protocol| KERBEROS_KDC
    
    style CLI fill:#e1f5ff
    style LIB fill:#fff4e1
    style EXT fill:#ffe1e1
    style PROXY_AUTH_IFACE fill:#e1ffe1
    style HTTP_CLIENT fill:#f0f0f0
```

#### Mermaid Diagram: Component Dependency Graph

```mermaid
graph LR
    %% om-cli Components
    User[User CLI Input]
    PivnetOptions[PivnetOptions<br/>Struct]
    DownloadProduct[download_product.go]
    PivnetClientAdapter[pivnet_client.go<br/>NewPivnetClient]
    
    %% go-pivnet Components
    ClientConfig[ClientConfig<br/>Struct]
    NewClient[NewClient<br/>pivnet.go]
    ProxyAuthFactory[NewProxyAuthenticator<br/>proxy_authenticator.go]
    BasicAuth[BasicProxyAuth<br/>proxy_auth_basic.go]
    SPNEGOAuth[SPNEGOProxyAuth<br/>proxy_auth_spnego.go]
    ProxyAuthIface[ProxyAuthenticator<br/>Interface]
    ProxyAuthTransport[ProxyAuthTransport<br/>proxy_auth.go]
    HTTPClient[http.Client]
    
    %% External
    HTTPProxy[HTTP Proxy]
    PivnetAPI[Pivnet API]
    KerberosKDC[Kerberos KDC]
    
    %% om-cli flow
    User -->|1. Flags| PivnetOptions
    PivnetOptions -->|2. Contains| DownloadProduct
    DownloadProduct -->|3. Calls| PivnetClientAdapter
    PivnetClientAdapter -->|4. Creates| ClientConfig
    
    %% Library boundary
    ClientConfig -.->|5. Passed to| NewClient
    
    %% go-pivnet flow
    NewClient -->|6. Creates| ProxyAuthFactory
    ProxyAuthFactory -->|7a. Creates| BasicAuth
    ProxyAuthFactory -->|7b. Creates| SPNEGOAuth
    BasicAuth -.->|8. Implements| ProxyAuthIface
    SPNEGOAuth -.->|8. Implements| ProxyAuthIface
    ProxyAuthIface -->|9. Used by| ProxyAuthTransport
    NewClient -->|10. Wraps with| ProxyAuthTransport
    ProxyAuthTransport -->|11. Sets| HTTPClient
    
    %% External flow
    HTTPClient -->|12. Requests via| HTTPProxy
    HTTPProxy -->|13. Forwards to| PivnetAPI
    SPNEGOAuth -.->|14. Kerberos| KerberosKDC
    
    %% Styling
    classDef omCli fill:#e1f5ff,stroke:#0066cc,stroke-width:2px
    classDef goPivnet fill:#fff4e1,stroke:#cc6600,stroke-width:2px
    classDef external fill:#ffe1e1,stroke:#cc0000,stroke-width:2px
    classDef interface fill:#e1ffe1,stroke:#00cc00,stroke-width:2px
    
    class User,PivnetOptions,DownloadProduct,PivnetClientAdapter omCli
    class ClientConfig,NewClient,ProxyAuthFactory,BasicAuth,SPNEGOAuth,ProxyAuthTransport,HTTPClient goPivnet
    class HTTPProxy,PivnetAPI,KerberosKDC external
    class ProxyAuthIface interface
```

#### Mermaid Diagram: Component Relationship Graph

```mermaid
graph TD
    %% Central components
    ClientConfig[ClientConfig<br/>Configuration Data]
    
    %% om-cli layer
    subgraph OMCLI["om-cli Layer"]
        direction TB
        CLI[CLI Flags]
        Cmd[Command Layer]
        Adapter[Adapter Layer]
    end
    
    %% go-pivnet layer
    subgraph GOPIVNET["go-pivnet Library"]
        direction TB
        Factory[Client Factory]
        AuthFactory[Auth Factory]
        AuthImpl[Auth Implementations]
        Transport[Transport Wrapper]
    end
    
    %% External services
    subgraph EXTERNAL["External Services"]
        direction TB
        Proxy[HTTP Proxy]
        API[Pivnet API]
        KDC[Kerberos KDC]
    end
    
    %% Relationships
    CLI -->|Provides| Cmd
    Cmd -->|Calls| Adapter
    Adapter -->|Creates| ClientConfig
    ClientConfig -->|Configures| Factory
    Factory -->|Uses| AuthFactory
    AuthFactory -->|Creates| AuthImpl
    AuthImpl -->|Implements| Transport
    Transport -->|Sends to| Proxy
    Proxy -->|Forwards to| API
    AuthImpl -.->|Authenticates with| KDC
    
    %% Styling
    style ClientConfig fill:#ffffcc,stroke:#ffcc00,stroke-width:3px
    style OMCLI fill:#e1f5ff,stroke:#0066cc,stroke-width:2px
    style GOPIVNET fill:#fff4e1,stroke:#cc6600,stroke-width:2px
    style EXTERNAL fill:#ffe1e1,stroke:#cc0000,stroke-width:2px
```

#### ASCII Diagram (Alternative View)

The following diagram illustrates the high-level architecture and component relationships:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           om-cli Application                                │
│                        (/Users/hyayi/work/om/)                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────┐   │
│  │                    User Interface Layer                             │   │
│  │  ┌──────────────────────────────────────────────────────────────┐ │   │
│  │  │  Command Line Interface (CLI)                                 │ │   │
│  │  │  - om download-product --proxy-url=...                        │ │   │
│  │  │  - Flag parsing and validation                                │ │   │
│  │  └───────────────────────────┬────────────────────────────────────┘ │   │
│  │                              │                                        │   │
│  │                              │ Command Options                        │   │
│  │                              ↓                                        │   │
│  │  ┌──────────────────────────────────────────────────────────────┐ │   │
│  │  │  commands/download_product.go                                │ │   │
│  │  │  ┌────────────────────────────────────────────────────────┐  │ │   │
│  │  │  │  PivnetOptions struct                                  │  │ │   │
│  │  │  │  - ProxyURL (flag)                                     │  │ │   │
│  │  │  │  - ProxyUsername (flag)                                │  │ │   │
│  │  │  │  - ProxyPassword (flag)                                │  │ │   │
│  │  │  │  - ProxyAuthType (flag)                                │  │ │   │
│  │  │  │  - ProxyKrb5Config (flag)                              │  │ │   │
│  │  │  └───────────────────┬────────────────────────────────────┘  │ │   │
│  │  │                      │                                        │ │   │
│  │  │                      │ newDownloadClientFromSource()          │ │   │
│  │  │                      ↓                                        │ │   │
│  │  └──────────────────────┼──────────────────────────────────────┘ │   │
│  │                         │                                        │   │
│  └─────────────────────────┼────────────────────────────────────────┘   │
│                             │                                            │
│                             │ Function Call                              │
│                             ↓                                            │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │              Download Clients Adapter Layer                        │ │
│  │  ┌──────────────────────────────────────────────────────────────┐ │ │
│  │  │  download_clients/pivnet_client.go                           │ │ │
│  │  │  ┌────────────────────────────────────────────────────────┐ │ │ │
│  │  │  │  NewPivnetClient()                                     │ │ │ │
│  │  │  │  - Receives proxy flags from command layer              │ │ │ │
│  │  │  │  - Validates and converts to ClientConfig              │ │ │ │
│  │  │  │  - Explicit validation: only set proxy if URL provided │ │ │ │
│  │  │  └───────────────────┬────────────────────────────────────┘ │ │ │
│  │  │                      │                                        │ │ │
│  │  │                      │ Creates ClientConfig                   │ │ │
│  │  │                      ↓                                        │ │ │
│  │  └──────────────────────┼──────────────────────────────────────┘ │ │
│  │                         │                                        │ │
│  └─────────────────────────┼────────────────────────────────────────┘ │
│                             │                                            │
│                             │ Library Boundary                           │
│                             │ (import "github.com/pivotal-cf/go-pivnet/v7")│
│                             ↓                                            │
└─────────────────────────────┼────────────────────────────────────────────┘
                              │
                              │ pivnet.NewClient(config, ...)
                              │
┌─────────────────────────────┼────────────────────────────────────────────┐
│                    go-pivnet Library                                     │
│              (/Users/hyayi/work/go-pivnet/)                             │
├─────────────────────────────┼────────────────────────────────────────────┤
│                             │                                            │
│                             ↓                                            │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │                    Client Factory Layer                           │  │
│  │  ┌──────────────────────────────────────────────────────────────┐ │  │
│  │  │  pivnet.go                                                   │ │  │
│  │  │  ┌────────────────────────────────────────────────────────┐  │ │  │
│  │  │  │  NewClient()                                           │  │ │  │
│  │  │  │  - Receives ClientConfig                               │  │ │  │
│  │  │  │  - Creates proxy function                              │  │ │  │
│  │  │  │  - Sets up base HTTP transport                          │  │ │  │
│  │  │  │  - Conditionally wraps with proxy auth                 │  │ │  │
│  │  │  └───────────────┬────────────────────────────────────────┘  │ │  │
│  │  │                  │                                            │ │  │
│  │  │                  │ If ProxyAuthType != ""                    │ │  │
│  │  │                  ↓                                            │ │  │
│  │  │  ┌────────────────────────────────────────────────────────┐  │ │  │
│  │  │  │  NewProxyAuthenticator()                               │  │ │  │
│  │  │  │  (proxy_authenticator.go)                              │  │ │  │
│  │  │  └───────────────┬────────────────────────────────────────┘  │ │  │
│  │  │                  │                                            │ │  │
│  │  │                  │ Factory Pattern                           │ │  │
│  │  │                  ↓                                            │ │  │
│  │  │  ┌──────────────────────┐  ┌─────────────────────────────┐ │ │  │
│  │  │  │  Basic Authentication │  │  SPNEGO Authentication      │ │ │  │
│  │  │  │  ┌──────────────────┐ │  │  ┌───────────────────────┐ │ │ │  │
│  │  │  │  │proxy_auth_basic.go│ │  │  │proxy_auth_spnego.go  │ │ │ │  │
│  │  │  │  │                  │ │  │  │                      │ │ │ │  │
│  │  │  │  │BasicProxyAuth     │ │  │  │SPNEGOProxyAuth        │ │ │ │  │
│  │  │  │  │- username         │ │  │  │- username            │ │ │ │  │
│  │  │  │  │- password         │ │  │  │- password            │ │ │ │  │
│  │  │  │  │                  │ │  │  │- kerberosClient      │ │ │ │  │
│  │  │  │  │Authenticate():   │ │  │  │- proxyURL            │ │ │ │  │
│  │  │  │  │  Base64 encode   │ │  │  │                      │ │ │ │  │
│  │  │  │  │  Set header      │ │  │  │Authenticate():       │ │ │ │  │
│  │  │  │  └──────────────────┘ │  │  │  Generate SPNEGO     │ │ │ │  │
│  │  │  └──────────────────────┘  │  │  │  Set header         │ │ │ │  │
│  │  │                              │  │  └───────────────────────┘ │ │ │  │
│  │  │                              │  └─────────────────────────────┘ │ │  │
│  │  │                              │                                    │ │  │
│  │  │                              │ Both implement                    │ │  │
│  │  │                              ↓                                    │ │  │
│  │  │  ┌────────────────────────────────────────────────────────┐  │ │  │
│  │  │  │  ProxyAuthenticator interface                          │  │ │  │
│  │  │  │  - Authenticate(req *http.Request) error                │  │ │  │
│  │  │  │  - Close() error                                        │  │ │  │
│  │  │  └───────────────────┬────────────────────────────────────┘  │ │  │
│  │  │                      │                                        │ │  │
│  │  │                      │ Wrapped by                             │ │  │
│  │  │                      ↓                                        │ │  │
│  │  │  ┌────────────────────────────────────────────────────────┐  │ │  │
│  │  │  │  proxy_auth.go                                         │  │ │  │
│  │  │  │  ┌──────────────────────────────────────────────────┐  │ │  │
│  │  │  │  │  ProxyAuthTransport                              │  │ │  │
│  │  │  │  │  - Transport (http.RoundTripper)                 │  │ │  │
│  │  │  │  │  - Authenticator (ProxyAuthenticator)             │  │ │  │
│  │  │  │  │                                                  │  │ │  │
│  │  │  │  │  RoundTrip():                                    │  │ │  │
│  │  │  │  │    1. Call Authenticator.Authenticate()           │  │ │  │
│  │  │  │  │    2. Call Transport.RoundTrip()                 │  │ │  │
│  │  │  │  └──────────────────────────────────────────────────┘  │ │  │
│  │  │  └────────────────────────────────────────────────────────┘  │ │  │
│  │  │                                                               │ │  │
│  │  │                                                               │ │  │
│  │  │  ┌────────────────────────────────────────────────────────┐  │ │  │
│  │  │  │  http.Client                                           │  │ │  │
│  │  │  │  - Transport: ProxyAuthTransport (or base transport)   │  │ │  │
│  │  │  │  - Used for API calls and downloads                    │  │ │  │
│  │  │  └────────────────────────────────────────────────────────┘  │ │  │
│  │  └──────────────────────────────────────────────────────────────┘ │  │
│  └────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                              │
                              │ HTTP Requests
                              │ (with Proxy-Authorization header)
                              ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                        External Components                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────┐              ┌──────────────────┐                   │
│  │  HTTP Proxy      │              │  Pivnet API      │                   │
│  │  (Enterprise)    │              │  (network.pivotal.io)                 │
│  │                  │              │                  │                   │
│  │  - Authenticates │              │  - Product API  │                   │
│  │    requests      │──────────────>│  - File downloads│                   │
│  │  - Forwards to   │              │                  │                   │
│  │    Pivnet        │              │                  │                   │
│  └──────────────────┘              └──────────────────┘                   │
│                                                                              │
│  ┌──────────────────┐                                                       │
│  │  Kerberos KDC    │                                                       │
│  │  (For SPNEGO)    │                                                       │
│  │                  │                                                       │
│  │  - Issues TGT    │                                                       │
│  │  - Validates     │                                                       │
│  │    credentials   │                                                       │
│  └──────────────────┘                                                       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Component Interaction Flow

#### Mermaid Diagram: Interaction Flow

```mermaid
sequenceDiagram
    participant User as User (CLI)
    participant CMD as download_product.go<br/>(om-cli)
    participant ADAPTER as pivnet_client.go<br/>(om-cli)
    participant PIVNET as pivnet.go<br/>(go-pivnet)
    participant AUTH_FACTORY as proxy_authenticator.go<br/>(go-pivnet)
    participant AUTH_IMPL as BasicProxyAuth<br/>or SPNEGOProxyAuth<br/>(go-pivnet)
    participant TRANSPORT as proxy_auth.go<br/>ProxyAuthTransport<br/>(go-pivnet)
    participant PROXY as HTTP Proxy<br/>(External)
    participant API as Pivnet API<br/>(External)
    
    User->>CMD: 1. om download-product<br/>--proxy-url=...<br/>--proxy-auth-type=basic
    Note over CMD: Parse flags<br/>Validate input
    CMD->>ADAPTER: 2. Call NewPivnetClient()
    Note over ADAPTER: Convert flags<br/>Build ClientConfig
    ADAPTER->>PIVNET: 3. pivnet.NewClient(config)
    Note over PIVNET: Create transport<br/>Check ProxyAuthType
    PIVNET->>AUTH_FACTORY: 4. NewProxyAuthenticator()
    Note over AUTH_FACTORY: Factory pattern
    AUTH_FACTORY->>AUTH_IMPL: 5. Create BasicProxyAuth<br/>or SPNEGOProxyAuth
    AUTH_IMPL-->>AUTH_FACTORY: Return authenticator
    AUTH_FACTORY-->>PIVNET: Return authenticator
    PIVNET->>TRANSPORT: Wrap transport with auth
    TRANSPORT-->>PIVNET: Return ProxyAuthTransport
    PIVNET-->>ADAPTER: Return Client
    ADAPTER-->>CMD: Return ProductDownloader
    CMD-->>User: Client ready
    
    Note over User,API: Runtime: HTTP Request Flow
    User->>TRANSPORT: HTTP Request
    TRANSPORT->>AUTH_IMPL: Authenticate(req)
    AUTH_IMPL->>AUTH_IMPL: Add Proxy-Authorization header
    AUTH_IMPL-->>TRANSPORT: Request authenticated
    TRANSPORT->>PROXY: 6. HTTP Request with<br/>Proxy-Authorization header
    PROXY->>PROXY: Validate authentication
    PROXY->>API: Forward to Pivnet API
    API-->>PROXY: Response
    PROXY-->>TRANSPORT: Response
    TRANSPORT-->>User: Response
```

#### ASCII Diagram (Alternative View)

```
┌─────────────┐
│   User      │
│  (CLI)      │
└──────┬──────┘
       │
       │ 1. om download-product --proxy-url=... --proxy-auth-type=basic
       ↓
┌─────────────────────────────────────────────────────────────────┐
│  om-cli Layer                                                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────┐                                      │
│  │ download_product.go  │                                      │
│  │ - Parse flags        │                                      │
│  │ - Validate input     │                                      │
│  └──────────┬───────────┘                                      │
│             │                                                    │
│             │ 2. Call NewPivnetClient()                          │
│             ↓                                                    │
│  ┌──────────────────────┐                                      │
│  │ pivnet_client.go     │                                      │
│  │ - Convert flags      │                                      │
│  │ - Build ClientConfig │                                      │
│  └──────────┬───────────┘                                      │
│             │                                                    │
└─────────────┼────────────────────────────────────────────────────┘
              │
              │ 3. pivnet.NewClient(config)
              ↓
┌─────────────────────────────────────────────────────────────────┐
│  go-pivnet Library                                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────┐                                      │
│  │ pivnet.go            │                                      │
│  │ - Create transport   │                                      │
│  │ - Check ProxyAuthType│                                      │
│  └──────────┬───────────┘                                      │
│             │                                                    │
│             │ 4. NewProxyAuthenticator()                         │
│             ↓                                                    │
│  ┌──────────────────────┐                                      │
│  │ proxy_authenticator │                                      │
│  │ .go                 │                                      │
│  │ - Factory pattern   │                                      │
│  └──────────┬───────────┘                                      │
│             │                                                    │
│             │ 5. Create BasicProxyAuth or SPNEGOProxyAuth        │
│             ↓                                                    │
│  ┌──────────────────────┐                                      │
│  │ proxy_auth.go        │                                      │
│  │ - Wrap transport     │                                      │
│  │ - Create client      │                                      │
│  └──────────┬───────────┘                                      │
│             │                                                    │
└─────────────┼────────────────────────────────────────────────────┘
              │
              │ 6. HTTP Request with Proxy-Authorization header
              ↓
┌─────────────────────────────────────────────────────────────────┐
│  External Services                                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐         ┌──────────────┐                    │
│  │ HTTP Proxy   │────────>│ Pivnet API   │                    │
│  │ (authenticated)│         │              │                    │
│  └──────────────┘         └──────────────┘                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities Summary

| Layer | Component | Responsibility | Ownership |
|-------|-----------|----------------|-----------|
| **User Interface** | CLI | Parse command-line flags | om-cli |
| **Command Layer** | `download_product.go` | Validate flags, route to client factory | om-cli |
| **Adapter Layer** | `pivnet_client.go` | Convert flags to ClientConfig, validate | om-cli |
| **Client Factory** | `pivnet.go` | Create HTTP client with proxy support | go-pivnet |
| **Auth Factory** | `proxy_authenticator.go` | Create appropriate authenticator | go-pivnet |
| **Basic Auth** | `proxy_auth_basic.go` | Implement Basic authentication | go-pivnet |
| **SPNEGO Auth** | `proxy_auth_spnego.go` | Implement SPNEGO authentication | go-pivnet |
| **Transport Wrapper** | `proxy_auth.go` | Wrap transport with authentication | go-pivnet |
| **HTTP Client** | `http.Client` | Execute HTTP requests | Go stdlib |

### om-cli Components

**Location:** `/Users/hyayi/work/om/`

#### Command Layer (`commands/download_product.go`)

```go
// om-cli: commands/download_product.go

type PivnetOptions struct {
    // ... other fields ...
    ProxyURL        string `long:"proxy-url"`
    ProxyUsername   string `long:"proxy-username"`
    ProxyPassword   string `long:"proxy-password"`
    ProxyAuthType   string `long:"proxy-auth-type"`
    ProxyKrb5Config string `long:"proxy-krb5-config"`
}

// Function that creates client from command options
func newDownloadClientFromSource(c DownloadProductOptions, ...) {
    // ...
    case "pivnet", "":
        return download_clients.NewPivnetClient(
            stdout, stderr,
            download_clients.DefaultPivnetFactory,
            c.PivnetToken,
            c.PivnetDisableSSL,
            c.PivnetHost,
            c.ProxyURL,           // ← om-cli passes flags
            c.ProxyUsername,      // ← om-cli passes flags
            c.ProxyPassword,      // ← om-cli passes flags
            c.ProxyAuthType,      // ← om-cli passes flags
            c.ProxyKrb5Config,    // ← om-cli passes flags
        ), nil
}
```

**Responsibilities:**
- Parse command-line flags
- Validate user input
- Pass proxy configuration to download_clients layer

#### Download Clients Layer (`download_clients/pivnet_client.go`)

```go
// File: om-cli/download_clients/pivnet_client.go
// Library: om-cli
// Location: /Users/hyayi/work/om/download_clients/pivnet_client.go

var NewPivnetClient = func(
    stdout *log.Logger,
    stderr *log.Logger,
    factory PivnetFactory,
    token string,
    skipSSL bool,
    pivnetHost string,
    proxyURL string,           // ← Received from om-cli command layer
    proxyUsername string,       // ← Received from om-cli command layer
    proxyPassword string,       // ← Received from om-cli command layer
    proxyAuthType string,       // ← Received from om-cli command layer
    proxyKrb5Config string,    // ← Received from om-cli command layer
) ProductDownloader {
    logger := logshim.NewLogShim(stdout, stderr, false)
    tokenGenerator := pivnet.NewAccessTokenOrLegacyToken(
        token, pivnetHost, skipSSL, userAgent,
    )
    
    // Create base config without proxy settings
    // [om-cli] Explicit validation: only set proxy if URL provided
    config := pivnet.ClientConfig{
        Host:              pivnetHost,
        UserAgent:         userAgent,
        SkipSSLValidation: skipSSL,
    }
    
    // Only configure proxy settings if proxy URL is provided
    // [om-cli] This explicit check ensures normal operation when flags not set
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
    
    // [om-cli → go-pivnet] Call go-pivnet library with config
    downloader := factory(tokenGenerator, config, logger)
    client := pivnet.NewClient(tokenGenerator, config, logger)
    
    return &pivnetClient{
        filter:     filter.NewFilter(logger),
        downloader: downloader,
        stderr:     stderr,
        client:     client,
    }
}
```

**Responsibilities:**
- Bridge between om-cli command layer and go-pivnet library
- Convert om-cli parameters to go-pivnet ClientConfig
- Explicit validation logic (only set proxy if URL provided)
- Maintains ProductDownloader interface for om-cli

### go-pivnet Library Components

**Location:** `/Users/hyayi/work/go-pivnet/`

#### Client Configuration (`pivnet.go`)

```go
// File: go-pivnet/pivnet.go
// Library: go-pivnet
// Location: /Users/hyayi/work/go-pivnet/pivnet.go

type ClientConfig struct {
    Host              string
    UserAgent         string
    SkipSSLValidation bool
    ProxyURL          string        // ← Set by om-cli, used by go-pivnet
    ProxyAuthType     ProxyAuthType // ← Set by om-cli, used by go-pivnet
    ProxyUsername     string        // ← Set by om-cli, used by go-pivnet
    ProxyPassword     string        // ← Set by om-cli, used by go-pivnet
    ProxyKrb5Config   string        // ← Set by om-cli, used by go-pivnet
}
```

**Note:** This struct is defined in go-pivnet but populated by om-cli through the `download_clients` layer.

#### Client Creation (`pivnet.go`)

```go
// File: go-pivnet/pivnet.go
// Library: go-pivnet
// Location: /Users/hyayi/work/go-pivnet/pivnet.go

func NewClient(
    token AccessTokenService,
    config ClientConfig,  // ← Contains proxy settings populated by om-cli
    logger logger.Logger,
) Client {
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
    // ...
}
```

#### Proxy Authentication Components

**All in go-pivnet library:**

1. **Proxy Authenticator Interface** (`proxy_authenticator.go`)
   ```go
   // File: go-pivnet/proxy_authenticator.go
   // Library: go-pivnet
   // Location: /Users/hyayi/work/go-pivnet/proxy_authenticator.go
   
   type ProxyAuthenticator interface {
       Authenticate(req *http.Request) error
       Close() error
   }
   ```

2. **Basic Authentication** (`proxy_auth_basic.go`)
   ```go
   // File: go-pivnet/proxy_auth_basic.go
   // Library: go-pivnet
   // Location: /Users/hyayi/work/go-pivnet/proxy_auth_basic.go
   
   type BasicProxyAuth struct { ... }
   func NewBasicProxyAuth(...) *BasicProxyAuth { ... }
   func (b *BasicProxyAuth) Authenticate(...) error { ... }
   ```

3. **SPNEGO Authentication** (`proxy_auth_spnego.go`)
   ```go
   // File: go-pivnet/proxy_auth_spnego.go
   // Library: go-pivnet
   // Location: /Users/hyayi/work/go-pivnet/proxy_auth_spnego.go
   
   type SPNEGOProxyAuth struct { ... }
   func NewSPNEGOProxyAuth(...) (*SPNEGOProxyAuth, error) { ... }
   func (s *SPNEGOProxyAuth) Authenticate(...) error { ... }
   ```

4. **Proxy Auth Transport** (`proxy_auth.go`)
   ```go
   // File: go-pivnet/proxy_auth.go
   // Library: go-pivnet
   // Location: /Users/hyayi/work/go-pivnet/proxy_auth.go
   
   type ProxyAuthTransport struct { ... }
   func NewProxyAuthTransport(...) (*ProxyAuthTransport, error) { ... }
   func (t *ProxyAuthTransport) RoundTrip(...) (*http.Response, error) { ... }
   ```

5. **Factory Function** (`proxy_authenticator.go`)
   ```go
   // File: go-pivnet/proxy_authenticator.go
   // Library: go-pivnet
   // Location: /Users/hyayi/work/go-pivnet/proxy_authenticator.go
   
   func NewProxyAuthenticator(...) (ProxyAuthenticator, error) { ... }
   ```

### Component Ownership Summary

| Component | File Path | Ownership | Purpose | Called By |
|-----------|-----------|-----------|---------|-----------|
| `PivnetOptions` | `om/commands/download_product.go` | om-cli | Command-line flag definitions | User (via CLI) |
| `newDownloadClientFromSource` | `om/commands/download_product.go` | om-cli | Creates client from command options | `DownloadProduct.Execute()` |
| `NewPivnetClient` | `om/download_clients/pivnet_client.go` | om-cli | Bridge/adapter layer | `newDownloadClientFromSource()` |
| `ClientConfig` | `go-pivnet/pivnet.go` | go-pivnet | Configuration structure | Populated by om-cli |
| `NewClient` | `go-pivnet/pivnet.go` | go-pivnet | Client factory with proxy support | `NewPivnetClient()` (om-cli) |
| `ProxyAuthenticator` | `go-pivnet/proxy_authenticator.go` | go-pivnet | Authentication interface | N/A (interface) |
| `BasicProxyAuth` | `go-pivnet/proxy_auth_basic.go` | go-pivnet | Basic auth implementation | `NewProxyAuthenticator()` |
| `SPNEGOProxyAuth` | `go-pivnet/proxy_auth_spnego.go` | go-pivnet | SPNEGO auth implementation | `NewProxyAuthenticator()` |
| `ProxyAuthTransport` | `go-pivnet/proxy_auth.go` | go-pivnet | Transport wrapper | `NewClient()` |
| `NewProxyAuthenticator` | `go-pivnet/proxy_authenticator.go` | go-pivnet | Authenticator factory | `NewClient()` |

### Data Flow: om-cli → go-pivnet

```
┌─────────────────────────────────────────────────────────────────┐
│                        om-cli Layer                              │
│                    (/Users/hyayi/work/om/)                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  User Input (Command Line)                                      │
│    ↓                                                            │
│  commands/download_product.go                                   │
│    [om-cli] PivnetOptions struct (flags)                        │
│    - Flag parsing and validation                                │
│    - Flag: --proxy-url, --proxy-username, etc.                 │
│    ↓                                                            │
│  download_clients/pivnet_client.go                               │
│    [om-cli] NewPivnetClient() function                          │
│    - Parameter conversion (string → ProxyAuthType)             │
│    - Explicit validation logic                                  │
│    - Only sets proxy config if proxyURL != ""                   │
│                                                                  │
└────────────────────────────┬────────────────────────────────────┘
                              │
                              │ ClientConfig struct
                              │ (proxy settings passed)
                              │ Interface Boundary
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    go-pivnet Library Layer                       │
│              (/Users/hyayi/work/go-pivnet/)                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  pivnet.go                                                      │
│    [go-pivnet] NewClient() function                             │
│    - Receives ClientConfig with proxy settings                  │
│    - Proxy function creation                                    │
│    - Transport setup                                            │
│    - Graceful degradation on error                              │
│    ↓                                                            │
│  proxy_authenticator.go                                         │
│    [go-pivnet] NewProxyAuthenticator() factory                  │
│    - Creates BasicProxyAuth or SPNEGOProxyAuth                  │
│    - Validates parameters                                       │
│    ↓                                                            │
│  proxy_auth_basic.go OR proxy_auth_spnego.go                    │
│    [go-pivnet] Authenticator implementation                     │
│    - Basic: Base64 encoding                                     │
│    - SPNEGO: Kerberos token generation                          │
│    ↓                                                            │
│  proxy_auth.go                                                  │
│    [go-pivnet] ProxyAuthTransport wrapper                       │
│    - Wraps http.Transport                                       │
│    - Calls Authenticator.Authenticate() per request             │
│    ↓                                                            │
│  HTTP Client (with authenticated transport)                     │
│    [go-pivnet] Ready for use                                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Code Boundary Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         om-cli Code                              │
│                    (User-facing layer)                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────┐              │
│  │ commands/download_product.go                  │              │
│  │  - PivnetOptions (flag definitions)          │              │
│  │  - newDownloadClientFromSource()              │              │
│  └──────────────────┬───────────────────────────┘              │
│                     │                                            │
│                     │ Calls                                      │
│                     ↓                                            │
│  ┌──────────────────────────────────────────────┐              │
│  │ download_clients/pivnet_client.go             │              │
│  │  - NewPivnetClient()                         │              │
│  │  - Converts flags → ClientConfig              │              │
│  │  - Explicit validation                        │              │
│  └──────────────────┬───────────────────────────┘              │
│                     │                                            │
│                     │ Calls                                      │
│                     ↓                                            │
│  ┌──────────────────────────────────────────────┐              │
│  │ import "github.com/pivotal-cf/go-pivnet/v7"   │              │
│  │                                               │              │
│  │ pivnet.NewClient(config, ...)                 │              │
│  └──────────────────────────────────────────────┘              │
│                                                                  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ Library Boundary
                             │ (go-pivnet package)
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│                      go-pivnet Library Code                     │
│                    (Reusable library)                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────┐              │
│  │ pivnet.go                                    │              │
│  │  - ClientConfig struct                       │              │
│  │  - NewClient() function                      │              │
│  └──────────────────┬───────────────────────────┘              │
│                     │                                            │
│                     │ Uses                                       │
│                     ↓                                            │
│  ┌──────────────────────────────────────────────┐              │
│  │ proxy_authenticator.go                       │              │
│  │  - ProxyAuthenticator interface              │              │
│  │  - NewProxyAuthenticator() factory            │              │
│  └──────┬───────────────────────┬───────────────┘              │
│         │                       │                               │
│         │ Creates                │ Creates                       │
│         ↓                        ↓                               │
│  ┌──────────────┐      ┌──────────────────────┐                │
│  │proxy_auth_   │      │proxy_auth_spnego.go  │                │
│  │basic.go      │      │                      │                │
│  │              │      │                      │                │
│  │BasicProxyAuth│      │SPNEGOProxyAuth       │                │
│  └──────┬───────┘      └──────────┬───────────┘                │
│         │                         │                             │
│         └──────────┬──────────────┘                             │
│                    │                                            │
│                    │ Used by                                    │
│                    ↓                                            │
│  ┌──────────────────────────────────────────────┐              │
│  │ proxy_auth.go                                │              │
│  │  - ProxyAuthTransport                        │              │
│  │  - Wraps transport with authentication       │              │
│  └──────────────────────────────────────────────┘              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Interface Boundary

The interface between om-cli and go-pivnet is:

**om-cli side:**
```go
// om-cli: download_clients/pivnet_client.go
download_clients.NewPivnetClient(
    stdout, stderr, factory,
    token, skipSSL, pivnetHost,
    proxyURL, proxyUsername, proxyPassword,
    proxyAuthType, proxyKrb5Config,
)
```

**go-pivnet side:**
```go
// go-pivnet: pivnet.go
pivnet.NewClient(
    token,
    ClientConfig{
        Host: ...,
        ProxyURL: ...,
        ProxyAuthType: ...,
        ProxyUsername: ...,
        ProxyPassword: ...,
        ProxyKrb5Config: ...,
    },
    logger,
)
```

**Key Points:**
- om-cli converts command flags to go-pivnet ClientConfig
- go-pivnet library handles all proxy authentication logic
- om-cli provides explicit validation (only set proxy if URL provided)
- go-pivnet provides graceful degradation (continues without auth on error)

---

## Data Structures

### Complete Data Structure Definitions

#### ClientConfig

```go
type ClientConfig struct {
    Host              string        // Pivnet API host (e.g., "https://network.pivotal.io")
    UserAgent         string        // User agent string for HTTP requests
    SkipSSLValidation bool          // TLS certificate validation flag
    ProxyURL          string        // Proxy server URL (e.g., "http://proxy.example.com:8080")
    ProxyAuthType     ProxyAuthType // Authentication mechanism type
    ProxyUsername     string        // Username for proxy authentication
    ProxyPassword     string        // Password for proxy authentication
    ProxyKrb5Config   string        // Path to Kerberos configuration file (krb5.conf)
}
```

**Field Specifications:**

| Field | Type | Zero Value | Validation | Constraints |
|-------|------|------------|------------|-------------|
| `Host` | `string` | `""` | Non-empty for valid client | Must be valid URL |
| `UserAgent` | `string` | `""` | Optional | Recommended format: "app/version" |
| `SkipSSLValidation` | `bool` | `false` | N/A | `true` disables TLS verification |
| `ProxyURL` | `string` | `""` | Valid URL if provided | Format: `http://host:port` or `https://host:port` |
| `ProxyAuthType` | `ProxyAuthType` | `""` | Must be `"basic"` or `"spnego"` if set | Enum: `ProxyAuthTypeBasic`, `ProxyAuthTypeSPNEGO` |
| `ProxyUsername` | `string` | `""` | Required if `ProxyAuthType` set | No length limit, UTF-8 encoded |
| `ProxyPassword` | `string` | `""` | Required if `ProxyAuthType` set | No length limit, UTF-8 encoded |
| `ProxyKrb5Config` | `string` | `""` | Valid file path if provided | Absolute or relative path to krb5.conf |

**Memory Layout (64-bit system):**

```
Offset  Size    Field              Type
------  ----    -----              ----
0       8       Host               *string (pointer)
8       8       UserAgent          *string (pointer)
16      8       SkipSSLValidation  bool (1 byte, 7 padding)
24      8       ProxyURL           *string (pointer)
32      8       ProxyAuthType      *string (pointer)
40      8       ProxyUsername      *string (pointer)
48      8       ProxyPassword      *string (pointer)
56      8       ProxyKrb5Config    *string (pointer)
Total: 64 bytes (struct) + string data (heap)
```

**String Memory (per string field):**
- Header: 16 bytes (length + data pointer)
- Data: N bytes (UTF-8 encoded, heap-allocated)

**Total Memory Estimate:**
- Struct: 64 bytes
- String headers: 8 × 16 = 128 bytes (if all fields set)
- String data: Variable (depends on content)
- **Minimum**: ~192 bytes (empty strings)
- **Typical**: ~500-1000 bytes (with typical values)

### ProxyAuthType

```go
type ProxyAuthType string

const (
    ProxyAuthTypeBasic  ProxyAuthType = "basic"  // 5 bytes + string overhead
    ProxyAuthTypeSPNEGO ProxyAuthType = "spnego" // 6 bytes + string overhead
)
```

**Type Characteristics:**
- Zero value: `""` (empty string)
- Comparison: string equality (`==`)
- Immutable: constants defined at compile time
- Type safety: Compile-time checked (cannot assign arbitrary string without conversion)

**Memory Layout:**
```
ProxyAuthTypeBasic:
  String header: 16 bytes
  Data: "basic" (5 bytes)
  Total: 21 bytes

ProxyAuthTypeSPNEGO:
  String header: 16 bytes
  Data: "spnego" (6 bytes)
  Total: 22 bytes
```

**Type Conversion Algorithm:**
```go
// String to ProxyAuthType (explicit conversion required)
authType := ProxyAuthType("basic")  // O(1) - pointer assignment

// ProxyAuthType to string (implicit conversion)
str := string(authType)  // O(1) - no copy, same underlying data
```

**Validation Function:**
```go
func IsValidProxyAuthType(s string) bool {
    return s == string(ProxyAuthTypeBasic) || 
           s == string(ProxyAuthTypeSPNEGO) ||
           s == ""
}
// Time: O(1), Space: O(1)
```

### ProxyAuthenticator Interface

```go
type ProxyAuthenticator interface {
    Authenticate(req *http.Request) error
    Close() error
}
```

**Interface Contract:**

| Method | Preconditions | Postconditions | Side Effects | Thread Safety |
|--------|---------------|----------------|--------------|---------------|
| `Authenticate(req)` | `req != nil` | `req.Header` contains auth header | Modifies `req.Header` | Must be thread-safe |
| `Close()` | Object initialized | All resources released | Network connections closed, memory freed | Must be idempotent |

**Interface Invariants:**
1. `Authenticate()` must be idempotent: calling multiple times produces same result
2. `Authenticate()` must not block indefinitely
3. `Close()` must be safe to call multiple times
4. After `Close()`, `Authenticate()` behavior is undefined (may panic)

**Interface Size:**
- Interface value: 16 bytes (2 pointers: type descriptor + value)
- Method table: Shared across all instances (no per-instance overhead)

### ProxyAuthTransport

```go
type ProxyAuthTransport struct {
    Transport     http.RoundTripper  // Wrapped transport (typically *http.Transport)
    Authenticator ProxyAuthenticator // Authentication mechanism implementation
}
```

**Memory Layout (64-bit system):**

```
Offset  Size    Field          Type
------  ----    -----          ----
0       8       Transport      *http.RoundTripper (interface pointer)
8       8       Authenticator  *ProxyAuthenticator (interface pointer)
Total: 16 bytes
```

**Field Specifications:**

| Field | Type | Ownership | Lifetime | Thread Safety |
|-------|------|-----------|----------|---------------|
| `Transport` | `http.RoundTripper` | Reference (not owned) | Managed by caller | Must be thread-safe |
| `Authenticator` | `ProxyAuthenticator` | Owned | Until `Close()` called | Must be thread-safe |

**Invariants:**
1. `Transport != nil` (enforced by construction)
2. `Authenticator != nil` (enforced by `NewProxyAuthTransport`)
3. `Transport` is not closed by this struct
4. `Authenticator` must be closed via `Close()` method

**Interface Implementation:**
```go
// Implements http.RoundTripper
var _ http.RoundTripper = (*ProxyAuthTransport)(nil)

// RoundTrip signature matches http.RoundTripper
func (t *ProxyAuthTransport) RoundTrip(req *http.Request) (*http.Response, error)
```

### BasicProxyAuth

```go
type BasicProxyAuth struct {
    username string  // Username for Basic authentication
    password string  // Password for Basic authentication
}
```

**Memory Layout (64-bit system):**

```
Offset  Size    Field      Type
------  ----    -----      ----
0       8       username   *string (pointer to string header)
8       8       password   *string (pointer to string header)
Total: 16 bytes (struct) + string data (heap)
```

**Field Specifications:**

| Field | Type | Zero Value | Validation | Encoding |
|-------|------|------------|------------|----------|
| `username` | `string` | `""` | Optional (empty allowed) | UTF-8 |
| `password` | `string` | `""` | Optional (empty allowed) | UTF-8 |

**Memory Characteristics:**
- Struct size: 16 bytes (2 pointers)
- String data: Heap-allocated, immutable after construction
- Total per instance: 16 bytes + string overhead (32 bytes) + string data
- No external resources: No cleanup required
- Stateless: No mutable state between calls

**Construction Algorithm:**
```go
func NewBasicProxyAuth(username, password string) *BasicProxyAuth {
    return &BasicProxyAuth{
        username: username,  // String assignment (O(1) - pointer copy)
        password: password, // String assignment (O(1) - pointer copy)
    }
}
// Time: O(1), Space: O(1) + string data
```

### SPNEGOProxyAuth

```go
type SPNEGOProxyAuth struct {
    username       string          // Username for Kerberos authentication
    password       string          // Password for Kerberos authentication
    proxyURL       string          // Proxy URL (used for SPN derivation)
    kerberosClient *client.Client  // Kerberos client (gokrb5 library)
}
```

**Memory Layout (64-bit system):**

```
Offset  Size    Field           Type
------  ----    -----           ----
0       8       username        *string (pointer)
8       8       password        *string (pointer)
16      8       proxyURL        *string (pointer)
24      8       kerberosClient  *client.Client (pointer)
Total: 32 bytes (struct) + string data + Kerberos client state (heap)
```

**Field Specifications:**

| Field | Type | Zero Value | Validation | Lifetime |
|-------|------|------------|------------|----------|
| `username` | `string` | `""` | Required, non-empty | Immutable after construction |
| `password` | `string` | `""` | Required, non-empty | Immutable after construction |
| `proxyURL` | `string` | `""` | Required, valid URL | Immutable after construction |
| `kerberosClient` | `*client.Client` | `nil` | Created during initialization | Owned, must be destroyed |

**Kerberos Client State (gokrb5 library):**
- TGT (Ticket Granting Ticket): ~1-4 KB
- Service tickets cache: Variable (typically 1-10 KB)
- Network connections: TCP sockets to KDC
- Cryptographic state: Session keys, nonces
- **Estimated total**: 5-20 KB per instance

**Memory Characteristics:**
- Struct size: 32 bytes
- String data: ~100-500 bytes (typical)
- Kerberos client: 5-20 KB
- **Total per instance**: ~5-20 KB

**State Transitions:**
```
kerberosClient == nil  →  Uninitialized
kerberosClient != nil   →  Initialized (ready for use)
After Close()           →  kerberosClient == nil (closed)
```

---

## Type System

### Type Hierarchy

```
ProxyAuthType (string)
    ├── ProxyAuthTypeBasic ("basic")
    └── ProxyAuthTypeSPNEGO ("spnego")

ProxyAuthenticator (interface)
    ├── *BasicProxyAuth (concrete type)
    └── *SPNEGOProxyAuth (concrete type)

http.RoundTripper (interface)
    └── *ProxyAuthTransport (wrapper implementation)
            └── *http.Transport (wrapped transport)
```

### Type Assertions and Conversions

```go
// Type conversion: string → ProxyAuthType
authType := ProxyAuthType("basic")  // Explicit conversion

// Type assertion: interface → concrete type
if basicAuth, ok := authenticator.(*BasicProxyAuth); ok {
    // Use basicAuth
}

// Interface satisfaction check (compile-time)
var _ ProxyAuthenticator = (*BasicProxyAuth)(nil)
var _ ProxyAuthenticator = (*SPNEGOProxyAuth)(nil)
```

---

## Component Architecture

### Component Dependency Graph

```
┌─────────────────────────────────────────────────────────────┐
│                    NewClient()                              │
│  - Input: ClientConfig                                      │
│  - Output: Client                                           │
└────────────┬───────────────────────────────────────────────┘
             │
             ├──────────────────────────────────────────────┐
             │                                               │
             ▼                                               ▼
┌────────────────────────────┐      ┌────────────────────────────┐
│  NewProxyAuthenticator()   │      │  http.ProxyURL()           │
│  - Factory function         │      │  - Creates proxy function  │
│  - Returns interface        │      │  - URL parsing             │
└────────────┬───────────────┘      └────────────┬───────────────┘
             │                                     │
             ├───────────────┐                     │
             │               │                     │
             ▼               ▼                     ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ BasicProxyAuth   │  │ SPNEGOProxyAuth  │  │ http.Transport   │
│ - No state       │  │ - Kerberos client│  │ - Base transport │
│ - Stateless      │  │ - Network conns  │  │ - TLS config     │
└────────┬─────────┘  └────────┬─────────┘  └────────┬───────────┘
         │                     │                     │
         └─────────────────────┴─────────────────────┘
                               │
                               ▼
                    ┌──────────────────────┐
                    │ ProxyAuthTransport   │
                    │ - Wraps transport    │
                    │ - Calls authenticator│
                    └──────────────────────┘
```

### Component Responsibilities

#### NewClient()
- **Input Validation**: Validates ClientConfig fields
- **Transport Creation**: Creates base HTTP transport with proxy configuration
- **Authenticator Creation**: Delegates to NewProxyAuthenticator()
- **Transport Wrapping**: Wraps transport with ProxyAuthTransport if needed
- **Client Assembly**: Constructs and returns Client struct

#### NewProxyAuthenticator()
- **Type Dispatch**: Routes to appropriate authenticator constructor
- **Parameter Validation**: Validates required parameters for each type
- **Error Handling**: Returns descriptive errors for invalid configurations

#### ProxyAuthTransport
- **Request Interception**: Implements http.RoundTripper interface
- **Authentication Injection**: Calls Authenticator.Authenticate() before each request
- **Error Propagation**: Returns errors from authentication or transport

#### BasicProxyAuth
- **Credential Encoding**: Base64 encodes username:password
- **Header Injection**: Sets Proxy-Authorization header
- **Stateless Operation**: No per-request state required

#### SPNEGOProxyAuth
- **Kerberos Initialization**: Authenticates with KDC during construction
- **Token Generation**: Generates SPNEGO token per request
- **Resource Management**: Manages Kerberos client lifecycle

---

## Initialization Sequence

### Sequence Diagram: Client Creation with Basic Auth

```
Client Code
    │
    │ 1. Create ClientConfig
    │    {ProxyURL: "http://proxy:8080",
    │     ProxyAuthType: "basic",
    │     ProxyUsername: "user",
    │     ProxyPassword: "pass"}
    │
    ▼
NewClient(token, config, logger)
    │
    │ 2. Parse ProxyURL
    │    url.Parse("http://proxy:8080")
    │    → *url.URL{...}
    │
    │ 3. Create proxy function
    │    http.ProxyURL(parsedURL)
    │    → func(*Request) (*url.URL, error)
    │
    │ 4. Create base transport
    │    &http.Transport{
    │        Proxy: proxyFunc,
    │        TLSClientConfig: {...}
    │    }
    │
    │ 5. Check ProxyAuthType != ""
    │    → true
    │
    │ 6. Call NewProxyAuthenticator()
    │    NewProxyAuthenticator("basic", "user", "pass", "http://proxy:8080", "")
    │    │
    │    │ 6.1. Switch on "basic"
    │    │ 6.2. Validate username != "" && password != ""
    │    │ 6.3. Return NewBasicProxyAuth("user", "pass")
    │    │      → &BasicProxyAuth{username: "user", password: "pass"}
    │    │
    │    ← *BasicProxyAuth
    │
    │ 7. Wrap transport
    │    NewProxyAuthTransport(baseTransport, authenticator)
    │    │
    │    │ 7.1. Validate authenticator != nil
    │    │ 7.2. Return &ProxyAuthTransport{
    │    │         Transport: baseTransport,
    │    │         Authenticator: authenticator
    │    │       }
    │    │
    │    ← *ProxyAuthTransport
    │
    │ 8. Create HTTP client
    │    &http.Client{Transport: proxyAuthTransport}
    │
    │ 9. Return Client{HTTP: httpClient, ...}
    │
    ← Client
```

### Sequence Diagram: Client Creation with SPNEGO Auth

```
Client Code
    │
    │ 1. Create ClientConfig
    │    {ProxyURL: "http://proxy.example.com:8080",
    │     ProxyAuthType: "spnego",
    │     ProxyUsername: "user",
    │     ProxyPassword: "pass",
    │     ProxyKrb5Config: "/etc/krb5.conf"}
    │
    ▼
NewClient(token, config, logger)
    │
    │ ... (steps 2-5 same as Basic Auth) ...
    │
    │ 6. Call NewProxyAuthenticator()
    │    NewProxyAuthenticator("spnego", "user", "pass", 
    │                          "http://proxy.example.com:8080", "/etc/krb5.conf")
    │    │
    │    │ 6.1. Switch on "spnego"
    │    │ 6.2. Call NewSPNEGOProxyAuth(...)
    │    │      │
    │    │      │ 6.2.1. Validate parameters
    │    │      │ 6.2.2. Parse proxy URL
    │    │      │        url.Parse("http://proxy.example.com:8080")
    │    │      │        → Hostname: "proxy.example.com"
    │    │      │
    │    │      │ 6.2.3. Extract domain
    │    │      │        Split("proxy.example.com", ".")
    │    │      │        → ["proxy", "example", "com"]
    │    │      │        Join(["example", "com"], ".")
    │    │      │        → "example.com"
    │    │      │
    │    │      │ 6.2.4. Load Kerberos config
    │    │      │        if krb5ConfigPath != "":
    │    │      │          config.Load("/etc/krb5.conf")
    │    │      │          → *config.Config{...}
    │    │      │        else:
    │    │      │          config.New()
    │    │      │          Set DefaultRealm = "EXAMPLE.COM"
    │    │      │          Add Realm{KDC: ["example.com:88"]}
    │    │      │          → *config.Config{...}
    │    │      │
    │    │      │ 6.2.5. Create Kerberos client
    │    │      │        client.NewWithPassword(
    │    │      │          "user",
    │    │      │          "EXAMPLE.COM",
    │    │      │          "pass",
    │    │      │          krb5conf,
    │    │      │          DisablePAFXFAST(true)
    │    │      │        )
    │    │      │        → *client.Client{...}
    │    │      │
    │    │      │ 6.2.6. Authenticate with KDC
    │    │      │        kerberosClient.Login()
    │    │      │        │
    │    │      │        │ Network: TCP connection to KDC:88
    │    │      │        │ Protocol: AS-REQ / AS-REP (Kerberos)
    │    │      │        │ Result: TGT (Ticket Granting Ticket) obtained
    │    │      │        │
    │    │      │        ← nil (success) or error
    │    │      │
    │    │      │ 6.2.7. Return SPNEGOProxyAuth
    │    │      │        &SPNEGOProxyAuth{
    │    │      │          kerberosClient: kerberosClient,
    │    │      │          ...
    │    │      │        }
    │    │      │
    │    │      ← *SPNEGOProxyAuth
    │    │
    │    ← *SPNEGOProxyAuth
    │
    │ 7. Wrap transport (same as Basic Auth)
    │
    │ 8. Create HTTP client
    │
    │ 9. Return Client
    │
    ← Client
```

---

## Request Flow Algorithm

### Algorithm: HTTP Request with Proxy Authentication

```
Algorithm: RoundTrip (ProxyAuthTransport)
Input: req *http.Request
Output: *http.Response, error

BEGIN
    // Step 1: Authenticate request
    err := t.Authenticator.Authenticate(req)
    IF err != nil THEN
        RETURN nil, fmt.Errorf("failed to authenticate: %w", err)
    END IF
    
    // Step 2: Execute request through underlying transport
    resp, err := t.Transport.RoundTrip(req)
    IF err != nil THEN
        RETURN nil, err
    END IF
    
    // Step 3: Return response
    RETURN resp, nil
END
```

#### Actual Go Implementation

```go
// File: go-pivnet/proxy_auth.go
// Library: go-pivnet
// Location: /Users/hyayi/work/go-pivnet/proxy_auth.go

// ProxyAuthTransport wraps an http.RoundTripper and adds proxy authentication
// using a pluggable ProxyAuthenticator interface
type ProxyAuthTransport struct {
	Transport     http.RoundTripper
	Authenticator ProxyAuthenticator
}

// NewProxyAuthTransport creates a new ProxyAuthTransport with the given authenticator
func NewProxyAuthTransport(transport http.RoundTripper, authenticator ProxyAuthenticator) (*ProxyAuthTransport, error) {
	if authenticator == nil {
		return nil, fmt.Errorf("authenticator cannot be nil")
	}

	return &ProxyAuthTransport{
		Transport:     transport,
		Authenticator: authenticator,
	}, nil
}

// RoundTrip executes a single HTTP transaction, adding proxy authentication
func (t *ProxyAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Step 1: Add authentication to the request
	if err := t.Authenticator.Authenticate(req); err != nil {
		return nil, fmt.Errorf("failed to authenticate proxy request: %w", err)
	}

	// Step 2: Execute the request with the underlying transport
	return t.Transport.RoundTrip(req)
}

// Close cleans up the authenticator resources
func (t *ProxyAuthTransport) Close() error {
	if t.Authenticator != nil {
		return t.Authenticator.Close()
	}
	return nil
}
```

### Algorithm: Basic Authentication

#### Complete Algorithm Specification

```
Algorithm: Authenticate (BasicProxyAuth)
Input: req *http.Request (must not be nil)
Output: error (nil on success)

Preconditions:
    - req != nil
    - b.username and b.password may be empty (handled gracefully)

Postconditions:
    - req.Header["Proxy-Authorization"] contains "Basic <base64-encoded-credentials>"
    - If username and password are empty, no header is set (returns nil)

BEGIN
    // Step 1: Early return if credentials are empty
    IF b.username == "" AND b.password == "" THEN
        RETURN nil  // No authentication needed
    END IF
    
    // Step 2: Construct credentials string
    // Format: "username:password"
    credentials := b.username + ":" + b.password
    // Example: "myuser:mypass"
    // Time: O(n) where n = len(username) + len(password)
    // Space: O(n) for new string allocation
    
    // Step 3: Convert to byte slice
    credentialsBytes := []byte(credentials)
    // Time: O(n), Space: O(n) for byte slice
    
    // Step 4: Base64 encode
    // Base64 encoding: 4 output bytes per 3 input bytes
    // Output size: ceil(n * 4 / 3)
    encoded := base64.StdEncoding.EncodeToString(credentialsBytes)
    // Example: "bXl1c2VyOm15cGFzcw=="
    // Time: O(n), Space: O(ceil(n * 4 / 3))
    
    // Step 5: Construct header value
    headerValue := "Basic " + encoded
    // Time: O(m) where m = len(encoded), Space: O(m)
    
    // Step 6: Set HTTP header
    req.Header.Set("Proxy-Authorization", headerValue)
    // Time: O(1) amortized (hash map insert), Space: O(m)
    
    RETURN nil
END
```

#### Actual Go Implementation

```go
// File: go-pivnet/proxy_auth_basic.go
// Library: go-pivnet
// Location: /Users/hyayi/work/go-pivnet/proxy_auth_basic.go

// BasicProxyAuth implements HTTP Basic authentication for proxies
type BasicProxyAuth struct {
	username string
	password string
}

// NewBasicProxyAuth creates a new BasicProxyAuth authenticator
func NewBasicProxyAuth(username, password string) *BasicProxyAuth {
	return &BasicProxyAuth{
		username: username,
		password: password,
	}
}

// Authenticate adds the Proxy-Authorization header with Basic auth
func (b *BasicProxyAuth) Authenticate(req *http.Request) error {
	// Step 1: Early return if credentials are empty
	if b.username == "" && b.password == "" {
		return nil
	}

	// Step 2: Construct credentials string
	// Format: "username:password"
	auth := b.username + ":" + b.password
	
	// Step 3: Base64 encode (converts string to []byte implicitly)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	
	// Step 4: Set HTTP header
	req.Header.Set("Proxy-Authorization", "Basic "+encodedAuth)
	
	return nil
}

// Close does nothing for basic auth
func (b *BasicProxyAuth) Close() error {
	return nil
}
```

**Complexity Analysis:**

| Step | Time Complexity | Space Complexity | Notes |
|------|----------------|------------------|-------|
| Empty check | O(1) | O(1) | Constant time comparison |
| String concatenation | O(n) | O(n) | n = len(username) + len(password) |
| Byte conversion | O(n) | O(n) | UTF-8 to bytes |
| Base64 encoding | O(n) | O(ceil(4n/3)) | Standard base64 |
| Header construction | O(m) | O(m) | m = encoded length |
| Header setting | O(1) | O(m) | Hash map insert |

**Overall Complexity:**
- **Time**: O(n) where n = len(username) + len(password)
- **Space**: O(n) for intermediate buffers

**Example Execution:**

```
Input:
  username = "myuser"
  password = "mypass"
  req = &http.Request{Header: make(http.Header)}

Step 1: credentials = "myuser:mypass" (13 bytes)
Step 2: credentialsBytes = [109, 121, 117, 115, 101, 114, 58, 109, 121, 112, 97, 115, 115]
Step 3: encoded = "bXl1c2VyOm15cGFzcw==" (24 bytes)
Step 4: headerValue = "Basic bXl1c2VyOm15cGFzcw==" (30 bytes)
Step 5: req.Header["Proxy-Authorization"] = "Basic bXl1c2VyOm15cGFzcw=="

Output: nil (success)
```

**Correctness Properties:**
1. **Idempotency**: Multiple calls produce same header value
2. **Thread Safety**: No shared mutable state
3. **Encoding Correctness**: Base64 encoding is reversible
4. **Header Format**: Matches RFC 7617 specification

### Algorithm: SPNEGO Authentication

#### Complete Algorithm Specification

```
Algorithm: Authenticate (SPNEGOProxyAuth)
Input: req *http.Request (must not be nil)
Output: error (nil on success)

Preconditions:
    - req != nil
    - s.kerberosClient != nil (initialized)
    - s.kerberosClient has valid TGT (Ticket Granting Ticket)

Postconditions:
    - req.Header["Proxy-Authorization"] contains "Negotiate <base64-encoded-spnego-token>"
    - Service ticket may be cached for future use

BEGIN
    // Step 1: Validate Kerberos client state
    IF s.kerberosClient == nil THEN
        RETURN fmt.Errorf("Kerberos client not initialized")
    END IF
    // Time: O(1), Space: O(1)
    
    // Step 2: Extract proxy hostname from proxyURL
    // Parse: "http://proxy.example.com:8080" → "proxy.example.com"
    parsedURL := url.Parse(s.proxyURL)
    proxyHostname := parsedURL.Hostname()
    // Time: O(m) where m = len(proxyURL), Space: O(m)
    
    // Step 3: Construct Service Principal Name (SPN)
    // Format: "HTTP/<hostname>"
    servicePrincipal := "HTTP/" + proxyHostname
    // Example: "HTTP/proxy.example.com"
    // Time: O(p) where p = len(proxyHostname), Space: O(p)
    
    // Step 4: Create SPNEGO client context
    spnegoClient := spnego.SPNEGOClient(s.kerberosClient, servicePrincipal)
    // This creates a GSS-API context for SPNEGO negotiation
    // Time: O(1), Space: O(1) (creates context object)
    
    // Step 5: Generate SPNEGO token
    // This may use cached service ticket or request new one from KDC
    token, err := spnegoClient.InitSecContext()
    IF err != nil THEN
        RETURN fmt.Errorf("failed to generate SPNEGO token: %w", err)
    END IF
    // Time: O(1) if cached, O(k) if KDC request needed (k = network RTT)
    // Space: O(t) where t = token size (~1-4 KB)
    
    // Step 6: Marshal token to ASN.1 DER format
    // ASN.1 (Abstract Syntax Notation One) DER (Distinguished Encoding Rules)
    tokenBytes, err := token.Marshal()
    IF err != nil THEN
        RETURN fmt.Errorf("failed to marshal SPNEGO token: %w", err)
    END IF
    // Time: O(t) where t = token size, Space: O(t) for DER encoding
    // Typical size: 1-4 KB
    
    // Step 7: Base64 encode DER bytes
    encodedToken := base64.StdEncoding.EncodeToString(tokenBytes)
    // Time: O(t), Space: O(ceil(4t/3)) for base64 output
    // Example: 2 KB token → ~2.7 KB base64 string
    
    // Step 8: Construct header value
    headerValue := "Negotiate " + encodedToken
    // Time: O(e) where e = len(encodedToken), Space: O(e)
    
    // Step 9: Set HTTP header
    req.Header.Set("Proxy-Authorization", headerValue)
    // Time: O(1) amortized, Space: O(e)
    
    RETURN nil
END
```

#### Actual Go Implementation

```go
// File: go-pivnet/proxy_auth_spnego.go
// Library: go-pivnet
// Location: /Users/hyayi/work/go-pivnet/proxy_auth_spnego.go

// SPNEGOProxyAuth implements Kerberos/SPNEGO authentication for proxies
type SPNEGOProxyAuth struct {
	username       string
	password       string
	proxyURL       string
	kerberosClient *client.Client
}

// Authenticate adds the Proxy-Authorization header with SPNEGO token
func (s *SPNEGOProxyAuth) Authenticate(req *http.Request) error {
	// Step 1: Validate Kerberos client state
	if s.kerberosClient == nil {
		return fmt.Errorf("Kerberos client not initialized")
	}

	// Step 2-4: Generate SPNEGO token for the proxy
	// The service principal name (SPN) for HTTP proxy is typically HTTP/<proxy-host>
	spnegoClient := spnego.SPNEGOClient(s.kerberosClient, "HTTP")

	// Step 5: Get the SPNEGO token
	token, err := spnegoClient.InitSecContext()
	if err != nil {
		return fmt.Errorf("failed to generate SPNEGO token: %w", err)
	}

	// Step 6: Marshal the token to bytes and encode as base64
	tokenBytes, err := token.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal SPNEGO token: %w", err)
	}
	encodedToken := base64.StdEncoding.EncodeToString(tokenBytes)

	// Step 7-8: Add the Proxy-Authorization header with Negotiate scheme
	req.Header.Set("Proxy-Authorization", "Negotiate "+encodedToken)

	return nil
}

// Close cleans up the Kerberos client resources
func (s *SPNEGOProxyAuth) Close() error {
	if s.kerberosClient != nil {
		s.kerberosClient.Destroy()
	}
	return nil
}
```

**Complexity Analysis:**

| Step | Time Complexity | Space Complexity | Notes |
|------|----------------|------------------|-------|
| Client validation | O(1) | O(1) | Nil check |
| URL parsing | O(m) | O(m) | m = len(proxyURL) |
| SPN construction | O(p) | O(p) | p = len(hostname) |
| SPNEGO client creation | O(1) | O(1) | Context object |
| Token generation | O(1) or O(k) | O(t) | O(1) if cached, O(k) if KDC request |
| Token marshaling | O(t) | O(t) | t = token size (~1-4 KB) |
| Base64 encoding | O(t) | O(ceil(4t/3)) | Standard base64 |
| Header construction | O(e) | O(e) | e = encoded length |
| Header setting | O(1) | O(e) | Hash map insert |

**Overall Complexity:**
- **Time**: O(t) where t = token size, or O(k) if KDC request needed
- **Space**: O(t) for token and encoding buffers

**Token Generation Details:**

```
SPNEGO Token Structure (ASN.1 DER):

GSS-API Initial Context Token {
    thisMech: OID (1.3.6.1.5.5.2)  // SPNEGO OID
    innerContextToken: NegTokenInit {
        mechTypes: [OID]             // Supported mechanisms
        mechToken: OCTET STRING {    // Kerberos token
            AP-REQ {
                pvno: INTEGER        // Protocol version
                msg-type: INTEGER    // Message type
                ap-options: BIT STRING
                ticket: Ticket {
                    tkt-vno: INTEGER
                    realm: Realm
                    sname: PrincipalName
                    enc-part: EncryptedData
                }
                authenticator: EncryptedData
            }
        }
        mechListMIC: OCTET STRING    // Optional
    }
}

Typical Sizes:
- Minimal token: ~500 bytes
- Typical token: 1-2 KB
- Large token (with MIC): 3-4 KB
```

**Example Execution:**

```
Input:
  proxyURL = "http://proxy.example.com:8080"
  kerberosClient = <initialized with TGT>
  req = &http.Request{Header: make(http.Header)}

Step 1: Validation passed
Step 2: proxyHostname = "proxy.example.com"
Step 3: servicePrincipal = "HTTP/proxy.example.com"
Step 4: spnegoClient created
Step 5: token = <SPNEGO token object> (uses cached service ticket)
Step 6: tokenBytes = [DER encoded bytes, ~1500 bytes]
Step 7: encodedToken = "YIIBhwYGKwYBBQUCoIIBezCCAXegGDAWBgkqhkiG9w0BBQ..." (~2000 chars)
Step 8: headerValue = "Negotiate YIIBhwYGKwYBBQUCoIIBezCCAXegGDAWBgkqhkiG9w0BBQ..."
Step 9: req.Header["Proxy-Authorization"] = headerValue

Output: nil (success)
```

**Caching Strategy:**

```
Service Ticket Cache (managed by gokrb5 library):

Cache Key: Service Principal Name (SPN)
  Example: "HTTP/proxy.example.com@EXAMPLE.COM"

Cache Entry:
  - Service Ticket: Encrypted ticket from KDC
  - Expiration Time: Ticket lifetime (typically 8-10 hours)
  - Session Key: Encryption key for ticket

Cache Lookup:
  IF cache contains valid ticket for SPN THEN
      Use cached ticket  // O(1)
  ELSE
      Request new ticket from KDC using TGT  // O(k) network RTT
      Cache new ticket
  END IF
```

**Correctness Properties:**
1. **Token Validity**: Token is valid for proxy authentication
2. **Format Compliance**: Matches RFC 4178 (SPNEGO) and RFC 4559 (HTTP Negotiate)
3. **Security**: Token contains encrypted service ticket
4. **Idempotency**: Multiple calls generate new tokens (may differ due to nonces)

### Detailed SPNEGO Token Generation

```
SPNEGO Token Generation Process:

1. Extract Service Principal Name (SPN)
   SPN = "HTTP/" + proxyHostname
   Example: "HTTP/proxy.example.com"

2. Request Service Ticket (if not cached)
   IF service ticket not in cache THEN
       a. Use TGT to request service ticket from KDC
       b. AP-REQ / AP-REP exchange
       c. Cache service ticket
   END IF

3. Create GSS-API Context Token
   a. Wrap service ticket in GSS-API structure
   b. Apply encryption/checksum as per Kerberos protocol
   c. Encode as ASN.1 DER

4. Base64 Encode
   Convert binary DER to base64 string
```

---

## State Management

### State Diagram: SPNEGOProxyAuth Lifecycle

```
[Uninitialized]
    │
    │ NewSPNEGOProxyAuth()
    │
    ▼
[Initializing]
    │
    │ ├─→ Parse URL
    │ ├─→ Load Kerberos Config
    │ ├─→ Create Kerberos Client
    │ └─→ Login to KDC
    │
    ├─→ [Error] ─→ [Failed]
    │
    └─→ [Initialized]
            │
            │ Authenticate() called
            │
            ├─→ [Generating Token]
            │       │
            │       ├─→ Generate SPNEGO token
            │       │   (may use cached service ticket)
            │       │
            │       └─→ [Token Generated]
            │               │
            │               └─→ Add header → [Ready]
            │
            └─→ [Ready]
                    │
                    │ Close() called
                    │
                    └─→ [Closing]
                            │
                            │ Destroy Kerberos client
                            │ Close network connections
                            │
                            └─→ [Closed]
```

### State Transitions

| State | Entry Condition | Exit Condition | Operations |
|-------|----------------|----------------|------------|
| Uninitialized | Object created | NewSPNEGOProxyAuth() called | None |
| Initializing | NewSPNEGOProxyAuth() | Login() completes | URL parsing, config loading, KDC authentication |
| Initialized | Login() succeeds | Authenticate() or Close() | None (idle) |
| Generating Token | Authenticate() called | Token generated | SPNEGO token creation |
| Ready | Header added | Next request or Close() | Request sent |
| Closing | Close() called | Resources released | Client destruction |
| Closed | Resources released | N/A (final state) | None |
| Failed | Error during init | Close() or retry | Error state |

### Caching Strategy

**SPNEGO Token Caching:**
- Service tickets cached by Kerberos client (gokrb5 library)
- Cache key: Service Principal Name (SPN)
- Cache lifetime: Determined by ticket expiration time
- Cache invalidation: On ticket expiration or explicit refresh

**Basic Auth:**
- No caching required (stateless)
- Credentials stored in memory (not cached per request)

---

## Error Handling

### Error Propagation Chain

```
Request Error Flow:

HTTP Request
    │
    ▼
ProxyAuthTransport.RoundTrip()
    │
    ├─→ Authenticator.Authenticate()
    │       │
    │       ├─→ [BasicAuth] Base64 encoding error
    │       │       └─→ Return error → RoundTrip fails
    │       │
    │       └─→ [SPNEGO] Token generation error
    │               ├─→ Kerberos client error
    │               ├─→ Token marshaling error
    │               └─→ Return error → RoundTrip fails
    │
    └─→ Transport.RoundTrip()
            │
            ├─→ Network error
            ├─→ Proxy error (407 if auth fails)
            └─→ Server error
```

### Error Types and Handling

#### Initialization Errors

```go
// URL Parsing Error
err := url.Parse(proxyURL)
if err != nil {
    logger.Info("Failed to parse proxy URL, using environment proxy", 
                logger.Data{"error": err.Error()})
    // Fallback: Use http.ProxyFromEnvironment
}

// Authenticator Creation Error
authenticator, err := NewProxyAuthenticator(...)
if err != nil {
    logger.Info("Failed to create proxy authenticator", 
                logger.Data{"error": err.Error()})
    // Fallback: Continue without authentication
    httpTransport = baseTransport
}
```

#### Runtime Errors

```go
// Authentication Error (in RoundTrip)
if err := t.Authenticator.Authenticate(req); err != nil {
    return nil, fmt.Errorf("failed to authenticate proxy request: %w", err)
    // Error propagated to caller, request not sent
}

// Transport Error
resp, err := t.Transport.RoundTrip(req)
if err != nil {
    return nil, err
    // Network/proxy/server error returned to caller
}
```

#### SPNEGO-Specific Errors

```go
// Kerberos Login Error
err = kerberosClient.Login()
if err != nil {
    return nil, fmt.Errorf("failed to login to Kerberos: %w", err)
    // Initialization fails, client not created
}

// Token Generation Error
token, err := spnegoClient.InitSecContext()
if err != nil {
    return fmt.Errorf("failed to generate SPNEGO token: %w", err)
    // Request fails, may indicate:
    // - Expired TGT (need re-authentication)
    // - Network issue with KDC
    // - Invalid service principal
}
```

### Error Recovery Strategies

1. **Graceful Degradation**: If proxy auth setup fails, fall back to base transport
2. **Logging**: All errors logged with context for debugging
3. **No Retry**: Authentication errors are not retried automatically (caller responsibility)
4. **Resource Cleanup**: Errors during initialization trigger cleanup

---

## Resource Lifecycle

### Resource Acquisition and Release

#### BasicProxyAuth

```
Lifecycle:
    Creation:  O(1) - String assignments
    Usage:     O(1) per request - Header setting
    Cleanup:  O(1) - No resources to release
```

**Memory:**
- Allocated: At construction time
- Released: When object is garbage collected
- No explicit cleanup required

#### SPNEGOProxyAuth

```
Lifecycle:
    Creation:  O(n) - Network I/O to KDC
    Usage:     O(1) per request - Token generation (may use cache)
    Cleanup:   O(1) - Client destruction
```

**Resources:**
- Kerberos client: Network connections, cryptographic state
- Service tickets: Cached in memory
- TGT: Cached in memory

**Cleanup Sequence:**
```go
func (s *SPNEGOProxyAuth) Close() error {
    if s.kerberosClient != nil {
        s.kerberosClient.Destroy()  // Closes connections, clears cache
        s.kerberosClient = nil      // Prevent use-after-close
    }
    return nil
}
```

### Resource Leak Prevention

1. **Explicit Close**: SPNEGOProxyAuth must be closed
2. **Nil Checks**: All methods check for nil before use
3. **Idempotent Close**: Safe to call Close() multiple times
4. **Finalizer**: Not used (rely on explicit cleanup)

### Memory Management

**Heap Allocations:**
- ClientConfig: All strings heap-allocated
- ProxyAuthTransport: 2 pointers (16 bytes)
- BasicProxyAuth: 2 string pointers (16 bytes)
- SPNEGOProxyAuth: 4 pointers + Kerberos client (variable)

**Stack Allocations:**
- Function parameters
- Local variables
- Temporary buffers (base64 encoding)

**Garbage Collection:**
- All authenticators are GC-eligible when no longer referenced
- Kerberos client resources released on Close()
- No finalizers required

---

## Thread Safety

### Concurrent Access Analysis

#### ProxyAuthTransport

```go
type ProxyAuthTransport struct {
    Transport     http.RoundTripper  // Read-only after construction
    Authenticator ProxyAuthenticator  // Must be thread-safe
}
```

**Thread Safety:**
- ✅ **Safe**: `RoundTrip()` can be called concurrently
- ✅ **Safe**: `Transport` is read-only
- ⚠️ **Requires**: `Authenticator.Authenticate()` must be thread-safe

#### BasicProxyAuth

```go
type BasicProxyAuth struct {
    username string  // Immutable after construction
    password string  // Immutable after construction
}
```

**Thread Safety:**
- ✅ **Safe**: All fields immutable
- ✅ **Safe**: `Authenticate()` is stateless
- ✅ **Safe**: No shared mutable state
- ✅ **Safe**: Concurrent calls to `Authenticate()` are safe

#### SPNEGOProxyAuth

```go
type SPNEGOProxyAuth struct {
    username       string
    password       string
    proxyURL       string
    kerberosClient *client.Client  // Shared mutable state
}
```

**Thread Safety:**
- ✅ **Safe**: String fields immutable
- ⚠️ **Requires**: `kerberosClient` must be thread-safe (gokrb5 library)
- ✅ **Safe**: `Authenticate()` uses client in read-only manner (token generation)
- ⚠️ **Unsafe**: `Close()` must not be called concurrently with `Authenticate()`

### Synchronization Requirements

**No Explicit Locks Required:**
- BasicProxyAuth: Stateless, no synchronization needed
- SPNEGOProxyAuth: Relies on gokrb5 library's thread safety
- ProxyAuthTransport: Read-only fields, authenticator must be thread-safe

**Critical Sections:**
- None (delegated to underlying libraries)

**Race Conditions:**
- Potential: `Close()` called during `Authenticate()`
- Mitigation: Application-level coordination required
- Recommendation: Close only after all requests complete

---

## Performance Considerations

### MVP Scope: Caching Limitations

**Important Note**: Explicit caching strategies are **out of scope for this MVP**. The current implementation:

- **No Application-Level Token Caching**: Authentication tokens (SPNEGO or Basic) are generated per-request without explicit caching
- **No Credential Caching**: Credentials are not cached or stored between requests
- **Library-Managed Caching Only**: The `gokrb5` library used for SPNEGO authentication may cache Kerberos tickets internally (TGT, service tickets), but this is managed by the library, not by our application code
- **No Persistent State**: No caching of authentication state across application sessions

**Performance Impact**:
- Basic authentication: Minimal impact (fast Base64 encoding per request)
- SPNEGO authentication: May have performance impact due to token generation per request, though the underlying Kerberos library may cache tickets internally

**Future Enhancements** (Post-MVP):
- Application-level token caching to reduce authentication overhead
- Intelligent token reuse strategies
- Secure credential storage and caching

### Time Complexity Analysis

| Operation | Basic Auth | SPNEGO Auth |
|-----------|-----------|-------------|
| Initialization | O(1) | O(n) - Network I/O |
| Per-Request Auth | O(n) - Base64 | O(1) - Token generation (library may cache tickets) |
| Token Generation | N/A | O(1) - If TGT cached by library, O(k) if KDC request needed |
| Header Setting | O(1) | O(1) |

Where:
- n = credential/token size
- k = network round-trip time to KDC

**Note**: SPNEGO token generation complexity assumes the underlying `gokrb5` library may cache Kerberos tickets internally. This is library-managed caching, not application-level caching. Application-level token caching is out of scope for this MVP.

### Space Complexity

| Component | Space | Notes |
|-----------|-------|-------|
| BasicProxyAuth | O(1) | Fixed size struct |
| SPNEGOProxyAuth | O(1) + O(m) | Struct + Kerberos state |
| Token Generation | O(k) | k = token size (~1-4KB) |
| Base64 Buffer | O(n) | n = input size |

### Performance Optimizations

1. **Token Caching**: SPNEGO tokens cached by Kerberos client
2. **Lazy Initialization**: Authenticator created only when needed
3. **Zero-Copy**: Header setting uses string references
4. **Connection Reuse**: HTTP transport reuses connections

### Benchmark Estimates

**Basic Auth (per request):**
- Base64 encoding: ~1-10μs (depends on credential length)
- Header setting: <1μs
- **Total overhead: ~1-10μs per request**

**SPNEGO Auth (per request):**
- Token generation (cached): ~10-100μs
- Token marshaling: ~10-50μs
- Base64 encoding: ~10-50μs
- **Total overhead: ~30-200μs per request**

**Initialization:**
- Basic: <1μs
- SPNEGO: 100-1000ms (network I/O to KDC)

---

## Additional Algorithms

### Algorithm: NewProxyAuthenticator (Factory Function)

```
Algorithm: NewProxyAuthenticator
Input: 
    authType: ProxyAuthType
    username: string
    password: string
    proxyURL: string
    krb5ConfigPath: string
Output: ProxyAuthenticator, error

Preconditions:
    - authType is valid (basic, spnego, or empty)
    - If authType is basic: username and password required
    - If authType is spnego: username, password, and proxyURL required

Postconditions:
    - Returns valid ProxyAuthenticator implementation
    - Returns error if validation fails

BEGIN
    // Step 1: Dispatch based on authentication type
    SWITCH authType DO
        CASE ProxyAuthTypeBasic:
            // Step 1.1: Validate Basic auth parameters
            IF username == "" OR password == "" THEN
                RETURN nil, fmt.Errorf("username and password are required for basic authentication")
            END IF
            // Time: O(1), Space: O(1)
            
            // Step 1.2: Create Basic authenticator
            authenticator := NewBasicProxyAuth(username, password)
            // Time: O(1), Space: O(1) + string data
            // Returns: *BasicProxyAuth
            
            RETURN authenticator, nil
            
        CASE ProxyAuthTypeSPNEGO:
            // Step 2.1: Validate SPNEGO auth parameters
            IF username == "" OR password == "" OR proxyURL == "" THEN
                RETURN nil, fmt.Errorf("username, password, and proxyURL are required for SPNEGO authentication")
            END IF
            // Time: O(1), Space: O(1)
            
            // Step 2.2: Create SPNEGO authenticator
            authenticator, err := NewSPNEGOProxyAuth(username, password, proxyURL, krb5ConfigPath)
            IF err != nil THEN
                RETURN nil, fmt.Errorf("failed to create SPNEGO authenticator: %w", err)
            END IF
            // Time: O(k) where k = KDC authentication time, Space: O(s) where s = Kerberos state
            // Returns: *SPNEGOProxyAuth
            
            RETURN authenticator, nil
            
        DEFAULT:
            // Step 3: Unsupported authentication type
            RETURN nil, fmt.Errorf("unsupported proxy authentication type: %s", authType)
    END SWITCH
END
```

#### Actual Go Implementation

```go
// File: go-pivnet/proxy_authenticator.go
// Library: go-pivnet
// Location: /Users/hyayi/work/go-pivnet/proxy_authenticator.go

// NewProxyAuthenticator creates a ProxyAuthenticator based on the auth type
func NewProxyAuthenticator(authType ProxyAuthType, username, password, proxyURL, krb5ConfigPath string) (ProxyAuthenticator, error) {
	switch authType {
	case ProxyAuthTypeBasic:
		// Step 1.1: Validate Basic auth parameters
		if username == "" || password == "" {
			return nil, fmt.Errorf("username and password are required for basic authentication")
		}
		
		// Step 1.2: Create Basic authenticator
		return NewBasicProxyAuth(username, password), nil

	case ProxyAuthTypeSPNEGO:
		// Step 2: Create SPNEGO authenticator (validation happens inside)
		return NewSPNEGOProxyAuth(username, password, proxyURL, krb5ConfigPath)

	default:
		// Step 3: Unsupported authentication type
		return nil, fmt.Errorf("unsupported proxy authentication type: %s", authType)
	}
}
```

**Complexity Analysis:**

| Auth Type | Time Complexity | Space Complexity | Notes |
|-----------|----------------|------------------|-------|
| Basic | O(1) | O(1) + string data | Simple struct creation |
| SPNEGO | O(k) | O(s) | k = KDC auth time, s = Kerberos state size |
| Invalid | O(1) | O(1) | Error return |

### Algorithm: NewSPNEGOProxyAuth (SPNEGO Initialization)

```
Algorithm: NewSPNEGOProxyAuth
Input:
    username: string
    password: string
    proxyURL: string
    krb5ConfigPath: string
Output: *SPNEGOProxyAuth, error

Preconditions:
    - username != "" AND password != "" AND proxyURL != ""
    - krb5ConfigPath may be empty (uses auto-generated config)

Postconditions:
    - Returns initialized SPNEGOProxyAuth with authenticated Kerberos client
    - Kerberos client has valid TGT (Ticket Granting Ticket)

BEGIN
    // Step 1: Parse proxy URL
    parsedURL, err := url.Parse(proxyURL)
    IF err != nil THEN
        RETURN nil, fmt.Errorf("failed to parse proxy URL: %w", err)
    END IF
    // Time: O(m) where m = len(proxyURL), Space: O(m)
    
    // Step 2: Extract hostname
    proxyHostname := parsedURL.Hostname()
    IF proxyHostname == "" THEN
        RETURN nil, fmt.Errorf("invalid proxy URL: no host found")
    END IF
    // Time: O(1), Space: O(1)
    
    // Step 3: Extract domain from hostname
    // Algorithm: Split by ".", take last two parts
    domainParts := strings.Split(proxyHostname, ".")
    // Time: O(h) where h = len(proxyHostname), Space: O(h)
    
    domain := ""
    IF len(domainParts) >= 2 THEN
        // Take last two parts: "example.com"
        domain = strings.Join(domainParts[len(domainParts)-2:], ".")
    ELSE
        // Single part: use as-is
        domain = proxyHostname
    END IF
    // Time: O(d) where d = len(domain), Space: O(d)
    // Example: "proxy.example.com" → "example.com"
    
    // Step 4: Load or create Kerberos configuration
    var krb5conf *config.Config
    IF krb5ConfigPath != "" THEN
        // Load from file
        krb5conf, err = config.Load(krb5ConfigPath)
        IF err != nil THEN
            RETURN nil, fmt.Errorf("failed to load Kerberos config from %s: %w", krb5ConfigPath, err)
        END IF
        // Time: O(f) where f = file size, Space: O(f)
    ELSE
        // Create default configuration
        krb5conf = config.New()
        // Time: O(1), Space: O(1)
        
        // Set default realm (uppercase domain)
        krb5conf.LibDefaults.DefaultRealm = strings.ToUpper(domain)
        // Example: "example.com" → "EXAMPLE.COM"
        // Time: O(d), Space: O(d)
        
        // Add realm configuration with KDC
        krb5conf.Realms = []config.Realm{
            {
                Realm: strings.ToUpper(domain),
                KDC:   []string{domain + ":88"},  // Default Kerberos port
            },
        }
        // Time: O(d), Space: O(d)
        
        // Add domain to realm mapping
        krb5conf.DomainRealm = config.DomainRealm{
            domain: strings.ToUpper(domain),
        }
        // Time: O(d), Space: O(d)
    END IF
    
    // Step 5: Create Kerberos client
    kerberosClient := client.NewWithPassword(
        username,                    // Principal name
        strings.ToUpper(domain),     // Realm
        password,                    // Password
        krb5conf,                    // Configuration
        client.DisablePAFXFAST(true), // Disable pre-authentication
    )
    // Time: O(1), Space: O(1) (client object creation)
    
    // Step 6: Authenticate with Kerberos KDC
    err = kerberosClient.Login()
    IF err != nil THEN
        RETURN nil, fmt.Errorf("failed to login to Kerberos: %w", err)
    END IF
    // Time: O(k) where k = network RTT + KDC processing time
    // Space: O(t) where t = TGT size (~1-4 KB)
    // Network: TCP connection to KDC:88, AS-REQ/AS-REP exchange
    
    // Step 7: Create and return SPNEGO authenticator
    RETURN &SPNEGOProxyAuth{
        username:       username,
        password:       password,
        proxyURL:       proxyURL,
        kerberosClient: kerberosClient,
    }, nil
    // Time: O(1), Space: O(1) + struct size
END
```

#### Actual Go Implementation

```go
// File: go-pivnet/proxy_auth_spnego.go

// NewSPNEGOProxyAuth creates a new SPNEGOProxyAuth authenticator
func NewSPNEGOProxyAuth(username, password, proxyURL, krb5ConfigPath string) (*SPNEGOProxyAuth, error) {
	// Step 1: Validate parameters
	if username == "" || password == "" || proxyURL == "" {
		return nil, fmt.Errorf("username, password, and proxyURL are required for SPNEGO authentication")
	}

	// Step 2: Parse the proxy URL to extract the domain/host
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
	}

	// Step 3: Extract domain from the proxy host
	proxyHost := parsedURL.Hostname()
	if proxyHost == "" {
		return nil, fmt.Errorf("invalid proxy URL: no host found")
	}

	// Step 4: Extract domain from hostname (e.g., proxy.example.com -> example.com)
	// For Kerberos, we need the realm which is typically the domain in uppercase
	domainParts := strings.Split(proxyHost, ".")
	var domain string
	if len(domainParts) >= 2 {
		// Take the last two parts as the domain (e.g., example.com)
		domain = strings.Join(domainParts[len(domainParts)-2:], ".")
	} else {
		domain = proxyHost
	}

	// Step 5: Create or load Kerberos configuration
	var krb5conf *config.Config
	if krb5ConfigPath != "" {
		// Load configuration from file
		krb5conf, err = config.Load(krb5ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load Kerberos config from %s: %w", krb5ConfigPath, err)
		}
	} else {
		// Create default configuration
		krb5conf = config.New()
		// Set the default realm to the proxy domain
		krb5conf.LibDefaults.DefaultRealm = strings.ToUpper(domain)
		
		// Add the realm configuration
		krb5conf.Realms = []config.Realm{
			{
				Realm: strings.ToUpper(domain),
				KDC:   []string{domain + ":88"},
			},
		}
		
		// Add domain to realm mapping
		krb5conf.DomainRealm = config.DomainRealm{
			domain: strings.ToUpper(domain),
		}
	}

	// Step 6: Create Kerberos client with username and password
	kerberosClient := client.NewWithPassword(
		username,
		strings.ToUpper(domain),
		password,
		krb5conf,
		client.DisablePAFXFAST(true),
	)

	// Step 7: Login to Kerberos
	err = kerberosClient.Login()
	if err != nil {
		return nil, fmt.Errorf("failed to login to Kerberos: %w", err)
	}

	// Step 8: Return initialized authenticator
	return &SPNEGOProxyAuth{
		username:       username,
		password:       password,
		proxyURL:       proxyURL,
		kerberosClient: kerberosClient,
	}, nil
}
```

**Complexity Analysis:**

| Step | Time Complexity | Space Complexity | Network I/O |
|------|----------------|------------------|-------------|
| URL parsing | O(m) | O(m) | None |
| Hostname extraction | O(1) | O(1) | None |
| Domain extraction | O(h) | O(h) | None |
| Config loading | O(f) or O(1) | O(f) or O(d) | File I/O (if file) |
| Client creation | O(1) | O(1) | None |
| KDC authentication | O(k) | O(t) | TCP to KDC:88 |
| Struct creation | O(1) | O(1) | None |

**Overall Complexity:**
- **Time**: O(k) where k = KDC authentication time (typically 100-1000ms)
- **Space**: O(s) where s = Kerberos state size (~5-20 KB)

**KDC Authentication Details:**

```
AS-REQ (Authentication Service Request) Flow:

1. Client → KDC: AS-REQ
   - Principal: username@REALM
   - Pre-authentication: PA-ENC-TIMESTAMP (encrypted timestamp)
   - Encryption: Based on password-derived key

2. KDC → Client: AS-REP
   - TGT (Ticket Granting Ticket): Encrypted with KDC secret
   - Session key: For future TGS requests
   - Lifetime: Typically 8-10 hours

Network Protocol:
  - Transport: TCP
  - Port: 88 (Kerberos)
  - Protocol: ASN.1 DER over TCP
  - Encryption: Kerberos encryption (AES, DES, etc.)
```

### Algorithm: NewClient (Client Creation with Proxy Auth)

```
Algorithm: NewClient (with proxy authentication support)
Input:
    token: AccessTokenService
    config: ClientConfig
    logger: logger.Logger
Output: Client

Preconditions:
    - token != nil
    - config.Host != ""
    - logger != nil

Postconditions:
    - Returns initialized Client with proxy authentication if configured
    - HTTP client uses ProxyAuthTransport if ProxyAuthType is set

BEGIN
    // Step 1: Construct base URL
    baseURL := fmt.Sprintf("%s%s", config.Host, apiVersion)
    // Time: O(h) where h = len(config.Host), Space: O(h)
    
    // Step 2: Create proxy function
    proxyFunc := http.ProxyFromEnvironment  // Default
    IF config.ProxyURL != "" THEN
        parsedURL, err := url.Parse(config.ProxyURL)
        IF err != nil THEN
            logger.Info("Failed to parse proxy URL, using environment proxy", 
                        logger.Data{"error": err.Error()})
        ELSE
            proxyFunc = http.ProxyURL(parsedURL)
            // Returns function: func(*Request) (*url.URL, error)
        END IF
    END IF
    // Time: O(m) where m = len(config.ProxyURL), Space: O(m)
    
    // Step 3: Create base HTTP transport
    baseTransport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: config.SkipSSLValidation,
        },
        Proxy: proxyFunc,
    }
    // Time: O(1), Space: O(1) + transport state
    
    // Step 4: Wrap with proxy authentication if configured
    httpTransport := baseTransport  // Default: no authentication
    IF config.ProxyAuthType != "" THEN
        // Step 4.1: Create authenticator
        authenticator, err := NewProxyAuthenticator(
            config.ProxyAuthType,
            config.ProxyUsername,
            config.ProxyPassword,
            config.ProxyURL,
            config.ProxyKrb5Config,
        )
        IF err != nil THEN
            logger.Info("Failed to create proxy authenticator", 
                        logger.Data{"error": err.Error()})
            // Continue without authentication
        ELSE
            // Step 4.2: Wrap transport with authentication
            proxyAuthTransport, err := NewProxyAuthTransport(baseTransport, authenticator)
            IF err != nil THEN
                logger.Info("Failed to initialize proxy authentication", 
                            logger.Data{"error": err.Error()})
                // Continue without authentication
            ELSE
                httpTransport = proxyAuthTransport
            END IF
        END IF
    END IF
    // Time: O(a) where a = authenticator creation time, Space: O(a)
    
    // Step 5: Create HTTP client for API requests
    httpClient := &http.Client{
        Timeout:   10 * time.Minute,
        Transport: httpTransport,
    }
    // Time: O(1), Space: O(1)
    
    // Step 6: Create download transport (same proxy configuration)
    downloadBaseTransport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: config.SkipSSLValidation,
        },
        Proxy: proxyFunc,
    }
    // Time: O(1), Space: O(1)
    
    // Step 7: Wrap download transport with authentication
    downloadTransport := downloadBaseTransport
    IF config.ProxyAuthType != "" THEN
        // Create separate authenticator for download client
        // (may be shared or separate instance)
        downloadAuthenticator, err := NewProxyAuthenticator(
            config.ProxyAuthType,
            config.ProxyUsername,
            config.ProxyPassword,
            config.ProxyURL,
            config.ProxyKrb5Config,
        )
        IF err == nil THEN
            downloadProxyAuthTransport, err := NewProxyAuthTransport(
                downloadBaseTransport, 
                downloadAuthenticator,
            )
            IF err == nil THEN
                downloadTransport = downloadProxyAuthTransport
            END IF
        END IF
    END IF
    
    // Step 8: Create download client
    downloadClient := &http.Client{
        Timeout:   0,  // No timeout for downloads
        Transport: downloadTransport,
    }
    // Time: O(1), Space: O(1)
    
    // Step 9: Create downloader with ranger
    ranger := download.NewRanger(concurrentDownloads)
    downloader := download.Client{
        HTTPClient: downloadClient,
        Ranger:     ranger,
        Logger:     logger,
        Timeout:    30 * time.Second,
    }
    // Time: O(1), Space: O(1)
    
    // Step 10: Create and return Client
    client := Client{
        baseURL:    baseURL,
        token:      token,
        userAgent:  config.UserAgent,
        logger:     logger,
        downloader: downloader,
        HTTP:       httpClient,
        // ... initialize service fields ...
    }
    // Time: O(1), Space: O(1) + client state
    
    RETURN client
END
```

#### Actual Go Implementation

```go
// File: go-pivnet/pivnet.go
// Library: go-pivnet
// Location: /Users/hyayi/work/go-pivnet/pivnet.go

func NewClient(
	token AccessTokenService,
	config ClientConfig,  // ← Receives config from om-cli via download_clients layer
	logger logger.Logger,
) Client {
	// Step 1: Construct base URL
	baseURL := fmt.Sprintf("%s%s", config.Host, apiVersion)

	// Step 2: Create proxy function based on ProxyURL or environment
	proxyFunc := http.ProxyFromEnvironment
	if config.ProxyURL != "" {
		proxyURL, err := url.Parse(config.ProxyURL)
		if err != nil {
			logger.Info("Failed to parse proxy URL, using environment proxy", logger.Data{"error": err.Error()})
		} else {
			proxyFunc = http.ProxyURL(proxyURL)
		}
	}

	// Step 3: Create base transport
	baseTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.SkipSSLValidation,
		},
		Proxy: proxyFunc,
	}

	// Step 4: Wrap with proxy authentication if configured
	var httpTransport http.RoundTripper = baseTransport
	if config.ProxyAuthType != "" {
		authenticator, err := NewProxyAuthenticator(
			config.ProxyAuthType,
			config.ProxyUsername,
			config.ProxyPassword,
			config.ProxyURL,
			config.ProxyKrb5Config,
		)
		if err != nil {
			logger.Info("Failed to create proxy authenticator", logger.Data{"error": err.Error()})
		} else {
			proxyAuthTransport, err := NewProxyAuthTransport(baseTransport, authenticator)
			if err != nil {
				logger.Info("Failed to initialize proxy authentication", logger.Data{"error": err.Error()})
			} else {
				httpTransport = proxyAuthTransport
			}
		}
	}

	// Step 5: Create HTTP client
	httpClient := &http.Client{
		Timeout:   10 * time.Minute,
		Transport: httpTransport,
	}

	// Step 6: Create download transport with same proxy configuration
	downloadBaseTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.SkipSSLValidation,
		},
		Proxy: proxyFunc,
	}

	// Step 7: Wrap download transport with proxy authentication if configured
	var downloadTransport http.RoundTripper = downloadBaseTransport
	if config.ProxyAuthType != "" {
		authenticator, err := NewProxyAuthenticator(
			config.ProxyAuthType,
			config.ProxyUsername,
			config.ProxyPassword,
			config.ProxyURL,
			config.ProxyKrb5Config,
		)
		if err != nil {
			logger.Info("Failed to create proxy authenticator for download client", logger.Data{"error": err.Error()})
		} else {
			proxyAuthTransport, err := NewProxyAuthTransport(downloadBaseTransport, authenticator)
			if err != nil {
				logger.Info("Failed to initialize proxy authentication for download client", logger.Data{"error": err.Error()})
			} else {
				downloadTransport = proxyAuthTransport
			}
		}
	}

	// Step 8: Create download client
	downloadClient := &http.Client{
		Timeout:   0,
		Transport: downloadTransport,
	}

	// Step 9: Create downloader
	ranger := download.NewRanger(concurrentDownloads)
	downloader := download.Client{
		HTTPClient: downloadClient,
		Ranger:     ranger,
		Logger:     logger,
		Timeout:    30 * time.Second,
	}

	// Step 10: Create and return Client
	client := Client{
		baseURL:    baseURL,
		token:      token,
		userAgent:  config.UserAgent,
		logger:     logger,
		downloader: downloader,
		HTTP:       httpClient,
	}

	// Initialize service fields
	client.Auth = &AuthService{client: client}
	client.EULA = &EULAsService{client: client}
	client.ProductFiles = &ProductFilesService{client: client}
	// ... other service initializations ...

	return client
}
```

**Complexity Analysis:**

| Step | Time Complexity | Space Complexity | Notes |
|------|----------------|------------------|-------|
| Base URL construction | O(h) | O(h) | h = len(Host) |
| Proxy function creation | O(m) | O(m) | m = len(ProxyURL) |
| Base transport creation | O(1) | O(1) | Transport struct |
| Authenticator creation | O(a) | O(a) | a = auth-specific |
| Transport wrapping | O(1) | O(1) | Wrapper struct |
| HTTP client creation | O(1) | O(1) | Client struct |
| Download setup | O(a) | O(a) | Similar to HTTP client |
| Client assembly | O(1) | O(1) | Final struct |

**Overall Complexity:**
- **Time**: O(a) where a = authenticator creation time (O(1) for Basic, O(k) for SPNEGO)
- **Space**: O(a) where a = authenticator state size

## Sequence Diagrams

This section provides detailed sequence diagrams for different proxy authentication flows, showing the interactions between all components in the system.

### Sequence Diagram: Client Initialization with Basic Authentication

**Legend:**
- `[om-cli]` = Component in om-cli
- `[go-pivnet]` = Component in go-pivnet library

```
┌─────────────┐    ┌──────────────┐    ┌──────────────────┐    ┌─────────────┐    ┌──────────────┐
│ Client Code │    │  NewClient   │    │NewProxyAuthentic │    │BasicProxyAuth│   │ProxyAuthTrans│
│  [om-cli]   │    │ [go-pivnet]  │    │   [go-pivnet]    │    │  [go-pivnet]│   │  [go-pivnet] │
└──────┬──────┘    └──────┬───────┘    └────────┬─────────┘    └──────┬───────┘    └──────┬───────┘
       │                  │                      │                      │                    │
       │ 1. download_clients.NewPivnetClient()   │                      │                    │
       │    [om-cli]                              │                      │                    │
       │─────────────────────────────────────────>│                      │                    │
       │  (proxyURL, proxyUsername, etc.)         │                      │                    │
       │                                          │                      │                    │
       │                                          │ 2. pivnet.NewClient()│                    │
       │                                          │    [go-pivnet]      │                    │
       │                                          │────────────────────>│                    │
       │                                          │  (ClientConfig)     │                    │
       │                                          │                      │                    │
       │                                          │ 3. url.Parse()       │                    │
       │                                          │    [go-pivnet]       │                    │
       │                                          │─────────────────────┼────────────────────┼──┐
       │                                          │                      │                    │  │
       │                                          │ 4. http.ProxyURL()   │                    │  │
       │                                          │    [go-pivnet]       │                    │  │
       │                                          │─────────────────────┼────────────────────┼──┤
       │                                          │                      │                    │  │
       │                                          │ 5. Create Transport  │                    │  │
       │                                          │    [go-pivnet]       │                    │  │
       │                                          │─────────────────────┼────────────────────┼──┤
       │                                          │                      │                    │  │
       │                                          │ 6. NewProxyAuthenticator()                │  │
       │                                          │    [go-pivnet]       │                    │  │
       │                                          │────────────────────>│                    │  │
       │                                          │  (authType="basic")  │                    │  │
       │                                          │                      │                    │  │
       │                                          │                      │ 7. Validate params  │  │
       │                                          │                      │    [go-pivnet]     │  │
       │                                          │                      │────────────────────┼──┤
       │                                          │                      │                    │  │
       │                                          │                      │ 8. NewBasicProxyAuth()│
       │                                          │                      │    [go-pivnet]     │  │
       │                                          │                      │────────────────────>│  │
       │                                          │                      │                    │  │
       │                                          │                      │ 9. Return BasicProxyAuth│
       │                                          │                      │    [go-pivnet]     │  │
       │                                          │                      │<────────────────────│  │
       │                                          │                      │                    │  │
       │                                          │ 10. Return authenticator                  │  │
       │                                          │     [go-pivnet]      │                    │  │
       │                                          │<─────────────────────│                    │  │
       │                                          │                      │                    │  │
       │                                          │ 11. NewProxyAuthTransport()               │  │
       │                                          │     [go-pivnet]      │                    │  │
       │                                          │───────────────────────────────────────────┼──>│
       │                                          │                      │                    │  │
       │                                          │                      │ 12. Validate & create│
       │                                          │                      │     [go-pivnet]     │  │
       │                                          │                      │<────────────────────│  │
       │                                          │                      │                    │  │
       │                                          │ 13. Return ProxyAuthTransport            │  │
       │                                          │     [go-pivnet]      │                    │  │
       │                                          │<───────────────────────────────────────────┘  │
       │                                          │                      │                    │  │
       │                                          │ 14. Create http.Client                   │  │
       │                                          │     [go-pivnet]      │                    │  │
       │                                          │─────────────────────┼────────────────────┼──┤
       │                                          │                      │                    │  │
       │                                          │ 15. Return Client    │                    │  │
       │                                          │     [go-pivnet]      │                    │  │
       │                                          │<─────────────────────┼────────────────────┼──┤
       │                                          │                      │                    │  │
       │ 16. Return ProductDownloader              │                      │                    │  │
       │     [om-cli]                             │                      │                    │  │
       │<─────────────────────────────────────────│                      │                    │  │
       │                                          │                      │                    │  │
```

### Sequence Diagram: Client Initialization with SPNEGO Authentication

**Legend:**
- `[om-cli]` = Component in om-cli
- `[go-pivnet]` = Component in go-pivnet library

```
┌─────────────┐    ┌──────────────┐    ┌──────────────────┐    ┌──────────────┐    ┌────────────┐    ┌──────────┐
│ Client Code │    │  NewClient   │    │NewProxyAuthentic │    │SPNEGOProxyAuth│   │gokrb5     │    │   KDC    │
│  [om-cli]   │    │ [go-pivnet]  │    │   [go-pivnet]    │    │  [go-pivnet] │   │[external] │    │[external]│
└──────┬──────┘    └──────┬───────┘    └────────┬─────────┘    └──────┬───────┘    └─────┬──────┘    └────┬─────┘
       │                  │                      │                      │                   │                │
       │ 1. download_clients.NewPivnetClient()   │                      │                   │                │
       │    [om-cli]                              │                      │                   │                │
       │─────────────────────────────────────────>│                      │                   │                │
       │  (proxyURL, proxyUsername, etc.)         │                      │                   │                │
       │                                          │                      │                   │                │
       │                                          │ 2. pivnet.NewClient()│                   │                │
       │                                          │    [go-pivnet]      │                   │                │
       │                                          │────────────────────>│                   │                │
       │                                          │  (ClientConfig)     │                   │                │
       │                                          │                      │                   │                │
       │                                          │ 3. Setup proxy & transport             │                │
       │                                          │    [go-pivnet]      │                   │                │
       │                                          │─────────────────────┼───────────────────┼────────────────┼──┐
       │                                          │                      │                   │                │  │
       │                                          │ 4. NewProxyAuthenticator()              │                │  │
       │                                          │    [go-pivnet]      │                   │                │  │
       │                                          │────────────────────>│                   │                │  │
       │                                          │  (authType="spnego")│                   │                │  │
       │                                          │                      │                   │                │  │
       │                                          │                      │ 5. NewSPNEGOProxyAuth()            │  │
       │                                          │                      │    [go-pivnet]    │                │  │
       │                                          │                      │──────────────────>│                │  │
       │                  │                      │                      │                   │                │  │
       │                  │                      │ 5. Parse proxy URL   │                   │                │  │
       │                  │                      │─────────────────────┼───────────────────┼────────────────┼──┤
       │                  │                      │                      │                   │                │  │
       │                  │                      │ 6. Extract domain    │                   │                │  │
       │                  │                      │─────────────────────┼───────────────────┼────────────────┼──┤
       │                  │                      │                      │                   │                │  │
       │                  │                      │ 7. Load/Create krb5 config              │                │  │
       │                  │                      │─────────────────────┼───────────────────┼────────────────┼──┤
       │                  │                      │                      │                   │                │  │
       │                  │                      │ 8. client.NewWithPassword()            │                │  │
       │                  │                      │────────────────────────────────────────>│                │  │
       │                  │                      │                      │                   │                │  │
       │                  │                      │                      │ 9. kerberosClient.Login()         │  │
       │                  │                      │                      │───────────────────────────────────>│  │
       │                  │                      │                      │                   │                │  │
       │                  │                      │                      │                   │ 10. AS-REQ     │  │
       │                  │                      │                      │                   │───────────────>│  │
       │                  │                      │                      │                   │                │  │
       │                  │                      │                      │                   │ 11. AS-REP     │  │
       │                  │                      │                      │                   │<───────────────│  │
       │                  │                      │                      │                   │  (TGT)         │  │
       │                  │                      │                      │                   │                │  │
       │                  │                      │                      │ 12. Return client│                │  │
       │                  │                      │                      │<───────────────────│                │  │
       │                  │                      │                      │                   │                │  │
       │                  │                      │ 13. Return SPNEGOProxyAuth             │                │  │
       │                  │                      │<──────────────────────│                   │                │  │
       │                  │                      │                      │                   │                │  │
       │                  │ 14. Return authenticator                     │                   │                │  │
       │                  │<─────────────────────│                      │                   │                │  │
       │                  │                      │                      │                   │                │  │
       │                  │ 15. Wrap transport & create client           │                   │                │  │
       │                  │─────────────────────┼──────────────────────┼───────────────────┼────────────────┼──┘
       │                  │                      │                      │                   │                │
       │                  │ 16. Return Client    │                      │                   │                │
       │<─────────────────│                      │                      │                   │                │
       │                  │                      │                      │                   │                │
```

### Sequence Diagram: HTTP Request with Basic Authentication

**Legend:**
- `[om-cli]` = Component in om-cli
- `[go-pivnet]` = Component in go-pivnet library
- `[stdlib]` = Go standard library

```
┌─────────────┐    ┌──────────────┐    ┌──────────────────┐    ┌─────────────┐    ┌──────────────┐    ┌──────────┐    ┌──────────┐
│ Client Code │    │ http.Client  │    │ProxyAuthTransport│    │BasicProxyAuth│   │http.Transport│    │  Proxy   │    │  Pivnet  │
│  [om-cli]   │    │  [stdlib]    │    │   [go-pivnet]    │    │  [go-pivnet]│   │  [stdlib]   │    │[external]│    │[external]│
└──────┬──────┘    └──────┬───────┘    └────────┬─────────┘    └──────┬───────┘    └──────┬───────┘    └────┬─────┘    └────┬─────┘
       │                  │                      │                      │                    │                  │                │
       │ 1. client.Products.List()              │                      │                    │                  │                │
       │─────────────────>│                      │                      │                    │                  │                │
       │                  │                      │                      │                    │                  │                │
       │                  │ 2. Create HTTP Request                      │                    │                  │                │
       │                  │─────────────────────┼──────────────────────┼───────────────────┼─────────────────┼────────────────┼──┐
       │                  │                      │                      │                    │                  │                │  │
       │                  │ 3. httpClient.Do(req)│                      │                    │                  │                │  │
       │                  │────────────────────>│                      │                    │                  │                │  │
       │                  │                      │                      │                    │                  │                │  │
       │                  │                      │ 4. RoundTrip(req)    │                    │                  │                │  │
       │                  │                      │─────────────────────┼────────────────────┼─────────────────┼────────────────┼──┤
       │                  │                      │                      │                    │                  │                │  │
       │                  │                      │ 5. Authenticate(req) │                    │                  │                │  │
       │                  │                      │─────────────────────>│                    │                  │                │  │
       │                  │                      │                      │                    │                  │                │  │
       │                  │                      │                      │ 6. Encode credentials│                  │                │  │
       │                  │                      │                      │    username:password│                  │                │  │
       │                  │                      │                      │─────────────────────┼─────────────────┼────────────────┼──┤
       │                  │                      │                      │                    │                  │                │  │
       │                  │                      │                      │ 7. Base64 encode    │                  │                │  │
       │                  │                      │                      │─────────────────────┼─────────────────┼────────────────┼──┤
       │                  │                      │                      │                    │                  │                │  │
       │                  │                      │                      │ 8. Set header       │                  │                │  │
       │                  │                      │                      │    Proxy-Authorization│                 │                │  │
       │                  │                      │                      │─────────────────────┼─────────────────┼────────────────┼──┤
       │                  │                      │                      │                    │                  │                │  │
       │                  │                      │ 9. Return (nil)      │                    │                  │                │  │
       │                  │                      │<─────────────────────│                    │                  │                │  │
       │                  │                      │                      │                    │                  │                │  │
       │                  │                      │ 10. Transport.RoundTrip(req)             │                  │                │  │
       │                  │                      │───────────────────────────────────────────>│                  │                │  │
       │                  │                      │                      │                    │                  │                │  │
       │                  │                      │                      │                    │ 11. Proxy function│                │  │
       │                  │                      │                      │                    │─────────────────>│                │  │
       │                  │                      │                      │                    │                  │                │  │
       │                  │                      │                      │                    │ 12. Connect to proxy│              │  │
       │                  │                      │                      │                    │───────────────────>│                │  │
       │                  │                      │                      │                    │                  │                │  │
       │                  │                      │                      │                    │ 13. HTTP CONNECT   │                │  │
       │                  │                      │                      │                    │    with auth header│                │  │
       │                  │                      │                      │                    │───────────────────>│                │  │
       │                  │                      │                      │                    │                  │                │  │
       │                  │                      │                      │                    │                  │ 14. Authenticate│  │
       │                  │                      │                      │                    │                  │─────────────────┼──┤
       │                  │                      │                      │                    │                  │                │  │
       │                  │                      │                      │                    │                  │ 15. Forward request│  │
       │                  │                      │                      │                    │                  │─────────────────>│  │
       │                  │                      │                      │                    │                  │                │  │
       │                  │                      │                      │                    │                  │                │ 16. Response│
       │                  │                      │                      │                    │                  │                │<───────────│
       │                  │                      │                      │                    │                  │                │            │
       │                  │                      │                      │                    │                  │ 17. Return response│          │
       │                  │                      │                      │                    │                  │<─────────────────│            │
       │                  │                      │                      │                    │                  │                │            │
       │                  │                      │                      │                    │ 18. Return response│                │            │
       │                  │                      │                      │                    │<───────────────────│                │            │
       │                  │                      │                      │                    │                  │                │            │
       │                  │                      │ 19. Return response  │                    │                  │                │            │
       │                  │                      │<─────────────────────│                    │                  │                │            │
       │                  │                      │                      │                    │                  │                │            │
       │                  │ 20. Return response  │                      │                    │                  │                │            │
       │                  │<─────────────────────│                      │                    │                  │                │            │
       │                  │                      │                      │                    │                  │                │            │
       │ 21. Return products│                    │                      │                    │                  │                │            │
       │<─────────────────│                      │                      │                    │                  │                │            │
       │                  │                      │                      │                    │                  │                │            │
```

### Sequence Diagram: HTTP Request with SPNEGO Authentication

**Legend:**
- `[om-cli]` = Component in om-cli
- `[go-pivnet]` = Component in go-pivnet library
- `[stdlib]` = Go standard library
- `[external]` = External library or service

```
┌─────────────┐    ┌──────────────┐    ┌──────────────────┐    ┌──────────────┐    ┌────────────┐    ┌──────────┐    ┌──────────┐
│ Client Code │    │ http.Client  │    │ProxyAuthTransport│    │SPNEGOProxyAuth│   │gokrb5      │    │  Proxy   │    │  Pivnet  │
│  [om-cli]   │    │  [stdlib]    │    │   [go-pivnet]    │    │  [go-pivnet] │   │[external]  │    │[external]│    │[external]│
└──────┬──────┘    └──────┬───────┘    └────────┬─────────┘    └──────┬───────┘    └─────┬──────┘    └────┬─────┘    └────┬─────┘
       │                  │                      │                      │                   │                │                │
       │ 1. client.Products.List()              │                      │                   │                │                │
       │─────────────────>│                      │                      │                   │                │                │
       │                  │                      │                      │                   │                │                │
       │                  │ 2. Create HTTP Request                      │                   │                │                │
       │                  │─────────────────────┼──────────────────────┼───────────────────┼────────────────┼────────────────┼──┐
       │                  │                      │                      │                   │                │                │  │
       │                  │ 3. httpClient.Do(req)│                      │                   │                │                │  │
       │                  │────────────────────>│                      │                   │                │                │  │
       │                  │                      │                      │                   │                │                │  │
       │                  │                      │ 4. RoundTrip(req)    │                   │                │                │  │
       │                  │                      │─────────────────────┼────────────────────┼────────────────┼────────────────┼──┤
       │                  │                      │                      │                   │                │                │  │
       │                  │                      │ 5. Authenticate(req) │                   │                │                │  │
       │                  │                      │─────────────────────>│                   │                │                │  │
       │                  │                      │                      │                   │                │                │  │
       │                  │                      │                      │ 6. spnego.SPNEGOClient()│          │                │  │
       │                  │                      │                      │───────────────────>│                │                │  │
       │                  │                      │                      │                   │                │                │  │
       │                  │                      │                      │ 7. InitSecContext()│                │                │  │
       │                  │                      │                      │───────────────────>│                │                │  │
       │                  │                      │                      │                   │                │                │  │
       │                  │                      │                      │                   │ 8. Check cache  │                │  │
       │                  │                      │                      │                   │─────────────────┼────────────────┼──┤
       │                  │                      │                      │                   │                │                │  │
       │                  │                      │                      │                   │ 9a. If cached:  │                │  │
       │                  │                      │                      │                   │    Use service  │                │  │
       │                  │                      │                      │                   │    ticket       │                │  │
       │                  │                      │                      │                   │─────────────────┼────────────────┼──┤
       │                  │                      │                      │                   │                │                │  │
       │                  │                      │                      │                   │ 9b. If not cached: Request TGS│  │
       │                  │                      │                      │                   │───────────────────────────────────>│  │
       │                  │                      │                      │                   │                │                │  │
       │                  │                      │                      │                   │                │                │ 10. TGS-REP│
       │                  │                      │                      │                   │                │                │<──────────│
       │                  │                      │                      │                   │                │                │            │
       │                  │                      │                      │                   │ 11. Return token│                │            │
       │                  │                      │                      │                   │<────────────────│                │            │
       │                  │                      │                      │                   │                │                │            │
       │                  │                      │                      │ 12. Marshal token │                │                │            │
       │                  │                      │                      │    (ASN.1 DER)    │                │                │            │
       │                  │                      │                      │───────────────────┼────────────────┼────────────────┼──┤
       │                  │                      │                      │                   │                │                │            │
       │                  │                      │                      │ 13. Base64 encode │                │                │            │
       │                  │                      │                      │───────────────────┼────────────────┼────────────────┼──┤
       │                  │                      │                      │                   │                │                │            │
       │                  │                      │                      │ 14. Set header    │                │                │            │
       │                  │                      │                      │    Proxy-Authorization│           │                │            │
       │                  │                      │                      │───────────────────┼────────────────┼────────────────┼──┤
       │                  │                      │                      │                   │                │                │            │
       │                  │                      │ 15. Return (nil)     │                   │                │                │            │
       │                  │                      │<─────────────────────│                   │                │                │            │
       │                  │                      │                      │                   │                │                │            │
       │                  │                      │ 16. Transport.RoundTrip(req)             │                │                │            │
       │                  │                      │──────────────────────────────────────────┼────────────────┼────────────────┼──┘
       │                  │                      │                      │                   │                │                │
       │                  │                      │                      │                   │ 17. Connect & send│              │
       │                  │                      │                      │                   │───────────────────>│              │
       │                  │                      │                      │                   │                  │              │
       │                  │                      │                      │                   │                  │ 18. Authenticate│
       │                  │                      │                      │                   │                  │─────────────────┼──┐
       │                  │                      │                      │                   │                  │                │  │
       │                  │                      │                      │                   │                  │ 19. Forward     │  │
       │                  │                      │                      │                   │                  │─────────────────>│  │
       │                  │                      │                      │                   │                  │                │  │
       │                  │                      │                      │                   │                  │                │ 20. Response│
       │                  │                      │                      │                   │                  │                │<───────────│
       │                  │                      │                      │                   │                  │                │            │
       │                  │                      │                      │                   │                  │ 21. Return     │            │
       │                  │                      │                      │                   │                  │<────────────────│            │
       │                  │                      │                      │                   │                  │                │            │
       │                  │                      │                      │                   │ 22. Return       │                │            │
       │                  │                      │                      │                   │<──────────────────│                │            │
       │                  │                      │                      │                   │                  │                │            │
       │                  │                      │ 23. Return response  │                   │                  │                │            │
       │                  │                      │<─────────────────────│                   │                  │                │            │
       │                  │                      │                      │                   │                  │                │            │
       │                  │ 24. Return response  │                      │                   │                  │                │            │
       │                  │<─────────────────────│                      │                   │                  │                │            │
       │                  │                      │                      │                   │                  │                │            │
       │ 25. Return products│                    │                      │                   │                  │                │            │
       │<─────────────────│                      │                      │                   │                  │                │            │
       │                  │                      │                      │                   │                  │                │            │
```

### Sequence Diagram: SPNEGO Token Generation (Detailed)

```
┌──────────────┐    ┌────────────┐    ┌──────────┐    ┌──────────┐
│SPNEGOProxyAuth│   │gokrb5      │    │  KDC     │    │  Cache   │
└──────┬───────┘    └─────┬──────┘    └────┬─────┘    └────┬─────┘
       │                  │                 │               │
       │ 1. Authenticate(req)               │               │
       │──────────────────>│                 │               │
       │                  │                 │               │
       │                  │ 2. spnego.SPNEGOClient()        │
       │                  │─────────────────┼───────────────┼──┐
       │                  │                 │               │  │
       │                  │ 3. InitSecContext()            │  │
       │                  │─────────────────┼───────────────┼──┤
       │                  │                 │               │  │
       │                  │ 4. Get SPN: "HTTP/proxy.example.com"│
       │                  │─────────────────┼───────────────┼──┤
       │                  │                 │               │  │
       │                  │ 5. Check cache for service ticket│  │
       │                  │─────────────────────────────────>│  │
       │                  │                 │               │  │
       │                  │                 │               │ 6. Cache lookup│
       │                  │                 │               │─────────────────┼──┤
       │                  │                 │               │                │  │
       │                  │                 │               │ 7a. If found & valid│
       │                  │                 │               │─────────────────┼──┤
       │                  │                 │               │                │  │
       │                  │                 │               │ 8. Return cached ticket│
       │                  │                 │               │<─────────────────│  │
       │                  │                 │               │                │  │
       │                  │                 │               │                │  │
       │                  │ 7b. If not cached: Request TGS│                │  │
       │                  │──────────────────────────────────────────────────>│  │
       │                  │                 │               │                │  │
       │                  │                 │ 9. TGS-REQ     │                │  │
       │                  │                 │───────────────>│                │  │
       │                  │                 │  (with TGT)    │                │  │
       │                  │                 │               │                │  │
       │                  │                 │ 10. TGS-REP   │                │  │
       │                  │                 │<───────────────│                │  │
       │                  │                 │  (service ticket)│             │  │
       │                  │                 │               │                │  │
       │                  │ 11. Cache service ticket         │                │  │
       │                  │─────────────────────────────────>│                │  │
       │                  │                 │               │                │  │
       │                  │ 12. Create AP-REQ with service ticket│          │  │
       │                  │─────────────────┼───────────────┼───────────────┼──┤
       │                  │                 │               │                │  │
       │                  │ 13. Wrap in GSS-API structure  │                │  │
       │                  │─────────────────┼───────────────┼───────────────┼──┤
       │                  │                 │               │                │  │
       │                  │ 14. Encode as ASN.1 DER         │                │  │
       │                  │─────────────────┼───────────────┼───────────────┼──┤
       │                  │                 │               │                │  │
       │                  │ 15. Return token                │                │  │
       │                  │<─────────────────┼───────────────┼───────────────┼──┘
       │                  │                 │               │                │
       │ 16. Marshal & encode token          │               │                │
       │──────────────────┼─────────────────┼───────────────┼────────────────┼──┐
       │                  │                 │               │                │  │
       │ 17. Set header & return            │               │                │  │
       │──────────────────┼─────────────────┼───────────────┼────────────────┼──┤
       │                  │                 │               │                │  │
```

### Sequence Diagram: Error Handling Flow

```
┌─────────────┐    ┌──────────────┐    ┌──────────────────┐    ┌──────────────┐    ┌────────────┐
│ Client Code │    │  NewClient   │    │NewProxyAuthentic │    │SPNEGOProxyAuth│   │   Logger   │
└──────┬──────┘    └──────┬───────┘    └────────┬─────────┘    └──────┬───────┘    └─────┬──────┘
       │                  │                      │                      │                   │
       │ 1. NewClient()   │                      │                      │                   │
       │─────────────────>│                      │                      │                   │
       │                  │                      │                      │                   │
       │                  │ 2. NewProxyAuthenticator()                  │                   │
       │                  │─────────────────────>│                      │                   │
       │                  │                      │                      │                   │
       │                  │                      │ 3. NewSPNEGOProxyAuth()                │
       │                  │                      │─────────────────────>│                   │
       │                  │                      │                      │                   │
       │                  │                      │                      │ 4. KDC Login()   │
       │                  │                      │                      │───────────────────┼──┐
       │                  │                      │                      │                   │  │
       │                  │                      │                      │ 5. Error: Login failed│
       │                  │                      │                      │<───────────────────│  │
       │                  │                      │                      │                   │  │
       │                  │                      │                      │ 6. Return error   │  │
       │                  │                      │                      │───────────────────┼──┤
       │                  │                      │                      │                   │  │
       │                  │                      │ 7. Return error       │                   │  │
       │                  │                      │<─────────────────────│                   │  │
       │                  │                      │                      │                   │  │
       │                  │ 8. Return error      │                      │                   │  │
       │                  │<─────────────────────│                      │                   │  │
       │                  │                      │                      │                   │  │
       │                  │ 9. Log error         │                      │                   │  │
       │                  │─────────────────────────────────────────────┼──────────────────>│  │
       │                  │                      │                      │                   │  │
       │                  │ 10. Continue without auth                    │                   │  │
       │                  │─────────────────────┼──────────────────────┼───────────────────┼──┤
       │                  │                      │                      │                   │  │
       │                  │ 11. Create client with base transport      │                   │  │
       │                  │─────────────────────┼──────────────────────┼───────────────────┼──┤
       │                  │                      │                      │                   │  │
       │                  │ 12. Return Client (no auth)                  │                   │  │
       │<─────────────────│                      │                      │                   │  │
       │                  │                      │                      │                   │  │
```

### Sequence Diagram: Resource Cleanup (SPNEGO)

```
┌─────────────┐    ┌──────────────┐    ┌──────────────┐    ┌────────────┐
│ Client Code │    │ProxyAuthTrans│    │SPNEGOProxyAuth│   │gokrb5      │
└──────┬──────┘    └──────┬───────┘    └──────┬───────┘    └─────┬──────┘
       │                  │                   │                   │
       │ 1. Close()       │                   │                   │
       │─────────────────>│                   │                   │
       │                  │                   │                   │
       │                  │ 2. Close()        │                   │
       │                  │──────────────────>│                   │
       │                  │                   │                   │
       │                  │                   │ 3. Close()        │
       │                  │                   │──────────────────>│
       │                  │                   │                   │
       │                  │                   │                   │ 4. Destroy client│
       │                  │                   │                   │───────────────────┼──┐
       │                  │                   │                   │                   │  │
       │                  │                   │                   │ 5. Close connections│
       │                  │                   │                   │───────────────────┼──┤
       │                  │                   │                   │                   │  │
       │                  │                   │                   │ 6. Clear cache    │  │
       │                  │                   │                   │───────────────────┼──┤
       │                  │                   │                   │                   │  │
       │                  │                   │                   │ 7. Release resources│
       │                  │                   │                   │───────────────────┼──┤
       │                  │                   │                   │                   │  │
       │                  │                   │                   │ 8. Return (nil)   │  │
       │                  │                   │                   │<───────────────────│  │
       │                  │                   │                   │                   │  │
       │                  │                   │ 9. Return (nil)   │                   │  │
       │                  │                   │<──────────────────│                   │  │
       │                  │                   │                   │                   │  │
       │                  │ 10. Return (nil)  │                   │                   │  │
       │                  │<──────────────────│                   │                   │  │
       │                  │                   │                   │                   │  │
       │ 11. Cleanup complete                │                   │                   │  │
       │<─────────────────│                   │                   │                   │  │
       │                  │                   │                   │                   │  │
```

### Sequence Diagram: Concurrent Requests with Basic Auth

```
┌─────────────┐    ┌──────────────┐    ┌──────────────────┐    ┌─────────────┐
│ Request 1   │    │ http.Client  │    │ProxyAuthTransport│    │BasicProxyAuth│
│ Request 2   │    │              │    │                  │    │             │
└──────┬──────┘    └──────┬───────┘    └────────┬─────────┘    └──────┬──────┘
       │                  │                      │                      │
       │ 1. Do(req1)      │                      │                      │
       │─────────────────>│                      │                      │
       │                  │                      │                      │
       │ 2. Do(req2)      │                      │                      │
       │─────────────────>│                      │                      │
       │                  │                      │                      │
       │                  │ 3. RoundTrip(req1)   │                      │
       │                  │─────────────────────>│                      │
       │                  │                      │                      │
       │                  │ 4. RoundTrip(req2)   │                      │
       │                  │─────────────────────>│                      │
       │                  │                      │                      │
       │                  │                      │ 5. Authenticate(req1) │
       │                  │                      │─────────────────────>│
       │                  │                      │                      │
       │                  │                      │ 6. Authenticate(req2) │
       │                  │                      │─────────────────────>│
       │                  │                      │                      │
       │                  │                      │                      │ 7. Encode & set header (req1)│
       │                  │                      │                      │─────────────────────┼──┐
       │                  │                      │                      │                   │  │
       │                  │                      │                      │ 8. Encode & set header (req2)│
       │                  │                      │                      │─────────────────────┼──┤
       │                  │                      │                      │                   │  │
       │                  │                      │ 9. Return (req1)     │                   │  │
       │                  │                      │<─────────────────────│                   │  │
       │                  │                      │                      │                   │  │
       │                  │                      │ 10. Return (req2)    │                   │  │
       │                  │                      │<─────────────────────│                   │  │
       │                  │                      │                      │                   │  │
       │                  │ 11. Execute req1 & req2 (parallel)          │                   │  │
       │                  │─────────────────────┼──────────────────────┼───────────────────┼──┤
       │                  │                      │                      │                   │  │
       │                  │ 12. Return responses │                      │                   │  │
       │<─────────────────│                      │                      │                   │  │
       │                  │                      │                      │                   │  │
```

**Key Points:**
- Both requests can be authenticated concurrently
- BasicProxyAuth is stateless and thread-safe
- No shared mutable state between requests
- Each request gets its own authentication header

### Sequence Diagram: Download Flow with Proxy Authentication

```
┌─────────────┐    ┌──────────────┐    ┌──────────────────┐    ┌──────────────┐    ┌────────────┐
│ Client Code │    │download.Client│   │ProxyAuthTransport│    │SPNEGOProxyAuth│   │  Download  │
└──────┬──────┘    └──────┬───────┘    └────────┬─────────┘    └──────┬───────┘    └─────┬──────┘
       │                  │                      │                      │                   │
       │ 1. DownloadProductFile()               │                      │                   │
       │─────────────────>│                      │                      │                   │
       │                  │                      │                      │                   │
       │                  │ 2. Get download link │                      │                   │
       │                  │─────────────────────┼──────────────────────┼──────────────────┼──┐
       │                  │                      │                      │                   │  │
       │                  │ 3. HEAD request     │                      │                   │  │
       │                  │─────────────────────>│                      │                   │  │
       │                  │                      │                      │                   │  │
       │                  │                      │ 4. Authenticate & send│                   │  │
       │                  │                      │─────────────────────┼──────────────────┼──┤
       │                  │                      │                      │                   │  │
       │                  │                      │                      │ 5. Generate token │  │
       │                  │                      │                      │───────────────────┼──┤
       │                  │                      │                      │                   │  │
       │                  │                      │ 6. Return response    │                   │  │
       │                  │                      │<─────────────────────┼──────────────────┼──┤
       │                  │                      │                      │                   │  │
       │                  │ 7. Return HEAD response                    │                   │  │
       │                  │<─────────────────────│                      │                   │  │
       │                  │                      │                      │                   │  │
       │                  │ 8. Build ranges      │                      │                   │  │
       │                  │─────────────────────┼──────────────────────┼──────────────────┼──┤
       │                  │                      │                      │                   │  │
       │                  │ 9. GET requests (parallel chunks)           │                   │  │
       │                  │─────────────────────>│                      │                   │  │
       │                  │                      │                      │                   │  │
       │                  │                      │ 10. Authenticate each request            │  │
       │                  │                      │─────────────────────>│                   │  │
       │                  │                      │                      │                   │  │
       │                  │                      │                      │ 11. Generate tokens│  │
       │                  │                      │                      │───────────────────┼──┤
       │                  │                      │                      │                   │  │
       │                  │                      │ 12. Execute downloads│                   │  │
       │                  │                      │─────────────────────┼───────────────────┼──┤
       │                  │                      │                      │                   │  │
       │                  │ 13. Return chunks    │                      │                   │  │
       │                  │<─────────────────────│                      │                   │  │
       │                  │                      │                      │                   │  │
       │ 14. Return file  │                      │                      │                   │  │
       │<─────────────────│                      │                      │                   │  │
       │                  │                      │                      │                   │  │
```

**Key Points:**
- Download uses separate HTTP client with same proxy authentication
- Multiple concurrent chunk downloads each get authenticated
- SPNEGO tokens may be cached and reused across chunks
- All chunks go through same proxy with authentication

## Implementation Details

### Code Flow: Complete Request with SPNEGO

```go
// 1. Client makes request
client.Products.List()
    │
    ▼
// 2. Client creates HTTP request
req := http.NewRequest("GET", "https://network.pivotal.io/api/v2/products", nil)
    │
    ▼
// 3. Client calls HTTP.Do()
httpClient.Do(req)
    │
    ▼
// 4. Transport intercepts
ProxyAuthTransport.RoundTrip(req)
    │
    ▼
// 5. Authenticate request
SPNEGOProxyAuth.Authenticate(req)
    │
    │ 5.1. Create SPNEGO client
    │      spnego.SPNEGOClient(kerberosClient, "HTTP")
    │
    │ 5.2. Generate token
    │      spnegoClient.InitSecContext()
    │      │
    │      │ Uses cached service ticket if available
    │      │ Otherwise requests from KDC using TGT
    │      │
    │      │ Returns: *spnego.Token
    │
    │ 5.3. Marshal token
    │      token.Marshal()
    │      │
    │      │ ASN.1 DER encoding
    │      │
    │      │ Returns: []byte
    │
    │ 5.4. Base64 encode
    │      base64.StdEncoding.EncodeToString(tokenBytes)
    │
    │ 5.5. Set header
    │      req.Header.Set("Proxy-Authorization", "Negotiate " + encodedToken)
    │
    ▼
// 6. Execute request
baseTransport.RoundTrip(req)
    │
    │ 6.1. Proxy function called
    │      proxyFunc(req)
    │      │
    │      │ Returns: *url.URL{proxy URL}
    │
    │ 6.2. Establish connection to proxy
    │      TCP connection to proxy:8080
    │
    │ 6.3. Send HTTP request through proxy
    │      GET https://network.pivotal.io/api/v2/products HTTP/1.1
    │      Host: network.pivotal.io
    │      Proxy-Authorization: Negotiate YIIBhwYGKwYBBQUCoIIBezCCAXe...
    │
    │ 6.4. Proxy authenticates and forwards
    │
    │ 6.5. Receive response
    │
    ▼
// 7. Return response
return resp, nil
```

### Memory Allocation Points

1. **ClientConfig creation**: Heap allocation for strings
2. **ProxyAuthTransport creation**: Stack allocation (small struct)
3. **BasicProxyAuth creation**: Stack allocation (small struct)
4. **SPNEGOProxyAuth creation**: 
   - Stack: Struct itself
   - Heap: Kerberos client, cached tickets
5. **Token generation**: Heap allocation for token bytes
6. **Base64 encoding**: Heap allocation for encoded string

### Network I/O Points

1. **SPNEGO Initialization**: 
   - KDC connection (TCP port 88)
   - AS-REQ / AS-REP exchange
2. **Token Generation** (if service ticket not cached):
   - KDC connection
   - TGS-REQ / TGS-REP exchange
3. **HTTP Request**:
   - Proxy connection (TCP)
   - HTTP request/response

---

## Testing Considerations

### Unit Test Coverage

**Components to Test:**
1. `NewProxyAuthenticator()` - Factory function
2. `BasicProxyAuth.Authenticate()` - Header generation
3. `SPNEGOProxyAuth.Authenticate()` - Token generation
4. `ProxyAuthTransport.RoundTrip()` - Request interception
5. `NewClient()` - Integration with proxy auth

**Test Cases:**
- Valid configurations
- Invalid configurations
- Error conditions
- Concurrent access
- Resource cleanup

### Integration Test Scenarios

1. **Basic Auth with Mock Proxy**: Verify header format
2. **SPNEGO with Mock KDC**: Verify token generation
3. **End-to-End**: Real proxy with authentication
4. **Error Recovery**: Failed authentication scenarios
5. **Performance**: Concurrent request handling

---

## Conclusion

This low-level design document provides a comprehensive technical specification for the proxy authentication mechanism in go-pivnet. The implementation follows a clean, extensible architecture that supports multiple authentication mechanisms while maintaining thread safety and efficient resource usage.

Key design principles:
- **Interface-based design**: Extensible authenticator interface
- **Separation of concerns**: Transport, authentication, and configuration are separate
- **Resource management**: Explicit lifecycle management for SPNEGO
- **Error handling**: Graceful degradation and comprehensive error reporting
- **Performance**: Token caching and efficient encoding

The design allows for easy extension with new authentication mechanisms while maintaining backward compatibility and performance characteristics.

