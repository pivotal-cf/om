package api

import "net/http"

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}
