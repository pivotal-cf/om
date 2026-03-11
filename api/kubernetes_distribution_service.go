package api

import (
	"encoding/json"
	"fmt"
)

type KubernetesDistributionAssociationsResponse struct {
	Products []KubernetesProductDistributionEntry `json:"products"`
	Library  []KubernetesDistributionLibraryEntry `json:"kubernetes_distribution_library"`
}

type KubernetesDistributionLibraryEntry struct {
	Identifier string `json:"identifier"`
	Version    string `json:"version"`
	Rank       int    `json:"rank"`
	Label      string `json:"label"`
}

type KubernetesProductDistributionEntry struct {
	GUID                             string                   `json:"guid"`
	ProductName                      string                   `json:"identifier"`
	StagedForDeletion                bool                     `json:"is_staged_for_deletion"`
	StagedKubernetesDistribution     *KubernetesDistribution  `json:"staged_kubernetes_distribution"`
	DeployedKubernetesDistribution   *KubernetesDistribution  `json:"deployed_kubernetes_distribution"`
	AvailableKubernetesDistributions []KubernetesDistribution `json:"available_kubernetes_distributions"`
}

type KubernetesDistribution struct {
	Identifier string `json:"identifier"`
	Version    string `json:"version"`
}

type AssignKubernetesDistributionInput struct {
	Products []AssignKubernetesDistributionProduct `json:"products"`
}

type AssignKubernetesDistributionProduct struct {
	GUID                   string                 `json:"guid"`
	KubernetesDistribution KubernetesDistribution `json:"kubernetes_distribution"`
}

func (a Api) ListKubernetesDistributions() (KubernetesDistributionAssociationsResponse, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/kubernetes_distribution_associations", nil)
	if err != nil {
		return KubernetesDistributionAssociationsResponse{}, fmt.Errorf("could not make api request to list kubernetes distributions: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err = validateStatusOK(resp); err != nil {
		return KubernetesDistributionAssociationsResponse{}, err
	}

	var result KubernetesDistributionAssociationsResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return KubernetesDistributionAssociationsResponse{}, fmt.Errorf("invalid JSON: %s", err)
	}

	return result, nil
}

func (a Api) AssignKubernetesDistribution(input AssignKubernetesDistributionInput) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return fmt.Errorf("could not marshal json: %w", err)
	}

	resp, err := a.sendAPIRequest("PATCH", "/api/v0/kubernetes_distribution_associations", jsonData)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
