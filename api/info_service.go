package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Info contains information about Ops Manager itself.
type Info struct {
	Version string `json:"version"`
}

func (i Info) VersionAtLeast(major, minor int) (bool, error) {
	// Given: X.Y-build.Z or X.Y.Z-build.A
	// Extract X and Y
	parts := strings.Split(i.Version, ".")
	if len(parts) < 2 {
		return false, fmt.Errorf("invalid version: '%s'", i.Version)
	}
	maj, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, fmt.Errorf("invalid version: '%s'", i.Version)
	}

	//remove "-build.A" information
	minParts := strings.Split(parts[1], "-")
	min, err := strconv.Atoi(minParts[0])
	if err != nil {
		return false, fmt.Errorf("invalid version: '%s'", i.Version)
	}

	if maj < major || (maj == major && min < minor) {
		return false, nil
	}
	return true, nil
}

// Info gets information about Ops Manager.
func (a Api) Info() (Info, error) {
	var r struct {
		Info Info `json:"info"`
	}

	resp, err := a.sendUnauthedAPIRequest("GET", "/api/v0/info", nil)
	if err != nil {
		return r.Info, errors.Wrap(err, "could not make request to info endpoint")
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return Info{}, err
	}

	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.Info, err
}
