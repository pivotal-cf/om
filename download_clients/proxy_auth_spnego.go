package download_clients

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/jcmturner/gokrb5/v8/client"
	krb5config "github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/spnego"
)

// SPNEGOProxyAuth implements SPNEGO/Kerberos proxy authentication
type SPNEGOProxyAuth struct {
	krb5Client *client.Client
	tokenCache map[string]string // Cache tokens per proxy host
	mu         sync.RWMutex
	config     ProxyAuthConfig
}

// NewSPNEGOProxyAuth creates a new SPNEGO proxy authenticator
func NewSPNEGOProxyAuth() *SPNEGOProxyAuth {
	return &SPNEGOProxyAuth{
		tokenCache: make(map[string]string),
	}
}

// Type returns the authentication type
func (s *SPNEGOProxyAuth) Type() ProxyAuthType {
	return ProxyAuthTypeSPNEGO
}

// Validate checks if the configuration is valid for SPNEGO auth
func (s *SPNEGOProxyAuth) Validate(config ProxyAuthConfig) error {
	if config.URL == "" {
		return fmt.Errorf("proxy URL is required")
	}
	if config.Domain == "" {
		return fmt.Errorf("domain is required for SPNEGO authentication")
	}
	if config.Username == "" {
		return fmt.Errorf("username is required for SPNEGO authentication")
	}
	if config.Password == "" {
		return fmt.Errorf("password is required for SPNEGO authentication")
	}
	return nil
}

// Initialize sets up the Kerberos client with the provided configuration
func (s *SPNEGOProxyAuth) Initialize(config ProxyAuthConfig) error {
	s.config = config

	// Load Kerberos configuration
	krb5Config, err := krb5config.Load(configPath())
	if err != nil {
		return fmt.Errorf("failed to load Kerberos config: %w", err)
	}

	// Create Kerberos client
	krb5Client := client.NewWithPassword(
		config.Username,
		config.Domain,
		config.Password,
		krb5Config,
		client.DisablePAFXFAST(true),
	)

	err = krb5Client.Login()
	if err != nil {
		return fmt.Errorf("failed to login to Kerberos: %w", err)
	}

	s.krb5Client = krb5Client
	return nil
}

// GetSPNEGOToken generates a SPNEGO token for the proxy
func (s *SPNEGOProxyAuth) GetSPNEGOToken(proxyHost string) (string, error) {
	s.mu.RLock()
	if token, ok := s.tokenCache[proxyHost]; ok {
		s.mu.RUnlock()
		return token, nil
	}
	s.mu.RUnlock()

	// Generate SPNEGO token
	// The service principal name for HTTP proxy is typically HTTP/proxy-host@REALM
	spn := fmt.Sprintf("HTTP/%s", proxyHost)

	// Create SPNEGO client
	spnegoClient := spnego.SPNEGOClient(s.krb5Client, spn)

	// Get the token
	contextToken, err := spnegoClient.InitSecContext()
	if err != nil {
		return "", fmt.Errorf("failed to initialize SPNEGO context: %w", err)
	}

	// Marshal the token to bytes
	tokenBytes, err := contextToken.Marshal()
	if err != nil {
		return "", fmt.Errorf("failed to marshal SPNEGO token: %w", err)
	}

	// Encode token to base64
	tokenBase64 := base64.StdEncoding.EncodeToString(tokenBytes)

	// Cache the token
	s.mu.Lock()
	s.tokenCache[proxyHost] = tokenBase64
	s.mu.Unlock()

	return tokenBase64, nil
}

// ConfigureTransport configures the transport for SPNEGO authentication
func (s *SPNEGOProxyAuth) ConfigureTransport(transport http.RoundTripper, proxyURL *url.URL) (http.RoundTripper, error) {
	// Initialize Kerberos client if not already done
	if s.krb5Client == nil {
		if err := s.Initialize(s.config); err != nil {
			return nil, fmt.Errorf("failed to initialize SPNEGO: %w", err)
		}
	}

	// Wrap transport to add Proxy-Authorization headers
	return NewProxyAuthRoundTripper(transport, s, proxyURL), nil
}

// ProxyAuthRoundTripper wraps an HTTP RoundTripper to add Proxy-Authorization headers
type ProxyAuthRoundTripper struct {
	rt        http.RoundTripper
	spnegoAuth *SPNEGOProxyAuth
	proxyURL  *url.URL
}

// NewProxyAuthRoundTripper creates a new RoundTripper that adds Proxy-Authorization headers
func NewProxyAuthRoundTripper(rt http.RoundTripper, spnegoAuth *SPNEGOProxyAuth, proxyURL *url.URL) *ProxyAuthRoundTripper {
	return &ProxyAuthRoundTripper{
		rt:        rt,
		spnegoAuth: spnegoAuth,
		proxyURL:  proxyURL,
	}
}

// RoundTrip implements http.RoundTripper
func (p *ProxyAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check if this request is going through a proxy
	if req.URL.Scheme != "" && p.proxyURL != nil && p.spnegoAuth != nil {
		// Get SPNEGO token for the proxy
		proxyHost := p.proxyURL.Hostname()
		token, err := p.spnegoAuth.GetSPNEGOToken(proxyHost)
		if err != nil {
			// Log error but continue - might fall back to basic auth
			log.Printf("Warning: Failed to get SPNEGO token: %v", err)
		} else {
			// Add Proxy-Authorization header
			req.Header.Set("Proxy-Authorization", fmt.Sprintf("Negotiate %s", token))
		}
	}

	resp, err := p.rt.RoundTrip(req)

	// Handle 407 Proxy Authentication Required - retry with new token
	if err == nil && resp != nil && resp.StatusCode == http.StatusProxyAuthRequired && p.spnegoAuth != nil {
		resp.Body.Close()

		// Clear cached token and get a new one
		proxyHost := p.proxyURL.Hostname()
		p.spnegoAuth.mu.Lock()
		delete(p.spnegoAuth.tokenCache, proxyHost)
		p.spnegoAuth.mu.Unlock()

		// Get new token
		token, err := p.spnegoAuth.GetSPNEGOToken(proxyHost)
		if err == nil {
			// Retry the request with new token
			req.Header.Set("Proxy-Authorization", fmt.Sprintf("Negotiate %s", token))
			return p.rt.RoundTrip(req)
		}
	}

	return resp, err
}

// configPath returns the path to krb5.conf
func configPath() string {
	// Try common locations for krb5.conf
	paths := []string{
		os.Getenv("KRB5_CONFIG"),
		"/etc/krb5.conf",
	}

	for _, p := range paths {
		if p != "" {
			// Check if file exists
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	// Return default path
	return "/etc/krb5.conf"
}

