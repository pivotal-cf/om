package jhanda

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Usage provides all of the details describing a Command, including a
// description, a shorter description (used when display a list of commands),
// and the flag options offered by the Command.
type Usage struct {
	Description      string
	ShortDescription string
	Flags            interface{}
}

// PrintUsage will return a string representation of the options provided by a
// Command flag set.
func PrintUsage(receiver interface{}) (string, error) {
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
		var kindParts []string
		if _, ok := field.Tag.Lookup("required"); ok {
			kindParts = append(kindParts, "required")
		}

		kind := field.Type.Kind().String()
		if kind == reflect.Slice.String() {
			kind = field.Type.Elem().Kind().String()
			kindParts = append(kindParts, "variadic")
		}

		if len(kindParts) > 0 {
			kind = fmt.Sprintf("%s (%s)", kind, strings.Join(kindParts, ", "))
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

	sort.Strings(usage)

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
