package api

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/pkg/errors"
)

func validateStatusOK(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return errors.Wrap(err, "request failed: unexpected response")
		}

		return fmt.Errorf("request failed: unexpected response:\n%s", out)
	}

	return nil
}

func validateStatusOKOrVerificationWarning(resp *http.Response, ignoreVerifierWarnings bool) error {
	if ignoreVerifierWarnings && resp.StatusCode == http.StatusMultiStatus {

		return nil
	}
	return validateStatusOK(resp)
}
