package api

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const pendingChangesEndpoint = "/api/v0/staged/pending_changes"

type PendingChangesOutput struct {
	ChangeList []ProductChange `json:"product_changes"`
}

type CompletenessChecks struct {
	ConfigurationComplete       bool `json:"configuration_complete"`
	StemcellPresent             bool `json:"stemcell_present"`
	ConfigurablePropertiesValid bool `json:"configurable_properties_valid"`
}

type ProductChange struct {
	GUID               string              `json:"guid"`
	Action             string              `json:"action"`
	Errands            []Errand            `json:"errands"`
	CompletenessChecks *CompletenessChecks `json:"completeness_checks,omitempty"`
}

func (a Api) ListStagedPendingChanges() (PendingChangesOutput, error) {
	resp, err := a.sendAPIRequest("GET", pendingChangesEndpoint, nil)
	if err != nil {
		return PendingChangesOutput{}, errors.Wrap(err, "failed to submit request")
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return PendingChangesOutput{}, err
	}

	var pendingChanges PendingChangesOutput
	if err := json.NewDecoder(resp.Body).Decode(&pendingChanges); err != nil {
		return PendingChangesOutput{}, errors.Wrap(err, "could not unmarshal pending_changes response")
	}

	return pendingChanges, nil
}
