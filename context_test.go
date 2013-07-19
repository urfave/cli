package cli

import (
	"flag"
	"testing"
)

func Test_Int(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int("myflag", 12, "doc")
	c := NewContext(set)
	expect(t, c.Int("myflag"), 12)
}

func Test_String(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("myflag", "hello world", "doc")
	c := NewContext(set)
	expect(t, c.String("myflag"), "hello world")
}

func Test_Bool(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	c := NewContext(set)
	expect(t, c.Bool("myflag"), false)
}

func Test_Args(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	c := NewContext(set)
	set.Parse([]string{"--myflag", "bat", "baz"})
	expect(t, len(c.Args()), 2)
	expect(t, c.Bool("myflag"), true)
}
