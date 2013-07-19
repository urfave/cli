package cli

import (
	"reflect"
	"testing"
)

func Test_SettingFlags(t *testing.T) {
	Flags = []Flag{
		StringFlag{"foo", "default", "a string flag"},
		IntFlag{"bar", 42, "an int flag"},
		BoolFlag{"bat", "a bool flag"},
	}
	Action = func(c *Context) {
		expect(t, c.String("foo"), "hello world")
		expect(t, c.Int("bar"), 245)
		expect(t, c.Bool("bat"), true)
	}
	Run([]string{"command", "--foo", "hello world", "--bar", "245", "--bat"})
}

func Test_FlagDefaults(t *testing.T) {
	Flags = []Flag{
		StringFlag{"foo", "default", "a string flag"},
		IntFlag{"bar", 42, "an int flag"},
		BoolFlag{"bat", "a bool flag"},
	}
	Action = func(c *Context) {
		expect(t, c.String("foo"), "default")
		expect(t, c.Int("bar"), 42)
		expect(t, c.Bool("bat"), false)
	}
	Run([]string{"command"})
}

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
