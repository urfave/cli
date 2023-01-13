package cli_test

import (
	"fmt"
	"os"
	"testing"

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

func (test boolWithInverseTestCase) Run(flagWithInverse *cli.BoolWithInverseFlag) error {
	app := cli.App{
		Flags:  []cli.Flag{flagWithInverse},
		Action: func(ctx *cli.Context) error { return nil },
	}

	for key, val := range test.envVars {
		os.Setenv(key, val)
		defer os.Unsetenv(key)
	}

	err := app.Run(append([]string{"prog"}, test.args...))
	if err != nil {
		return err
	}

	if flagWithInverse.IsSet() != test.toBeSet {
		return fmt.Errorf("flag should be set %t, but got %t", test.toBeSet, flagWithInverse.IsSet())
	}

	if flagWithInverse.Value() != test.value {
		return fmt.Errorf("flag value should be %t, but got %t", test.value, flagWithInverse.Value())
	}

	return nil
}

func runTests(newFlagMethod func() *cli.BoolWithInverseFlag, cases []boolWithInverseTestCase) error {
	for _, test := range cases {
		flag := newFlagMethod()

		err := test.Run(flag)
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

	testCases := []boolWithInverseTestCase{
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

	err := runTests(flagMethod, testCases)
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

	testCases := []boolWithInverseTestCase{
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

	err := runTests(flagMethod, testCases)
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

	testCases := []boolWithInverseTestCase{
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

	err := runTests(flagMethod, testCases)
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
				EnvVars: []string{"ENV"},
			},
		}
	}

	testCases := []boolWithInverseTestCase{
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

	err := runTests(flagMethod, testCases)
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

	testCases := []boolWithInverseTestCase{
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

	err := runTests(flagMethod, testCases)
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

	app := cli.App{
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

	_ = app.Run([]string{"prog", "--no-env"})
	_ = app.Run([]string{"prog", "--env"})

	fmt.Println("flags:", len(flagWithInverse.Flags()))

	// Output:
	// no-env is set
	// env is set
	// flags: 2
}
