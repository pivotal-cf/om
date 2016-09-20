package flags

import "fmt"

type Bool struct {
	Value bool

	short        string
	long         string
	defaultValue bool
	description  string
}

func NewBool(short, long string, defaultValue bool, description string) *Bool {
	return &Bool{
		Value:        defaultValue,
		short:        short,
		long:         long,
		defaultValue: defaultValue,
		description:  description,
	}
}

func (b Bool) Help() string {
	return fmt.Sprintf("-%s, --%s %s", b.short, b.long, b.description)
}
