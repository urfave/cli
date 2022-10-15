package altsrc

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/urfave/cli/v2"
)

const (
	fileName   = "current.json"
	simpleJSON = `{"test": 15, "testb": false}`
	nestedJSON = `{"top": {"test": 15}}`
)

func TestCommandJSONFileTest(t *testing.T) {
	cleanup := writeTempFile(t, fileName, simpleJSON)
	defer cleanup()

	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	test := []string{"test-cmd", "--load", fileName}
	_ = set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) error {
			val := c.Int("test")
			expect(t, val, 15)

			valb := c.Bool("testb")
			expect(t, valb, false)
			return nil
		},
		Flags: []cli.Flag{
			NewIntFlag(&cli.IntFlag{Name: "test"}),
			&cli.StringFlag{Name: "load"},
			NewBoolFlag(&cli.BoolFlag{Name: "testb", Value: true}),
		},
	}
	command.Before = InitInputSourceWithContext(command.Flags, NewJSONSourceFromFlagFunc("load"))
	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestCommandJSONFileTestGlobalEnvVarWins(t *testing.T) {
	cleanup := writeTempFile(t, fileName, simpleJSON)
	defer cleanup()

	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	_ = os.Setenv("THE_TEST", "10")
	defer os.Setenv("THE_TEST", "")

	test := []string{"test-cmd", "--load", fileName}
	_ = set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) error {
			val := c.Int("test")
			expect(t, val, 10)
			return nil
		},
		Flags: []cli.Flag{
			NewIntFlag(&cli.IntFlag{Name: "test", EnvVars: []string{"THE_TEST"}}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, NewJSONSourceFromFlagFunc("load"))

	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestCommandJSONFileTestGlobalEnvVarWinsNested(t *testing.T) {
	cleanup := writeTempFile(t, fileName, nestedJSON)
	defer cleanup()

	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	_ = os.Setenv("THE_TEST", "10")
	defer os.Setenv("THE_TEST", "")

	test := []string{"test-cmd", "--load", fileName}
	_ = set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) error {
			val := c.Int("top.test")
			expect(t, val, 10)
			return nil
		},
		Flags: []cli.Flag{
			NewIntFlag(&cli.IntFlag{Name: "top.test", EnvVars: []string{"THE_TEST"}}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, NewJSONSourceFromFlagFunc("load"))

	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestCommandJSONFileTestSpecifiedFlagWins(t *testing.T) {
	cleanup := writeTempFile(t, fileName, simpleJSON)
	defer cleanup()

	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	test := []string{"test-cmd", "--load", fileName, "--test", "7"}
	_ = set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) error {
			val := c.Int("test")
			expect(t, val, 7)
			return nil
		},
		Flags: []cli.Flag{
			NewIntFlag(&cli.IntFlag{Name: "test"}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, NewJSONSourceFromFlagFunc("load"))

	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestCommandJSONFileTestSpecifiedFlagWinsNested(t *testing.T) {
	cleanup := writeTempFile(t, fileName, nestedJSON)
	defer cleanup()

	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	test := []string{"test-cmd", "--load", fileName, "--top.test", "7"}
	_ = set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) error {
			val := c.Int("top.test")
			expect(t, val, 7)
			return nil
		},
		Flags: []cli.Flag{
			NewIntFlag(&cli.IntFlag{Name: "top.test"}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, NewJSONSourceFromFlagFunc("load"))

	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestCommandJSONFileTestDefaultValueFileWins(t *testing.T) {
	cleanup := writeTempFile(t, fileName, simpleJSON)
	defer cleanup()

	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	test := []string{"test-cmd", "--load", fileName}
	_ = set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) error {
			val := c.Int("test")
			expect(t, val, 15)
			return nil
		},
		Flags: []cli.Flag{
			NewIntFlag(&cli.IntFlag{Name: "test", Value: 7}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, NewJSONSourceFromFlagFunc("load"))

	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestCommandJSONFileTestDefaultValueFileWinsNested(t *testing.T) {
	cleanup := writeTempFile(t, fileName, nestedJSON)
	defer cleanup()

	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	test := []string{"test-cmd", "--load", fileName}
	_ = set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) error {
			val := c.Int("top.test")
			expect(t, val, 15)
			return nil
		},
		Flags: []cli.Flag{
			NewIntFlag(&cli.IntFlag{Name: "top.test", Value: 7}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, NewJSONSourceFromFlagFunc("load"))

	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestCommandJSONFileFlagHasDefaultGlobalEnvJSONSetGlobalEnvWins(t *testing.T) {
	cleanup := writeTempFile(t, fileName, simpleJSON)
	defer cleanup()

	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	_ = os.Setenv("THE_TEST", "11")
	defer os.Setenv("THE_TEST", "")

	test := []string{"test-cmd", "--load", fileName}
	_ = set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) error {
			val := c.Int("test")
			expect(t, val, 11)
			return nil
		},
		Flags: []cli.Flag{
			NewIntFlag(&cli.IntFlag{Name: "test", Value: 7, EnvVars: []string{"THE_TEST"}}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, NewJSONSourceFromFlagFunc("load"))
	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestCommandJSONFileFlagHasDefaultGlobalEnvJSONSetGlobalEnvWinsNested(t *testing.T) {
	cleanup := writeTempFile(t, fileName, nestedJSON)
	defer cleanup()

	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	_ = os.Setenv("THE_TEST", "11")
	defer os.Setenv("THE_TEST", "")

	test := []string{"test-cmd", "--load", fileName}
	_ = set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) error {
			val := c.Int("top.test")
			expect(t, val, 11)
			return nil
		},
		Flags: []cli.Flag{
			NewIntFlag(&cli.IntFlag{Name: "top.test", Value: 7, EnvVars: []string{"THE_TEST"}}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, NewJSONSourceFromFlagFunc("load"))
	err := command.Run(c, test...)

	expect(t, err, nil)
}

func writeTempFile(t *testing.T, name string, content string) func() {
	if err := ioutil.WriteFile(name, []byte(content), 0666); err != nil {
		t.Fatalf("cannot write %q: %v", name, err)
	}
	return func() {
		if err := os.Remove(name); err != nil {
			t.Errorf("cannot remove %q: %v", name, err)
		}
	}
}
