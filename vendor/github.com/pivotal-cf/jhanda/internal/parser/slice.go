package parser

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
)

type Slice struct {
	Set   *flag.FlagSet
	Field reflect.Value
	Tags  reflect.StructTag
}

func (s Slice) Execute() error {
	collection := s.Field.Addr().Interface().(*[]string)

	defaultSlice, ok := s.Tags.Lookup("default")
	if ok {
		separated := strings.Split(defaultSlice, ",")
		*collection = append(*collection, separated...)
	}

	slice := StringSlice{collection}

	short, ok := s.Tags.Lookup("short")
	if ok {
		s.Set.Var(&slice, short, "")
	}

	long, ok := s.Tags.Lookup("long")
	if ok {
		s.Set.Var(&slice, long, "")
	}

	return nil
}

type StringSlice struct {
	slice *[]string
}

func (ss *StringSlice) String() string {
	return fmt.Sprintf("%s", ss.slice)
}

func (ss *StringSlice) Set(item string) error {
	*ss.slice = append(*ss.slice, item)
	return nil
}
