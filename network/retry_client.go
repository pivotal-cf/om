package network

import (
	"net/http"
	"time"
)

type RetryClient struct {
	client     httpClient
	retryCount int
	retryDelay time.Duration
}

func NewRetryClient(client httpClient, retryCount int, retryDelay time.Duration) RetryClient {
	return RetryClient{
		client:     client,
		retryCount: retryCount,
		retryDelay: retryDelay,
	}
}

func (c RetryClient) Do(request *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 0; i < c.retryCount+1; i++ {
		resp, err = c.client.Do(request)
		if err != nil || resp.StatusCode >= 500 {
			time.Sleep(c.retryDelay)
		} else {
			break
		}
	}

	return resp, err
}
