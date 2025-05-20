package cli

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var errBothEnvFlagsAreSet = fmt.Errorf("cannot set both flags `--env` and `--no-env`")

type boolWithInverseTestCase struct {
	args    []string
	toBeSet bool
	value   bool
	err     error
	envVars map[string]string
}

func (tc *boolWithInverseTestCase) Run(t *testing.T, flagWithInverse *BoolWithInverseFlag) error {
	cmd := &Command{
		Flags:  []Flag{flagWithInverse},
		Action: func(context.Context, *Command) error { return nil },
	}

	for key, val := range tc.envVars {
		t.Setenv(key, val)
	}

	err := cmd.Run(buildTestContext(t), append([]string{"prog"}, tc.args...))
	if err != nil {
		return err
	}

	if flagWithInverse.IsSet() != tc.toBeSet {
		return fmt.Errorf("flag should be set %t, but got %t", tc.toBeSet, flagWithInverse.IsSet())
	}

	if flagWithInverse.Get() != tc.value {
		return fmt.Errorf("flag value should be %t, but got %t", tc.value, flagWithInverse.Get())
	}

	return nil
}

func runBoolWithInverseFlagTests(t *testing.T, newFlagMethod func() *BoolWithInverseFlag, cases []*boolWithInverseTestCase) error {
	for _, tc := range cases {
		t.Run(strings.Join(tc.args, " ")+fmt.Sprintf("%[1]v %[2]v %[3]v", tc.value, tc.toBeSet, tc.err), func(t *testing.T) {
			r := require.New(t)

			fl := newFlagMethod()

			err := tc.Run(t, fl)
			if err != nil && tc.err == nil {
				r.NoError(err)
			}

			if err == nil && tc.err != nil {
				r.Error(err)
			}

			if err != nil && tc.err != nil {
				r.ErrorContains(err, tc.err.Error())
			}
		})
	}

	return nil
}

