package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

var (
	lastExitCode = 0
	fakeOsExiter = func(rc int) {
		lastExitCode = rc
	}
	fakeErrWriter = &bytes.Buffer{}
)

func init() {
	OsExiter = fakeOsExiter
	ErrWriter = fakeErrWriter
}

type opCounts struct {
	Total, ShellComplete, OnUsageError, Before, CommandNotFound, Action, After, SubCommand int
}

func ExampleApp_Run() {
	// set args for examples sake
	os.Args = []string{"greet", "--name", "Jeremy"}

	app := &App{
		Name: "greet",
		Flags: []Flag{
			&StringFlag{Name: "name", Value: "bob", Usage: "a name to say"},
		},
		Action: func(c *Context) error {
			fmt.Printf("Hello %v\n", c.String("name"))
			return nil
		},
		UsageText: "app [first_arg] [second_arg]",
		Authors:   []*Author{{Name: "Oliver Allen", Email: "oliver@toyshop.example.com"}},
	}

	app.Run(os.Args)
	// Output:
	// Hello Jeremy
}

func ExampleApp_Run_subcommand() {
	// set args for examples sake
	os.Args = []string{"say", "hi", "english", "--name", "Jeremy"}
	app := &App{
		Name: "say",
		Commands: []*Command{
			{
				Name:        "hello",
				Aliases:     []string{"hi"},
				Usage:       "use it to see a description",
				Description: "This is how we describe hello the function",
				Subcommands: []*Command{
					{
						Name:        "english",
						Aliases:     []string{"en"},
						Usage:       "sends a greeting in english",
						Description: "greets someone in english",
						Flags: []Flag{
							&StringFlag{
								Name:  "name",
								Value: "Bob",
								Usage: "Name of the person to greet",
							},
						},
						Action: func(c *Context) error {
							fmt.Println("Hello,", c.String("name"))
							return nil
						},
					},
				},
			},
		},
	}

	_ = app.Run(os.Args)
	// Output:
	// Hello, Jeremy
}

func ExampleApp_Run_appHelp() {
	// set args for examples sake
	os.Args = []string{"greet", "help"}

	app := &App{
		Name:        "greet",
		Version:     "0.1.0",
		Description: "This is how we describe greet the app",
		Authors: []*Author{
			{Name: "Harrison", Email: "harrison@lolwut.com"},
			{Name: "Oliver Allen", Email: "oliver@toyshop.com"},
		},
		Flags: []Flag{
			&StringFlag{Name: "name", Value: "bob", Usage: "a name to say"},
		},
		Commands: []*Command{
			{
				Name:        "describeit",
				Aliases:     []string{"d"},
				Usage:       "use it to see a description",
				Description: "This is how we describe describeit the function",
				Action: func(c *Context) error {
					fmt.Printf("i like to describe things")
					return nil
				},
			},
		},
	}
	_ = app.Run(os.Args)
	// Output:
	// NAME:
	//    greet - A new cli application
	//
	// USAGE:
	//    greet [global options] command [command options] [arguments...]
	//
	// VERSION:
	//    0.1.0
	//
	// DESCRIPTION:
	//    This is how we describe greet the app
	//
	// AUTHORS:
	//    Harrison <harrison@lolwut.com>
	//    Oliver Allen <oliver@toyshop.com>
	//
	// COMMANDS:
	//    describeit, d  use it to see a description
	//    help, h        Shows a list of commands or help for one command
	//
	// GLOBAL OPTIONS:
	//    --help, -h     show help (default: false)
	//    --name value   a name to say (default: "bob")
	//    --version, -v  print the version (default: false)
}

func ExampleApp_Run_commandHelp() {
	// set args for examples sake
	os.Args = []string{"greet", "h", "describeit"}

	app := &App{
		Name: "greet",
		Flags: []Flag{
			&StringFlag{Name: "name", Value: "bob", Usage: "a name to say"},
		},
		Commands: []*Command{
			{
				Name:        "describeit",
				Aliases:     []string{"d"},
				Usage:       "use it to see a description",
				Description: "This is how we describe describeit the function",
				Action: func(c *Context) error {
					fmt.Printf("i like to describe things")
					return nil
				},
			},
		},
	}
	_ = app.Run(os.Args)
	// Output:
	// NAME:
	//    greet describeit - use it to see a description
	//
	// USAGE:
	//    greet describeit [arguments...]
	//
	// DESCRIPTION:
	//    This is how we describe describeit the function
}

func ExampleApp_Run_noAction() {
	app := App{}
	app.Name = "greet"
	_ = app.Run([]string{"greet"})
	// Output:
	// NAME:
	//    greet - A new cli application
	//
	// USAGE:
	//    greet [global options] command [command options] [arguments...]
	//
	// COMMANDS:
	//    help, h  Shows a list of commands or help for one command
	//
	// GLOBAL OPTIONS:
	//    --help, -h  show help (default: false)
}

func ExampleApp_Run_subcommandNoAction() {
	app := &App{
		Name: "greet",
		Commands: []*Command{
			{
				Name:        "describeit",
				Aliases:     []string{"d"},
				Usage:       "use it to see a description",
				Description: "This is how we describe describeit the function",
			},
		},
	}
	_ = app.Run([]string{"greet", "describeit"})
	// Output:
	// NAME:
	//    greet describeit - use it to see a description
	//
	// USAGE:
	//    greet describeit [command options] [arguments...]
	//
	// DESCRIPTION:
	//    This is how we describe describeit the function
	//
	// OPTIONS:
	//    --help, -h  show help (default: false)

}

func ExampleApp_Run_bashComplete_withShortFlag() {
	os.Setenv("SHELL", "bash")
	os.Args = []string{"greet", "-", "--generate-bash-completion"}

	app := NewApp()
	app.Name = "greet"
	app.EnableBashCompletion = true
	app.Flags = []Flag{
		&IntFlag{
			Name:    "other",
			Aliases: []string{"o"},
		},
		&StringFlag{
			Name:    "xyz",
			Aliases: []string{"x"},
		},
	}

	_ = app.Run(os.Args)
	// Output:
	// --other
	// -o
	// --xyz
	// -x
	// --help
	// -h
}

func ExampleApp_Run_bashComplete_withLongFlag() {
	os.Setenv("SHELL", "bash")
	os.Args = []string{"greet", "--s", "--generate-bash-completion"}

	app := NewApp()
	app.Name = "greet"
	app.EnableBashCompletion = true
	app.Flags = []Flag{
		&IntFlag{
			Name:    "other",
			Aliases: []string{"o"},
		},
		&StringFlag{
			Name:    "xyz",
			Aliases: []string{"x"},
		},
		&StringFlag{
			Name: "some-flag,s",
		},
		&StringFlag{
			Name: "similar-flag",
		},
	}

	_ = app.Run(os.Args)
	// Output:
	// --some-flag
	// --similar-flag
}
func ExampleApp_Run_bashComplete_withMultipleLongFlag() {
	os.Setenv("SHELL", "bash")
	os.Args = []string{"greet", "--st", "--generate-bash-completion"}

	app := NewApp()
	app.Name = "greet"
	app.EnableBashCompletion = true
	app.Flags = []Flag{
		&IntFlag{
			Name:    "int-flag",
			Aliases: []string{"i"},
		},
		&StringFlag{
			Name:    "string",
			Aliases: []string{"s"},
		},
		&StringFlag{
			Name: "string-flag-2",
		},
		&StringFlag{
			Name: "similar-flag",
		},
		&StringFlag{
			Name: "some-flag",
		},
	}

	_ = app.Run(os.Args)
	// Output:
	// --string
	// --string-flag-2
}

