package api

import "net/http"

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
	RoundTrip(*http.Request) (*http.Response, error)
}
