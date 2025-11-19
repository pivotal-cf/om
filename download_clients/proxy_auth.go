package download_clients

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/spnego"
)

// ProxyAuthConfig holds configuration for proxy authentication
type ProxyAuthConfig struct {
	AuthType string // "spnego", "basic", etc.
	Username string
	Password string
	Domain   string
	ProxyURL string      // Proxy URL (e.g., http://proxy.example.com:8080)
	Logger   *log.Logger // Optional logger for debugging (nil = no logging)
}

// ProxyAuthenticator is the interface for proxy authentication mechanisms
type ProxyAuthenticator interface {
	// Authenticate adds the appropriate Proxy-Authorization header to the request
	Authenticate(req *http.Request) error
}

// ProxyAuthenticatorFactory creates a new ProxyAuthenticator from a config
type ProxyAuthenticatorFactory func(config *ProxyAuthConfig) (ProxyAuthenticator, error)

// proxyAuthRegistry holds registered authentication mechanisms
// This is only written to during init() and read from afterwards, so no mutex needed
var proxyAuthRegistry = make(map[string]ProxyAuthenticatorFactory)

// RegisterProxyAuth registers a new proxy authentication mechanism
func RegisterProxyAuth(name string, factory ProxyAuthenticatorFactory) {
	proxyAuthRegistry[strings.ToLower(name)] = factory
}

// GetProxyAuthFactory retrieves a factory for the given auth type
func GetProxyAuthFactory(authType string) (ProxyAuthenticatorFactory, error) {
	factory, exists := proxyAuthRegistry[strings.ToLower(authType)]
	if !exists {
		return nil, fmt.Errorf("unknown proxy authentication type: %s", authType)
	}
	return factory, nil
}

// logf is a helper function to log messages if logger is available
func logf(logger *log.Logger, format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[proxy-auth] "+format, args...)
	}
}

// ProxyAuthTransport wraps an HTTP transport to add Proxy-Authorization header
type ProxyAuthTransport struct {
	Transport     http.RoundTripper
	authenticator ProxyAuthenticator
	logger        *log.Logger
}

// NewProxyAuthTransport creates a new transport with proxy authentication
func NewProxyAuthTransport(baseTransport http.RoundTripper, proxyConfig *ProxyAuthConfig) (*ProxyAuthTransport, error) {
	logf(proxyConfig.Logger, "Creating proxy auth transport")

	if proxyConfig == nil || proxyConfig.Username == "" {
		// No proxy authentication configured
		logf(proxyConfig.Logger, "No proxy authentication configured (no username provided)")
		return &ProxyAuthTransport{
			Transport: baseTransport,
		}, nil
	}

	// Default to "spnego" if not specified
	authType := proxyConfig.AuthType
	if authType == "" {
		authType = "spnego"
		logf(proxyConfig.Logger, "Auth type not specified, defaulting to 'spnego'")
	} else {
		logf(proxyConfig.Logger, "Using auth type: %s", authType)
	}

	// Get the authenticator factory from registry
	logf(proxyConfig.Logger, "Looking up authenticator factory for type: %s", authType)
	factory, err := GetProxyAuthFactory(authType)
	if err != nil {
		logf(proxyConfig.Logger, "Failed to get proxy auth factory: %v", err)
		return nil, fmt.Errorf("failed to get proxy auth factory: %w", err)
	}
	logf(proxyConfig.Logger, "Found authenticator factory for type: %s", authType)

	// Create the authenticator
	logf(proxyConfig.Logger, "Creating authenticator with username: %s, domain: %s, proxy: %s",
		proxyConfig.Username, proxyConfig.Domain, proxyConfig.ProxyURL)
	authenticator, err := factory(proxyConfig)
	if err != nil {
		logf(proxyConfig.Logger, "Failed to create proxy authenticator: %v", err)
		return nil, fmt.Errorf("failed to create proxy authenticator: %w", err)
	}
	logf(proxyConfig.Logger, "Successfully created proxy authenticator")

	return &ProxyAuthTransport{
		Transport:     baseTransport,
		authenticator: authenticator,
		logger:        proxyConfig.Logger,
	}, nil
}

