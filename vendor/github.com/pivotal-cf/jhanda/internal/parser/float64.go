package parser

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
)

type Float64 struct{
	Set   *flag.FlagSet
	Field reflect.Value
	Tags  reflect.StructTag
}

func (f Float64) Execute() error {
	var defaultValue float64
	defaultStr, ok := f.Tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = strconv.ParseFloat(defaultStr, 64)
		if err != nil {
			return fmt.Errorf("could not parse float64 default value %q: %s", defaultStr, err)
		}
	}

	short, ok := f.Tags.Lookup("short")
	if ok {
		f.Set.Float64Var(f.Field.Addr().Interface().(*float64), short, defaultValue, "")
	}

	long, ok := f.Tags.Lookup("long")
	if ok {
		f.Set.Float64Var(f.Field.Addr().Interface().(*float64), long, defaultValue, "")
	}

	return nil
}