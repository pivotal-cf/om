package parser

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
)

type Int64 struct{
	Set *flag.FlagSet
	Field reflect.Value
	Tags reflect.StructTag
}

func (i Int64) Execute() error {
	var defaultValue int64
	defaultStr, ok := i.Tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = strconv.ParseInt(defaultStr, 0, 64)
		if err != nil {
			return fmt.Errorf("could not parse int64 default value %q: %s", defaultStr, err)
		}
	}

	short, ok := i.Tags.Lookup("short")
	if ok {
		i.Set.Int64Var(i.Field.Addr().Interface().(*int64), short, defaultValue, "")
	}

	long, ok := i.Tags.Lookup("long")
	if ok {
		i.Set.Int64Var(i.Field.Addr().Interface().(*int64), long, defaultValue, "")
	}

	return nil
}