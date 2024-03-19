package network

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/cloudfoundry-community/go-uaa"
	"golang.org/x/oauth2"
)

// CachedResolver caches DNS lookups
type CachedResolver struct {
	cache      map[string]string
	cacheMutex sync.Mutex
	ttl        time.Duration
}

func NewCachedResolver(ttl time.Duration) *CachedResolver {
	return &CachedResolver{
		cache: make(map[string]string),
		ttl:   ttl,
	}
}

func (r *CachedResolver) Resolve(host string) (string, error) {
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()

	if ip, found := r.cache[host]; found {
		return ip, nil
	}

	ips, err := net.LookupHost(host)
	if err != nil {
		return "", err
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no IPs found for host: %s", host)
	}

	r.cache[host] = ips[0]

	go func() {
		time.Sleep(r.ttl)
		r.cacheMutex.Lock()
		delete(r.cache, host)
		r.cacheMutex.Unlock()
	}()

	return ips[0], nil
}

type OAuthClient struct {
	caCert             string
	clientID           string
	clientSecret       string
	insecureSkipVerify bool
	password           string
	target             string
	token              *oauth2.Token
	username           string
	connectTimeout     time.Duration
	requestTimeout     time.Duration
	resolver           *CachedResolver // Added resolver
}

func NewOAuthClient(
	target, username, password string,
	clientID, clientSecret string,
	insecureSkipVerify bool,
	caCert string,
	connectTimeout, requestTimeout time.Duration,
) (*OAuthClient, error) {
	resolver := NewCachedResolver(5 * time.Minute) // Example TTL
	return &OAuthClient{
		caCert:             caCert,
		clientID:           clientID,
		clientSecret:       clientSecret,
		insecureSkipVerify: insecureSkipVerify,
		password:           password,
		target:             target,
		username:           username,
		connectTimeout:     connectTimeout,
		requestTimeout:     requestTimeout,
		resolver:           resolver, // Initialize resolver
	}, nil
}

func (oc *OAuthClient) customDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	separator := strings.LastIndex(addr, ":")
	host, port := addr[:separator], addr[separator+1:]

	resolvedIP, err := oc.resolver.Resolve(host)
	if err != nil {
		return nil, err
	}

	return net.Dial(network, resolvedIP+":"+port)
}

func (oc *OAuthClient) Do(request *http.Request) (*http.Response, error) {
	token := oc.token
	target := oc.target

	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}

	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("could not parse target URL: %w", err)
	}

	targetURL.Path = "/uaa"

	request.URL.Scheme = targetURL.Scheme
	request.URL.Host = targetURL.Host

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: oc.customDialContext,
		},
		Timeout: oc.requestTimeout,
	}

	if token != nil && token.Valid() {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
		return client.Do(request)
	}

	options := []uaa.Option{
		uaa.WithSkipSSLValidation(oc.insecureSkipVerify),
		uaa.WithClient(client),
	}

	var authOption uaa.AuthenticationOption

	if oc.username != "" && oc.password != "" {
		authOption = uaa.WithPasswordCredentials("opsman", "", oc.username, oc.password, uaa.JSONWebToken)
	} else {
		authOption = uaa.WithClientCredentials(oc.clientID, oc.clientSecret, uaa.JSONWebToken)
	}

	api, err := uaa.New(targetURL.String(), authOption, options...)
	if err != nil {
		return nil, fmt.Errorf("could not init UAA client: %w", err)
	}

	for i := 0; i <= 2; i++ {
		token, err = api.Token(request.Context())
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("token could not be retrieved from target URL: %w", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	oc.token = token

	return client.Do(request)
}
