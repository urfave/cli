package cli_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v3"
)

var (
	bothEnvFlagsAreSetError = fmt.Errorf("cannot set both flags `--env` and `--no-env`")
)

type boolWithInverseTestCase struct {
	args    []string
	toBeSet bool
	value   bool
	err     error
	envVars map[string]string
}

func (tc *boolWithInverseTestCase) Run(t *testing.T, flagWithInverse *cli.BoolWithInverseFlag) error {
	cmd := &cli.Command{
		Flags:  []cli.Flag{flagWithInverse},
		Action: func(ctx *cli.Context) error { return nil },
	}

	for key, val := range tc.envVars {
		os.Setenv(key, val)
		defer os.Unsetenv(key)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, append([]string{"prog"}, tc.args...))
	if err != nil {
		return err
	}

	if flagWithInverse.IsSet() != tc.toBeSet {
		return fmt.Errorf("flag should be set %t, but got %t", tc.toBeSet, flagWithInverse.IsSet())
	}

	if flagWithInverse.Value() != tc.value {
		return fmt.Errorf("flag value should be %t, but got %t", tc.value, flagWithInverse.Value())
	}

	return nil
}

func runTests(t *testing.T, newFlagMethod func() *cli.BoolWithInverseFlag, cases []*boolWithInverseTestCase) error {
	for _, test := range cases {
		fl := newFlagMethod()

		err := test.Run(t, fl)
		if err != nil && test.err == nil {
			return err
		}

		if err == nil && test.err != nil {
			return fmt.Errorf("expected error %q, but got nil", test.err)
		}

		if err != nil && test.err != nil && err.Error() != test.err.Error() {
			return fmt.Errorf("expected error %q, but got %q", test.err, err)
		}

	}

	return nil
}

func TestBoolWithInverseBasic(t *testing.T) {
	flagMethod := func() *cli.BoolWithInverseFlag {
		return &cli.BoolWithInverseFlag{
			BoolFlag: &cli.BoolFlag{
				Name: "env",
			},
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
			err:  bothEnvFlagsAreSetError,
		},
	}

	err := runTests(t, flagMethod, testCases)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestBoolWithInverseAction(t *testing.T) {
	flagMethod := func() *cli.BoolWithInverseFlag {
		return &cli.BoolWithInverseFlag{
			BoolFlag: &cli.BoolFlag{
				Name: "env",

				// Setting env to the opposite to test flag Action is working as intended
				Action: func(ctx *cli.Context, value bool) error {
					if value {
						return ctx.Set("env", "false")
					}

					return ctx.Set("env", "true")
				},
			},
		}
	}

	testCases := []*boolWithInverseTestCase{
		{
			args:    []string{"--no-env"},
			toBeSet: true,
			value:   true,
		},
		{
			args:    []string{"--env"},
			toBeSet: true,
			value:   false,
		},

		// This test is not inverse because the flag action is never called
		{
			toBeSet: false,
			value:   false,
		},
		{
			args: []string{"--env", "--no-env"},
			err:  bothEnvFlagsAreSetError,
		},
	}

	err := runTests(t, flagMethod, testCases)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestBoolWithInverseAlias(t *testing.T) {
	flagMethod := func() *cli.BoolWithInverseFlag {
		return &cli.BoolWithInverseFlag{
			BoolFlag: &cli.BoolFlag{
				Name:    "env",
				Aliases: []string{"e", "do-env"},
			},
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
			err:  bothEnvFlagsAreSetError,
		},
	}

	err := runTests(t, flagMethod, testCases)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestBoolWithInverseEnvVars(t *testing.T) {
	flagMethod := func() *cli.BoolWithInverseFlag {
		return &cli.BoolWithInverseFlag{
			BoolFlag: &cli.BoolFlag{
				Name:    "env",
				Sources: cli.ValueSources{cli.EnvSource("ENV")},
			},
		}
	}

	testCases := []*boolWithInverseTestCase{
		{
			toBeSet: true,
			value:   false,
			envVars: map[string]string{
				"NO-ENV": "true",
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
		{
			err: bothEnvFlagsAreSetError,
			envVars: map[string]string{
				"ENV":    "true",
				"NO-ENV": "true",
			},
		},
	}

	err := runTests(t, flagMethod, testCases)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestBoolWithInverseWithPrefix(t *testing.T) {
	flagMethod := func() *cli.BoolWithInverseFlag {
		return &cli.BoolWithInverseFlag{
			BoolFlag: &cli.BoolFlag{
				Name: "env",
			},
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
	}

	err := runTests(t, flagMethod, testCases)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestBoolWithInverseRequired(t *testing.T) {
	flagMethod := func() *cli.BoolWithInverseFlag {
		return &cli.BoolWithInverseFlag{
			BoolFlag: &cli.BoolFlag{
				Name:     "env",
				Required: true,
			},
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
			err:     fmt.Errorf(`Required flag "env" not set`),
		},
		{
			args: []string{"--env", "--no-env"},
			err:  bothEnvFlagsAreSetError,
		},
	}

	err := runTests(t, flagMethod, testCases)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestBoolWithInverseNames(t *testing.T) {
	flag := &cli.BoolWithInverseFlag{
		BoolFlag: &cli.BoolFlag{
			Name:     "env",
			Required: true,
		},
	}
	names := flag.Names()

	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
		return
	}

	if names[0] != "env" {
		t.Errorf("expected first name to be `env`, got `%s`", names[0])
		return
	}

	if names[1] != "no-env" {
		t.Errorf("expected first name to be `no-env`, got `%s`", names[1])
		return
	}

	flagString := flag.String()
	if strings.Contains(flagString, "--env") == false {
		t.Errorf("expected `%s` to contain `--env`", flagString)
		return
	}

	if strings.Contains(flagString, "--no-env") == false {
		t.Errorf("expected `%s` to contain `--no-env`", flagString)
		return
	}
}

func TestBoolWithInverseDestination(t *testing.T) {
	destination := new(bool)
	count := new(int)

	flagMethod := func() *cli.BoolWithInverseFlag {
		return &cli.BoolWithInverseFlag{
			BoolFlag: &cli.BoolFlag{
				Name:        "env",
				Destination: destination,
				Config: cli.BoolConfig{
					Count: count,
				},
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
}

func ExampleBoolWithInverseFlag() {
	flagWithInverse := &cli.BoolWithInverseFlag{
		BoolFlag: &cli.BoolFlag{
			Name: "env",
		},
	}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			flagWithInverse,
		},
		Action: func(ctx *cli.Context) error {
			if flagWithInverse.IsSet() {
				if flagWithInverse.Value() {
					fmt.Println("env is set")
				} else {
					fmt.Println("no-env is set")
				}
			}

			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = cmd.Run(ctx, []string{"prog", "--no-env"})
	_ = cmd.Run(ctx, []string{"prog", "--env"})

	fmt.Println("flags:", len(flagWithInverse.Flags()))

	// Output:
	// no-env is set
	// env is set
	// flags: 2
}