func ExampleApp_Run_bashComplete() {
	os.Setenv("SHELL", "bash")
	os.Args = []string{"greet", "--generate-bash-completion"}

	app := &App{
		Name:                 "greet",
		EnableBashCompletion: true,
		Commands: []*Command{
			{
				Name:        "describeit",
				Aliases:     []string{"d"},
				Usage:       "use it to see a description",
				Description: "This is how we describe describeit the function",
				Action: func(c *Context) error {
					fmt.Printf("i like to describe things")
					return nil
				},
			}, {
				Name:        "next",
				Usage:       "next example",
				Description: "more stuff to see when generating shell completion",
				Action: func(c *Context) error {
					fmt.Printf("the next example")
					return nil
				},
			},
		},
	}

	_ = app.Run(os.Args)
	// Output:
	// describeit
	// d
	// next
	// help
	// h
}

func ExampleApp_Run_zshComplete() {
	// set args for examples sake
	os.Args = []string{"greet", "--generate-bash-completion"}
	_ = os.Setenv("SHELL", "/usr/bin/zsh")

	app := NewApp()
	app.Name = "greet"
	app.EnableBashCompletion = true
	app.Commands = []*Command{
		{
			Name:        "describeit",
			Aliases:     []string{"d"},
			Usage:       "use it to see a description",
			Description: "This is how we describe describeit the function",
			Action: func(c *Context) error {
				fmt.Printf("i like to describe things")
				return nil
			},
		}, {
			Name:        "next",
			Usage:       "next example",
			Description: "more stuff to see when generating bash completion",
			Action: func(c *Context) error {
				fmt.Printf("the next example")
				return nil
			},
		},
	}

	_ = app.Run(os.Args)
	// Output:
	// describeit:use it to see a description
	// d:use it to see a description
	// next:next example
	// help:Shows a list of commands or help for one command
	// h:Shows a list of commands or help for one command
}

func ExampleApp_Run_sliceValues() {
	// set args for examples sake
	os.Args = []string{"multi_values",
		"--stringSclice", "parsed1,parsed2", "--stringSclice", "parsed3,parsed4",
		"--float64Sclice", "13.3,14.4", "--float64Sclice", "15.5,16.6",
		"--int64Sclice", "13,14", "--int64Sclice", "15,16",
		"--intSclice", "13,14", "--intSclice", "15,16",
	}
	app := NewApp()
	app.Name = "multi_values"
	app.Flags = []Flag{
		&StringSliceFlag{Name: "stringSclice"},
		&Float64SliceFlag{Name: "float64Sclice"},
		&Int64SliceFlag{Name: "int64Sclice"},
		&IntSliceFlag{Name: "intSclice"},
	}
	app.Action = func(ctx *Context) error {
		for i, v := range ctx.FlagNames() {
			fmt.Printf("%d-%s %#v\n", i, v, ctx.Value(v))
		}
		err := ctx.Err()
		fmt.Println("error:", err)
		return err
	}

	_ = app.Run(os.Args)
	// Output:
	// 0-float64Sclice cli.Float64Slice{slice:[]float64{13.3, 14.4, 15.5, 16.6}, hasBeenSet:true}
	// 1-int64Sclice cli.Int64Slice{slice:[]int64{13, 14, 15, 16}, hasBeenSet:true}
	// 2-intSclice cli.IntSlice{slice:[]int{13, 14, 15, 16}, hasBeenSet:true}
	// 3-stringSclice cli.StringSlice{slice:[]string{"parsed1", "parsed2", "parsed3", "parsed4"}, hasBeenSet:true}
	// error: <nil>
}

func TestApp_Run(t *testing.T) {
	s := ""

	app := &App{
		Action: func(c *Context) error {
			s = s + c.Args().First()
			return nil
		},
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
	app := &App{
		Commands: []*Command{
			{Name: "foobar", Aliases: []string{"f"}},
			{Name: "batbaz", Aliases: []string{"b"}},
		},
	}

	for _, test := range commandAppTests {
		expect(t, app.Command(test.name) != nil, test.expected)
	}
}

var defaultCommandAppTests = []struct {
	cmdName    string
	defaultCmd string
	expected   bool
}{
	{"foobar", "foobar", true},
	{"batbaz", "foobar", true},
	{"b", "", true},
	{"f", "", true},
	{"", "foobar", true},
	{"", "", true},
	{" ", "", false},
	{"bat", "batbaz", false},
	{"nothing", "batbaz", false},
	{"nothing", "", false},
}

func TestApp_RunDefaultCommand(t *testing.T) {
	for _, test := range defaultCommandAppTests {
		testTitle := fmt.Sprintf("command=%[1]s-default=%[2]s", test.cmdName, test.defaultCmd)
		t.Run(testTitle, func(t *testing.T) {
			app := &App{
				DefaultCommand: test.defaultCmd,
				Commands: []*Command{
					{Name: "foobar", Aliases: []string{"f"}},
					{Name: "batbaz", Aliases: []string{"b"}},
				},
			}

			err := app.Run([]string{"c", test.cmdName})
			expect(t, err == nil, test.expected)
		})
	}
}

var defaultCommandSubCmdAppTests = []struct {
	cmdName    string
	subCmd     string
	defaultCmd string
	expected   bool
}{
	{"foobar", "", "foobar", true},
	{"foobar", "carly", "foobar", true},
	{"batbaz", "", "foobar", true},
	{"b", "", "", true},
	{"f", "", "", true},
	{"", "", "foobar", true},
	{"", "", "", true},
	{"", "jimbob", "foobar", true},
	{"", "j", "foobar", true},
	{"", "carly", "foobar", true},
	{"", "jimmers", "foobar", true},
	{"", "jimmers", "", true},
	{" ", "jimmers", "foobar", false},
	{"", "", "", true},
	{" ", "", "", false},
	{" ", "j", "", false},
	{"bat", "", "batbaz", false},
	{"nothing", "", "batbaz", false},
	{"nothing", "", "", false},
	{"nothing", "j", "batbaz", false},
	{"nothing", "carly", "", false},
}

func TestApp_RunDefaultCommandWithSubCommand(t *testing.T) {
	for _, test := range defaultCommandSubCmdAppTests {
		testTitle := fmt.Sprintf("command=%[1]s-subcmd=%[2]s-default=%[3]s", test.cmdName, test.subCmd, test.defaultCmd)
		t.Run(testTitle, func(t *testing.T) {
			app := &App{
				DefaultCommand: test.defaultCmd,
				Commands: []*Command{
					{
						Name:    "foobar",
						Aliases: []string{"f"},
						Subcommands: []*Command{
							{Name: "jimbob", Aliases: []string{"j"}},
							{Name: "carly"},
						},
					},
					{Name: "batbaz", Aliases: []string{"b"}},
				},
			}

			err := app.Run([]string{"c", test.cmdName, test.subCmd})
			expect(t, err == nil, test.expected)
		})
	}
}

var defaultCommandFlagAppTests = []struct {
	cmdName    string
	flag       string
	defaultCmd string
	expected   bool
}{
	{"foobar", "", "foobar", true},
	{"foobar", "-c derp", "foobar", true},
	{"batbaz", "", "foobar", true},
	{"b", "", "", true},
	{"f", "", "", true},
	{"", "", "foobar", true},
	{"", "", "", true},
	{"", "-j", "foobar", true},
	{"", "-j", "foobar", true},
	{"", "-c derp", "foobar", true},
	{"", "--carly=derp", "foobar", true},
	{"", "-j", "foobar", true},
	{"", "-j", "", true},
	{" ", "-j", "foobar", false},
	{"", "", "", true},
	{" ", "", "", false},
	{" ", "-j", "", false},
	{"bat", "", "batbaz", false},
	{"nothing", "", "batbaz", false},
	{"nothing", "", "", false},
	{"nothing", "--jimbob", "batbaz", false},
	{"nothing", "--carly", "", false},
}

func TestApp_RunDefaultCommandWithFlags(t *testing.T) {
	for _, test := range defaultCommandFlagAppTests {
		testTitle := fmt.Sprintf("command=%[1]s-flag=%[2]s-default=%[3]s", test.cmdName, test.flag, test.defaultCmd)
		t.Run(testTitle, func(t *testing.T) {
			app := &App{
				DefaultCommand: test.defaultCmd,
				Flags: []Flag{
					&StringFlag{
						Name:     "carly",
						Aliases:  []string{"c"},
						Required: false,
					},
					&BoolFlag{
						Name:     "jimbob",
						Aliases:  []string{"j"},
						Required: false,
						Value:    true,
					},
				},
				Commands: []*Command{
					{
						Name:    "foobar",
						Aliases: []string{"f"},
					},
					{Name: "batbaz", Aliases: []string{"b"}},
				},
			}

			appArgs := []string{"c"}

			if test.flag != "" {
				flags := strings.Split(test.flag, " ")
				if len(flags) > 1 {
					appArgs = append(appArgs, flags...)
				}

				flags = strings.Split(test.flag, "=")
				if len(flags) > 1 {
					appArgs = append(appArgs, flags...)
				}
			}

			appArgs = append(appArgs, test.cmdName)

			err := app.Run(appArgs)
			expect(t, err == nil, test.expected)
		})
	}
}

func TestApp_Setup_defaultsReader(t *testing.T) {
	app := &App{}
	app.Setup()
	expect(t, app.Reader, os.Stdin)
}

func TestApp_Setup_defaultsWriter(t *testing.T) {
	app := &App{}
	app.Setup()
	expect(t, app.Writer, os.Stdout)
}

func TestApp_RunAsSubcommandParseFlags(t *testing.T) {
	var cCtx *Context

	a := &App{
		Commands: []*Command{
			{
				Name: "foo",
				Action: func(c *Context) error {
					cCtx = c
					return nil
				},
				Flags: []Flag{
					&StringFlag{
						Name:  "lang",
						Value: "english",
						Usage: "language for the greeting",
					},
				},
				Before: func(_ *Context) error { return nil },
			},
		},
	}
	_ = a.Run([]string{"", "foo", "--lang", "spanish", "abcd"})

	expect(t, cCtx.Args().Get(0), "abcd")
	expect(t, cCtx.String("lang"), "spanish")
}

func TestApp_RunAsSubCommandIncorrectUsage(t *testing.T) {
	a := App{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{Name: "foo"},
		},
		Writer: bytes.NewBufferString(""),
	}

	set := flag.NewFlagSet("", flag.ContinueOnError)
	_ = set.Parse([]string{"", "-bar"})
	c := &Context{flagSet: set}

	err := a.RunAsSubcommand(c)

	expect(t, err.Error(), "flag provided but not defined: -bar")
}

