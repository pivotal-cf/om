package download_clients

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

// ProxyAuthType represents the type of proxy authentication mechanism
type ProxyAuthType string

const (
	ProxyAuthTypeNone   ProxyAuthType = "none"   // No authentication
	ProxyAuthTypeBasic  ProxyAuthType = "basic"  // Basic HTTP authentication (username/password in URL)
	ProxyAuthTypeSPNEGO ProxyAuthType = "spnego" // SPNEGO/Kerberos authentication
	// Future: ProxyAuthTypeNTLM, ProxyAuthTypeOAuth, etc.
)

// ProxyAuthConfig holds all proxy configuration parameters
type ProxyAuthConfig struct {
	URL      string
	Username string
	Password string
	Domain   string        // Used for SPNEGO/Kerberos
	Type     ProxyAuthType // If empty, will be auto-determined
}

// ProxyAuth is the interface that all proxy authentication mechanisms must implement
type ProxyAuth interface {
	// ConfigureTransport configures the HTTP transport to use this authentication mechanism
	// It should set environment variables, wrap transports, or configure clients as needed
	ConfigureTransport(transport http.RoundTripper, proxyURL *url.URL) (http.RoundTripper, error)
	
	// Type returns the authentication type this mechanism implements
	Type() ProxyAuthType
	
	// Validate checks if the configuration is valid for this mechanism
	Validate(config ProxyAuthConfig) error
}

// ProxyAuthRegistry manages available proxy authentication mechanisms
type ProxyAuthRegistry struct {
	mechanisms map[ProxyAuthType]ProxyAuth
}

// NewProxyAuthRegistry creates a new registry with default mechanisms
func NewProxyAuthRegistry() *ProxyAuthRegistry {
	registry := &ProxyAuthRegistry{
		mechanisms: make(map[ProxyAuthType]ProxyAuth),
	}
	
	// Register default mechanisms
	registry.Register(NewBasicProxyAuth())
	registry.Register(NewSPNEGOProxyAuth())
	
	return registry
}

// Register adds a new authentication mechanism to the registry
func (r *ProxyAuthRegistry) Register(auth ProxyAuth) {
	r.mechanisms[auth.Type()] = auth
}

// Get returns the authentication mechanism for the given type
func (r *ProxyAuthRegistry) Get(authType ProxyAuthType) (ProxyAuth, error) {
	auth, ok := r.mechanisms[authType]
	if !ok {
		return nil, fmt.Errorf("unknown proxy authentication type: %s", authType)
	}
	return auth, nil
}

// DetermineAuthType determines the appropriate authentication type based on configuration
func DetermineAuthType(config ProxyAuthConfig) ProxyAuthType {
	if config.URL == "" {
		return ProxyAuthTypeNone
	}
	
	// If type is explicitly set, use it
	if config.Type != "" {
		return config.Type
	}
	
	// SPNEGO takes precedence if domain is provided
	if config.Domain != "" {
		return ProxyAuthTypeSPNEGO
	}
	
	// Basic auth if username/password provided
	if config.Username != "" || config.Password != "" {
		return ProxyAuthTypeBasic
	}
	
	// No authentication
	return ProxyAuthTypeNone
}

// ConfigureProxyAuth configures proxy authentication based on the provided config
func ConfigureProxyAuth(config ProxyAuthConfig, registry *ProxyAuthRegistry, stderr *log.Logger) error {
	if config.URL == "" {
		return nil // No proxy configured
	}
	
	authType := DetermineAuthType(config)
	if authType == ProxyAuthTypeNone {
		// Set proxy without authentication
		os.Setenv("HTTP_PROXY", config.URL)
		os.Setenv("HTTPS_PROXY", config.URL)
		os.Setenv("http_proxy", config.URL)
		os.Setenv("https_proxy", config.URL)
		return nil
	}
	
	auth, err := registry.Get(authType)
	if err != nil {
		return fmt.Errorf("failed to get proxy auth mechanism: %w", err)
	}
	
	if err := auth.Validate(config); err != nil {
		if stderr != nil {
			stderr.Printf("Warning: Invalid proxy auth configuration: %s. Falling back to basic auth.", err)
		}
		// Fall back to basic auth
		authType = ProxyAuthTypeBasic
		auth, err = registry.Get(authType)
		if err != nil {
			return fmt.Errorf("failed to get basic auth mechanism: %w", err)
		}
	}
	
	parsedURL, err := url.Parse(config.URL)
	if err != nil {
		return fmt.Errorf("failed to parse proxy URL: %w", err)
	}
	
	// For SPNEGO, we need to initialize it with the config first
	if authType == ProxyAuthTypeSPNEGO {
		spnegoAuth, ok := auth.(*SPNEGOProxyAuth)
		if ok {
			if err := spnegoAuth.Initialize(config); err != nil {
				if stderr != nil {
					stderr.Printf("Warning: Failed to initialize SPNEGO authentication: %s. Falling back to basic auth.", err)
				}
				// Fall back to basic auth
				authType = ProxyAuthTypeBasic
				auth, err = registry.Get(authType)
				if err != nil {
					return fmt.Errorf("failed to get basic auth mechanism: %w", err)
				}
			}
		}
	}
	
	// Configure the transport
	transport := http.DefaultTransport
	if transport == nil {
		transport = &http.Transport{}
	}
	
	newTransport, err := auth.ConfigureTransport(transport, parsedURL)
	if err != nil {
		return fmt.Errorf("failed to configure proxy auth transport: %w", err)
	}
	
	http.DefaultTransport = newTransport
	
	// Set proxy environment variables
	proxyURLForEnv := config.URL
	if authType == ProxyAuthTypeBasic {
		// Basic auth embeds credentials in URL
		proxyURLForEnv = BuildProxyURLWithAuth(config.URL, config.Username, config.Password)
	}
	// SPNEGO handles auth via headers, so use URL without credentials
	
	os.Setenv("HTTP_PROXY", proxyURLForEnv)
	os.Setenv("HTTPS_PROXY", proxyURLForEnv)
	os.Setenv("http_proxy", proxyURLForEnv)
	os.Setenv("https_proxy", proxyURLForEnv)
	
	return nil
}

