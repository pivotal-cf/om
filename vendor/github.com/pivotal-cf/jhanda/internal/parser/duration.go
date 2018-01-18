package parser

import (
	"flag"
	"fmt"
	"reflect"
	"time"
)

func NewDuration(set *flag.FlagSet, field reflect.Value, tags reflect.StructTag) (*Flag, error) {
	var defaultValue time.Duration
	defaultStr, ok := tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = time.ParseDuration(defaultStr)
		if err != nil {
			return &Flag{}, fmt.Errorf("could not parse duration default value %q: %s", defaultStr, err)
		}
	}

	var f Flag
	short, ok := tags.Lookup("short")
	if ok {
		set.DurationVar(field.Addr().Interface().(*time.Duration), short, defaultValue, "")
		f.flags = append(f.flags, set.Lookup(short))
		f.name = fmt.Sprintf("-%s", short)
	}

	long, ok := tags.Lookup("long")
	if ok {
		set.DurationVar(field.Addr().Interface().(*time.Duration), long, defaultValue, "")
		f.flags = append(f.flags, set.Lookup(long))
		f.name = fmt.Sprintf("--%s", long)
	}

	_, f.required = tags.Lookup("required")

	return &f, nil
}
