package models

import "time"

type Installation struct {
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Id         int        `json:"id"`
	StartedAt  *time.Time `json:"started_at"`
	Status     string     `json:"status"`
	User       string     `json:"user"`
}

type Product struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ProductsVersionsDisplay struct {
	ProductVersions []ProductVersions `json:"product_versions"`
	Available       bool
	Staged          bool
	Deployed        bool
}

type ProductVersions struct {
	Name      string   `json:"name"`
	Available []string `json:"available,omitempty"`
	Staged    string   `json:"staged,omitempty"`
	Deployed  string   `json:"deployed,omitempty"`
}

type Errand struct {
	Name              string `json:"name"`
	PostDeployEnabled string `json:"post_deploy_enabled,omitempty"`
	PreDeleteEnabled  string `json:"pre_delete_enabled,omitempty"`
}
