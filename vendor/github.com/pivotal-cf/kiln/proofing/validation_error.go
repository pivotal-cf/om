package proofing

import (
	"fmt"
	"reflect"
	"strings"
)

type ValidationError struct {
	Kind    interface{}
	Message string
}

func NewValidationError(kind interface{}, message string) ValidationError {
	return ValidationError{
		Kind:    kind,
		Message: message,
	}
}

func (ve ValidationError) Error() string {
	t := reflect.TypeOf(ve.Kind)
	return fmt.Sprintf("%s %s", strings.ToLower(t.Name()), ve.Message)
}

func ValidatePresence(err error, v interface{}, field string) error {
	value := reflect.ValueOf(v).FieldByName(field)
	if value.Len() == 0 {
		validationError := NewValidationError(v, fmt.Sprintf("%s must be present", strings.ToLower(field)))
		switch e := err.(type) {
		case *CompoundError:
			e.Add(validationError)
		case ValidationError:
			err = &CompoundError{err, validationError}
		default:
			err = validationError
		}
	}

	return err
}
