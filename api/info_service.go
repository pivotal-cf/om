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
	idx := strings.Index(i.Version, ".")
	if idx == -1 {
		return false, fmt.Errorf("invalid version: '%s'", i.Version)
	}

	majv := i.Version[:idx] // take substring up to '.'
	toDash := strings.Index(i.Version, "-")
	if toDash == -1 {
		toDash = len(i.Version) - 1
	}
	legacyMinv := i.Version[idx+1 : toDash] // take substring between '.' and '-' if dash is there

	maj, err := strconv.Atoi(majv)
	if err != nil {
		return false, fmt.Errorf("invalid version: '%s'", i.Version)
	}

	min, err := strconv.Atoi(legacyMinv)
	if err != nil {
		semverMinv := legacyMinv[:strings.Index(legacyMinv, ".")] // take substring up to '.'
		min, err = strconv.Atoi(semverMinv)
		if err != nil {
			return false, fmt.Errorf("invalid version: '%s'", i.Version)
		}
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

	resp, err := a.sendAPIRequest("GET", "/api/v0/info", nil)
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
