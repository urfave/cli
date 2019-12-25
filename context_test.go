package cli

import (
	"context"
	"flag"
	"os"
	"sort"
	"strings"
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

func TestContext_Path(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("path", "path/to/file", "path to file")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.String("top-path", "path/to/top/file", "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	c := NewContext(nil, set, parentCtx)
	expect(t, c.Path("path"), "path/to/file")
	expect(t, c.Path("top-path"), "path/to/top/file")
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
	_ = set.Parse([]string{"--myflag", "bat", "baz"})
	expect(t, c.Args().Len(), 2)
	expect(t, c.Bool("myflag"), true)
}

func TestContext_NArg(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	c := NewContext(nil, set, nil)
	_ = set.Parse([]string{"--myflag", "bat", "baz"})
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

	_ = set.Parse([]string{"--one-flag", "--two-flag", "--three-flag", "frob"})
	_ = parentSet.Parse([]string{"--top-flag"})

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
		timeoutIsSet, tIsSet    bool
		noEnvVarIsSet, nIsSet   bool
		passwordIsSet, pIsSet   bool
		unparsableIsSet, uIsSet bool
	)

	os.Clearenv()
	_ = os.Setenv("APP_TIMEOUT_SECONDS", "15.5")
	_ = os.Setenv("APP_PASSWORD", "")
	a := App{
		Flags: []Flag{
			&Float64Flag{Name: "timeout", Aliases: []string{"t"}, EnvVars: []string{"APP_TIMEOUT_SECONDS"}},
			&StringFlag{Name: "password", Aliases: []string{"p"}, EnvVars: []string{"APP_PASSWORD"}},
			&Float64Flag{Name: "unparsable", Aliases: []string{"u"}, EnvVars: []string{"APP_UNPARSABLE"}},
			&Float64Flag{Name: "no-env-var", Aliases: []string{"n"}},
		},
		Action: func(ctx *Context) error {
			timeoutIsSet = ctx.IsSet("timeout")
			tIsSet = ctx.IsSet("t")
			passwordIsSet = ctx.IsSet("password")
			pIsSet = ctx.IsSet("p")
			unparsableIsSet = ctx.IsSet("unparsable")
			uIsSet = ctx.IsSet("u")
			noEnvVarIsSet = ctx.IsSet("no-env-var")
			nIsSet = ctx.IsSet("n")
			return nil
		},
	}
	_ = a.Run([]string{"run"})
	expect(t, timeoutIsSet, true)
	expect(t, tIsSet, true)
	expect(t, passwordIsSet, true)
	expect(t, pIsSet, true)
	expect(t, noEnvVarIsSet, false)
	expect(t, nIsSet, false)

	_ = os.Setenv("APP_UNPARSABLE", "foobar")
	_ = a.Run([]string{"run"})
	expect(t, unparsableIsSet, false)
	expect(t, uIsSet, false)
}

func TestContext_NumFlags(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	set.String("otherflag", "hello world", "doc")
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("myflagGlobal", true, "doc")
	globalCtx := NewContext(nil, globalSet, nil)
	c := NewContext(nil, set, globalCtx)
	_ = set.Parse([]string{"--myflag", "--otherflag=foo"})
	_ = globalSet.Parse([]string{"--myflagGlobal"})
	expect(t, c.NumFlags(), 2)
}

func TestContext_Set(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int("int", 5, "an int")
	c := NewContext(nil, set, nil)

	expect(t, c.IsSet("int"), false)
	_ = c.Set("int", "1")
	expect(t, c.Int("int"), 1)
	expect(t, c.IsSet("int"), true)
}

func TestContext_LocalFlagNames(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("one-flag", false, "doc")
	set.String("two-flag", "hello world", "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	_ = set.Parse([]string{"--one-flag", "--two-flag=foo"})
	_ = parentSet.Parse([]string{"--top-flag"})

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
	_ = set.Parse([]string{"--one-flag", "--two-flag=foo"})
	_ = parentSet.Parse([]string{"--top-flag"})

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
	_ = set.Parse([]string{"--local-flag"})
	_ = parentSet.Parse([]string{"--top-flag"})

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
	_ = set.Parse([]string{"--local-flag"})
	_ = parentSet.Parse([]string{"--top-flag"})

	fs := lookupFlagSet("top-flag", ctx)
	expect(t, fs, parentCtx.flagSet)

	fs = lookupFlagSet("local-flag", ctx)
	expect(t, fs, ctx.flagSet)

	if fs := lookupFlagSet("frob", ctx); fs != nil {
		t.Fail()
	}
}

func TestNonNilContext(t *testing.T) {
	ctx := NewContext(nil, nil, nil)
	if ctx.Context == nil {
		t.Fatal("expected a non nil context when no parent is present")
	}
}

// TestContextPropagation tests that
// *cli.Context always has a valid
// context.Context
func TestContextPropagation(t *testing.T) {
	parent := NewContext(nil, nil, nil)
	parent.Context = context.WithValue(context.Background(), "key", "val")
	ctx := NewContext(nil, nil, parent)
	val := ctx.Context.Value("key")
	if val == nil {
		t.Fatal("expected a parent context to be inherited but got nil")
	}
	valstr, _ := val.(string)
	if valstr != "val" {
		t.Fatalf("expected the context value to be %q but got %q", "val", valstr)
	}
	parent = NewContext(nil, nil, nil)
	parent.Context = nil
	ctx = NewContext(nil, nil, parent)
	if ctx.Context == nil {
		t.Fatal("expected context to not be nil even if the parent's context is nil")
	}
}

func TestContextAttributeAccessing(t *testing.T) {
	tdata := []struct {
		testCase        string
		setBoolInput    string
		ctxBoolInput    string
		newContextInput *Context
	}{
		{
			testCase:        "empty",
			setBoolInput:    "",
			ctxBoolInput:    "",
			newContextInput: nil,
		},
		{
			testCase:        "empty_with_background_context",
			setBoolInput:    "",
			ctxBoolInput:    "",
			newContextInput: &Context{Context: context.Background()},
		},
		{
			testCase:        "empty_set_bool_and_present_ctx_bool",
			setBoolInput:    "",
			ctxBoolInput:    "ctx-bool",
			newContextInput: nil,
		},
		{
			testCase:        "present_set_bool_and_present_ctx_bool_with_background_context",
			setBoolInput:    "",
			ctxBoolInput:    "ctx-bool",
			newContextInput: &Context{Context: context.Background()},
		},
		{
			testCase:        "present_set_bool_and_present_ctx_bool",
			setBoolInput:    "ctx-bool",
			ctxBoolInput:    "ctx-bool",
			newContextInput: nil,
		},
		{
			testCase:        "present_set_bool_and_present_ctx_bool_with_background_context",
			setBoolInput:    "ctx-bool",
			ctxBoolInput:    "ctx-bool",
			newContextInput: &Context{Context: context.Background()},
		},
		{
			testCase:        "present_set_bool_and_different_ctx_bool",
			setBoolInput:    "ctx-bool",
			ctxBoolInput:    "not-ctx-bool",
			newContextInput: nil,
		},
		{
			testCase:        "present_set_bool_and_different_ctx_bool_with_background_context",
			setBoolInput:    "ctx-bool",
			ctxBoolInput:    "not-ctx-bool",
			newContextInput: &Context{Context: context.Background()},
		},
	}

	for _, test := range tdata {
		t.Run(test.testCase, func(t *testing.T) {
			// setup
			set := flag.NewFlagSet("some-flag-set-name", 0)
			set.Bool(test.setBoolInput, false, "usage documentation")
			ctx := NewContext(nil, set, test.newContextInput)

			// logic under test
			value := ctx.Bool(test.ctxBoolInput)

			// assertions
			if value != false {
				t.Errorf("expected \"value\" to be false, but it was not")
			}
		})
	}
}

func TestCheckRequiredFlags(t *testing.T) {
	tdata := []struct {
		testCase              string
		parseInput            []string
		envVarInput           [2]string
		flags                 []Flag
		expectedAnError       bool
		expectedErrorContents []string
	}{
		{
			testCase: "empty",
		},
		{
			testCase: "optional",
			flags: []Flag{
				&StringFlag{Name: "optionalFlag"},
			},
		},
		{
			testCase: "required",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
			},
			expectedAnError:       true,
			expectedErrorContents: []string{"requiredFlag"},
		},
		{
			testCase: "required_and_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
			},
			parseInput: []string{"--requiredFlag", "myinput"},
		},
		{
			testCase: "required_and_present_via_env_var",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true, EnvVars: []string{"REQUIRED_FLAG"}},
			},
			envVarInput: [2]string{"REQUIRED_FLAG", "true"},
		},
		{
			testCase: "required_and_optional",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "optionalFlag"},
			},
			expectedAnError: true,
		},
		{
			testCase: "required_and_optional_and_optional_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "optionalFlag"},
			},
			parseInput:      []string{"--optionalFlag", "myinput"},
			expectedAnError: true,
		},
		{
			testCase: "required_and_optional_and_optional_present_via_env_var",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "optionalFlag", EnvVars: []string{"OPTIONAL_FLAG"}},
			},
			envVarInput:     [2]string{"OPTIONAL_FLAG", "true"},
			expectedAnError: true,
		},
		{
			testCase: "required_and_optional_and_required_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "optionalFlag"},
			},
			parseInput: []string{"--requiredFlag", "myinput"},
		},
		{
			testCase: "two_required",
			flags: []Flag{
				&StringFlag{Name: "requiredFlagOne", Required: true},
				&StringFlag{Name: "requiredFlagTwo", Required: true},
			},
			expectedAnError:       true,
			expectedErrorContents: []string{"requiredFlagOne", "requiredFlagTwo"},
		},
		{
			testCase: "two_required_and_one_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "requiredFlagTwo", Required: true},
			},
			parseInput:      []string{"--requiredFlag", "myinput"},
			expectedAnError: true,
		},
		{
			testCase: "two_required_and_both_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "requiredFlagTwo", Required: true},
			},
			parseInput: []string{"--requiredFlag", "myinput", "--requiredFlagTwo", "myinput"},
		},
		{
			testCase: "required_flag_with_short_name",
			flags: []Flag{
				&StringSliceFlag{Name: "names", Aliases: []string{"N"}, Required: true},
			},
			parseInput: []string{"-N", "asd", "-N", "qwe"},
		},
		{
			testCase: "required_flag_with_multiple_short_names",
			flags: []Flag{
				&StringSliceFlag{Name: "names", Aliases: []string{"N", "n"}, Required: true},
			},
			parseInput: []string{"-n", "asd", "-n", "qwe"},
		},
	}

	for _, test := range tdata {
		t.Run(test.testCase, func(t *testing.T) {
			// setup
			if test.envVarInput[0] != "" {
				os.Clearenv()
				_ = os.Setenv(test.envVarInput[0], test.envVarInput[1])
			}

			set := flag.NewFlagSet("test", 0)
			for _, flags := range test.flags {
				_ = flags.Apply(set)
			}
			_ = set.Parse(test.parseInput)

			c := &Context{}
			ctx := NewContext(c.App, set, c)
			ctx.Command.Flags = test.flags

			// logic under test
			err := checkRequiredFlags(test.flags, ctx)

			// assertions
			if test.expectedAnError && err == nil {
				t.Errorf("expected an error, but there was none")
			}
			if !test.expectedAnError && err != nil {
				t.Errorf("did not expected an error, but there was one: %s", err)
			}
			for _, errString := range test.expectedErrorContents {
				if err != nil {
					if !strings.Contains(err.Error(), errString) {
						t.Errorf("expected error %q to contain %q, but it didn't!", err.Error(), errString)
					}
				}
			}
		})
	}
}
