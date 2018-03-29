package jhanda

import (
	"flag"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/pivotal-cf/jhanda/internal/parser"
)

// Parse will populate the "receiver" object with the parsed values of the
// given "args". If the parser encounters a non-flag value, it will stop
// parsing and return the remainer of arguments. Parse will return an error in
// the case that the flags cannot be parsed, or a required flag is missing.
// The receiver is expected to be a pointer.
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

	var flags []*parser.Flag

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		var (
			f   *parser.Flag
			err error
		)

		switch {
		case field.Type.Kind() == reflect.Bool:
			f, err = parser.NewBool(set, v.Field(i), field.Tag)

		case field.Type.Kind() == reflect.Float64:
			f, err = parser.NewFloat64(set, v.Field(i), field.Tag)

		case field.Type == reflect.TypeOf(time.Duration(0)):
			f, err = parser.NewDuration(set, v.Field(i), field.Tag)

		case field.Type.Kind() == reflect.Int64:
			f, err = parser.NewInt64(set, v.Field(i), field.Tag)

		case field.Type.Kind() == reflect.Int:
			f, err = parser.NewInt(set, v.Field(i), field.Tag)

		case field.Type.Kind() == reflect.String:
			f, err = parser.NewString(set, v.Field(i), field.Tag)

		case field.Type.Kind() == reflect.Uint64:
			f, err = parser.NewUint64(set, v.Field(i), field.Tag)

		case field.Type.Kind() == reflect.Uint:
			f, err = parser.NewUint(set, v.Field(i), field.Tag)

		case field.Type.Kind() == reflect.Slice:
			f, err = parser.NewSlice(set, v.Field(i), field.Tag)

		default:
			return nil, fmt.Errorf("unexpected flag receiver field type %s", field.Type.Kind())
		}
		if err != nil {
			return nil, err
		}

		flags = append(flags, f)
	}

	err := set.Parse(args)
	if err != nil {
		return nil, err
	}

	set.Visit(func(ef *flag.Flag) {
		for _, ff := range flags {
			ff.SetIfMatched(ef)
		}
	})

	for _, ff := range flags {
		if err := ff.Validate(); err != nil {
			return nil, err
		}
	}

	return set.Args(), nil
}
