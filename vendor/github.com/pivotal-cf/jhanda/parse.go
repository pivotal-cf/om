package jhanda

import (
	"flag"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/pivotal-cf/jhanda/internal/parser"
)

type parsable interface{
	Execute() error
}

func Parse(receiver interface{}, args []string) ([]string, error) {
	set := flag.NewFlagSet("", flag.ContinueOnError)
	set.SetOutput(ioutil.Discard)
	set.Usage = func() {}

	ptr := reflect.ValueOf(receiver)
	if ptr.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("unexpected non-pointer type %s for flag receiver", ptr.Kind())
	}

	v := ptr.Elem()
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unexpected pointer to non-struct type %s", t.Kind())
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		var p parsable

		switch {
		case field.Type.Kind() == reflect.Bool:
			p = parser.Bool{set, v.Field(i), field.Tag}
		case field.Type.Kind() == reflect.Float64:
			p = parser.Float64{set, v.Field(i), field.Tag}
		case field.Type == reflect.TypeOf(time.Duration(0)):
			p = parser.Duration{set, v.Field(i), field.Tag}
		case field.Type.Kind() == reflect.Int64:
			p = parser.Int64{set, v.Field(i), field.Tag}
		case field.Type.Kind() == reflect.Int:
			p = parser.Int{set, v.Field(i), field.Tag}
		case field.Type.Kind() == reflect.String:
			p = parser.String{set, v.Field(i), field.Tag}
		case field.Type.Kind() == reflect.Uint64:
			p = parser.Uint64{set, v.Field(i), field.Tag}
		case field.Type.Kind() == reflect.Uint:
			p = parser.Uint{set, v.Field(i), field.Tag}
		case field.Type.Kind() == reflect.Slice:
			p = parser.Slice{set, v.Field(i), field.Tag}
		default:
			return nil, fmt.Errorf("unexpected flag receiver field type %s", field.Type.Kind())
		}

		err := p.Execute()
		if err != nil {
			return nil, err
		}
	}

	err := set.Parse(args)
	if err != nil {
		return nil, err
	}

	return set.Args(), nil
}
