package network

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

type UnauthenticatedClient struct {
	target *url.URL
	client *http.Client
}

func NewUnauthenticatedClient(target *url.URL, insecureSkipVerify bool, caCert string, connectTimeout time.Duration, requestTimeout time.Duration) (UnauthenticatedClient, error) {
	if target == nil {
		return UnauthenticatedClient{}, errors.New("expected a non-nil target")
	}

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
	request.URL.Scheme = c.target.Scheme
	request.URL.Host = c.target.Host

	return c.client.Do(request)
}
