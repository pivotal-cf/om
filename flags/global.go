package flags

import (
	"flag"
	"io/ioutil"
)

type Global struct {
	Help    *Bool
	Version *Bool
}

func NewGlobal() Global {
	return Global{
		Help:    NewBool("h", "help", false, "    prints this usage information"),
		Version: NewBool("v", "version", false, " prints the om release version"),
	}
}

func (g Global) Parse(args []string) ([]string, error) {
	set := flag.NewFlagSet("global", flag.ContinueOnError)
	set.SetOutput(ioutil.Discard)
	set.Usage = func() {}

	set.BoolVar(&g.Help.Value, g.Help.short, false, g.Help.description)
	set.BoolVar(&g.Help.Value, g.Help.long, false, g.Help.description)

	set.BoolVar(&g.Version.Value, g.Version.short, false, g.Version.description)
	set.BoolVar(&g.Version.Value, g.Version.long, false, g.Version.description)

	err := set.Parse(args)
	if err != nil {
		return nil, err
	}

	return set.Args(), nil
}
