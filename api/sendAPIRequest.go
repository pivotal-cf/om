package api

import (
	"bytes"
	"fmt"
	"net/http"
)

func (a Api) sendAPIRequest(method, endpoint string, jsonData []byte) (*http.Response, error) {
	return sendRequestNew(a.client, method, endpoint, jsonData)
}

func (a Api) sendProgressAPIRequest(method, endpoint string, jsonData []byte) (*http.Response, error) {
	return sendRequestNew(a.progressClient, method, endpoint, jsonData)
}

func (a Api) sendUnauthedAPIRequest(method, endpoint string, jsonData []byte) (*http.Response, error) {
	return sendRequest(a.unauthedClient, method, endpoint, jsonData)
}

func sendRequest(client httpClient, method, endpoint string, jsonData []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("could not create api request %s %s: %w", method, endpoint, err)
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send api request to %s %s: %w", method, endpoint, err)
	}

	return resp, nil
}

func sendRequestNew(client httpClient, method, endpoint string, jsonData []byte) (*http.Response, error) {
	const (
		retries = 2
	)
	for i := 0; i < retries; i++ {
		req, err := http.NewRequest(method, endpoint, bytes.NewReader(jsonData))
		if err != nil {
			return nil, fmt.Errorf("could not create api request %s %s: %w", method, endpoint, err)
		}
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err == nil {
			return resp, nil // Request was successful
		}
		// If there is an error and we have more retries left, log the error and retry
		fmt.Printf("Error sending API request to %s %s: %v. Retrying...\n", method, endpoint, err)
	}
	return nil, fmt.Errorf("could not send API request after %d retries", retries)

}
