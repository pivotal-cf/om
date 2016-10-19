package api

const (
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

type InstallationsService struct {
	client httpClient
}

type InstallationsServiceOutput struct {
	ID     int
	Status string
	Logs   string
}

func NewInstallationsService(client httpClient) InstallationsService {
	return InstallationsService{}
}
