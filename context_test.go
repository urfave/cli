package cli

import (
	"sort"
	"testing"
	"time"
)

func TestNewContext(t *testing.T) {
	set := NewFlagSet("test", []Flag{
		&IntFlag{
			Name:  "myflag",
			Value: 12,
		},
		&Float64Flag{
			Name:  "myflag64",
			Value: float64(17),
		},
	}, nil)

	topSet := NewFlagSet("test", []Flag{
		&IntFlag{
			Name:  "myflag",
			Value: 42,
		},
		&Float64Flag{
			Name:  "myflag64",
			Value: float64(47),
		},
	}, nil)

	topCtx := NewContext(nil, topSet, nil)
	command := &Command{Name: "mycommand"}
	c := NewContext(nil, set, topCtx)
	c.Command = command

	expect(t, c.Int("myflag"), 12)
	expect(t, c.Float64("myflag64"), float64(17))
	expect(t, c.Command.Name, "mycommand")
}

func TestContext_Int(t *testing.T) {
	set := NewFlagSet("test",
		[]Flag{&IntFlag{Name: "myflag", Value: 12}}, nil)
	c := NewContext(nil, set, nil)
	expect(t, c.Int("myflag"), 12)
}

func TestContext_Float64(t *testing.T) {
	set := NewFlagSet("test",
		[]Flag{&Float64Flag{Name: "myflag", Value: float64(17)}}, nil)
	c := NewContext(nil, set, nil)
	expect(t, c.Float64("myflag"), float64(17))
}

func TestContext_Duration(t *testing.T) {
	set := NewFlagSet("test",
		[]Flag{&DurationFlag{Name: "myflag", Value: time.Duration(12 * time.Second)}}, nil)
	c := NewContext(nil, set, nil)
	expect(t, c.Duration("myflag"), time.Duration(12*time.Second))
}

func TestContext_String(t *testing.T) {
	set := NewFlagSet("test",
		[]Flag{&StringFlag{Name: "myflag", Value: "hello world"}}, nil)
	c := NewContext(nil, set, nil)
	expect(t, c.String("myflag"), "hello world")
}

func TestContext_Bool(t *testing.T) {
	set := NewFlagSet("test", []Flag{&BoolFlag{Name: "myflag"}}, nil)
	c := NewContext(nil, set, nil)
	expect(t, c.Bool("myflag"), false)
}

func TestContext_Args(t *testing.T) {
	set := NewFlagSet("test",
		[]Flag{&BoolFlag{Name: "myflag"}},
		[]string{"--myflag", "bat", "baz"})
	err := set.Parse()
	expect(t, err, nil)

	c := NewContext(nil, set, nil)
	expect(t, c.NumArgs(), 2)
	expect(t, c.Bool("myflag"), true)
}

func TestContext_NumArgs(t *testing.T) {
	set := NewFlagSet("test",
		[]Flag{&BoolFlag{Name: "myflag"}},
		[]string{"--myflag", "bat", "baz"})
	err := set.Parse()
	expect(t, err, nil)
	c := NewContext(nil, set, nil)
	expect(t, c.NumArgs(), 2)
}

func TestContext_IsSet(t *testing.T) {
	set := NewFlagSet("test", []Flag{
		&BoolFlag{Name: "one-flag"},
		&BoolFlag{Name: "two-flag"},
		&StringFlag{Name: "three-flag", Value: "hello world"},
	}, []string{"--one-flag", "--two-flag", "--three-flag", "frob"})
	err := set.Parse()
	expect(t, err, nil)

	parentSet := NewFlagSet("test",
		[]Flag{
			&BoolFlag{Name: "top-flag", Value: true},
		}, []string{"--top-flag"})
	err = parentSet.Parse()
	expect(t, err, nil)

	ctx := NewContext(nil, set, NewContext(nil, parentSet, nil))

	expect(t, ctx.IsSet("one-flag"), true)
	expect(t, ctx.IsSet("two-flag"), true)
	expect(t, ctx.IsSet("three-flag"), true)
	expect(t, ctx.IsSet("top-flag"), true)
	expect(t, ctx.IsSet("bogus"), false)
}

