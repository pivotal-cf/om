package network

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

type AuthenticatedClient struct {
	target string
	client *http.Client
}

func NewAuthenticatedClient(target, username, password string, insecureSkipVerify bool) (AuthenticatedClient, error) {
	conf := &oauth2.Config{
		ClientID:     "opsman",
		ClientSecret: "",
		Endpoint: oauth2.Endpoint{
			TokenURL: fmt.Sprintf("%s/uaa/oauth/token", target),
		},
	}

	httpclient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			ResponseHeaderTimeout: 15 * time.Minute,
		},
		Timeout: 30 * time.Minute,
	}

	insecureContext := context.Background()
	insecureContext = context.WithValue(insecureContext, oauth2.HTTPClient, httpclient)

	token, err := conf.PasswordCredentialsToken(insecureContext, username, password)
	if err != nil {
		return AuthenticatedClient{}, fmt.Errorf("token could not be retrieved from target url: %s", err)
	}

	return AuthenticatedClient{
		target: target,
		client: conf.Client(insecureContext, token),
	}, nil
}

func (ac AuthenticatedClient) Do(request *http.Request) (*http.Response, error) {
	targetURL, err := url.Parse(ac.target)
	if err != nil {
		return nil, fmt.Errorf("could not parse target url: %s", err)
	}

	request.URL.Scheme = targetURL.Scheme
	request.URL.Host = targetURL.Host

	return ac.client.Do(request)
}

func (ac AuthenticatedClient) RoundTrip(request *http.Request) (*http.Response, error) {
	targetURL, err := url.Parse(ac.target)
	if err != nil {
		return nil, fmt.Errorf("could not parse target url: %s", err)
	}

	request.URL.Scheme = targetURL.Scheme
	request.URL.Host = targetURL.Host

	return ac.client.Transport.RoundTrip(request)
}
