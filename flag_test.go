package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

var boolFlagTests = []struct {
	name     string
	expected string
}{
	{"help", "--help\t"},
	{"h", "-h\t"},
}

func TestBoolFlagHelpOutput(t *testing.T) {
	for _, test := range boolFlagTests {
		fl := &BoolFlag{Name: test.name}
		output := fl.String()
		assert.Equal(t, test.expected, output)
	}
}

func TestBoolFlagApply_SetsAllNames(t *testing.T) {
	v := false
	cmd := buildMinimalTestCommand()
	cmd.Flags = []Flag{
		&BoolFlag{Name: "wat", Aliases: []string{"W", "huh"}, Destination: &v},
	}

	err := cmd.Run(buildTestContext(t), []string{"", "--wat", "-W", "--huh"})
	assert.NoError(t, err)
	assert.True(t, v)
}

func TestBoolFlagValueFromCommand(t *testing.T) {
	tf := &BoolFlag{Name: "trueflag"}
	ff := &BoolFlag{Name: "falseflag"}

	cmd := buildMinimalTestCommand()
	cmd.Flags = []Flag{
		tf,
		ff,
	}

	r := require.New(t)
	r.NoError(cmd.Set(tf.Name, "true"))
	r.NoError(cmd.Set(ff.Name, "false"))
	r.True(cmd.Bool(tf.Name))
	r.False(cmd.Bool(ff.Name))
}

func TestBoolFlagApply_SetsCount(t *testing.T) {
	v := false
	count := 0
	cmd := buildMinimalTestCommand()
	cmd.Flags = []Flag{
		&BoolFlag{Name: "wat", Aliases: []string{"W", "huh"}, Destination: &v, Config: BoolConfig{Count: &count}},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{"", "--wat", "-W", "--huh"}))
	assert.True(t, v)
	assert.Equal(t, 3, count)
}

func TestBoolFlagCountFromCommand(t *testing.T) {
	boolCountTests := []struct {
		name          string
		input         []string
		expectedVal   bool
		expectedCount int
	}{
		{
			name:          "3 count",
			input:         []string{"main", "-tf", "-w", "-huh"},
			expectedVal:   true,
			expectedCount: 3,
		},
		{
			name:          "single count",
			input:         []string{"main", "-huh"},
			expectedVal:   true,
			expectedCount: 1,
		},
		{
			name:          "zero count",
			input:         []string{"main"},
			expectedVal:   false,
			expectedCount: 0,
		},
	}

	flags := func() []Flag {
		return []Flag{
			&BoolFlag{Name: "tf", Aliases: []string{"w", "huh"}},
			&BoolWithInverseFlag{Name: "tf", Aliases: []string{"w", "huh"}},
		}
	}
	for index := range flags() {
		for _, bct := range boolCountTests {
			t.Run(bct.name, func(t *testing.T) {
				bf := flags()[index]
				cmd := &Command{
					Flags: []Flag{
						bf,
					},
				}
				r := require.New(t)

				r.NoError(cmd.Run(buildTestContext(t), bct.input))

				for _, alias := range bf.Names() {
					r.Equal(bct.expectedCount, cmd.Count(alias))
					r.Equal(bct.expectedVal, cmd.Value(alias))
				}
			})
		}
	}
}

