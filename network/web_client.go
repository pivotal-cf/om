package network

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"time"
)

type WebClient struct {
	target     string
	username   string
	password   string
	HTTPClient *http.Client
}

func NewWebClient(target, username, password string, insecureSkipVerify bool, requestTimeout time.Duration) (*WebClient, error) {
	cookieJar, _ := cookiejar.New(nil)
	client := &WebClient{
		target:   target,
		username: username,
		password: password,
		HTTPClient: &http.Client{
			Jar: cookieJar,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: insecureSkipVerify,
				},
			},
			Timeout: requestTimeout,
		},
	}
	return client, client.authenticate()
}

func (c WebClient) authenticate() error {
	var err error
	var resp *http.Response
	var XUaaCsrf, UAAToken string
	if resp, err = c.HTTPClient.Get(fmt.Sprintf("%s/auth/cloudfoundry", c.target)); err == nil {
		defer resp.Body.Close()
		if err = c.handleResponse(resp); err != nil {
			return err
		}
		cookies := resp.Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "X-Uaa-Csrf" {
				XUaaCsrf = cookie.Value
			}
		}
		if XUaaCsrf == "" {
			return fmt.Errorf("X-Uaa-Csrf cookie not found in response")
		}
		data := url.Values{}
		data.Set("username", c.username)
		data.Add("password", c.password)
		data.Add("X-Uaa-Csrf", XUaaCsrf)
		if resp, err = c.HTTPClient.PostForm(fmt.Sprintf("%s/uaa/login.do", c.target), data); err == nil {
			defer resp.Body.Close()
			if err = c.handleResponse(resp); err != nil {
				return err
			}
			hostURL, _ := url.Parse(c.target)
			cookies := c.HTTPClient.Jar.Cookies(hostURL)
			for _, cookie := range cookies {
				if cookie.Name == "uaa_access_token" {
					UAAToken = cookie.Value
				}
			}
			if UAAToken == "" {
				out, _ := httputil.DumpResponse(resp, true)
				return fmt.Errorf("Authentication failed %s", out)
			}
		} else {
			return err
		}
	} else {
		return err
	}

	return nil
}

func (c WebClient) handleResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return fmt.Errorf("request failed: unexpected response: %s", err)
		}
		return fmt.Errorf("request failed: unexpected response:\n%s", out)
	}
	return nil
}

func (c WebClient) Do(request *http.Request) (*http.Response, error) {
	targetURL, err := url.Parse(c.target)
	if err != nil {
		return nil, fmt.Errorf("could not parse target url: %s", err)
	}

	request.URL.Scheme = targetURL.Scheme
	request.URL.Host = targetURL.Host

	return c.HTTPClient.Do(request)
}

func (c WebClient) RoundTrip(request *http.Request) (*http.Response, error) {
	targetURL, err := url.Parse(c.target)
	if err != nil {
		return nil, fmt.Errorf("could not parse target url: %s", err)
	}

	request.URL.Scheme = targetURL.Scheme
	request.URL.Host = targetURL.Host

	return c.HTTPClient.Transport.RoundTrip(request)
}
