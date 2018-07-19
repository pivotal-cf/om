package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Info contains information about Ops Manager itself.
type Info struct {
	Version string `json:"version"`
}

// Info gets information about Ops Manager.
func (a Api) Info() (Info, error) {
	var r struct {
		Info Info `json:"info"`
	}
	req, err := http.NewRequest("GET", "/api/v0/info", nil)
	if err != nil {
		return r.Info, err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return r.Info, fmt.Errorf("could not make request to info endpoint: %v", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return r.Info, err
	}
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.Info, err
}
