package download_clients

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/spnego"
)

// ProxyAuthConfig holds configuration for proxy authentication
type ProxyAuthConfig struct {
	Username string
	Password string
	Domain   string
}

// ProxyAuthTransport wraps an HTTP transport to add Proxy-Authorization header
type ProxyAuthTransport struct {
	Transport  http.RoundTripper
	config     *ProxyAuthConfig
	krb5Client *client.Client
}

// NewProxyAuthTransport creates a new transport with proxy authentication
func NewProxyAuthTransport(baseTransport http.RoundTripper, proxyConfig *ProxyAuthConfig) (*ProxyAuthTransport, error) {
	if proxyConfig == nil || proxyConfig.Username == "" {
		// No proxy authentication configured
		return &ProxyAuthTransport{
			Transport: baseTransport,
		}, nil
	}

	// Load Kerberos configuration
	krb5Config, err := config.Load(configPath())
	if err != nil {
		return nil, fmt.Errorf("failed to load Kerberos config: %w", err)
	}

	// Create Kerberos client
	krb5Client := client.NewWithPassword(
		proxyConfig.Username,
		proxyConfig.Domain,
		proxyConfig.Password,
		krb5Config,
		client.DisablePAFXFAST(true),
	)

	// Login to Kerberos
	err = krb5Client.Login()
	if err != nil {
		return nil, fmt.Errorf("failed to login to Kerberos: %w", err)
	}

	return &ProxyAuthTransport{
		Transport:  baseTransport,
		config:     proxyConfig,
		krb5Client: krb5Client,
	}, nil
}

// RoundTrip implements http.RoundTripper interface
func (t *ProxyAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// If proxy authentication is configured, add the header
	if t.config != nil && t.config.Username != "" && t.krb5Client != nil {
		// Get proxy hostname from environment or request
		proxyHost := getProxyHost(req)
		
		// Construct SPN for proxy (HTTP/proxy-hostname@REALM)
		spn := fmt.Sprintf("HTTP/%s@%s", proxyHost, strings.ToUpper(t.config.Domain))
		
		// Generate SPNEGO token
		spnegoClient := spnego.SPNEGOClient(t.krb5Client, spn)
		token, err := spnegoClient.InitSecContext()
		if err != nil {
			return nil, fmt.Errorf("failed to generate SPNEGO token: %w", err)
		}

		tokenBytes, err := token.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal SPNEGO token: %w", err)
		}

		// Base64 encode and set header
		tokenBase64 := base64.StdEncoding.EncodeToString(tokenBytes)
		req.Header.Set("Proxy-Authorization", fmt.Sprintf("Negotiate %s", tokenBase64))
	}

	// Use the underlying transport
	if t.Transport != nil {
		return t.Transport.RoundTrip(req)
	}
	return http.DefaultTransport.RoundTrip(req)
}

// getProxyHost extracts the proxy hostname from environment variables or request
func getProxyHost(req *http.Request) string {
	// Try to get proxy from environment variables
	proxyURL := os.Getenv("HTTP_PROXY")
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTPS_PROXY")
	}
	if proxyURL == "" {
		proxyURL = os.Getenv("http_proxy")
	}
	if proxyURL == "" {
		proxyURL = os.Getenv("https_proxy")
	}

	if proxyURL != "" {
		parsedURL, err := url.Parse(proxyURL)
		if err == nil && parsedURL.Host != "" {
			host := parsedURL.Hostname()
			if host != "" {
				return host
			}
		}
	}

	// Fallback: try to extract from request URL if it's a proxy request
	if req.URL != nil && req.URL.Host != "" {
		return req.URL.Hostname()
	}

	// Default fallback
	return "proxy"
}

// configPath returns the path to Kerberos configuration file
func configPath() string {
	// Try common locations for krb5.conf
	paths := []string{
		"/etc/krb5.conf",
		"/usr/local/etc/krb5.conf",
		"/opt/homebrew/etc/krb5.conf",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return "/etc/krb5.conf"
}

