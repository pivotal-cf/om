package api

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/pkg/errors"
)

func validateStatusOK(resp *http.Response) error {
	return validateStatus(resp, http.StatusOK)
}

func validateStatusOKOrVerificationWarning(resp *http.Response, ignoreVerifierWarnings bool) error {
	if ignoreVerifierWarnings && resp.StatusCode == http.StatusMultiStatus {

		return nil
	}
	return validateStatusOK(resp)
}

func validateStatus(resp *http.Response, status int) error {
	if resp.StatusCode != status {
		var requestURL string
		if resp.Request != nil {
			requestURL = fmt.Sprintf(" from %s", resp.Request.URL.Path)
		}

		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("request failed: unexpected response%s", requestURL))
		}

		return fmt.Errorf("request failed: unexpected response%s:\n%s", requestURL, out)
	}

	return nil
}
