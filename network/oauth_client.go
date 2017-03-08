package network

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

type OAuthClient struct {
	oauthConfig *oauth2.Config
	jar         *cookiejar.Jar
	context     context.Context
	username    string
	password    string
	target      string
	timeout     time.Duration
}

func NewOAuthClient(target, username, password string, insecureSkipVerify bool, includeCookies bool, requestTimeout time.Duration) (OAuthClient, error) {
	conf := &oauth2.Config{
		ClientID:     "opsman",
		ClientSecret: "",
		Endpoint: oauth2.Endpoint{
			TokenURL: fmt.Sprintf("%s/uaa/oauth/token", target),
		},
	}

	httpclient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureSkipVerify,
			},
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
		},
	}

	var jar *cookiejar.Jar
	if includeCookies {
		var err error
		jar, err = cookiejar.New(nil)
		if err != nil {
			return OAuthClient{}, fmt.Errorf("could not create cookie jar")
		}
	}

	insecureContext := context.Background()
	insecureContext = context.WithValue(insecureContext, oauth2.HTTPClient, httpclient)

	return OAuthClient{
		oauthConfig: conf,
		jar:         jar,
		context:     insecureContext,
		username:    username,
		password:    password,
		target:      target,
		timeout:     requestTimeout,
	}, nil
}

func (oc OAuthClient) Do(request *http.Request) (*http.Response, error) {
	if oc.target == "" {
		return nil, fmt.Errorf("target flag is required. Run `om help` for more info.")
	}

	token, err := retrieveTokenWithRetry(oc.oauthConfig, oc.context, oc.username, oc.password)
	if err != nil {
		return nil, err
	}

	client := oc.oauthConfig.Client(oc.context, token)
	client.Timeout = oc.timeout

	if oc.jar != nil {
		client.Jar = oc.jar
	}

	targetURL, err := url.Parse(oc.target)
	if err != nil {
		return nil, fmt.Errorf("could not parse target url: %s", err)
	}

	request.URL.Scheme = targetURL.Scheme
	request.URL.Host = targetURL.Host

	// we only want to retry non-modifying actions
	if request.Method == "GET" {
		return httpResponseWithRetry(client, request)
	}

	return client.Do(request)
}

func retrieveTokenWithRetry(config *oauth2.Config, ctx context.Context, username, password string) (*oauth2.Token, error) {
retry:
	token, err := config.PasswordCredentialsToken(ctx, username, password)
	if canRetry(err) {
		goto retry
	}

	if err != nil {
		return nil, fmt.Errorf("token could not be retrieved from target url: %s", err)
	}

	return token, err
}

func httpResponseWithRetry(client *http.Client, request *http.Request) (*http.Response, error) {
retry:
	resp, err := client.Do(request)
	if client.Timeout == 0 {
		if canRetry(err) {
			goto retry
		}
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func canRetry(err error) bool {
	if err != nil {
		if ne, ok := err.(net.Error); ok {
			if ne.Temporary() {
				return true
			}
		}

		if err == io.EOF {
			return true
		}

		return false
	}

	return false
}
