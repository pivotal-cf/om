package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type DashboardService struct {
	client httpClient
}

func NewDashboardService(client httpClient) DashboardService {
	return DashboardService{
		client: client,
	}
}

func (bs DashboardService) GetInstallForm() (Form, error) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		return Form{}, err
	}

	resp, err := bs.client.Do(req)
	if err != nil {
		return Form{}, fmt.Errorf("failed during request: %s", err)
	}
	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return Form{}, err
	}

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return Form{}, err // cannot test
	}

	var action, authenticityToken, railsMethod string
	var tokenFound, methodFound bool
	document.Find("form").Each(func(index int, sel *goquery.Selection) {
		formAction, _ := sel.Attr("action")
		if formAction == "/install" {
			action = "/install"
			authenticityToken, tokenFound = sel.Find(`input[name="authenticity_token"]`).Attr("value")
			railsMethod, methodFound = sel.Find(`input[name="_method"]`).Attr("value")
		}
	})

	if !tokenFound {
		return Form{}, errors.New("could not find the form authenticity token")
	}

	return Form{
		Action:            action,
		AuthenticityToken: authenticityToken,
		RailsMethod:       railsMethod,
	}, nil
}

func (bs DashboardService) PostInstallForm(input PostFormInput) error {
	req, err := http.NewRequest("POST", "/installation", strings.NewReader(input.EncodedPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := bs.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to POST form: %s", err)
	}

	return ValidateStatusOK(resp)
}
