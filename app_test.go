package cli_test

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"testing"
)

func ExampleApp() {
	// set args for examples sake
	os.Args = []string{"greet", "--name", "Jeremy"}

	app := cli.NewApp()
	app.Name = "greet"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "name", Value: "bob", Usage: "a name to say"},
	}
	app.Action = func(c *cli.Context) {
		fmt.Printf("Hello %v\n", c.String("name"))
	}
	app.Run(os.Args)
	// Output:
	// Hello Jeremy
}

func TestApp_Run(t *testing.T) {
	s := ""

	app := cli.NewApp()
	app.Action = func(c *cli.Context) {
		s = s + c.Args().First()
	}

	err := app.Run([]string{"command", "foo"})
	expect(t, err, nil)
	err = app.Run([]string{"command", "bar"})
	expect(t, err, nil)
	expect(t, s, "foobar")
}

var commandAppTests = []struct {
	name     string
	expected bool
}{
	{"foobar", true},
	{"batbaz", true},
	{"b", true},
	{"f", true},
	{"bat", false},
	{"nothing", false},
}

func TestApp_Command(t *testing.T) {
	app := cli.NewApp()
	fooCommand := cli.Command{Name: "foobar", ShortName: "f"}
	batCommand := cli.Command{Name: "batbaz", ShortName: "b"}
	app.Commands = []cli.Command{
		fooCommand,
		batCommand,
	}

	for _, test := range commandAppTests {
		expect(t, app.Command(test.name) != nil, test.expected)
	}
}

func TestApp_CommandWithArgBeforeFlags(t *testing.T) {
	var parsedOption, firstArg string

	app := cli.NewApp()
	command := cli.Command{
		Name: "cmd",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "option", Value: "", Usage: "some option"},
		},
		Action: func(c *cli.Context) {
			parsedOption = c.String("option")
			firstArg = c.Args().First()
		},
	}
	app.Commands = []cli.Command{command}

	app.Run([]string{"", "cmd", "my-arg", "--option", "my-option"})

	expect(t, parsedOption, "my-option")
	expect(t, firstArg, "my-arg")
}

func TestApp_Float64Flag(t *testing.T) {
	var meters float64

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.Float64Flag{Name: "height", Value: 1.5, Usage: "Set the height, in meters"},
	}
	app.Action = func(c *cli.Context) {
		meters = c.Float64("height")
	}

	app.Run([]string{"", "--height", "1.93"})
	expect(t, meters, 1.93)
}

func TestApp_ParseSliceFlags(t *testing.T) {
	var parsedOption, firstArg string
	var parsedIntSlice []int
	var parsedStringSlice []string

	app := cli.NewApp()
	command := cli.Command{
		Name: "cmd",
		Flags: []cli.Flag{
			cli.IntSliceFlag{Name: "p", Value: &cli.IntSlice{}, Usage: "set one or more ip addr"},
			cli.StringSliceFlag{Name: "ip", Value: &cli.StringSlice{}, Usage: "set one or more ports to open"},
		},
		Action: func(c *cli.Context) {
			parsedIntSlice = c.IntSlice("p")
			parsedStringSlice = c.StringSlice("ip")
			parsedOption = c.String("option")
			firstArg = c.Args().First()
		},
	}
	app.Commands = []cli.Command{command}

	app.Run([]string{"", "cmd", "my-arg", "-p", "22", "-p", "80", "-ip", "8.8.8.8", "-ip", "8.8.4.4"})

	IntsEquals := func(a, b []int) bool {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
		return true
	}

	StrsEquals := func(a, b []string) bool {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
		return true
	}
	var expectedIntSlice = []int{22, 80}
	var expectedStringSlice = []string{"8.8.8.8", "8.8.4.4"}

	if !IntsEquals(parsedIntSlice, expectedIntSlice) {
		t.Errorf("%s does not match %s", parsedIntSlice, expectedIntSlice)
	}

	if !StrsEquals(parsedStringSlice, expectedStringSlice) {
		t.Errorf("%s does not match %s", parsedStringSlice, expectedStringSlice)
	}
}
