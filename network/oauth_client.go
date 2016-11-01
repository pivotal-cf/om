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

type OAuthClient struct {
	oauthConfig *oauth2.Config
	context     context.Context
	username    string
	password    string
	target      string
	client      *http.Client
	timeout     time.Duration
}

func NewOAuthClient(target, username, password string, insecureSkipVerify bool, requestTimeout time.Duration) (OAuthClient, error) {
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
		},
	}

	insecureContext := context.Background()
	insecureContext = context.WithValue(insecureContext, oauth2.HTTPClient, httpclient)

	return OAuthClient{
		oauthConfig: conf,
		context:     insecureContext,
		username:    username,
		password:    password,
		target:      target,
		client:      nil,
		timeout:     requestTimeout,
	}, nil
}

func (oc OAuthClient) Do(request *http.Request) (*http.Response, error) {
	token, err := oc.oauthConfig.PasswordCredentialsToken(oc.context, oc.username, oc.password)
	if err != nil {
		return nil, fmt.Errorf("token could not be retrieved from target url: %s", err)
	}

	client := oc.oauthConfig.Client(oc.context, token)
	client.Timeout =  oc.timeout

	targetURL, err := url.Parse(oc.target)
	if err != nil {
		return nil, fmt.Errorf("could not parse target url: %s", err)
	}

	request.URL.Scheme = targetURL.Scheme
	request.URL.Host = targetURL.Host

	return client.Do(request)
}

func (oc OAuthClient) RoundTrip(request *http.Request) (*http.Response, error) {
	token, err := oc.oauthConfig.PasswordCredentialsToken(oc.context, oc.username, oc.password)
	if err != nil {
		return nil, fmt.Errorf("token could not be retrieved from target url: %s", err)
	}

	client := oc.oauthConfig.Client(oc.context, token)
	client.Timeout =  oc.timeout

	targetURL, err := url.Parse(oc.target)
	if err != nil {
		return nil, fmt.Errorf("could not parse target url: %s", err)
	}

	request.URL.Scheme = targetURL.Scheme
	request.URL.Host = targetURL.Host

	return client.Transport.RoundTrip(request)
}
