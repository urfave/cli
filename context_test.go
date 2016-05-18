package cli

import (
	"flag"
	"sort"
	"testing"
	"time"
)

func TestNewContext(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int("myflag", 12, "doc")
	set.Float64("myflag64", float64(17), "doc")
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Int("myflag", 42, "doc")
	globalSet.Float64("myflag64", float64(47), "doc")
	globalCtx := NewContext(nil, globalSet, nil)
	command := Command{Name: "mycommand"}
	c := NewContext(nil, set, globalCtx)
	c.Command = command
	expect(t, c.Int("myflag"), 12)
	expect(t, c.Float64("myflag64"), float64(17))
	expect(t, c.Command.Name, "mycommand")
}

func TestContext_Int(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int("myflag", 12, "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.Int("myflag"), 12)
}

func TestContext_Float64(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Float64("myflag", float64(17), "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.Float64("myflag"), float64(17))
}

func TestContext_Duration(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Duration("myflag", time.Duration(12*time.Second), "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.Duration("myflag"), time.Duration(12*time.Second))
}

func TestContext_String(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("myflag", "hello world", "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.String("myflag"), "hello world")
}

func TestContext_Bool(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.Bool("myflag"), false)
}

func TestContext_Args(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	c := NewContext(nil, set, nil)
	set.Parse([]string{"--myflag", "bat", "baz"})
	expect(t, len(c.Args()), 2)
	expect(t, c.Bool("myflag"), true)
}

func TestContext_NArg(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	c := NewContext(nil, set, nil)
	set.Parse([]string{"--myflag", "bat", "baz"})
	expect(t, c.NArg(), 2)
}

func TestContext_IsSet(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("one-flag", false, "doc")
	set.Bool("two-flag", false, "doc")
	set.String("three-flag", "hello world", "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)

	set.Parse([]string{"--one-flag", "--two-flag", "--three-flag", "frob"})
	parentSet.Parse([]string{"--top-flag"})

	expect(t, ctx.IsSet("one-flag"), true)
	expect(t, ctx.IsSet("two-flag"), true)
	expect(t, ctx.IsSet("three-flag"), true)
	expect(t, ctx.IsSet("top-flag"), true)
	expect(t, ctx.IsSet("bogus"), false)
}

func TestContext_NumFlags(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	set.String("otherflag", "hello world", "doc")
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("myflagGlobal", true, "doc")
	globalCtx := NewContext(nil, globalSet, nil)
	c := NewContext(nil, set, globalCtx)
	set.Parse([]string{"--myflag", "--otherflag=foo"})
	globalSet.Parse([]string{"--myflagGlobal"})
	expect(t, c.NumFlags(), 2)
}

func TestContext_Set(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int("int", 5, "an int")
	c := NewContext(nil, set, nil)

	c.Set("int", "1")
	expect(t, c.Int("int"), 1)
}

func TestContext_LocalFlagNames(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("one-flag", false, "doc")
	set.String("two-flag", "hello world", "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	set.Parse([]string{"--one-flag", "--two-flag=foo"})
	parentSet.Parse([]string{"--top-flag"})

	actualFlags := ctx.LocalFlagNames()
	sort.Strings(actualFlags)

	expect(t, actualFlags, []string{"one-flag", "two-flag"})
}

func TestContext_FlagNames(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("one-flag", false, "doc")
	set.String("two-flag", "hello world", "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	set.Parse([]string{"--one-flag", "--two-flag=foo"})
	parentSet.Parse([]string{"--top-flag"})

	actualFlags := ctx.FlagNames()
	sort.Strings(actualFlags)

	expect(t, actualFlags, []string{"one-flag", "top-flag", "two-flag"})
}

func TestContext_Lineage(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("local-flag", false, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	set.Parse([]string{"--local-flag"})
	parentSet.Parse([]string{"--top-flag"})

	lineage := ctx.Lineage()
	expect(t, len(lineage), 2)
	expect(t, lineage[0], ctx)
	expect(t, lineage[1], parentCtx)
}

func TestContext_lookupFlagSet(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("local-flag", false, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	set.Parse([]string{"--local-flag"})
	parentSet.Parse([]string{"--top-flag"})

	fs := lookupFlagSet("top-flag", ctx)
	expect(t, fs, parentCtx.flagSet)

	fs = lookupFlagSet("local-flag", ctx)
	expect(t, fs, ctx.flagSet)

	if fs := lookupFlagSet("frob", ctx); fs != nil {
		t.Fail()
	}
}
