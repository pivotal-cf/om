package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type SetupInput struct {
	IdentityProvider                 string `json:"identity_provider"`
	AdminUserName                    string `json:"admin_user_name,omitempty"`
	AdminPassword                    string `json:"admin_password,omitempty"`
	AdminPasswordConfirmation        string `json:"admin_password_confirmation,omitempty"`
	DecryptionPassphrase             string `json:"decryption_passphrase"`
	DecryptionPassphraseConfirmation string `json:"decryption_passphrase_confirmation"`
	EULAAccepted                     string `json:"eula_accepted"`
	HTTPProxyURL                     string `json:"http_proxy,omitempty"`
	HTTPSProxyURL                    string `json:"https_proxy,omitempty"`
	NoProxy                          string `json:"no_proxy,omitempty"`
	IDPMetadata                      string `json:"idp_metadata,omitempty"`
	BoshIDPMetadata                  string `json:"bosh_idp_metadata,omitempty"`
	RBACAdminGroup                   string `json:"rbac_saml_admin_group,omitempty"`
	RBACGroupsAttribute              string `json:"rbac_saml_groups_attribute,omitempty"`
}

type SetupOutput struct{}

type setup struct {
	SetupInput `json:"setup"`
}

func (a Api) Setup(input SetupInput) (SetupOutput, error) {
	payload, err := json.Marshal(setup{input})
	if err != nil {
		return SetupOutput{}, err
	}

	request, err := http.NewRequest("POST", "/api/v0/setup", bytes.NewReader(payload))
	if err != nil {
		return SetupOutput{}, err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := a.unauthedClient.Do(request)
	if err != nil {
		return SetupOutput{}, fmt.Errorf("could not make api request to setup endpoint: %s", err)
	}

	defer response.Body.Close()

	if err = validateStatusOK(response); err != nil {
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

func (a Api) EnsureAvailability(input EnsureAvailabilityInput) (EnsureAvailabilityOutput, error) {
	request, err := http.NewRequest("GET", "/login/ensure_availability", nil)
	if err != nil {
		return EnsureAvailabilityOutput{}, err
	}

	response, err := a.unauthedClient.Do(request)
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
		} else {
			return EnsureAvailabilityOutput{}, fmt.Errorf("Unexpected redirect location: %s", location.Path)
		}

	case response.StatusCode == http.StatusOK:
		respBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return EnsureAvailabilityOutput{}, err
		}

		if strings.Contains(string(respBody), "Waiting for authentication system to start...") {
			status = EnsureAvailabilityStatusPending
		} else {
			return EnsureAvailabilityOutput{}, fmt.Errorf("Received OK with an unexpected body: %s", string(respBody))
		}

	default:
		return EnsureAvailabilityOutput{}, fmt.Errorf("Unexpected response code: %d %s", response.StatusCode, http.StatusText(response.StatusCode))
	}

	return EnsureAvailabilityOutput{Status: status}, nil
}
