package flags

import (
	"fmt"
	"reflect"
	"strings"
)

func Usage(receiver interface{}) (string, error) {
	v := reflect.ValueOf(receiver)
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("unexpected pointer to non-struct type %s", t.Kind())
	}

	var fields []reflect.StructField

	for i := 0; i < t.NumField(); i++ {
		fields = append(fields, t.Field(i))
	}

	var usage []string
	var length int
	for _, field := range fields {
		var longShort string
		long, ok := field.Tag.Lookup("long")
		if ok {
			longShort += fmt.Sprintf("--%s", long)
		}

		short, ok := field.Tag.Lookup("short")
		if ok {
			if longShort != "" {
				longShort += ", "
			}
			longShort += fmt.Sprintf("-%s", short)
		}

		if len(longShort) > length {
			length = len(longShort)
		}

		usage = append(usage, longShort)
	}

	for i, line := range usage {
		usage[i] = pad(line, " ", length)
	}

	for i, field := range fields {
		kind := field.Type.Kind().String()
		if kind == reflect.Slice.String() {
			kind = fmt.Sprintf("%s (variadic)", field.Type.Elem().Kind().String())
		}

		line := fmt.Sprintf("%s  %s", usage[i], kind)

		if len(line) > length {
			length = len(line)
		}

		usage[i] = line
	}

	for i, line := range usage {
		usage[i] = pad(line, " ", length)
	}

	for i, field := range fields {
		description, ok := field.Tag.Lookup("description")
		if ok {
			usage[i] = fmt.Sprintf("%s  %s", usage[i], description)
		}
	}

	for i, field := range fields {
		defaultValue, ok := field.Tag.Lookup("default")
		if ok {
			usage[i] = fmt.Sprintf("%s (default: %s)", usage[i], defaultValue)
		}
	}

	return strings.Join(usage, "\n"), nil
}

func pad(str, pad string, length int) string {
	for {
		str += pad
		if len(str) > length {
			return str[0:length]
		}
	}
}