func TestApp_CommandWithFlagBeforeTerminator(t *testing.T) {
	var parsedOption string
	var args Args

	app := &App{
		Commands: []*Command{
			{
				Name: "cmd",
				Flags: []Flag{
					&StringFlag{Name: "option", Value: "", Usage: "some option"},
				},
				Action: func(c *Context) error {
					parsedOption = c.String("option")
					args = c.Args()
					return nil
				},
			},
		},
	}

	_ = app.Run([]string{"", "cmd", "--option", "my-option", "my-arg", "--", "--notARealFlag"})

	expect(t, parsedOption, "my-option")
	expect(t, args.Get(0), "my-arg")
	expect(t, args.Get(1), "--")
	expect(t, args.Get(2), "--notARealFlag")
}

func TestApp_CommandWithDash(t *testing.T) {
	var args Args

	app := &App{
		Commands: []*Command{
			{
				Name: "cmd",
				Action: func(c *Context) error {
					args = c.Args()
					return nil
				},
			},
		},
	}

	_ = app.Run([]string{"", "cmd", "my-arg", "-"})

	expect(t, args.Get(0), "my-arg")
	expect(t, args.Get(1), "-")
}

func TestApp_CommandWithNoFlagBeforeTerminator(t *testing.T) {
	var args Args

	app := &App{
		Commands: []*Command{
			{
				Name: "cmd",
				Action: func(c *Context) error {
					args = c.Args()
					return nil
				},
			},
		},
	}

	_ = app.Run([]string{"", "cmd", "my-arg", "--", "notAFlagAtAll"})

	expect(t, args.Get(0), "my-arg")
	expect(t, args.Get(1), "--")
	expect(t, args.Get(2), "notAFlagAtAll")
}

func TestApp_VisibleCommands(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Name:     "frob",
				HelpName: "foo frob",
				Action:   func(_ *Context) error { return nil },
			},
			{
				Name:     "frib",
				HelpName: "foo frib",
				Hidden:   true,
				Action:   func(_ *Context) error { return nil },
			},
		},
	}

	app.Setup()
	expected := []*Command{
		app.Commands[0],
		app.Commands[2], // help
	}
	actual := app.VisibleCommands()
	expect(t, len(expected), len(actual))
	for i, actualCommand := range actual {
		expectedCommand := expected[i]

		if expectedCommand.Action != nil {
			// comparing func addresses is OK!
			expect(t, fmt.Sprintf("%p", expectedCommand.Action), fmt.Sprintf("%p", actualCommand.Action))
		}

		func() {
			// nil out funcs, as they cannot be compared
			// (https://github.com/golang/go/issues/8554)
			expectedAction := expectedCommand.Action
			actualAction := actualCommand.Action
			defer func() {
				expectedCommand.Action = expectedAction
				actualCommand.Action = actualAction
			}()
			expectedCommand.Action = nil
			actualCommand.Action = nil

			if !reflect.DeepEqual(expectedCommand, actualCommand) {
				t.Errorf("expected\n%#v\n!=\n%#v", expectedCommand, actualCommand)
			}
		}()
	}
}

func TestApp_UseShortOptionHandling(t *testing.T) {
	var one, two bool
	var name string
	expected := "expectedName"

	app := newTestApp()
	app.UseShortOptionHandling = true
	app.Flags = []Flag{
		&BoolFlag{Name: "one", Aliases: []string{"o"}},
		&BoolFlag{Name: "two", Aliases: []string{"t"}},
		&StringFlag{Name: "name", Aliases: []string{"n"}},
	}
	app.Action = func(c *Context) error {
		one = c.Bool("one")
		two = c.Bool("two")
		name = c.String("name")
		return nil
	}

	_ = app.Run([]string{"", "-on", expected})
	expect(t, one, true)
	expect(t, two, false)
	expect(t, name, expected)
}

func TestApp_UseShortOptionHandling_missing_value(t *testing.T) {
	app := newTestApp()
	app.UseShortOptionHandling = true
	app.Flags = []Flag{
		&StringFlag{Name: "name", Aliases: []string{"n"}},
	}

	err := app.Run([]string{"", "-n"})
	expect(t, err, errors.New("flag needs an argument: -n"))
}

func TestApp_UseShortOptionHandlingCommand(t *testing.T) {
	var one, two bool
	var name string
	expected := "expectedName"

	app := newTestApp()
	app.UseShortOptionHandling = true
	command := &Command{
		Name: "cmd",
		Flags: []Flag{
			&BoolFlag{Name: "one", Aliases: []string{"o"}},
			&BoolFlag{Name: "two", Aliases: []string{"t"}},
			&StringFlag{Name: "name", Aliases: []string{"n"}},
		},
		Action: func(c *Context) error {
			one = c.Bool("one")
			two = c.Bool("two")
			name = c.String("name")
			return nil
		},
	}
	app.Commands = []*Command{command}

	_ = app.Run([]string{"", "cmd", "-on", expected})
	expect(t, one, true)
	expect(t, two, false)
	expect(t, name, expected)
}

func TestApp_UseShortOptionHandlingCommand_missing_value(t *testing.T) {
	app := newTestApp()
	app.UseShortOptionHandling = true
	command := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{Name: "name", Aliases: []string{"n"}},
		},
	}
	app.Commands = []*Command{command}

	err := app.Run([]string{"", "cmd", "-n"})
	expect(t, err, errors.New("flag needs an argument: -n"))
}