func TestContext_NumFlags(t *testing.T) {
	set := NewFlagSet("test", []Flag{
		&BoolFlag{Name: "myflag"},
		&StringFlag{Name: "otherflag", Value: "hello world"},
	}, []string{"--myflag", "--otherflag=foo"})
	err := set.Parse()
	expect(t, err, nil)

	parentSet := NewFlagSet("test", []Flag{
		&BoolFlag{Name: "myflagGlobal", Value: true},
	}, []string{"--myflagGlobal"})
	err = parentSet.Parse()
	expect(t, err, nil)

	c := NewContext(nil, set, NewContext(nil, parentSet, nil))
	expect(t, c.NumFlags(), 2)
}

func TestContext_Set(t *testing.T) {
	set := NewFlagSet("test", []Flag{&IntFlag{Name: "int", Value: 5}}, nil)
	c := NewContext(nil, set, nil)

	c.Set("int", "1")
	expect(t, c.Int("int"), 1)
}

func TestContext_LocalFlagNames(t *testing.T) {
	set := NewFlagSet("test", []Flag{
		&BoolFlag{Name: "one-flag"},
		&StringFlag{Name: "two-flag", Value: "hello world"},
	}, []string{"--one-flag", "--two-flag=foo"})
	err := set.Parse()
	expect(t, err, nil)

	parentSet := NewFlagSet("test", []Flag{
		&BoolFlag{Name: "top-flag", Value: true},
	}, []string{"--top-flag"})
	err = parentSet.Parse()
	expect(t, err, nil)

	ctx := NewContext(nil, set, NewContext(nil, parentSet, nil))

	actualFlags := ctx.LocalFlagNames()
	sort.Strings(actualFlags)

	expect(t, actualFlags, []string{"one-flag", "two-flag"})
}

func TestContext_FlagNames(t *testing.T) {
	set := NewFlagSet("test", []Flag{
		&BoolFlag{Name: "one-flag"},
		&StringFlag{Name: "two-flag", Value: "hello world"},
	}, []string{"--one-flag", "--two-flag=foo"})
	err := set.Parse()
	expect(t, err, nil)

	parentSet := NewFlagSet("test", []Flag{
		&BoolFlag{Name: "top-flag", Value: true},
	}, []string{"--top-flag"})
	err = parentSet.Parse()
	expect(t, err, nil)

	ctx := NewContext(nil, set, NewContext(nil, parentSet, nil))

	actualFlags := ctx.FlagNames()
	sort.Strings(actualFlags)

	expect(t, actualFlags, []string{"one-flag", "top-flag", "two-flag"})
}

func TestContext_Lineage(t *testing.T) {
	set := NewFlagSet("test", []Flag{
		&BoolFlag{Name: "local-flag"},
	}, []string{"--local-flag"})
	err := set.Parse()
	expect(t, err, nil)

	parentSet := NewFlagSet("test", []Flag{
		&BoolFlag{Name: "top-flag", Value: true},
	}, []string{"--top-flag"})
	err = parentSet.Parse()
	expect(t, err, nil)

	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)

	lineage := ctx.Lineage()
	expect(t, len(lineage), 2)
	expect(t, lineage[0], ctx)
	expect(t, lineage[1], parentCtx)
}

func TestContext_lookupFlagSet(t *testing.T) {
	set := NewFlagSet("test", []Flag{
		&BoolFlag{Name: "local-flag"},
	}, []string{"--local-flag"})
	err := set.Parse()
	expect(t, err, nil)

	parentSet := NewFlagSet("test", []Flag{
		&BoolFlag{Name: "top-flag", Value: true},
	}, []string{"--top-flag"})
	err = parentSet.Parse()
	expect(t, err, nil)

	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)

	fs := ctx.lookupFlagSet("top-flag")
	expect(t, fs, parentCtx.flagSet)

	fs = ctx.lookupFlagSet("local-flag")
	expect(t, fs, ctx.flagSet)

	if fs := ctx.lookupFlagSet("frob"); fs != nil {
		t.Fail()
	}
}
