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

type Errand struct {
	Name              string `json:"name"`
	PostDeployEnabled string `json:"post_deploy_enabled,omitempty"`
	PreDeleteEnabled  string `json:"pre_delete_enabled,omitempty"`
}
