package network

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// parseURL takes a candidate target URL and attempts to parse it with sane defaults
func parseURL(u string) (*url.URL, error) {
	// default the target protocol to https if none specified
	if !strings.Contains(u, "://") {
		u = "https://" + u
	}

	targetURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	// at a minimum ensure we have a host with http(s) protocol
	if targetURL.Scheme != "https" && targetURL.Scheme != "http" {
		return nil, fmt.Errorf("error parsing URL, expected http(s) protocol but got %s", targetURL.Scheme)
	}
	if targetURL.Host == "" {
		return nil, errors.New("target flag is required, run `om help` for more info")
	}

	return targetURL, nil
}

// parseOpsmanAndUAAURLs takes a candidate OpsMan and UAA target URLs and attempts to parse both of them, defaulting
// the UAA target to the /uaa path under the OpsMan target if none specified.
func parseOpsmanAndUAAURLs(opsmanTarget, uaaTarget string) (*url.URL, *url.URL, error) {
	opsmanURL, err := parseURL(opsmanTarget)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse Opsman target URL: %w", err)
	}

	var uaaURL *url.URL
	if uaaTarget != "" {
		uaaURL, err = parseURL(uaaTarget)
		if err != nil {
			return nil, nil, fmt.Errorf("could not parse UAA target URL: %w", err)
		}
	} else {
		// default to opsman URL with /uaa path (shallow copy)
		t := *opsmanURL
		t.Path = "/uaa"
		uaaURL = &t
	}

	return opsmanURL, uaaURL, nil
}
