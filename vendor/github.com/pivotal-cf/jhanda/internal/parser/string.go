package parser

import (
	"flag"
	"reflect"
)

type String struct {
	Set   *flag.FlagSet
	Field reflect.Value
	Tags  reflect.StructTag
}

func (s String) Execute() error {
	var defaultValue string
	defaultStr, ok := s.Tags.Lookup("default")
	if ok {
		defaultValue = defaultStr
	}

	short, ok := s.Tags.Lookup("short")
	if ok {
		s.Set.StringVar(s.Field.Addr().Interface().(*string), short, defaultValue, "")
	}

	long, ok := s.Tags.Lookup("long")
	if ok {
		s.Set.StringVar(s.Field.Addr().Interface().(*string), long, defaultValue, "")
	}

	return nil
}
