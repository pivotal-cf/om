package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	availabilityZonesConfigurationPath = "/infrastructure/availability_zones/edit"
	networkAssignmentConfigurationPath = "/infrastructure/director/az_and_network_assignment/edit"
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

	if err = ValidateStatusOK(resp); err != nil {
		return Form{}, err
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

	if err = ValidateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (bs BoshFormService) AvailabilityZones() (map[string]string, error) {
	zones := make(map[string]string)

	req, err := http.NewRequest("GET", availabilityZonesConfigurationPath, nil)
	if err != nil {
		return zones, err // cannot test
	}

	resp, err := bs.client.Do(req)
	if err != nil {
		return zones, fmt.Errorf("failed during request: %s", err)
	}
	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return zones, err
	}

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return zones, err // cannot test
	}

	azNames := document.Find("form").Find(`input[name*='iaas_identifier']`).Map(
		func(i int, s *goquery.Selection) string {
			v, _ := s.Attr("value")
			return v
		},
	)

	if len(azNames) == 0 {
		azNames = document.Find("form").Find(`input[name*='name']`).Map(
			func(i int, s *goquery.Selection) string {
				v, _ := s.Attr("value")
				return v
			},
		)
	}

	azGuids := document.Find("form").Find(`input[name*='guid'][type=hidden]`).Map(
		func(i int, s *goquery.Selection) string {
			v, _ := s.Attr("value")
			return v
		},
	)

	if len(azGuids) != len(azNames) {
		return zones, fmt.Errorf("failed constructing AZ map - mismatched # of AZ names to GUIDs")
	}

	for i, azName := range azNames {
		zones[azName] = azGuids[i]
	}

	return zones, nil
}

func (bs BoshFormService) Networks() (map[string]string, error) {
	networks := make(map[string]string)

	req, err := http.NewRequest("GET", networkAssignmentConfigurationPath, nil)
	if err != nil {
		return networks, err // shall not test
	}

	resp, err := bs.client.Do(req)
	if err != nil {
		return networks, fmt.Errorf("failed during request: %s", err)
	}
	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return networks, err
	}

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return networks, err
	}

	document.Find(`select[id*=network] option:not([value=''])`).Each(
		func(i int, s *goquery.Selection) {
			name := s.Text()
			guid, _ := s.Attr("value")

			networks[name] = guid
		},
	)

	return networks, nil
}
