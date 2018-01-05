package parser

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
)

type Uint64 struct {
	Set *flag.FlagSet
	Field reflect.Value
	Tags reflect.StructTag
}

func (u Uint64) Execute() error {
	var defaultValue uint64
	defaultStr, ok := u.Tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = strconv.ParseUint(defaultStr, 0, 64)
		if err != nil {
			return fmt.Errorf("could not parse uint64 default value %q: %s", defaultStr, err)
		}
	}

	short, ok := u.Tags.Lookup("short")
	if ok {
		u.Set.Uint64Var(u.Field.Addr().Interface().(*uint64), short, defaultValue, "")
	}

	long, ok := u.Tags.Lookup("long")
	if ok {
		u.Set.Uint64Var(u.Field.Addr().Interface().(*uint64), long, defaultValue, "")
	}

	return nil
}