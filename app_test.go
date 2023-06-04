package cli

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/mail"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
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

func ExampleCommand_Run() {
	// set args for examples sake
	os.Args = []string{"greet", "--name", "Jeremy"}

	cmd := &Command{
		Name: "greet",
		Flags: []Flag{
			&StringFlag{Name: "name", Value: "bob", Usage: "a name to say"},
		},
		Action: func(c *Context) error {
			fmt.Printf("Hello %v\n", c.String("name"))
			return nil
		},
		UsageText: "app [first_arg] [second_arg]",
		Authors:   []any{&mail.Address{Name: "Oliver Allen", Address: "oliver@toyshop.example.com"}, "gruffalo@soup-world.example.org"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := cmd.Run(ctx, os.Args); err != nil {
		return
	}
	// Output:
	// Hello Jeremy
}

func ExampleCommand_Run_subcommand() {
	// set args for examples sake
	os.Args = []string{"say", "hi", "english", "--name", "Jeremy"}
	cmd := &Command{
		Name: "say",
		Commands: []*Command{
			{
				Name:        "hello",
				Aliases:     []string{"hi"},
				Usage:       "use it to see a description",
				Description: "This is how we describe hello the function",
				Commands: []*Command{
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = cmd.Run(ctx, os.Args)
	// Output:
	// Hello, Jeremy
}

func ExampleCommand_Run_appHelp() {
	// set args for examples sake
	os.Args = []string{"greet", "help"}

	cmd := &Command{
		Name:        "greet",
		Version:     "0.1.0",
		Description: "This is how we describe greet the app",
		Authors: []any{
			&mail.Address{Name: "Harrison", Address: "harrison@lolwut.example.com"},
			"Oliver Allen  <oliver@toyshop.example.com>",
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
				Action: func(*Context) error {
					fmt.Printf("i like to describe things")
					return nil
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = cmd.Run(ctx, os.Args)
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
	//    "Harrison" <harrison@lolwut.example.com>
	//    Oliver Allen  <oliver@toyshop.example.com>
	//
	// COMMANDS:
	//    describeit, d  use it to see a description
	//    help, h        Shows a list of commands or help for one command
	//
	// GLOBAL OPTIONS:
	//    --name value   a name to say (default: "bob")
	//    --help, -h     show help (default: false)
	//    --version, -v  print the version (default: false)
}

func ExampleCommand_Run_commandHelp() {
	// set args for examples sake
	os.Args = []string{"greet", "h", "describeit"}

	cmd := &Command{
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
				Action: func(*Context) error {
					fmt.Printf("i like to describe things")
					return nil
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = cmd.Run(ctx, os.Args)
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

func ExampleCommand_Run_noAction() {
	cmd := &Command{}
	cmd.Name = "greet"

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = cmd.Run(ctx, []string{"greet"})
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

func ExampleCommand_Run_subcommandNoAction() {
	cmd := &Command{
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = cmd.Run(ctx, []string{"greet", "describeit"})
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

func ExampleCommand_Run_bashComplete_withShortFlag() {
	os.Setenv("SHELL", "bash")
	os.Args = []string{"greet", "-", "--generate-shell-completion"}

	cmd := &Command{
		Name:                  "greet",
		EnableShellCompletion: true,
		Flags: []Flag{
			&IntFlag{
				Name:    "other",
				Aliases: []string{"o"},
			},
			&StringFlag{
				Name:    "xyz",
				Aliases: []string{"x"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = cmd.Run(ctx, os.Args)
	// Output:
	// --other
	// -o
	// --xyz
	// -x
	// --help
	// -h
}

func ExampleCommand_Run_bashComplete_withLongFlag() {
	os.Setenv("SHELL", "bash")
	os.Args = []string{"greet", "--s", "--generate-shell-completion"}

	cmd := &Command{
		Name:                  "greet",
		EnableShellCompletion: true,
		Flags: []Flag{
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
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = cmd.Run(ctx, os.Args)
	// Output:
	// --some-flag
	// --similar-flag
}

func ExampleCommand_Run_bashComplete_withMultipleLongFlag() {
	os.Setenv("SHELL", "bash")
	os.Args = []string{"greet", "--st", "--generate-shell-completion"}

	cmd := &Command{
		Name:                  "greet",
		EnableShellCompletion: true,
		Flags: []Flag{
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
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = cmd.Run(ctx, os.Args)
	// Output:
	// --string
	// --string-flag-2
}

func ExampleCommand_Run_bashComplete() {
	os.Setenv("SHELL", "bash")
	os.Args = []string{"greet", "--generate-shell-completion"}

	cmd := &Command{
		Name:                  "greet",
		EnableShellCompletion: true,
		Commands: []*Command{
			{
				Name:        "describeit",
				Aliases:     []string{"d"},
				Usage:       "use it to see a description",
				Description: "This is how we describe describeit the function",
				Action: func(*Context) error {
					fmt.Printf("i like to describe things")
					return nil
				},
			}, {
				Name:        "next",
				Usage:       "next example",
				Description: "more stuff to see when generating shell completion",
				Action: func(*Context) error {
					fmt.Printf("the next example")
					return nil
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = cmd.Run(ctx, os.Args)
	// Output:
	// describeit
	// d
	// next
	// help
	// h
}

func ExampleCommand_Run_zshComplete() {
	// set args for examples sake
	os.Args = []string{"greet", "--generate-shell-completion"}
	_ = os.Setenv("SHELL", "/usr/bin/zsh")

	cmd := &Command{
		Name:                  "greet",
		EnableShellCompletion: true,
		Commands: []*Command{
			{
				Name:        "describeit",
				Aliases:     []string{"d"},
				Usage:       "use it to see a description",
				Description: "This is how we describe describeit the function",
				Action: func(*Context) error {
					fmt.Printf("i like to describe things")
					return nil
				},
			}, {
				Name:        "next",
				Usage:       "next example",
				Description: "more stuff to see when generating bash completion",
				Action: func(*Context) error {
					fmt.Printf("the next example")
					return nil
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = cmd.Run(ctx, os.Args)
	// Output:
	// describeit:use it to see a description
	// d:use it to see a description
	// next:next example
	// help:Shows a list of commands or help for one command
	// h:Shows a list of commands or help for one command
}

func ExampleCommand_Run_sliceValues() {
	// set args for examples sake
	os.Args = []string{
		"multi_values",
		"--stringSclice", "parsed1,parsed2", "--stringSclice", "parsed3,parsed4",
		"--float64Sclice", "13.3,14.4", "--float64Sclice", "15.5,16.6",
		"--int64Sclice", "13,14", "--int64Sclice", "15,16",
		"--intSclice", "13,14", "--intSclice", "15,16",
	}
	cmd := &Command{
		Name: "multi_values",
		Flags: []Flag{
			&StringSliceFlag{Name: "stringSclice"},
			&Float64SliceFlag{Name: "float64Sclice"},
			&Int64SliceFlag{Name: "int64Sclice"},
			&IntSliceFlag{Name: "intSclice"},
		},
	}
	cmd.Action = func(ctx *Context) error {
		for i, v := range ctx.FlagNames() {
			fmt.Printf("%d-%s %#v\n", i, v, ctx.Value(v))
		}
		err := ctx.Err()
		fmt.Println("error:", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = cmd.Run(ctx, os.Args)
	// Output:
	// 0-float64Sclice []float64{13.3, 14.4, 15.5, 16.6}
	// 1-int64Sclice []int64{13, 14, 15, 16}
	// 2-intSclice []int{13, 14, 15, 16}
	// 3-stringSclice []string{"parsed1", "parsed2", "parsed3", "parsed4"}
	// error: <nil>
}

func ExampleCommand_Run_mapValues() {
	// set args for examples sake
	os.Args = []string{
		"multi_values",
		"--stringMap", "parsed1=parsed two", "--stringMap", "parsed3=",
	}
	cmd := &Command{
		Name: "multi_values",
		Flags: []Flag{
			&StringMapFlag{Name: "stringMap"},
		},
		Action: func(ctx *Context) error {
			for i, v := range ctx.FlagNames() {
				fmt.Printf("%d-%s %#v\n", i, v, ctx.StringMap(v))
			}
			fmt.Printf("notfound %#v\n", ctx.StringMap("notfound"))
			err := ctx.Err()
			fmt.Println("error:", err)
			return err
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = cmd.Run(ctx, os.Args)
	// Output:
	// 0-stringMap map[string]string{"parsed1":"parsed two", "parsed3":""}
	// notfound map[string]string(nil)
	// error: <nil>
}

func TestApp_Run(t *testing.T) {
	s := ""

	cmd := &Command{
		Action: func(c *Context) error {
			s = s + c.Args().First()
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"command", "foo"})
	expect(t, err, nil)
	err = cmd.Run(ctx, []string{"command", "bar"})
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
	cmd := &Command{
		Commands: []*Command{
			{Name: "foobar", Aliases: []string{"f"}},
			{Name: "batbaz", Aliases: []string{"b"}},
		},
	}

	for _, test := range commandAppTests {
		expect(t, cmd.Command(test.name) != nil, test.expected)
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
			cmd := &Command{
				DefaultCommand: test.defaultCmd,
				Commands: []*Command{
					{Name: "foobar", Aliases: []string{"f"}},
					{Name: "batbaz", Aliases: []string{"b"}},
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			err := cmd.Run(ctx, []string{"c", test.cmdName})
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
			cmd := &Command{
				DefaultCommand: test.defaultCmd,
				Commands: []*Command{
					{
						Name:    "foobar",
						Aliases: []string{"f"},
						Commands: []*Command{
							{Name: "jimbob", Aliases: []string{"j"}},
							{Name: "carly"},
						},
					},
					{Name: "batbaz", Aliases: []string{"b"}},
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			err := cmd.Run(ctx, []string{"c", test.cmdName, test.subCmd})
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
			cmd := &Command{
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

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			err := cmd.Run(ctx, appArgs)
			expect(t, err == nil, test.expected)
		})
	}
}

func TestApp_FlagsFromExtPackage(t *testing.T) {
	var someint int
	flag.IntVar(&someint, "epflag", 2, "ext package flag usage")

	// Based on source code we can reset the global flag parsing this way
	defer func() {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	cmd := &Command{
		AllowExtFlags: true,
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
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "-c", "cly", "--epflag", "10"})
	if err != nil {
		t.Error(err)
	}

	if someint != 10 {
		t.Errorf("Expected 10 got %d for someint", someint)
	}

	cmd = &Command{
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
	}

	// this should return an error since epflag shouldnt be registered
	err = cmd.Run(ctx, []string{"foo", "-c", "cly", "--epflag", "10"})
	if err == nil {
		t.Error("Expected error")
	}
}

func TestApp_Setup_defaultsReader(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults()
	expect(t, cmd.Reader, os.Stdin)
}

func TestApp_Setup_defaultsWriter(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults()
	expect(t, cmd.Writer, os.Stdout)
}

func TestApp_RunAsSubcommandParseFlags(t *testing.T) {
	var cCtx *Context

	cmd := &Command{
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"", "foo", "--lang", "spanish", "abcd"})

	expect(t, cCtx.Args().Get(0), "abcd")
	expect(t, cCtx.String("lang"), "spanish")
}

func TestApp_CommandWithFlagBeforeTerminator(t *testing.T) {
	var parsedOption string
	var args Args

	cmd := &Command{
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"", "cmd", "--option", "my-option", "my-arg", "--", "--notARealFlag"})

	expect(t, parsedOption, "my-option")
	expect(t, args.Get(0), "my-arg")
	expect(t, args.Get(1), "--")
	expect(t, args.Get(2), "--notARealFlag")
}

func TestApp_CommandWithDash(t *testing.T) {
	var args Args

	cmd := &Command{
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"", "cmd", "my-arg", "-"})

	expect(t, args.Get(0), "my-arg")
	expect(t, args.Get(1), "-")
}

func TestApp_CommandWithNoFlagBeforeTerminator(t *testing.T) {
	var args Args

	cmd := &Command{
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"", "cmd", "my-arg", "--", "notAFlagAtAll"})

	expect(t, args.Get(0), "my-arg")
	expect(t, args.Get(1), "--")
	expect(t, args.Get(2), "notAFlagAtAll")
}

func TestApp_SkipFlagParsing(t *testing.T) {
	var args Args

	cmd := &Command{
		SkipFlagParsing: true,
		Action: func(c *Context) error {
			args = c.Args()
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"", "--", "my-arg", "notAFlagAtAll"})

	expect(t, args.Get(0), "--")
	expect(t, args.Get(1), "my-arg")
	expect(t, args.Get(2), "notAFlagAtAll")
}

func TestApp_VisibleCommands(t *testing.T) {
	cmd := &Command{
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

	cmd.setupDefaults()
	expected := []*Command{
		cmd.Commands[0],
		cmd.Commands[2], // help
	}
	actual := cmd.VisibleCommands()
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

	cmd := newTestCommand()
	cmd.UseShortOptionHandling = true
	cmd.Flags = []Flag{
		&BoolFlag{Name: "one", Aliases: []string{"o"}},
		&BoolFlag{Name: "two", Aliases: []string{"t"}},
		&StringFlag{Name: "name", Aliases: []string{"n"}},
	}
	cmd.Action = func(c *Context) error {
		one = c.Bool("one")
		two = c.Bool("two")
		name = c.String("name")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"", "-on", expected})
	expect(t, one, true)
	expect(t, two, false)
	expect(t, name, expected)
}

func TestApp_UseShortOptionHandling_missing_value(t *testing.T) {
	cmd := newTestCommand()
	cmd.UseShortOptionHandling = true
	cmd.Flags = []Flag{
		&StringFlag{Name: "name", Aliases: []string{"n"}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"", "-n"})
	expect(t, err, errors.New("flag needs an argument: -n"))
}

func TestApp_UseShortOptionHandlingCommand(t *testing.T) {
	var one, two bool
	var name string
	expected := "expectedName"

	cmd := newTestCommand()
	cmd.UseShortOptionHandling = true
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
	cmd.Commands = []*Command{command}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"", "cmd", "-on", expected})
	expect(t, one, true)
	expect(t, two, false)
	expect(t, name, expected)
}

func TestApp_UseShortOptionHandlingCommand_missing_value(t *testing.T) {
	cmd := newTestCommand()
	cmd.UseShortOptionHandling = true
	command := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{Name: "name", Aliases: []string{"n"}},
		},
	}
	cmd.Commands = []*Command{command}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"", "cmd", "-n"})
	expect(t, err, errors.New("flag needs an argument: -n"))
}

func TestApp_UseShortOptionHandlingSubCommand(t *testing.T) {
	var one, two bool
	var name string
	expected := "expectedName"

	cmd := newTestCommand()
	cmd.UseShortOptionHandling = true
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
	command.Commands = []*Command{subCommand}
	cmd.Commands = []*Command{command}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"", "cmd", "sub", "-on", expected})
	expect(t, err, nil)
	expect(t, one, true)
	expect(t, two, false)
	expect(t, name, expected)
}

func TestApp_UseShortOptionHandlingSubCommand_missing_value(t *testing.T) {
	cmd := newTestCommand()
	cmd.UseShortOptionHandling = true
	command := &Command{
		Name: "cmd",
	}
	subCommand := &Command{
		Name: "sub",
		Flags: []Flag{
			&StringFlag{Name: "name", Aliases: []string{"n"}},
		},
	}
	command.Commands = []*Command{subCommand}
	cmd.Commands = []*Command{command}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"", "cmd", "sub", "-n"})
	expect(t, err, errors.New("flag needs an argument: -n"))
}

func TestApp_UseShortOptionAfterSliceFlag(t *testing.T) {
	var one, two bool
	var name string
	var sliceValDest []string
	var sliceVal []string
	expected := "expectedName"

	cmd := newTestCommand()
	cmd.UseShortOptionHandling = true
	cmd.Flags = []Flag{
		&StringSliceFlag{Name: "env", Aliases: []string{"e"}, Destination: &sliceValDest},
		&BoolFlag{Name: "one", Aliases: []string{"o"}},
		&BoolFlag{Name: "two", Aliases: []string{"t"}},
		&StringFlag{Name: "name", Aliases: []string{"n"}},
	}
	cmd.Action = func(c *Context) error {
		sliceVal = c.StringSlice("env")
		one = c.Bool("one")
		two = c.Bool("two")
		name = c.String("name")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"", "-e", "foo", "-on", expected})
	expect(t, sliceVal, []string{"foo"})
	expect(t, sliceValDest, []string{"foo"})
	expect(t, one, true)
	expect(t, two, false)
	expect(t, name, expected)
}

func TestApp_Float64Flag(t *testing.T) {
	var meters float64

	cmd := &Command{
		Flags: []Flag{
			&Float64Flag{Name: "height", Value: 1.5, Usage: "Set the height, in meters"},
		},
		Action: func(c *Context) error {
			meters = c.Float64("height")
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"", "--height", "1.93"})
	expect(t, meters, 1.93)
}

func TestApp_ParseSliceFlags(t *testing.T) {
	var parsedIntSlice []int
	var parsedStringSlice []string

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "cmd",
				Flags: []Flag{
					&IntSliceFlag{Name: "p", Value: []int{}, Usage: "set one or more ip addr"},
					&StringSliceFlag{Name: "ip", Value: []string{}, Usage: "set one or more ports to open"},
				},
				Action: func(c *Context) error {
					parsedIntSlice = c.IntSlice("p")
					parsedStringSlice = c.StringSlice("ip")
					return nil
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"", "cmd", "-p", "22", "-p", "80", "-ip", "8.8.8.8", "-ip", "8.8.4.4"})

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
	expectedIntSlice := []int{22, 80}
	expectedStringSlice := []string{"8.8.8.8", "8.8.4.4"}

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

	cmd := &Command{
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"", "cmd", "-a", "2", "-str", "A"})

	expectedIntSlice := []int{2}
	expectedStringSlice := []string{"A"}

	if parsedIntSlice[0] != expectedIntSlice[0] {
		t.Errorf("%v does not match %v", parsedIntSlice[0], expectedIntSlice[0])
	}

	if parsedStringSlice[0] != expectedStringSlice[0] {
		t.Errorf("%v does not match %v", parsedIntSlice[0], expectedIntSlice[0])
	}
}

func TestApp_DefaultStdin(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults()

	if cmd.Reader != os.Stdin {
		t.Error("Default input reader not set.")
	}
}

func TestApp_DefaultStdout(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults()

	if cmd.Writer != os.Stdout {
		t.Error("Default output writer not set.")
	}
}

func TestApp_SetStdin(t *testing.T) {
	buf := make([]byte, 12)

	cmd := &Command{
		Name:   "test",
		Reader: strings.NewReader("Hello World!"),
		Action: func(c *Context) error {
			_, err := c.Command.Reader.Read(buf)
			return err
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"help"})
	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if string(buf) != "Hello World!" {
		t.Error("App did not read input from desired reader.")
	}
}

func TestApp_SetStdin_Subcommand(t *testing.T) {
	buf := make([]byte, 12)

	cmd := &Command{
		Name:   "test",
		Reader: strings.NewReader("Hello World!"),
		Commands: []*Command{
			{
				Name: "command",
				Commands: []*Command{
					{
						Name: "subcommand",
						Action: func(c *Context) error {
							_, err := c.Command.Reader.Read(buf)
							return err
						},
					},
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"test", "command", "subcommand"})
	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if string(buf) != "Hello World!" {
		t.Error("App did not read input from desired reader.")
	}
}

func TestApp_SetStdout(t *testing.T) {
	var w bytes.Buffer

	cmd := &Command{
		Name:   "test",
		Writer: &w,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"help"})
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

	cmd := &Command{
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
				Action: func(*Context) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		},
		Flags: []Flag{
			&StringFlag{Name: "opt"},
		},
		Writer: io.Discard,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	// run with the Before() func succeeding
	err = cmd.Run(ctx, []string{"command", "--opt", "succeed", "sub"})

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
	err = cmd.Run(ctx, []string{"command", "--opt", "fail", "sub"})

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
	cmd.After = func(_ *Context) error {
		return afterError
	}

	// run with the Before() func failing, wrapped by After()
	err = cmd.Run(ctx, []string{"command", "--opt", "fail", "sub"})

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

	cmd := &Command{
		EnableShellCompletion: true,
		Before: func(*Context) error {
			counts.Total++
			counts.Before = counts.Total
			return nil
		},
		After: func(*Context) error {
			counts.Total++
			counts.After = counts.Total
			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(*Context) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		},
		Flags: []Flag{
			&StringFlag{Name: "opt"},
		},
		Writer: io.Discard,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	// run with the Before() func succeeding
	err = cmd.Run(ctx, []string{"command", "--opt", "succeed", "sub", "--generate-shell-completion"})

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

	cmd := &Command{
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
				Action: func(*Context) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	// run with the After() func succeeding
	err = cmd.Run(ctx, []string{"command", "--opt", "succeed", "sub"})

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
	err = cmd.Run(ctx, []string{"command", "--opt", "fail", "sub"})

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
	// reset the flags since they are set previously
	cmd.Flags = []Flag{
		&StringFlag{Name: "opt"},
	}

	// run with none args
	err = cmd.Run(ctx, []string{"command"})

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

	cmd := &Command{Writer: io.Discard}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"test", "-h"})

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
				Commands: []*Command{{
					Name:  "mySubCommand",
					Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
				}},
			}},
			expectedAnError: true,
		},
		// assertion: inputting --help, when a required flag is present, does not error
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
				Commands: []*Command{{
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
				Commands: []*Command{{
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
				Commands: []*Command{{
					Name:  "mySubCommand",
					Flags: []Flag{&StringFlag{Name: "requiredFlag", Required: true}},
					Action: func(c *Context) error {
						return nil
					},
				}},
			}},
		},
	}
	for _, test := range tdata {
		t.Run(test.testCase, func(t *testing.T) {
			// setup
			cmd := newTestCommand()
			cmd.Flags = test.appFlags
			cmd.Commands = test.appCommands

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			// logic under test
			err := cmd.Run(ctx, test.appRunInput)

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

	wasCalled := false
	HelpPrinter = func(io.Writer, string, interface{}) {
		wasCalled = true
	}

	cmd := &Command{}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"-h"})

	if wasCalled == false {
		t.Errorf("Help printer expected to be called, but was not")
	}
}

func TestApp_VersionPrinter(t *testing.T) {
	oldPrinter := VersionPrinter
	defer func() {
		VersionPrinter = oldPrinter
	}()

	wasCalled := false
	VersionPrinter = func(*Context) {
		wasCalled = true
	}

	cmd := &Command{}
	ctx := NewContext(cmd, nil, nil)
	ShowVersion(ctx)

	if wasCalled == false {
		t.Errorf("Version printer expected to be called, but was not")
	}
}

func TestApp_CommandNotFound(t *testing.T) {
	counts := &opCounts{}
	cmd := &Command{
		CommandNotFound: func(*Context, string) {
			counts.Total++
			counts.CommandNotFound = counts.Total
		},
		Commands: []*Command{
			{
				Name: "bar",
				Action: func(*Context) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"command", "foo"})

	expect(t, counts.CommandNotFound, 1)
	expect(t, counts.SubCommand, 0)
	expect(t, counts.Total, 1)
}

func TestApp_OrderOfOperations(t *testing.T) {
	counts := &opCounts{}

	resetCounts := func() { counts = &opCounts{} }

	cmd := &Command{
		EnableShellCompletion: true,
		ShellComplete: func(*Context) {
			counts.Total++
			counts.ShellComplete = counts.Total
		},
		OnUsageError: func(*Context, error, bool) error {
			counts.Total++
			counts.OnUsageError = counts.Total
			return errors.New("hay OnUsageError")
		},
		Writer: io.Discard,
	}

	beforeNoError := func(*Context) error {
		counts.Total++
		counts.Before = counts.Total
		return nil
	}

	beforeError := func(*Context) error {
		counts.Total++
		counts.Before = counts.Total
		return errors.New("hay Before")
	}

	cmd.Before = beforeNoError
	cmd.CommandNotFound = func(*Context, string) {
		counts.Total++
		counts.CommandNotFound = counts.Total
	}

	afterNoError := func(*Context) error {
		counts.Total++
		counts.After = counts.Total
		return nil
	}

	afterError := func(*Context) error {
		counts.Total++
		counts.After = counts.Total
		return errors.New("hay After")
	}

	cmd.After = afterNoError
	cmd.Commands = []*Command{
		{
			Name: "bar",
			Action: func(*Context) error {
				counts.Total++
				counts.SubCommand = counts.Total
				return nil
			},
		},
	}

	cmd.Action = func(*Context) error {
		counts.Total++
		counts.Action = counts.Total
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"command", "--nope"})
	expect(t, counts.OnUsageError, 1)
	expect(t, counts.Total, 1)

	resetCounts()

	_ = cmd.Run(ctx, []string{"command", fmt.Sprintf("--%s", "generate-shell-completion")})
	expect(t, counts.ShellComplete, 1)
	expect(t, counts.Total, 1)

	resetCounts()

	oldOnUsageError := cmd.OnUsageError
	cmd.OnUsageError = nil
	_ = cmd.Run(ctx, []string{"command", "--nope"})
	expect(t, counts.Total, 0)
	cmd.OnUsageError = oldOnUsageError

	resetCounts()

	_ = cmd.Run(ctx, []string{"command", "foo"})
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.CommandNotFound, 0)
	expect(t, counts.Action, 2)
	expect(t, counts.After, 3)
	expect(t, counts.Total, 3)

	resetCounts()

	cmd.Before = beforeError
	_ = cmd.Run(ctx, []string{"command", "bar"})
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.After, 2)
	expect(t, counts.Total, 2)
	cmd.Before = beforeNoError

	resetCounts()

	cmd.After = nil
	_ = cmd.Run(ctx, []string{"command", "bar"})
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.SubCommand, 2)
	expect(t, counts.Total, 2)
	cmd.After = afterNoError

	resetCounts()

	cmd.After = afterError
	err := cmd.Run(ctx, []string{"command", "bar"})
	if err == nil {
		t.Fatalf("expected a non-nil error")
	}
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.SubCommand, 2)
	expect(t, counts.After, 3)
	expect(t, counts.Total, 3)
	cmd.After = afterNoError

	resetCounts()

	oldCommands := cmd.Commands
	cmd.Commands = nil
	_ = cmd.Run(ctx, []string{"command"})
	expect(t, counts.OnUsageError, 0)
	expect(t, counts.Before, 1)
	expect(t, counts.Action, 2)
	expect(t, counts.After, 3)
	expect(t, counts.Total, 3)
	cmd.Commands = oldCommands
}

func TestApp_Run_CommandWithSubcommandHasHelpTopic(t *testing.T) {
	subcommandHelpTopics := [][]string{
		{"foo", "--help"},
		{"foo", "-h"},
		{"foo", "help"},
	}

	for _, flagSet := range subcommandHelpTopics {
		t.Run(fmt.Sprintf("checking with flags %v", flagSet), func(t *testing.T) {
			buf := new(bytes.Buffer)

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
				Commands:    []*Command{subCmdBar, subCmdBaz},
				Action:      func(c *Context) error { return nil },
				Writer:      buf,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			err := cmd.Run(ctx, flagSet)
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
	buf := new(bytes.Buffer)

	subCmd := &Command{
		Name:  "bar",
		Usage: "does bar things",
	}

	cmd := &Command{
		Name:        "foo",
		Description: "foo commands",
		Commands:    []*Command{subCmd},
		Writer:      buf,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "bar", "--help"})
	if err != nil {
		t.Error(err)
	}

	output := buf.String()
	expected := "foo bar - does bar things"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %s", expected, output)
	}

	expected = "foo bar [command options] [arguments...]"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %s", expected, output)
	}
}

func TestApp_Run_SubcommandHelpName(t *testing.T) {
	buf := new(bytes.Buffer)

	subCmd := &Command{
		Name:     "bar",
		HelpName: "custom",
		Usage:    "does bar things",
	}

	cmd := &Command{
		Name:        "foo",
		Description: "foo commands",
		Commands:    []*Command{subCmd},
		Writer:      buf,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "bar", "--help"})
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
	buf := new(bytes.Buffer)

	subCmd := &Command{
		Name:  "bar",
		Usage: "does bar things",
	}

	cmd := &Command{
		Name:        "foo",
		HelpName:    "custom",
		Description: "foo commands",
		Commands:    []*Command{subCmd},
		Writer:      buf,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "bar", "--help"})
	if err != nil {
		t.Error(err)
	}

	output := buf.String()

	expected := "custom bar - does bar things"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %s", expected, output)
	}

	expected = "custom bar [command options] [arguments...]"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %s", expected, output)
	}
}

func TestApp_Run_CommandSubcommandHelpName(t *testing.T) {
	buf := new(bytes.Buffer)

	subCmd := &Command{
		Name:     "bar",
		HelpName: "custom",
		Usage:    "does bar things",
	}

	cmd := &Command{
		Name:        "foo",
		Usage:       "foo commands",
		Description: "This is a description",
		Commands:    []*Command{subCmd},
		Writer:      buf,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "--help"})
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
	tests := []struct {
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

			cmd := &Command{
				Name:     "boom",
				Usage:    "make an explosive entrance",
				Writer:   buf,
				HideHelp: tt.hideHelp,
				Action: func(*Context) error {
					buf.WriteString("boom I say!")
					return nil
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			err := cmd.Run(ctx, tt.helpArguments)
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
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
	versionArguments := [][]string{{"boom", "--version"}, {"boom", "-v"}}

	for _, args := range versionArguments {
		t.Run(fmt.Sprintf("checking with arguments %v", args), func(t *testing.T) {
			buf := new(bytes.Buffer)

			cmd := &Command{
				Name:    "boom",
				Usage:   "make an explosive entrance",
				Version: "0.1.0",
				Writer:  buf,
				Action: func(*Context) error {
					buf.WriteString("boom I say!")
					return nil
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			err := cmd.Run(ctx, args)
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

	cmd := &Command{
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{"categories"})

	expect := commandCategories([]*commandCategory{
		{
			name: "1",
			commands: []*Command{
				cmd.Commands[0],
				cmd.Commands[1],
			},
		},
		{
			name: "2",
			commands: []*Command{
				cmd.Commands[2],
			},
		},
	})

	if !reflect.DeepEqual(cmd.categories, &expect) {
		t.Fatalf("expected categories %#v, to equal %#v", cmd.categories, &expect)
	}

	output := buf.String()

	if !strings.Contains(output, "1:\n     command1") {
		t.Errorf("want buffer to include category %q, did not: \n%q", "1:\n     command1", output)
	}
}

func TestApp_VisibleCategories(t *testing.T) {
	cmd := &Command{
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
				cmd.Commands[1],
			},
		},
		&commandCategory{
			name: "3",
			commands: []*Command{
				cmd.Commands[2],
			},
		},
	}

	cmd.setupDefaults()
	expect(t, expected, cmd.VisibleCategories())

	cmd = &Command{
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
				cmd.Commands[2],
			},
		},
	}

	cmd.setupDefaults()
	expect(t, expected, cmd.VisibleCategories())

	cmd = &Command{
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

	cmd.setupDefaults()
	expect(t, []CommandCategory{}, cmd.VisibleCategories())
}

func TestApp_VisibleFlagCategories(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&StringFlag{
				Name: "strd", // no category set
			},
			&Int64Flag{
				Name:     "intd",
				Aliases:  []string{"altd1", "altd2"},
				Category: "cat1",
			},
		},
	}
	cmd.setupDefaults()
	vfc := cmd.VisibleFlagCategories()
	if len(vfc) != 1 {
		t.Fatalf("unexpected visible flag categories %+v", vfc)
	}
	if vfc[0].Name() != "cat1" {
		t.Errorf("expected category name cat1 got %s", vfc[0].Name())
	}
	if len(vfc[0].Flags()) != 1 {
		t.Fatalf("expected flag category to have just one flag got %+v", vfc[0].Flags())
	}

	fl := vfc[0].Flags()[0]
	if !reflect.DeepEqual(fl.Names(), []string{"intd", "altd1", "altd2"}) {
		t.Errorf("unexpected flag %+v", fl.Names())
	}
}

func TestApp_Run_DoesNotOverwriteErrorFromBefore(t *testing.T) {
	cmd := &Command{
		Action: func(c *Context) error { return nil },
		Before: func(c *Context) error { return fmt.Errorf("before error") },
		After:  func(c *Context) error { return fmt.Errorf("after error") },
		Writer: io.Discard,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo"})
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
	cmd := &Command{
		Commands: []*Command{
			{
				Commands: []*Command{
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "bar"})
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
	cmd := &Command{
		Flags: []Flag{
			&IntFlag{Name: "flag"},
		},
		OnUsageError: func(_ *Context, err error, isSubcommand bool) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "--flag=wrong"})
	if err == nil {
		t.Fatalf("expected to receive error from Run, got none")
	}

	if !strings.HasPrefix(err.Error(), "intercepted: invalid value") {
		t.Errorf("Expect an intercepted error, but got \"%v\"", err)
	}
}

func TestApp_OnUsageError_WithWrongFlagValue_ForSubcommand(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&IntFlag{Name: "flag"},
		},
		OnUsageError: func(_ *Context, err error, isSubcommand bool) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "--flag=wrong", "bar"})
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

func (c *customBoolFlag) RunAction(*Context) error {
	return nil
}

func (c *customBoolFlag) IsSet() bool {
	return false
}

func (c *customBoolFlag) IsRequired() bool {
	return false
}

func (c *customBoolFlag) IsVisible() bool {
	return false
}

func (c *customBoolFlag) GetCategory() string {
	return ""
}

func (c *customBoolFlag) GetEnvVars() []string {
	return nil
}

func (c *customBoolFlag) GetDefaultText() string {
	return ""
}

func TestCustomFlagsUnused(t *testing.T) {
	cmd := &Command{
		Flags:  []Flag{&customBoolFlag{"custom"}},
		Writer: io.Discard,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo"})
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestCustomFlagsUsed(t *testing.T) {
	cmd := &Command{
		Flags:  []Flag{&customBoolFlag{"custom"}},
		Writer: io.Discard,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "--custom=bar"})
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestCustomHelpVersionFlags(t *testing.T) {
	cmd := &Command{
		Writer: io.Discard,
	}

	// Be sure to reset the global flags
	defer func(helpFlag Flag, versionFlag Flag) {
		HelpFlag = helpFlag.(*BoolFlag)
		VersionFlag = versionFlag.(*BoolFlag)
	}(HelpFlag, VersionFlag)

	HelpFlag = &customBoolFlag{"help-custom"}
	VersionFlag = &customBoolFlag{"version-custom"}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "--help-custom=bar"})
	if err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestHandleExitCoder_Default(t *testing.T) {
	app := newTestCommand()
	fs, err := flagSet(app.Name, app.Flags)
	if err != nil {
		t.Errorf("error creating FlagSet: %s", err)
	}

	ctx := NewContext(app, fs, nil)
	app.handleExitCoder(ctx, Exit("Default Behavior Error", 42))

	output := fakeErrWriter.String()
	if !strings.Contains(output, "Default") {
		t.Fatalf("Expected Default Behavior from Error Handler but got: %s", output)
	}
}

func TestHandleExitCoder_Custom(t *testing.T) {
	cmd := newTestCommand()
	fs, err := flagSet(cmd.Name, cmd.Flags)
	if err != nil {
		t.Errorf("error creating FlagSet: %s", err)
	}

	cmd.ExitErrHandler = func(_ *Context, _ error) {
		_, _ = fmt.Fprintln(ErrWriter, "I'm a Custom error handler, I print what I want!")
	}

	ctx := NewContext(cmd, fs, nil)
	cmd.handleExitCoder(ctx, Exit("Default Behavior Error", 42))

	output := fakeErrWriter.String()
	if !strings.Contains(output, "Custom") {
		t.Fatalf("Expected Custom Behavior from Error Handler but got: %s", output)
	}
}

func TestShellCompletionForIncompleteFlags(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&IntFlag{
				Name: "test-completion",
			},
		},
		EnableShellCompletion: true,
		ShellComplete: func(ctx *Context) {
			for _, command := range ctx.Command.Commands {
				if command.Hidden {
					continue
				}

				for _, name := range command.Names() {
					_, _ = fmt.Fprintln(ctx.Command.Writer, name)
				}
			}

			for _, fl := range ctx.Command.Flags {
				for _, name := range fl.Names() {
					if name == BashCompletionFlag.Names()[0] {
						continue
					}

					switch name = strings.TrimSpace(name); len(name) {
					case 0:
					case 1:
						_, _ = fmt.Fprintln(ctx.Command.Writer, "-"+name)
					default:
						_, _ = fmt.Fprintln(ctx.Command.Writer, "--"+name)
					}
				}
			}
		},
		Action: func(ctx *Context) error {
			return fmt.Errorf("should not get here")
		},
		Writer: io.Discard,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"", "--test-completion", "--" + "generate-shell-completion"})
	if err != nil {
		t.Errorf("app should not return an error: %s", err)
	}
}

func TestWhenExitSubCommandWithCodeThenAppQuitUnexpectedly(t *testing.T) {
	testCode := 104

	cmd := newTestCommand()
	cmd.Commands = []*Command{
		{
			Name: "cmd",
			Commands: []*Command{
				{
					Name: "subcmd",
					Action: func(c *Context) error {
						return Exit("exit error", testCode)
					},
				},
			},
		},
	}

	// set user function as ExitErrHandler
	var exitCodeFromExitErrHandler int
	cmd.ExitErrHandler = func(_ *Context, err error) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	_ = cmd.Run(ctx, []string{
		"myapp",
		"cmd",
		"subcmd",
	})

	if exitCodeFromOsExiter != 0 {
		t.Errorf("exitCodeFromExitErrHandler should not change, but its value is %v", exitCodeFromOsExiter)
	}

	if exitCodeFromExitErrHandler != testCode {
		t.Errorf("exitCodeFromOsExiter value should be %v, but its value is %v", testCode, exitCodeFromExitErrHandler)
	}
}

func newTestCommand() *Command {
	return &Command{Writer: io.Discard}
}

func TestSetupInitializesBothWriters(t *testing.T) {
	cmd := &Command{}

	cmd.setupDefaults()

	if cmd.ErrWriter != os.Stderr {
		t.Errorf("expected a.ErrWriter to be os.Stderr")
	}

	if cmd.Writer != os.Stdout {
		t.Errorf("expected a.Writer to be os.Stdout")
	}
}

func TestSetupInitializesOnlyNilWriters(t *testing.T) {
	wr := &bytes.Buffer{}
	cmd := &Command{
		ErrWriter: wr,
	}

	cmd.setupDefaults()

	if cmd.ErrWriter != wr {
		t.Errorf("expected a.ErrWriter to be a *bytes.Buffer instance")
	}

	if cmd.Writer != os.Stdout {
		t.Errorf("expected a.Writer to be os.Stdout")
	}
}

func TestFlagAction(t *testing.T) {
	stringFlag := &StringFlag{
		Name: "f_string",
		Action: func(c *Context, v string) error {
			if v == "" {
				return fmt.Errorf("empty string")
			}
			_, err := c.Command.Writer.Write([]byte(v + " "))
			return err
		},
	}
	cmd := &Command{
		Name: "app",
		Commands: []*Command{
			{
				Name:   "c1",
				Flags:  []Flag{stringFlag},
				Action: func(ctx *Context) error { return nil },
				Commands: []*Command{
					{
						Name:   "sub1",
						Action: func(ctx *Context) error { return nil },
						Flags:  []Flag{stringFlag},
					},
				},
			},
		},
		Flags: []Flag{
			stringFlag,
			&StringFlag{
				Name: "f_no_action",
			},
			&StringSliceFlag{
				Name: "f_string_slice",
				Action: func(c *Context, v []string) error {
					if v[0] == "err" {
						return fmt.Errorf("error string slice")
					}
					_, err := c.Command.Writer.Write([]byte(fmt.Sprintf("%v ", v)))
					return err
				},
			},
			&BoolFlag{
				Name: "f_bool",
				Action: func(c *Context, v bool) error {
					if !v {
						return fmt.Errorf("value is false")
					}
					_, err := c.Command.Writer.Write([]byte(fmt.Sprintf("%t ", v)))
					return err
				},
			},
			&DurationFlag{
				Name: "f_duration",
				Action: func(c *Context, v time.Duration) error {
					if v == 0 {
						return fmt.Errorf("empty duration")
					}
					_, err := c.Command.Writer.Write([]byte(v.String() + " "))
					return err
				},
			},
			&Float64Flag{
				Name: "f_float64",
				Action: func(c *Context, v float64) error {
					if v < 0 {
						return fmt.Errorf("negative float64")
					}
					_, err := c.Command.Writer.Write([]byte(strconv.FormatFloat(v, 'f', -1, 64) + " "))
					return err
				},
			},
			&Float64SliceFlag{
				Name: "f_float64_slice",
				Action: func(c *Context, v []float64) error {
					if len(v) > 0 && v[0] < 0 {
						return fmt.Errorf("invalid float64 slice")
					}
					_, err := c.Command.Writer.Write([]byte(fmt.Sprintf("%v ", v)))
					return err
				},
			},
			&IntFlag{
				Name: "f_int",
				Action: func(c *Context, v int) error {
					if v < 0 {
						return fmt.Errorf("negative int")
					}
					_, err := c.Command.Writer.Write([]byte(fmt.Sprintf("%v ", v)))
					return err
				},
			},
			&IntSliceFlag{
				Name: "f_int_slice",
				Action: func(c *Context, v []int) error {
					if len(v) > 0 && v[0] < 0 {
						return fmt.Errorf("invalid int slice")
					}
					_, err := c.Command.Writer.Write([]byte(fmt.Sprintf("%v ", v)))
					return err
				},
			},
			&Int64Flag{
				Name: "f_int64",
				Action: func(c *Context, v int64) error {
					if v < 0 {
						return fmt.Errorf("negative int64")
					}
					_, err := c.Command.Writer.Write([]byte(fmt.Sprintf("%v ", v)))
					return err
				},
			},
			&Int64SliceFlag{
				Name: "f_int64_slice",
				Action: func(c *Context, v []int64) error {
					if len(v) > 0 && v[0] < 0 {
						return fmt.Errorf("invalid int64 slice")
					}
					_, err := c.Command.Writer.Write([]byte(fmt.Sprintf("%v ", v)))
					return err
				},
			},
			&TimestampFlag{
				Name: "f_timestamp",
				Config: TimestampConfig{
					Layout: "2006-01-02 15:04:05",
				},
				Action: func(c *Context, v time.Time) error {
					if v.IsZero() {
						return fmt.Errorf("zero timestamp")
					}
					_, err := c.Command.Writer.Write([]byte(v.Format(time.RFC3339) + " "))
					return err
				},
			},
			&UintFlag{
				Name: "f_uint",
				Action: func(c *Context, v uint) error {
					if v == 0 {
						return fmt.Errorf("zero uint")
					}
					_, err := c.Command.Writer.Write([]byte(fmt.Sprintf("%v ", v)))
					return err
				},
			},
			&Uint64Flag{
				Name: "f_uint64",
				Action: func(c *Context, v uint64) error {
					if v == 0 {
						return fmt.Errorf("zero uint64")
					}
					_, err := c.Command.Writer.Write([]byte(fmt.Sprintf("%v ", v)))
					return err
				},
			},
			&StringMapFlag{
				Name: "f_string_map",
				Action: func(c *Context, v map[string]string) error {
					if _, ok := v["err"]; ok {
						return fmt.Errorf("error string map")
					}
					_, err := c.Command.Writer.Write([]byte(fmt.Sprintf("%v", v)))
					return err
				},
			},
		},
		Action: func(ctx *Context) error { return nil },
	}

	tests := []struct {
		name string
		args []string
		err  error
		exp  string
	}{
		{
			name: "flag_string",
			args: []string{"app", "--f_string=string"},
			exp:  "string ",
		},
		{
			name: "flag_string_error",
			args: []string{"app", "--f_string="},
			err:  fmt.Errorf("empty string"),
		},
		{
			name: "flag_string_slice",
			args: []string{"app", "--f_string_slice=s1,s2,s3"},
			exp:  "[s1 s2 s3] ",
		},
		{
			name: "flag_string_slice_error",
			args: []string{"app", "--f_string_slice=err"},
			err:  fmt.Errorf("error string slice"),
		},
		{
			name: "flag_bool",
			args: []string{"app", "--f_bool"},
			exp:  "true ",
		},
		{
			name: "flag_bool_error",
			args: []string{"app", "--f_bool=false"},
			err:  fmt.Errorf("value is false"),
		},
		{
			name: "flag_duration",
			args: []string{"app", "--f_duration=1h30m20s"},
			exp:  "1h30m20s ",
		},
		{
			name: "flag_duration_error",
			args: []string{"app", "--f_duration=0"},
			err:  fmt.Errorf("empty duration"),
		},
		{
			name: "flag_float64",
			args: []string{"app", "--f_float64=3.14159"},
			exp:  "3.14159 ",
		},
		{
			name: "flag_float64_error",
			args: []string{"app", "--f_float64=-1"},
			err:  fmt.Errorf("negative float64"),
		},
		{
			name: "flag_float64_slice",
			args: []string{"app", "--f_float64_slice=1.1,2.2,3.3"},
			exp:  "[1.1 2.2 3.3] ",
		},
		{
			name: "flag_float64_slice_error",
			args: []string{"app", "--f_float64_slice=-1"},
			err:  fmt.Errorf("invalid float64 slice"),
		},
		{
			name: "flag_int",
			args: []string{"app", "--f_int=1"},
			exp:  "1 ",
		},
		{
			name: "flag_int_error",
			args: []string{"app", "--f_int=-1"},
			err:  fmt.Errorf("negative int"),
		},
		{
			name: "flag_int_slice",
			args: []string{"app", "--f_int_slice=1,2,3"},
			exp:  "[1 2 3] ",
		},
		{
			name: "flag_int_slice_error",
			args: []string{"app", "--f_int_slice=-1"},
			err:  fmt.Errorf("invalid int slice"),
		},
		{
			name: "flag_int64",
			args: []string{"app", "--f_int64=1"},
			exp:  "1 ",
		},
		{
			name: "flag_int64_error",
			args: []string{"app", "--f_int64=-1"},
			err:  fmt.Errorf("negative int64"),
		},
		{
			name: "flag_int64_slice",
			args: []string{"app", "--f_int64_slice=1,2,3"},
			exp:  "[1 2 3] ",
		},
		{
			name: "flag_int64_slice",
			args: []string{"app", "--f_int64_slice=-1"},
			err:  fmt.Errorf("invalid int64 slice"),
		},
		{
			name: "flag_timestamp",
			args: []string{"app", "--f_timestamp", "2022-05-01 02:26:20"},
			exp:  "2022-05-01T02:26:20Z ",
		},
		{
			name: "flag_timestamp_error",
			args: []string{"app", "--f_timestamp", "0001-01-01 00:00:00"},
			err:  fmt.Errorf("zero timestamp"),
		},
		{
			name: "flag_uint",
			args: []string{"app", "--f_uint=1"},
			exp:  "1 ",
		},
		{
			name: "flag_uint_error",
			args: []string{"app", "--f_uint=0"},
			err:  fmt.Errorf("zero uint"),
		},
		{
			name: "flag_uint64",
			args: []string{"app", "--f_uint64=1"},
			exp:  "1 ",
		},
		{
			name: "flag_uint64_error",
			args: []string{"app", "--f_uint64=0"},
			err:  fmt.Errorf("zero uint64"),
		},
		{
			name: "flag_no_action",
			args: []string{"app", "--f_no_action="},
			exp:  "",
		},
		{
			name: "command_flag",
			args: []string{"app", "c1", "--f_string=c1"},
			exp:  "c1 ",
		},
		{
			name: "subCommand_flag",
			args: []string{"app", "c1", "sub1", "--f_string=sub1"},
			exp:  "sub1 ",
		},
		{
			name: "mixture",
			args: []string{"app", "--f_string=app", "--f_uint=1", "--f_int_slice=1,2,3", "--f_duration=1h30m20s", "c1", "--f_string=c1", "sub1", "--f_string=sub1"},
			exp:  "app 1h30m20s [1 2 3] 1 c1 sub1 ",
		},
		{
			name: "flag_string_map",
			args: []string{"app", "--f_string_map=s1=s2,s3="},
			exp:  "map[s1:s2 s3:]",
		},
		{
			name: "flag_string_map_error",
			args: []string{"app", "--f_string_map=err="},
			err:  fmt.Errorf("error string map"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			cmd.Writer = buf

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			err := cmd.Run(ctx, test.args)
			if test.err != nil {
				expect(t, err, test.err)
			} else {
				expect(t, err, nil)
				expect(t, buf.String(), test.exp)
			}
		})
	}
}

func TestPersistentFlag(t *testing.T) {
	var topInt, topPersistentInt, subCommandInt, appOverrideInt int
	var appFlag string
	var appOverrideCmdInt int64
	var appSliceFloat64 []float64
	var persistentAppSliceInt []int64

	cmd := &Command{
		Flags: []Flag{
			&StringFlag{
				Name:        "persistentAppFlag",
				Persistent:  true,
				Destination: &appFlag,
			},
			&Int64SliceFlag{
				Name:        "persistentAppSliceFlag",
				Persistent:  true,
				Destination: &persistentAppSliceInt,
			},
			&Float64SliceFlag{
				Name:       "persistentAppFloatSliceFlag",
				Persistent: true,
				Value:      []float64{11.3, 12.5},
			},
			&IntFlag{
				Name:        "persistentAppOverrideFlag",
				Persistent:  true,
				Destination: &appOverrideInt,
			},
		},
		Commands: []*Command{
			{
				Name: "cmd",
				Flags: []Flag{
					&IntFlag{
						Name:        "cmdFlag",
						Destination: &topInt,
					},
					&IntFlag{
						Name:        "cmdPersistentFlag",
						Persistent:  true,
						Destination: &topPersistentInt,
					},
					&Int64Flag{
						Name:        "paof",
						Aliases:     []string{"persistentAppOverrideFlag"},
						Destination: &appOverrideCmdInt,
					},
				},
				Commands: []*Command{
					{
						Name: "subcmd",
						Flags: []Flag{
							&IntFlag{
								Name:        "cmdFlag",
								Destination: &subCommandInt,
							},
						},
						Action: func(ctx *Context) error {
							appSliceFloat64 = ctx.Float64Slice("persistentAppFloatSliceFlag")
							return nil
						},
					},
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"app",
		"--persistentAppFlag", "hello",
		"--persistentAppSliceFlag", "100",
		"--persistentAppOverrideFlag", "102",
		"cmd",
		"--cmdFlag", "12",
		"--persistentAppSliceFlag", "102",
		"--persistentAppFloatSliceFlag", "102.455",
		"--paof", "105",
		"subcmd",
		"--cmdPersistentFlag", "20",
		"--cmdFlag", "11",
		"--persistentAppFlag", "bar",
		"--persistentAppSliceFlag", "130",
		"--persistentAppFloatSliceFlag", "3.1445",
	})

	if err != nil {
		t.Fatal(err)
	}

	if appFlag != "bar" {
		t.Errorf("Expected 'bar' got %s", appFlag)
	}

	if topInt != 12 {
		t.Errorf("Expected 12 got %d", topInt)
	}

	if topPersistentInt != 20 {
		t.Errorf("Expected 20 got %d", topPersistentInt)
	}

	// this should be changed from app since
	// cmd overrides it
	if appOverrideInt != 102 {
		t.Errorf("Expected 102 got %d", appOverrideInt)
	}

	if subCommandInt != 11 {
		t.Errorf("Expected 11 got %d", subCommandInt)
	}

	if appOverrideCmdInt != 105 {
		t.Errorf("Expected 105 got %d", appOverrideCmdInt)
	}

	expectedInt := []int64{100, 102, 130}
	if !reflect.DeepEqual(persistentAppSliceInt, expectedInt) {
		t.Errorf("Expected %v got %d", expectedInt, persistentAppSliceInt)
	}

	expectedFloat := []float64{102.455, 3.1445}
	if !reflect.DeepEqual(appSliceFloat64, expectedFloat) {
		t.Errorf("Expected %f got %f", expectedFloat, appSliceFloat64)
	}

}

func TestFlagDuplicates(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&StringFlag{
				Name:     "sflag",
				OnlyOnce: true,
			},
			&Int64SliceFlag{
				Name: "isflag",
			},
			&Float64SliceFlag{
				Name:     "fsflag",
				OnlyOnce: true,
			},
			&IntFlag{
				Name: "iflag",
			},
		},
		Action: func(ctx *Context) error {
			return nil
		},
	}

	tests := []struct {
		name        string
		args        []string
		errExpected bool
	}{
		{
			name: "all args present once",
			args: []string{"foo", "--sflag", "hello", "--isflag", "1", "--isflag", "2", "--fsflag", "2.0", "--iflag", "10"},
		},
		{
			name: "duplicate non slice flag(duplicatable)",
			args: []string{"foo", "--sflag", "hello", "--isflag", "1", "--isflag", "2", "--fsflag", "2.0", "--iflag", "10", "--iflag", "20"},
		},
		{
			name:        "duplicate non slice flag(non duplicatable)",
			args:        []string{"foo", "--sflag", "hello", "--isflag", "1", "--isflag", "2", "--fsflag", "2.0", "--iflag", "10", "--sflag", "trip"},
			errExpected: true,
		},
		{
			name:        "duplicate slice flag(non duplicatable)",
			args:        []string{"foo", "--sflag", "hello", "--isflag", "1", "--isflag", "2", "--fsflag", "2.0", "--fsflag", "3.0", "--iflag", "10"},
			errExpected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			err := cmd.Run(ctx, test.args)
			if test.errExpected && err == nil {
				t.Error("expected error")
			} else if !test.errExpected && err != nil {
				t.Error(err)
			}
		})
	}
}

func TestShorthandCommand(t *testing.T) {
	af := func(p *int) ActionFunc {
		return func(ctx *Context) error {
			*p = *p + 1
			return nil
		}
	}

	var cmd1, cmd2 int

	cmd := &Command{
		PrefixMatchCommands: true,
		Commands: []*Command{
			{
				Name:    "cthdisd",
				Aliases: []string{"cth"},
				Action:  af(&cmd1),
			},
			{
				Name:    "cthertoop",
				Aliases: []string{"cer"},
				Action:  af(&cmd2),
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "cth"})
	if err != nil {
		t.Error(err)
	}

	if cmd1 != 1 && cmd2 != 0 {
		t.Errorf("Expected command1 to be trigerred once but didnt %d %d", cmd1, cmd2)
	}

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(ctx, []string{"foo", "cthd"})
	if err != nil {
		t.Error(err)
	}

	if cmd1 != 1 && cmd2 != 0 {
		t.Errorf("Expected command1 to be trigerred once but didnt %d %d", cmd1, cmd2)
	}

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(ctx, []string{"foo", "cthe"})
	if err != nil {
		t.Error(err)
	}

	if cmd1 != 1 && cmd2 != 0 {
		t.Errorf("Expected command1 to be trigerred once but didnt %d %d", cmd1, cmd2)
	}

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(ctx, []string{"foo", "cthert"})
	if err != nil {
		t.Error(err)
	}

	if cmd1 != 0 && cmd2 != 1 {
		t.Errorf("Expected command1 to be trigerred once but didnt %d %d", cmd1, cmd2)
	}

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(ctx, []string{"foo", "cthet"})
	if err != nil {
		t.Error(err)
	}

	if cmd1 != 0 && cmd2 != 1 {
		t.Errorf("Expected command1 to be trigerred once but didnt %d %d", cmd1, cmd2)
	}
}
