package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Form struct {
	Action            string
	AuthenticityToken string
	RailsMethod       string
}

type PostFormInput struct {
	Form
	EncodedPayload string
}

type BoshFormService struct {
	client httpClient
}

func NewBoshFormService(client httpClient) BoshFormService {
	return BoshFormService{
		client: client,
	}
}

func (bs BoshFormService) GetForm(path string) (Form, error) {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return Form{}, err
	}

	resp, err := bs.client.Do(req)
	if err != nil {
		return Form{}, fmt.Errorf("failed during request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return Form{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return Form{}, fmt.Errorf("request failed: unexpected response:\n%s", out)
	}

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return Form{}, err // cannot test
	}

	action, _ := document.Find("form").Attr("action")
	authenticityToken, ok := document.Find(`input[name="authenticity_token"]`).Attr("value")
	if !ok {
		return Form{}, errors.New("could not find the form authenticity token")
	}

	railsMethod, _ := document.Find(`input[name="_method"]`).Attr("value")

	return Form{
		Action:            action,
		AuthenticityToken: authenticityToken,
		RailsMethod:       railsMethod,
	}, nil
}

func (bs BoshFormService) PostForm(input PostFormInput) error {
	req, err := http.NewRequest("POST", input.Action, strings.NewReader(input.EncodedPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := bs.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to POST form: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return fmt.Errorf("request failed: unexpected response:\n%s", out)
	}

	return nil
}
