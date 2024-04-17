package network

import (
	"net/http"
	"time"
)

type UnauthenticatedClient struct {
	target string
	client *http.Client
}

func NewUnauthenticatedClient(target string, insecureSkipVerify bool, caCert string, connectTimeout time.Duration, requestTimeout time.Duration) (UnauthenticatedClient, error) {
	client, err := newHTTPClient(insecureSkipVerify, caCert, requestTimeout, connectTimeout)
	if err != nil {
		return UnauthenticatedClient{}, err
	}

	return UnauthenticatedClient{
		target: target,
		client: client,
	}, nil
}

func (c UnauthenticatedClient) Do(request *http.Request) (*http.Response, error) {
	targetURL, err := parseURL(c.target)
	if err != nil {
		return nil, err
	}

	request.URL.Scheme = targetURL.Scheme
	request.URL.Host = targetURL.Host

	return c.client.Do(request)
}
