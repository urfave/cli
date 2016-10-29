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

	app.Run(os.Args)
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
	app.Run(os.Args)
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
	//      describeit, d  use it to see a description
	//      help, h        Shows a list of commands or help for one command
	//
	// GLOBAL OPTIONS:
	//    --name value   a name to say (default: "bob")
	//    --help, -h     show help (default: false)
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
	app.Run(os.Args)
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

func ExampleApp_Run_shellComplete() {
	// set args for examples sake
	os.Args = []string{"greet", "--generate-completion"}

	app := &App{
		Name: "greet",
		EnableShellCompletion: true,
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

	app.Run(os.Args)
	// Output:
	// describeit
	// d
	// next
	// help
	// h
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
	app := &App{}
	fooCommand := &Command{Name: "foobar", Aliases: []string{"f"}}
	batCommand := &Command{Name: "batbaz", Aliases: []string{"b"}}
	app.Commands = []*Command{
		fooCommand,
		batCommand,
	}

	for _, test := range commandAppTests {
		expect(t, app.Command(test.name) != nil, test.expected)
	}
}

func TestApp_RunAsSubcommandParseFlags(t *testing.T) {
	var context *Context

	a := &App{
		Commands: []*Command{
			{
				Name: "foo",
				Action: func(c *Context) error {
					context = c
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
	a.Run([]string{"", "foo", "--lang", "spanish", "abcd"})

	expect(t, context.Args().Get(0), "abcd")
	expect(t, context.String("lang"), "spanish")
}

func TestApp_RunAsSubCommandIncorrectUsage(t *testing.T) {
	a := App{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{Name: "--foo"},
		},
		Writer: bytes.NewBufferString(""),
	}

	set := flag.NewFlagSet("", flag.ContinueOnError)
	set.Parse([]string{"", "---foo"})
	c := &Context{flagSet: set}

	err := a.RunAsSubcommand(c)

	expect(t, err, errors.New("bad flag syntax: ---foo"))
}

func TestApp_CommandWithFlagBeforeTerminator(t *testing.T) {
	var parsedOption string
	var args Args

	app := &App{}
	command := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{Name: "option", Value: "", Usage: "some option"},
		},
		Action: func(c *Context) error {
			parsedOption = c.String("option")
			args = c.Args()
			return nil
		},
	}
	app.Commands = []*Command{command}

	app.Run([]string{"", "cmd", "--option", "my-option", "my-arg", "--", "--notARealFlag"})

	expect(t, parsedOption, "my-option")
	expect(t, args.Get(0), "my-arg")
	expect(t, args.Get(1), "--")
	expect(t, args.Get(2), "--notARealFlag")
}

func TestApp_CommandWithDash(t *testing.T) {
	var args Args

	app := &App{}
	command := &Command{
		Name: "cmd",
		Action: func(c *Context) error {
			args = c.Args()
			return nil
		},
	}
	app.Commands = []*Command{command}

	app.Run([]string{"", "cmd", "my-arg", "-"})

	expect(t, args.Get(0), "my-arg")
	expect(t, args.Get(1), "-")
}

func TestApp_CommandWithNoFlagBeforeTerminator(t *testing.T) {
	var args Args

	app := &App{}
	command := &Command{
		Name: "cmd",
		Action: func(c *Context) error {
			args = c.Args()
			return nil
		},
	}
	app.Commands = []*Command{command}

	app.Run([]string{"", "cmd", "my-arg", "--", "notAFlagAtAll"})

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

	app.Run([]string{"", "--height", "1.93"})
	expect(t, meters, 1.93)
}

func TestApp_ParseSliceFlags(t *testing.T) {
	var parsedOption, firstArg string
	var parsedIntSlice []int
	var parsedStringSlice []string

	app := &App{}
	command := &Command{
		Name: "cmd",
		Flags: []Flag{
			&IntSliceFlag{Name: "p", Value: NewIntSlice(), Usage: "set one or more ip addr"},
			&StringSliceFlag{Name: "ip", Value: NewStringSlice(), Usage: "set one or more ports to open"},
		},
		Action: func(c *Context) error {
			parsedIntSlice = c.IntSlice("p")
			parsedStringSlice = c.StringSlice("ip")
			parsedOption = c.String("option")
			firstArg = c.Args().First()
			return nil
		},
	}
	app.Commands = []*Command{command}

	app.Run([]string{"", "cmd", "-p", "22", "-p", "80", "-ip", "8.8.8.8", "-ip", "8.8.4.4", "my-arg"})

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

	app := &App{}
	command := &Command{
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
	}
	app.Commands = []*Command{command}

	app.Run([]string{"", "cmd", "-a", "2", "-str", "A", "my-arg"})

	var expectedIntSlice = []int{2}
	var expectedStringSlice = []string{"A"}

	if parsedIntSlice[0] != expectedIntSlice[0] {
		t.Errorf("%v does not match %v", parsedIntSlice[0], expectedIntSlice[0])
	}

	if parsedStringSlice[0] != expectedStringSlice[0] {
		t.Errorf("%v does not match %v", parsedIntSlice[0], expectedIntSlice[0])
	}
}

func TestApp_DefaultStdout(t *testing.T) {
	app := &App{}
	app.Setup()

	if app.Writer != os.Stdout {
		t.Error("Default output writer not set.")
	}
}

type mockWriter struct {
	written []byte
}

func (fw *mockWriter) Write(p []byte) (n int, err error) {
	if fw.written == nil {
		fw.written = p
	} else {
		fw.written = append(fw.written, p...)
	}

	return len(p), nil
}

func (fw *mockWriter) GetWritten() (b []byte) {
	return fw.written
}

func TestApp_SetStdout(t *testing.T) {
	w := &mockWriter{}

	app := &App{
		Name:   "test",
		Writer: w,
	}

	err := app.Run([]string{"help"})

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if len(w.written) == 0 {
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
	app.Run([]string{"-h"})

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

	app.Run([]string{"command", "foo"})

	expect(t, counts.CommandNotFound, 1)
	expect(t, counts.SubCommand, 0)
	expect(t, counts.Total, 1)
}

func TestApp_OrderOfOperations(t *testing.T) {
	counts := &opCounts{}

	resetCounts := func() { counts = &opCounts{} }

	app := &App{
		EnableShellCompletion: true,
		ShellComplete: func(c *Context) {
			counts.Total++
			counts.ShellComplete = counts.Total
		},
		OnUsageError: func(c *Context, err error, isSubcommand bool) error {
			counts.Total++
			counts.OnUsageError = counts.Total
			return errors.New("hay OnUsageError")
		},
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

	_ = app.Run([]string{"command", "--generate-completion"})
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
		t.Logf("==> checking with flags %v", flagSet)

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
		t.Logf("output: %q\n", buf.Bytes())

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
		Description: "foo commands",
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

	expected = "base foo command [command options] [arguments...]"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %q", expected, output)
	}
}

func TestApp_Run_Help(t *testing.T) {
	var helpArguments = [][]string{{"boom", "--help"}, {"boom", "-h"}, {"boom", "help"}}

	for _, args := range helpArguments {
		buf := new(bytes.Buffer)

		t.Logf("==> checking with arguments %v", args)

		app := &App{
			Name:   "boom",
			Usage:  "make an explosive entrance",
			Writer: buf,
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
		t.Logf("output: %q\n", buf.Bytes())

		if !strings.Contains(output, "boom - make an explosive entrance") {
			t.Errorf("want help to contain %q, did not: \n%q", "boom - make an explosive entrance", output)
		}
	}
}

func TestApp_Run_Version(t *testing.T) {
	var versionArguments = [][]string{{"boom", "--version"}, {"boom", "-v"}}

	for _, args := range versionArguments {
		buf := new(bytes.Buffer)

		t.Logf("==> checking with arguments %v", args)

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
		t.Logf("output: %q\n", buf.Bytes())

		if !strings.Contains(output, "0.1.0") {
			t.Errorf("want version to contain %q, did not: \n%q", "0.1.0", output)
		}
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

	app.Run([]string{"categories"})

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

	if !reflect.DeepEqual(app.Categories, &expect) {
		t.Fatalf("expected categories %#v, to equal %#v", app.Categories, &expect)
	}

	output := buf.String()
	t.Logf("output: %q\n", buf.Bytes())

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

func TestApp_Run_DoesNotOverwriteErrorFromBefore(t *testing.T) {
	app := &App{
		Action: func(c *Context) error { return nil },
		Before: func(c *Context) error { return fmt.Errorf("before error") },
		After:  func(c *Context) error { return fmt.Errorf("after error") },
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