func TestFlagsFromEnv(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		output      any
		fl          Flag
		errContains string
	}{
		{
			name:   "BoolFlag valid true",
			input:  "1",
			output: true,
			fl:     &BoolFlag{Name: "debug", Sources: EnvVars("DEBUG")},
		},
		{
			name:   "BoolFlag valid false",
			input:  "false",
			output: false,
			fl:     &BoolFlag{Name: "debug", Sources: EnvVars("DEBUG")},
		},
		{
			name:   "BoolFlag invalid",
			input:  "foobar",
			output: true,
			fl:     &BoolFlag{Name: "debug", Sources: EnvVars("DEBUG")},
			errContains: `could not parse "foobar" as bool value from environment variable ` +
				`"DEBUG" for flag debug:`,
		},
		{
			name:   "BoolInverse Empty",
			output: false,
			fl:     &BoolWithInverseFlag{Name: "debug", Sources: EnvVars("DEBUG")},
		},
		{
			name:   "DurationFlag valid",
			input:  "1s",
			output: 1 * time.Second,
			fl:     &DurationFlag{Name: "time", Sources: EnvVars("TIME")},
		},
		{
			name:   "DurationFlag invalid",
			input:  "foobar",
			output: false,
			fl:     &DurationFlag{Name: "time", Sources: EnvVars("TIME")},
			errContains: `could not parse "foobar" as time.Duration value from environment ` +
				`variable "TIME" for flag time:`,
		},

		{
			name:   "Float64Flag valid",
			input:  "1.2",
			output: 1.2,
			fl:     &FloatFlag{Name: "seconds", Sources: EnvVars("SECONDS")},
		},
		{
			name:   "Float64Flag valid from int",
			input:  "1",
			output: 1.0,
			fl:     &FloatFlag{Name: "seconds", Sources: EnvVars("SECONDS")},
		},
		{
			name:   "Float64Flag invalid",
			input:  "foobar",
			output: 0,
			fl:     &FloatFlag{Name: "seconds", Sources: EnvVars("SECONDS")},
			errContains: `could not parse "foobar" as float64 value from environment variable ` +
				`"SECONDS" for flag seconds:`,
		},

		{
			name:   "Int64Flag valid",
			input:  "1",
			output: int64(1),
			fl:     &Int64Flag{Name: "seconds", Sources: EnvVars("SECONDS")},
		},
		{
			name:   "Int64Flag invalid from float",
			input:  "1.2",
			output: 0,
			fl:     &Int64Flag{Name: "seconds", Sources: EnvVars("SECONDS")},
			errContains: `could not parse "1.2" as int64 value from environment variable ` +
				`"SECONDS" for flag seconds:`,
		},
		{
			name:   "Int64Flag invalid",
			input:  "foobar",
			output: 0,
			fl:     &Int64Flag{Name: "seconds", Sources: EnvVars("SECONDS")},
			errContains: `could not parse "foobar" as int64 value from environment variable ` +
				`"SECONDS" for flag seconds:`,
		},
		{
			name:   "Int64Flag valid from hex",
			input:  "deadBEEF",
			output: int64(3735928559),
			fl:     &Int64Flag{Name: "seconds", Sources: EnvVars("SECONDS"), Config: IntegerConfig{Base: 16}},
		},
		{
			name:   "Int64Flag invalid from octal",
			input:  "08",
			output: 0,
			fl:     &Int64Flag{Name: "seconds", Sources: EnvVars("SECONDS"), Config: IntegerConfig{Base: 8}},
			errContains: `could not parse "08" as int64 value from environment variable ` +
				`"SECONDS" for flag seconds:`,
		},

		{
			name:   "Float64SliceFlag valid",
			input:  "1.0,2",
			output: []float64{1, 2},
			fl:     &FloatSliceFlag{Name: "seconds", Sources: EnvVars("SECONDS")},
		},
		{
			name:   "Float64SliceFlag invalid",
			input:  "foobar",
			output: []float64{},
			fl:     &FloatSliceFlag{Name: "seconds", Sources: EnvVars("SECONDS")},
			errContains: `could not parse "foobar" as []float64 value from environment ` +
				`variable "SECONDS" for flag seconds:`,
		},
		{
			name:   "Generic",
			input:  "foo,bar",
			output: &Parser{"foo", "bar"},
			fl:     &GenericFlag{Name: "names", Value: &Parser{}, Sources: EnvVars("NAMES")},
		},
		{
			name:   "Int64SliceFlag valid",
			input:  "1,2",
			output: []int64{1, 2},
			fl:     &Int64SliceFlag{Name: "seconds", Sources: EnvVars("SECONDS")},
		},
		{
			name:   "Int64SliceFlag invalid from float",
			input:  "1.2,2",
			output: []int64{},
			fl:     &Int64SliceFlag{Name: "seconds", Sources: EnvVars("SECONDS")},
			errContains: `could not parse "1.2,2" as []int64 value from environment variable ` +
				`"SECONDS" for flag seconds:`,
		},
		{
			name:   "Int64SliceFlag invalid",
			input:  "foobar",
			output: []int64{},
			fl:     &Int64SliceFlag{Name: "seconds", Sources: EnvVars("SECONDS")},
			errContains: `could not parse "foobar" as []int64 value from environment variable ` +
				`"SECONDS" for flag seconds:`,
		},

		{
			name:   "Uint64SliceFlag valid",
			input:  "1,2",
			output: []uint64{1, 2},
			fl:     &Uint64SliceFlag{Name: "seconds", Sources: EnvVars("SECONDS")},
		},
		{
			name:   "Uint64SliceFlag invalid with float",
			input:  "1.2,2",
			output: []uint64{},
			fl:     &Uint64SliceFlag{Name: "seconds", Sources: EnvVars("SECONDS")},
			errContains: `could not parse "1.2,2" as []uint64 value from environment variable ` +
				`"SECONDS" for flag seconds:`,
		},
		{
			name:   "Uint64SliceFlag invalid",
			input:  "foobar",
			output: []uint64{},
			fl:     &Uint64SliceFlag{Name: "seconds", Sources: EnvVars("SECONDS")},
			errContains: `could not parse "foobar" as []uint64 value from environment variable ` +
				`"SECONDS" for flag seconds:`,
		},

		{
			name:   "StringFlag valid",
			input:  "foo",
			output: "foo",
			fl:     &StringFlag{Name: "name", Sources: EnvVars("NAME")},
		},
		{
			name:   "StringFlag valid with TrimSpace",
			input:  " foo",
			output: "foo",
			fl:     &StringFlag{Name: "names", Sources: EnvVars("NAMES"), Config: StringConfig{TrimSpace: true}},
		},

		{
			name:   "StringSliceFlag valid",
			input:  "foo,bar",
			output: []string{"foo", "bar"},
			fl:     &StringSliceFlag{Name: "names", Sources: EnvVars("NAMES")},
		},
		{
			name:   "StringSliceFlag valid with TrimSpace",
			input:  "foo , bar ",
			output: []string{"foo", "bar"},
			fl:     &StringSliceFlag{Name: "names", Sources: EnvVars("NAMES"), Config: StringConfig{TrimSpace: true}},
		},
		{
			name:   "StringSliceFlag valid without TrimSpace",
			input:  "foo , bar ",
			output: []string{"foo ", " bar "},
			fl:     &StringSliceFlag{Name: "names", Sources: EnvVars("NAMES")},
		},

		{
			name:   "StringMapFlag valid",
			input:  "foo=bar,empty=",
			output: map[string]string{"foo": "bar", "empty": ""},
			fl:     &StringMapFlag{Name: "names", Sources: EnvVars("NAMES")},
		},
		{
			name:   "StringMapFlag valid with TrimSpace",
			input:  "foo= bar ",
			output: map[string]string{"foo": "bar"},
			fl:     &StringMapFlag{Name: "names", Sources: EnvVars("NAMES"), Config: StringConfig{TrimSpace: true}},
		},

		{
			name:   "Uint64Flag valid",
			input:  "1",
			output: uint64(1),
			fl:     &Uint64Flag{Name: "seconds", Sources: EnvVars("SECONDS")},
		},
		{
			name:   "Uint64Flag valid leading zero",
			input:  "08",
			output: uint64(8),
			fl:     &Uint64Flag{Name: "seconds", Sources: EnvVars("SECONDS"), Config: IntegerConfig{Base: 10}},
		},
		{
			name:   "Uint64Flag valid from octal",
			input:  "755",
			output: uint64(493),
			fl:     &Uint64Flag{Name: "seconds", Sources: EnvVars("SECONDS"), Config: IntegerConfig{Base: 8}},
		},
		{
			name:   "Uint64Flag valid from hex",
			input:  "deadBEEF",
			output: uint64(3735928559),
			fl:     &Uint64Flag{Name: "seconds", Sources: EnvVars("SECONDS"), Config: IntegerConfig{Base: 16}},
		},
		{
			name:   "Uint64Flag invalid octal",
			input:  "08",
			output: 0,
			fl:     &Uint64Flag{Name: "seconds", Sources: EnvVars("SECONDS"), Config: IntegerConfig{Base: 8}},
			errContains: `could not parse "08" as uint64 value from environment variable ` +
				`"SECONDS" for flag seconds:`,
		},
		{
			name:   "Uint64Flag invalid float",
			input:  "1.2",
			output: 0,
			fl:     &Uint64Flag{Name: "seconds", Sources: EnvVars("SECONDS")},
			errContains: `could not parse "1.2" as uint64 value from environment variable ` +
				`"SECONDS" for flag seconds:`,
		},
		{
			name:   "Uint64Flag invalid",
			input:  "foobar",
			output: 0,
			fl:     &Uint64Flag{Name: "seconds", Sources: EnvVars("SECONDS")},
			errContains: `could not parse "foobar" as uint64 value from environment variable ` +
				`"SECONDS" for flag seconds:`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			r.Implements((*DocGenerationFlag)(nil), tc.fl)
			f := tc.fl.(DocGenerationFlag)

			envVarSlice := f.GetEnvVars()
			t.Setenv(envVarSlice[0], tc.input)

			cmd := &Command{
				Flags: []Flag{tc.fl},
				Action: func(_ context.Context, cmd *Command) error {
					r.Equal(tc.output, cmd.Value(tc.fl.Names()[0]))
					r.True(tc.fl.IsSet())
					r.Equal(tc.fl.Names(), cmd.FlagNames())

					return nil
				},
			}

			err := cmd.Run(buildTestContext(t), []string{"run"})

			if tc.errContains != "" {
				r.NotNil(err)
				r.ErrorContains(err, tc.errContains)

				return
			}

			r.NoError(err)
		})
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
			expected: "--vividly\t",
		},
		{
			name:     "bool-flag-with-default-text",
			fl:       &BoolFlag{Name: "wildly", DefaultText: "scrambled"},
			expected: "--wildly\t(default: scrambled)",
		},
		{
			name:     "bool-inv-flag",
			fl:       &BoolWithInverseFlag{Name: "vividly"},
			expected: "--vividly, --no-vividly\t(default: false)",
		},
		{
			name:     "duration-flag",
			fl:       &DurationFlag{Name: "scream-for"},
			expected: "--scream-for duration\t(default: 0s)",
		},
		{
			name:     "duration-flag-with-default-text",
			fl:       &DurationFlag{Name: "feels-about", DefaultText: "whimsically"},
			expected: "--feels-about duration\t(default: whimsically)",
		},
		{
			name:     "float64-flag",
			fl:       &FloatFlag{Name: "arduous"},
			expected: "--arduous float\t(default: 0)",
		},
		{
			name:     "float64-flag-with-default-text",
			fl:       &FloatFlag{Name: "filibuster", DefaultText: "42"},
			expected: "--filibuster float\t(default: 42)",
		},
		{
			name:     "float64-slice-flag",
			fl:       &FloatSliceFlag{Name: "pizzas"},
			expected: "--pizzas float [ --pizzas float ]\t",
		},
		{
			name:     "float64-slice-flag-with-default-text",
			fl:       &FloatSliceFlag{Name: "pepperonis", DefaultText: "shaved"},
			expected: "--pepperonis float [ --pepperonis float ]\t(default: shaved)",
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
			fl:       &Int64Flag{Name: "grubs"},
			expected: "--grubs int\t(default: 0)",
		},
		{
			name:     "int-flag-with-default-text",
			fl:       &Int64Flag{Name: "poisons", DefaultText: "11ty"},
			expected: "--poisons int\t(default: 11ty)",
		},
		{
			name:     "int-slice-flag",
			fl:       &Int64SliceFlag{Name: "pencils"},
			expected: "--pencils int [ --pencils int ]\t",
		},
		{
			name:     "int-slice-flag-with-default-text",
			fl:       &Int64Flag{Name: "pens", DefaultText: "-19"},
			expected: "--pens int\t(default: -19)",
		},
		{
			name:     "uint-slice-flag",
			fl:       &Uint64SliceFlag{Name: "pencils"},
			expected: "--pencils uint [ --pencils uint ]\t",
		},
		{
			name:     "uint-slice-flag-with-default-text",
			fl:       &Uint64Flag{Name: "pens", DefaultText: "29"},
			expected: "--pens uint\t(default: 29)",
		},
		{
			name:     "int64-flag",
			fl:       &Int64Flag{Name: "flume"},
			expected: "--flume int\t(default: 0)",
		},
		{
			name:     "int64-flag-with-default-text",
			fl:       &Int64Flag{Name: "shattering", DefaultText: "22"},
			expected: "--shattering int\t(default: 22)",
		},
		{
			name:     "uint64-slice-flag",
			fl:       &Uint64SliceFlag{Name: "drawers"},
			expected: "--drawers uint [ --drawers uint ]\t",
		},
		{
			name:     "uint64-slice-flag-with-default-text",
			fl:       &Uint64SliceFlag{Name: "handles", DefaultText: "-2"},
			expected: "--handles uint [ --handles uint ]\t(default: -2)",
		},
		{
			name:     "string-flag",
			fl:       &StringFlag{Name: "arf-sound"},
			expected: "--arf-sound string\t",
		},
		{
			name:     "string-flag-with-default-text",
			fl:       &StringFlag{Name: "woof-sound", DefaultText: "urp"},
			expected: "--woof-sound string\t(default: urp)",
		},
		{
			name:     "string-slice-flag",
			fl:       &StringSliceFlag{Name: "meow-sounds"},
			expected: "--meow-sounds string [ --meow-sounds string ]\t",
		},
		{
			name:     "string-slice-flag-with-default-text",
			fl:       &StringSliceFlag{Name: "moo-sounds", DefaultText: "awoo"},
			expected: "--moo-sounds string [ --moo-sounds string ]\t(default: awoo)",
		},
		{
			name:     "timestamp-flag",
			fl:       &TimestampFlag{Name: "eating"},
			expected: "--eating time\t",
		},
		{
			name:     "timestamp-flag-with-default-text",
			fl:       &TimestampFlag{Name: "sleeping", DefaultText: "earlier"},
			expected: "--sleeping time\t(default: earlier)",
		},
		{
			name:     "uint-flag",
			fl:       &Uint64Flag{Name: "jars"},
			expected: "--jars uint\t(default: 0)",
		},
		{
			name:     "uint-flag-with-default-text",
			fl:       &Uint64Flag{Name: "bottles", DefaultText: "99"},
			expected: "--bottles uint\t(default: 99)",
		},
		{
			name:     "uint64-flag",
			fl:       &Uint64Flag{Name: "cans"},
			expected: "--cans uint\t(default: 0)",
		},
		{
			name:     "uint64-flag-with-default-text",
			fl:       &Uint64Flag{Name: "tubes", DefaultText: "13"},
			expected: "--tubes uint\t(default: 13)",
		},
		{
			name:     "nodoc-flag",
			fl:       &nodocFlag{Name: "scarecrow"},
			expected: "",
		},
	} {
		t.Run(tc.name, func(ct *testing.T) {
			s := stringifyFlag(tc.fl)
			assert.Equal(t, tc.expected, s, "stringified flag %q does not match expected", s)
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
	{"foo", nil, "", "", "--foo string\t"},
	{"f", nil, "", "", "-f string\t"},
	{"f", nil, "The total `foo` desired", "all", "-f foo\tThe total foo desired (default: \"all\")"},
	{"test", nil, "", "Something", "--test string\t(default: \"Something\")"},
	{"config", []string{"c"}, "Load configuration from `FILE`", "", "--config FILE, -c FILE\tLoad configuration from FILE"},
	{"config", []string{"c"}, "Load configuration from `CONFIG`", "config.json", "--config CONFIG, -c CONFIG\tLoad configuration from CONFIG (default: \"config.json\")"},
}

func TestStringFlagHelpOutput(t *testing.T) {
	for _, test := range stringFlagTests {
		fl := &StringFlag{Name: test.name, Aliases: test.aliases, Usage: test.usage, Value: test.value}
		assert.Equal(t, test.expected, fl.String())
	}
}

func TestStringFlagDefaultText(t *testing.T) {
	fl := &StringFlag{Name: "foo", Aliases: nil, Usage: "amount of `foo` requested", Value: "none", DefaultText: "all of it"}
	expected := "--foo foo\tamount of foo requested (default: all of it)"
	assert.Equal(t, expected, fl.String())
}

func TestStringFlagWithEnvVarHelpOutput(t *testing.T) {
	t.Setenv("APP_FOO", "derp")

	for _, test := range stringFlagTests {
		fl := &StringFlag{Name: test.name, Aliases: test.aliases, Value: test.value, Sources: EnvVars("APP_FOO")}
		output := fl.String()

		expectedSuffix := withEnvHint([]string{"APP_FOO"}, "")
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

var _ = []struct {
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
	cmd := &Command{
		Flags: []Flag{
			&StringFlag{Name: "hay", Aliases: []string{"H", "hayyy"}, Destination: &v},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"", "--hay", "u", "-H", "yuu", "--hayyy", "YUUUU"})
	assert.NoError(t, err)
	assert.Equal(t, "YUUUU", v)
}

func TestStringFlagValueFromCommand(t *testing.T) {
	f := &StringFlag{Name: "myflag"}
	cmd := &Command{
		Flags: []Flag{
			f,
		},
	}
	require.NoError(t, cmd.Set("myflag", "foobar"))
	require.Equal(t, "foobar", cmd.String(f.Name))
}

var _ = []struct {
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

// func TestFlagEnvHinter(t *testing.T) {
//	defer func() {
//		FlagEnvHinter = withEnvHint
//	}()
//
//	for _, test := range envHintFlagTests {
//		FlagEnvHinter = test.hinter
//		fl := StringFlag{Name: test.name, Sources: ValueSources{test.env}}
//		output := fl.String()
//		if output != test.expected {
//			t.Errorf("%q does not match %q", output, test.expected)
//		}
//	}
// }

var stringSliceFlagTests = []struct {
	name     string
	aliases  []string
	value    []string
	expected string
}{
	{"foo", nil, []string{}, "--foo string [ --foo string ]\t"},
	{"f", nil, []string{}, "-f string [ -f string ]\t"},
	{"f", nil, []string{"Lipstick"}, "-f string [ -f string ]\t(default: \"Lipstick\")"},
	{"test", nil, []string{"Something"}, "--test string [ --test string ]\t(default: \"Something\")"},
	{"dee", []string{"d"}, []string{"Inka", "Dinka", "dooo"}, "--dee string, -d string [ --dee string, -d string ]\t(default: \"Inka\", \"Dinka\", \"dooo\")"},
}

func TestStringSliceFlagHelpOutput(t *testing.T) {
	for _, test := range stringSliceFlagTests {
		f := &StringSliceFlag{Name: test.name, Aliases: test.aliases, Value: test.value}
		assert.Equal(t, test.expected, f.String())
	}
}

func TestStringSliceFlagWithEnvVarHelpOutput(t *testing.T) {
	t.Setenv("APP_QWWX", "11,4")

	for _, test := range stringSliceFlagTests {
		fl := &StringSliceFlag{Name: test.name, Aliases: test.aliases, Value: test.value, Sources: EnvVars("APP_QWWX")}
		output := fl.String()

		expectedSuffix := withEnvHint([]string{"APP_QWWX"}, "")
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

func TestStringSliceFlagApply_SetsAllNames(t *testing.T) {
	fl := &StringSliceFlag{Name: "goat", Aliases: []string{"G", "gooots"}}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"", "--goat", "aaa", "-G", "bbb", "--gooots", "eeeee"})
	assert.NoError(t, err)
}

func TestStringSliceFlagApply_UsesEnvValues_noDefault(t *testing.T) {
	t.Setenv("MY_GOAT", "vincent van goat,scape goat")
	fl := &StringSliceFlag{Name: "goat", Sources: EnvVars("MY_GOAT")}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}

	err := cmd.Run(buildTestContext(t), []string{""})
	assert.NoError(t, err)
	assert.Equal(t, []string{"vincent van goat", "scape goat"}, cmd.Value("goat"))
}

func TestStringSliceFlagApply_UsesEnvValues_withDefault(t *testing.T) {
	t.Setenv("MY_GOAT", "vincent van goat,scape goat")
	val := []string{`some default`, `values here`}
	fl := &StringSliceFlag{Name: "goat", Sources: EnvVars("MY_GOAT"), Value: val}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}
	err := cmd.Run(buildTestContext(t), []string{""})
	assert.NoError(t, err)
	assert.Equal(t, []string{"vincent van goat", "scape goat"}, cmd.Value("goat"))
}

func TestStringSliceFlagApply_DefaultValueWithDestination(t *testing.T) {
	defValue := []string{"UA", "US"}
	dest := []string{"CA"}

	fl := StringSliceFlag{Name: "country", Value: defValue, Destination: &dest}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	err := cmd.Run(buildTestContext(t), []string{""})
	assert.NoError(t, err)
	assert.Equal(t, defValue, dest)
}

func TestStringSliceFlagValueFromCommand(t *testing.T) {
	f := &StringSliceFlag{Name: "myflag"}
	cmd := &Command{
		Flags: []Flag{
			f,
		},
	}

	require.NoError(t, cmd.Set("myflag", "a"))
	require.NoError(t, cmd.Set("myflag", "b"))
	require.NoError(t, cmd.Set("myflag", "c"))
	require.Equal(t, []string{"a", "b", "c"}, cmd.StringSlice(f.Name))
}

var intFlagTests = []struct {
	name     string
	expected string
}{
	{"hats", "--hats int\t(default: 9)"},
	{"H", "-H int\t(default: 9)"},
}

func TestIntFlagHelpOutput(t *testing.T) {
	for _, test := range intFlagTests {
		fl := &Int64Flag{Name: test.name, Value: 9}
		assert.Equal(t, test.expected, fl.String())
	}
}

func TestIntFlagWithEnvVarHelpOutput(t *testing.T) {
	t.Setenv("APP_BAR", "2")

	for _, test := range intFlagTests {
		fl := &Int64Flag{Name: test.name, Sources: EnvVars("APP_BAR")}
		output := fl.String()

		expectedSuffix := withEnvHint([]string{"APP_BAR"}, "")
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

func TestIntFlagApply_SetsAllNames(t *testing.T) {
	v := int64(3)
	cmd := &Command{
		Flags: []Flag{
			&Int64Flag{Name: "banana", Aliases: []string{"B", "banannanana"}, Destination: &v},
		},
	}
	r := require.New(t)
	r.NoError(cmd.Run(buildTestContext(t), []string{"", "--banana", "1", "-B", "2", "--banannanana", "5"}))
	r.Equal(int64(5), v)
}

func TestIntFlagValueFromCommand(t *testing.T) {
	fl := &Int64Flag{Name: "myflag"}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}
	require.NoError(t, cmd.Set("myflag", "42"))
	require.Equal(t, int64(42), cmd.Int64(fl.Name))
}

var uintFlagTests = []struct {
	name     string
	expected string
}{
	{"nerfs", "--nerfs uint\t(default: 41)"},
	{"N", "-N uint\t(default: 41)"},
}

func TestUintFlagHelpOutput(t *testing.T) {
	for _, test := range uintFlagTests {
		fl := &Uint64Flag{Name: test.name, Value: 41}
		assert.Equal(t, test.expected, fl.String())
	}
}

func TestUintFlagWithEnvVarHelpOutput(t *testing.T) {
	t.Setenv("APP_BAR", "2")

	for _, test := range uintFlagTests {
		fl := &Uint64Flag{Name: test.name, Sources: EnvVars("APP_BAR")}
		output := fl.String()

		expectedSuffix := withEnvHint([]string{"APP_BAR"}, "")
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

func TestUintFlagValueFromCommand(t *testing.T) {
	fl := &Uint64Flag{Name: "myflag"}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}
	require.NoError(t, cmd.Set("myflag", "42"))
	require.Equal(t, uint64(42), cmd.Uint64(fl.Name))
}

var uint64FlagTests = []struct {
	name     string
	expected string
}{
	{"gerfs", "--gerfs uint\t(default: 8589934582)"},
	{"G", "-G uint\t(default: 8589934582)"},
}

func TestUint64FlagHelpOutput(t *testing.T) {
	for _, test := range uint64FlagTests {
		fl := Uint64Flag{Name: test.name, Value: 8589934582}
		assert.Equal(t, test.expected, fl.String())
	}
}

func TestUint64FlagWithEnvVarHelpOutput(t *testing.T) {
	t.Setenv("APP_BAR", "2")

	for _, test := range uint64FlagTests {
		fl := &Uint64Flag{Name: test.name, Sources: EnvVars("APP_BAR")}
		output := fl.String()

		expectedSuffix := withEnvHint([]string{"APP_BAR"}, "")
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

func TestUint64FlagValueFromCommand(t *testing.T) {
	f := &Uint64Flag{Name: "myflag"}
	cmd := &Command{
		Flags: []Flag{
			f,
		},
	}
	require.NoError(t, cmd.Set("myflag", "42"))
	require.Equal(t, uint64(42), cmd.Uint64(f.Name))
}

var durationFlagTests = []struct {
	name     string
	expected string
}{
	{"hooting", "--hooting duration\t(default: 1s)"},
	{"H", "-H duration\t(default: 1s)"},
}

func TestDurationFlagHelpOutput(t *testing.T) {
	for _, test := range durationFlagTests {
		fl := &DurationFlag{Name: test.name, Value: 1 * time.Second}
		assert.Equal(t, test.expected, fl.String())
	}
}

func TestDurationFlagWithEnvVarHelpOutput(t *testing.T) {
	t.Setenv("APP_BAR", "2h3m6s")

	for _, test := range durationFlagTests {
		fl := &DurationFlag{Name: test.name, Sources: EnvVars("APP_BAR")}
		output := fl.String()

		expectedSuffix := withEnvHint([]string{"APP_BAR"}, "")
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

func TestDurationFlagApply_SetsAllNames(t *testing.T) {
	v := time.Second * 20
	cmd := &Command{
		Flags: []Flag{
			&DurationFlag{Name: "howmuch", Aliases: []string{"H", "whyyy"}, Destination: &v},
		},
	}
	err := cmd.Run(buildTestContext(t), []string{"", "--howmuch", "30s", "-H", "5m", "--whyyy", "30h"})
	assert.NoError(t, err)
	assert.Equal(t, time.Hour*30, v)
}

func TestDurationFlagValueFromCommand(t *testing.T) {
	f := &DurationFlag{Name: "myflag"}
	cmd := &Command{
		Flags: []Flag{
			f,
		},
	}
	require.NoError(t, cmd.Set("myflag", "42s"))
	require.Equal(t, 42*time.Second, cmd.Duration(f.Name))
}

var intSliceFlagTests = []struct {
	name     string
	aliases  []string
	value    []int64
	expected string
}{
	{"heads", nil, []int64{}, "--heads int [ --heads int ]\t"},
	{"H", nil, []int64{}, "-H int [ -H int ]\t"},
	{"H", []string{"heads"}, []int64{9, 3}, "-H int, --heads int [ -H int, --heads int ]\t(default: 9, 3)"},
}

func TestIntSliceFlagHelpOutput(t *testing.T) {
	for _, test := range intSliceFlagTests {
		fl := &Int64SliceFlag{Name: test.name, Aliases: test.aliases, Value: test.value}
		assert.Equal(t, test.expected, fl.String())
	}
}

func TestIntSliceFlagWithEnvVarHelpOutput(t *testing.T) {
	t.Setenv("APP_SMURF", "42,3")

	for _, test := range intSliceFlagTests {
		fl := &Int64SliceFlag{Name: test.name, Aliases: test.aliases, Value: test.value, Sources: EnvVars("APP_SMURF")}
		output := fl.String()

		expectedSuffix := withEnvHint([]string{"APP_SMURF"}, "")
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

func TestIntSliceFlagApply_SetsAllNames(t *testing.T) {
	fl := &Int64SliceFlag{Name: "bits", Aliases: []string{"B", "bips"}}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}
	err := cmd.Run(buildTestContext(t), []string{"", "--bits", "23", "-B", "3", "--bips", "99"})
	assert.NoError(t, err)
}

func TestIntSliceFlagApply_UsesEnvValues_noDefault(t *testing.T) {
	t.Setenv("MY_GOAT", "1 , 2")

	fl := &Int64SliceFlag{Name: "goat", Sources: EnvVars("MY_GOAT")}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}
	r := require.New(t)
	r.NoError(cmd.Run(buildTestContext(t), []string{""}))
	r.NoError(fl.PostParse())
	r.Equal([]int64{1, 2}, cmd.Value("goat"))
}

func TestIntSliceFlagApply_UsesEnvValues_withDefault(t *testing.T) {
	t.Setenv("MY_GOAT", "1 , 2")
	val := []int64{3, 4}
	fl := &Int64SliceFlag{Name: "goat", Sources: EnvVars("MY_GOAT"), Value: val}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}
	r := require.New(t)
	r.NoError(cmd.Run(buildTestContext(t), []string{""}))
	r.Equal([]int64{3, 4}, val)
	r.Equal([]int64{1, 2}, cmd.Value("goat"))
}

func TestIntSliceFlagApply_DefaultValueWithDestination(t *testing.T) {
	defValue := []int64{1, 2}
	dest := []int64{3}

	fl := &Int64SliceFlag{Name: "country", Value: defValue, Destination: &dest}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}
	err := cmd.Run(buildTestContext(t), []string{""})
	assert.NoError(t, err)
	assert.Equal(t, defValue, dest)
}

func TestIntSliceFlagApply_ParentContext(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&Int64SliceFlag{Name: "numbers", Aliases: []string{"n"}, Value: []int64{1, 2, 3}},
		},
		Commands: []*Command{
			{
				Name: "child",
				Action: func(_ context.Context, cmd *Command) error {
					require.Equalf(t, []int64{1, 2, 3}, cmd.Int64Slice("numbers"), "child context unable to view parent flag")

					return nil
				},
			},
		},
	}).Run(buildTestContext(t), []string{"run", "child"})
}

func TestIntSliceFlag_SetFromParentCommand(t *testing.T) {
	fl := &Int64SliceFlag{Name: "numbers", Aliases: []string{"n"}, Value: []int64{1, 2, 3, 4}}

	cmd := &Command{
		parent: &Command{
			Flags: []Flag{
				fl,
			},
		},
	}

	require.Equalf(t, []int64{1, 2, 3, 4}, cmd.Int64Slice("numbers"), "child context unable to view parent flag")
}

func TestIntSliceFlagValueFromCommand(t *testing.T) {
	f := &Int64SliceFlag{Name: "myflag"}
	cmd := &Command{
		Flags: []Flag{
			f,
		},
	}
	require.NoError(t, cmd.Set("myflag", "1"))
	require.NoError(t, cmd.Set("myflag", "2"))
	require.NoError(t, cmd.Set("myflag", "3"))
	require.Equal(t, []int64{1, 2, 3}, cmd.Int64Slice(f.Name))
}

var uintSliceFlagTests = []struct {
	name     string
	aliases  []string
	value    []uint64
	expected string
}{
	{"heads", nil, []uint64{}, "--heads uint [ --heads uint ]\t"},
	{"H", nil, []uint64{}, "-H uint [ -H uint ]\t"},
	{
		"heads",
		[]string{"H"},
		[]uint64{2, 17179869184},
		"--heads uint, -H uint [ --heads uint, -H uint ]\t(default: 2, 17179869184)",
	},
}

func TestUintSliceFlagHelpOutput(t *testing.T) {
	for _, test := range uintSliceFlagTests {
		t.Run(test.name, func(t *testing.T) {
			fl := &Uint64SliceFlag{Name: test.name, Aliases: test.aliases, Value: test.value}
			require.Equal(t, test.expected, fl.String())
		})
	}
}

func TestUintSliceFlagWithEnvVarHelpOutput(t *testing.T) {
	t.Setenv("APP_SMURF", "42,17179869184")

	for _, test := range uintSliceFlagTests {
		fl := &Uint64SliceFlag{Name: test.name, Value: test.value, Sources: EnvVars("APP_SMURF")}
		output := fl.String()

		expectedSuffix := withEnvHint([]string{"APP_SMURF"}, "")
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

func TestUintSliceFlagApply_SetsAllNames(t *testing.T) {
	fl := &Uint64SliceFlag{Name: "bits", Aliases: []string{"B", "bips"}}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}
	err := cmd.Run(buildTestContext(t), []string{"", "--bits", "23", "-B", "3", "--bips", "99"})
	assert.NoError(t, err)
}

func TestUintSliceFlagApply_UsesEnvValues_noDefault(t *testing.T) {
	t.Setenv("MY_GOAT", "1 , 2")

	fl := &Uint64SliceFlag{Name: "goat", Sources: EnvVars("MY_GOAT")}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}
	r := require.New(t)
	r.NoError(cmd.Run(buildTestContext(t), []string{""}))
	r.Equal([]uint64{1, 2}, cmd.Value("goat"))
}

func TestUintSliceFlagApply_UsesEnvValues_withDefault(t *testing.T) {
	t.Setenv("MY_GOAT", "1 , 2")
	val := NewUint64Slice(3, 4)
	fl := &Uint64SliceFlag{Name: "goat", Sources: EnvVars("MY_GOAT"), Value: val.Value()}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}
	r := require.New(t)
	r.NoError(cmd.Run(buildTestContext(t), []string{""}))
	r.Equal([]uint64{3, 4}, val.Value())
	r.Equal([]uint64{1, 2}, cmd.Value("goat"))
}

func TestUintSliceFlagApply_DefaultValueWithDestination(t *testing.T) {
	defValue := []uint64{1, 2}
	var dest []uint64

	fl := &Uint64SliceFlag{Name: "country", Value: defValue, Destination: &dest}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}
	err := cmd.Run(buildTestContext(t), []string{""})
	assert.NoError(t, err)
	assert.Equal(t, defValue, dest)
}

func TestUint64SliceFlagApply_ParentContext(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&Uint64SliceFlag{Name: "numbers", Aliases: []string{"n"}, Value: []uint64{1, 2, 3}},
		},
		Commands: []*Command{
			{
				Name: "child",
				Action: func(_ context.Context, cmd *Command) error {
					require.Equalf(
						t, []uint64{1, 2, 3}, cmd.Uint64Slice("numbers"),
						"child context unable to view parent flag",
					)
					return nil
				},
			},
		},
	}).Run(buildTestContext(t), []string{"run", "child"})
}