func TestApp_UseShortOptionHandlingSubCommand(t *testing.T) {
	var one, two bool
	var name string
	expected := "expectedName"

	app := newTestApp()
	app.UseShortOptionHandling = true
	command := &Command{
		Name: "cmd",
	}
	subCommand := &Command{
		Name: "sub",
		Flags: []Flag{
			&BoolFlag{Name: "one", Aliases: []string{"o"}},
			&BoolFlag{Name: "two", Aliases: []string{"t"}},
			&StringFlag{Name: "name", Aliases: []string{"n"}},
		},
		Action: func(c *Context) error {
			one = c.Bool("one")
			two = c.Bool("two")
			name = c.String("name")
			return nil
		},
	}
	command.Subcommands = []*Command{subCommand}
	app.Commands = []*Command{command}

	err := app.Run([]string{"", "cmd", "sub", "-on", expected})
	expect(t, err, nil)
	expect(t, one, true)
	expect(t, two, false)
	expect(t, name, expected)
}

func TestApp_UseShortOptionHandlingSubCommand_missing_value(t *testing.T) {
	app := newTestApp()
	app.UseShortOptionHandling = true
	command := &Command{
		Name: "cmd",
	}
	subCommand := &Command{
		Name: "sub",
		Flags: []Flag{
			&StringFlag{Name: "name", Aliases: []string{"n"}},
		},
	}
	command.Subcommands = []*Command{subCommand}
	app.Commands = []*Command{command}

	err := app.Run([]string{"", "cmd", "sub", "-n"})
	expect(t, err, errors.New("flag needs an argument: -n"))
}

func TestApp_Float64Flag(t *testing.T) {
	var meters float64

	app := &App{
		Flags: []Flag{
			&Float64Flag{Name: "height", Value: 1.5, Usage: "Set the height, in meters"},
		},
		Action: func(c *Context) error {
			meters = c.Float64("height")
			return nil
		},
	}

	_ = app.Run([]string{"", "--height", "1.93"})
	expect(t, meters, 1.93)
}

func TestApp_ParseSliceFlags(t *testing.T) {
	var parsedIntSlice []int
	var parsedStringSlice []string

	app := &App{
		Commands: []*Command{
			{
				Name: "cmd",
				Flags: []Flag{
					&IntSliceFlag{Name: "p", Value: NewIntSlice(), Usage: "set one or more ip addr"},
					&StringSliceFlag{Name: "ip", Value: NewStringSlice(), Usage: "set one or more ports to open"},
				},
				Action: func(c *Context) error {
					parsedIntSlice = c.IntSlice("p")
					parsedStringSlice = c.StringSlice("ip")
					return nil
				},
			},
		},
	}

	_ = app.Run([]string{"", "cmd", "-p", "22", "-p", "80", "-ip", "8.8.8.8", "-ip", "8.8.4.4"})

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
		t.Errorf("%v does not match %v", parsedIntSlice, expectedIntSlice)
	}

	if !StrsEquals(parsedStringSlice, expectedStringSlice) {
		t.Errorf("%v does not match %v", parsedStringSlice, expectedStringSlice)
	}
}

func TestApp_ParseSliceFlagsWithMissingValue(t *testing.T) {
	var parsedIntSlice []int
	var parsedStringSlice []string

	app := &App{
		Commands: []*Command{
			{
				Name: "cmd",
				Flags: []Flag{
					&IntSliceFlag{Name: "a", Usage: "set numbers"},
					&StringSliceFlag{Name: "str", Usage: "set strings"},
				},
				Action: func(c *Context) error {
					parsedIntSlice = c.IntSlice("a")
					parsedStringSlice = c.StringSlice("str")
					return nil
				},
			},
		},
	}

	_ = app.Run([]string{"", "cmd", "-a", "2", "-str", "A"})

	var expectedIntSlice = []int{2}
	var expectedStringSlice = []string{"A"}

	if parsedIntSlice[0] != expectedIntSlice[0] {
		t.Errorf("%v does not match %v", parsedIntSlice[0], expectedIntSlice[0])
	}

	if parsedStringSlice[0] != expectedStringSlice[0] {
		t.Errorf("%v does not match %v", parsedIntSlice[0], expectedIntSlice[0])
	}
}

func TestApp_DefaultStdin(t *testing.T) {
	app := &App{}
	app.Setup()

	if app.Reader != os.Stdin {
		t.Error("Default input reader not set.")
	}
}

func TestApp_DefaultStdout(t *testing.T) {
	app := &App{}
	app.Setup()

	if app.Writer != os.Stdout {
		t.Error("Default output writer not set.")
	}
}

func TestApp_SetStdin(t *testing.T) {
	buf := make([]byte, 12)

	app := &App{
		Name:   "test",
		Reader: strings.NewReader("Hello World!"),
		Action: func(c *Context) error {
			_, err := c.App.Reader.Read(buf)
			return err
		},
	}

	err := app.Run([]string{"help"})

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if string(buf) != "Hello World!" {
		t.Error("App did not read input from desired reader.")
	}
}

func TestApp_SetStdin_Subcommand(t *testing.T) {
	buf := make([]byte, 12)

	app := &App{
		Name:   "test",
		Reader: strings.NewReader("Hello World!"),
		Commands: []*Command{
			{
				Name: "command",
				Subcommands: []*Command{
					{
						Name: "subcommand",
						Action: func(c *Context) error {
							_, err := c.App.Reader.Read(buf)
							return err
						},
					},
				},
			},
		},
	}

	err := app.Run([]string{"test", "command", "subcommand"})

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if string(buf) != "Hello World!" {
		t.Error("App did not read input from desired reader.")
	}
}

func TestApp_SetStdout(t *testing.T) {
	var w bytes.Buffer

	app := &App{
		Name:   "test",
		Writer: &w,
	}

	err := app.Run([]string{"help"})

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if w.Len() == 0 {
		t.Error("App did not write output to desired writer.")
	}
}

