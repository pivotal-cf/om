package network

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type UnauthenticatedClient struct {
	target string
	client *http.Client
}

func NewUnauthenticatedClient(target string, insecureSkipVerify bool, caCert string, connectTimeout time.Duration, requestTimeout time.Duration) (UnauthenticatedClient, error) {
	client, err := newHTTPClient(insecureSkipVerify, caCert, requestTimeout, connectTimeout)
	if err != nil {
		return UnauthenticatedClient{}, nil
	}

	return UnauthenticatedClient{
		target: target,
		client: client,
	}, nil
}

func (c UnauthenticatedClient) Do(request *http.Request) (*http.Response, error) {
	candidateURL := c.target
	if !strings.Contains(candidateURL, "//") {
		candidateURL = fmt.Sprintf("//%s", candidateURL)
	}

	targetURL, err := url.Parse(candidateURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse target url: %s", err)
	}

	if targetURL.Scheme == "" {
		targetURL.Scheme = "https"
	}

	if targetURL.Host == "" {
		return nil, fmt.Errorf("target flag is required. Run `om help` for more info.")
	}

	request.URL.Scheme = targetURL.Scheme
	request.URL.Host = targetURL.Host

	return c.client.Do(request)
}
