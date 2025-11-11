package download_clients

import (
	"fmt"
	"net/http"
	"net/url"
)

// BasicProxyAuth implements basic HTTP proxy authentication (username/password in URL)
type BasicProxyAuth struct{}

// NewBasicProxyAuth creates a new basic proxy authenticator
func NewBasicProxyAuth() *BasicProxyAuth {
	return &BasicProxyAuth{}
}

// Type returns the authentication type
func (b *BasicProxyAuth) Type() ProxyAuthType {
	return ProxyAuthTypeBasic
}

// Validate checks if the configuration is valid for basic auth
func (b *BasicProxyAuth) Validate(config ProxyAuthConfig) error {
	if config.URL == "" {
		return fmt.Errorf("proxy URL is required")
	}
	// Basic auth doesn't require username/password, but it's recommended
	return nil
}

// ConfigureTransport configures the transport for basic authentication
// Basic auth embeds credentials in the proxy URL, so we just need to set environment variables
func (b *BasicProxyAuth) ConfigureTransport(transport http.RoundTripper, proxyURL *url.URL) (http.RoundTripper, error) {
	// For basic auth, credentials are embedded in the URL via environment variables
	// The transport doesn't need wrapping, just return the original
	return transport, nil
}

// BuildProxyURLWithAuth constructs a proxy URL with authentication credentials
func BuildProxyURLWithAuth(proxyURL, username, password string) string {
	if username == "" && password == "" {
		return proxyURL
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		// If URL parsing fails, return original URL
		return proxyURL
	}

	if username != "" {
		if password != "" {
			parsedURL.User = url.UserPassword(username, password)
		} else {
			parsedURL.User = url.User(username)
		}
	}

	return parsedURL.String()
}


