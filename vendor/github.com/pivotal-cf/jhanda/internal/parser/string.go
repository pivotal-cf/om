package parser

import (
	"flag"
	"fmt"
	"reflect"
)

func NewString(set *flag.FlagSet, field reflect.Value, tags reflect.StructTag) (*Flag, error) {
	var defaultValue string
	defaultStr, ok := tags.Lookup("default")
	if ok {
		defaultValue = defaultStr
	}

	var f Flag
	short, ok := tags.Lookup("short")
	if ok {
		set.StringVar(field.Addr().Interface().(*string), short, defaultValue, "")
		f.flags = append(f.flags, set.Lookup(short))
		f.name = fmt.Sprintf("-%s", short)
	}

	long, ok := tags.Lookup("long")
	if ok {
		set.StringVar(field.Addr().Interface().(*string), long, defaultValue, "")
		f.flags = append(f.flags, set.Lookup(long))
		f.name = fmt.Sprintf("--%s", long)
	}

	_, f.required = tags.Lookup("required")

	return &f, nil
}
