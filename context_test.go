package cli

import (
  "flag"
  "reflect"
  "testing"
)

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

func Test_IntFlag(t *testing.T) {
  set := flag.NewFlagSet("test", 0)
  set.Int("myflag", 12, "doc")
  c := NewContext(set)
  expect(t, c.IntFlag("myflag"), 12)
}

func Test_StringFlag(t *testing.T) {
  set := flag.NewFlagSet("test", 0)
  set.String("myflag", "hello world", "doc")
  c := NewContext(set)
  expect(t, c.StringFlag("myflag"), "hello world")
}

func Test_BoolFlag(t *testing.T) {
  set := flag.NewFlagSet("test", 0)
  set.Bool("myflag", false, "doc")
  c := NewContext(set)
  expect(t, c.BoolFlag("myflag"), false)
}

func Test_Args(t *testing.T) {
  set := flag.NewFlagSet("test", 0)
  set.Bool("myflag", false, "doc")
  c := NewContext(set)
  set.Parse([]string{"--myflag", "bat", "baz"})
  expect(t, len(c.Args()), 2)
  expect(t, c.BoolFlag("myflag"), true)
}
