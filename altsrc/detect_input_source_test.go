package altsrc

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestDetectsConfCorrectly(t *testing.T) {
	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	_ = os.WriteFile("current.conf", []byte("test = 15"), 0666)
	defer os.Remove("current.conf")
	test := []string{"test-cmd", "--load", "current.conf"}
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
			NewIntFlag(&cli.IntFlag{Name: "test"}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, DetectNewSourceFromFlagFunc("load"))
	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestDetectsJsonCorrectly(t *testing.T) {
	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	_ = os.WriteFile("current.json", []byte("{\"test\":15}"), 0666)
	defer os.Remove("current.json")
	test := []string{"test-cmd", "--load", "current.json"}
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
			NewIntFlag(&cli.IntFlag{Name: "test"}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, DetectNewSourceFromFlagFunc("load"))
	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestDetectsTomlCorrectly(t *testing.T) {
	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	_ = os.WriteFile("current.toml", []byte("test = 15"), 0666)
	defer os.Remove("current.toml")
	test := []string{"test-cmd", "--load", "current.toml"}
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
			NewIntFlag(&cli.IntFlag{Name: "test"}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, DetectNewSourceFromFlagFunc("load"))
	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestDetectsYamlCorrectly(t *testing.T) {
	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	_ = os.WriteFile("current.yaml", []byte("test: 15"), 0666)
	defer os.Remove("current.yaml")
	test := []string{"test-cmd", "--load", "current.yaml"}
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
			NewIntFlag(&cli.IntFlag{Name: "test"}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, DetectNewSourceFromFlagFunc("load"))
	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestDetectsYmlCorrectly(t *testing.T) {
	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	_ = os.WriteFile("current.yml", []byte("test: 15"), 0666)
	defer os.Remove("current.yml")
	test := []string{"test-cmd", "--load", "current.yml"}
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
			NewIntFlag(&cli.IntFlag{Name: "test"}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, DetectNewSourceFromFlagFunc("load"))
	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestHandlesCustomTypeCorrectly(t *testing.T) {
	app := &cli.App{}
	app.RegisterDetectableSource(".custom", NewYamlSourceFromFlagFunc)
	set := flag.NewFlagSet("test", 0)
	_ = os.WriteFile("current.custom", []byte("test: 15"), 0666)
	defer os.Remove("current.custom")
	test := []string{"test-cmd", "--load", "current.custom"}
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
			NewIntFlag(&cli.IntFlag{Name: "test"}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, DetectNewSourceFromFlagFunc("load"))
	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestAllowsOverrides(t *testing.T) {
	app := &cli.App{}
	app.RegisterDetectableSource(".conf", NewYamlSourceFromFlagFunc)
	set := flag.NewFlagSet("test", 0)
	_ = os.WriteFile("current.conf", []byte("test: 15"), 0666)
	defer os.Remove("current.conf")
	test := []string{"test-cmd", "--load", "current.conf"}
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
			NewIntFlag(&cli.IntFlag{Name: "test"}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, DetectNewSourceFromFlagFunc("load"))
	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestFailsOnUnrocegnized(t *testing.T) {
	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	_ = os.WriteFile("current.fake", []byte("test: 15"), 0666)
	defer os.Remove("current.fake")
	test := []string{"test-cmd", "--load", "current.fake"}
	_ = set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) error {
			val := c.Int("test")
			expect(t, val, 0)
			return nil
		},
		Flags: []cli.Flag{
			NewIntFlag(&cli.IntFlag{Name: "test"}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, DetectNewSourceFromFlagFunc("load"))
	err := command.Run(c, test...)

	expect(t, err, fmt.Errorf("Unable to create input source with context: inner error: \n'Unable to determine config file type from extension.\nMust be one of [.conf .json .toml .yaml .yml]'"))
}

func TestSilentNoOpWithoutFlag(t *testing.T) {
	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	_ = os.WriteFile("current.conf", []byte("test = 15"), 0666)
	defer os.Remove("current.conf")
	test := []string{"test-cmd"}
	_ = set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) error {
			val := c.Int("test")
			expect(t, val, 0)
			return nil
		},
		Flags: []cli.Flag{
			NewIntFlag(&cli.IntFlag{Name: "test"}),
			&cli.StringFlag{Name: "load"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, DetectNewSourceFromFlagFunc("load"))
	err := command.Run(c, test...)

	expect(t, err, nil)
}

func TestLoadDefaultConfig(t *testing.T) {
	t.Skip("Fix parent implementation for default Flag values to get this working")

	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	_ = os.WriteFile("current.conf", []byte("test = 15"), 0666)
	defer os.Remove("current.conf")
	test := []string{"test-cmd"}
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
			NewIntFlag(&cli.IntFlag{Name: "test"}),
			&cli.StringFlag{Name: "load", Value: "current.conf"}},
	}
	command.Before = InitInputSourceWithContext(command.Flags, DetectNewSourceFromFlagFunc("load"))
	err := command.Run(c, test...)

	expect(t, err, nil)
}
