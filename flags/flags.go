package flags

import (
	"flag"
	"io"
	"os"
)

type Flags struct {
	set *flag.FlagSet
}

func New(name string) Flags {
	return Flags{
		set: flag.NewFlagSet(name, flag.ContinueOnError),
	}
}

func (f Flags) Bool(variable *bool, short, long string, defaultValue bool, description string) {
	f.set.BoolVar(variable, short, defaultValue, description)
	f.set.BoolVar(variable, long, defaultValue, description)
}

func (f Flags) Parse() ([]string, error) {
	if err := f.set.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	return f.set.Args(), nil
}

func (f Flags) WriteUsageTo(output io.Writer) {
	f.set.SetOutput(output)
	f.set.PrintDefaults()
}
