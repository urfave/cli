package cli

import (
	"os"
	"reflect"
	"testing"
)

func init() {
	_ = os.Setenv("CLI_TEMPLATE_REPANIC", "1")
}

func expect(t *testing.T, a interface{}, b interface{}) {
	t.Helper()

	if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func expectNotEqual(t *testing.T, a interface{}, b interface{}) {
	t.Helper()

	if reflect.DeepEqual(a, b) {
		t.Errorf("Expected not equal, but got: %v (type %v), %v (type %v) ", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
