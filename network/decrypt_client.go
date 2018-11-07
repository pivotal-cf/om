package network

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type DecryptClient struct {
	unauthedClient httpClient
	authedClient   httpClient

	tried bool // to enforce only unlock once in the entire run

	decryptionPassphrase string
	writer               io.Writer
}

func NewDecryptClient(authdClient httpClient, unAuthedClient httpClient, decryptionPassphrase string, writer io.Writer) *DecryptClient {
	return &DecryptClient{
		authedClient:         authdClient,
		unauthedClient:       unAuthedClient,
		decryptionPassphrase: decryptionPassphrase,
		writer:               writer,
	}
}

func (c *DecryptClient) Do(request *http.Request) (*http.Response, error) {
	if !c.tried {
		if err := c.decrypt(); err != nil {
			return nil, err
		}
	}
	c.tried = true

	return c.authedClient.Do(request)
}

func (c *DecryptClient) decrypt() error {
	const unlock = "/api/v0/unlock"

	var err error
	var resp *http.Response

	for retries := 0; retries < 3; retries++ {
		var request *http.Request
		request, err = http.NewRequest("PUT", unlock, bytes.NewBufferString(fmt.Sprintf("{\"passphrase\": \"%s\"}", c.decryptionPassphrase)))
		if err != nil {
			return err
		}

		request.Header.Set("Content-Type", "application/json")
		resp, err = c.unauthedClient.Do(request)

		if err == nil {
			break
		}

		if e, ok := err.(net.Error); ok && e.Timeout() {
			//jitter on the retry
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			continue
		}

		if err != nil {
			return fmt.Errorf("could not make api request to unlock endpoint: %s", err)
		}
	}

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
	var trial = 1
	for {
		if trial == 2 {
			c.writer.Write([]byte("Waiting for Ops Manager's auth systems to start. This may take a few minutes...\n"))
		}

		err := c.checkAvailability()
		if err == nil {
			return nil
		}

		if !IsTemporary(err) {
			return fmt.Errorf("could not check Ops Manager Status: %s", err)
		}

		trial++
	}
}

type RetryError struct {
	Err       error
	Retryable bool
}

func (e RetryError) Error() string {
	return e.Err.Error()
}

func (e RetryError) Temporary() bool {
	return e.Retryable
}

func RetryableError(err error) *RetryError {
	if err == nil {
		return nil
	}
	return &RetryError{Err: err, Retryable: true}
}

func NonRetryableError(err error) *RetryError {
	if err == nil {
		return nil
	}
	return &RetryError{Err: err, Retryable: false}
}

type temporary interface {
	Temporary() bool
}

func IsTemporary(err error) bool {
	te, ok := err.(temporary)
	return ok && te.Temporary()
}

func (c DecryptClient) checkAvailability() error {
	// the below code is copied from api/setup_service. Don't really want to import api here as it will break
	// dag dependency graph. It's probably make sense to separate that logic from api package into a standalone one
	// to just maintain the dependencies.

	request, err := http.NewRequest("GET", "/login/ensure_availability", nil)
	if err != nil {
		return NonRetryableError(err)
	}

	response, err := c.unauthedClient.Do(request)
	if err != nil {
		return NonRetryableError(fmt.Errorf("could not make request round trip: %s", err))
	}

	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusFound:
		location, err := url.Parse(response.Header.Get("Location"))
		if err != nil {
			return NonRetryableError(fmt.Errorf("could not parse redirect url: %s", err))
		}

		switch location.Path {
		case "/auth/cloudfoundry":
			return nil
		case "/setup":
			return RetryableError(errors.New("Status was unstarted"))
		default:
			return NonRetryableError(fmt.Errorf("Unexpected redirect location: %s", location.Path))
		}

	case http.StatusOK:
		respBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return NonRetryableError(err)
		}

		if strings.Contains(string(respBody), "Waiting for authentication system to start...") {
			return RetryableError(errors.New("Status is pending"))
		}

		return NonRetryableError(fmt.Errorf("Received OK with an unexpected body: %s", string(respBody)))

	default:
		return NonRetryableError(fmt.Errorf("Unexpected response code: %d %s", response.StatusCode, http.StatusText(response.StatusCode)))
	}
}
