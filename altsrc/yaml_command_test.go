// Disabling building of yaml support in cases where golang is 1.0 or 1.1
// as the encoding library is not implemented or supported.

// +build go1.2

package altsrc

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/codegangsta/cli"
)

func TestCommandYamlFileTest(t *testing.T) {
	app := cli.NewApp()
	set := cli.NewFlagSet("test",
		[]cli.Flag{
			NewIntFlag(&cli.IntFlag{Name: "test"}),
			&cli.StringFlag{Name: "load"},
		}, []string{"test-cmd", "--load", "current.yaml"})
	err := set.Parse()
	expect(t, err, nil)

	ioutil.WriteFile("current.yaml", []byte("test: 15"), 0666)
	defer os.Remove("current.yaml")

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Before:      InitInputSourceWithContext(set.Flags, NewYamlSourceFromFlagFunc("load")),
		Action: func(c *cli.Context) error {
			val := c.Int("test")
			expect(t, val, 15)
			return nil
		},
		Flags: set.Flags,
	}
	err = command.Run(c)
	expect(t, err, nil)
}

func TestCommandYamlFileTestGlobalEnvVarWins(t *testing.T) {
	app := cli.NewApp()
	set := cli.NewFlagSet("test", []cli.Flag{
		NewIntFlag(&cli.IntFlag{Name: "test", EnvVars: []string{"THE_TEST"}}),
		&cli.StringFlag{Name: "load"},
	}, []string{"test-cmd", "--load", "current.yaml"})
	err := set.Parse()
	expect(t, err, nil)

	ioutil.WriteFile("current.yaml", []byte("test: 15"), 0666)
	defer os.Remove("current.yaml")

	os.Setenv("THE_TEST", "10")
	defer os.Setenv("THE_TEST", "")

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Before:      InitInputSourceWithContext(set.Flags, NewYamlSourceFromFlagFunc("load")),
		Action: func(c *cli.Context) error {
			val := c.Int("test")
			expect(t, val, 10)
			return nil
		},
		Flags: set.Flags,
	}
	err = command.Run(c)
	expect(t, err, nil)
}

func TestCommandYamlFileTestGlobalEnvVarWinsNested(t *testing.T) {
	app := cli.NewApp()
	set := cli.NewFlagSet("test", []cli.Flag{
		NewIntFlag(&cli.IntFlag{Name: "top.test", EnvVars: []string{"THE_TEST"}}),
		&cli.StringFlag{Name: "load"},
	}, []string{"test-cmd", "--load", "current.yaml"})
	err := set.Parse()
	expect(t, err, nil)

	ioutil.WriteFile("current.yaml", []byte("top:\n  test: 15"), 0666)
	defer os.Remove("current.yaml")

	os.Setenv("THE_TEST", "10")
	defer os.Setenv("THE_TEST", "")

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Before:      InitInputSourceWithContext(set.Flags, NewYamlSourceFromFlagFunc("load")),
		Action: func(c *cli.Context) error {
			val := c.Int("top.test")
			expect(t, val, 10)
			return nil
		},
		Flags: set.Flags,
	}

	err = command.Run(c)
	expect(t, err, nil)
}

func TestCommandYamlFileTestSpecifiedFlagWins(t *testing.T) {
	app := cli.NewApp()
	set := cli.NewFlagSet("test", []cli.Flag{
		NewIntFlag(&cli.IntFlag{Name: "test"}),
		&cli.StringFlag{Name: "load"},
	}, []string{"test-cmd", "--load", "current.yaml", "--test", "7"})
	err := set.Parse()
	expect(t, err, nil)

	ioutil.WriteFile("current.yaml", []byte("test: 15"), 0666)
	defer os.Remove("current.yaml")

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Before:      InitInputSourceWithContext(set.Flags, NewYamlSourceFromFlagFunc("load")),
		Action: func(c *cli.Context) error {
			val := c.Int("test")
			expect(t, val, 7)
			return nil
		},
		Flags: set.Flags,
	}

	err = command.Run(c)
	expect(t, err, nil)
}