func TestApp_BeforeFunc(t *testing.T) {
	counts := &opCounts{}
	beforeError := fmt.Errorf("fail")
	var err error

	app := &App{
		Before: func(c *Context) error {
			counts.Total++
			counts.Before = counts.Total
			s := c.String("opt")
			if s == "fail" {
				return beforeError
			}

			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(c *Context) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		},
		Flags: []Flag{
			&StringFlag{Name: "opt"},
		},
		Writer: ioutil.Discard,
	}

	// run with the Before() func succeeding
	err = app.Run([]string{"command", "--opt", "succeed", "sub"})

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if counts.Before != 1 {
		t.Errorf("Before() not executed when expected")
	}

	if counts.SubCommand != 2 {
		t.Errorf("Subcommand not executed when expected")
	}

	// reset
	counts = &opCounts{}

	// run with the Before() func failing
	err = app.Run([]string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	if err != beforeError {
		t.Errorf("Run error expected, but not received")
	}

	if counts.Before != 1 {
		t.Errorf("Before() not executed when expected")
	}

	if counts.SubCommand != 0 {
		t.Errorf("Subcommand executed when NOT expected")
	}

	// reset
	counts = &opCounts{}

	afterError := errors.New("fail again")
	app.After = func(_ *Context) error {
		return afterError
	}

	// run with the Before() func failing, wrapped by After()
	err = app.Run([]string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	if _, ok := err.(MultiError); !ok {
		t.Errorf("MultiError expected, but not received")
	}

	if counts.Before != 1 {
		t.Errorf("Before() not executed when expected")
	}

	if counts.SubCommand != 0 {
		t.Errorf("Subcommand executed when NOT expected")
	}
}

func TestApp_BeforeAfterFuncShellCompletion(t *testing.T) {
	counts := &opCounts{}
	var err error

	app := &App{
		EnableBashCompletion: true,
		Before: func(c *Context) error {
			counts.Total++
			counts.Before = counts.Total
			return nil
		},
		After: func(c *Context) error {
			counts.Total++
			counts.After = counts.Total
			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(c *Context) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		},
		Flags: []Flag{
			&StringFlag{Name: "opt"},
		},
		Writer: ioutil.Discard,
	}

	// run with the Before() func succeeding
	err = app.Run([]string{"command", "--opt", "succeed", "sub", "--generate-bash-completion"})

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if counts.Before != 0 {
		t.Errorf("Before() executed when not expected")
	}

	if counts.After != 0 {
		t.Errorf("After() executed when not expected")
	}

	if counts.SubCommand != 0 {
		t.Errorf("Subcommand executed more than expected")
	}
}

func TestApp_AfterFunc(t *testing.T) {
	counts := &opCounts{}
	afterError := fmt.Errorf("fail")
	var err error

	app := &App{
		After: func(c *Context) error {
			counts.Total++
			counts.After = counts.Total
			s := c.String("opt")
			if s == "fail" {
				return afterError
			}

			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(c *Context) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		},
		Flags: []Flag{
			&StringFlag{Name: "opt"},
		},
	}

	// run with the After() func succeeding
	err = app.Run([]string{"command", "--opt", "succeed", "sub"})

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if counts.After != 2 {
		t.Errorf("After() not executed when expected")
	}

	if counts.SubCommand != 1 {
		t.Errorf("Subcommand not executed when expected")
	}

	// reset
	counts = &opCounts{}

	// run with the Before() func failing
	err = app.Run([]string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	if err != afterError {
		t.Errorf("Run error expected, but not received")
	}

	if counts.After != 2 {
		t.Errorf("After() not executed when expected")
	}

	if counts.SubCommand != 1 {
		t.Errorf("Subcommand not executed when expected")
	}

	/*
		reset
	*/
	counts = &opCounts{}

	// run with none args
	err = app.Run([]string{"command"})

	// should be the same error produced by the Before func
	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if counts.After != 1 {
		t.Errorf("After() not executed when expected")
	}

	if counts.SubCommand != 0 {
		t.Errorf("Subcommand not executed when expected")
	}
}

func TestAppNoHelpFlag(t *testing.T) {
	oldFlag := HelpFlag
	defer func() {
		HelpFlag = oldFlag
	}()

	HelpFlag = nil

	app := &App{Writer: ioutil.Discard}
	err := app.Run([]string{"test", "-h"})

	if err != flag.ErrHelp {
		t.Errorf("expected error about missing help flag, but got: %s (%T)", err, err)
	}
}

func TestRequiredFlagAppRunBehavior(t *testing.T) {
	tdata := []struct {
		testCase        string
		appFlags        []Flag
		appRunInput     []string
		appCommands     []*Command
		expectedAnError bool
	}{
		// assertion: empty input, when a required flag is present, errors
		{
			testCase:        "error_case_empty_input_with_required_flag_on_app",
			appRunInput:     []string{"myCLI"},
			appFlags:        []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
			expectedAnError: true,
		},
		{
			testCase:    "error_case_empty_input_with_required_flag_on_command",
			appRunInput: []string{"myCLI", "myCommand"},
			appCommands: []*Command{{
				Name:  "myCommand",
				Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
			}},
			expectedAnError: true,
		},
		{
			testCase:    "error_case_empty_input_with_required_flag_on_subcommand",
			appRunInput: []string{"myCLI", "myCommand", "mySubCommand"},
			appCommands: []*Command{{
				Name: "myCommand",
				Subcommands: []*Command{{
					Name:  "mySubCommand",
					Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
				}},
			}},
			expectedAnError: true,
		},
		// assertion: inputing --help, when a required flag is present, does not error
		{
			testCase:    "valid_case_help_input_with_required_flag_on_app",
			appRunInput: []string{"myCLI", "--help"},
			appFlags:    []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
		},
		{
			testCase:    "valid_case_help_input_with_required_flag_on_command",
			appRunInput: []string{"myCLI", "myCommand", "--help"},
			appCommands: []*Command{{
				Name:  "myCommand",
				Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
			}},
		},
		{
			testCase:    "valid_case_help_input_with_required_flag_on_subcommand",
			appRunInput: []string{"myCLI", "myCommand", "mySubCommand", "--help"},
			appCommands: []*Command{{
				Name: "myCommand",
				Subcommands: []*Command{{
					Name:  "mySubCommand",
					Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
				}},
			}},
		},
		// assertion: giving optional input, when a required flag is present, errors
		{
			testCase:        "error_case_optional_input_with_required_flag_on_app",
			appRunInput:     []string{"myCLI", "--optional", "cats"},
			appFlags:        []Flag{&StringFlag{Name: "requiredFlag", Required: true}, &StringFlag{Name: "optional"}},
			expectedAnError: true,
		},
		{
			testCase:    "error_case_optional_input_with_required_flag_on_command",
			appRunInput: []string{"myCLI", "myCommand", "--optional", "cats"},
			appCommands: []*Command{{
				Name:  "myCommand",
				Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}, &StringFlag{Name: "optional"}},
			}},
			expectedAnError: true,
		},
		{
			testCase:    "error_case_optional_input_with_required_flag_on_subcommand",
			appRunInput: []string{"myCLI", "myCommand", "mySubCommand", "--optional", "cats"},
			appCommands: []*Command{{
				Name: "myCommand",
				Subcommands: []*Command{{
					Name:  "mySubCommand",
					Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}, &StringFlag{Name: "optional"}},
				}},
			}},
			expectedAnError: true,
		},
		// assertion: when a required flag is present, inputting that required flag does not error
		{
			testCase:    "valid_case_required_flag_input_on_app",
			appRunInput: []string{"myCLI", "--requiredFlag", "cats"},
			appFlags:    []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
		},
		{
			testCase:    "valid_case_required_flag_input_on_command",
			appRunInput: []string{"myCLI", "myCommand", "--requiredFlag", "cats"},
			appCommands: []*Command{{
				Name:  "myCommand",
				Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
			}},
		},
		{
			testCase:    "valid_case_required_flag_input_on_subcommand",
			appRunInput: []string{"myCLI", "myCommand", "mySubCommand", "--requiredFlag", "cats"},
			appCommands: []*Command{{
				Name: "myCommand",
				Subcommands: []*Command{{
					Name:  "mySubCommand",
					Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
				}},
			}},
		},
	}
	for _, test := range tdata {
		t.Run(test.testCase, func(t *testing.T) {
			// setup
			app := newTestApp()
			app.Flags = test.appFlags
			app.Commands = test.appCommands

			// logic under test
			err := app.Run(test.appRunInput)

			// assertions
			if test.expectedAnError && err == nil {
				t.Errorf("expected an error, but there was none")
			}
			if _, ok := err.(requiredFlagsErr); test.expectedAnError && !ok {
				t.Errorf("expected a requiredFlagsErr, but got: %s", err)
			}
			if !test.expectedAnError && err != nil {
				t.Errorf("did not expected an error, but there was one: %s", err)
			}
		})
	}
}

func TestAppHelpPrinter(t *testing.T) {
	oldPrinter := HelpPrinter
	defer func() {
		HelpPrinter = oldPrinter
	}()

	var wasCalled = false
	HelpPrinter = func(w io.Writer, template string, data interface{}) {
		wasCalled = true
	}

	app := &App{}
	_ = app.Run([]string{"-h"})

	if wasCalled == false {
		t.Errorf("Help printer expected to be called, but was not")
	}
}

func TestApp_VersionPrinter(t *testing.T) {
	oldPrinter := VersionPrinter
	defer func() {
		VersionPrinter = oldPrinter
	}()

	var wasCalled = false
	VersionPrinter = func(c *Context) {
		wasCalled = true
	}

	app := &App{}
	ctx := NewContext(app, nil, nil)
	ShowVersion(ctx)

	if wasCalled == false {
		t.Errorf("Version printer expected to be called, but was not")
	}
}

