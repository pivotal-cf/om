package api

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func ValidateStatusOK(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return fmt.Errorf("request failed: unexpected response:\n%s", out)
	}
	return nil
}
