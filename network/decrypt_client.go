package network

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type DecryptClient struct {
	unauthedClient httpClient
	authedClient   httpClient

	decryptionPassphrase string
	writer               io.Writer
}

func NewDecryptClient(authdClient httpClient, unAuthedClient httpClient, decryptionPassphrase string, writer io.Writer) DecryptClient {
	return DecryptClient{
		authedClient:         authdClient,
		unauthedClient:       unAuthedClient,
		decryptionPassphrase: decryptionPassphrase,
		writer:               writer,
	}
}

func (c DecryptClient) Do(request *http.Request) (*http.Response, error) {
	if err := c.decrypt(); err != nil {
		return nil, err
	}

	return c.authedClient.Do(request)
}

func (c DecryptClient) decrypt() error {
	const unlock = "/api/v0/unlock"
	request, err := http.NewRequest("PUT", unlock, bytes.NewBufferString(fmt.Sprintf("{\"passphrase\": \"%s\"}", c.decryptionPassphrase)))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	resp, err := c.unauthedClient.Do(request)
	if err != nil {
		return fmt.Errorf("could not make api request to unlock endpoint: %s", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not unlock ops manager, check if the decryption passphrase is correct")
	}

	return c.waitUntilAvailable()
}

func (c DecryptClient) waitUntilAvailable() error {
	var status = "unknown"
	var err error

	var trial = 1
	for status != "complete" {
		if trial == 2 {
			c.writer.Write([]byte("Waiting for Ops Manager's auth systems to start. This may take a few minutes...\n"))
		}
		status, err = c.checkAvailability()
		if err != nil {
			return fmt.Errorf("could not check Ops Manager Status: %s", err)
		}
		trial += 1
	}

	return nil
}

func (c DecryptClient) checkAvailability() (string, error) {
	// the below code is copied from api/setup_service. Don't really want to import api here as it will break
	// dag dependency graph. It's probably make sense to separate that logic from api package into a standalone one
	// to just maintain the dependencies.

	request, err := http.NewRequest("GET", "/login/ensure_availability", nil)
	if err != nil {
		return "", err
	}

	response, err := c.unauthedClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("could not make request round trip: %s", err)
	}

	defer response.Body.Close()

	status := "unknown"
	switch {
	case response.StatusCode == http.StatusFound:
		location, err := url.Parse(response.Header.Get("Location"))
		if err != nil {
			return "", fmt.Errorf("could not parse redirect url: %s", err)
		}

		if location.Path == "/setup" {
			status = "unstarted"
		} else if location.Path == "/auth/cloudfoundry" {
			status = "complete"
		} else {
			return "", fmt.Errorf("Unexpected redirect location: %s", location.Path)
		}

	case response.StatusCode == http.StatusOK:
		respBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", err
		}

		if strings.Contains(string(respBody), "Waiting for authentication system to start...") {
			status = "pending"
		} else {
			return "", fmt.Errorf("Received OK with an unexpected body: %s", string(respBody))
		}

	default:
		return "", fmt.Errorf("Unexpected response code: %d %s", response.StatusCode, http.StatusText(response.StatusCode))
	}

	return status, nil
}
