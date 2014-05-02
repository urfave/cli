package cli_test

import (
	"reflect"
	"testing"
)

type (
	FlagTestString struct {
		name     string
		value    string
		expected string
	}

	FlagTest struct {
		name     string
		expected string
	}

	FlagTestBool struct {
		name     string
		expected bool
	}
)

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
