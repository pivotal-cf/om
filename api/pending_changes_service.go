package api

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/pkg/errors"
)

const pendingChangesEndpoint = "/api/v0/staged/pending_changes"

type PendingChangesOutput struct {
	ChangeList []ProductChange `json:"product_changes"`
	FullReport string
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

	reportBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var pendingChanges PendingChangesOutput
	if err := json.Unmarshal(reportBytes, &pendingChanges); err != nil {
		return PendingChangesOutput{}, errors.Wrap(err, "could not unmarshal pending_changes response")
	}

	var pendingChangesFull *struct {
		PendingChanges json.RawMessage `json:"product_changes"`
	}
	if err := json.Unmarshal(reportBytes, &pendingChangesFull); err != nil {
		return PendingChangesOutput{}, errors.Wrap(err, "could not unmarshal pending_changes response")
	}

	pendingChanges.FullReport = string(pendingChangesFull.PendingChanges)
	return pendingChanges, nil
}
