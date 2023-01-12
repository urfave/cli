package cli_test

import (
	"fmt"
	"testing"

	"github.com/urfave/cli/v3"
)

type boolWithInverseTestCase struct {
	args    []string
	toBeSet bool
	value   bool
}

func (test boolWithInverseTestCase) Run() error {
	flagWithInverse := cli.NewBoolWithInverse(cli.BoolFlag{
		Name: "env",
	})

	app := cli.App{
		Flags: []cli.Flag{
			flagWithInverse,
		},
		Action: func(ctx *cli.Context) error {
			return nil
		},
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

func TestBoolWithInverse(t *testing.T) {
	err := boolWithInverseTestCase{
		args:    []string{"--no-env"},
		toBeSet: true,
		value:   false,
	}.Run()
	if err != nil {
		t.Error(err)
		return
	}

	err = boolWithInverseTestCase{
		args:    []string{"--env"},
		toBeSet: true,
		value:   true,
	}.Run()
	if err != nil {
		t.Error(err)
		return
	}

	err = boolWithInverseTestCase{
		toBeSet: false,
		value:   false,
	}.Run()
	if err != nil {
		t.Error(err)
		return
	}

	expectedError := fmt.Errorf("cannot set both flags `--env` and `--no-env`")
	err = boolWithInverseTestCase{
		args: []string{"--env", "--no-env"},
	}.Run()
	if err == nil || err.Error() != expectedError.Error() {
		t.Errorf("expected error %q, but got %q", expectedError, err)
		return
	}
}

func ExampleNewBoolWithInverse() {
	flagWithInverse := cli.NewBoolWithInverse(cli.BoolFlag{
		Name: "env",
	})

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

	// Output:
	// no-env is set
	// env is set
}
