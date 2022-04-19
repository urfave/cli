package cli

import (
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
	{"help", "--help\t"},
	{"h", "-h\t"},
}

func TestBoolFlagHelpOutput(t *testing.T) {
	for _, test := range boolFlagTests {
		flag := BoolFlag{Name: test.name}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestFlagsFromEnv(t *testing.T) {
	var flagTests = []struct {
		input     string
		output    interface{}
		flag      Flag
		errRegexp string
	}{
		{"", false, BoolFlag{Name: "debug", EnvVar: "DEBUG"}, ""},
		{"1", true, BoolFlag{Name: "debug", EnvVar: "DEBUG"}, ""},
		{"false", false, BoolFlag{Name: "debug", EnvVar: "DEBUG"}, ""},
		{"foobar", true, BoolFlag{Name: "debug", EnvVar: "DEBUG"}, fmt.Sprintf(`could not parse foobar as bool value for flag debug: .*`)},

		{"", false, BoolTFlag{Name: "debug", EnvVar: "DEBUG"}, ""},
		{"1", true, BoolTFlag{Name: "debug", EnvVar: "DEBUG"}, ""},
		{"false", false, BoolTFlag{Name: "debug", EnvVar: "DEBUG"}, ""},
		{"foobar", true, BoolTFlag{Name: "debug", EnvVar: "DEBUG"}, fmt.Sprintf(`could not parse foobar as bool value for flag debug: .*`)},

		{"1s", 1 * time.Second, DurationFlag{Name: "time", EnvVar: "TIME"}, ""},
		{"foobar", false, DurationFlag{Name: "time", EnvVar: "TIME"}, fmt.Sprintf(`could not parse foobar as duration for flag time: .*`)},

		{"1.2", 1.2, Float64Flag{Name: "seconds", EnvVar: "SECONDS"}, ""},
		{"1", 1.0, Float64Flag{Name: "seconds", EnvVar: "SECONDS"}, ""},
		{"foobar", 0, Float64Flag{Name: "seconds", EnvVar: "SECONDS"}, fmt.Sprintf(`could not parse foobar as float64 value for flag seconds: .*`)},

		{"1", int64(1), Int64Flag{Name: "seconds", EnvVar: "SECONDS"}, ""},
		{"1.2", 0, Int64Flag{Name: "seconds", EnvVar: "SECONDS"}, fmt.Sprintf(`could not parse 1.2 as int value for flag seconds: .*`)},
		{"foobar", 0, Int64Flag{Name: "seconds", EnvVar: "SECONDS"}, fmt.Sprintf(`could not parse foobar as int value for flag seconds: .*`)},

		{"1", 1, IntFlag{Name: "seconds", EnvVar: "SECONDS"}, ""},
		{"1.2", 0, IntFlag{Name: "seconds", EnvVar: "SECONDS"}, fmt.Sprintf(`could not parse 1.2 as int value for flag seconds: .*`)},
		{"foobar", 0, IntFlag{Name: "seconds", EnvVar: "SECONDS"}, fmt.Sprintf(`could not parse foobar as int value for flag seconds: .*`)},

		{"1,2", IntSlice{1, 2}, IntSliceFlag{Name: "seconds", EnvVar: "SECONDS"}, ""},
		{"1.2,2", IntSlice{}, IntSliceFlag{Name: "seconds", EnvVar: "SECONDS"}, fmt.Sprintf(`could not parse 1.2,2 as int slice value for flag seconds: .*`)},
		{"foobar", IntSlice{}, IntSliceFlag{Name: "seconds", EnvVar: "SECONDS"}, fmt.Sprintf(`could not parse foobar as int slice value for flag seconds: .*`)},

		{"1,2", Int64Slice{1, 2}, Int64SliceFlag{Name: "seconds", EnvVar: "SECONDS"}, ""},
		{"1.2,2", Int64Slice{}, Int64SliceFlag{Name: "seconds", EnvVar: "SECONDS"}, fmt.Sprintf(`could not parse 1.2,2 as int64 slice value for flag seconds: .*`)},
		{"foobar", Int64Slice{}, Int64SliceFlag{Name: "seconds", EnvVar: "SECONDS"}, fmt.Sprintf(`could not parse foobar as int64 slice value for flag seconds: .*`)},

		{"foo", "foo", StringFlag{Name: "name", EnvVar: "NAME"}, ""},

		{"foo,bar", StringSlice{"foo", "bar"}, StringSliceFlag{Name: "names", EnvVar: "NAMES"}, ""},

		{"1", uint(1), UintFlag{Name: "seconds", EnvVar: "SECONDS"}, ""},
		{"1.2", 0, UintFlag{Name: "seconds", EnvVar: "SECONDS"}, fmt.Sprintf(`could not parse 1.2 as uint value for flag seconds: .*`)},
		{"foobar", 0, UintFlag{Name: "seconds", EnvVar: "SECONDS"}, fmt.Sprintf(`could not parse foobar as uint value for flag seconds: .*`)},

		{"1", uint64(1), Uint64Flag{Name: "seconds", EnvVar: "SECONDS"}, ""},
		{"1.2", 0, Uint64Flag{Name: "seconds", EnvVar: "SECONDS"}, fmt.Sprintf(`could not parse 1.2 as uint64 value for flag seconds: .*`)},
		{"foobar", 0, Uint64Flag{Name: "seconds", EnvVar: "SECONDS"}, fmt.Sprintf(`could not parse foobar as uint64 value for flag seconds: .*`)},

		{"foo,bar", &Parser{"foo", "bar"}, GenericFlag{Name: "names", Value: &Parser{}, EnvVar: "NAMES"}, ""},
	}

	for _, test := range flagTests {
		os.Clearenv()
		_ = os.Setenv(reflect.ValueOf(test.flag).FieldByName("EnvVar").String(), test.input)
		a := App{
			Flags: []Flag{test.flag},
			Action: func(ctx *Context) error {
				if !reflect.DeepEqual(ctx.value(test.flag.GetName()), test.output) {
					t.Errorf("expected %+v to be parsed as %+v, instead was %+v", test.input, test.output, ctx.value(test.flag.GetName()))
				}
				return nil
			},
		}

		err := a.Run([]string{"run"})

		if test.errRegexp != "" {
			if err == nil {
				t.Errorf("expected error to match %s, got none", test.errRegexp)
			} else {
				if matched, _ := regexp.MatchString(test.errRegexp, err.Error()); !matched {
					t.Errorf("expected error to match %s, got error %s", test.errRegexp, err)
				}
			}
		} else {
			if err != nil && test.errRegexp == "" {
				t.Errorf("expected no error got %s", err)
			}
		}
	}
}

var stringFlagTests = []struct {
	name     string
	usage    string
	value    string
	expected string
}{
	{"foo", "", "", "--foo value\t"},
	{"f", "", "", "-f value\t"},
	{"f", "The total `foo` desired", "all", "-f foo\tThe total foo desired (default: \"all\")"},
	{"test", "", "Something", "--test value\t(default: \"Something\")"},
	{"config,c", "Load configuration from `FILE`", "", "--config FILE, -c FILE\tLoad configuration from FILE"},
	{"config,c", "Load configuration from `CONFIG`", "config.json", "--config CONFIG, -c CONFIG\tLoad configuration from CONFIG (default: \"config.json\")"},
}

func TestStringFlagHelpOutput(t *testing.T) {
	for _, test := range stringFlagTests {
		flag := StringFlag{Name: test.name, Usage: test.usage, Value: test.value}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestStringFlagWithEnvVarHelpOutput(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_FOO", "derp")
	for _, test := range stringFlagTests {
		flag := StringFlag{Name: test.name, Value: test.value, EnvVar: "APP_FOO"}
		output := flag.String()

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
	usage    string
	value    string
	prefixer FlagNamePrefixFunc
	expected string
}{
	{"foo", "", "", func(a, b string) string {
		return fmt.Sprintf("name: %s, ph: %s", a, b)
	}, "name: foo, ph: value\t"},
	{"f", "", "", func(a, b string) string {
		return fmt.Sprintf("name: %s, ph: %s", a, b)
	}, "name: f, ph: value\t"},
	{"f", "The total `foo` desired", "all", func(a, b string) string {
		return fmt.Sprintf("name: %s, ph: %s", a, b)
	}, "name: f, ph: foo\tThe total foo desired (default: \"all\")"},
	{"test", "", "Something", func(a, b string) string {
		return fmt.Sprintf("name: %s, ph: %s", a, b)
	}, "name: test, ph: value\t(default: \"Something\")"},
	{"config,c", "Load configuration from `FILE`", "", func(a, b string) string {
		return fmt.Sprintf("name: %s, ph: %s", a, b)
	}, "name: config,c, ph: FILE\tLoad configuration from FILE"},
	{"config,c", "Load configuration from `CONFIG`", "config.json", func(a, b string) string {
		return fmt.Sprintf("name: %s, ph: %s", a, b)
	}, "name: config,c, ph: CONFIG\tLoad configuration from CONFIG (default: \"config.json\")"},
}

func TestFlagNamePrefixer(t *testing.T) {
	defer func() {
		FlagNamePrefixer = prefixedNames
	}()

	for _, test := range prefixStringFlagTests {
		FlagNamePrefixer = test.prefixer
		flag := StringFlag{Name: test.name, Usage: test.usage, Value: test.value}
		output := flag.String()
		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

var envHintFlagTests = []struct {
	name     string
	env      string
	hinter   FlagEnvHintFunc
	expected string
}{
	{"foo", "", func(a, b string) string {
		return fmt.Sprintf("env: %s, str: %s", a, b)
	}, "env: , str: --foo value\t"},
	{"f", "", func(a, b string) string {
		return fmt.Sprintf("env: %s, str: %s", a, b)
	}, "env: , str: -f value\t"},
	{"foo", "ENV_VAR", func(a, b string) string {
		return fmt.Sprintf("env: %s, str: %s", a, b)
	}, "env: ENV_VAR, str: --foo value\t"},
	{"f", "ENV_VAR", func(a, b string) string {
		return fmt.Sprintf("env: %s, str: %s", a, b)
	}, "env: ENV_VAR, str: -f value\t"},
}

func TestFlagEnvHinter(t *testing.T) {
	defer func() {
		FlagEnvHinter = withEnvHint
	}()

	for _, test := range envHintFlagTests {
		FlagEnvHinter = test.hinter
		flag := StringFlag{Name: test.name, EnvVar: test.env}
		output := flag.String()
		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

var stringSliceFlagTests = []struct {
	name     string
	value    *StringSlice
	expected string
}{
	{"foo", func() *StringSlice {
		s := &StringSlice{}
		_ = s.Set("")
		return s
	}(), "--foo value\t"},
	{"f", func() *StringSlice {
		s := &StringSlice{}
		_ = s.Set("")
		return s
	}(), "-f value\t"},
	{"f", func() *StringSlice {
		s := &StringSlice{}
		_ = s.Set("Lipstick")
		return s
	}(), "-f value\t(default: \"Lipstick\")"},
	{"test", func() *StringSlice {
		s := &StringSlice{}
		_ = s.Set("Something")
		return s
	}(), "--test value\t(default: \"Something\")"},
}

func TestStringSliceFlagHelpOutput(t *testing.T) {
	for _, test := range stringSliceFlagTests {
		flag := StringSliceFlag{Name: test.name, Value: test.value}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestStringSliceFlagWithEnvVarHelpOutput(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_QWWX", "11,4")
	for _, test := range stringSliceFlagTests {
		flag := StringSliceFlag{Name: test.name, Value: test.value, EnvVar: "APP_QWWX"}
		output := flag.String()

		expectedSuffix := " [$APP_QWWX]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_QWWX%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
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
		flag := IntFlag{Name: test.name, Value: 9}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

func TestIntFlagWithEnvVarHelpOutput(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_BAR", "2")
	for _, test := range intFlagTests {
		flag := IntFlag{Name: test.name, EnvVar: "APP_BAR"}
		output := flag.String()

		expectedSuffix := " [$APP_BAR]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAR%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
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
		flag := Int64Flag{Name: test.name, Value: 8589934592}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

func TestInt64FlagWithEnvVarHelpOutput(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_BAR", "2")
	for _, test := range int64FlagTests {
		flag := IntFlag{Name: test.name, EnvVar: "APP_BAR"}
		output := flag.String()

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
		flag := UintFlag{Name: test.name, Value: 41}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

func TestUintFlagWithEnvVarHelpOutput(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_BAR", "2")
	for _, test := range uintFlagTests {
		flag := UintFlag{Name: test.name, EnvVar: "APP_BAR"}
		output := flag.String()

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
		flag := Uint64Flag{Name: test.name, Value: 8589934582}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

func TestUint64FlagWithEnvVarHelpOutput(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_BAR", "2")
	for _, test := range uint64FlagTests {
		flag := UintFlag{Name: test.name, EnvVar: "APP_BAR"}
		output := flag.String()

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
		flag := DurationFlag{Name: test.name, Value: 1 * time.Second}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestDurationFlagWithEnvVarHelpOutput(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_BAR", "2h3m6s")
	for _, test := range durationFlagTests {
		flag := DurationFlag{Name: test.name, EnvVar: "APP_BAR"}
		output := flag.String()

		expectedSuffix := " [$APP_BAR]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAR%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

var intSliceFlagTests = []struct {
	name     string
	value    *IntSlice
	expected string
}{
	{"heads", &IntSlice{}, "--heads value\t"},
	{"H", &IntSlice{}, "-H value\t"},
	{"H, heads", func() *IntSlice {
		i := &IntSlice{}
		_ = i.Set("9")
		_ = i.Set("3")
		return i
	}(), "-H value, --heads value\t(default: 9, 3)"},
}

func TestIntSliceFlagHelpOutput(t *testing.T) {
	for _, test := range intSliceFlagTests {
		flag := IntSliceFlag{Name: test.name, Value: test.value}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestIntSliceFlagWithEnvVarHelpOutput(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_SMURF", "42,3")
	for _, test := range intSliceFlagTests {
		flag := IntSliceFlag{Name: test.name, Value: test.value, EnvVar: "APP_SMURF"}
		output := flag.String()

		expectedSuffix := " [$APP_SMURF]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_SMURF%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

var int64SliceFlagTests = []struct {
	name     string
	value    *Int64Slice
	expected string
}{
	{"heads", &Int64Slice{}, "--heads value\t"},
	{"H", &Int64Slice{}, "-H value\t"},
	{"H, heads", func() *Int64Slice {
		i := &Int64Slice{}
		_ = i.Set("2")
		_ = i.Set("17179869184")
		return i
	}(), "-H value, --heads value\t(default: 2, 17179869184)"},
}

func TestInt64SliceFlagHelpOutput(t *testing.T) {
	for _, test := range int64SliceFlagTests {
		flag := Int64SliceFlag{Name: test.name, Value: test.value}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestInt64SliceFlagWithEnvVarHelpOutput(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_SMURF", "42,17179869184")
	for _, test := range int64SliceFlagTests {
		flag := Int64SliceFlag{Name: test.name, Value: test.value, EnvVar: "APP_SMURF"}
		output := flag.String()

		expectedSuffix := " [$APP_SMURF]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_SMURF%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
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
		flag := Float64Flag{Name: test.name, Value: 0.1}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestFloat64FlagWithEnvVarHelpOutput(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_BAZ", "99.4")
	for _, test := range float64FlagTests {
		flag := Float64Flag{Name: test.name, EnvVar: "APP_BAZ"}
		output := flag.String()

		expectedSuffix := " [$APP_BAZ]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAZ%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
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
		flag := GenericFlag{Name: test.name, Value: test.value, Usage: "test flag"}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestGenericFlagWithEnvVarHelpOutput(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_ZAP", "3")
	for _, test := range genericFlagTests {
		flag := GenericFlag{Name: test.name, EnvVar: "APP_ZAP"}
		output := flag.String()

		expectedSuffix := " [$APP_ZAP]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_ZAP%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

func TestParseMultiString(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			StringFlag{Name: "serve, s"},
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
			StringFlag{
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
	os.Clearenv()
	_ = os.Setenv("APP_COUNT", "20")
	_ = (&App{
		Flags: []Flag{
			StringFlag{Name: "count, c", EnvVar: "APP_COUNT"},
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
	os.Clearenv()
	_ = os.Setenv("APP_COUNT", "20")
	_ = (&App{
		Flags: []Flag{
			StringFlag{Name: "count, c", EnvVar: "COMPAT_COUNT,APP_COUNT"},
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
			StringSliceFlag{Name: "serve, s", Value: &StringSlice{}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.StringSlice("serve"), []string{"10", "20"}) {
				t.Errorf("main name not set")
			}
			if !reflect.DeepEqual(ctx.StringSlice("s"), []string{"10", "20"}) {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiStringSliceFromEnv(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&App{
		Flags: []Flag{
			StringSliceFlag{Name: "intervals, i", Value: &StringSlice{}, EnvVar: "APP_INTERVALS"},
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
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&App{
		Flags: []Flag{
			StringSliceFlag{Name: "intervals, i", Value: &StringSlice{}, EnvVar: "COMPAT_INTERVALS,APP_INTERVALS"},
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

func TestParseMultiInt(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			IntFlag{Name: "serve, s"},
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
			IntFlag{
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
	os.Clearenv()
	_ = os.Setenv("APP_TIMEOUT_SECONDS", "10")
	_ = (&App{
		Flags: []Flag{
			IntFlag{Name: "timeout, t", EnvVar: "APP_TIMEOUT_SECONDS"},
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
	os.Clearenv()
	_ = os.Setenv("APP_TIMEOUT_SECONDS", "10")
	_ = (&App{
		Flags: []Flag{
			IntFlag{Name: "timeout, t", EnvVar: "COMPAT_TIMEOUT_SECONDS,APP_TIMEOUT_SECONDS"},
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
			IntSliceFlag{Name: "serve, s", Value: &IntSlice{}},
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

func TestParseMultiIntSliceFromEnv(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&App{
		Flags: []Flag{
			IntSliceFlag{Name: "intervals, i", Value: &IntSlice{}, EnvVar: "APP_INTERVALS"},
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
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&App{
		Flags: []Flag{
			IntSliceFlag{Name: "intervals, i", Value: &IntSlice{}, EnvVar: "COMPAT_INTERVALS,APP_INTERVALS"},
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
			Int64SliceFlag{Name: "serve, s", Value: &Int64Slice{}},
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
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,17179869184")

	_ = (&App{
		Flags: []Flag{
			Int64SliceFlag{Name: "intervals, i", Value: &Int64Slice{}, EnvVar: "APP_INTERVALS"},
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
	os.Clearenv()
	_ = os.Setenv("APP_INTERVALS", "20,30,17179869184")

	_ = (&App{
		Flags: []Flag{
			Int64SliceFlag{Name: "intervals, i", Value: &Int64Slice{}, EnvVar: "COMPAT_INTERVALS,APP_INTERVALS"},
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
			Float64Flag{Name: "serve, s"},
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
			Float64Flag{
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
	os.Clearenv()
	_ = os.Setenv("APP_TIMEOUT_SECONDS", "15.5")
	_ = (&App{
		Flags: []Flag{
			Float64Flag{Name: "timeout, t", EnvVar: "APP_TIMEOUT_SECONDS"},
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
	os.Clearenv()
	_ = os.Setenv("APP_TIMEOUT_SECONDS", "15.5")
	_ = (&App{
		Flags: []Flag{
			Float64Flag{Name: "timeout, t", EnvVar: "COMPAT_TIMEOUT_SECONDS,APP_TIMEOUT_SECONDS"},
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

func TestParseMultiBool(t *testing.T) {
	_ = (&App{
		Flags: []Flag{
			BoolFlag{Name: "serve, s"},
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
		Commands: []Command{
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
					BoolFlag{Name: "serve, s"},
					BoolFlag{Name: "option, o"},
				},
			},
		},
	}).Run([]string{"run", "foobar", "-so"})
}

func TestParseDestinationBool(t *testing.T) {
	var dest bool
	_ = (&App{
		Flags: []Flag{
			BoolFlag{
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
	os.Clearenv()
	_ = os.Setenv("APP_DEBUG", "1")
	_ = (&App{
		Flags: []Flag{
			BoolFlag{Name: "debug, d", EnvVar: "APP_DEBUG"},
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
	os.Clearenv()
	os.Setenv("APP_DEBUG", "1")
	_ = (&App{
		Flags: []Flag{
			BoolFlag{Name: "debug, d", EnvVar: "COMPAT_DEBUG,APP_DEBUG"},
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

func TestParseBoolTFromEnv(t *testing.T) {
	var boolTFlagTests = []struct {
		input  string
		output bool
	}{
		{"", false},
		{"1", true},
		{"false", false},
		{"true", true},
	}

	for _, test := range boolTFlagTests {
		os.Clearenv()
		_ = os.Setenv("DEBUG", test.input)
		_ = (&App{
			Flags: []Flag{
				BoolTFlag{Name: "debug, d", EnvVar: "DEBUG"},
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
			BoolTFlag{Name: "serve, s"},
		},
		Action: func(ctx *Context) error {
			if ctx.BoolT("serve") != true {
				t.Errorf("main name not set")
			}
			if ctx.BoolT("s") != true {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "--serve"})
}

func TestParseDestinationBoolT(t *testing.T) {
	var dest bool
	_ = (&App{
		Flags: []Flag{
			BoolTFlag{
				Name:        "dest",
				Destination: &dest,
			},
		},
		Action: func(ctx *Context) error {
			if dest != true {
				t.Errorf("expected destination BoolT true")
			}
			return nil
		},
	}).Run([]string{"run", "--dest"})
}

func TestParseMultiBoolTFromEnv(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_DEBUG", "0")
	_ = (&App{
		Flags: []Flag{
			BoolTFlag{Name: "debug, d", EnvVar: "APP_DEBUG"},
		},
		Action: func(ctx *Context) error {
			if ctx.BoolT("debug") != false {
				t.Errorf("main name not set from env")
			}
			if ctx.BoolT("d") != false {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiBoolTFromEnvCascade(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_DEBUG", "0")
	_ = (&App{
		Flags: []Flag{
			BoolTFlag{Name: "debug, d", EnvVar: "COMPAT_DEBUG,APP_DEBUG"},
		},
		Action: func(ctx *Context) error {
			if ctx.BoolT("debug") != false {
				t.Errorf("main name not set from env")
			}
			if ctx.BoolT("d") != false {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
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
			GenericFlag{Name: "serve, s", Value: &Parser{}},
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
	os.Clearenv()
	_ = os.Setenv("APP_SERVE", "20,30")
	_ = (&App{
		Flags: []Flag{
			GenericFlag{Name: "serve, s", Value: &Parser{}, EnvVar: "APP_SERVE"},
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
	os.Clearenv()
	_ = os.Setenv("APP_FOO", "99,2000")
	_ = (&App{
		Flags: []Flag{
			GenericFlag{Name: "foos", Value: &Parser{}, EnvVar: "COMPAT_FOO,APP_FOO"},
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
	os.Clearenv()
	os.Setenv("APP_FOO", "123")

	temp, err := ioutil.TempFile("", "urfave_cli_test")
	if err != nil {
		t.Error(err)
		return
	}
	_, _ = io.WriteString(temp, "abc")
	_ = temp.Close()
	defer func() {
		_ = os.Remove(temp.Name())
	}()

	var filePathTests = []struct {
		path     string
		name     string
		expected string
	}{
		{"file-does-not-exist", "APP_BAR", ""},
		{"file-does-not-exist", "APP_FOO", "123"},
		{"file-does-not-exist", "APP_FOO,APP_BAR", "123"},
		{temp.Name(), "APP_FOO", "123"},
		{temp.Name(), "APP_BAR", "abc"},
	}

	for _, filePathTest := range filePathTests {
		got, _ := flagFromFileEnv(filePathTest.path, filePathTest.name)
		if want := filePathTest.expected; got != want {
			t.Errorf("Did not expect %v - Want %v", got, want)
		}
	}
}

func TestSliceFlag_WithDefaults(t *testing.T) {
	tests := []struct {
		args   []string
		app    *App
	}{
		{
			args: []string{""},
			app: &App{
				Flags: []Flag{
					StringSliceFlag{Name: "names, n", Value: &StringSlice{"john"}},
					IntSliceFlag{Name: "userIds, u", Value: &IntSlice{3}},
					Int64SliceFlag{Name: "phoneNumbers, p", Value: &Int64Slice{123456789}},
				},
				Action: func(ctx *Context) error {
					expect(t, len(ctx.StringSlice("n")), 1)
					for _, name := range ctx.StringSlice("names") {
						expect(t, name == "john", true)
					}

					expect(t, len(ctx.IntSlice("u")), 1)
					for _, userId := range ctx.IntSlice("userIds") {
						expect(t, userId == 3, true)
					}

					expect(t, len(ctx.Int64Slice("p")), 1)
					for _, phoneNumber := range ctx.Int64Slice("phoneNumbers") {
						expect(t, phoneNumber == 123456789, true)
					}
					return nil
				},
			},
		},
		{
			args: []string{"", "-n", "jane", "-n", "bob", "-u", "5", "-u", "10", "-p", "987654321"},
			app: &App{
				Flags: []Flag{
					StringSliceFlag{Name: "names, n", Value: &StringSlice{"john"}},
					IntSliceFlag{Name: "userIds, u", Value: &IntSlice{3}},
					Int64SliceFlag{Name: "phoneNumbers, p", Value: &Int64Slice{123456789}},
				},
				Action: func(ctx *Context) error {
					expect(t, len(ctx.StringSlice("n")), 2)
					for _, name := range ctx.StringSlice("names") {
						expect(t, name != "john", true)
					}

					expect(t, len(ctx.IntSlice("u")), 2)
					for _, userId := range ctx.IntSlice("userIds") {
						expect(t, userId != 3, true)
					}

					expect(t, len(ctx.Int64Slice("p")), 1)
					for _, phoneNumber := range ctx.Int64Slice("phoneNumbers") {
						expect(t, phoneNumber != 123456789, true)
					}
					return nil
				},
			},
		},
		{
			args: []string{"", "--names", "john", "--userIds", "3", "--phoneNumbers", "123456789"},
			app: &App{
				Flags: []Flag{
					StringSliceFlag{Name: "names, n", Value: &StringSlice{"john"}},
					IntSliceFlag{Name: "userIds, u", Value: &IntSlice{3}},
					Int64SliceFlag{Name: "phoneNumbers, p", Value: &Int64Slice{123456789}},
				},
				Action: func(ctx *Context) error {
					expect(t, len(ctx.StringSlice("n")), 1)
					for _, name := range ctx.StringSlice("names") {
						expect(t, name == "john", true)
					}

					expect(t, len(ctx.IntSlice("u")), 1)
					for _, userId := range ctx.IntSlice("userIds") {
						expect(t, userId == 3, true)
					}

					expect(t, len(ctx.Int64Slice("p")), 1)
					for _, phoneNumber := range ctx.Int64Slice("phoneNumbers") {
						expect(t, phoneNumber == 123456789, true)
					}
					return nil
				},
			},
		},
	}

	for _, tt := range tests {
		_ = tt.app.Run(tt.args)
	}
}
