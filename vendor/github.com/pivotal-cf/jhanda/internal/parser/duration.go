package parser

import (
	"flag"
	"fmt"
	"reflect"
	"time"
)

type Duration struct{
	Set *flag.FlagSet
	Field reflect.Value
	Tags reflect.StructTag
}

func (d Duration) Execute() error {
	var defaultValue time.Duration
	defaultStr, ok := d.Tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = time.ParseDuration(defaultStr)
		if err != nil {
			return fmt.Errorf("could not parse duration default value %q: %s", defaultStr, err)
		}
	}

	short, ok := d.Tags.Lookup("short")
	if ok {
		d.Set.DurationVar(d.Field.Addr().Interface().(*time.Duration), short, defaultValue, "")
	}

	long, ok := d.Tags.Lookup("long")
	if ok {
		d.Set.DurationVar(d.Field.Addr().Interface().(*time.Duration), long, defaultValue, "")
	}

	return nil
}