func TestApp_CommandNotFound(t *testing.T) {
	counts := &opCounts{}
	app := &App{
		CommandNotFound: func(c *Context, command string) {
			counts.Total++
			counts.CommandNotFound = counts.Total
		},
		Commands: []*Command{
			{
				Name: "bar",
				Action: func(c *Context) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		},
	}

	_ = app.Run([]string{"command", "foo"})

	expect(t, counts.CommandNotFound, 1)
	expect(t, counts.SubCommand, 0)
	expect(t, counts.Total, 1)
}

func TestApp_OrderOfOperations(t *testing.T) {
	counts := &opCounts{}

	resetCounts := func() { counts = &opCounts{} }

	app := &App{
		EnableBashCompletion: true,
		BashComplete: func(c *Context) {
			counts.Total++
			counts.ShellComplete = counts.Total
		},
		OnUsageError: func(c *Context, err error, isSubcommand bool) error {
			counts.Total++
			counts.OnUsageError = counts.Total
			return errors.New("hay OnUsageError")
		},
		Writer: ioutil.Discard,
	}

	beforeNoError := func(c *Context) error {
		counts.Total++
		counts.Before = counts.Total
		return nil
	}

	beforeError := func(c *Context) error {
		counts.Total++
		counts.Before = counts.Total
		return errors.New("hay Before")
	}

	app.Before = beforeNoError
	app.CommandNotFound = func(c *Context, command string) {
		counts.Total++
		counts.CommandNotFound = counts.Total
	}

	afterNoError := func(c *Context) error {
		counts.Total++
		counts.After = counts.Total
		return nil
	}

	afterError := func(c *Context) error {
		counts.Total++
		counts.After = counts.Total
		return errors.New("hay After")
	}

	app.After = afterNoError
	app.Commands = []*Command{
		{
			Name: "bar",
			Action: func(c *Context) error {
				counts.Total++
				counts.SubCommand = counts.Total
				return nil
			},
		},
	}

	app.Action = func(c *Context) error {
		counts.Total++
		counts.Action = counts.Total
		return nil
	}

	_ = app.Run([]string{"command", "--nope"})
	expect(t, counts.OnUsageError, 1)
	expect(t, counts.Total, 1)

	resetCounts()

	_ = app.Run([]string{"command", fmt.Sprintf("--%s", "generate-bash-completion")})
	expect(t, counts.ShellComplete, 1)
	expect(t, counts.Total, 1)

	resetCounts()

	oldOnUsageError := app.OnUsageError
	app.OnUsageError = nil
	_ = app.Run([]string{"command", "--nope"})
	expect(t, counts.Total, 0)
	app.OnUsageError = oldOnUsageError

	resetCounts()

	_ = app.Run([]string{"command", "foo"})
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.CommandNotFound, 0)
	expect(t, counts.Action, 2)
	expect(t, counts.After, 3)
	expect(t, counts.Total, 3)

	resetCounts()

	app.Before = beforeError
	_ = app.Run([]string{"command", "bar"})
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.After, 2)
	expect(t, counts.Total, 2)
	app.Before = beforeNoError

	resetCounts()

	app.After = nil
	_ = app.Run([]string{"command", "bar"})
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.SubCommand, 2)
	expect(t, counts.Total, 2)
	app.After = afterNoError

	resetCounts()

	app.After = afterError
	err := app.Run([]string{"command", "bar"})
	if err == nil {
		t.Fatalf("expected a non-nil error")
	}
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.SubCommand, 2)
	expect(t, counts.After, 3)
	expect(t, counts.Total, 3)
	app.After = afterNoError

	resetCounts()

	oldCommands := app.Commands
	app.Commands = nil
	_ = app.Run([]string{"command"})
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.Action, 2)
	expect(t, counts.After, 3)
	expect(t, counts.Total, 3)
	app.Commands = oldCommands
}

func TestApp_Run_CommandWithSubcommandHasHelpTopic(t *testing.T) {
	var subcommandHelpTopics = [][]string{
		{"command", "foo", "--help"},
		{"command", "foo", "-h"},
		{"command", "foo", "help"},
	}

	for _, flagSet := range subcommandHelpTopics {
		t.Run(fmt.Sprintf("checking with flags %v", flagSet), func(t *testing.T) {

			app := &App{}
			buf := new(bytes.Buffer)
			app.Writer = buf

			subCmdBar := &Command{
				Name:  "bar",
				Usage: "does bar things",
			}
			subCmdBaz := &Command{
				Name:  "baz",
				Usage: "does baz things",
			}
			cmd := &Command{
				Name:        "foo",
				Description: "descriptive wall of text about how it does foo things",
				Subcommands: []*Command{subCmdBar, subCmdBaz},
				Action:      func(c *Context) error { return nil },
			}

			app.Commands = []*Command{cmd}
			err := app.Run(flagSet)

			if err != nil {
				t.Error(err)
			}

			output := buf.String()

			if strings.Contains(output, "No help topic for") {
				t.Errorf("expect a help topic, got none: \n%q", output)
			}

			for _, shouldContain := range []string{
				cmd.Name, cmd.Description,
				subCmdBar.Name, subCmdBar.Usage,
				subCmdBaz.Name, subCmdBaz.Usage,
			} {
				if !strings.Contains(output, shouldContain) {
					t.Errorf("want help to contain %q, did not: \n%q", shouldContain, output)
				}
			}
		})
	}
}

func TestApp_Run_SubcommandFullPath(t *testing.T) {
	app := &App{}
	buf := new(bytes.Buffer)
	app.Writer = buf
	app.Name = "command"
	subCmd := &Command{
		Name:  "bar",
		Usage: "does bar things",
	}
	cmd := &Command{
		Name:        "foo",
		Description: "foo commands",
		Subcommands: []*Command{subCmd},
	}
	app.Commands = []*Command{cmd}

	err := app.Run([]string{"command", "foo", "bar", "--help"})
	if err != nil {
		t.Error(err)
	}

	output := buf.String()
	expected := "command foo bar - does bar things"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %s", expected, output)
	}

	expected = "command foo bar [command options] [arguments...]"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %s", expected, output)
	}
}

func TestApp_Run_SubcommandHelpName(t *testing.T) {
	app := &App{}
	buf := new(bytes.Buffer)
	app.Writer = buf
	app.Name = "command"
	subCmd := &Command{
		Name:     "bar",
		HelpName: "custom",
		Usage:    "does bar things",
	}
	cmd := &Command{
		Name:        "foo",
		Description: "foo commands",
		Subcommands: []*Command{subCmd},
	}
	app.Commands = []*Command{cmd}

	err := app.Run([]string{"command", "foo", "bar", "--help"})
	if err != nil {
		t.Error(err)
	}

	output := buf.String()

	expected := "custom - does bar things"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %s", expected, output)
	}

	expected = "custom [command options] [arguments...]"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %s", expected, output)
	}
}

func TestApp_Run_CommandHelpName(t *testing.T) {
	app := &App{}
	buf := new(bytes.Buffer)
	app.Writer = buf
	app.Name = "command"
	subCmd := &Command{
		Name:  "bar",
		Usage: "does bar things",
	}
	cmd := &Command{
		Name:        "foo",
		HelpName:    "custom",
		Description: "foo commands",
		Subcommands: []*Command{subCmd},
	}
	app.Commands = []*Command{cmd}

	err := app.Run([]string{"command", "foo", "bar", "--help"})
	if err != nil {
		t.Error(err)
	}

	output := buf.String()

	expected := "command foo bar - does bar things"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %s", expected, output)
	}

	expected = "command foo bar [command options] [arguments...]"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %s", expected, output)
	}
}

func TestApp_Run_CommandSubcommandHelpName(t *testing.T) {
	app := &App{}
	buf := new(bytes.Buffer)
	app.Writer = buf
	app.Name = "base"
	subCmd := &Command{
		Name:     "bar",
		HelpName: "custom",
		Usage:    "does bar things",
	}
	cmd := &Command{
		Name:        "foo",
		Usage:       "foo commands",
		Description: "This is a description",
		Subcommands: []*Command{subCmd},
	}
	app.Commands = []*Command{cmd}

	err := app.Run([]string{"command", "foo", "--help"})
	if err != nil {
		t.Error(err)
	}

	output := buf.String()

	expected := "base foo - foo commands"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %q", expected, output)
	}

	expected = "DESCRIPTION:\n   This is a description\n"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %q", expected, output)
	}

	expected = "base foo command [command options] [arguments...]"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %q", expected, output)
	}
}

