package network

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	target             string
	token              *oauth2.Token
	username           string
	connectTimeout     time.Duration
	requestTimeout     time.Duration
}

func NewOAuthClient(
	target, username, password string,
	clientID, clientSecret string,
	insecureSkipVerify bool,
	caCert string,
	connectTimeout time.Duration,
	requestTimeout time.Duration,
) (*OAuthClient, error) {
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
	}, nil
}

func (oc *OAuthClient) Do(request *http.Request) (*http.Response, error) {
	token := oc.token
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

	client, err := newHTTPClient(
		oc.insecureSkipVerify,
		oc.caCert,
		oc.requestTimeout,
		oc.connectTimeout,
	)

	if err != nil {
		return nil, err
	}

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
	var targetString string
	if len(os.Getenv("OM_UAA_HOST")) != 0 {
		uaaTargetURL, err := url.Parse(os.Getenv("OM_UAA_HOST"))
		if err != nil {
			return nil, fmt.Errorf("could not reset UAA Host: %w", err)
		}
		targetString = uaaTargetURL.String()
	} else {
		targetString = targetURL.String()
	}

	api, err := uaa.New(
		targetString,
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
