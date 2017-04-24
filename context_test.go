package cli

import (
	"flag"
	"os"
	"sort"
	"testing"
	"time"
)

func TestNewContext(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int("myflag", 12, "doc")
	set.Int64("myflagInt64", int64(12), "doc")
	set.Uint("myflagUint", uint(93), "doc")
	set.Uint64("myflagUint64", uint64(93), "doc")
	set.Float64("myflag64", float64(17), "doc")
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Int("myflag", 42, "doc")
	globalSet.Int64("myflagInt64", int64(42), "doc")
	globalSet.Uint("myflagUint", uint(33), "doc")
	globalSet.Uint64("myflagUint64", uint64(33), "doc")
	globalSet.Float64("myflag64", float64(47), "doc")
	globalCtx := NewContext(nil, globalSet, nil)
	command := &Command{Name: "mycommand"}
	c := NewContext(nil, set, globalCtx)
	c.Command = command
	expect(t, c.Int("myflag"), 12)
	expect(t, c.Int64("myflagInt64"), int64(12))
	expect(t, c.Uint("myflagUint"), uint(93))
	expect(t, c.Uint64("myflagUint64"), uint64(93))
	expect(t, c.Float64("myflag64"), float64(17))
	expect(t, c.Command.Name, "mycommand")
}

func TestContext_Int(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int("myflag", 12, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Int("top-flag", 13, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	c := NewContext(nil, set, parentCtx)
	expect(t, c.Int("myflag"), 12)
	expect(t, c.Int("top-flag"), 13)
}

func TestContext_Int64(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int64("myflagInt64", 12, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Int64("top-flag", 13, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	c := NewContext(nil, set, parentCtx)
	expect(t, c.Int64("myflagInt64"), int64(12))
	expect(t, c.Int64("top-flag"), int64(13))
}

func TestContext_Uint(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Uint("myflagUint", uint(13), "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Uint("top-flag", uint(14), "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	c := NewContext(nil, set, parentCtx)
	expect(t, c.Uint("myflagUint"), uint(13))
	expect(t, c.Uint("top-flag"), uint(14))
}

func TestContext_Uint64(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Uint64("myflagUint64", uint64(9), "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Uint64("top-flag", uint64(10), "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	c := NewContext(nil, set, parentCtx)
	expect(t, c.Uint64("myflagUint64"), uint64(9))
	expect(t, c.Uint64("top-flag"), uint64(10))
}

func TestContext_Float64(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Float64("myflag", float64(17), "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Float64("top-flag", float64(18), "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	c := NewContext(nil, set, parentCtx)
	expect(t, c.Float64("myflag"), float64(17))
	expect(t, c.Float64("top-flag"), float64(18))
}

func TestContext_Duration(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Duration("myflag", 12*time.Second, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Duration("top-flag", 13*time.Second, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	c := NewContext(nil, set, parentCtx)
	expect(t, c.Duration("myflag"), 12*time.Second)
	expect(t, c.Duration("top-flag"), 13*time.Second)
}

func TestContext_String(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("myflag", "hello world", "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.String("top-flag", "hai veld", "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	c := NewContext(nil, set, parentCtx)
	expect(t, c.String("myflag"), "hello world")
	expect(t, c.String("top-flag"), "hai veld")
}

func TestContext_Bool(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	c := NewContext(nil, set, parentCtx)
	expect(t, c.Bool("myflag"), false)
	expect(t, c.Bool("top-flag"), true)
}

func TestContext_Args(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	c := NewContext(nil, set, nil)
	set.Parse([]string{"--myflag", "bat", "baz"})
	expect(t, c.Args().Len(), 2)
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

// XXX Corresponds to hack in context.IsSet for flags with EnvVar field
// Should be moved to `flag_test` in v2
func TestContext_IsSet_fromEnv(t *testing.T) {
	var (
		timeoutIsSet, tIsSet, noEnvVarIsSet, nIsSet             bool
		globalTimeoutIsSet, TIsSet, globalNoEnvVarIsSet, NIsSet bool
	)

	os.Clearenv()
	os.Setenv("GLOBAL_APP_TIMEOUT_SECONDS", "15.5")
	os.Setenv("APP_TIMEOUT_SECONDS", "15.5")
	a := App{
		Flags: []Flag{
			&Float64Flag{
				Name:    "global-timeout",
				Aliases: []string{"T"},
				EnvVars: []string{"GLOBAL_APP_TIMEOUT_SECONDS"},
			},
			&Float64Flag{
				Name:    "global-no-env-var",
				Aliases: []string{"N"},
			},
		},
		Commands: []*Command{
			{
				Name: "hello",
				Flags: []Flag{
					&Float64Flag{
						Name:    "timeout",
						Aliases: []string{"t"},
						EnvVars: []string{"APP_TIMEOUT_SECONDS"},
					},
					&Float64Flag{
						Name:    "no-env-var",
						Aliases: []string{"n"},
					},
				},
				Action: func(ctx *Context) error {
					globalTimeoutIsSet = ctx.IsSet("global-timeout")
					TIsSet = ctx.IsSet("T")
					globalNoEnvVarIsSet = ctx.IsSet("global-no-env-var")
					NIsSet = ctx.IsSet("N")
					timeoutIsSet = ctx.IsSet("timeout")
					tIsSet = ctx.IsSet("t")
					noEnvVarIsSet = ctx.IsSet("no-env-var")
					nIsSet = ctx.IsSet("n")
					return nil
				},
			},
		},
	}
	a.Run([]string{"run", "hello"})
	expect(t, globalTimeoutIsSet, true)
	expect(t, TIsSet, true)
	expect(t, globalNoEnvVarIsSet, false)
	expect(t, NIsSet, false)
	expect(t, timeoutIsSet, true)
	expect(t, tIsSet, true)
	expect(t, noEnvVarIsSet, false)
	expect(t, nIsSet, false)
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
