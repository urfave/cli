package cli

import (
  "reflect"
  "testing"
)

func Test_SimpleCLIFlags(t *testing.T) {
  Flags = []Flag{
    StringFlag{"foo", "default", "a foo flag"},
  }
  Action = func(c *Context) {
    expect(t, c.String("foo"), "hello world")
  }
  Run([]string{ "command", "--foo", "hello world" })
}

/* Helpers */

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
