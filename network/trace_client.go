package network

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

//go:generate counterfeiter -o ./fakes/httpclient.go --fake-name HttpClient . httpClient

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type TraceClient struct {
	client httpClient
	writer io.Writer
}

func NewTraceClient(client httpClient, writer io.Writer) *TraceClient {
	return &TraceClient{
		client: client,
		writer: writer,
	}
}

func (c *TraceClient) Do(request *http.Request) (*http.Response, error) {
	requestOutput, err := httputil.DumpRequest(request, true)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(c.writer, "%s\n", string(requestOutput))

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}

	responseOutput, err := httputil.DumpResponse(response, true)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(c.writer, "%s\n", string(responseOutput))

	return response, nil
}
