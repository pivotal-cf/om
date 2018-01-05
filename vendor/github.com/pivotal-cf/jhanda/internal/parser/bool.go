package parser

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
)

type Bool struct {
	Set   *flag.FlagSet
	Field reflect.Value
	Tags  reflect.StructTag
}

func (b Bool) Execute() error {

	var defaultValue bool
	defaultStr, ok := b.Tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = strconv.ParseBool(defaultStr)
		if err != nil {
			return fmt.Errorf("could not parse bool default value %q: %s", defaultStr, err)
		}
	}

	short, ok := b.Tags.Lookup("short")
	if ok {
		b.Set.BoolVar(b.Field.Addr().Interface().(*bool), short, defaultValue, "")
	}

	long, ok := b.Tags.Lookup("long")
	if ok {
		b.Set.BoolVar(b.Field.Addr().Interface().(*bool), long, defaultValue, "")
	}

	return nil
}