func TestApp_Run_Help(t *testing.T) {
	var tests = []struct {
		helpArguments []string
		hideHelp      bool
		wantContains  string
		wantErr       error
	}{
		{
			helpArguments: []string{"boom", "--help"},
			hideHelp:      false,
			wantContains:  "boom - make an explosive entrance",
		},
		{
			helpArguments: []string{"boom", "-h"},
			hideHelp:      false,
			wantContains:  "boom - make an explosive entrance",
		},
		{
			helpArguments: []string{"boom", "help"},
			hideHelp:      false,
			wantContains:  "boom - make an explosive entrance",
		},
		{
			helpArguments: []string{"boom", "--help"},
			hideHelp:      true,
			wantErr:       fmt.Errorf("flag: help requested"),
		},
		{
			helpArguments: []string{"boom", "-h"},
			hideHelp:      true,
			wantErr:       fmt.Errorf("flag: help requested"),
		},
		{
			helpArguments: []string{"boom", "help"},
			hideHelp:      true,
			wantContains:  "boom I say!",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("checking with arguments %v", tt.helpArguments), func(t *testing.T) {
			buf := new(bytes.Buffer)

			app := &App{
				Name:     "boom",
				Usage:    "make an explosive entrance",
				Writer:   buf,
				HideHelp: tt.hideHelp,
				Action: func(c *Context) error {
					buf.WriteString("boom I say!")
					return nil
				},
			}

			err := app.Run(tt.helpArguments)
			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("want err: %s, did note %s\n", tt.wantErr, err)
			}

			output := buf.String()

			if !strings.Contains(output, tt.wantContains) {
				t.Errorf("want help to contain %q, did not: \n%q", "boom - make an explosive entrance", output)
			}
		})
	}
}

func TestApp_Run_Version(t *testing.T) {
	var versionArguments = [][]string{{"boom", "--version"}, {"boom", "-v"}}

	for _, args := range versionArguments {
		t.Run(fmt.Sprintf("checking with arguments %v", args), func(t *testing.T) {

			buf := new(bytes.Buffer)

			app := &App{
				Name:    "boom",
				Usage:   "make an explosive entrance",
				Version: "0.1.0",
				Writer:  buf,
				Action: func(c *Context) error {
					buf.WriteString("boom I say!")
					return nil
				},
			}

			err := app.Run(args)
			if err != nil {
				t.Error(err)
			}

			output := buf.String()

			if !strings.Contains(output, "0.1.0") {
				t.Errorf("want version to contain %q, did not: \n%q", "0.1.0", output)
			}
		})
	}
}

func TestApp_Run_Categories(t *testing.T) {
	buf := new(bytes.Buffer)

	app := &App{
		Name:     "categories",
		HideHelp: true,
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
			},
			{
				Name:     "command2",
				Category: "1",
			},
			{
				Name:     "command3",
				Category: "2",
			},
		},
		Writer: buf,
	}

	_ = app.Run([]string{"categories"})

	expect := commandCategories([]*commandCategory{
		{
			name: "1",
			commands: []*Command{
				app.Commands[0],
				app.Commands[1],
			},
		},
		{
			name: "2",
			commands: []*Command{
				app.Commands[2],
			},
		},
	})

	if !reflect.DeepEqual(app.categories, &expect) {
		t.Fatalf("expected categories %#v, to equal %#v", app.categories, &expect)
	}

	output := buf.String()

	if !strings.Contains(output, "1:\n     command1") {
		t.Errorf("want buffer to include category %q, did not: \n%q", "1:\n     command1", output)
	}
}

func TestApp_VisibleCategories(t *testing.T) {
	app := &App{
		Name:     "visible-categories",
		HideHelp: true,
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
				HelpName: "foo command1",
				Hidden:   true,
			},
			{
				Name:     "command2",
				Category: "2",
				HelpName: "foo command2",
			},
			{
				Name:     "command3",
				Category: "3",
				HelpName: "foo command3",
			},
		},
	}

	expected := []CommandCategory{
		&commandCategory{
			name: "2",
			commands: []*Command{
				app.Commands[1],
			},
		},
		&commandCategory{
			name: "3",
			commands: []*Command{
				app.Commands[2],
			},
		},
	}

	app.Setup()
	expect(t, expected, app.VisibleCategories())

	app = &App{
		Name:     "visible-categories",
		HideHelp: true,
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
				HelpName: "foo command1",
				Hidden:   true,
			},
			{
				Name:     "command2",
				Category: "2",
				HelpName: "foo command2",
				Hidden:   true,
			},
			{
				Name:     "command3",
				Category: "3",
				HelpName: "foo command3",
			},
		},
	}

	expected = []CommandCategory{
		&commandCategory{
			name: "3",
			commands: []*Command{
				app.Commands[2],
			},
		},
	}

	app.Setup()
	expect(t, expected, app.VisibleCategories())

	app = &App{
		Name:     "visible-categories",
		HideHelp: true,
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
				HelpName: "foo command1",
				Hidden:   true,
			},
			{
				Name:     "command2",
				Category: "2",
				HelpName: "foo command2",
				Hidden:   true,
			},
			{
				Name:     "command3",
				Category: "3",
				HelpName: "foo command3",
				Hidden:   true,
			},
		},
	}

	app.Setup()
	expect(t, []CommandCategory{}, app.VisibleCategories())
}

func TestApp_VisibleFlagCategories(t *testing.T) {
	app := &App{}
	vfc := app.VisibleFlagCategories()
	if len(vfc) != 0 {
		t.Errorf("unexpected visible flag categories %+v", vfc)
	}
}

func TestApp_Run_DoesNotOverwriteErrorFromBefore(t *testing.T) {
	app := &App{
		Action: func(c *Context) error { return nil },
		Before: func(c *Context) error { return fmt.Errorf("before error") },
		After:  func(c *Context) error { return fmt.Errorf("after error") },
		Writer: ioutil.Discard,
	}

	err := app.Run([]string{"foo"})
	if err == nil {
		t.Fatalf("expected to receive error from Run, got none")
	}

	if !strings.Contains(err.Error(), "before error") {
		t.Errorf("expected text of error from Before method, but got none in \"%v\"", err)
	}
	if !strings.Contains(err.Error(), "after error") {
		t.Errorf("expected text of error from After method, but got none in \"%v\"", err)
	}
}

func TestApp_Run_SubcommandDoesNotOverwriteErrorFromBefore(t *testing.T) {
	app := &App{
		Commands: []*Command{
			{
				Subcommands: []*Command{
					{
						Name: "sub",
					},
				},
				Name:   "bar",
				Before: func(c *Context) error { return fmt.Errorf("before error") },
				After:  func(c *Context) error { return fmt.Errorf("after error") },
			},
		},
	}

	err := app.Run([]string{"foo", "bar"})
	if err == nil {
		t.Fatalf("expected to receive error from Run, got none")
	}

	if !strings.Contains(err.Error(), "before error") {
		t.Errorf("expected text of error from Before method, but got none in \"%v\"", err)
	}
	if !strings.Contains(err.Error(), "after error") {
		t.Errorf("expected text of error from After method, but got none in \"%v\"", err)
	}
}

func TestApp_OnUsageError_WithWrongFlagValue(t *testing.T) {
	app := &App{
		Flags: []Flag{
			&IntFlag{Name: "flag"},
		},
		OnUsageError: func(c *Context, err error, isSubcommand bool) error {
			if isSubcommand {
				t.Errorf("Expect no subcommand")
			}
			if !strings.HasPrefix(err.Error(), "invalid value \"wrong\"") {
				t.Errorf("Expect an invalid value error, but got \"%v\"", err)
			}
			return errors.New("intercepted: " + err.Error())
		},
		Commands: []*Command{
			{
				Name: "bar",
			},
		},
	}

	err := app.Run([]string{"foo", "--flag=wrong"})
	if err == nil {
		t.Fatalf("expected to receive error from Run, got none")
	}

	if !strings.HasPrefix(err.Error(), "intercepted: invalid value") {
		t.Errorf("Expect an intercepted error, but got \"%v\"", err)
	}
}

