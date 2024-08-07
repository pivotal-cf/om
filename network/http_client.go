package network

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func newHTTPClient(insecureSkipVerify bool, caCert string, requestTimeout time.Duration, connectTimeout time.Duration) (*http.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}
	err := setCACert(caCert, tlsConfig)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: tlsConfig,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				d := net.Dialer{
					Timeout:   connectTimeout,
					KeepAlive: 30 * time.Second,
				}
				return d.DialContext(ctx, network, addr)
			},
		},
		Timeout: requestTimeout,
	}, nil
}

func setCACert(caCert string, tlsConfig *tls.Config) error {
	if caCert == "" {
		return nil
	}

	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		caCertPool = x509.NewCertPool()
	}
	if !strings.Contains(caCert, "BEGIN") {
		contents, err := os.ReadFile(caCert)
		if err != nil {
			return fmt.Errorf("could not load ca cert from file: %s", err)
		}
		caCert = string(contents)
	}
	if ok := caCertPool.AppendCertsFromPEM([]byte(caCert)); !ok {
		return errors.New("could not use ca cert")
	}

	tlsConfig.RootCAs = caCertPool
	return nil
}
