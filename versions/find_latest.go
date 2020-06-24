package versions

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"log"
	"regexp"
	"sort"
)

func FindLatestVersionFromRegex(productVersions []string, regex string, stderr *log.Logger) (string ,error) {
	re, err := regexp.Compile(regex)
	if err != nil {
		return "", fmt.Errorf("could not compile regex '%s': %s", regex, err)
	}

	var versions version.Collection
	for _, productVersion := range productVersions {
		if !re.MatchString(productVersion) {
			continue
		}

		v, err := version.NewVersion(productVersion)
		if err != nil {
			stderr.Printf(fmt.Sprintf("warning: could not parse semver version from: %s", productVersion))
			continue
		}
		versions = append(versions, v)
	}

	sort.Sort(versions)

	if len(versions) == 0 {
		return "", fmt.Errorf("not available version found")
	}

	return versions[len(versions)-1].Original(), nil
}