func TestApp_OnUsageError_WithWrongFlagValue_ForSubcommand(t *testing.T) {
	app := &App{
		Flags: []Flag{
			&IntFlag{Name: "flag"},
		},
		OnUsageError: func(c *Context, err error, isSubcommand bool) error {
			if isSubcommand {
				t.Errorf("Expect subcommand")
			}
			if !strings.HasPrefix(err.Error(), "invalid value \"wrong\"") {
				t.Errorf("Expect an invalid value error, but got \"%v\"", err)
			}
			return errors.New("intercepted: " + err.Error())
		},
		Commands: []*Command{
			{
				Name: "bar",
			},
		},
	}

	err := app.Run([]string{"foo", "--flag=wrong", "bar"})
	if err == nil {
		t.Fatalf("expected to receive error from Run, got none")
	}

	if !strings.HasPrefix(err.Error(), "intercepted: invalid value") {
		t.Errorf("Expect an intercepted error, but got \"%v\"", err)
	}
}

// A custom flag that conforms to the relevant interfaces, but has none of the
// fields that the other flag types do.
type customBoolFlag struct {
	Nombre string
}

// Don't use the normal FlagStringer
func (c *customBoolFlag) String() string {
	return "***" + c.Nombre + "***"
}

func (c *customBoolFlag) Names() []string {
	return []string{c.Nombre}
}

func (c *customBoolFlag) TakesValue() bool {
	return false
}

func (c *customBoolFlag) GetValue() string {
	return "value"
}

func (c *customBoolFlag) GetUsage() string {
	return "usage"
}

func (c *customBoolFlag) Apply(set *flag.FlagSet) error {
	set.String(c.Nombre, c.Nombre, "")
	return nil
}

func (c *customBoolFlag) IsSet() bool {
	return false
}

func TestCustomFlagsUnused(t *testing.T) {
	app := &App{
		Flags:  []Flag{&customBoolFlag{"custom"}},
		Writer: ioutil.Discard,
	}

	err := app.Run([]string{"foo"})
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestCustomFlagsUsed(t *testing.T) {
	app := &App{
		Flags:  []Flag{&customBoolFlag{"custom"}},
		Writer: ioutil.Discard,
	}

	err := app.Run([]string{"foo", "--custom=bar"})
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestCustomHelpVersionFlags(t *testing.T) {
	app := &App{
		Writer: ioutil.Discard,
	}

	// Be sure to reset the global flags
	defer func(helpFlag Flag, versionFlag Flag) {
		HelpFlag = helpFlag.(*BoolFlag)
		VersionFlag = versionFlag.(*BoolFlag)
	}(HelpFlag, VersionFlag)

	HelpFlag = &customBoolFlag{"help-custom"}
	VersionFlag = &customBoolFlag{"version-custom"}

	err := app.Run([]string{"foo", "--help-custom=bar"})
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestHandleExitCoder_Default(t *testing.T) {
	app := newTestApp()
	fs, err := flagSet(app.Name, app.Flags)
	if err != nil {
		t.Errorf("error creating FlagSet: %s", err)
	}

	ctx := NewContext(app, fs, nil)
	app.handleExitCoder(ctx, NewExitError("Default Behavior Error", 42))

	output := fakeErrWriter.String()
	if !strings.Contains(output, "Default") {
		t.Fatalf("Expected Default Behavior from Error Handler but got: %s", output)
	}
}

func TestHandleExitCoder_Custom(t *testing.T) {
	app := newTestApp()
	fs, err := flagSet(app.Name, app.Flags)
	if err != nil {
		t.Errorf("error creating FlagSet: %s", err)
	}

	app.ExitErrHandler = func(_ *Context, _ error) {
		_, _ = fmt.Fprintln(ErrWriter, "I'm a Custom error handler, I print what I want!")
	}

	ctx := NewContext(app, fs, nil)
	app.handleExitCoder(ctx, NewExitError("Default Behavior Error", 42))

	output := fakeErrWriter.String()
	if !strings.Contains(output, "Custom") {
		t.Fatalf("Expected Custom Behavior from Error Handler but got: %s", output)
	}
}

func TestShellCompletionForIncompleteFlags(t *testing.T) {
	app := &App{
		Flags: []Flag{
			&IntFlag{
				Name: "test-completion",
			},
		},
		EnableBashCompletion: true,
		BashComplete: func(ctx *Context) {
			for _, command := range ctx.App.Commands {
				if command.Hidden {
					continue
				}

				for _, name := range command.Names() {
					_, _ = fmt.Fprintln(ctx.App.Writer, name)
				}
			}

			for _, fl := range ctx.App.Flags {
				for _, name := range fl.Names() {
					if name == BashCompletionFlag.Names()[0] {
						continue
					}

					switch name = strings.TrimSpace(name); len(name) {
					case 0:
					case 1:
						_, _ = fmt.Fprintln(ctx.App.Writer, "-"+name)
					default:
						_, _ = fmt.Fprintln(ctx.App.Writer, "--"+name)
					}
				}
			}
		},
		Action: func(ctx *Context) error {
			return fmt.Errorf("should not get here")
		},
		Writer: ioutil.Discard,
	}
	err := app.Run([]string{"", "--test-completion", "--" + "generate-bash-completion"})
	if err != nil {
		t.Errorf("app should not return an error: %s", err)
	}
}

func TestWhenExitSubCommandWithCodeThenAppQuitUnexpectedly(t *testing.T) {
	testCode := 104

	app := newTestApp()
	app.Commands = []*Command{
		{
			Name: "cmd",
			Subcommands: []*Command{
				{
					Name: "subcmd",
					Action: func(c *Context) error {
						return NewExitError("exit error", testCode)
					},
				},
			},
		},
	}

	// set user function as ExitErrHandler
	var exitCodeFromExitErrHandler int
	app.ExitErrHandler = func(c *Context, err error) {
		if exitErr, ok := err.(ExitCoder); ok {
			exitCodeFromExitErrHandler = exitErr.ExitCode()
		}
	}

	// keep and restore original OsExiter
	origExiter := OsExiter
	defer func() {
		OsExiter = origExiter
	}()

	// set user function as OsExiter
	var exitCodeFromOsExiter int
	OsExiter = func(exitCode int) {
		exitCodeFromOsExiter = exitCode
	}

	_ = app.Run([]string{
		"myapp",
		"cmd",
		"subcmd",
	})

	if exitCodeFromOsExiter != 0 {
		t.Errorf("exitCodeFromExitErrHandler should not change, but its value is %v", exitCodeFromOsExiter)
	}

	if exitCodeFromExitErrHandler != testCode {
		t.Errorf("exitCodeFromOsExiter valeu should be %v, but its value is %v", testCode, exitCodeFromExitErrHandler)
	}
}

func newTestApp() *App {
	a := NewApp()
	a.Writer = ioutil.Discard
	return a
}

func TestSetupInitializesBothWriters(t *testing.T) {
	a := &App{}

	a.Setup()

	if a.ErrWriter != os.Stderr {
		t.Errorf("expected a.ErrWriter to be os.Stderr")
	}

	if a.Writer != os.Stdout {
		t.Errorf("expected a.Writer to be os.Stdout")
	}
}

func TestSetupInitializesOnlyNilWriters(t *testing.T) {
	wr := &bytes.Buffer{}
	a := &App{
		ErrWriter: wr,
	}

	a.Setup()

	if a.ErrWriter != wr {
		t.Errorf("expected a.ErrWriter to be a *bytes.Buffer instance")
	}

	if a.Writer != os.Stdout {
		t.Errorf("expected a.Writer to be os.Stdout")
	}
}
