package api

import (
	"encoding/json"
	"fmt"
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

func (a Api) ListStagedPendingChanges() (PendingChangesOutput, error) {
	resp, err := a.sendAPIRequest("GET", pendingChangesEndpoint, nil)
	if err != nil {
		return PendingChangesOutput{}, fmt.Errorf("failed to submit request: %s", err)
	}
	defer resp.Body.Close()

	var pendingChanges PendingChangesOutput
	if err := json.NewDecoder(resp.Body).Decode(&pendingChanges); err != nil {
		return PendingChangesOutput{}, fmt.Errorf("could not unmarshal pending_changes response: %s", err)
	}

	return pendingChanges, nil
}
