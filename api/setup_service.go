package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type SetupService struct {
	client httpClient
}

func NewSetupService(client httpClient) SetupService {
	return SetupService{
		client: client,
	}
}

type SetupInput struct {
	IdentityProvider                 string
	AdminUserName                    string
	AdminPassword                    string
	AdminPasswordConfirmation        string
	DecryptionPassphrase             string
	DecryptionPassphraseConfirmation string
	EULAAccepted                     bool
}

type SetupOutput struct{}

func (ss SetupService) Setup(input SetupInput) (SetupOutput, error) {
	var setup struct {
		Setup struct {
			IdentityProvider                 string `json:"identity_provider"`
			AdminUserName                    string `json:"admin_user_name"`
			AdminPassword                    string `json:"admin_password"`
			AdminPasswordConfirmation        string `json:"admin_password_confirmation"`
			DecryptionPassphrase             string `json:"decryption_passphrase"`
			DecryptionPassphraseConfirmation string `json:"decryption_passphrase_confirmation"`
			EULAAccepted                     string `json:"eula_accepted"`
		} `json:"setup"`
	}

	setup.Setup.IdentityProvider = input.IdentityProvider
	setup.Setup.AdminUserName = input.AdminUserName
	setup.Setup.AdminPassword = input.AdminPassword
	setup.Setup.AdminPasswordConfirmation = input.AdminPasswordConfirmation
	setup.Setup.DecryptionPassphrase = input.DecryptionPassphrase
	setup.Setup.DecryptionPassphraseConfirmation = input.DecryptionPassphraseConfirmation
	setup.Setup.EULAAccepted = strconv.FormatBool(input.EULAAccepted)

	payload, err := json.Marshal(setup)
	if err != nil {
		return SetupOutput{}, err
	}

	request, err := http.NewRequest("POST", "/api/v0/setup", bytes.NewReader(payload))
	if err != nil {
		return SetupOutput{}, err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := ss.client.Do(request)
	if err != nil {
		return SetupOutput{}, fmt.Errorf("could not make api request to setup endpoint: %s", err)
	}

	defer response.Body.Close()

	if err = ValidateStatusOK(response); err != nil {
		return SetupOutput{}, err
	}

	return SetupOutput{}, nil
}

const (
	EnsureAvailabilityStatusUnstarted = "unstarted"
	EnsureAvailabilityStatusPending   = "pending"
	EnsureAvailabilityStatusComplete  = "complete"
	EnsureAvailabilityStatusUnknown   = "unknown"
)

type EnsureAvailabilityInput struct{}
type EnsureAvailabilityOutput struct {
	Status string
}

func (ss SetupService) EnsureAvailability(input EnsureAvailabilityInput) (EnsureAvailabilityOutput, error) {
	request, err := http.NewRequest("GET", "/login/ensure_availability", nil)
	if err != nil {
		return EnsureAvailabilityOutput{}, err
	}

	response, err := ss.client.RoundTrip(request)
	if err != nil {
		return EnsureAvailabilityOutput{}, fmt.Errorf("could not make request round trip: %s", err)
	}

	defer response.Body.Close()

	status := EnsureAvailabilityStatusUnknown
	switch {
	case response.StatusCode == http.StatusFound:
		location, err := url.Parse(response.Header.Get("Location"))
		if err != nil {
			return EnsureAvailabilityOutput{}, fmt.Errorf("could not parse redirect url: %s", err)
		}

		if location.Path == "/setup" {
			status = EnsureAvailabilityStatusUnstarted
		} else if location.Path == "/auth/cloudfoundry" {
			status = EnsureAvailabilityStatusComplete
		}

	case response.StatusCode == http.StatusOK:
		respBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return EnsureAvailabilityOutput{}, err
		}

		if strings.Contains(string(respBody), "Waiting for authentication system to start...") {
			status = EnsureAvailabilityStatusPending
		}
	}

	return EnsureAvailabilityOutput{Status: status}, nil
}
