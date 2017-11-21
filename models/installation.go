package models

type Installation struct {
	Id         string
	User       string
	Status     string
	StartedAt  string
	FinishedAt string
}

type Product struct {
	Name    string
	Version string
}
