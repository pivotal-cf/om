package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const pendingChangesEndpoint = "/api/v0/staged/pending_changes"

type PendingChangesOutput struct {
	ChangeList []ProductChange `json:"product_changes"`
}

type ProductChange struct {
	Product string   `json:"guid"`
	Errands []Errand `json:"errands"`
	Action  string   `json:"action"`
}

type PendingChangesService struct {
	client httpClient
}

func NewPendingChangesService(client httpClient) PendingChangesService {
	return PendingChangesService{
		client: client,
	}
}

func (pc PendingChangesService) List() (PendingChangesOutput, error) {
	pcReq, err := http.NewRequest("GET", pendingChangesEndpoint, nil)
	if err != nil {
		return PendingChangesOutput{}, err
	}

	resp, err := pc.client.Do(pcReq)
	if err != nil {
		return PendingChangesOutput{}, fmt.Errorf("could not make api request to pending_changes endpoint: %s", err)
	}
	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return PendingChangesOutput{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return PendingChangesOutput{}, err
	}

	var pendingChanges PendingChangesOutput
	err = json.Unmarshal(respBody, &pendingChanges)
	if err != nil {
		return PendingChangesOutput{}, fmt.Errorf("could not unmarshal pending_changes response: %s", err)
	}

	return pendingChanges, nil
}