func TestUintSliceFlag_SetFromParentCommand(t *testing.T) {
	fl := &UintSliceFlag{Name: "numbers", Aliases: []string{"n"}, Value: []uint{1, 2, 3, 4}}
	cmd := &Command{
		parent: &Command{
			Flags: []Flag{
				fl,
			},
		},
	}

	r := require.New(t)
	r.Equalf(
		[]uint{1, 2, 3, 4},
		cmd.UintSlice("numbers"),
		"child context unable to view parent flag",
	)
}

func TestUintSliceFlag_ReturnNil(t *testing.T) {
	fl := &Uint64SliceFlag{}

	r := require.New(t)
	cmd := &Command{
		parent: &Command{
			Flags: []Flag{
				fl,
			},
		},
	}
	r.Equalf(
		[]uint64(nil),
		cmd.Uint64Slice("numbers"),
		"child context unable to view parent flag",
	)
}

var uint64SliceFlagTests = []struct {
	name     string
	aliases  []string
	value    []uint64
	expected string
}{
	{"heads", nil, []uint64{}, "--heads uint [ --heads uint ]\t"},
	{"H", nil, []uint64{}, "-H uint [ -H uint ]\t"},
	{
		"heads",
		[]string{"H"},
		[]uint64{2, 17179869184},
		"--heads uint, -H uint [ --heads uint, -H uint ]\t(default: 2, 17179869184)",
	},
}

