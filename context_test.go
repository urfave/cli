package cli

import (
	"flag"
	"os"
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
	command := Command{Name: "mycommand"}
	c := NewContext(nil, set, globalCtx)
	c.Command = command
	expect(t, c.Int("myflag"), 12)
	expect(t, c.Int64("myflagInt64"), int64(12))
	expect(t, c.Uint("myflagUint"), uint(93))
	expect(t, c.Uint64("myflagUint64"), uint64(93))
	expect(t, c.Float64("myflag64"), float64(17))
	expect(t, c.GlobalInt("myflag"), 42)
	expect(t, c.GlobalInt64("myflagInt64"), int64(42))
	expect(t, c.GlobalUint("myflagUint"), uint(33))
	expect(t, c.GlobalUint64("myflagUint64"), uint64(33))
	expect(t, c.GlobalFloat64("myflag64"), float64(47))
	expect(t, c.Command.Name, "mycommand")
}

func TestContext_Int(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int("myflag", 12, "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.Int("myflag"), 12)
}

func TestContext_Int64(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int64("myflagInt64", 12, "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.Int64("myflagInt64"), int64(12))
}

func TestContext_Uint(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Uint("myflagUint", uint(13), "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.Uint("myflagUint"), uint(13))
}

func TestContext_Uint64(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Uint64("myflagUint64", uint64(9), "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.Uint64("myflagUint64"), uint64(9))
}

func TestContext_GlobalInt(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int("myflag", 12, "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.GlobalInt("myflag"), 12)
	expect(t, c.GlobalInt("nope"), 0)
}

func TestContext_GlobalInt64(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Int64("myflagInt64", 12, "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.GlobalInt64("myflagInt64"), int64(12))
	expect(t, c.GlobalInt64("nope"), int64(0))
}

func TestContext_Float64(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Float64("myflag", float64(17), "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.Float64("myflag"), float64(17))
}

func TestContext_GlobalFloat64(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Float64("myflag", float64(17), "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.GlobalFloat64("myflag"), float64(17))
	expect(t, c.GlobalFloat64("nope"), float64(0))
}

func TestContext_Duration(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Duration("myflag", 12*time.Second, "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.Duration("myflag"), 12*time.Second)
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

func TestContext_BoolT(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", true, "doc")
	c := NewContext(nil, set, nil)
	expect(t, c.BoolT("myflag"), true)
}

func TestContext_GlobalBool(t *testing.T) {
	set := flag.NewFlagSet("test", 0)

	globalSet := flag.NewFlagSet("test-global", 0)
	globalSet.Bool("myflag", false, "doc")
	globalCtx := NewContext(nil, globalSet, nil)

	c := NewContext(nil, set, globalCtx)
	expect(t, c.GlobalBool("myflag"), false)
	expect(t, c.GlobalBool("nope"), false)
}

func TestContext_GlobalBoolT(t *testing.T) {
	set := flag.NewFlagSet("test", 0)

	globalSet := flag.NewFlagSet("test-global", 0)
	globalSet.Bool("myflag", true, "doc")
	globalCtx := NewContext(nil, globalSet, nil)

	c := NewContext(nil, set, globalCtx)
	expect(t, c.GlobalBoolT("myflag"), true)
	expect(t, c.GlobalBoolT("nope"), false)
}

func TestContext_Args(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	c := NewContext(nil, set, nil)
	_ = set.Parse([]string{"--myflag", "bat", "baz"})
	expect(t, len(c.Args()), 2)
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
	set.Bool("myflag", false, "doc")
	set.String("otherflag", "hello world", "doc")
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("myflagGlobal", true, "doc")
	globalCtx := NewContext(nil, globalSet, nil)
	c := NewContext(nil, set, globalCtx)
	_ = set.Parse([]string{"--myflag", "bat", "baz"})
	_ = globalSet.Parse([]string{"--myflagGlobal", "bat", "baz"})
	expect(t, c.IsSet("myflag"), true)
	expect(t, c.IsSet("otherflag"), false)
	expect(t, c.IsSet("bogusflag"), false)
	expect(t, c.IsSet("myflagGlobal"), false)
}

func TestContext_IsSet_ShortAndFull_FlagNames(t *testing.T) {
	var (
		numberIsSet, nIsSet bool
		tempIsSet, tIsSet bool
		usernameIsSet, uIsSet bool
		debugIsSet, dIsSet bool
	)

	a := App {
		Flags: []Flag{
			IntFlag{Name: "number, n"},
			Float64Flag{Name: "temp, t"},
			StringFlag{Name: "username, u"},
			BoolFlag{Name: "debug, d"},
		},
		Action: func(ctx *Context) error {
			numberIsSet = ctx.IsSet("number")
			nIsSet = ctx.IsSet("n")
			tempIsSet = ctx.IsSet("temp")
			tIsSet = ctx.IsSet("t")
			usernameIsSet = ctx.IsSet("username")
			uIsSet = ctx.IsSet("u")
			debugIsSet = ctx.IsSet("debug")
			dIsSet = ctx.IsSet("d")
			return nil
		},
	}

	tests := []struct {
		args[]string
	}{
		{args: []string{"", "--number", "5", "--temp", "5.2", "--username", "ajitem", "--debug"}},
		{args: []string{"", "-n", "5", "-t", "5.2", "-u", "ajitem", "-d"}},
	}

	for _, tt := range tests {
		_ = a.Run(tt.args)

		expect(t, numberIsSet == nIsSet, true)
		expect(t, tempIsSet == tIsSet, true)
		expect(t, usernameIsSet == uIsSet, true)
		expect(t, debugIsSet == dIsSet, true)
	}
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

	clearenv()
	_ = os.Setenv("APP_TIMEOUT_SECONDS", "15.5")
	_ = os.Setenv("APP_PASSWORD", "")
	a := App{
		Flags: []Flag{
			Float64Flag{Name: "timeout, t", EnvVar: "APP_TIMEOUT_SECONDS"},
			StringFlag{Name: "password, p", EnvVar: "APP_PASSWORD"},
			Float64Flag{Name: "unparsable, u", EnvVar: "APP_UNPARSABLE"},
			Float64Flag{Name: "no-env-var, n"},
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

func TestContext_GlobalIsSet(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	set.String("otherflag", "hello world", "doc")
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("myflagGlobal", true, "doc")
	globalSet.Bool("myflagGlobalUnset", true, "doc")
	globalCtx := NewContext(nil, globalSet, nil)
	c := NewContext(nil, set, globalCtx)
	_ = set.Parse([]string{"--myflag", "bat", "baz"})
	_ = globalSet.Parse([]string{"--myflagGlobal", "bat", "baz"})
	expect(t, c.GlobalIsSet("myflag"), false)
	expect(t, c.GlobalIsSet("otherflag"), false)
	expect(t, c.GlobalIsSet("bogusflag"), false)
	expect(t, c.GlobalIsSet("myflagGlobal"), true)
	expect(t, c.GlobalIsSet("myflagGlobalUnset"), false)
	expect(t, c.GlobalIsSet("bogusGlobal"), false)
}

// XXX Corresponds to hack in context.IsSet for flags with EnvVar field
// Should be moved to `flag_test` in v2
func TestContext_GlobalIsSet_fromEnv(t *testing.T) {
	var (
		timeoutIsSet, tIsSet    bool
		noEnvVarIsSet, nIsSet   bool
		passwordIsSet, pIsSet   bool
		passwordValue           string
		unparsableIsSet, uIsSet bool
		overrideIsSet, oIsSet   bool
		overrideValue           string
	)

	clearenv()
	_ = os.Setenv("APP_TIMEOUT_SECONDS", "15.5")
	_ = os.Setenv("APP_PASSWORD", "badpass")
	_ = os.Setenv("APP_OVERRIDE", "overridden")
	a := App{
		Flags: []Flag{
			Float64Flag{Name: "timeout, t", EnvVar: "APP_TIMEOUT_SECONDS"},
			StringFlag{Name: "password, p", EnvVar: "APP_PASSWORD"},
			Float64Flag{Name: "no-env-var, n"},
			Float64Flag{Name: "unparsable, u", EnvVar: "APP_UNPARSABLE"},
			StringFlag{Name: "overrides-default, o", Value: "default", EnvVar: "APP_OVERRIDE"},
		},
		Commands: []Command{
			{
				Name: "hello",
				Action: func(ctx *Context) error {
					timeoutIsSet = ctx.GlobalIsSet("timeout")
					tIsSet = ctx.GlobalIsSet("t")
					passwordIsSet = ctx.GlobalIsSet("password")
					pIsSet = ctx.GlobalIsSet("p")
					passwordValue = ctx.GlobalString("password")
					unparsableIsSet = ctx.GlobalIsSet("unparsable")
					uIsSet = ctx.GlobalIsSet("u")
					noEnvVarIsSet = ctx.GlobalIsSet("no-env-var")
					nIsSet = ctx.GlobalIsSet("n")
					overrideIsSet = ctx.GlobalIsSet("overrides-default")
					oIsSet = ctx.GlobalIsSet("o")
					overrideValue = ctx.GlobalString("overrides-default")
					return nil
				},
			},
		},
	}
	if err := a.Run([]string{"run", "hello"}); err != nil {
		t.Logf("error running Run(): %+v", err)
	}
	expect(t, timeoutIsSet, true)
	expect(t, tIsSet, true)
	expect(t, passwordIsSet, true)
	expect(t, pIsSet, true)
	expect(t, passwordValue, "badpass")
	expect(t, unparsableIsSet, false)
	expect(t, noEnvVarIsSet, false)
	expect(t, nIsSet, false)
	expect(t, overrideIsSet, true)
	expect(t, oIsSet, true)
	expect(t, overrideValue, "overridden")

	_ = os.Setenv("APP_UNPARSABLE", "foobar")
	if err := a.Run([]string{"run"}); err != nil {
		t.Logf("error running Run(): %+v", err)
	}
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

func TestContext_GlobalFlag(t *testing.T) {
	var globalFlag string
	var globalFlagSet bool
	app := NewApp()
	app.Flags = []Flag{
		StringFlag{Name: "global, g", Usage: "global"},
	}
	app.Action = func(c *Context) error {
		globalFlag = c.GlobalString("global")
		globalFlagSet = c.GlobalIsSet("global")
		return nil
	}
	_ = app.Run([]string{"command", "-g", "foo"})
	expect(t, globalFlag, "foo")
	expect(t, globalFlagSet, true)

}

func TestContext_GlobalFlagsInSubcommands(t *testing.T) {
	subcommandRun := false
	parentFlag := false
	app := NewApp()

	app.Flags = []Flag{
		BoolFlag{Name: "debug, d", Usage: "Enable debugging"},
	}

	app.Commands = []Command{
		{
			Name: "foo",
			Flags: []Flag{
				BoolFlag{Name: "parent, p", Usage: "Parent flag"},
			},
			Subcommands: []Command{
				{
					Name: "bar",
					Action: func(c *Context) error {
						if c.GlobalBool("debug") {
							subcommandRun = true
						}
						if c.GlobalBool("parent") {
							parentFlag = true
						}
						return nil
					},
				},
			},
		},
	}

	_ = app.Run([]string{"command", "-d", "foo", "-p", "bar"})

	expect(t, subcommandRun, true)
	expect(t, parentFlag, true)
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

func TestContext_GlobalSet(t *testing.T) {
	gSet := flag.NewFlagSet("test", 0)
	gSet.Int("int", 5, "an int")

	set := flag.NewFlagSet("sub", 0)
	set.Int("int", 3, "an int")

	pc := NewContext(nil, gSet, nil)
	c := NewContext(nil, set, pc)

	_ = c.Set("int", "1")
	expect(t, c.Int("int"), 1)
	expect(t, c.GlobalInt("int"), 5)

	expect(t, c.GlobalIsSet("int"), false)
	_ = c.GlobalSet("int", "1")
	expect(t, c.Int("int"), 1)
	expect(t, c.GlobalInt("int"), 1)
	expect(t, c.GlobalIsSet("int"), true)
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
				StringFlag{Name: "optionalFlag"},
			},
		},
		{
			testCase: "required",
			flags: []Flag{
				StringFlag{Name: "requiredFlag", Required: true},
			},
			expectedAnError:       true,
			expectedErrorContents: []string{"requiredFlag"},
		},
		{
			testCase: "required_and_present",
			flags: []Flag{
				StringFlag{Name: "requiredFlag", Required: true},
			},
			parseInput: []string{"--requiredFlag", "myinput"},
		},
		{
			testCase: "required_and_present_via_env_var",
			flags: []Flag{
				StringFlag{Name: "requiredFlag", Required: true, EnvVar: "REQUIRED_FLAG"},
			},
			envVarInput: [2]string{"REQUIRED_FLAG", "true"},
		},
		{
			testCase: "required_and_optional",
			flags: []Flag{
				StringFlag{Name: "requiredFlag", Required: true},
				StringFlag{Name: "optionalFlag"},
			},
			expectedAnError: true,
		},
		{
			testCase: "required_and_optional_and_optional_present",
			flags: []Flag{
				StringFlag{Name: "requiredFlag", Required: true},
				StringFlag{Name: "optionalFlag"},
			},
			parseInput:      []string{"--optionalFlag", "myinput"},
			expectedAnError: true,
		},
		{
			testCase: "required_and_optional_and_optional_present_via_env_var",
			flags: []Flag{
				StringFlag{Name: "requiredFlag", Required: true},
				StringFlag{Name: "optionalFlag", EnvVar: "OPTIONAL_FLAG"},
			},
			envVarInput:     [2]string{"OPTIONAL_FLAG", "true"},
			expectedAnError: true,
		},
		{
			testCase: "required_and_optional_and_required_present",
			flags: []Flag{
				StringFlag{Name: "requiredFlag", Required: true},
				StringFlag{Name: "optionalFlag"},
			},
			parseInput: []string{"--requiredFlag", "myinput"},
		},
		{
			testCase: "two_required",
			flags: []Flag{
				StringFlag{Name: "requiredFlagOne", Required: true},
				StringFlag{Name: "requiredFlagTwo", Required: true},
			},
			expectedAnError:       true,
			expectedErrorContents: []string{"requiredFlagOne", "requiredFlagTwo"},
		},
		{
			testCase: "two_required_and_one_present",
			flags: []Flag{
				StringFlag{Name: "requiredFlag", Required: true},
				StringFlag{Name: "requiredFlagTwo", Required: true},
			},
			parseInput:      []string{"--requiredFlag", "myinput"},
			expectedAnError: true,
		},
		{
			testCase: "two_required_and_both_present",
			flags: []Flag{
				StringFlag{Name: "requiredFlag", Required: true},
				StringFlag{Name: "requiredFlagTwo", Required: true},
			},
			parseInput: []string{"--requiredFlag", "myinput", "--requiredFlagTwo", "myinput"},
		},
		{
			testCase: "required_flag_with_short_name",
			flags: []Flag{
				StringSliceFlag{Name: "names, N", Required: true},
			},
			parseInput: []string{"-N", "asd", "-N", "qwe"},
		},
		{
			testCase: "required_flag_with_multiple_short_names",
			flags: []Flag{
				StringSliceFlag{Name: "names, N, n", Required: true},
			},
			parseInput: []string{"-n", "asd", "-n", "qwe"},
		},
		{
			testCase: "required_flag_with_short_alias_not_printed_on_error",
			expectedAnError: true,
			expectedErrorContents: []string{"Required flag \"names\" not set"},
			flags: []Flag{
				StringSliceFlag{Name: "names, n", Required: true},
			},
		},
	}
	for _, test := range tdata {
		t.Run(test.testCase, func(t *testing.T) {
			// setup
			set := flag.NewFlagSet("test", 0)
			for _, flags := range test.flags {
				flags.Apply(set)
			}
			_ = set.Parse(test.parseInput)
			if test.envVarInput[0] != "" {
				os.Clearenv()
				_ = os.Setenv(test.envVarInput[0], test.envVarInput[1])
			}
			ctx := &Context{}
			context := NewContext(ctx.App, set, ctx)
			context.Command.Flags = test.flags

			// logic under test
			err := checkRequiredFlags(test.flags, context)

			// assertions
			if test.expectedAnError && err == nil {
				t.Errorf("expected an error, but there was none")
			}
			if !test.expectedAnError && err != nil {
				t.Errorf("did not expected an error, but there was one: %s", err)
			}
			for _, errString := range test.expectedErrorContents {
				if !strings.Contains(err.Error(), errString) {
					t.Errorf("expected error %q to contain %q, but it didn't!", err.Error(), errString)
				}
			}
		})
	}
}
