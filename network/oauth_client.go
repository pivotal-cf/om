package network

import (
	"fmt"
	"github.com/cloudfoundry-community/go-uaa"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OAuthClient struct {
	client             *http.Client
	clientID           string
	clientSecret       string
	insecureSkipVerify bool
	password           string
	target             string
	token              *oauth2.Token
	username           string
}

func NewOAuthClient(
	target, username, password string,
	clientID, clientSecret string,
	insecureSkipVerify bool,
	caCert string,
	connectTimeout time.Duration,
	requestTimeout time.Duration,
) (*OAuthClient, error) {
	httpclient, err := newHTTPClient(insecureSkipVerify, caCert, requestTimeout, connectTimeout)
	if err != nil {
		return nil, err
	}

	return &OAuthClient{
		client:             httpclient,
		clientID:           clientID,
		clientSecret:       clientSecret,
		insecureSkipVerify: insecureSkipVerify,
		password:           password,
		target:             target,
		username:           username,
	}, nil
}

func (oc *OAuthClient) Do(request *http.Request) (*http.Response, error) {
	token := oc.token
	client := oc.client
	target := oc.target

	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}

	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("could not parse target url: %s", err)
	}

	targetURL.Path = "/uaa"

	request.URL.Scheme = targetURL.Scheme
	request.URL.Host = targetURL.Host

	if token != nil && token.Valid() {
		request.Header.Set(
			"Authorization",
			fmt.Sprintf("Bearer %s", token.AccessToken),
		)
		return client.Do(request)
	}

	options := []uaa.Option{
		uaa.WithSkipSSLValidation(oc.insecureSkipVerify),
		uaa.WithClient(client),
	}

	var authOption uaa.AuthenticationOption

	if oc.username != "" && oc.password != "" {
		authOption = uaa.WithPasswordCredentials(
			"opsman",
			"",
			oc.username,
			oc.password,
			uaa.OpaqueToken,
		)
	} else {
		authOption = uaa.WithClientCredentials(
			oc.clientID,
			oc.clientSecret,
			uaa.OpaqueToken,
		)
	}

	api, err := uaa.New(
		targetURL.String(),
		authOption,
		options...,
	)
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
		return nil, fmt.Errorf("token could not be retrieved from target url: %w", err)
	}

	request.Header.Set(
		"Authorization",
		fmt.Sprintf("Bearer %s", token.AccessToken),
	)

	oc.token = token

	return client.Do(request)
}
