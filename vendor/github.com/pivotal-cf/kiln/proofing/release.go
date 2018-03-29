package proofing

import (
	"fmt"
	"strings"
)

type Release struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	File    string `yaml:"file"`

	SHA1 string `yaml:"sha1"` // NOTE: this only exists because of kiln

	// TODO: validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/release.rb#L8-L15
}

type CompoundError []error

func (ce *CompoundError) Add(err error) {
	*ce = append(*ce, err)
}

func (ce *CompoundError) Error() string {
	var messages []string
	for _, e := range *ce {
		messages = append(messages, fmt.Sprintf("- %s", e))
	}
	return strings.Join(messages, "\n")
}

func (r Release) Validate() error {
	var err error
	err = ValidatePresence(err, r, "Name")
	err = ValidatePresence(err, r, "File")
	err = ValidatePresence(err, r, "Version")
	return err
}
