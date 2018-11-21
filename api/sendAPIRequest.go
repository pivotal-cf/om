package api

import (
	"bytes"
	"fmt"
	"net/http"
)

func (a Api) sendAPIRequest(method, endpoint string, jsonData []byte) (*http.Response, error) {
	return sendRequest(a.client, method, endpoint, jsonData)
}

func (a Api) sendProgressAPIRequest(method, endpoint string, jsonData []byte) (*http.Response, error) {
	return sendRequest(a.progressClient, method, endpoint, jsonData)
}

func (a Api) sendUnauthedAPIRequest(method, endpoint string, jsonData []byte) (*http.Response, error) {
	return sendRequest(a.unauthedClient, method, endpoint, jsonData)
}

func sendRequest(client httpClient, method, endpoint string, jsonData []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("could not create api request %s %s: %s", method, endpoint, err.Error())
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return resp, fmt.Errorf("could not send api request to %s %s: %s", method, endpoint, err.Error())
	}

	err = validateStatusOK(resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
