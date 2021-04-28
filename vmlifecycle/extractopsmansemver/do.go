package extractopsmansemver

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blang/semver"
)

var (
	extractSemverVersionRegex    = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)(-build\.\d+)?`)
	extractOldOpsmanVersionRegex = regexp.MustCompile(`(\d+)\.(\d+)-build\.(\d+)`)
)

func Do(s string) (semver.Version, error) {
	name := regexp.MustCompile(`^\[.*?]`).ReplaceAllString(filepath.Base(s), "")
	foundVersion := ""

	extractedVersion := extractSemverVersionRegex.FindStringSubmatch(name)
	if extractedVersion != nil {
		foundVersion = extractedVersion[0]
	} else {
		extractedVersion = extractOldOpsmanVersionRegex.FindStringSubmatch(name)
		if extractedVersion != nil {
			foundVersion = strings.Join([]string{extractedVersion[1], extractedVersion[2], extractedVersion[3]}, ".")
		}
	}

	if foundVersion == "" {
		return semver.Version{}, errors.New("cannot find version from string")
	}

	return semver.Make(foundVersion)
}
