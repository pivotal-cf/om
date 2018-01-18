package parser

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
)

func NewInt(set *flag.FlagSet, field reflect.Value, tags reflect.StructTag) (*Flag, error) {
	var defaultValue int64
	defaultStr, ok := tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = strconv.ParseInt(defaultStr, 0, 0)
		if err != nil {
			return &Flag{}, fmt.Errorf("could not parse int default value %q: %s", defaultStr, err)
		}
	}

	var f Flag
	short, ok := tags.Lookup("short")
	if ok {
		set.IntVar(field.Addr().Interface().(*int), short, int(defaultValue), "")
		f.flags = append(f.flags, set.Lookup(short))
		f.name = fmt.Sprintf("-%s", short)
	}

	long, ok := tags.Lookup("long")
	if ok {
		set.IntVar(field.Addr().Interface().(*int), long, int(defaultValue), "")
		f.flags = append(f.flags, set.Lookup(long))
		f.name = fmt.Sprintf("--%s", long)
	}

	_, f.required = tags.Lookup("required")

	return &f, nil
}
