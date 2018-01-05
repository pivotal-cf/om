package parser

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
)

type Uint struct {
	Set *flag.FlagSet
	Field reflect.Value
	Tags reflect.StructTag
}

func (u Uint) Execute() error {
	var defaultValue uint64
	defaultStr, ok := u.Tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = strconv.ParseUint(defaultStr, 0, 0)
		if err != nil {
			return fmt.Errorf("could not parse uint default value %q: %s", defaultStr, err)
		}
	}

	short, ok := u.Tags.Lookup("short")
	if ok {
		u.Set.UintVar(u.Field.Addr().Interface().(*uint), short, uint(defaultValue), "")
	}

	long, ok := u.Tags.Lookup("long")
	if ok {
		u.Set.UintVar(u.Field.Addr().Interface().(*uint), long, uint(defaultValue), "")
	}

	return nil
}