func TestUint64SliceFlagHelpOutput(t *testing.T) {
	for _, test := range uint64SliceFlagTests {
		fl := Uint64SliceFlag{Name: test.name, Aliases: test.aliases, Value: test.value}
		assert.Equal(t, test.expected, fl.String())
	}
}

func TestUint64SliceFlagWithEnvVarHelpOutput(t *testing.T) {
	t.Setenv("APP_SMURF", "42,17179869184")

	for _, test := range uint64SliceFlagTests {
		fl := Uint64SliceFlag{Name: test.name, Value: test.value, Sources: EnvVars("APP_SMURF")}
		output := fl.String()

		expectedSuffix := withEnvHint([]string{"APP_SMURF"}, "")
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

func TestUint64SliceFlagApply_SetsAllNames(t *testing.T) {
	fl := Uint64SliceFlag{Name: "bits", Aliases: []string{"B", "bips"}}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}
	err := cmd.Run(buildTestContext(t), []string{"", "--bits", "23", "-B", "3", "--bips", "99"})
	assert.NoError(t, err)
}

func TestUint64SliceFlagApply_UsesEnvValues_noDefault(t *testing.T) {
	t.Setenv("MY_GOAT", "1 , 2")
	fl := Uint64SliceFlag{Name: "goat", Sources: EnvVars("MY_GOAT")}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}
	assert.NoError(t, cmd.Run(buildTestContext(t), []string{""}))
	assert.Equal(t, []uint64{1, 2}, cmd.Value("goat"))
}

func TestUint64SliceFlagApply_UsesEnvValues_withDefault(t *testing.T) {
	t.Setenv("MY_GOAT", "1 , 2")
	val := []uint64{3, 4}
	fl := Uint64SliceFlag{Name: "goat", Sources: EnvVars("MY_GOAT"), Value: val}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}
	assert.NoError(t, cmd.Run(buildTestContext(t), []string{""}))
	assert.Equal(t, []uint64{1, 2}, cmd.Value("goat"))
}

func TestUint64SliceFlagApply_DefaultValueWithDestination(t *testing.T) {
	defValue := []uint64{1, 2}
	dest := []uint64{3}

	fl := Uint64SliceFlag{Name: "country", Value: defValue, Destination: &dest}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	err := cmd.Run(buildTestContext(t), []string{""})
	assert.NoError(t, err)

	assert.Equal(t, defValue, dest)
}

func TestUint64SliceFlagApply_ParentCommand(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&Uint64SliceFlag{Name: "numbers", Aliases: []string{"n"}, Value: []uint64{1, 2, 3}},
		},
		Commands: []*Command{
			{
				Name: "child",
				Action: func(_ context.Context, cmd *Command) error {
					require.Equalf(
						t, []uint64{1, 2, 3}, cmd.Uint64Slice("numbers"),
						"child context unable to view parent flag",
					)
					return nil
				},
			},
		},
	}).Run(buildTestContext(t), []string{"run", "child"})
}

func TestUint64SliceFlag_SetFromParentCommand(t *testing.T) {
	fl := &Uint64SliceFlag{Name: "numbers", Aliases: []string{"n"}, Value: []uint64{1, 2, 3, 4}}
	cmd := &Command{
		parent: &Command{
			Flags: []Flag{
				fl,
			},
		},
	}
	r := require.New(t)
	r.Equalf(
		[]uint64{1, 2, 3, 4}, cmd.Uint64Slice("numbers"),
		"child context unable to view parent flag",
	)
}

func TestUint64SliceFlag_ReturnNil(t *testing.T) {
	fl := &Uint64SliceFlag{}
	cmd := &Command{
		parent: &Command{
			Flags: []Flag{
				fl,
			},
		},
	}
	r := require.New(t)
	r.Equalf(
		[]uint64(nil), cmd.Uint64Slice("numbers"),
		"child context unable to view parent flag",
	)
}

var float64FlagTests = []struct {
	name     string
	expected string
}{
	{"hooting", "--hooting float\t(default: 0.1)"},
	{"H", "-H float\t(default: 0.1)"},
}

func TestFloat64FlagHelpOutput(t *testing.T) {
	for _, test := range float64FlagTests {
		f := &FloatFlag{Name: test.name, Value: 0.1}
		assert.Equal(t, test.expected, f.String())
	}
}

func TestFloat64FlagWithEnvVarHelpOutput(t *testing.T) {
	t.Setenv("APP_BAZ", "99.4")

	for _, test := range float64FlagTests {
		fl := &FloatFlag{Name: test.name, Sources: EnvVars("APP_BAZ")}
		output := fl.String()

		expectedSuffix := withEnvHint([]string{"APP_BAZ"}, "")
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

func TestFloat64FlagApply_SetsAllNames(t *testing.T) {
	v := 99.1
	fl := FloatFlag{Name: "noodles", Aliases: []string{"N", "nurbles"}, Destination: &v}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{"", "--noodles", "1.3", "-N", "11", "--nurbles", "43.33333"}))
	assert.Equal(t, float64(43.33333), v)
}

func TestFloat64FlagValueFromCommand(t *testing.T) {
	fl := &FloatFlag{Name: "myflag"}
	cmd := &Command{
		Flags: []Flag{
			fl,
		},
	}
	require.NoError(t, cmd.Set("myflag", "1.23"))
	require.Equal(t, 1.23, cmd.Float(fl.Name))
}

var float64SliceFlagTests = []struct {
	name     string
	aliases  []string
	value    []float64
	expected string
}{
	{"heads", nil, []float64{}, "--heads float [ --heads float ]\t"},
	{"H", nil, []float64{}, "-H float [ -H float ]\t"},
	{
		"heads",
		[]string{"H"},
		[]float64{0.1234, -10.5},
		"--heads float, -H float [ --heads float, -H float ]\t(default: 0.1234, -10.5)",
	},
}

func TestFloat64SliceFlagHelpOutput(t *testing.T) {
	for _, test := range float64SliceFlagTests {
		fl := FloatSliceFlag{Name: test.name, Aliases: test.aliases, Value: test.value}
		assert.Equal(t, test.expected, fl.String())
	}
}

func TestFloat64SliceFlagWithEnvVarHelpOutput(t *testing.T) {
	t.Setenv("APP_SMURF", "0.1234,-10.5")
	for _, test := range float64SliceFlagTests {
		fl := FloatSliceFlag{Name: test.name, Value: test.value, Sources: EnvVars("APP_SMURF")}
		output := fl.String()

		expectedSuffix := withEnvHint([]string{"APP_SMURF"}, "")
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

func TestFloat64SliceFlagApply_SetsAllNames(t *testing.T) {
	fl := FloatSliceFlag{Name: "bits", Aliases: []string{"B", "bips"}}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{"", "--bits", "23", "-B", "3", "--bips", "99"}))
}

func TestFloat64SliceFlagApply_UsesEnvValues_noDefault(t *testing.T) {
	t.Setenv("MY_GOAT", "1.0 , 2.0")
	fl := FloatSliceFlag{Name: "goat", Sources: EnvVars("MY_GOAT")}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{""}))
	assert.Equal(t, []float64{1, 2}, cmd.Value("goat"))
}

func TestFloat64SliceFlagApply_UsesEnvValues_withDefault(t *testing.T) {
	t.Setenv("MY_GOAT", "1.0 , 2.0")
	val := []float64{3.0, 4.0}
	fl := FloatSliceFlag{Name: "goat", Sources: EnvVars("MY_GOAT"), Value: val}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{""}))
	assert.Equal(t, []float64{1, 2}, cmd.Value("goat"))
}

func TestFloat64SliceFlagApply_DefaultValueWithDestination(t *testing.T) {
	defValue := []float64{1.0, 2.0}
	dest := []float64{3}

	fl := FloatSliceFlag{Name: "country", Value: defValue, Destination: &dest}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{""}))
	assert.Equal(t, defValue, dest)
}

func TestFloat64SliceFlagValueFromCommand(t *testing.T) {
	fl := FloatSliceFlag{Name: "myflag"}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}
	assert.NoError(t, cmd.Run(buildTestContext(t), []string{""}))
	require.NoError(t, cmd.Set("myflag", "1.23"))
	require.NoError(t, cmd.Set("myflag", "4.56"))
	require.Equal(t, []float64{1.23, 4.56}, cmd.FloatSlice(fl.Name))
}

func TestFloat64SliceFlagApply_ParentCommand(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&FloatSliceFlag{Name: "numbers", Aliases: []string{"n"}, Value: []float64{1.0, 2.0, 3.0}},
		},
		Commands: []*Command{
			{
				Name: "child",
				Action: func(_ context.Context, cmd *Command) error {
					require.Equalf(t, []float64{1.0, 2.0, 3.0}, cmd.FloatSlice("numbers"), "child context unable to view parent flag")
					return nil
				},
			},
		},
	}).Run(buildTestContext(t), []string{"run", "child"})
}

var genericFlagTests = []struct {
	name     string
	value    Value
	expected string
}{
	{"toads", &Parser{"abc", "def"}, "--toads value\ttest flag (default: abc,def)"},
	{"t", &Parser{"abc", "def"}, "-t value\ttest flag (default: abc,def)"},
}

func TestGenericFlagHelpOutput(t *testing.T) {
	for _, test := range genericFlagTests {
		fl := &GenericFlag{Name: test.name, Value: test.value, Usage: "test flag"}
		assert.Equal(t, test.expected, fl.String())
	}
}

func TestGenericFlagWithEnvVarHelpOutput(t *testing.T) {
	t.Setenv("APP_ZAP", "3")

	for _, test := range genericFlagTests {
		fl := &GenericFlag{Name: test.name, Sources: EnvVars("APP_ZAP")}
		output := fl.String()

		expectedSuffix := withEnvHint([]string{"APP_ZAP"}, "")
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

func TestGenericFlagApply_SetsAllNames(t *testing.T) {
	fl := GenericFlag{Name: "orbs", Aliases: []string{"O", "obrs"}, Value: &Parser{}}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{"", "--orbs", "eleventy,3", "-O", "4,bloop", "--obrs", "19,s"}))
}

func TestGenericFlagValueFromCommand(t *testing.T) {
	cmd := &Command{
		Name: "foo",
		Flags: []Flag{
			&GenericFlag{Name: "myflag", Value: &Parser{}},
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "--myflag", "abc,def"}))
	assert.Equal(t, &Parser{"abc", "def"}, cmd.Generic("myflag"))
	assert.Nil(t, cmd.Generic("someother"))
}

