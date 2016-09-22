package flags

import (
	"flag"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"time"
)

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

		switch field.Type.Kind() {
		case reflect.Bool:
			err := parseBool(set, v.Field(i), field.Tag)
			if err != nil {
				return nil, err
			}

		case reflect.Float64:
			err := parseFloat64(set, v.Field(i), field.Tag)
			if err != nil {
				return nil, err
			}

		case reflect.Int64:
			if t.Field(i).Type == reflect.TypeOf(time.Duration(0)) {
				err := parseDuration(set, v.Field(i), field.Tag)
				if err != nil {
					return nil, err
				}
			} else {
				err := parseInt64(set, v.Field(i), field.Tag)
				if err != nil {
					return nil, err
				}
			}

		case reflect.Int:
			err := parseInt(set, v.Field(i), field.Tag)
			if err != nil {
				return nil, err
			}

		case reflect.String:
			parseString(set, v.Field(i), field.Tag)

		case reflect.Uint64:
			err := parseUint64(set, v.Field(i), field.Tag)
			if err != nil {
				return nil, err
			}

		case reflect.Uint:
			err := parseUint(set, v.Field(i), field.Tag)
			if err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("unexpected flag receiver field type %s", field.Type.Kind())
		}
	}

	err := set.Parse(args)
	if err != nil {
		return nil, err
	}

	return set.Args(), nil
}

func parseBool(set *flag.FlagSet, field reflect.Value, tags reflect.StructTag) error {
	var defaultValue bool
	defaultStr, ok := tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = strconv.ParseBool(defaultStr)
		if err != nil {
			return fmt.Errorf("could not parse bool default value %q: %s", defaultStr, err)
		}
	}

	short, ok := tags.Lookup("short")
	if ok {
		set.BoolVar(field.Addr().Interface().(*bool), short, defaultValue, "")
	}

	long, ok := tags.Lookup("long")
	if ok {
		set.BoolVar(field.Addr().Interface().(*bool), long, defaultValue, "")
	}

	return nil
}

func parseFloat64(set *flag.FlagSet, field reflect.Value, tags reflect.StructTag) error {
	var defaultValue float64
	defaultStr, ok := tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = strconv.ParseFloat(defaultStr, 64)
		if err != nil {
			return fmt.Errorf("could not parse float64 default value %q: %s", defaultStr, err)
		}
	}

	short, ok := tags.Lookup("short")
	if ok {
		set.Float64Var(field.Addr().Interface().(*float64), short, defaultValue, "")
	}

	long, ok := tags.Lookup("long")
	if ok {
		set.Float64Var(field.Addr().Interface().(*float64), long, defaultValue, "")
	}

	return nil
}

func parseInt64(set *flag.FlagSet, field reflect.Value, tags reflect.StructTag) error {
	var defaultValue int64
	defaultStr, ok := tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = strconv.ParseInt(defaultStr, 0, 64)
		if err != nil {
			return fmt.Errorf("could not parse int64 default value %q: %s", defaultStr, err)
		}
	}

	short, ok := tags.Lookup("short")
	if ok {
		set.Int64Var(field.Addr().Interface().(*int64), short, defaultValue, "")
	}

	long, ok := tags.Lookup("long")
	if ok {
		set.Int64Var(field.Addr().Interface().(*int64), long, defaultValue, "")
	}

	return nil
}

func parseDuration(set *flag.FlagSet, field reflect.Value, tags reflect.StructTag) error {
	var defaultValue time.Duration
	defaultStr, ok := tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = time.ParseDuration(defaultStr)
		if err != nil {
			return fmt.Errorf("could not parse duration default value %q: %s", defaultStr, err)
		}
	}

	short, ok := tags.Lookup("short")
	if ok {
		set.DurationVar(field.Addr().Interface().(*time.Duration), short, defaultValue, "")
	}

	long, ok := tags.Lookup("long")
	if ok {
		set.DurationVar(field.Addr().Interface().(*time.Duration), long, defaultValue, "")
	}

	return nil
}

func parseInt(set *flag.FlagSet, field reflect.Value, tags reflect.StructTag) error {
	var defaultValue int64
	defaultStr, ok := tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = strconv.ParseInt(defaultStr, 0, 0)
		if err != nil {
			return fmt.Errorf("could not parse int default value %q: %s", defaultStr, err)
		}
	}

	short, ok := tags.Lookup("short")
	if ok {
		set.IntVar(field.Addr().Interface().(*int), short, int(defaultValue), "")
	}

	long, ok := tags.Lookup("long")
	if ok {
		set.IntVar(field.Addr().Interface().(*int), long, int(defaultValue), "")
	}

	return nil
}

func parseString(set *flag.FlagSet, field reflect.Value, tags reflect.StructTag) {
	var defaultValue string
	defaultStr, ok := tags.Lookup("default")
	if ok {
		defaultValue = defaultStr
	}

	short, ok := tags.Lookup("short")
	if ok {
		set.StringVar(field.Addr().Interface().(*string), short, defaultValue, "")
	}

	long, ok := tags.Lookup("long")
	if ok {
		set.StringVar(field.Addr().Interface().(*string), long, defaultValue, "")
	}
}

func parseUint64(set *flag.FlagSet, field reflect.Value, tags reflect.StructTag) error {
	var defaultValue uint64
	defaultStr, ok := tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = strconv.ParseUint(defaultStr, 0, 64)
		if err != nil {
			return fmt.Errorf("could not parse uint64 default value %q: %s", defaultStr, err)
		}
	}

	short, ok := tags.Lookup("short")
	if ok {
		set.Uint64Var(field.Addr().Interface().(*uint64), short, defaultValue, "")
	}

	long, ok := tags.Lookup("long")
	if ok {
		set.Uint64Var(field.Addr().Interface().(*uint64), long, defaultValue, "")
	}

	return nil
}

func parseUint(set *flag.FlagSet, field reflect.Value, tags reflect.StructTag) error {
	var defaultValue uint64
	defaultStr, ok := tags.Lookup("default")
	if ok {
		var err error
		defaultValue, err = strconv.ParseUint(defaultStr, 0, 0)
		if err != nil {
			return fmt.Errorf("could not parse uint default value %q: %s", defaultStr, err)
		}
	}

	short, ok := tags.Lookup("short")
	if ok {
		set.UintVar(field.Addr().Interface().(*uint), short, uint(defaultValue), "")
	}

	long, ok := tags.Lookup("long")
	if ok {
		set.UintVar(field.Addr().Interface().(*uint), long, uint(defaultValue), "")
	}

	return nil
}
