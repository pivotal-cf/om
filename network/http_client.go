package network

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

func newHTTPClient(insecureSkipVerify bool, connectTimeout time.Duration, requestTimeout time.Duration) *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureSkipVerify,
				MinVersion:         tls.VersionTLS12,
			},
			Dial: (&net.Dialer{
				Timeout:   connectTimeout,
				KeepAlive: 30 * time.Second,
			}).Dial,
		},
		Timeout: requestTimeout,
	}
}