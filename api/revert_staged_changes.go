package api

import (
	"errors"
	"fmt"
	"net/http"
)

func (a Api) RevertStagedChanges() (bool, error) {
	request, err := http.NewRequest("DELETE", "/api/v0/staged", nil)
	if err != nil {
		panic(err)
	}

	response, err := a.client.Do(request)
	if err != nil {
		return false, fmt.Errorf("could not revert staged changes: %w", err)
	}

	if response.StatusCode == http.StatusNotModified {
		return false, nil
	}

	if response.StatusCode == http.StatusNotFound {
		return false, errors.New("The revert staged changes endpoint is not available in the version of Ops Manager.\nThis endpoint was not available until Ops Manager 2.5.21+, 2.6.13+, or 2.7.2+.")
	}

	if err := validateStatus(response, http.StatusNoContent); err != nil {
		return false, err
	}

	return true, nil
}
