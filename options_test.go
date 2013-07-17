package cli

import "testing"
import "reflect"

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (%v) - Got %v (%v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (%v) - Got %v (%v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func Test_Int(t *testing.T) {
	opts := Options{
		"foo": 1,
		"bar": 2,
		"bat": 3,
	}

	expect(t, opts.Int("foo"), 1)
	expect(t, opts.Int("bar"), 2)
	expect(t, opts.Int("bat"), 3)
	refute(t, opts.Int("foo"), "1")
	expect(t, opts.Int("nope"), 0)
}

func Test_String(t *testing.T) {
	opts := Options{
		"foo": "bar",
		"bat": "baz",
	}

	expect(t, opts.String("foo"), "bar")
	expect(t, opts.String("bat"), "baz")
	expect(t, opts.String("nope"), "")
}

func Test_Bool(t *testing.T) {
	opts := Options{
		"foo": false,
		"bar": true,
	}

	expect(t, opts.Bool("foo"), false)
	expect(t, opts.Bool("bar"), true)
	expect(t, opts.Bool("nope"), false)
}
