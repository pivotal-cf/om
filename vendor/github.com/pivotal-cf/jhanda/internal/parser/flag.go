package parser

import (
	"flag"
	"fmt"
)

type Flag struct {
	flags []*flag.Flag
	name  string

	set      bool
	required bool
}

func (f *Flag) SetIfMatched(ef *flag.Flag) {
	for _, ff := range f.flags {
		if ff == ef {
			f.set = true
		}
	}
}

func (f Flag) Validate() error {
	if f.required && !f.set {
		return fmt.Errorf("missing required flag %q", f.name)
	}

	return nil
}
