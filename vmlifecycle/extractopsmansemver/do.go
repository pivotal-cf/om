package extractopsmansemver

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blang/semver"
)

var (
	oldOpsmanBuildFormatRegex = regexp.MustCompile(`(\d+)\.(\d+)-build\.(\d+)`)
	semverRegex               = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
)

func Do(s string) (semver.Version, error) {
	name := regexp.MustCompile(`^\[.*?]`).ReplaceAllString(filepath.Base(s), "")

	extractedVersion := semverRegex.FindStringSubmatch(name)
	if extractedVersion == nil {
		extractedVersion = oldOpsmanBuildFormatRegex.FindStringSubmatch(name)
	}

	if extractedVersion == nil {
		return semver.Version{}, errors.New("cannot find version from string")
	}

	foundVersion := strings.Join([]string{extractedVersion[1], extractedVersion[2], extractedVersion[3]}, ".")

	return semver.Make(foundVersion)
}
