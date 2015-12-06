package inputfilesupport

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/codegangsta/cli"
)

func TestCommandYamlFileTest(t *testing.T) {
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	ioutil.WriteFile("current.yaml", []byte("test: 15"), 0666)
	defer os.Remove("current.yaml")
	test := []string{"test-cmd", "--load", "current.yaml"}
	set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) {
			val := c.Int("test")
			expect(t, val, 15)
		},
		Flags: []cli.Flag{
			NewIntFlag(cli.IntFlag{Name: "test"}),
			cli.StringFlag{Name: "load"}},
	}
	command.Before = InitializeYaml("load", command.Flags)
	err := command.Run(c)

	expect(t, err, nil)
}

func TestCommandYamlFileTestGlobalEnvVarWins(t *testing.T) {
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	ioutil.WriteFile("current.yaml", []byte("test: 15"), 0666)
	defer os.Remove("current.yaml")

	os.Setenv("THE_TEST", "10")
	defer os.Setenv("THE_TEST", "")
	test := []string{"test-cmd", "--load", "current.yaml"}
	set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) {
			val := c.Int("test")
			expect(t, val, 10)
		},
		Flags: []cli.Flag{
			NewIntFlag(cli.IntFlag{Name: "test", EnvVar: "THE_TEST"}),
			cli.StringFlag{Name: "load"}},
	}
	command.Before = InitializeYaml("load", command.Flags)

	err := command.Run(c)

	expect(t, err, nil)
}

func TestCommandYamlFileTestSpecifiedFlagWins(t *testing.T) {
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	ioutil.WriteFile("current.yaml", []byte("test: 15"), 0666)
	defer os.Remove("current.yaml")

	test := []string{"test-cmd", "--load", "current.yaml", "--test", "7"}
	set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) {
			val := c.Int("test")
			expect(t, val, 7)
		},
		Flags: []cli.Flag{
			NewIntFlag(cli.IntFlag{Name: "test"}),
			cli.StringFlag{Name: "load"}},
	}
	command.Before = InitializeYaml("load", command.Flags)

	err := command.Run(c)

	expect(t, err, nil)
}

func TestCommandYamlFileTestDefaultValueFileWins(t *testing.T) {
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	ioutil.WriteFile("current.yaml", []byte("test: 15"), 0666)
	defer os.Remove("current.yaml")

	test := []string{"test-cmd", "--load", "current.yaml"}
	set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) {
			val := c.Int("test")
			expect(t, val, 15)
		},
		Flags: []cli.Flag{
			NewIntFlag(cli.IntFlag{Name: "test", Value: 7}),
			cli.StringFlag{Name: "load"}},
	}
	command.Before = InitializeYaml("load", command.Flags)

	err := command.Run(c)

	expect(t, err, nil)
}

func TestCommandYamlFileFlagHasDefaultGlobalEnvYamlSetGlobalEnvWins(t *testing.T) {
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	ioutil.WriteFile("current.yaml", []byte("test: 15"), 0666)
	defer os.Remove("current.yaml")

	os.Setenv("THE_TEST", "11")
	defer os.Setenv("THE_TEST", "")

	test := []string{"test-cmd", "--load", "current.yaml"}
	set.Parse(test)

	c := cli.NewContext(app, set, nil)

	command := &cli.Command{
		Name:        "test-cmd",
		Aliases:     []string{"tc"},
		Usage:       "this is for testing",
		Description: "testing",
		Action: func(c *cli.Context) {
			val := c.Int("test")
			expect(t, val, 11)
		},
		Flags: []cli.Flag{
			NewIntFlag(cli.IntFlag{Name: "test", Value: 7, EnvVar: "THE_TEST"}),
			cli.StringFlag{Name: "load"}},
	}
	command.Before = InitializeYaml("load", command.Flags)
	err := command.Run(c)

	expect(t, err, nil)
}