// RoundTrip implements http.RoundTripper interface
func (t *ProxyAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// If proxy authentication is configured, add the header
	if t.authenticator != nil {
		logf(t.logger, "Authenticating request to %s %s", req.Method, req.URL.String())
		if err := t.authenticator.Authenticate(req); err != nil {
			logf(t.logger, "Failed to authenticate proxy request: %v", err)
			return nil, fmt.Errorf("failed to authenticate proxy request: %w", err)
		}
		logf(t.logger, "Successfully added Proxy-Authorization header, sending request")
	} else {
		logf(t.logger, "No authenticator configured, proceeding without proxy auth")
	}

	// Use the underlying transport
	if t.Transport != nil {
		resp, err := t.Transport.RoundTrip(req)
		if err != nil {
			logf(t.logger, "Request failed: %v", err)
		} else {
			logf(t.logger, "Request completed with status: %s", resp.Status)
		}
		return resp, err
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
	// Check KRB5_CONFIG environment variable first (works on all platforms)
	if krb5Config := os.Getenv("KRB5_CONFIG"); krb5Config != "" {
		if _, err := os.Stat(krb5Config); err == nil {
			return krb5Config
		}
	}

	// Platform-specific paths
	if runtime.GOOS == "windows" {
		// Windows paths (check both krb5.conf and krb5.ini)
		paths := []string{
			filepath.Join(os.Getenv("ProgramData"), "Kerberos", "krb5.conf"),
			filepath.Join(os.Getenv("ProgramData"), "MIT", "Kerberos5", "krb5.conf"),
			filepath.Join(os.Getenv("ProgramData"), "MIT", "Kerberos", "krb5.conf"),
			filepath.Join(os.Getenv("WINDIR"), "krb5.ini"),
			filepath.Join(os.Getenv("ProgramData"), "Kerberos", "krb5.ini"),
			filepath.Join(os.Getenv("ProgramData"), "MIT", "Kerberos5", "krb5.ini"),
			filepath.Join(os.Getenv("ProgramData"), "MIT", "Kerberos", "krb5.ini"),
		}

		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
		// Default Windows path
		return filepath.Join(os.Getenv("ProgramData"), "MIT", "Kerberos5", "krb5.conf")
	}

	// Unix/Linux/macOS paths
	paths := []string{
		"/Users/hyayi/work/tpi-telemetry-cli/test-integration/krb5-host.conf",
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

// NegotiateProxyAuthenticator implements SPNEGO/Kerberos proxy authentication
type NegotiateProxyAuthenticator struct {
	krb5Client *client.Client
	domain     string
	proxyURL   string
	logger     *log.Logger
}

// NewNegotiateProxyAuthenticator creates a new Negotiate authenticator
func NewNegotiateProxyAuthenticator(proxyConfig *ProxyAuthConfig) (ProxyAuthenticator, error) {
	logf(proxyConfig.Logger, "Initializing SPNEGO proxy authenticator")

	if proxyConfig == nil || proxyConfig.Username == "" || proxyConfig.Password == "" {
		logf(proxyConfig.Logger, "Missing required credentials for SPNEGO authentication")
		return nil, fmt.Errorf("username and password are required for SPNEGO authentication")
	}

	// Load Kerberos configuration
	krb5ConfigPath := configPath()
	logf(proxyConfig.Logger, "Loading Kerberos config from: %s", krb5ConfigPath)
	krb5Config, err := config.Load(krb5ConfigPath)
	if err != nil {
		logf(proxyConfig.Logger, "Failed to load Kerberos config: %v", err)
		return nil, fmt.Errorf("failed to load Kerberos config: %w", err)
	}
	logf(proxyConfig.Logger, "Successfully loaded Kerberos config")

	// Determine domain/realm: use provided domain or default realm from config
	domain := proxyConfig.Domain
	if domain == "" {
		// Extract default realm from krb5.conf
		domain = krb5Config.LibDefaults.DefaultRealm
		logf(proxyConfig.Logger, "No domain provided, using default realm from config: %s", domain)
		if domain == "" {
			logf(proxyConfig.Logger, "No default realm found in config and no domain provided")
			return nil, fmt.Errorf("domain/realm is required: either provide --proxy-domain or configure default_realm in krb5.conf")
		}
	} else {
		logf(proxyConfig.Logger, "Using provided domain: %s", domain)
	}

	// Create Kerberos client
	logf(proxyConfig.Logger, "Creating Kerberos client for user: %s@%s", proxyConfig.Username, domain)
	krb5Client := client.NewWithPassword(
		proxyConfig.Username,
		domain,
		proxyConfig.Password,
		krb5Config,
		client.DisablePAFXFAST(true),
	)

	// Login to Kerberos
	logf(proxyConfig.Logger, "Attempting Kerberos login...")
	err = krb5Client.Login()
	if err != nil {
		logf(proxyConfig.Logger, "Kerberos login failed: %v", err)
		return nil, fmt.Errorf("failed to login to Kerberos: %w", err)
	}
	logf(proxyConfig.Logger, "Kerberos login successful")

	return &NegotiateProxyAuthenticator{
		krb5Client: krb5Client,
		domain:     domain,
		proxyURL:   proxyConfig.ProxyURL,
		logger:     proxyConfig.Logger,
	}, nil
}

// Authenticate adds SPNEGO authentication header to the request
func (a *NegotiateProxyAuthenticator) Authenticate(req *http.Request) error {
	logf(a.logger, "Starting SPNEGO authentication for request: %s %s", req.Method, req.URL.String())

	// Get proxy hostname from config, environment, or request
	proxyHost := a.getProxyHost(req)
	logf(a.logger, "Resolved proxy hostname: %s", proxyHost)

	// Construct SPN for proxy (HTTP/proxy-hostname)
	// Note: gokrb5's SPNEGOClient adds the realm automatically from the krb5Client's realm,
	// so we only need to provide the service/hostname part without the realm
	spn := fmt.Sprintf("HTTP/%s", proxyHost)
	logf(a.logger, "Constructed SPN: %s (realm %s will be added by library)", spn, a.domain)

	// Generate SPNEGO token
	logf(a.logger, "Creating SPNEGO client for SPN: %s", spn)
	spnegoClient := spnego.SPNEGOClient(a.krb5Client, spn)

	logf(a.logger, "Initializing SPNEGO security context...")
	token, err := spnegoClient.InitSecContext()
	if err != nil {
		logf(a.logger, "Failed to generate SPNEGO token: %v", err)
		return fmt.Errorf("failed to generate SPNEGO token for SPN %s: %w. Ensure the SPN is registered in the KDC and the proxy hostname is correct", spn, err)
	}
	logf(a.logger, "SPNEGO token generated successfully")

	logf(a.logger, "Marshaling SPNEGO token...")
	tokenBytes, err := token.Marshal()
	if err != nil {
		logf(a.logger, "Failed to marshal SPNEGO token: %v", err)
		return fmt.Errorf("failed to marshal SPNEGO token: %w", err)
	}
	logf(a.logger, "SPNEGO token marshaled, size: %d bytes", len(tokenBytes))

	// Base64 encode and set header
	tokenBase64 := base64.StdEncoding.EncodeToString(tokenBytes)
	authHeader := fmt.Sprintf("Negotiate %s", tokenBase64)
	req.Header.Set("Proxy-Authorization", authHeader)
	logf(a.logger, "Set Proxy-Authorization header (token length: %d chars)", len(tokenBase64))

	return nil
}

// getProxyHost gets the proxy hostname from config, environment, or request
func (a *NegotiateProxyAuthenticator) getProxyHost(req *http.Request) string {
	// Priority 1: Use proxy URL from config if available
	if a.proxyURL != "" {
		logf(a.logger, "Using proxy URL from config: %s", a.proxyURL)
		parsedURL, err := url.Parse(a.proxyURL)
		if err == nil && parsedURL.Host != "" {
			host := parsedURL.Hostname()
			if host != "" {
				logf(a.logger, "Extracted hostname from config URL: %s", host)
				return host
			}
		}
		logf(a.logger, "Failed to parse proxy URL from config, falling back to environment/request")
	}

	// Priority 2: Fall back to standard getProxyHost function
	host := getProxyHost(req)
	logf(a.logger, "Using hostname from environment/request: %s", host)
	return host
}

// BasicProxyAuthenticator implements Basic proxy authentication
type BasicProxyAuthenticator struct {
	username string
	password string
	logger   *log.Logger
}

// NewBasicProxyAuthenticator creates a new Basic authenticator
func NewBasicProxyAuthenticator(proxyConfig *ProxyAuthConfig) (ProxyAuthenticator, error) {
	logf(proxyConfig.Logger, "Initializing Basic proxy authenticator")

	if proxyConfig == nil || proxyConfig.Username == "" || proxyConfig.Password == "" {
		logf(proxyConfig.Logger, "Missing required credentials for Basic authentication")
		return nil, fmt.Errorf("username and password are required for Basic authentication")
	}

	logf(proxyConfig.Logger, "Basic authenticator created for user: %s", proxyConfig.Username)
	return &BasicProxyAuthenticator{
		username: proxyConfig.Username,
		password: proxyConfig.Password,
		logger:   proxyConfig.Logger,
	}, nil
}

// Authenticate adds Basic authentication header to the request
func (a *BasicProxyAuthenticator) Authenticate(req *http.Request) error {
	logf(a.logger, "Starting Basic authentication for request: %s %s", req.Method, req.URL.String())

	credentials := fmt.Sprintf("%s:%s", a.username, a.password)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	authHeader := fmt.Sprintf("Basic %s", encoded)
	req.Header.Set("Proxy-Authorization", authHeader)

	logf(a.logger, "Set Proxy-Authorization header with Basic authentication")
	return nil
}

// init registers the built-in authentication mechanisms
func init() {
	RegisterProxyAuth("spnego", NewNegotiateProxyAuthenticator) // primary name
	RegisterProxyAuth("basic", NewBasicProxyAuthenticator)
}