func TestCommandYamlFileTestSpecifiedFlagWinsNested(t *testing.T) {
	app := cli.NewApp()
	set := cli.NewFlagSet("test", []cli.Flag{
		NewIntFlag(&cli.IntFlag{Name: "top.test"}),
		&cli.StringFlag{Name: "load"},
	}, []string{"test-cmd", "--load", "current.yaml", "--top.test", "7"})
	err := set.Parse()
	expect(t, err, nil)

	ioutil.WriteFile("current.yaml", []byte("top:\n  test: 15"), 0666)
	defer os.Remove("current.yaml")

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Before:      InitInputSourceWithContext(set.Flags, NewYamlSourceFromFlagFunc("load")),
		Action: func(c *cli.Context) error {
			val := c.Int("top.test")
			expect(t, val, 7)
			return nil
		},
		Flags: set.Flags,
	}

	err = command.Run(c)
	expect(t, err, nil)
}

func TestCommandYamlFileTestDefaultValueFileWins(t *testing.T) {
	app := cli.NewApp()
	set := cli.NewFlagSet("test", []cli.Flag{
		NewIntFlag(&cli.IntFlag{Name: "test", Value: 7}),
		&cli.StringFlag{Name: "load"},
	}, []string{"test-cmd", "--load", "current.yaml"})
	err := set.Parse()
	expect(t, err, nil)

	ioutil.WriteFile("current.yaml", []byte("test: 15"), 0666)
	defer os.Remove("current.yaml")

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Before:      InitInputSourceWithContext(set.Flags, NewYamlSourceFromFlagFunc("load")),
		Action: func(c *cli.Context) error {
			val := c.Int("test")
			expect(t, val, 15)
			return nil
		},
		Flags: set.Flags,
	}

	err = command.Run(c)
	expect(t, err, nil)
}

func TestCommandYamlFileTestDefaultValueFileWinsNested(t *testing.T) {
	app := cli.NewApp()
	set := cli.NewFlagSet("test", []cli.Flag{
		NewIntFlag(&cli.IntFlag{Name: "top.test", Value: 7}),
		&cli.StringFlag{Name: "load"},
	}, []string{"test-cmd", "--load", "current.yaml"})
	err := set.Parse()
	expect(t, err, nil)

	ioutil.WriteFile("current.yaml", []byte("top:\n  test: 15"), 0666)
	defer os.Remove("current.yaml")

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Before:      InitInputSourceWithContext(set.Flags, NewYamlSourceFromFlagFunc("load")),
		Action: func(c *cli.Context) error {
			val := c.Int("top.test")
			expect(t, val, 15)
			return nil
		},
		Flags: set.Flags,
	}

	err = command.Run(c)
	expect(t, err, nil)
}

func TestCommandYamlFileFlagHasDefaultGlobalEnvYamlSetGlobalEnvWins(t *testing.T) {
	app := cli.NewApp()
	set := cli.NewFlagSet("test", []cli.Flag{
		NewIntFlag(&cli.IntFlag{Name: "test", Value: 7, EnvVars: []string{"THE_TEST"}}),
		&cli.StringFlag{Name: "load"},
	}, []string{"test-cmd", "--load", "current.yaml"})
	err := set.Parse()
	expect(t, err, nil)
	ioutil.WriteFile("current.yaml", []byte("test: 15"), 0666)
	defer os.Remove("current.yaml")

	os.Setenv("THE_TEST", "11")
	defer os.Setenv("THE_TEST", "")

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Before:      InitInputSourceWithContext(set.Flags, NewYamlSourceFromFlagFunc("load")),
		Action: func(c *cli.Context) error {
			val := c.Int("test")
			expect(t, val, 11)
			return nil
		},
		Flags: set.Flags,
	}

	err = command.Run(c)
	expect(t, err, nil)
}

func TestCommandYamlFileFlagHasDefaultGlobalEnvYamlSetGlobalEnvWinsNested(t *testing.T) {
	app := cli.NewApp()
	set := cli.NewFlagSet("test", []cli.Flag{
		NewIntFlag(&cli.IntFlag{Name: "top.test", Value: 7, EnvVars: []string{"THE_TEST"}}),
		&cli.StringFlag{Name: "load"},
	}, []string{"test-cmd", "--load", "current.yaml"})
	err := set.Parse()
	expect(t, err, nil)

	ioutil.WriteFile("current.yaml", []byte("top:\n  test: 15"), 0666)
	defer os.Remove("current.yaml")

	os.Setenv("THE_TEST", "11")
	defer os.Setenv("THE_TEST", "")

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Before:      InitInputSourceWithContext(set.Flags, NewYamlSourceFromFlagFunc("load")),
		Action: func(c *cli.Context) error {
			val := c.Int("top.test")
			expect(t, val, 11)
			return nil
		},
		Flags: set.Flags,
	}

	err = command.Run(c)
	expect(t, err, nil)
}
