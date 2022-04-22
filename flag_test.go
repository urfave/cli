package cli

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

var boolFlagTests = []struct {
	name     string
	expected string
}{
	{"help", "--help\t(default: false)"},
	{"h", "-h\t(default: false)"},
}

func resetEnv(env []string) {
	for _, e := range env {
		fields := strings.SplitN(e, "=", 2)
		os.Setenv(fields[0], fields[1])
	}
}

func TestBoolFlagHelpOutput(t *testing.T) {
	for _, test := range boolFlagTests {
		fl := &BoolFlag{Name: test.name}
		output := fl.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestBoolFlagApply_SetsAllNames(t *testing.T) {
	v := false
	fl := BoolFlag{Name: "wat", Aliases: []string{"W", "huh"}, Destination: &v}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse([]string{"--wat", "-W", "--huh"})
	expect(t, err, nil)
	expect(t, v, true)
}

func TestFlagsFromEnv(t *testing.T) {
	newSetFloat64Slice := func(defaults ...float64) Float64Slice {
		s := NewFloat64Slice(defaults...)
		s.hasBeenSet = false
		return *s
	}

	newSetIntSlice := func(defaults ...int) IntSlice {
		s := NewIntSlice(defaults...)
		s.hasBeenSet = false
		return *s
	}

	newSetInt64Slice := func(defaults ...int64) Int64Slice {
		s := NewInt64Slice(defaults...)
		s.hasBeenSet = false
		return *s
	}

	newSetStringSlice := func(defaults ...string) StringSlice {
		s := NewStringSlice(defaults...)
		s.hasBeenSet = false
		return *s
	}

	var flagTests = []struct {
		input     string
		output    interface{}
		flag      Flag
		errRegexp string
	}{
		{"", false, &BoolFlag{Name: "debug", EnvVars: []string{"DEBUG"}}, ""},
		{"1", true, &BoolFlag{Name: "debug", EnvVars: []string{"DEBUG"}}, ""},
		{"false", false, &BoolFlag{Name: "debug", EnvVars: []string{"DEBUG"}}, ""},
		{"foobar", true, &BoolFlag{Name: "debug", EnvVars: []string{"DEBUG"}}, `could not parse "foobar" as bool value for flag debug: .*`},

		{"1s", 1 * time.Second, &DurationFlag{Name: "time", EnvVars: []string{"TIME"}}, ""},
		{"foobar", false, &DurationFlag{Name: "time", EnvVars: []string{"TIME"}}, `could not parse "foobar" as duration value for flag time: .*`},

		{"1.2", 1.2, &Float64Flag{Name: "seconds", EnvVars: []string{"SECONDS"}}, ""},
		{"1", 1.0, &Float64Flag{Name: "seconds", EnvVars: []string{"SECONDS"}}, ""},
		{"foobar", 0, &Float64Flag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "foobar" as float64 value for flag seconds: .*`},

		{"1", int64(1), &Int64Flag{Name: "seconds", EnvVars: []string{"SECONDS"}}, ""},
		{"1.2", 0, &Int64Flag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "1.2" as int value for flag seconds: .*`},
		{"foobar", 0, &Int64Flag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "foobar" as int value for flag seconds: .*`},

		{"1", 1, &IntFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, ""},
		{"1.2", 0, &IntFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "1.2" as int value for flag seconds: .*`},
		{"foobar", 0, &IntFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "foobar" as int value for flag seconds: .*`},

		{"1.0,2", newSetFloat64Slice(1, 2), &Float64SliceFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, ""},
		{"foobar", newSetFloat64Slice(), &Float64SliceFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "\[\]float64{}" as float64 slice value for flag seconds: .*`},

		{"1,2", newSetIntSlice(1, 2), &IntSliceFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, ""},
		{"1.2,2", newSetIntSlice(), &IntSliceFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "1.2,2" as int slice value for flag seconds: .*`},
		{"foobar", newSetIntSlice(), &IntSliceFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "foobar" as int slice value for flag seconds: .*`},

		{"1,2", newSetInt64Slice(1, 2), &Int64SliceFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, ""},
		{"1.2,2", newSetInt64Slice(), &Int64SliceFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "1.2,2" as int64 slice value for flag seconds: .*`},
		{"foobar", newSetInt64Slice(), &Int64SliceFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "foobar" as int64 slice value for flag seconds: .*`},

		{"foo", "foo", &StringFlag{Name: "name", EnvVars: []string{"NAME"}}, ""},
		{"path", "path", &PathFlag{Name: "path", EnvVars: []string{"PATH"}}, ""},

		{"foo,bar", newSetStringSlice("foo", "bar"), &StringSliceFlag{Name: "names", EnvVars: []string{"NAMES"}}, ""},

		{"1", uint(1), &UintFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, ""},
		{"1.2", 0, &UintFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "1.2" as uint value for flag seconds: .*`},
		{"foobar", 0, &UintFlag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "foobar" as uint value for flag seconds: .*`},

		{"1", uint64(1), &Uint64Flag{Name: "seconds", EnvVars: []string{"SECONDS"}}, ""},
		{"1.2", 0, &Uint64Flag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "1.2" as uint64 value for flag seconds: .*`},
		{"foobar", 0, &Uint64Flag{Name: "seconds", EnvVars: []string{"SECONDS"}}, `could not parse "foobar" as uint64 value for flag seconds: .*`},

		{"foo,bar", &Parser{"foo", "bar"}, &GenericFlag{Name: "names", Value: &Parser{}, EnvVars: []string{"NAMES"}}, ""},
	}

	for i, test := range flagTests {
		defer resetEnv(os.Environ())
		os.Clearenv()

		f, ok := test.flag.(DocGenerationFlag)
		if !ok {
			t.Errorf("flag %v needs to implement DocGenerationFlag to retrieve env vars", test.flag)
		}
		envVarSlice := f.GetEnvVars()
		_ = os.Setenv(envVarSlice[0], test.input)

		a := App{
			Flags: []Flag{test.flag},
			Action: func(ctx *Context) error {
				if !reflect.DeepEqual(ctx.Value(test.flag.Names()[0]), test.output) {
					t.Errorf("ex:%01d expected %q to be parsed as %#v, instead was %#v", i, test.input, test.output, ctx.Value(test.flag.Names()[0]))
				}
				return nil
			},
		}

		err := a.Run([]string{"run"})

		if test.errRegexp != "" {
			if err == nil {
				t.Errorf("expected error to match %q, got none", test.errRegexp)
			} else {
				if matched, _ := regexp.MatchString(test.errRegexp, err.Error()); !matched {
					t.Errorf("expected error to match %q, got error %s", test.errRegexp, err)
				}
			}
		} else {
			if err != nil && test.errRegexp == "" {
				t.Errorf("expected no error got %q", err)
			}
		}
	}
}

type nodocFlag struct {
	Flag

	Name string
}

func TestFlagStringifying(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fl       Flag
		expected string
	}{
		{
			name:     "bool-flag",
			fl:       &BoolFlag{Name: "vividly"},
			expected: "--vividly\t(default: false)",
		},
		{
			name:     "bool-flag-with-default-text",
			fl:       &BoolFlag{Name: "wildly", DefaultText: "scrambled"},
			expected: "--wildly\t(default: scrambled)",
		},
		{
			name:     "duration-flag",
			fl:       &DurationFlag{Name: "scream-for"},
			expected: "--scream-for value\t(default: 0s)",
		},
		{
			name:     "duration-flag-with-default-text",
			fl:       &DurationFlag{Name: "feels-about", DefaultText: "whimsically"},
			expected: "--feels-about value\t(default: whimsically)",
		},
		{
			name:     "float64-flag",
			fl:       &Float64Flag{Name: "arduous"},
			expected: "--arduous value\t(default: 0)",
		},
		{
			name:     "float64-flag-with-default-text",
			fl:       &Float64Flag{Name: "filibuster", DefaultText: "42"},
			expected: "--filibuster value\t(default: 42)",
		},
		{
			name:     "float64-slice-flag",
			fl:       &Float64SliceFlag{Name: "pizzas"},
			expected: "--pizzas value\t",
		},
		{
			name:     "float64-slice-flag-with-default-text",
			fl:       &Float64SliceFlag{Name: "pepperonis", DefaultText: "shaved"},
			expected: "--pepperonis value\t(default: shaved)",
		},
		{
			name:     "generic-flag",
			fl:       &GenericFlag{Name: "yogurt"},
			expected: "--yogurt value\t",
		},
		{
			name:     "generic-flag-with-default-text",
			fl:       &GenericFlag{Name: "ricotta", DefaultText: "plops"},
			expected: "--ricotta value\t(default: plops)",
		},
		{
			name:     "int-flag",
			fl:       &IntFlag{Name: "grubs"},
			expected: "--grubs value\t(default: 0)",
		},
		{
			name:     "int-flag-with-default-text",
			fl:       &IntFlag{Name: "poisons", DefaultText: "11ty"},
			expected: "--poisons value\t(default: 11ty)",
		},
		{
			name:     "int-slice-flag",
			fl:       &IntSliceFlag{Name: "pencils"},
			expected: "--pencils value\t",
		},
		{
			name:     "int-slice-flag-with-default-text",
			fl:       &IntFlag{Name: "pens", DefaultText: "-19"},
			expected: "--pens value\t(default: -19)",
		},
		{
			name:     "int64-flag",
			fl:       &Int64Flag{Name: "flume"},
			expected: "--flume value\t(default: 0)",
		},
		{
			name:     "int64-flag-with-default-text",
			fl:       &Int64Flag{Name: "shattering", DefaultText: "22"},
			expected: "--shattering value\t(default: 22)",
		},
		{
			name:     "int64-slice-flag",
			fl:       &Int64SliceFlag{Name: "drawers"},
			expected: "--drawers value\t",
		},
		{
			name:     "int64-slice-flag-with-default-text",
			fl:       &Int64SliceFlag{Name: "handles", DefaultText: "-2"},
			expected: "--handles value\t(default: -2)",
		},
		{
			name:     "path-flag",
			fl:       &PathFlag{Name: "soup"},
			expected: "--soup value\t",
		},
		{
			name:     "path-flag-with-default-text",
			fl:       &PathFlag{Name: "stew", DefaultText: "charred/beans"},
			expected: "--stew value\t(default: charred/beans)",
		},
		{
			name:     "string-flag",
			fl:       &StringFlag{Name: "arf-sound"},
			expected: "--arf-sound value\t",
		},
		{
			name:     "string-flag-with-default-text",
			fl:       &StringFlag{Name: "woof-sound", DefaultText: "urp"},
			expected: "--woof-sound value\t(default: urp)",
		},
		{
			name:     "string-slice-flag",
			fl:       &StringSliceFlag{Name: "meow-sounds"},
			expected: "--meow-sounds value\t",
		},
		{
			name:     "string-slice-flag-with-default-text",
			fl:       &StringSliceFlag{Name: "moo-sounds", DefaultText: "awoo"},
			expected: "--moo-sounds value\t(default: awoo)",
		},
		{
			name:     "timestamp-flag",
			fl:       &TimestampFlag{Name: "eating"},
			expected: "--eating value\t",
		},
		{
			name:     "timestamp-flag-with-default-text",
			fl:       &TimestampFlag{Name: "sleeping", DefaultText: "earlier"},
			expected: "--sleeping value\t(default: earlier)",
		},
		{
			name:     "uint-flag",
			fl:       &UintFlag{Name: "jars"},
			expected: "--jars value\t(default: 0)",
		},
		{
			name:     "uint-flag-with-default-text",
			fl:       &UintFlag{Name: "bottles", DefaultText: "99"},
			expected: "--bottles value\t(default: 99)",
		},
		{
			name:     "uint64-flag",
			fl:       &Uint64Flag{Name: "cans"},
			expected: "--cans value\t(default: 0)",
		},
		{
			name:     "uint64-flag-with-default-text",
			fl:       &UintFlag{Name: "tubes", DefaultText: "13"},
			expected: "--tubes value\t(default: 13)",
		},
		{
			name:     "nodoc-flag",
			fl:       &nodocFlag{Name: "scarecrow"},
			expected: "",
		},
	} {
		t.Run(tc.name, func(ct *testing.T) {
			s := stringifyFlag(tc.fl)
			if s != tc.expected {
				ct.Errorf("stringified flag %q does not match expected %q", s, tc.expected)
			}
		})
	}
}

var stringFlagTests = []struct {
	name     string
	aliases  []string
	usage    string
	value    string
	expected string
}{
	{"foo", nil, "", "", "--foo value\t"},
	{"f", nil, "", "", "-f value\t"},
	{"f", nil, "The total `foo` desired", "all", "-f foo\tThe total foo desired (default: \"all\")"},
	{"test", nil, "", "Something", "--test value\t(default: \"Something\")"},
	{"config", []string{"c"}, "Load configuration from `FILE`", "", "--config FILE, -c FILE\tLoad configuration from FILE"},
	{"config", []string{"c"}, "Load configuration from `CONFIG`", "config.json", "--config CONFIG, -c CONFIG\tLoad configuration from CONFIG (default: \"config.json\")"},
}

func TestStringFlagHelpOutput(t *testing.T) {
	for _, test := range stringFlagTests {
		fl := &StringFlag{Name: test.name, Aliases: test.aliases, Usage: test.usage, Value: test.value}
		output := fl.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestStringFlagDefaultText(t *testing.T) {
	fl := &StringFlag{Name: "foo", Aliases: nil, Usage: "amount of `foo` requested", Value: "none", DefaultText: "all of it"}
	expected := "--foo foo\tamount of foo requested (default: all of it)"
	output := fl.String()

	if output != expected {
		t.Errorf("%q does not match %q", output, expected)
	}
}

func TestStringFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_FOO", "derp")

	for _, test := range stringFlagTests {
		fl := &StringFlag{Name: test.name, Aliases: test.aliases, Value: test.value, EnvVars: []string{"APP_FOO"}}
		output := fl.String()

		expectedSuffix := " [$APP_FOO]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_FOO%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

var prefixStringFlagTests = []struct {
	name     string
	aliases  []string
	usage    string
	value    string
	prefixer FlagNamePrefixFunc
	expected string
}{
	{name: "foo", usage: "", value: "", prefixer: func(a []string, b string) string {
		return fmt.Sprintf("name: %s, ph: %s", a, b)
	}, expected: "name: foo, ph: value\t"},
	{name: "f", usage: "", value: "", prefixer: func(a []string, b string) string {
		return fmt.Sprintf("name: %s, ph: %s", a, b)
	}, expected: "name: f, ph: value\t"},
	{name: "f", usage: "The total `foo` desired", value: "all", prefixer: func(a []string, b string) string {
		return fmt.Sprintf("name: %s, ph: %s", a, b)
	}, expected: "name: f, ph: foo\tThe total foo desired (default: \"all\")"},
	{name: "test", usage: "", value: "Something", prefixer: func(a []string, b string) string {
		return fmt.Sprintf("name: %s, ph: %s", a, b)
	}, expected: "name: test, ph: value\t(default: \"Something\")"},
	{name: "config", aliases: []string{"c"}, usage: "Load configuration from `FILE`", value: "", prefixer: func(a []string, b string) string {
		return fmt.Sprintf("name: %s, ph: %s", a, b)
	}, expected: "name: config,c, ph: FILE\tLoad configuration from FILE"},
	{name: "config", aliases: []string{"c"}, usage: "Load configuration from `CONFIG`", value: "config.json", prefixer: func(a []string, b string) string {
		return fmt.Sprintf("name: %s, ph: %s", a, b)
	}, expected: "name: config,c, ph: CONFIG\tLoad configuration from CONFIG (default: \"config.json\")"},
}

func TestStringFlagApply_SetsAllNames(t *testing.T) {
	v := "mmm"
	fl := StringFlag{Name: "hay", Aliases: []string{"H", "hayyy"}, Destination: &v}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse([]string{"--hay", "u", "-H", "yuu", "--hayyy", "YUUUU"})
	expect(t, err, nil)
	expect(t, v, "YUUUU")
}

var pathFlagTests = []struct {
	name     string
	aliases  []string
	usage    string
	value    string
	expected string
}{
	{"f", nil, "", "", "-f value\t"},
	{"f", nil, "Path is the `path` of file", "/path/to/file", "-f path\tPath is the path of file (default: \"/path/to/file\")"},
}

func TestPathFlagHelpOutput(t *testing.T) {
	for _, test := range pathFlagTests {
		fl := &PathFlag{Name: test.name, Aliases: test.aliases, Usage: test.usage, Value: test.value}
		output := fl.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestPathFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_PATH", "/path/to/file")
	for _, test := range pathFlagTests {
		fl := &PathFlag{Name: test.name, Aliases: test.aliases, Value: test.value, EnvVars: []string{"APP_PATH"}}
		output := fl.String()

		expectedSuffix := " [$APP_PATH]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_PATH%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

func TestPathFlagApply_SetsAllNames(t *testing.T) {
	v := "mmm"
	fl := PathFlag{Name: "path", Aliases: []string{"p", "PATH"}, Destination: &v}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse([]string{"--path", "/path/to/file/path", "-p", "/path/to/file/p", "--PATH", "/path/to/file/PATH"})
	expect(t, err, nil)
	expect(t, v, "/path/to/file/PATH")
}

var envHintFlagTests = []struct {
	name     string
	env      string
	hinter   FlagEnvHintFunc
	expected string
}{
	{"foo", "", func(a []string, b string) string {
		return fmt.Sprintf("env: %s, str: %s", a, b)
	}, "env: , str: --foo value\t"},
	{"f", "", func(a []string, b string) string {
		return fmt.Sprintf("env: %s, str: %s", a, b)
	}, "env: , str: -f value\t"},
	{"foo", "ENV_VAR", func(a []string, b string) string {
		return fmt.Sprintf("env: %s, str: %s", a, b)
	}, "env: ENV_VAR, str: --foo value\t"},
	{"f", "ENV_VAR", func(a []string, b string) string {
		return fmt.Sprintf("env: %s, str: %s", a, b)
	}, "env: ENV_VAR, str: -f value\t"},
}

//func TestFlagEnvHinter(t *testing.T) {
//	defer func() {
//		FlagEnvHinter = withEnvHint
//	}()
//
//	for _, test := range envHintFlagTests {
//		FlagEnvHinter = test.hinter
//		fl := StringFlag{Name: test.name, EnvVars: []string{test.env}}
//		output := fl.String()
//		if output != test.expected {
//			t.Errorf("%q does not match %q", output, test.expected)
//		}
//	}
//}

var stringSliceFlagTests = []struct {
	name     string
	aliases  []string
	value    *StringSlice
	expected string
}{
	{"foo", nil, NewStringSlice(""), "--foo value\t(accepts multiple inputs)"},
	{"f", nil, NewStringSlice(""), "-f value\t(accepts multiple inputs)"},
	{"f", nil, NewStringSlice("Lipstick"), "-f value\t(default: \"Lipstick\")\t(accepts multiple inputs)"},
	{"test", nil, NewStringSlice("Something"), "--test value\t(default: \"Something\")\t(accepts multiple inputs)"},
	{"dee", []string{"d"}, NewStringSlice("Inka", "Dinka", "dooo"), "--dee value, -d value\t(default: \"Inka\", \"Dinka\", \"dooo\")\t(accepts multiple inputs)"},
}

func TestStringSliceFlagHelpOutput(t *testing.T) {
	for _, test := range stringSliceFlagTests {
		f := &StringSliceFlag{Name: test.name, Aliases: test.aliases, Value: test.value}
		output := f.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestStringSliceFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_QWWX", "11,4")

	for _, test := range stringSliceFlagTests {
		fl := &StringSliceFlag{Name: test.name, Aliases: test.aliases, Value: test.value, EnvVars: []string{"APP_QWWX"}}
		output := fl.String()

		expectedSuffix := " [$APP_QWWX]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_QWWX%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

func TestStringSliceFlagApply_SetsAllNames(t *testing.T) {
	fl := StringSliceFlag{Name: "goat", Aliases: []string{"G", "gooots"}}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse([]string{"--goat", "aaa", "-G", "bbb", "--gooots", "eeeee"})
	expect(t, err, nil)
}

func TestStringSliceFlagApply_UsesEnvValues(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("MY_GOAT", "vincent van goat,scape goat")
	var val StringSlice
	fl := StringSliceFlag{Name: "goat", EnvVars: []string{"MY_GOAT"}, Value: &val}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse(nil)
	expect(t, err, nil)
	expect(t, val.Value(), NewStringSlice("vincent van goat", "scape goat").Value())
}

func TestStringSliceFlagApply_DefaultValueWithDestination(t *testing.T) {
	defValue := []string{"UA", "US"}

	fl := StringSliceFlag{Name: "country", Value: NewStringSlice(defValue...), Destination: NewStringSlice("CA")}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse([]string{})
	expect(t, err, nil)
	expect(t, defValue, fl.Destination.Value())
}

var intFlagTests = []struct {
	name     string
	expected string
}{
	{"hats", "--hats value\t(default: 9)"},
	{"H", "-H value\t(default: 9)"},
}

func TestIntFlagHelpOutput(t *testing.T) {
	for _, test := range intFlagTests {
		fl := &IntFlag{Name: test.name, Value: 9}
		output := fl.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

func TestIntFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_BAR", "2")

	for _, test := range intFlagTests {
		fl := &IntFlag{Name: test.name, EnvVars: []string{"APP_BAR"}}
		output := fl.String()

		expectedSuffix := " [$APP_BAR]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAR%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

func TestIntFlagApply_SetsAllNames(t *testing.T) {
	v := 3
	fl := IntFlag{Name: "banana", Aliases: []string{"B", "banannanana"}, Destination: &v}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse([]string{"--banana", "1", "-B", "2", "--banannanana", "5"})
	expect(t, err, nil)
	expect(t, v, 5)
}

var int64FlagTests = []struct {
	name     string
	expected string
}{
	{"hats", "--hats value\t(default: 8589934592)"},
	{"H", "-H value\t(default: 8589934592)"},
}

func TestInt64FlagHelpOutput(t *testing.T) {
	for _, test := range int64FlagTests {
		fl := Int64Flag{Name: test.name, Value: 8589934592}
		output := fl.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

func TestInt64FlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_BAR", "2")

	for _, test := range int64FlagTests {
		fl := IntFlag{Name: test.name, EnvVars: []string{"APP_BAR"}}
		output := fl.String()

		expectedSuffix := " [$APP_BAR]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAR%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

var uintFlagTests = []struct {
	name     string
	expected string
}{
	{"nerfs", "--nerfs value\t(default: 41)"},
	{"N", "-N value\t(default: 41)"},
}

func TestUintFlagHelpOutput(t *testing.T) {
	for _, test := range uintFlagTests {
		fl := UintFlag{Name: test.name, Value: 41}
		output := fl.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

func TestUintFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_BAR", "2")

	for _, test := range uintFlagTests {
		fl := UintFlag{Name: test.name, EnvVars: []string{"APP_BAR"}}
		output := fl.String()

		expectedSuffix := " [$APP_BAR]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAR%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

var uint64FlagTests = []struct {
	name     string
	expected string
}{
	{"gerfs", "--gerfs value\t(default: 8589934582)"},
	{"G", "-G value\t(default: 8589934582)"},
}

func TestUint64FlagHelpOutput(t *testing.T) {
	for _, test := range uint64FlagTests {
		fl := Uint64Flag{Name: test.name, Value: 8589934582}
		output := fl.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

func TestUint64FlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_BAR", "2")

	for _, test := range uint64FlagTests {
		fl := UintFlag{Name: test.name, EnvVars: []string{"APP_BAR"}}
		output := fl.String()

		expectedSuffix := " [$APP_BAR]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAR%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

var durationFlagTests = []struct {
	name     string
	expected string
}{
	{"hooting", "--hooting value\t(default: 1s)"},
	{"H", "-H value\t(default: 1s)"},
}

func TestDurationFlagHelpOutput(t *testing.T) {
	for _, test := range durationFlagTests {
		fl := &DurationFlag{Name: test.name, Value: 1 * time.Second}
		output := fl.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestDurationFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_BAR", "2h3m6s")

	for _, test := range durationFlagTests {
		fl := &DurationFlag{Name: test.name, EnvVars: []string{"APP_BAR"}}
		output := fl.String()

		expectedSuffix := " [$APP_BAR]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAR%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

func TestDurationFlagApply_SetsAllNames(t *testing.T) {
	v := time.Second * 20
	fl := DurationFlag{Name: "howmuch", Aliases: []string{"H", "whyyy"}, Destination: &v}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse([]string{"--howmuch", "30s", "-H", "5m", "--whyyy", "30h"})
	expect(t, err, nil)
	expect(t, v, time.Hour*30)
}

var intSliceFlagTests = []struct {
	name     string
	aliases  []string
	value    *IntSlice
	expected string
}{
	{"heads", nil, NewIntSlice(), "--heads value\t(accepts multiple inputs)"},
	{"H", nil, NewIntSlice(), "-H value\t(accepts multiple inputs)"},
	{"H", []string{"heads"}, NewIntSlice(9, 3), "-H value, --heads value\t(default: 9, 3)\t(accepts multiple inputs)"},
}

func TestIntSliceFlagHelpOutput(t *testing.T) {
	for _, test := range intSliceFlagTests {
		fl := &IntSliceFlag{Name: test.name, Aliases: test.aliases, Value: test.value}
		output := fl.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestIntSliceFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_SMURF", "42,3")

	for _, test := range intSliceFlagTests {
		fl := &IntSliceFlag{Name: test.name, Aliases: test.aliases, Value: test.value, EnvVars: []string{"APP_SMURF"}}
		output := fl.String()

		expectedSuffix := " [$APP_SMURF]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_SMURF%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

func TestIntSliceFlagApply_SetsAllNames(t *testing.T) {
	fl := IntSliceFlag{Name: "bits", Aliases: []string{"B", "bips"}}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse([]string{"--bits", "23", "-B", "3", "--bips", "99"})
	expect(t, err, nil)
}

func TestIntSliceFlagApply_ParentContext(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&IntSliceFlag{Name: "numbers", Aliases: []string{"n"}, Value: NewIntSlice(1, 2, 3)},
		},
		Commands: []*Command{
			{
				Name: "child",
				Action: func(ctx *Context) error {
					expected := []int{1, 2, 3}
					if !reflect.DeepEqual(ctx.IntSlice("numbers"), expected) {
						t.Errorf("child context unable to view parent flag: %v != %v", expected, ctx.IntSlice("numbers"))
					}
					if !reflect.DeepEqual(ctx.IntSlice("n"), expected) {
						t.Errorf("child context unable to view parent flag: %v != %v", expected, ctx.IntSlice("n"))
					}
					return nil
				},
			},
		},
	}).Run([]string{"run", "child"})
}

func TestIntSliceFlag_SetFromParentContext(t *testing.T) {
	fl := &IntSliceFlag{Name: "numbers", Aliases: []string{"n"}, Value: NewIntSlice(1, 2, 3, 4)}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)
	ctx := &Context{
		parentContext: &Context{
			flagSet: set,
		},
		flagSet: flag.NewFlagSet("empty", 0),
	}
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(ctx.IntSlice("numbers"), expected) {
		t.Errorf("child context unable to view parent flag: %v != %v", expected, ctx.IntSlice("numbers"))
	}
}

var int64SliceFlagTests = []struct {
	name     string
	aliases  []string
	value    *Int64Slice
	expected string
}{
	{"heads", nil, NewInt64Slice(), "--heads value\t(accepts multiple inputs)"},
	{"H", nil, NewInt64Slice(), "-H value\t(accepts multiple inputs)"},
	{"heads", []string{"H"}, NewInt64Slice(int64(2), int64(17179869184)),
		"--heads value, -H value\t(default: 2, 17179869184)\t(accepts multiple inputs)"},
}

func TestInt64SliceFlagHelpOutput(t *testing.T) {
	for _, test := range int64SliceFlagTests {
		fl := Int64SliceFlag{Name: test.name, Aliases: test.aliases, Value: test.value}
		output := fl.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestInt64SliceFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_SMURF", "42,17179869184")

	for _, test := range int64SliceFlagTests {
		fl := Int64SliceFlag{Name: test.name, Value: test.value, EnvVars: []string{"APP_SMURF"}}
		output := fl.String()

		expectedSuffix := " [$APP_SMURF]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_SMURF%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

func TestInt64SliceFlagApply_ParentContext(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&Int64SliceFlag{Name: "numbers", Aliases: []string{"n"}, Value: NewInt64Slice(1, 2, 3)},
		},
		Commands: []*Command{
			{
				Name: "child",
				Action: func(ctx *Context) error {
					expected := []int64{1, 2, 3}
					if !reflect.DeepEqual(ctx.Int64Slice("numbers"), expected) {
						t.Errorf("child context unable to view parent flag: %v != %v", expected, ctx.Int64Slice("numbers"))
					}
					if !reflect.DeepEqual(ctx.Int64Slice("n"), expected) {
						t.Errorf("child context unable to view parent flag: %v != %v", expected, ctx.Int64Slice("n"))
					}
					return nil
				},
			},
		},
	}).Run([]string{"run", "child"})
}

func TestInt64SliceFlag_SetFromParentContext(t *testing.T) {
	fl := &Int64SliceFlag{Name: "numbers", Aliases: []string{"n"}, Value: NewInt64Slice(1, 2, 3, 4)}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)
	ctx := &Context{
		parentContext: &Context{
			flagSet: set,
		},
		flagSet: flag.NewFlagSet("empty", 0),
	}
	expected := []int64{1, 2, 3, 4}
	if !reflect.DeepEqual(ctx.Int64Slice("numbers"), expected) {
		t.Errorf("child context unable to view parent flag: %v != %v", expected, ctx.Int64Slice("numbers"))
	}
}
func TestInt64SliceFlag_ReturnNil(t *testing.T) {
	fl := &Int64SliceFlag{}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)
	ctx := &Context{
		parentContext: &Context{
			flagSet: set,
		},
		flagSet: flag.NewFlagSet("empty", 0),
	}
	expected := []int64(nil)
	if !reflect.DeepEqual(ctx.Int64Slice("numbers"), expected) {
		t.Errorf("child context unable to view parent flag: %v != %v", expected, ctx.Int64Slice("numbers"))
	}
}

var float64FlagTests = []struct {
	name     string
	expected string
}{
	{"hooting", "--hooting value\t(default: 0.1)"},
	{"H", "-H value\t(default: 0.1)"},
}

func TestFloat64FlagHelpOutput(t *testing.T) {
	for _, test := range float64FlagTests {
		f := &Float64Flag{Name: test.name, Value: 0.1}
		output := f.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestFloat64FlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_BAZ", "99.4")

	for _, test := range float64FlagTests {
		fl := &Float64Flag{Name: test.name, EnvVars: []string{"APP_BAZ"}}
		output := fl.String()

		expectedSuffix := " [$APP_BAZ]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAZ%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

func TestFloat64FlagApply_SetsAllNames(t *testing.T) {
	v := 99.1
	fl := Float64Flag{Name: "noodles", Aliases: []string{"N", "nurbles"}, Destination: &v}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse([]string{"--noodles", "1.3", "-N", "11", "--nurbles", "43.33333"})
	expect(t, err, nil)
	expect(t, v, float64(43.33333))
}

var float64SliceFlagTests = []struct {
	name     string
	aliases  []string
	value    *Float64Slice
	expected string
}{
	{"heads", nil, NewFloat64Slice(), "--heads value\t(accepts multiple inputs)"},
	{"H", nil, NewFloat64Slice(), "-H value\t(accepts multiple inputs)"},
	{"heads", []string{"H"}, NewFloat64Slice(0.1234, -10.5),
		"--heads value, -H value\t(default: 0.1234, -10.5)\t(accepts multiple inputs)"},
}

func TestFloat64SliceFlagHelpOutput(t *testing.T) {
	for _, test := range float64SliceFlagTests {
		fl := Float64SliceFlag{Name: test.name, Aliases: test.aliases, Value: test.value}
		output := fl.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestFloat64SliceFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_SMURF", "0.1234,-10.5")
	for _, test := range float64SliceFlagTests {
		fl := Float64SliceFlag{Name: test.name, Value: test.value, EnvVars: []string{"APP_SMURF"}}
		output := fl.String()

		expectedSuffix := " [$APP_SMURF]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_SMURF%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

var genericFlagTests = []struct {
	name     string
	value    Generic
	expected string
}{
	{"toads", &Parser{"abc", "def"}, "--toads value\ttest flag (default: abc,def)"},
	{"t", &Parser{"abc", "def"}, "-t value\ttest flag (default: abc,def)"},
}

func TestGenericFlagHelpOutput(t *testing.T) {
	for _, test := range genericFlagTests {
		fl := &GenericFlag{Name: test.name, Value: test.value, Usage: "test flag"}
		output := fl.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestGenericFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_ZAP", "3")

	for _, test := range genericFlagTests {
		fl := &GenericFlag{Name: test.name, EnvVars: []string{"APP_ZAP"}}
		output := fl.String()

		expectedSuffix := " [$APP_ZAP]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_ZAP%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

func TestGenericFlagApply_SetsAllNames(t *testing.T) {
	fl := GenericFlag{Name: "orbs", Aliases: []string{"O", "obrs"}, Value: &Parser{}}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse([]string{"--orbs", "eleventy,3", "-O", "4,bloop", "--obrs", "19,s"})
	expect(t, err, nil)
}

func TestParseMultiString(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&StringFlag{Name: "serve", Aliases: []string{"s"}},
		},
		Action: func(ctx *Context) error {
			if ctx.String("serve") != "10" {
				t.Errorf("main name not set")
			}
			if ctx.String("s") != "10" {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10"})
}

func TestParseDestinationString(t *testing.T) {
	var dest string
	_ = (&App{
		Flags: []Flag{
			&StringFlag{
				Name:        "dest",
				Destination: &dest,
			},
		},
		Action: func(ctx *Context) error {
			if dest != "10" {
				t.Errorf("expected destination String 10")
			}
			return nil
		},
	}).Run([]string{"run", "--dest", "10"})
}

func TestParseMultiStringFromEnv(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_COUNT", "20")
	_ = (&App{
		Flags: []Flag{
			&StringFlag{Name: "count", Aliases: []string{"c"}, EnvVars: []string{"APP_COUNT"}},
		},
		Action: func(ctx *Context) error {
			if ctx.String("count") != "20" {
				t.Errorf("main name not set")
			}
			if ctx.String("c") != "20" {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiStringFromEnvCascade(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_COUNT", "20")
	_ = (&App{
		Flags: []Flag{
			&StringFlag{Name: "count", Aliases: []string{"c"}, EnvVars: []string{"COMPAT_COUNT", "APP_COUNT"}},
		},
		Action: func(ctx *Context) error {
			if ctx.String("count") != "20" {
				t.Errorf("main name not set")
			}
			if ctx.String("c") != "20" {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiStringSlice(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&StringSliceFlag{Name: "serve", Aliases: []string{"s"}, Value: NewStringSlice()},
		},
		Action: func(ctx *Context) error {
			expected := []string{"10", "20"}
			if !reflect.DeepEqual(ctx.StringSlice("serve"), expected) {
				t.Errorf("main name not set: %v != %v", expected, ctx.StringSlice("serve"))
			}
			if !reflect.DeepEqual(ctx.StringSlice("s"), expected) {
				t.Errorf("short name not set: %v != %v", expected, ctx.StringSlice("s"))
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiStringSliceWithDefaults(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&StringSliceFlag{Name: "serve", Aliases: []string{"s"}, Value: NewStringSlice("9", "2")},
		},
		Action: func(ctx *Context) error {
			expected := []string{"10", "20"}
			if !reflect.DeepEqual(ctx.StringSlice("serve"), expected) {
				t.Errorf("main name not set: %v != %v", expected, ctx.StringSlice("serve"))
			}
			if !reflect.DeepEqual(ctx.StringSlice("s"), expected) {
				t.Errorf("short name not set: %v != %v", expected, ctx.StringSlice("s"))
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiStringSliceWithDestination(t *testing.T) {
	dest := &StringSlice{}
	_ = (&App{
		Flags: []Flag{
			&StringSliceFlag{Name: "serve", Aliases: []string{"s"}, Destination: dest},
		},
		Action: func(ctx *Context) error {
			expected := []string{"10", "20"}
			if !reflect.DeepEqual(dest.slice, expected) {
				t.Errorf("main name not set: %v != %v", expected, ctx.StringSlice("serve"))
			}
			if !reflect.DeepEqual(dest.slice, expected) {
				t.Errorf("short name not set: %v != %v", expected, ctx.StringSlice("s"))
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiStringSliceWithDestinationAndEnv(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,40")

	dest := &StringSlice{}
	_ = (&App{
		Flags: []Flag{
			&StringSliceFlag{Name: "serve", Aliases: []string{"s"}, Destination: dest, EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			expected := []string{"10", "20"}
			if !reflect.DeepEqual(dest.slice, expected) {
				t.Errorf("main name not set: %v != %v", expected, ctx.StringSlice("serve"))
			}
			if !reflect.DeepEqual(dest.slice, expected) {
				t.Errorf("short name not set: %v != %v", expected, ctx.StringSlice("s"))
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiStringSliceWithDefaultsUnset(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&StringSliceFlag{Name: "serve", Aliases: []string{"s"}, Value: NewStringSlice("9", "2")},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.StringSlice("serve"), []string{"9", "2"}) {
				t.Errorf("main name not set: %v", ctx.StringSlice("serve"))
			}
			if !reflect.DeepEqual(ctx.StringSlice("s"), []string{"9", "2"}) {
				t.Errorf("short name not set: %v", ctx.StringSlice("s"))
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiStringSliceFromEnv(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&App{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: NewStringSlice(), EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.StringSlice("intervals"), []string{"20", "30", "40"}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.StringSlice("i"), []string{"20", "30", "40"}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiStringSliceFromEnvWithDefaults(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&App{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: NewStringSlice("1", "2", "5"), EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.StringSlice("intervals"), []string{"20", "30", "40"}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.StringSlice("i"), []string{"20", "30", "40"}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiStringSliceFromEnvCascade(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&App{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: NewStringSlice(), EnvVars: []string{"COMPAT_INTERVALS", "APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.StringSlice("intervals"), []string{"20", "30", "40"}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.StringSlice("i"), []string{"20", "30", "40"}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiStringSliceFromEnvCascadeWithDefaults(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&App{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: NewStringSlice("1", "2", "5"), EnvVars: []string{"COMPAT_INTERVALS", "APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.StringSlice("intervals"), []string{"20", "30", "40"}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.StringSlice("i"), []string{"20", "30", "40"}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiStringSliceFromEnvWithDestination(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,40")

	dest := &StringSlice{}
	_ = (&App{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Destination: dest, EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(dest.slice, []string{"20", "30", "40"}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(dest.slice, []string{"20", "30", "40"}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiInt(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&IntFlag{Name: "serve", Aliases: []string{"s"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Int("serve") != 10 {
				t.Errorf("main name not set")
			}
			if ctx.Int("s") != 10 {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10"})
}

func TestParseDestinationInt(t *testing.T) {
	var dest int
	_ = (&App{
		Flags: []Flag{
			&IntFlag{
				Name:        "dest",
				Destination: &dest,
			},
		},
		Action: func(ctx *Context) error {
			if dest != 10 {
				t.Errorf("expected destination Int 10")
			}
			return nil
		},
	}).Run([]string{"run", "--dest", "10"})
}

func TestParseMultiIntFromEnv(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_TIMEOUT_SECONDS", "10")
	_ = (&App{
		Flags: []Flag{
			&IntFlag{Name: "timeout", Aliases: []string{"t"}, EnvVars: []string{"APP_TIMEOUT_SECONDS"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Int("timeout") != 10 {
				t.Errorf("main name not set")
			}
			if ctx.Int("t") != 10 {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiIntFromEnvCascade(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_TIMEOUT_SECONDS", "10")
	_ = (&App{
		Flags: []Flag{
			&IntFlag{Name: "timeout", Aliases: []string{"t"}, EnvVars: []string{"COMPAT_TIMEOUT_SECONDS", "APP_TIMEOUT_SECONDS"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Int("timeout") != 10 {
				t.Errorf("main name not set")
			}
			if ctx.Int("t") != 10 {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiIntSlice(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&IntSliceFlag{Name: "serve", Aliases: []string{"s"}, Value: NewIntSlice()},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.IntSlice("serve"), []int{10, 20}) {
				t.Errorf("main name not set")
			}
			if !reflect.DeepEqual(ctx.IntSlice("s"), []int{10, 20}) {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiIntSliceWithDefaults(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&IntSliceFlag{Name: "serve", Aliases: []string{"s"}, Value: NewIntSlice(9, 2)},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.IntSlice("serve"), []int{10, 20}) {
				t.Errorf("main name not set")
			}
			if !reflect.DeepEqual(ctx.IntSlice("s"), []int{10, 20}) {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiIntSliceWithDefaultsUnset(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&IntSliceFlag{Name: "serve", Aliases: []string{"s"}, Value: NewIntSlice(9, 2)},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.IntSlice("serve"), []int{9, 2}) {
				t.Errorf("main name not set")
			}
			if !reflect.DeepEqual(ctx.IntSlice("s"), []int{9, 2}) {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiIntSliceFromEnv(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&App{
		Flags: []Flag{
			&IntSliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: NewIntSlice(), EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.IntSlice("intervals"), []int{20, 30, 40}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.IntSlice("i"), []int{20, 30, 40}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiIntSliceFromEnvWithDefaults(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&App{
		Flags: []Flag{
			&IntSliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: NewIntSlice(1, 2, 5), EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.IntSlice("intervals"), []int{20, 30, 40}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.IntSlice("i"), []int{20, 30, 40}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiIntSliceFromEnvCascade(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&App{
		Flags: []Flag{
			&IntSliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: NewIntSlice(), EnvVars: []string{"COMPAT_INTERVALS", "APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.IntSlice("intervals"), []int{20, 30, 40}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.IntSlice("i"), []int{20, 30, 40}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiInt64Slice(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&Int64SliceFlag{Name: "serve", Aliases: []string{"s"}, Value: NewInt64Slice()},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Int64Slice("serve"), []int64{10, 17179869184}) {
				t.Errorf("main name not set")
			}
			if !reflect.DeepEqual(ctx.Int64Slice("s"), []int64{10, 17179869184}) {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10", "-s", "17179869184"})
}

func TestParseMultiInt64SliceFromEnv(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,17179869184")

	_ = (&App{
		Flags: []Flag{
			&Int64SliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: NewInt64Slice(), EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Int64Slice("intervals"), []int64{20, 30, 17179869184}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.Int64Slice("i"), []int64{20, 30, 17179869184}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiInt64SliceFromEnvCascade(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,17179869184")

	_ = (&App{
		Flags: []Flag{
			&Int64SliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: NewInt64Slice(), EnvVars: []string{"COMPAT_INTERVALS", "APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Int64Slice("intervals"), []int64{20, 30, 17179869184}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.Int64Slice("i"), []int64{20, 30, 17179869184}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiFloat64(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&Float64Flag{Name: "serve", Aliases: []string{"s"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Float64("serve") != 10.2 {
				t.Errorf("main name not set")
			}
			if ctx.Float64("s") != 10.2 {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10.2"})
}

func TestParseDestinationFloat64(t *testing.T) {
	var dest float64
	_ = (&App{
		Flags: []Flag{
			&Float64Flag{
				Name:        "dest",
				Destination: &dest,
			},
		},
		Action: func(ctx *Context) error {
			if dest != 10.2 {
				t.Errorf("expected destination Float64 10.2")
			}
			return nil
		},
	}).Run([]string{"run", "--dest", "10.2"})
}

func TestParseMultiFloat64FromEnv(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_TIMEOUT_SECONDS", "15.5")
	_ = (&App{
		Flags: []Flag{
			&Float64Flag{Name: "timeout", Aliases: []string{"t"}, EnvVars: []string{"APP_TIMEOUT_SECONDS"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Float64("timeout") != 15.5 {
				t.Errorf("main name not set")
			}
			if ctx.Float64("t") != 15.5 {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiFloat64FromEnvCascade(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_TIMEOUT_SECONDS", "15.5")
	_ = (&App{
		Flags: []Flag{
			&Float64Flag{Name: "timeout", Aliases: []string{"t"}, EnvVars: []string{"COMPAT_TIMEOUT_SECONDS", "APP_TIMEOUT_SECONDS"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Float64("timeout") != 15.5 {
				t.Errorf("main name not set")
			}
			if ctx.Float64("t") != 15.5 {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiFloat64SliceFromEnv(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "0.1,-10.5")

	_ = (&App{
		Flags: []Flag{
			&Float64SliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: NewFloat64Slice(), EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Float64Slice("intervals"), []float64{0.1, -10.5}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.Float64Slice("i"), []float64{0.1, -10.5}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiFloat64SliceFromEnvCascade(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "0.1234,-10.5")

	_ = (&App{
		Flags: []Flag{
			&Float64SliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: NewFloat64Slice(), EnvVars: []string{"COMPAT_INTERVALS", "APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Float64Slice("intervals"), []float64{0.1234, -10.5}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.Float64Slice("i"), []float64{0.1234, -10.5}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiBool(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&BoolFlag{Name: "serve", Aliases: []string{"s"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("serve") != true {
				t.Errorf("main name not set")
			}
			if ctx.Bool("s") != true {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "--serve"})
}

func TestParseBoolShortOptionHandle(t *testing.T) {
	_ = (&App{
		Commands: []*Command{
			{
				Name:                   "foobar",
				UseShortOptionHandling: true,
				Action: func(ctx *Context) error {
					if ctx.Bool("serve") != true {
						t.Errorf("main name not set")
					}
					if ctx.Bool("option") != true {
						t.Errorf("short name not set")
					}
					return nil
				},
				Flags: []Flag{
					&BoolFlag{Name: "serve", Aliases: []string{"s"}},
					&BoolFlag{Name: "option", Aliases: []string{"o"}},
				},
			},
		},
	}).Run([]string{"run", "foobar", "-so"})
}

func TestParseDestinationBool(t *testing.T) {
	var dest bool
	_ = (&App{
		Flags: []Flag{
			&BoolFlag{
				Name:        "dest",
				Destination: &dest,
			},
		},
		Action: func(ctx *Context) error {
			if dest != true {
				t.Errorf("expected destination Bool true")
			}
			return nil
		},
	}).Run([]string{"run", "--dest"})
}

func TestParseMultiBoolFromEnv(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_DEBUG", "1")
	_ = (&App{
		Flags: []Flag{
			&BoolFlag{Name: "debug", Aliases: []string{"d"}, EnvVars: []string{"APP_DEBUG"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("debug") != true {
				t.Errorf("main name not set from env")
			}
			if ctx.Bool("d") != true {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiBoolFromEnvCascade(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_DEBUG", "1")
	_ = (&App{
		Flags: []Flag{
			&BoolFlag{Name: "debug", Aliases: []string{"d"}, EnvVars: []string{"COMPAT_DEBUG", "APP_DEBUG"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("debug") != true {
				t.Errorf("main name not set from env")
			}
			if ctx.Bool("d") != true {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseBoolFromEnv(t *testing.T) {
	var boolFlagTests = []struct {
		input  string
		output bool
	}{
		{"", false},
		{"1", true},
		{"false", false},
		{"true", true},
	}

	for _, test := range boolFlagTests {
		defer resetEnv(os.Environ())
		os.Clearenv()
		_ = os.Setenv("DEBUG", test.input)
		_ = (&App{
			Flags: []Flag{
				&BoolFlag{Name: "debug", Aliases: []string{"d"}, EnvVars: []string{"DEBUG"}},
			},
			Action: func(ctx *Context) error {
				if ctx.Bool("debug") != test.output {
					t.Errorf("expected %+v to be parsed as %+v, instead was %+v", test.input, test.output, ctx.Bool("debug"))
				}
				if ctx.Bool("d") != test.output {
					t.Errorf("expected %+v to be parsed as %+v, instead was %+v", test.input, test.output, ctx.Bool("d"))
				}
				return nil
			},
		}).Run([]string{"run"})
	}
}

func TestParseMultiBoolT(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&BoolFlag{Name: "implode", Aliases: []string{"i"}, Value: true},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("implode") {
				t.Errorf("main name not set")
			}
			if ctx.Bool("i") {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "--implode=false"})
}

type Parser [2]string

func (p *Parser) Set(value string) error {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format")
	}

	(*p)[0] = parts[0]
	(*p)[1] = parts[1]

	return nil
}

func (p *Parser) String() string {
	return fmt.Sprintf("%s,%s", p[0], p[1])
}

func (p *Parser) Get() interface{} {
	return p
}

func TestParseGeneric(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			&GenericFlag{Name: "serve", Aliases: []string{"s"}, Value: &Parser{}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Generic("serve"), &Parser{"10", "20"}) {
				t.Errorf("main name not set")
			}
			if !reflect.DeepEqual(ctx.Generic("s"), &Parser{"10", "20"}) {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10,20"})
}

func TestParseGenericFromEnv(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_SERVE", "20,30")
	_ = (&App{
		Flags: []Flag{
			&GenericFlag{
				Name:    "serve",
				Aliases: []string{"s"},
				Value:   &Parser{},
				EnvVars: []string{"APP_SERVE"},
			},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Generic("serve"), &Parser{"20", "30"}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.Generic("s"), &Parser{"20", "30"}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseGenericFromEnvCascade(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("APP_FOO", "99,2000")
	_ = (&App{
		Flags: []Flag{
			&GenericFlag{
				Name:    "foos",
				Value:   &Parser{},
				EnvVars: []string{"COMPAT_FOO", "APP_FOO"},
			},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Generic("foos"), &Parser{"99", "2000"}) {
				t.Errorf("value not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestFlagFromFile(t *testing.T) {
	temp, err := ioutil.TempFile("", "urfave_cli_test")
	if err != nil {
		t.Error(err)
		return
	}

	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_FOO", "123")

	_, _ = io.WriteString(temp, "abc")
	_ = temp.Close()
	defer func() {
		_ = os.Remove(temp.Name())
	}()

	var filePathTests = []struct {
		path     string
		name     []string
		expected string
	}{
		{"file-does-not-exist", []string{"APP_BAR"}, ""},
		{"file-does-not-exist", []string{"APP_FOO"}, "123"},
		{"file-does-not-exist", []string{"APP_FOO", "APP_BAR"}, "123"},
		{temp.Name(), []string{"APP_FOO"}, "123"},
		{temp.Name(), []string{"APP_BAR"}, "abc"},
	}

	for _, filePathTest := range filePathTests {
		got, _ := flagFromEnvOrFile(filePathTest.name, filePathTest.path)
		if want := filePathTest.expected; got != want {
			t.Errorf("Did not expect %v - Want %v", got, want)
		}
	}
}

func TestStringSlice_Serialized_Set(t *testing.T) {
	sl0 := NewStringSlice("a", "b")
	ser0 := sl0.Serialize()

	if len(ser0) < len(slPfx) {
		t.Fatalf("serialized shorter than expected: %q", ser0)
	}

	sl1 := NewStringSlice("c", "d")
	_ = sl1.Set(ser0)

	if sl0.String() != sl1.String() {
		t.Fatalf("pre and post serialization do not match: %v != %v", sl0, sl1)
	}
}

func TestIntSlice_Serialized_Set(t *testing.T) {
	sl0 := NewIntSlice(1, 2)
	ser0 := sl0.Serialize()

	if len(ser0) < len(slPfx) {
		t.Fatalf("serialized shorter than expected: %q", ser0)
	}

	sl1 := NewIntSlice(3, 4)
	_ = sl1.Set(ser0)

	if sl0.String() != sl1.String() {
		t.Fatalf("pre and post serialization do not match: %v != %v", sl0, sl1)
	}
}

func TestInt64Slice_Serialized_Set(t *testing.T) {
	sl0 := NewInt64Slice(int64(1), int64(2))
	ser0 := sl0.Serialize()

	if len(ser0) < len(slPfx) {
		t.Fatalf("serialized shorter than expected: %q", ser0)
	}

	sl1 := NewInt64Slice(int64(3), int64(4))
	_ = sl1.Set(ser0)

	if sl0.String() != sl1.String() {
		t.Fatalf("pre and post serialization do not match: %v != %v", sl0, sl1)
	}
}

func TestTimestamp_set(t *testing.T) {
	ts := Timestamp{
		timestamp:  nil,
		hasBeenSet: false,
		layout:     "Jan 2, 2006 at 3:04pm (MST)",
	}

	time1 := "Feb 3, 2013 at 7:54pm (PST)"
	if err := ts.Set(time1); err != nil {
		t.Fatalf("Failed to parse time %s with layout %s", time1, ts.layout)
	}
	if ts.hasBeenSet == false {
		t.Fatalf("hasBeenSet is not true after setting a time")
	}

	ts.hasBeenSet = false
	ts.SetLayout(time.RFC3339)
	time2 := "2006-01-02T15:04:05Z"
	if err := ts.Set(time2); err != nil {
		t.Fatalf("Failed to parse time %s with layout %s", time2, ts.layout)
	}
	if ts.hasBeenSet == false {
		t.Fatalf("hasBeenSet is not true after setting a time")
	}
}

func TestTimestampFlagApply(t *testing.T) {
	expectedResult, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	fl := TimestampFlag{Name: "time", Aliases: []string{"t"}, Layout: time.RFC3339}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse([]string{"--time", "2006-01-02T15:04:05Z"})
	expect(t, err, nil)
	expect(t, *fl.Value.timestamp, expectedResult)
}

func TestTimestampFlagApplyValue(t *testing.T) {
	expectedResult, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	fl := TimestampFlag{Name: "time", Aliases: []string{"t"}, Layout: time.RFC3339, Value: NewTimestamp(expectedResult)}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse([]string{""})
	expect(t, err, nil)
	expect(t, *fl.Value.timestamp, expectedResult)
}

func TestTimestampFlagApply_Fail_Parse_Wrong_Layout(t *testing.T) {
	fl := TimestampFlag{Name: "time", Aliases: []string{"t"}, Layout: "randomlayout"}
	set := flag.NewFlagSet("test", 0)
	set.SetOutput(ioutil.Discard)
	_ = fl.Apply(set)

	err := set.Parse([]string{"--time", "2006-01-02T15:04:05Z"})
	expect(t, err, fmt.Errorf("invalid value \"2006-01-02T15:04:05Z\" for flag -time: parsing time \"2006-01-02T15:04:05Z\" as \"randomlayout\": cannot parse \"2006-01-02T15:04:05Z\" as \"randomlayout\""))
}

func TestTimestampFlagApply_Fail_Parse_Wrong_Time(t *testing.T) {
	fl := TimestampFlag{Name: "time", Aliases: []string{"t"}, Layout: "Jan 2, 2006 at 3:04pm (MST)"}
	set := flag.NewFlagSet("test", 0)
	set.SetOutput(ioutil.Discard)
	_ = fl.Apply(set)

	err := set.Parse([]string{"--time", "2006-01-02T15:04:05Z"})
	expect(t, err, fmt.Errorf("invalid value \"2006-01-02T15:04:05Z\" for flag -time: parsing time \"2006-01-02T15:04:05Z\" as \"Jan 2, 2006 at 3:04pm (MST)\": cannot parse \"2006-01-02T15:04:05Z\" as \"Jan\""))
}

type flagDefaultTestCase struct {
	name    string
	flag    Flag
	toParse []string
	expect  string
}

func TestFlagDefaultValue(t *testing.T) {
	cases := []*flagDefaultTestCase{
		&flagDefaultTestCase{
			name:    "stringSclice",
			flag:    &StringSliceFlag{Name: "flag", Value: NewStringSlice("default1", "default2")},
			toParse: []string{"--flag", "parsed"},
			expect: `--flag value	(default: "default1", "default2")	(accepts multiple inputs)`,
		},
		&flagDefaultTestCase{
			name:    "float64Sclice",
			flag:    &Float64SliceFlag{Name: "flag", Value: NewFloat64Slice(1.1, 2.2)},
			toParse: []string{"--flag", "13.3"},
			expect: `--flag value	(default: 1.1, 2.2)	(accepts multiple inputs)`,
		},
		&flagDefaultTestCase{
			name:    "int64Sclice",
			flag:    &Int64SliceFlag{Name: "flag", Value: NewInt64Slice(1, 2)},
			toParse: []string{"--flag", "13"},
			expect: `--flag value	(default: 1, 2)	(accepts multiple inputs)`,
		},
		&flagDefaultTestCase{
			name:    "intSclice",
			flag:    &IntSliceFlag{Name: "flag", Value: NewIntSlice(1, 2)},
			toParse: []string{"--flag", "13"},
			expect: `--flag value	(default: 1, 2)	(accepts multiple inputs)`,
		},
		&flagDefaultTestCase{
			name:    "string",
			flag:    &StringFlag{Name: "flag", Value: "default"},
			toParse: []string{"--flag", "parsed"},
			expect: `--flag value	(default: "default")`,
		},
		&flagDefaultTestCase{
			name:    "bool",
			flag:    &BoolFlag{Name: "flag", Value: true},
			toParse: []string{"--flag", "false"},
			expect: `--flag	(default: true)`,
		},
		&flagDefaultTestCase{
			name:    "uint64",
			flag:    &Uint64Flag{Name: "flag", Value: 1},
			toParse: []string{"--flag", "13"},
			expect: `--flag value	(default: 1)`,
		},
	}
	for i, v := range cases {
		set := flag.NewFlagSet("test", 0)
		set.SetOutput(ioutil.Discard)
		_ = v.flag.Apply(set)
		if err := set.Parse(v.toParse); err != nil {
			t.Error(err)
		}
		if got := v.flag.String(); got != v.expect {
			t.Errorf("TestFlagDefaultValue %d %s\nexpect:%s\ngot:%s", i, v.name, v.expect, got)
		}
	}
}

func TestTimestampFlagApply_WithDestination(t *testing.T) {
	var destination Timestamp
	expectedResult, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	fl := TimestampFlag{Name: "time", Aliases: []string{"t"}, Layout: time.RFC3339, Destination: &destination}
	set := flag.NewFlagSet("test", 0)
	_ = fl.Apply(set)

	err := set.Parse([]string{"--time", "2006-01-02T15:04:05Z"})
	expect(t, err, nil)
	expect(t, *fl.Destination.timestamp, expectedResult)
}
