package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type GetBoshEnvironmentOutput struct {
	Client       string
	ClientSecret string
	Environment  string
}

type credential struct {
	Credential string `json:"credential"`
}

func (a Api) GetBoshEnvironment() (GetBoshEnvironmentOutput, error) {
	req, err := http.NewRequest("GET", "/api/v0/deployed/director/credentials/bosh_commandline_credentials", nil)
	if err != nil {
		return GetBoshEnvironmentOutput{}, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return GetBoshEnvironmentOutput{}, fmt.Errorf("could not make api request to director credentials endpoint: %s", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return GetBoshEnvironmentOutput{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return GetBoshEnvironmentOutput{}, err
	}

	output := credential{}
	err = json.Unmarshal(respBody, &output)
	if err != nil {
		return GetBoshEnvironmentOutput{}, fmt.Errorf("could not unmarshal director credentials response: %s", err)
	}
	if err != nil {
		return GetBoshEnvironmentOutput{}, err
	}
	keyValues := parseKeyValues(output.Credential)
	return GetBoshEnvironmentOutput{
		Client:       keyValues["BOSH_CLIENT"],
		ClientSecret: keyValues["BOSH_CLIENT_SECRET"],
		Environment:  keyValues["BOSH_ENVIRONMENT"],
	}, nil
}

func parseKeyValues(credentials string) map[string]string {
	values := make(map[string]string)
	kvs := strings.Split(credentials, " ")
	for _, kv := range kvs {
		if strings.Contains(kv, "=") {
			k := strings.Split(kv, "=")[0]
			v := strings.Split(kv, "=")[1]
			values[k] = v
		}
	}
	return values
}
