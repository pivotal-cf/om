package network

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cloudfoundry-community/go-uaa"
	"golang.org/x/oauth2"
)

type OAuthClient struct {
	caCert             string
	clientID           string
	clientSecret       string
	insecureSkipVerify bool
	password           string
	target             *url.URL
	uaaTarget          *url.URL
	token              *oauth2.Token
	username           string
	connectTimeout     time.Duration
	requestTimeout     time.Duration
}

func NewOAuthClient(
	uaaTarget, opsmanTarget *url.URL,
	username, password string,
	clientID, clientSecret string,
	insecureSkipVerify bool,
	caCert string,
	connectTimeout time.Duration,
	requestTimeout time.Duration,
) (*OAuthClient, error) {
	if uaaTarget == nil {
		return nil, errors.New("expected a non-nil UAA target")
	}
	if opsmanTarget == nil {
		return nil, errors.New("expected a non-nil target")
	}

	return &OAuthClient{
		caCert:             caCert,
		clientID:           clientID,
		clientSecret:       clientSecret,
		insecureSkipVerify: insecureSkipVerify,
		password:           password,
		uaaTarget:          uaaTarget,
		target:             opsmanTarget,
		username:           username,
		connectTimeout:     connectTimeout,
		requestTimeout:     requestTimeout,
	}, nil
}

func (oc *OAuthClient) Do(request *http.Request) (*http.Response, error) {
	request.URL.Scheme = oc.target.Scheme
	request.URL.Host = oc.target.Host

	client, err := newHTTPClient(
		oc.insecureSkipVerify,
		oc.caCert,
		oc.requestTimeout,
		oc.connectTimeout,
	)
	if err != nil {
		return nil, err
	}

	token := oc.token
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
			uaa.JSONWebToken,
		)
	} else {
		authOption = uaa.WithClientCredentials(
			oc.clientID,
			oc.clientSecret,
			uaa.JSONWebToken,
		)
	}

	api, err := uaa.New(
		oc.uaaTarget.String(),
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
		if oauthErr, ok := err.(uaa.RequestError); ok {
			return nil, fmt.Errorf("token could not be retrieved from target: %w: %s",
				err, string(oauthErr.ErrorResponse))
		}
		return nil, fmt.Errorf("token could not be retrieved from target : %w", err)
	}

	request.Header.Set(
		"Authorization",
		fmt.Sprintf("Bearer %s", token.AccessToken),
	)

	oc.token = token

	return client.Do(request)
}
