package parser

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
)

type Int struct {
	Set *flag.FlagSet
	Field reflect.Value
	Tags reflect.StructTag
}

func (i Int) Execute() error {
	var defaultValue int64
	defaultStr, ok := i.Tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = strconv.ParseInt(defaultStr, 0, 0)
		if err != nil {
			return fmt.Errorf("could not parse int default value %q: %s", defaultStr, err)
		}
	}

	short, ok := i.Tags.Lookup("short")
	if ok {
		i.Set.IntVar(i.Field.Addr().Interface().(*int), short, int(defaultValue), "")
	}

	long, ok := i.Tags.Lookup("long")
	if ok {
		i.Set.IntVar(i.Field.Addr().Interface().(*int), long, int(defaultValue), "")
	}

	return nil
}