func TestBoolWithInverseBasic(t *testing.T) {
	flagMethod := func() *BoolWithInverseFlag {
		return &BoolWithInverseFlag{
			Name: "env",
		}
	}

	testCases := []*boolWithInverseTestCase{
		{
			args:    []string{"--no-env"},
			toBeSet: true,
			value:   false,
		},
		{
			args:    []string{"--env"},
			toBeSet: true,
			value:   true,
		},
		{
			toBeSet: false,
			value:   false,
		},
		{
			args: []string{"--env", "--no-env"},
			err:  errBothEnvFlagsAreSet,
		},
	}

	err := runBoolWithInverseFlagTests(t, flagMethod, testCases)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestBoolWithInverseAction(t *testing.T) {
	err := fmt.Errorf("action called")
	flagMethod := func() *BoolWithInverseFlag {
		bif := &BoolWithInverseFlag{
			Name: "env",

			// Setting env to the opposite to test flag Action is working as intended
			Action: func(_ context.Context, cmd *Command, value bool) error {
				return err
			},
		}
		return bif
	}

	testCases := []*boolWithInverseTestCase{
		{
			args:    []string{"--no-env"},
			toBeSet: true,
			value:   false,
			err:     err,
		},
		{
			args:    []string{"--env"},
			toBeSet: true,
			value:   true,
			err:     err,
		},

		// This test is not inverse because the flag action is never called
		{
			toBeSet: false,
			value:   false,
		},
		{
			args: []string{"--env", "--no-env"},
			err:  errBothEnvFlagsAreSet,
		},
	}

	errr := runBoolWithInverseFlagTests(t, flagMethod, testCases)
	if errr != nil {
		t.Error(errr)
		return
	}
}

func TestBoolWithInverseAlias(t *testing.T) {
	flagMethod := func() *BoolWithInverseFlag {
		return &BoolWithInverseFlag{
			Name:    "env",
			Aliases: []string{"e", "do-env"},
		}
	}

	testCases := []*boolWithInverseTestCase{
		{
			args:    []string{"--no-e"},
			toBeSet: true,
			value:   false,
		},
		{
			args:    []string{"--e"},
			toBeSet: true,
			value:   true,
		},
		{
			toBeSet: false,
			value:   false,
		},
		{
			args: []string{"--do-env", "--no-do-env"},
			err:  errBothEnvFlagsAreSet,
		},
	}

	err := runBoolWithInverseFlagTests(t, flagMethod, testCases)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestBoolWithInverseEnvVars(t *testing.T) {
	flagMethod := func() *BoolWithInverseFlag {
		return &BoolWithInverseFlag{
			Name:    "env",
			Sources: EnvVars("ENV", "NO-ENV"),
			Local:   true,
		}
	}

	testCases := []*boolWithInverseTestCase{
		{
			toBeSet: true,
			value:   false,
			envVars: map[string]string{
				"NO-ENV": "false",
			},
		},
		{
			toBeSet: true,
			value:   true,
			envVars: map[string]string{
				"ENV": "true",
			},
		},
		{
			toBeSet: true,
			value:   false,
			envVars: map[string]string{
				"ENV": "false",
			},
		},
		{
			toBeSet: false,
			value:   false,
		},
		// TODO
		/*{
			err: errBothEnvFlagsAreSet,
			envVars: map[string]string{
				"ENV":    "true",
				"NO-ENV": "true",
			},
		},*/
		{
			err: fmt.Errorf("could not parse \"true_env\" as bool value from environment variable \"ENV\" for flag env: parse error"),
			envVars: map[string]string{
				"ENV": "true_env",
			},
		},
		{
			err: fmt.Errorf("could not parse \"false_env\" as bool value from environment variable \"NO-ENV\" for flag env: parse error"),
			envVars: map[string]string{
				"NO-ENV": "false_env",
			},
		},
	}

	err := runBoolWithInverseFlagTests(t, flagMethod, testCases)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestBoolWithInverseWithPrefix(t *testing.T) {
	flagMethod := func() *BoolWithInverseFlag {
		return &BoolWithInverseFlag{
			Name:          "env",
			InversePrefix: "without-",
		}
	}

	testCases := []*boolWithInverseTestCase{
		{
			args:    []string{"--without-env"},
			toBeSet: true,
			value:   false,
		},
		{
			args:    []string{"--env"},
			toBeSet: true,
			value:   true,
		},
		{
			toBeSet: false,
			value:   false,
		},
		{
			args: []string{"--env", "--without-env"},
			err:  fmt.Errorf("cannot set both flags `--env` and `--without-env`"),
		},
		{
			args: []string{"--without-env", "--env"},
			err:  fmt.Errorf("cannot set both flags `--env` and `--without-env`"),
		},
	}

	err := runBoolWithInverseFlagTests(t, flagMethod, testCases)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestBoolWithInverseRequired(t *testing.T) {
	flagMethod := func() *BoolWithInverseFlag {
		return &BoolWithInverseFlag{
			Name:     "env",
			Required: true,
		}
	}

	testCases := []*boolWithInverseTestCase{
		{
			args:    []string{"--no-env"},
			toBeSet: true,
			value:   false,
		},
		{
			args:    []string{"--env"},
			toBeSet: true,
			value:   true,
		},
		{
			args: []string{"--env", "--no-env"},
			err:  errBothEnvFlagsAreSet,
		},
	}

	err := runBoolWithInverseFlagTests(t, flagMethod, testCases)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestBoolWithInverseNames(t *testing.T) {
	flag := &BoolWithInverseFlag{
		Name:     "env",
		Required: true,
	}
	names := flag.Names()

	require.Len(t, names, 2)
	require.Equal(t, "env", names[0], "expected first name to be `env`")
	require.Equal(t, "no-env", names[1], "expected first name to be `no-env`")

	var d DocGenerationFlag = flag
	require.Equal(t, "bool", d.TypeName())
}

func TestBoolWithInverseString(t *testing.T) {
	tcs := []struct {
		testName      string
		flagName      string
		required      bool
		usage         string
		inversePrefix string
		expected      string
	}{
		{
			testName: "empty inverse prefix no flag",
			flagName: "",
			required: true,
			expected: "--[no-]\t",
		},
		{
			testName: "single-char flag name",
			flagName: "e",
			required: true,
			expected: "-[no-]e\t",
		},
		{
			testName: "multi-char flag name",
			flagName: "env",
			required: true,
			expected: "--[no-]env\t",
		},
		{
			testName: "required with usage",
			flagName: "env",
			required: true,
			usage:    "env usage",
			expected: "--[no-]env\tenv usage",
		},
		{
			testName: "required without usage",
			flagName: "env",
			required: true,
			expected: "--[no-]env\t",
		},
		{
			testName: "not required with default usage",
			flagName: "env",
			required: false,
			expected: "--[no-]env\t(default: false)",
		},
		{
			testName:      "custom inverse prefix",
			flagName:      "env",
			required:      true,
			inversePrefix: "nope-",
			expected:      "--[nope-]env\t",
		},
		{
			testName: "empty inverse prefix",
			flagName: "env",
			required: true,
			expected: "--[no-]env\t",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			flag := &BoolWithInverseFlag{
				Name:          tc.flagName,
				Usage:         tc.usage,
				Required:      tc.required,
				InversePrefix: tc.inversePrefix,
			}

			require.Equal(t, tc.expected, flag.String())
		})
	}
}

func TestBoolWithInverseDestination(t *testing.T) {
	destination := new(bool)
	count := new(int)

	flagMethod := func() *BoolWithInverseFlag {
		return &BoolWithInverseFlag{
			Name:        "env",
			Destination: destination,
			Config: BoolConfig{
				Count: count,
			},
		}
	}

	checkAndReset := func(expectedCount int, expectedValue bool) error {
		if *count != expectedCount {
			return fmt.Errorf("expected count to be %d, got %d", expectedCount, *count)
		}

		if *destination != expectedValue {
			return fmt.Errorf("expected destination to be %t, got %t", expectedValue, *destination)
		}

		*count = 0
		*destination = false

		return nil
	}

	err := (&boolWithInverseTestCase{
		args:    []string{"--env"},
		toBeSet: true,
		value:   true,
	}).Run(t, flagMethod())
	if err != nil {
		t.Error(err)
		return
	}

	err = checkAndReset(1, true)
	if err != nil {
		t.Error(err)
		return
	}

	err = (&boolWithInverseTestCase{
		args:    []string{"--no-env"},
		toBeSet: true,
		value:   false,
	}).Run(t, flagMethod())
	if err != nil {
		t.Error(err)
		return
	}

	err = checkAndReset(1, false)
	if err != nil {
		t.Error(err)
		return
	}

	err = (&boolWithInverseTestCase{
		args:    []string{},
		toBeSet: false,
		value:   false,
	}).Run(t, flagMethod())
	if err != nil {
		t.Error(err)
		return
	}

	err = checkAndReset(0, false)
	if err != nil {
		t.Error(err)
		return
	}

	f := flagMethod()
	f.Value = true
	err = (&boolWithInverseTestCase{
		args:    []string{},
		toBeSet: false,
		value:   true,
	}).Run(t, f)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkAndReset(0, true)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestBoolWithInverseFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f RequiredFlag = &BoolWithInverseFlag{}

	_ = f.IsRequired()
}

func TestBoolWithInverseFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f VisibleFlag = &BoolWithInverseFlag{}

	_ = f.IsVisible()
}
