package network

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type UnauthenticatedClient struct {
	target string
	client *http.Client
}

func NewUnauthenticatedClient(target string, insecureSkipVerify bool, requestTimeout time.Duration) UnauthenticatedClient {
	return UnauthenticatedClient{
		target: target,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: insecureSkipVerify,
				},
				Dial: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
			},
			Timeout: requestTimeout,
		},
	}
}

func (c UnauthenticatedClient) Do(request *http.Request) (*http.Response, error) {
	targetURL, err := url.Parse(c.target)
	if err != nil {
		return nil, fmt.Errorf("could not parse target url: %s", err)
	}

	request.URL.Scheme = targetURL.Scheme
	request.URL.Host = targetURL.Host

	return c.client.Do(request)
}

func (c UnauthenticatedClient) RoundTrip(request *http.Request) (*http.Response, error) {
	targetURL, err := url.Parse(c.target)
	if err != nil {
		return nil, fmt.Errorf("could not parse target url: %s", err)
	}

	request.URL.Scheme = targetURL.Scheme
	request.URL.Host = targetURL.Host

	return c.client.Transport.RoundTrip(request)
}
