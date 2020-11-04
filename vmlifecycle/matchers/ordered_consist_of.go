package matchers

import (
	"fmt"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"reflect"
)

type orderconsistofMatcher struct {
	Elements []interface{}
}

func OrderedConsistOf(elements ...interface{}) types.GomegaMatcher {
	return &orderconsistofMatcher{
		Elements: elements,
	}
}

func isArrayOrSlice(a interface{}) bool {
	if a == nil {
		return false
	}
	switch reflect.TypeOf(a).Kind() {
	case reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}

func (matcher *orderconsistofMatcher) Match(actual interface{}) (success bool, err error) {
	if !isArrayOrSlice(actual) {
		return false, fmt.Errorf("ConsistOf matcher expects an array/slice/map.  Got:\n%s", format.Object(actual, 1))
	}

	elements := matcher.Elements
	if len(matcher.Elements) == 1 && isArrayOrSlice(matcher.Elements[0]) {
		elements = []interface{}{}
		value := reflect.ValueOf(matcher.Elements[0])
		for i := 0; i < value.Len(); i++ {
			elements = append(elements, value.Index(i).Interface())
		}
	}

	values := []interface{}{}
	value := reflect.ValueOf(actual)
	for i := 0; i < value.Len(); i++ {
		values = append(values, value.Index(i).Interface())
	}

	matchers := []types.GomegaMatcher{}
	for _, element := range elements {
		matcher, isMatcher := element.(types.GomegaMatcher)
		if !isMatcher {
			matcher = gomega.BeEquivalentTo(element)
		}
		matchers = append(matchers, matcher)
	}

	if reflect.ValueOf(actual).Len() != len(matchers) {
		return false, nil
	}

	for j, matcher := range matchers {
		matched, err := matcher.Match(values[j])
		if err != nil {
			return false, nil
		}
		if !matched {
			return false, nil
		}
	}

	return true, nil
}

func (matcher *orderconsistofMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to consist exactly of", matcher.Elements)

}

func (matcher *orderconsistofMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to consist exactly of", matcher.Elements)

}
