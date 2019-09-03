package api

import "net/http"

//counterfeiter:generate -o ./fakes/httpclient.go --fake-name HttpClient . httpClient
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}