func TestParseGenericFromEnv(t *testing.T) {
	t.Setenv("APP_SERVE", "20,30")
	cmd := &Command{
		Flags: []Flag{
			&GenericFlag{
				Name:    "serve",
				Aliases: []string{"s"},
				Value:   &Parser{},
				Sources: EnvVars("APP_SERVE"),
			},
		},
		Action: func(ctx context.Context, cmd *Command) error {
			if !reflect.DeepEqual(cmd.Generic("serve"), &Parser{"20", "30"}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(cmd.Generic("s"), &Parser{"20", "30"}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}
	assert.NoError(t, cmd.Run(buildTestContext(t), []string{"run"}))
}

func TestFlagActionFromEnv(t *testing.T) {
	t.Setenv("X", "42")
	x := 0

	cmd := &Command{
		Flags: []Flag{
			&IntFlag{
				Name:    "x",
				Sources: EnvVars("X"),
				Action: func(ctx context.Context, cmd *Command, v int) error {
					x = v
					return nil
				},
			},
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{"run"}))
	assert.Equal(t, cmd.Int("x"), 42)
	assert.Equal(t, x, 42)
}

func TestParseShortOptionBoolError(t *testing.T) {
	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	cmd.Flags = []Flag{
		&BoolFlag{Name: "debug", Aliases: []string{"d"}},
		&BoolFlag{Name: "verbose", Aliases: []string{"v"}},
	}

	err := cmd.Run(buildTestContext(t), []string{"run", "-vd=notabool"})
	assert.Error(t, err, "expected error parsing invalid bool")
}

func TestParseShortOptionIntError(t *testing.T) {
	cmd := buildMinimalTestCommand()
	cmd.Flags = []Flag{
		&IntFlag{Name: "port", Aliases: []string{"p"}},
		&BoolFlag{Name: "debug", Aliases: []string{"d"}},
	}

	err := cmd.Run(buildTestContext(t), []string{"run", "-dp=notanint"})
	assert.Error(t, err, "expected error parsing invalid int")
}

func TestParseMultiString(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&StringFlag{Name: "serve", Aliases: []string{"s"}},
		},
		Action: func(_ context.Context, cmd *Command) error {
			assert.Equal(t, "10", cmd.String("serve"), "main name not set")
			assert.Equal(t, "10", cmd.String("s"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "-s", "10"})
}

func TestParseDestinationString(t *testing.T) {
	var dest string
	_ = (&Command{
		Flags: []Flag{
			&StringFlag{
				Name:        "dest",
				Destination: &dest,
			},
		},
		Action: func(context.Context, *Command) error {
			assert.Equal(t, "10", dest, "expected destination String 10")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "--dest", "10"})
}

func TestParseMultiStringFromEnv(t *testing.T) {
	t.Setenv("APP_COUNT", "20")

	_ = (&Command{
		Flags: []Flag{
			&StringFlag{Name: "count", Aliases: []string{"c"}, Sources: EnvVars("APP_COUNT")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			assert.Equal(t, "20", cmd.String("count"), "main name not set")
			assert.Equal(t, "20", cmd.String("c"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiStringFromEnvCascade(t *testing.T) {
	t.Setenv("APP_COUNT", "20")

	_ = (&Command{
		Flags: []Flag{
			&StringFlag{Name: "count", Aliases: []string{"c"}, Sources: EnvVars("COMPAT_COUNT", "APP_COUNT")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			assert.Equal(t, "20", cmd.String("count"), "main name not set")
			assert.Equal(t, "20", cmd.String("c"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiStringSlice(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&StringSliceFlag{Name: "serve", Aliases: []string{"s"}, Value: []string{}},
		},
		Action: func(_ context.Context, cmd *Command) error {
			expected := []string{"10", "20"}
			assert.Equal(t, expected, cmd.StringSlice("serve"), "main name not set")
			assert.Equal(t, expected, cmd.StringSlice("s"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiStringSliceWithDefaults(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&StringSliceFlag{Name: "serve", Aliases: []string{"s"}, Value: []string{"9", "2"}},
		},
		Action: func(_ context.Context, cmd *Command) error {
			expected := []string{"10", "20"}
			assert.Equal(t, expected, cmd.StringSlice("serve"), "main name not set")
			assert.Equal(t, expected, cmd.StringSlice("s"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiStringSliceWithDestination(t *testing.T) {
	dest := []string{}

	_ = (&Command{
		Flags: []Flag{
			&StringSliceFlag{Name: "serve", Aliases: []string{"s"}, Destination: &dest},
		},
		Action: func(_ context.Context, cmd *Command) error {
			expected := []string{"10", "20"}
			assert.Equal(t, expected, dest, "destination val not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiStringSliceWithDestinationAndEnv(t *testing.T) {
	t.Setenv("APP_INTERVALS", "20,30,40")

	dest := []string{}
	_ = (&Command{
		Flags: []Flag{
			&StringSliceFlag{Name: "serve", Aliases: []string{"s"}, Destination: &dest, Sources: EnvVars("APP_INTERVALS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			expected := []string{"10", "20"}
			assert.Equal(t, expected, dest, "destination val not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiFloat64SliceWithDestinationAndEnv(t *testing.T) {
	t.Setenv("APP_INTERVALS", "20,30,40")

	dest := []float64{}
	_ = (&Command{
		Flags: []Flag{
			&FloatSliceFlag{Name: "serve", Aliases: []string{"s"}, Destination: &dest, Sources: EnvVars("APP_INTERVALS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			expected := []float64{10, 20}
			assert.Equal(t, expected, dest, "destination val not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiIntSliceWithDestinationAndEnv(t *testing.T) {
	t.Setenv("APP_INTERVALS", "20,30,40")

	dest := []int64{}
	_ = (&Command{
		Flags: []Flag{
			&Int64SliceFlag{Name: "serve", Aliases: []string{"s"}, Destination: &dest, Sources: EnvVars("APP_INTERVALS")},
		},
		Action: func(context.Context, *Command) error {
			require.Equalf(t, []int64{10, 20}, dest, "main name not set")

			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiStringSliceWithDefaultsUnset(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&StringSliceFlag{Name: "serve", Aliases: []string{"s"}, Value: []string{"9", "2"}},
		},
		Action: func(_ context.Context, cmd *Command) error {
			expected := []string{"9", "2"}
			assert.Equal(t, expected, cmd.StringSlice("serve"), "main name not set")
			assert.Equal(t, expected, cmd.StringSlice("s"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiStringSliceFromEnv(t *testing.T) {
	t.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&Command{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: []string{}, Sources: EnvVars("APP_INTERVALS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			expected := []string{"20", "30", "40"}
			assert.Equal(t, expected, cmd.StringSlice("intervals"), "main name not set from env")
			assert.Equal(t, expected, cmd.StringSlice("i"), "short name not set from env")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiStringSliceFromEnvWithDefaults(t *testing.T) {
	t.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&Command{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: []string{"1", "2", "5"}, Sources: EnvVars("APP_INTERVALS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			expected := []string{"20", "30", "40"}
			assert.Equal(t, expected, cmd.StringSlice("intervals"), "main name not set from env")
			assert.Equal(t, expected, cmd.StringSlice("i"), "short name not set from env")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiStringSliceFromEnvCascade(t *testing.T) {
	t.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&Command{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: []string{}, Sources: EnvVars("COMPAT_INTERVALS", "APP_INTERVALS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			expected := []string{"20", "30", "40"}
			assert.Equal(t, expected, cmd.StringSlice("intervals"), "main name not set from env")
			assert.Equal(t, expected, cmd.StringSlice("i"), "short name not set from env")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiStringSliceFromEnvCascadeWithDefaults(t *testing.T) {
	t.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&Command{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: []string{"1", "2", "5"}, Sources: EnvVars("COMPAT_INTERVALS", "APP_INTERVALS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			expected := []string{"20", "30", "40"}
			assert.Equal(t, expected, cmd.StringSlice("intervals"), "main name not set from env")
			assert.Equal(t, expected, cmd.StringSlice("i"), "short name not set from env")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiStringSliceFromEnvWithDestination(t *testing.T) {
	t.Setenv("APP_INTERVALS", "20,30,40")

	dest := []string{}
	_ = (&Command{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Destination: &dest, Sources: EnvVars("APP_INTERVALS")},
		},
		Action: func(context.Context, *Command) error {
			assert.Equal(t, []string{"20", "30", "40"}, dest, "destination value not set from env")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiInt(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&Int64Flag{Name: "serve", Aliases: []string{"s"}},
		},
		Action: func(_ context.Context, cmd *Command) error {
			assert.Equal(t, int64(10), cmd.Int64("serve"), "main name not set")
			assert.Equal(t, int64(10), cmd.Int64("s"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "-s", "10"})
}

func TestParseDestinationInt(t *testing.T) {
	var dest int64
	_ = (&Command{
		Flags: []Flag{
			&Int64Flag{
				Name:        "dest",
				Destination: &dest,
			},
		},
		Action: func(context.Context, *Command) error {
			assert.Equal(t, int64(10), dest, "expected destination Int64 10")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "--dest", "10"})
}

func TestParseMultiIntFromEnv(t *testing.T) {
	t.Setenv("APP_TIMEOUT_SECONDS", "10")
	_ = (&Command{
		Flags: []Flag{
			&Int64Flag{Name: "timeout", Aliases: []string{"t"}, Sources: EnvVars("APP_TIMEOUT_SECONDS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			assert.Equal(t, int64(10), cmd.Int64("timeout"), "main name not set")
			assert.Equal(t, int64(10), cmd.Int64("t"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiIntFromEnvCascade(t *testing.T) {
	t.Setenv("APP_TIMEOUT_SECONDS", "10")
	_ = (&Command{
		Flags: []Flag{
			&Int64Flag{Name: "timeout", Aliases: []string{"t"}, Sources: EnvVars("COMPAT_TIMEOUT_SECONDS", "APP_TIMEOUT_SECONDS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			assert.Equal(t, int64(10), cmd.Int64("timeout"), "main name not set")
			assert.Equal(t, int64(10), cmd.Int64("t"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiIntSlice(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&Int64SliceFlag{Name: "serve", Aliases: []string{"s"}, Value: []int64{}},
		},
		Action: func(_ context.Context, cmd *Command) error {
			r := require.New(t)

			r.Equalf([]int64{10, 20}, cmd.Int64Slice("serve"), "main name not set")
			r.Equalf([]int64{10, 20}, cmd.Int64Slice("s"), "short name not set")

			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiIntSliceWithDefaults(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&Int64SliceFlag{Name: "serve", Aliases: []string{"s"}, Value: []int64{9, 2}},
		},
		Action: func(_ context.Context, cmd *Command) error {
			r := require.New(t)

			r.Equalf([]int64{10, 20}, cmd.Int64Slice("serve"), "main name not set")
			r.Equalf([]int64{10, 20}, cmd.Int64Slice("s"), "short name not set")

			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiIntSliceWithDefaultsUnset(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&Int64SliceFlag{Name: "serve", Aliases: []string{"s"}, Value: []int64{9, 2}},
		},
		Action: func(_ context.Context, cmd *Command) error {
			expected := []int64{9, 2}
			assert.Equal(t, expected, cmd.Int64Slice("serve"), "main name not set")
			assert.Equal(t, expected, cmd.Int64Slice("s"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiIntSliceFromEnv(t *testing.T) {
	t.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&Command{
		Flags: []Flag{
			&Int64SliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: []int64{}, Sources: EnvVars("APP_INTERVALS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			r := require.New(t)

			r.Equalf([]int64{20, 30, 40}, cmd.Int64Slice("intervals"), "main name not set from env")
			r.Equalf([]int64{20, 30, 40}, cmd.Int64Slice("i"), "short name not set from env")

			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiIntSliceFromEnvWithDefaults(t *testing.T) {
	t.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&Command{
		Flags: []Flag{
			&Int64SliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: []int64{1, 2, 5}, Sources: EnvVars("APP_INTERVALS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			r := require.New(t)

			r.Equalf([]int64{20, 30, 40}, cmd.Int64Slice("intervals"), "main name not set from env")
			r.Equalf([]int64{20, 30, 40}, cmd.Int64Slice("i"), "short name not set from env")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiIntSliceFromEnvCascade(t *testing.T) {
	t.Setenv("APP_INTERVALS", "20,30,40")

	_ = (&Command{
		Flags: []Flag{
			&Int64SliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: []int64{}, Sources: EnvVars("COMPAT_INTERVALS", "APP_INTERVALS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			r := require.New(t)

			r.Equalf([]int64{20, 30, 40}, cmd.Int64Slice("intervals"), "main name not set from env")
			r.Equalf([]int64{20, 30, 40}, cmd.Int64Slice("i"), "short name not set from env")

			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiFloat64(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&FloatFlag{Name: "serve", Aliases: []string{"s"}},
		},
		Action: func(_ context.Context, cmd *Command) error {
			assert.Equal(t, 10.2, cmd.Float("serve"), "main name not set")
			assert.Equal(t, 10.2, cmd.Float("s"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "-s", "10.2"})
}

func TestParseDestinationFloat64(t *testing.T) {
	var dest float64
	_ = (&Command{
		Flags: []Flag{
			&FloatFlag{
				Name:        "dest",
				Destination: &dest,
			},
		},
		Action: func(context.Context, *Command) error {
			assert.Equal(t, 10.2, dest, "expected destination Float64 10.2")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "--dest", "10.2"})
}

func TestParseMultiFloat64FromEnv(t *testing.T) {
	t.Setenv("APP_TIMEOUT_SECONDS", "15.5")
	_ = (&Command{
		Flags: []Flag{
			&FloatFlag{Name: "timeout", Aliases: []string{"t"}, Sources: EnvVars("APP_TIMEOUT_SECONDS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			assert.Equal(t, 15.5, cmd.Float("timeout"), "main name not set")
			assert.Equal(t, 15.5, cmd.Float("t"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiFloat64FromEnvCascade(t *testing.T) {
	t.Setenv("APP_TIMEOUT_SECONDS", "15.5")

	_ = (&Command{
		Flags: []Flag{
			&FloatFlag{Name: "timeout", Aliases: []string{"t"}, Sources: EnvVars("COMPAT_TIMEOUT_SECONDS", "APP_TIMEOUT_SECONDS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			assert.Equal(t, 15.5, cmd.Float("timeout"), "main name not set")
			assert.Equal(t, 15.5, cmd.Float("t"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiFloat64SliceFromEnv(t *testing.T) {
	t.Setenv("APP_INTERVALS", "0.1,-10.5")

	_ = (&Command{
		Flags: []Flag{
			&FloatSliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: []float64{}, Sources: EnvVars("APP_INTERVALS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			require.Equalf(t, []float64{0.1, -10.5}, cmd.FloatSlice("intervals"), "main name not set from env")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiFloat64SliceFromEnvCascade(t *testing.T) {
	t.Setenv("APP_INTERVALS", "0.1234,-10.5")

	_ = (&Command{
		Flags: []Flag{
			&FloatSliceFlag{Name: "intervals", Aliases: []string{"i"}, Value: []float64{}, Sources: EnvVars("COMPAT_INTERVALS", "APP_INTERVALS")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			require.Equalf(t, []float64{0.1234, -10.5}, cmd.FloatSlice("intervals"), "main name not set from env")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiBool(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&BoolFlag{Name: "serve", Aliases: []string{"s"}},
		},
		Action: func(_ context.Context, cmd *Command) error {
			assert.True(t, cmd.Bool("serve"), "main name not set")
			assert.True(t, cmd.Bool("s"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "--serve"})
}

func TestParseBoolShortOptionHandle(t *testing.T) {
	_ = (&Command{
		Commands: []*Command{
			{
				Name:                   "foobar",
				UseShortOptionHandling: true,
				Action: func(_ context.Context, cmd *Command) error {
					assert.True(t, cmd.Bool("serve"), "main name not set")
					assert.True(t, cmd.Bool("option"), "short name not set")
					return nil
				},
				Flags: []Flag{
					&BoolFlag{Name: "serve", Aliases: []string{"s"}},
					&BoolFlag{Name: "option", Aliases: []string{"o"}},
				},
			},
		},
	}).Run(buildTestContext(t), []string{"run", "foobar", "-so"})
}

func TestParseDestinationBool(t *testing.T) {
	var dest bool
	_ = (&Command{
		Flags: []Flag{
			&BoolFlag{
				Name:        "dest",
				Destination: &dest,
			},
		},
		Action: func(context.Context, *Command) error {
			assert.True(t, dest, "expected destination Bool true")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "--dest"})
}

func TestParseMultiBoolFromEnv(t *testing.T) {
	t.Setenv("APP_DEBUG", "1")
	_ = (&Command{
		Flags: []Flag{
			&BoolFlag{Name: "debug", Aliases: []string{"d"}, Sources: EnvVars("APP_DEBUG")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			assert.True(t, cmd.Bool("debug"), "main name not set")
			assert.True(t, cmd.Bool("d"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseMultiBoolFromEnvCascade(t *testing.T) {
	t.Setenv("APP_DEBUG", "1")
	_ = (&Command{
		Flags: []Flag{
			&BoolFlag{Name: "debug", Aliases: []string{"d"}, Sources: EnvVars("COMPAT_DEBUG", "APP_DEBUG")},
		},
		Action: func(_ context.Context, cmd *Command) error {
			assert.True(t, cmd.Bool("debug"), "main name not set from env")
			assert.True(t, cmd.Bool("d"), "short name not set from env")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run"})
}

func TestParseBoolFromEnv(t *testing.T) {
	boolFlagTests := []struct {
		input  string
		output bool
	}{
		{"", false},
		{"1", true},
		{"false", false},
		{"true", true},
	}

	for _, test := range boolFlagTests {
		t.Run(fmt.Sprintf("%[1]q %[2]v", test.input, test.output), func(t *testing.T) {
			t.Setenv("DEBUG", test.input)
			_ = (&Command{
				Flags: []Flag{
					&BoolFlag{Name: "debug", Aliases: []string{"d"}, Sources: EnvVars("DEBUG")},
				},
				Action: func(_ context.Context, cmd *Command) error {
					assert.Equal(t, test.output, cmd.Bool("debug"))
					assert.Equal(t, test.output, cmd.Bool("d"))
					return nil
				},
			}).Run(buildTestContext(t), []string{"run"})
		})
	}
}

func TestParseMultiBoolT(t *testing.T) {
	_ = (&Command{
		Flags: []Flag{
			&BoolFlag{Name: "implode", Aliases: []string{"i"}, Value: true},
		},
		Action: func(_ context.Context, cmd *Command) error {
			assert.False(t, cmd.Bool("implode"), "main name not set")
			assert.False(t, cmd.Bool("i"), "short name not set")
			return nil
		},
	}).Run(buildTestContext(t), []string{"run", "--implode=false"})
}

func TestStringSlice_Serialized_Set(t *testing.T) {
	sl0 := NewStringSlice("a", "b")
	ser0 := sl0.Serialize()

	require.GreaterOrEqual(t, len(ser0), len(slPfx), "serialized shorter than expected")

	sl1 := NewStringSlice("c", "d")
	_ = sl1.Set(ser0)

	require.Equal(t, sl0.String(), sl1.String(), "pre and post serialization do not match")
}

func TestIntSlice_Serialized_Set(t *testing.T) {
	sl0 := NewInt64Slice(1, 2)
	ser0 := sl0.Serialize()

	require.GreaterOrEqual(t, len(ser0), len(slPfx), "serialized shorter than expected")

	sl1 := NewInt64Slice(3, 4)
	_ = sl1.Set(ser0)

	require.Equal(t, sl0.String(), sl1.String(), "pre and post serialization do not match")
}

func TestUintSlice_Serialized_Set(t *testing.T) {
	sl0 := NewUint64Slice(1, 2)
	ser0 := sl0.Serialize()

	require.GreaterOrEqual(t, len(ser0), len(slPfx), "serialized shorter than expected")

	sl1 := NewUint64Slice(3, 4)
	_ = sl1.Set(ser0)

	require.Equal(t, sl0.String(), sl1.String(), "pre and post serialization do not match")
}

func TestUint64Slice_Serialized_Set(t *testing.T) {
	sl0 := NewUint64Slice(1, 2)
	ser0 := sl0.Serialize()

	require.GreaterOrEqual(t, len(ser0), len(slPfx), "serialized shorter than expected")

	sl1 := NewUint64Slice(3, 4)
	_ = sl1.Set(ser0)

	require.Equal(t, sl0.String(), sl1.String(), "pre and post serialization do not match")
}

func TestStringMap_Serialized_Set(t *testing.T) {
	m0 := NewStringMap(map[string]string{"a": "b"})
	ser0 := m0.Serialize()

	require.GreaterOrEqual(t, len(ser0), len(slPfx), "serialized shorter than expected")

	m1 := NewStringMap(map[string]string{"c": "d"})
	_ = m1.Set(ser0)

	require.Equal(t, m0.String(), m1.String(), "pre and post serialization do not match")
}

var timestampFlagTests = []struct {
	name     string
	aliases  []string
	usage    string
	expected string
}{
	{"foo", nil, "", "--foo time\t(default: 2020-04-10 01:01:01.000000001 +0000 UTC)"},
	{"f", nil, "all", "-f time\tall (default: 2020-04-10 01:01:01.000000001 +0000 UTC)"},
}

func TestTimestampFlagHelpOutput(t *testing.T) {
	tl, err := time.LoadLocation("UTC")
	assert.NoError(t, err)
	for _, test := range timestampFlagTests {
		value := time.Date(2020, time.April, 10, 1, 1, 1, 1, tl)
		fl := &TimestampFlag{Name: test.name, Aliases: test.aliases, Usage: test.usage, Value: value}
		assert.Equal(t, test.expected, fl.String())
	}
}

func TestTimestamp_set(t *testing.T) {
	ts := timestampValue{
		timestamp:  nil,
		hasBeenSet: false,
		layouts:    []string{"Jan 2, 2006 at 3:04pm (MST)"},
	}

	time1 := "Feb 3, 2013 at 7:54pm (PST)"
	require.NoError(t, ts.Set(time1), "Failed to parse time %s with layouts %v", time1, ts.layouts)
	require.True(t, ts.hasBeenSet, "hasBeenSet is not true after setting a time")

	ts.hasBeenSet = false
	ts.layouts = []string{time.RFC3339}
	time2 := "2006-01-02T15:04:05Z"
	require.NoError(t, ts.Set(time2), "Failed to parse time %s with layout %v", time2, ts.layouts)
	require.True(t, ts.hasBeenSet, "hasBeenSet is not true after setting a time")
}

func TestTimestampFlagApply_SingleFormat(t *testing.T) {
	expectedResult, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	fl := TimestampFlag{Name: "time", Aliases: []string{"t"}, Config: TimestampConfig{Layouts: []string{time.RFC3339}}}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{"", "--time", "2006-01-02T15:04:05Z"}))
	assert.Equal(t, expectedResult, cmd.Value("time"))
}

func TestTimestampFlagApply_MultipleFormats(t *testing.T) {
	now := time.Now().UTC()

	testCases := []struct {
		caseName          string
		layoutsPrecisions map[string]time.Duration
		expRes            time.Time
		expErrValidation  func(err error) (validation error)
	}{
		{
			caseName: "all_valid_layouts",
			layoutsPrecisions: map[string]time.Duration{
				time.RFC3339:  time.Second,
				time.DateTime: time.Second,
				time.RFC1123:  time.Second,
			},
			expRes: now.Truncate(time.Second),
		},
		{
			caseName: "one_invalid_layout",
			layoutsPrecisions: map[string]time.Duration{
				time.RFC3339:  time.Second,
				time.DateTime: time.Second,
				"foo":         0,
			},
			expRes: now.Truncate(time.Second),
		},
		{
			caseName: "multiple_invalid_layouts",
			layoutsPrecisions: map[string]time.Duration{
				time.RFC3339:  time.Second,
				"foo":         0,
				time.DateTime: time.Second,
				"bar":         0,
			},
			expRes: now.Truncate(time.Second),
		},
		{
			caseName: "all_invalid_layouts",
			layoutsPrecisions: map[string]time.Duration{
				"foo":                      0,
				"2024-08-07 74:01:82Z-100": 0,
				"25:70":                    0,
				"":                         0,
			},
			expErrValidation: func(err error) error {
				if err == nil {
					return errors.New("got nil err")
				}

				found := regexp.MustCompile(`(cannot parse ".+" as ".*")|(extra text: ".+")`).Match([]byte(err.Error()))
				if !found {
					return fmt.Errorf("given error does not satisfy pattern: %w", err)
				}

				return nil
			},
		},
		{
			caseName: "empty_layout",
			layoutsPrecisions: map[string]time.Duration{
				"": 0,
			},
			expErrValidation: func(err error) error {
				if err == nil {
					return errors.New("got nil err")
				}

				found := regexp.MustCompile(`extra text: ".+"`).Match([]byte(err.Error()))
				if !found {
					return fmt.Errorf("given error does not satisfy pattern: %w", err)
				}

				return nil
			},
		},
		{
			caseName: "nil_layouts_slice",
			expErrValidation: func(err error) error {
				if err == nil {
					return errors.New("got nil err")
				}

				found := regexp.MustCompile(`got nil/empty layouts slice`).Match([]byte(err.Error()))
				if !found {
					return fmt.Errorf("given error does not satisfy pattern: %w", err)
				}

				return nil
			},
		},
		{
			caseName:          "empty_layouts_slice",
			layoutsPrecisions: map[string]time.Duration{},
			expErrValidation: func(err error) error {
				if err == nil {
					return errors.New("got nil err")
				}

				found := regexp.MustCompile(`got nil/empty layouts slice`).Match([]byte(err.Error()))
				if !found {
					return fmt.Errorf("given error does not satisfy pattern: %w", err)
				}

				return nil
			},
		},
	}

	// TODO: replace with maps.Keys() (go >= ), lo.Keys() if acceptable
	getKeys := func(m map[string]time.Duration) []string {
		if m == nil {
			return nil
		}

		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	}

	for idx := range testCases {
		testCase := testCases[idx]
		t.Run(testCase.caseName, func(t *testing.T) {
			// t.Parallel()
			fl := TimestampFlag{
				Name: "time",
				Config: TimestampConfig{
					Layouts: getKeys(testCase.layoutsPrecisions),
				},
			}

			if len(testCase.layoutsPrecisions) == 0 {
				err := fl.Set(fl.Name, now.Format(time.RFC3339))
				if testCase.expErrValidation != nil {
					assert.NoError(t, testCase.expErrValidation(err))
				}
			}

			validLayouts := make([]string, 0, len(testCase.layoutsPrecisions))
			invalidLayouts := make([]string, 0, len(testCase.layoutsPrecisions))

			// TODO: replace with lo.Filter if acceptable
			for layout, prec := range testCase.layoutsPrecisions {
				v, err := time.Parse(layout, now.Format(layout))
				if err != nil || prec == 0 || now.Truncate(prec).UnixNano() != v.Truncate(prec).UnixNano() {
					invalidLayouts = append(invalidLayouts, layout)
					continue
				}
				validLayouts = append(validLayouts, layout)
			}

			for _, layout := range validLayouts {
				err := fl.Set(fl.Name, now.Format(layout))
				assert.NoError(t, err)
				if !testCase.expRes.IsZero() {
					assert.Equal(t, testCase.expRes, fl.value.Get())
				}
			}

			for range invalidLayouts {
				err := fl.Set(fl.Name, now.Format(time.RFC3339))
				if testCase.expErrValidation != nil {
					assert.NoError(t, testCase.expErrValidation(err))
				}
			}
		})
	}
}

func TestTimestampFlagApply_ShortenedLayouts(t *testing.T) {
	now := time.Now().UTC()

	shortenedLayoutsPrecisions := map[string]time.Duration{
		time.Kitchen:    time.Minute,
		time.Stamp:      time.Second,
		time.StampMilli: time.Millisecond,
		time.StampMicro: time.Microsecond,
		time.StampNano:  time.Nanosecond,
		time.TimeOnly:   time.Second,
		"15:04":         time.Minute,
	}

	// TODO: replace with maps.Keys() (go >= ), lo.Keys() if acceptable
	getKeys := func(m map[string]time.Duration) []string {
		if m == nil {
			return nil
		}

		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	}

	fl := TimestampFlag{
		Name: "time",
		Config: TimestampConfig{
			Layouts: getKeys(shortenedLayoutsPrecisions),
		},
	}

	for layout, prec := range shortenedLayoutsPrecisions {
		err := fl.Set(fl.Name, now.Format(layout))
		assert.NoError(t, err)
		assert.Equal(t, now.Truncate(prec), fl.value.Get())
	}
}

func TestTimestampFlagApplyValue(t *testing.T) {
	expectedResult, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	fl := TimestampFlag{Name: "time", Aliases: []string{"t"}, Config: TimestampConfig{Layouts: []string{time.RFC3339}}, Value: expectedResult}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{""}))
	assert.Equal(t, expectedResult, cmd.Value("time"))
}

func TestTimestampFlagApply_Fail_Parse_Wrong_Layout(t *testing.T) {
	fl := TimestampFlag{Name: "time", Aliases: []string{"t"}, Config: TimestampConfig{Layouts: []string{"randomlayout"}}}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"", "--time", "2006-01-02T15:04:05Z"})
	assert.EqualError(t, err, "invalid value \"2006-01-02T15:04:05Z\" for flag -time: parsing time \"2006-01-02T15:04:05Z\" as \"randomlayout\": cannot parse \"2006-01-02T15:04:05Z\" as \"randomlayout\"")
}

func TestTimestampFlagApply_Fail_Parse_Wrong_Time(t *testing.T) {
	fl := TimestampFlag{Name: "time", Aliases: []string{"t"}, Config: TimestampConfig{Layouts: []string{"Jan 2, 2006 at 3:04pm (MST)"}}}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}
	err := cmd.Set("time", "2006-01-02T15:04:05Z")
	assert.EqualError(t, err, "parsing time \"2006-01-02T15:04:05Z\" as \"Jan 2, 2006 at 3:04pm (MST)\": cannot parse \"2006-01-02T15:04:05Z\" as \"Jan\"")
}

func TestTimestampFlagApply_Timezoned(t *testing.T) {
	pdt := time.FixedZone("PDT", -7*60*60)
	expectedResult, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	fl := TimestampFlag{Name: "time", Aliases: []string{"t"}, Config: TimestampConfig{Layouts: []string{time.ANSIC}, Timezone: pdt}}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{"", "--time", "Mon Jan 2 08:04:05 2006"}))
	assert.Equal(t, expectedResult.In(pdt), cmd.Value("time"))
}

func TestTimestampFlagValueFromCommand(t *testing.T) {
	now := time.Now()
	cmd := &Command{
		Flags: []Flag{
			&TimestampFlag{
				Name:  "myflag",
				Value: now,
			},
		},
	}
	f := &TimestampFlag{Name: "myflag"}
	require.Equal(t, now, cmd.Timestamp(f.Name))
}

type flagDefaultTestCase struct {
	name    string
	flag    Flag
	toParse []string
	expect  string
}

func TestFlagDefaultValue(t *testing.T) {
	cases := []*flagDefaultTestCase{
		{
			name:    "stringSlice",
			flag:    &StringSliceFlag{Name: "flag", Value: []string{"default1", "default2"}},
			toParse: []string{"--flag", "parsed"},
			expect:  `--flag string [ --flag string ]	(default: "default1", "default2")`,
		},
		{
			name:    "floatSlice",
			flag:    &FloatSliceFlag{Name: "flag", Value: []float64{1.1, 2.2}},
			toParse: []string{"--flag", "13.3"},
			expect:  `--flag float [ --flag float ]	(default: 1.1, 2.2)`,
		},
		{
			name:    "float32Slice",
			flag:    &Float32SliceFlag{Name: "flag", Value: []float32{1.1, 2.2}},
			toParse: []string{"--flag", "13.3"},
			expect:  `--flag float [ --flag float ]	(default: 1.1, 2.2)`,
		},
		{
			name:    "float64Slice",
			flag:    &Float64SliceFlag{Name: "flag", Value: []float64{1.1, 2.2}},
			toParse: []string{"--flag", "13.3"},
			expect:  `--flag float [ --flag float ]	(default: 1.1, 2.2)`,
		},
		{
			name:    "intSlice",
			flag:    &Int64SliceFlag{Name: "flag", Value: []int64{1, 2}},
			toParse: []string{"--flag", "13"},
			expect:  `--flag int [ --flag int ]	(default: 1, 2)`,
		},
		{
			name:    "uintSlice",
			flag:    &Uint64SliceFlag{Name: "flag", Value: []uint64{1, 2}},
			toParse: []string{"--flag", "13"},
			expect:  `--flag uint [ --flag uint ]	(default: 1, 2)`,
		},
		{
			name:    "string",
			flag:    &StringFlag{Name: "flag", Value: "default"},
			toParse: []string{"--flag", "parsed"},
			expect:  `--flag string	(default: "default")`,
		},
		{
			name:    "bool",
			flag:    &BoolFlag{Name: "flag", Value: true},
			toParse: []string{"--flag=false"},
			expect:  `--flag	`,
		},
		{
			name:    "uint",
			flag:    &UintFlag{Name: "flag", Value: 1},
			toParse: []string{"--flag", "13"},
			expect:  `--flag uint	(default: 1)`,
		},
		{
			name:    "uint64",
			flag:    &Uint64Flag{Name: "flag", Value: 1},
			toParse: []string{"--flag", "13"},
			expect:  `--flag uint	(default: 1)`,
		},
		{
			name:    "stringMap",
			flag:    &StringMapFlag{Name: "flag", Value: map[string]string{"default1": "default2"}},
			toParse: []string{"--flag", "parsed="},
			expect:  `--flag string=string [ --flag string=string ]	(default: default1="default2")`,
		},
	}
	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			cmd := &Command{
				Flags: []Flag{
					v.flag,
				},
			}
			assert.NoError(t, cmd.Run(buildTestContext(t), append([]string{""}, v.toParse...)))
			assert.Equal(t, v.expect, v.flag.String())
		})
	}
}

type flagDefaultTestCaseWithEnv struct {
	name    string
	flag    Flag
	toParse []string
	expect  string
	environ map[string]string
}

func TestFlagDefaultValueWithEnv(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2005-01-02T15:04:05Z")
	require.NoError(t, err)
	cases := []*flagDefaultTestCaseWithEnv{
		{
			name:    "stringSlice",
			flag:    &StringSliceFlag{Name: "flag", Value: []string{"default1", "default2"}, Sources: EnvVars("ssflag")},
			toParse: []string{"--flag", "parsed"},
			expect:  `--flag string [ --flag string ]	(default: "default1", "default2")` + withEnvHint([]string{"ssflag"}, ""),
			environ: map[string]string{
				"ssflag": "some-other-env_value",
			},
		},
		{
			name:    "float64Slice",
			flag:    &FloatSliceFlag{Name: "flag", Value: []float64{1.1, 2.2}, Sources: EnvVars("fsflag")},
			toParse: []string{"--flag", "13.3"},
			expect:  `--flag float [ --flag float ]	(default: 1.1, 2.2)` + withEnvHint([]string{"fsflag"}, ""),
			environ: map[string]string{
				"fsflag": "20304.222",
			},
		},
		{
			name:    "intSlice",
			flag:    &Int64SliceFlag{Name: "flag", Value: []int64{1, 2}, Sources: EnvVars("isflag")},
			toParse: []string{"--flag", "13"},
			expect:  `--flag int [ --flag int ]	(default: 1, 2)` + withEnvHint([]string{"isflag"}, ""),
			environ: map[string]string{
				"isflag": "101",
			},
		},
		{
			name:    "uintSlice",
			flag:    &Uint64SliceFlag{Name: "flag", Value: []uint64{1, 2}, Sources: EnvVars("uisflag")},
			toParse: []string{"--flag", "13"},
			expect:  `--flag uint [ --flag uint ]	(default: 1, 2)` + withEnvHint([]string{"uisflag"}, ""),
			environ: map[string]string{
				"uisflag": "3",
			},
		},
		{
			name:    "string",
			flag:    &StringFlag{Name: "flag", Value: "default", Sources: EnvVars("uflag")},
			toParse: []string{"--flag", "parsed"},
			expect:  `--flag string	(default: "default")` + withEnvHint([]string{"uflag"}, ""),
			environ: map[string]string{
				"uflag": "some-other-string",
			},
		},
		{
			name:    "bool",
			flag:    &BoolFlag{Name: "flag", Value: true, Sources: EnvVars("uflag")},
			toParse: []string{"--flag=false"},
			expect:  `--flag	` + withEnvHint([]string{"uflag"}, ""),
			environ: map[string]string{
				"uflag": "false",
			},
		},
		{
			name:   "bool",
			flag:   &BoolWithInverseFlag{Name: "flag", Value: true, Sources: EnvVars("uflag")},
			expect: `--[no-]flag	(default: true)` + withEnvHint([]string{"uflag"}, ""),
			environ: map[string]string{
				"uflag": "false",
			},
		},
		{
			name:    "uint64",
			flag:    &Uint64Flag{Name: "flag", Value: 1, Sources: EnvVars("uflag")},
			toParse: []string{"--flag", "13"},
			expect:  `--flag uint	(default: 1)` + withEnvHint([]string{"uflag"}, ""),
			environ: map[string]string{
				"uflag": "10",
			},
		},
		{
			name:    "uint",
			flag:    &Uint64Flag{Name: "flag", Value: 1, Sources: EnvVars("uflag")},
			toParse: []string{"--flag", "13"},
			expect:  `--flag uint	(default: 1)` + withEnvHint([]string{"uflag"}, ""),
			environ: map[string]string{
				"uflag": "10",
			},
		},
		{
			name:    "int64",
			flag:    &Int64Flag{Name: "flag", Value: 1, Sources: EnvVars("uflag")},
			toParse: []string{"--flag", "13"},
			expect:  `--flag int	(default: 1)` + withEnvHint([]string{"uflag"}, ""),
			environ: map[string]string{
				"uflag": "10",
			},
		},
		{
			name:    "int",
			flag:    &Int64Flag{Name: "flag", Value: 1, Sources: EnvVars("uflag")},
			toParse: []string{"--flag", "13"},
			expect:  `--flag int	(default: 1)` + withEnvHint([]string{"uflag"}, ""),
			environ: map[string]string{
				"uflag": "10",
			},
		},
		{
			name:    "duration",
			flag:    &DurationFlag{Name: "flag", Value: time.Second, Sources: EnvVars("uflag")},
			toParse: []string{"--flag", "2m"},
			expect:  `--flag duration	(default: 1s)` + withEnvHint([]string{"uflag"}, ""),
			environ: map[string]string{
				"uflag": "2h4m10s",
			},
		},
		{
			name:    "timestamp",
			flag:    &TimestampFlag{Name: "flag", Value: ts, Config: TimestampConfig{Layouts: []string{time.RFC3339}}, Sources: EnvVars("tflag")},
			toParse: []string{"--flag", "2006-11-02T15:04:05Z"},
			expect:  `--flag time	(default: 2005-01-02 15:04:05 +0000 UTC)` + withEnvHint([]string{"tflag"}, ""),
			environ: map[string]string{
				"tflag": "2010-01-02T15:04:05Z",
			},
		},
		{
			name:    "stringMap",
			flag:    &StringMapFlag{Name: "flag", Value: map[string]string{"default1": "default2"}, Sources: EnvVars("ssflag")},
			toParse: []string{"--flag", "parsed="},
			expect:  `--flag string=string [ --flag string=string ]	(default: default1="default2")` + withEnvHint([]string{"ssflag"}, ""),
			environ: map[string]string{
				"ssflag": "some-other-env_value=",
			},
		},
		// TODO
		/*{
			name:    "generic",
			flag:    &GenericFlag{Name: "flag", Value: &Parser{"11", "12"}, Sources: EnvVars("gflag")},
			toParse: []string{"--flag", "15,16"},
			expect:  `--flag value	(default: 11,12)` + withEnvHint([]string{"gflag"}, ""),
			environ: map[string]string{
				"gflag": "13,14",
			},
		},*/
	}
	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			for key, val := range v.environ {
				t.Setenv(key, val)
			}
			cmd := &Command{
				Flags: []Flag{
					v.flag,
				},
			}
			require.NoError(t, cmd.Run(buildTestContext(t), append([]string{""}, v.toParse...)))
			assert.Equal(t, v.expect, v.flag.String())
		})
	}
}

type flagValueTestCase struct {
	name    string
	flag    Flag
	toParse []string
	expect  string
}

func TestFlagValue(t *testing.T) {
	t.SkipNow()
	cases := []*flagValueTestCase{
		{
			name:    "stringSlice",
			flag:    &StringSliceFlag{Name: "flag", Value: []string{"default1", "default2"}},
			toParse: []string{"--flag", "parsed,parsed2", "--flag", "parsed3,parsed4"},
			expect:  `[parsed parsed2 parsed3 parsed4]`,
		},
		{
			name:    "float64Slice",
			flag:    &FloatSliceFlag{Name: "flag", Value: []float64{1.1, 2.2}},
			toParse: []string{"--flag", "13.3,14.4", "--flag", "15.5,16.6"},
			expect:  `[]float64{13.3, 14.4, 15.5, 16.6}`,
		},
		{
			name:    "intSlice",
			flag:    &Int64SliceFlag{Name: "flag", Value: []int64{1, 2}},
			toParse: []string{"--flag", "13,14", "--flag", "15,16"},
			expect:  `[]int64{13, 14, 15, 16}`,
		},
		{
			name:    "uintSlice",
			flag:    &Uint64SliceFlag{Name: "flag", Value: []uint64{1, 2}},
			toParse: []string{"--flag", "13,14", "--flag", "15,16"},
			expect:  `[]uint64{13, 14, 15, 16}`,
		},
		{
			name:    "stringMap",
			flag:    &StringMapFlag{Name: "flag", Value: map[string]string{"default1": "default2"}},
			toParse: []string{"--flag", "parsed=parsed2", "--flag", "parsed3=parsed4"},
			expect:  `map[parsed:parsed2 parsed3:parsed4]`,
		},
		{
			name:    "int",
			flag:    &IntFlag{Name: "flag", Value: 1},
			toParse: []string{"--flag", "42"},
			expect:  `int(42)`,
		},
		{
			name:    "uint",
			flag:    &UintFlag{Name: "flag", Value: 1},
			toParse: []string{"--flag", "42"},
			expect:  `uint(42)`,
		},
	}
	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			cmd := &Command{
				Flags: []Flag{
					v.flag,
				},
			}
			assert.NoError(t, cmd.Run(buildTestContext(t), append([]string{""}, v.toParse...)))
			f := cmd.lookupFlag("flag")
			require.Equal(t, v.expect, f.String())
		})
	}
}

func TestTimestampFlagApply_WithDestination(t *testing.T) {
	var destination time.Time
	expectedResult, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	fl := TimestampFlag{Name: "time", Aliases: []string{"t"}, Config: TimestampConfig{Layouts: []string{time.RFC3339}}, Destination: &destination}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{"", "--time", "2006-01-02T15:04:05Z"}))
	assert.Equal(t, expectedResult, destination)
}

// Test issue #1254
// StringSlice() with UseShortOptionHandling causes duplicated entries, depending on the ordering of the flags
func TestSliceShortOptionHandle(t *testing.T) {
	wasCalled := false
	err := (&Command{
		Name:                   "foobar",
		UseShortOptionHandling: true,
		Action: func(_ context.Context, cmd *Command) error {
			wasCalled = true

			if !cmd.Bool("i") {
				return fmt.Errorf("bool i not set")
			}

			if !cmd.Bool("t") {
				return fmt.Errorf("bool i not set")
			}

			ss := cmd.StringSlice("net")
			if !reflect.DeepEqual(ss, []string{"foo"}) {
				return fmt.Errorf("got different slice %q than expected", ss)
			}

			return nil
		},
		Flags: []Flag{
			&StringSliceFlag{Name: "net"},
			&BoolFlag{Name: "i"},
			&BoolFlag{Name: "t"},
		},
	}).Run(buildTestContext(t), []string{"foobar", "--net=foo", "-it"})

	r := require.New(t)

	r.NoError(err)
	r.Truef(wasCalled, "action callback was never called")
}

// Test issue #1541
func TestCustomizedSliceFlagSeparator(t *testing.T) {
	opts := []string{"opt1", "opt2", "opt3,op", "opt4"}
	ret := flagSplitMultiValues(strings.Join(opts, ";"), ";", disableSliceFlagSeparator)
	require.Equal(t, 4, len(ret), "split slice flag failed")
	for idx, r := range ret {
		require.Equal(t, opts[idx], r, "get %dth failed", idx)
	}
}

func TestFlagSplitMultiValues_Disabled(t *testing.T) {
	opts := []string{"opt1", "opt2", "opt3,op", "opt4"}
	ret := flagSplitMultiValues(strings.Join(opts, defaultSliceFlagSeparator), defaultSliceFlagSeparator, true)
	require.Equal(t, 1, len(ret), "failed to disable split slice flag")
	require.Equal(t, strings.Join(opts, defaultSliceFlagSeparator), ret[0])
}

var stringMapFlagTests = []struct {
	name     string
	aliases  []string
	value    map[string]string
	expected string
}{
	{"foo", nil, nil, "--foo string=string [ --foo string=string ]\t"},
	{"f", nil, nil, "-f string=string [ -f string=string ]\t"},
	{"f", nil, map[string]string{"Lipstick": ""}, "-f string=string [ -f string=string ]\t(default: Lipstick=)"},
	{"test", nil, map[string]string{"Something": ""}, "--test string=string [ --test string=string ]\t(default: Something=)"},
	{"dee", []string{"d"}, map[string]string{"Inka": "Dinka", "dooo": ""}, "--dee string=string, -d string=string [ --dee string=string, -d string=string ]\t(default: Inka=\"Dinka\", dooo=)"},
}

func TestStringMapFlagHelpOutput(t *testing.T) {
	for _, test := range stringMapFlagTests {
		f := &StringMapFlag{Name: test.name, Aliases: test.aliases, Value: test.value}
		assert.Equal(t, test.expected, f.String())
	}
}

func TestStringMapFlagWithEnvVarHelpOutput(t *testing.T) {
	t.Setenv("APP_QWWX", "11,4")

	for _, test := range stringMapFlagTests {
		fl := &StringMapFlag{Name: test.name, Aliases: test.aliases, Value: test.value, Sources: EnvVars("APP_QWWX")}
		output := fl.String()

		expectedSuffix := withEnvHint([]string{"APP_QWWX"}, "")
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

func TestStringMapFlagApply_SetsAllNames(t *testing.T) {
	fl := StringMapFlag{Name: "goat", Aliases: []string{"G", "gooots"}}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{"", "--goat", "aaa=", "-G", "bbb=", "--gooots", "eeeee="}))
}

func TestStringMapFlagApply_UsesEnvValues_noDefault(t *testing.T) {
	t.Setenv("MY_GOAT", "vincent van goat=scape goat")
	var val map[string]string
	fl := StringMapFlag{Name: "goat", Sources: EnvVars("MY_GOAT"), Value: val}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{""}))
	assert.Nil(t, val)
	assert.Equal(t, map[string]string{"vincent van goat": "scape goat"}, cmd.Value("goat"))
}

func TestStringMapFlagApply_UsesEnvValues_withDefault(t *testing.T) {
	t.Setenv("MY_GOAT", "vincent van goat=scape goat")
	val := map[string]string{`some default`: `values here`}
	fl := StringMapFlag{Name: "goat", Sources: EnvVars("MY_GOAT"), Value: val}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{""}))
	assert.Equal(t, map[string]string{`some default`: `values here`}, val)
	assert.Equal(t, map[string]string{"vincent van goat": "scape goat"}, cmd.Value("goat"))
}

func TestStringMapFlagApply_DefaultValueWithDestination(t *testing.T) {
	defValue := map[string]string{"UA": "US"}

	fl := StringMapFlag{Name: "country", Value: defValue, Destination: &map[string]string{"CA": ""}}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.NoError(t, cmd.Run(buildTestContext(t), []string{""}))
	assert.Equal(t, defValue, *fl.Destination)
}

func TestStringMapFlagValueFromCommand(t *testing.T) {
	f := &StringMapFlag{Name: "myflag"}
	cmd := &Command{
		Flags: []Flag{
			f,
		},
	}
	assert.NoError(t, cmd.Run(buildTestContext(t), []string{""}))
	require.NoError(t, cmd.Set("myflag", "a=b"))
	require.NoError(t, cmd.Set("myflag", "c="))

	require.Equal(t, map[string]string{"a": "b", "c": ""}, cmd.StringMap(f.Name))
}

func TestStringMapFlagApply_Error(t *testing.T) {
	fl := StringMapFlag{Name: "goat"}
	cmd := &Command{
		Flags: []Flag{
			&fl,
		},
	}

	assert.Error(t, cmd.Run(buildTestContext(t), []string{"", "--goat", "aaa", "bbb="}))
}

func TestZeroValueMutexFlag(t *testing.T) {
	var fl MutuallyExclusiveFlags
	assert.NoError(t, fl.check(&Command{}))
}

func TestExtFlag(t *testing.T) {
	var iv intValue[int64]
	var ipv int64

	f := &flag.Flag{
		Name:     "bar",
		Usage:    "bar usage",
		Value:    iv.Create(11, &ipv, IntegerConfig{}),
		DefValue: "10",
	}

	extF := &extFlag{
		f: f,
	}

	assert.Equal(t, []string{"bar"}, extF.Names())
	assert.True(t, extF.IsVisible())
	assert.False(t, extF.IsSet())
	assert.False(t, extF.TakesValue())
	assert.Equal(t, "bar usage", extF.GetUsage())
	assert.Equal(t, "11", extF.GetValue())
	assert.Equal(t, "10", extF.GetDefaultText())
	assert.Nil(t, extF.GetEnvVars())
}

func TestSliceValuesNil(t *testing.T) {
	assert.Equal(t, []float64(nil), NewFloatSlice().Value())
	assert.Equal(t, []float32(nil), NewFloat32Slice().Value())
	assert.Equal(t, []float64(nil), NewFloat64Slice().Value())
	assert.Equal(t, []int64(nil), NewInt64Slice().Value())
	assert.Equal(t, []uint64(nil), NewUint64Slice().Value())
	assert.Equal(t, []string(nil), NewStringSlice().Value())

	assert.Equal(t, []float64(nil), (&FloatSlice{}).Value())
	assert.Equal(t, []float32(nil), (&Float32Slice{}).Value())
	assert.Equal(t, []float64(nil), (&Float64Slice{}).Value())
	assert.Equal(t, []int64(nil), (&Int64Slice{}).Value())
	assert.Equal(t, []uint64(nil), (&Uint64Slice{}).Value())
	assert.Equal(t, []string(nil), (&StringSlice{}).Value())
}

func TestFileHint(t *testing.T) {
	assert.Equal(t, "", withFileHint("", ""))
	assert.Equal(t, " [/tmp/foo.txt]", withFileHint("/tmp/foo.txt", ""))
	assert.Equal(t, "foo", withFileHint("", "foo"))
	assert.Equal(t, "bar [/tmp/foo.txt]", withFileHint("/tmp/foo.txt", "bar"))
}

func TestHasFlags(t *testing.T) {
	flagToCheck := &StringFlag{Name: "foo"}
	flags := []Flag{
		&StringFlag{Name: "bar"},
		&Int64Flag{Name: "baz"},
		flagToCheck,
	}

	assert.True(t, hasFlag(flags, flagToCheck))
}

func TestFlagsByName(t *testing.T) {
	flags := []Flag{
		&StringFlag{
			Name: "b2",
		},
		&Int64Flag{
			Name: "a0",
		},
		&FloatFlag{
			Name: "b1",
		},
	}

	flagsByName := FlagsByName(flags)
	sort.Sort(flagsByName)

	assert.Equal(t, len(flags), flagsByName.Len())

	var prev Flag
	for _, f := range flags {
		if prev != nil {
			assert.LessOrEqual(t, prev.Names()[0], f.Names()[0])
		}
		prev = f
	}
}

func TestNonStringMap(t *testing.T) {
	type (
		floatMap = MapBase[float64, NoConfig, floatValue[float64]]
	)

	p := map[string]float64{}

	var fv floatValue[float64]

	f := &floatMap{
		value: &fv,
	}

	assert.Equal(t, map[string]float64{}, f.Value())
	f.dict = &p
	assert.Equal(t, map[string]float64{}, f.Value())
	assert.Equal(t, "map[string]float64{}", f.String())

	assert.ErrorContains(t, f.Set("invalid=value"), "ParseFloat")
}

func TestUnquoteUsage(t *testing.T) {
	tests := []struct {
		str      string
		expStr   string
		expUsage string
	}{
		{"foo", "", "foo"},
		{"foo something", "", "foo something"},
		{"foo `bar 11`", "bar 11", "foo bar 11"},
		{"foo `bar 11` sobar", "bar 11", "foo bar 11 sobar"},
		{"foo `bar 11", "", "foo `bar 11"},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("unquote %d", i), func(t *testing.T) {
			str, usage := unquoteUsage(test.str)
			assert.Equal(t, test.expStr, str)
			assert.Equal(t, test.expUsage, usage)
		})
	}
}

func TestEnvHintWindows(t *testing.T) {
	if runtime.GOOS == "windows" && os.Getenv("PSHOME") == "" {
		assert.Equal(t, "something [%foo%, %bar%, %ss%]", withEnvHint([]string{"foo", "bar", "ss"}, "something"))
	}
}

func TestDocGetValue(t *testing.T) {
	assert.Equal(t, "true", (&BoolFlag{Name: "foo", Value: true}).GetValue())
	assert.Equal(t, "false", (&BoolFlag{Name: "foo", Value: false}).GetValue())
	assert.Equal(t, "\"bar\"", (&StringFlag{Name: "foo", Value: "bar"}).GetValue())
	assert.Equal(t, "", (&BoolWithInverseFlag{Name: "foo", Value: false}).GetValue())
}

func TestGenericFlag_SatisfiesFlagInterface(t *testing.T) {
	var f Flag = &GenericFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestGenericValue_SatisfiesBoolInterface(t *testing.T) {
	var f boolFlag = &genericValue{}
	var fpv float64

	assert.False(t, f.IsBoolFlag())

	fv := floatValue[float64]{val: &fpv}
	f = &genericValue{
		val: &fv,
	}

	assert.False(t, f.IsBoolFlag())

	f = &genericValue{
		val: &boolValue{},
	}
	assert.True(t, f.IsBoolFlag())
}

func TestGenericFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &GenericFlag{}

	_ = f.String()
}

func TestGenericFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f RequiredFlag = &GenericFlag{}

	_ = f.IsRequired()
}

func TestGenericFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f VisibleFlag = &GenericFlag{}

	_ = f.IsVisible()
}

func TestGenericFlag_SatisfiesDocFlagInterface(t *testing.T) {
	var f DocGenerationFlag = &GenericFlag{}

	_ = f.GetUsage()
}

func TestGenericValue(t *testing.T) {
	g := &genericValue{}
	assert.NoError(t, g.Set("something"))
	assert.Nil(t, g.Get())
	assert.Empty(t, g.String())
}
