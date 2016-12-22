package commands

import (
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/pivotal-cf/om/api"
)

type RevertStagedChanges struct {
	service dashboardService
	logger  logger
}

//go:generate counterfeiter -o ./fakes/dashboard_service.go --fake-name DashboardService . dashboardService
type dashboardService interface {
	GetInstallForm() (api.Form, error)
	PostInstallForm(api.PostFormInput) error
}

func NewRevertStagedChanges(s dashboardService, l logger) RevertStagedChanges {
	return RevertStagedChanges{service: s, logger: l}
}

func (c RevertStagedChanges) Execute(args []string) error {
	form, err := c.service.GetInstallForm()
	if err != nil {
		return fmt.Errorf("could not fetch form: %s", err)
	}

	var formConfig CommonConfiguration
	formConfig.AuthenticityToken = form.AuthenticityToken
	formConfig.Method = "delete"
	formConfig.Commit = "Confirm"

	formValues, err := query.Values(formConfig)
	if err != nil {
		return err // cannot be tested
	}

	c.logger.Printf("reverting staged changes on the targeted Ops Manager")
	err = c.service.PostInstallForm(api.PostFormInput{Form: form, EncodedPayload: formValues.Encode()})
	if err != nil {
		return fmt.Errorf("failed to revert staged changes: %s", err)
	}
	c.logger.Printf("done")

	return nil
}

func (c RevertStagedChanges) Usage() Usage {
	return Usage{
		Description:      "reverts staged changes on the installation dashboard page in the target Ops Manager",
		ShortDescription: "reverts staged changes on the Ops Manager targeted",
	}
}
