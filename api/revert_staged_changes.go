package api

import (
	"github.com/pkg/errors"
	"net/http"
)

func (a Api) RevertStagedChanges() (bool, error) {
	request, err := http.NewRequest("DELETE", "/api/v0/staged", nil)
	if err != nil {
		panic(err)
	}

	response, err := a.client.Do(request)
	if err != nil {
		return false, errors.Wrap(err, "could not revert staged changes")
	}

	if response.StatusCode == http.StatusNotModified {
		return false, nil
	}

	if err := validateStatusOK(response); err != nil {
		return false, err
	}


	return true, nil
}
