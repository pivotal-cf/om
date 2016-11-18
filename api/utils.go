package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"regexp"
)

func GetBooleanAsFormValue(val bool) string {
	if val {
		return "1"
	}
	return "0"
}

func GetCSRFToken(resp *http.Response) (string, error) {
	r, err := regexp.Compile("name=\"csrf-token\"(\\s*)content=\"(.*)\"")
	if err != nil {
		return "", err
	}
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return "", err
	} else {
		//Evaluate regex and get results
		matches := r.FindAllStringSubmatch(string(body), 1)
		if len(matches) == 0 {
			return "", fmt.Errorf("Unable to find token")
		} else {
			return matches[0][2], nil
		}
	}
}

func HandleResponse(resp *http.Response, expectedStatus int) error {
	if resp.StatusCode != expectedStatus {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return fmt.Errorf("request failed: unexpected response: %s", err)
		}
		return fmt.Errorf("request failed: unexpected response:\n%s", out)
	}
	return nil
}
