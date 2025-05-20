package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/mail"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func buildExtendedTestCommand() *Command {
	cmd := buildMinimalTestCommand()
	cmd.Name = "greet"
	cmd.Flags = []Flag{
		&StringFlag{
			Name:      "socket",
			Aliases:   []string{"s"},
			Usage:     "some 'usage' text",
			Value:     "value",
			TakesFile: true,
		},
		&StringFlag{Name: "flag", Aliases: []string{"fl", "f"}},
		&BoolFlag{
			Name:    "another-flag",
			Aliases: []string{"b"},
			Usage:   "another usage text",
			Sources: EnvVars("EXAMPLE_VARIABLE_NAME"),
		},
		&BoolFlag{
			Name:   "hidden-flag",
			Hidden: true,
		},
	}
	cmd.Commands = []*Command{{
		Aliases: []string{"c"},
		Flags: []Flag{
			&StringFlag{
				Name:      "flag",
				Aliases:   []string{"fl", "f"},
				TakesFile: true,
			},
			&BoolFlag{
				Name:    "another-flag",
				Aliases: []string{"b"},
				Usage:   "another usage text",
			},
		},
		Name:  "config",
		Usage: "another usage test",
		Commands: []*Command{{
			Aliases: []string{"s", "ss"},
			Flags: []Flag{
				&StringFlag{Name: "sub-flag", Aliases: []string{"sub-fl", "s"}},
				&BoolFlag{
					Name:    "sub-command-flag",
					Aliases: []string{"s"},
					Usage:   "some usage text",
				},
			},
			Name:  "sub-config",
			Usage: "another usage test",
		}},
	}, {
		Aliases: []string{"i", "in"},
		Name:    "info",
		Usage:   "retrieve generic information",
	}, {
		Name: "some-command",
	}, {
		Name:   "hidden-command",
		Hidden: true,
		Flags: []Flag{
			&BoolFlag{
				Name: "completable",
			},
		},
	}, {
		Aliases: []string{"u"},
		Flags: []Flag{
			&StringFlag{
				Name:      "flag",
				Aliases:   []string{"fl", "f"},
				TakesFile: true,
			},
			&BoolFlag{
				Name:    "another-flag",
				Aliases: []string{"b"},
				Usage:   "another usage text",
			},
		},
		Name:  "usage",
		Usage: "standard usage text",
		UsageText: `
Usage for the usage text
- formatted:  Based on the specified ConfigMap and summon secrets.yml
- list:       Inspect the environment for a specific process running on a Pod
- for_effect: Compare 'namespace' environment with 'local'

` + "```" + `
func() { ... }
` + "```" + `

Should be a part of the same code block
`,
		Commands: []*Command{{
			Aliases: []string{"su"},
			Flags: []Flag{
				&BoolFlag{
					Name:    "sub-command-flag",
					Aliases: []string{"s"},
					Usage:   "some usage text",
				},
			},
			Name:      "sub-usage",
			Usage:     "standard usage text",
			UsageText: "Single line of UsageText",
		}},
	}}
	cmd.UsageText = "app [first_arg] [second_arg]"
	cmd.Description = `Description of the application.`
	cmd.Usage = "Some app"
	cmd.Authors = []any{
		"Harrison <harrison@lolwut.example.com>",
		&mail.Address{Name: "Oliver Allen", Address: "oliver@toyshop.com"},
	}

	return cmd
}

func TestCommandFlagParsing(t *testing.T) {
	cases := []struct {
		testArgs               []string
		skipFlagParsing        bool
		useShortOptionHandling bool
		expectedErr            string
	}{
		// Test normal "not ignoring flags" flow
		{testArgs: []string{"test-cmd", "-break", "blah", "blah"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: "flag provided but not defined: -break"},
		{testArgs: []string{"test-cmd", "blah", "blah"}, skipFlagParsing: true, useShortOptionHandling: false},                                        // Test SkipFlagParsing without any args that look like flags
		{testArgs: []string{"test-cmd", "blah", "-break"}, skipFlagParsing: true, useShortOptionHandling: false},                                      // Test SkipFlagParsing with random flag arg
		{testArgs: []string{"test-cmd", "blah", "-help"}, skipFlagParsing: true, useShortOptionHandling: false},                                       // Test SkipFlagParsing with "special" help flag arg
		{testArgs: []string{"test-cmd", "blah", "-h"}, skipFlagParsing: false, useShortOptionHandling: true, expectedErr: "No help topic for 'blah'"}, // Test UseShortOptionHandling
	}

	for _, c := range cases {
		t.Run(strings.Join(c.testArgs, " "), func(t *testing.T) {
			cmd := &Command{
				Writer:          io.Discard,
				Name:            "test-cmd",
				Aliases:         []string{"tc"},
				Usage:           "this is for testing",
				Description:     "testing",
				Action:          func(context.Context, *Command) error { return nil },
				SkipFlagParsing: c.skipFlagParsing,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			r := require.New(t)

			err := cmd.Run(ctx, c.testArgs)

			if c.expectedErr != "" {
				r.EqualError(err, c.expectedErr)
			} else {
				r.NoError(err)
			}
		})
	}
}

func TestParseAndRunShortOpts(t *testing.T) {
	testCases := []struct {
		testArgs     *stringSliceArgs
		expectedErr  string
		expectedArgs Args
	}{
		{testArgs: &stringSliceArgs{v: []string{"test", "-a"}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-c", "arg1", "arg2"}}, expectedArgs: &stringSliceArgs{v: []string{"arg1", "arg2"}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-f"}}, expectedArgs: &stringSliceArgs{v: []string{}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-ac", "--fgh"}}, expectedArgs: &stringSliceArgs{v: []string{}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-af"}}, expectedArgs: &stringSliceArgs{v: []string{}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-cf"}}, expectedArgs: &stringSliceArgs{v: []string{}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-acf"}}, expectedArgs: &stringSliceArgs{v: []string{}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "--acf"}}, expectedErr: "flag provided but not defined: -acf"},
		{testArgs: &stringSliceArgs{v: []string{"test", "-invalid"}}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: &stringSliceArgs{v: []string{"test", "-acf", "-invalid"}}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: &stringSliceArgs{v: []string{"test", "--invalid"}}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: &stringSliceArgs{v: []string{"test", "-acf", "--invalid"}}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: &stringSliceArgs{v: []string{"test", "-acf", "arg1", "-invalid"}}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: &stringSliceArgs{v: []string{"test", "-acf", "arg1", "--invalid"}}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: &stringSliceArgs{v: []string{"test", "-acfi", "not-arg", "arg1", "-invalid"}}, expectedErr: "flag provided but not defined: -invalid"},
		{testArgs: &stringSliceArgs{v: []string{"test", "-i", "ivalue"}}, expectedArgs: &stringSliceArgs{v: []string{}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-i", "ivalue", "arg1"}}, expectedArgs: &stringSliceArgs{v: []string{"arg1"}}},
		{testArgs: &stringSliceArgs{v: []string{"test", "-i"}}, expectedErr: "flag needs an argument: -i"},
	}

	for _, tc := range testCases {
		t.Run(strings.Join(tc.testArgs.v, " "), func(t *testing.T) {
			state := map[string]Args{"args": nil}

			cmd := &Command{
				Name:        "test",
				Usage:       "this is for testing",
				Description: "testing",
				Action: func(_ context.Context, cmd *Command) error {
					state["args"] = cmd.Args()
					return nil
				},
				UseShortOptionHandling: true,
				Writer:                 io.Discard,
				Flags: []Flag{
					&BoolFlag{Name: "abc", Aliases: []string{"a"}},
					&BoolFlag{Name: "cde", Aliases: []string{"c"}},
					&BoolFlag{Name: "fgh", Aliases: []string{"f"}},
					&StringFlag{Name: "ijk", Aliases: []string{"i"}},
				},
			}

			err := cmd.Run(buildTestContext(t), tc.testArgs.Slice())

			r := require.New(t)

			if tc.expectedErr == "" {
				r.NoError(err)
			} else {
				r.ErrorContains(err, tc.expectedErr)
			}

			if tc.expectedArgs == nil {
				if state["args"] != nil {
					r.Len(state["args"].Slice(), 0)
				} else {
					r.Nil(state["args"])
				}
			} else {
				r.Equal(tc.expectedArgs, state["args"])
			}
		})
	}
}

func TestCommand_Run_DoesNotOverwriteErrorFromBefore(t *testing.T) {
	cmd := &Command{
		Name: "bar",
		Before: func(context.Context, *Command) (context.Context, error) {
			return nil, fmt.Errorf("before error")
		},
		After: func(context.Context, *Command) error {
			return fmt.Errorf("after error")
		},
		Writer: io.Discard,
	}

	err := cmd.Run(buildTestContext(t), []string{"bar"})

	require.ErrorContains(t, err, "before error")
	require.ErrorContains(t, err, "after error")
}

func TestCommand_Run_BeforeSavesMetadata(t *testing.T) {
	var receivedMsgFromAction string
	var receivedMsgFromAfter string

	cmd := &Command{
		Name: "bar",
		Before: func(ctx context.Context, cmd *Command) (context.Context, error) {
			cmd.Metadata["msg"] = "hello world"
			return nil, nil
		},
		Action: func(ctx context.Context, cmd *Command) error {
			msg, ok := cmd.Metadata["msg"]
			if !ok {
				return errors.New("msg not found")
			}
			receivedMsgFromAction = msg.(string)

			return nil
		},
		After: func(_ context.Context, cmd *Command) error {
			msg, ok := cmd.Metadata["msg"]
			if !ok {
				return errors.New("msg not found")
			}
			receivedMsgFromAfter = msg.(string)
			return nil
		},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "bar"}))
	require.Equal(t, "hello world", receivedMsgFromAction)
	require.Equal(t, "hello world", receivedMsgFromAfter)
}

func TestCommand_Run_BeforeReturnNewContext(t *testing.T) {
	var receivedValFromAction, receivedValFromAfter string
	type key string

	bkey := key("bkey")

	cmd := &Command{
		Name: "bar",
		Before: func(ctx context.Context, cmd *Command) (context.Context, error) {
			return context.WithValue(ctx, bkey, "bval"), nil
		},
		Action: func(ctx context.Context, cmd *Command) error {
			if val := ctx.Value(bkey); val == nil {
				return errors.New("bkey value not found")
			} else {
				receivedValFromAction = val.(string)
			}
			return nil
		},
		After: func(ctx context.Context, cmd *Command) error {
			if val := ctx.Value(bkey); val == nil {
				return errors.New("bkey value not found")
			} else {
				receivedValFromAfter = val.(string)
			}
			return nil
		},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "bar"}))
	require.Equal(t, "bval", receivedValFromAfter)
	require.Equal(t, "bval", receivedValFromAction)
}

type ctxKey string

// ctxCollector is a small helper to collect context values.
type ctxCollector struct {
	// keys are the keys to check the context for.
	keys []ctxKey

	// m maps from function name to context name to value.
	m map[string]map[ctxKey]string
}

func (cc *ctxCollector) collect(ctx context.Context, fnName string) {
	if cc.m == nil {
		cc.m = make(map[string]map[ctxKey]string)
	}

	if _, ok := cc.m[fnName]; !ok {
		cc.m[fnName] = make(map[ctxKey]string)
	}

	for _, k := range cc.keys {
		if val := ctx.Value(k); val != nil {
			cc.m[fnName][k] = val.(string)
		}
	}
}

func TestCommand_Run_BeforeReturnNewContextSubcommand(t *testing.T) {
	bkey := ctxKey("bkey")
	bkey2 := ctxKey("bkey2")

	cc := &ctxCollector{keys: []ctxKey{bkey, bkey2}}
	cmd := &Command{
		Name: "bar",
		Before: func(ctx context.Context, cmd *Command) (context.Context, error) {
			return context.WithValue(ctx, bkey, "bval"), nil
		},
		After: func(ctx context.Context, cmd *Command) error {
			cc.collect(ctx, "bar.After")
			return nil
		},
		Commands: []*Command{
			{
				Name: "baz",
				Before: func(ctx context.Context, cmd *Command) (context.Context, error) {
					return context.WithValue(ctx, bkey2, "bval2"), nil
				},
				Action: func(ctx context.Context, cmd *Command) error {
					cc.collect(ctx, "baz.Action")
					return nil
				},
				After: func(ctx context.Context, cmd *Command) error {
					cc.collect(ctx, "baz.After")
					return nil
				},
			},
		},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"bar", "baz"}))
	expected := map[string]map[ctxKey]string{
		"bar.After": {
			bkey:  "bval",
			bkey2: "bval2",
		},
		"baz.Action": {
			bkey:  "bval",
			bkey2: "bval2",
		},
		"baz.After": {
			bkey:  "bval",
			bkey2: "bval2",
		},
	}
	require.Equal(t, expected, cc.m)
}

func TestCommand_Run_FlagActionContext(t *testing.T) {
	bkey := ctxKey("bkey")
	bkey2 := ctxKey("bkey2")

	cc := &ctxCollector{keys: []ctxKey{bkey, bkey2}}
	cmd := &Command{
		Name: "bar",
		Before: func(ctx context.Context, cmd *Command) (context.Context, error) {
			return context.WithValue(ctx, bkey, "bval"), nil
		},
		Flags: []Flag{
			&StringFlag{
				Name: "foo",
				Action: func(ctx context.Context, cmd *Command, _ string) error {
					cc.collect(ctx, "bar.foo.Action")
					return nil
				},
			},
		},
		Commands: []*Command{
			{
				Name: "baz",
				Before: func(ctx context.Context, cmd *Command) (context.Context, error) {
					return context.WithValue(ctx, bkey2, "bval2"), nil
				},
				Flags: []Flag{
					&StringFlag{
						Name: "goo",
						Action: func(ctx context.Context, cmd *Command, _ string) error {
							cc.collect(ctx, "baz.goo.Action")
							return nil
						},
					},
				},
				Action: func(ctx context.Context, cmd *Command) error {
					return nil
				},
			},
		},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"bar", "--foo", "value", "baz", "--goo", "value"}))
	expected := map[string]map[ctxKey]string{
		"bar.foo.Action": {
			bkey:  "bval",
			bkey2: "bval2",
		},
		"baz.goo.Action": {
			bkey:  "bval",
			bkey2: "bval2",
		},
	}
	require.Equal(t, expected, cc.m)
}

func TestCommand_OnUsageError_hasCommandContext(t *testing.T) {
	cmd := &Command{
		Name: "bar",
		Flags: []Flag{
			&Int64Flag{Name: "flag"},
		},
		OnUsageError: func(_ context.Context, cmd *Command, err error, _ bool) error {
			return fmt.Errorf("intercepted in %s: %s", cmd.Name, err.Error())
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"bar", "--flag=wrong"})
	assert.ErrorContains(t, err, "intercepted in bar")
}

func TestCommand_OnUsageError_WithWrongFlagValue(t *testing.T) {
	cmd := &Command{
		Name: "bar",
		Flags: []Flag{
			&Int64Flag{Name: "flag"},
		},
		OnUsageError: func(_ context.Context, _ *Command, err error, _ bool) error {
			assert.ErrorContains(t, err, "strconv.ParseInt: parsing \"wrong\"")
			return errors.New("intercepted: " + err.Error())
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"bar", "--flag=wrong"})
	assert.ErrorContains(t, err, "intercepted: invalid value \"wrong\" for flag -flag: strconv.ParseInt: parsing \"wrong\"")
}

func TestCommand_OnUsageError_WithSubcommand(t *testing.T) {
	cmd := &Command{
		Name: "bar",
		Commands: []*Command{
			{
				Name: "baz",
			},
		},
		Flags: []Flag{
			&Int64Flag{Name: "flag"},
		},
		OnUsageError: func(_ context.Context, _ *Command, err error, _ bool) error {
			assert.ErrorContains(t, err, "parsing \"wrong\": invalid syntax")
			return errors.New("intercepted: " + err.Error())
		},
	}

	require.ErrorContains(t, cmd.Run(buildTestContext(t), []string{"bar", "--flag=wrong"}),
		"intercepted: invalid value \"wrong\" for flag -flag: strconv.ParseInt: parsing \"wrong\": invalid syntax")
}

func TestCommand_Run_SubcommandsCanUseErrWriter(t *testing.T) {
	cmd := &Command{
		ErrWriter: io.Discard,
		Name:      "bar",
		Usage:     "this is for testing",
		Commands: []*Command{
			{
				Name:  "baz",
				Usage: "this is for testing",
				Action: func(_ context.Context, cmd *Command) error {
					require.Equal(t, io.Discard, cmd.Root().ErrWriter)

					return nil
				},
			},
		},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"bar", "baz"}))
}

func TestCommandSkipFlagParsing(t *testing.T) {
	cases := []struct {
		testArgs     *stringSliceArgs
		expectedArgs *stringSliceArgs
		expectedErr  error
	}{
		{testArgs: &stringSliceArgs{v: []string{"some-command", "some-arg", "--flag", "foo"}}, expectedArgs: &stringSliceArgs{v: []string{"some-arg", "--flag", "foo"}}, expectedErr: nil},
		{testArgs: &stringSliceArgs{v: []string{"some-command", "some-arg", "--flag=foo"}}, expectedArgs: &stringSliceArgs{v: []string{"some-arg", "--flag=foo"}}, expectedErr: nil},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.testArgs.Slice(), " "), func(t *testing.T) {
			var args Args
			cmd := &Command{
				SkipFlagParsing: true,
				Name:            "some-command",
				Flags: []Flag{
					&StringFlag{Name: "flag"},
				},
				Action: func(_ context.Context, cmd *Command) error {
					args = cmd.Args()
					return nil
				},
				Writer: io.Discard,
			}

			err := cmd.Run(buildTestContext(t), c.testArgs.Slice())
			assert.Equal(t, c.expectedErr, err)
			assert.Equal(t, c.expectedArgs, args)
		})
	}
}

func TestCommand_Run_CustomShellCompleteAcceptsMalformedFlags(t *testing.T) {
	cases := []struct {
		testArgs    *stringSliceArgs
		expectedOut string
	}{
		{testArgs: &stringSliceArgs{v: []string{"--undefined"}}, expectedOut: "found 0 args"},
		{testArgs: &stringSliceArgs{v: []string{"--number"}}, expectedOut: "found 0 args"},
		{testArgs: &stringSliceArgs{v: []string{"--number", "forty-two"}}, expectedOut: "found 0 args"},
		{testArgs: &stringSliceArgs{v: []string{"--number", "42"}}, expectedOut: "found 0 args"},
		{testArgs: &stringSliceArgs{v: []string{"--number", "42", "newArg"}}, expectedOut: "found 1 args"},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.testArgs.Slice(), " "), func(t *testing.T) {
			out := &bytes.Buffer{}
			cmd := &Command{
				Writer:                out,
				EnableShellCompletion: true,
				Name:                  "bar",
				Usage:                 "this is for testing",
				Flags: []Flag{
					&Int64Flag{
						Name:  "number",
						Usage: "A number to parse",
					},
				},
				ShellComplete: func(_ context.Context, cmd *Command) {
					fmt.Fprintf(cmd.Root().Writer, "found %[1]d args", cmd.NArg())
				},
			}

			osArgs := &stringSliceArgs{v: []string{"bar"}}
			osArgs.v = append(osArgs.v, c.testArgs.Slice()...)
			osArgs.v = append(osArgs.v, completionFlag)

			r := require.New(t)

			r.NoError(cmd.Run(buildTestContext(t), osArgs.Slice()))
			r.Equal(c.expectedOut, out.String())
		})
	}
}

func TestCommand_CanAddVFlagOnSubCommands(t *testing.T) {
	cmd := &Command{
		Version: "some version",
		Writer:  io.Discard,
		Name:    "foo",
		Usage:   "this is for testing",
		Commands: []*Command{
			{
				Name: "bar",
				Flags: []Flag{
					&BoolFlag{Name: "v"},
				},
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "bar"})
	assert.NoError(t, err)
}

func TestCommand_VisibleSubcCommands(t *testing.T) {
	subc1 := &Command{
		Name:  "subc1",
		Usage: "subc1 command1",
	}
	subc3 := &Command{
		Name:  "subc3",
		Usage: "subc3 command2",
	}
	cmd := &Command{
		Name:  "bar",
		Usage: "this is for testing",
		Commands: []*Command{
			subc1,
			{
				Name:   "subc2",
				Usage:  "subc2 command2",
				Hidden: true,
			},
			subc3,
		},
	}

	assert.Equal(t, cmd.VisibleCommands(), []*Command{subc1, subc3})
}

func TestCommand_VisibleFlagCategories(t *testing.T) {
	cmd := &Command{
		Name:  "bar",
		Usage: "this is for testing",
		Flags: []Flag{
			&StringFlag{
				Name: "strd", // no category set
			},
			&StringFlag{
				Name:   "strd1", // no category set and also hidden
				Hidden: true,
			},
			&Int64Flag{
				Name:     "intd",
				Aliases:  []string{"altd1", "altd2"},
				Category: "cat1",
			},
			&StringFlag{
				Name:     "sfd",
				Category: "cat2", // category set and hidden
				Hidden:   true,
			},
		},
		MutuallyExclusiveFlags: []MutuallyExclusiveFlags{{
			Category: "cat2",
			Flags: [][]Flag{
				{
					&StringFlag{
						Name: "mutex",
					},
				},
			},
		}},
	}

	cmd.MutuallyExclusiveFlags[0].propagateCategory()

	vfc := cmd.VisibleFlagCategories()
	require.Len(t, vfc, 3)

	assert.Equal(t, vfc[0].Name(), "", "expected category name to be empty")
	assert.Equal(t, vfc[0].Flags()[0].Names(), []string{"strd"})

	assert.Equal(t, vfc[1].Name(), "cat1", "expected category name cat1")
	require.Len(t, vfc[1].Flags(), 1, "expected flag category to have one flag")
	assert.Equal(t, vfc[1].Flags()[0].Names(), []string{"intd", "altd1", "altd2"})

	assert.Equal(t, vfc[2].Name(), "cat2", "expected category name cat2")
	require.Len(t, vfc[2].Flags(), 1, "expected flag category to have one flag")
	assert.Equal(t, vfc[2].Flags()[0].Names(), []string{"mutex"})
}

func TestCommand_RunSubcommandWithDefault(t *testing.T) {
	cmd := &Command{
		Version:        "some version",
		Name:           "app",
		DefaultCommand: "foo",
		Commands: []*Command{
			{
				Name: "foo",
				Action: func(context.Context, *Command) error {
					return errors.New("should not run this subcommand")
				},
			},
			{
				Name:     "bar",
				Usage:    "this is for testing",
				Commands: []*Command{{}}, // some subcommand
				Action: func(context.Context, *Command) error {
					return nil
				},
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"app", "bar"})
	assert.NoError(t, err)

	err = cmd.Run(buildTestContext(t), []string{"app"})
	assert.EqualError(t, err, "should not run this subcommand")
}

func TestCommand_Run(t *testing.T) {
	s := ""

	cmd := &Command{
		Action: func(_ context.Context, cmd *Command) error {
			s = s + cmd.Args().First()
			return nil
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"command", "foo"})
	assert.NoError(t, err)
	err = cmd.Run(buildTestContext(t), []string{"command", "bar"})
	assert.NoError(t, err)
	assert.Equal(t, s, "foobar")
}

var commandTests = []struct {
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

func TestCommand_Command(t *testing.T) {
	cmd := &Command{
		Commands: []*Command{
			{Name: "foobar", Aliases: []string{"f"}},
			{Name: "batbaz", Aliases: []string{"b"}},
		},
	}

	for _, test := range commandTests {
		if test.expected {
			assert.NotEmpty(t, cmd.Command(test.name))
		} else {
			assert.Empty(t, cmd.Command(test.name))
		}
	}
}

var defaultCommandTests = []struct {
	cmdName        string
	defaultCmd     string
	errNotExpected bool
}{
	{"foobar", "foobar", true},
	{"batbaz", "foobar", true},
	{"b", "", true},
	{"f", "", true},
	{"", "foobar", true},
	// TBD
	//{"", "", true},
	//{" ", "", false},
	{"bat", "batbaz", true},
	{"nothing", "batbaz", true},
	{"nothing", "", false},
}

func TestCommand_RunDefaultCommand(t *testing.T) {
	for _, test := range defaultCommandTests {
		testTitle := fmt.Sprintf("command=%[1]s-default=%[2]s", test.cmdName, test.defaultCmd)
		t.Run(testTitle, func(t *testing.T) {
			cmd := &Command{
				DefaultCommand: test.defaultCmd,
				Commands: []*Command{
					{Name: "foobar", Aliases: []string{"f"}},
					{Name: "batbaz", Aliases: []string{"b"}},
				},
			}

			err := cmd.Run(buildTestContext(t), []string{"c", test.cmdName})
			if test.errNotExpected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

var defaultCommandSubCommandTests = []struct {
	cmdName        string
	subCmd         string
	defaultCmd     string
	errNotExpected bool
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
	{" ", "jimmers", "foobar", true},
	/*{"", "", "", true},
	{" ", "", "", false},
	{" ", "j", "", false},*/
	{"bat", "", "batbaz", true},
	{"nothing", "", "batbaz", true},
	{"nothing", "", "", false},
	{"nothing", "j", "batbaz", false},
	{"nothing", "carly", "", false},
}

func TestCommand_RunDefaultCommandWithSubCommand(t *testing.T) {
	for _, test := range defaultCommandSubCommandTests {
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

			err := cmd.Run(buildTestContext(t), []string{"c", test.cmdName, test.subCmd})
			if test.errNotExpected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

var defaultCommandFlagTests = []struct {
	cmdName        string
	flag           string
	defaultCmd     string
	errNotExpected bool
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
	{" ", "-j", "foobar", true},
	{"", "", "", true},
	{" ", "", "", true},
	{" ", "-j", "", true},
	{"bat", "", "batbaz", true},
	{"nothing", "", "batbaz", true},
	{"nothing", "", "", false},
	{"nothing", "--jimbob", "batbaz", true},
	{"nothing", "--carly", "", false},
}

func TestCommand_RunDefaultCommandWithFlags(t *testing.T) {
	for _, test := range defaultCommandFlagTests {
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

			err := cmd.Run(buildTestContext(t), appArgs)
			if test.errNotExpected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestCommand_FlagsFromExtPackage(t *testing.T) {
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

	err := cmd.Run(buildTestContext(t), []string{"foo", "-c", "cly", "--epflag", "10"})
	assert.NoError(t, err)

	assert.Equal(t, int(10), someint)
	// this exercises the extFlag.Get()
	assert.Equal(t, int(10), cmd.Value("epflag"))

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
	err = cmd.Run(buildTestContext(t), []string{"foo", "-c", "cly", "--epflag", "10"})
	assert.Error(t, err)
}

func TestCommand_Setup_defaultsReader(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults([]string{"test"})
	assert.Equal(t, cmd.Reader, os.Stdin)
}

func TestCommand_Setup_defaultsWriter(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults([]string{"test"})
	assert.Equal(t, cmd.Writer, os.Stdout)
}

func TestCommand_CommandWithFlagBeforeTerminator(t *testing.T) {
	var parsedOption string
	var args Args

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "cmd",
				Flags: []Flag{
					&StringFlag{Name: "option", Value: "", Usage: "some option"},
				},
				Action: func(_ context.Context, cmd *Command) error {
					parsedOption = cmd.String("option")
					args = cmd.Args()
					return nil
				},
			},
		},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"", "cmd", "--option", "my-option", "my-arg", "--", "--notARealFlag"}))

	require.Equal(t, "my-option", parsedOption)
	require.Equal(t, "my-arg", args.Get(0))
	require.Equal(t, "--notARealFlag", args.Get(1))
}

func TestCommand_CommandWithDash(t *testing.T) {
	var args Args

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "cmd",
				Action: func(_ context.Context, cmd *Command) error {
					args = cmd.Args()
					return nil
				},
			},
		},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"", "cmd", "my-arg", "-"}))
	require.NotNil(t, args)
	require.Equal(t, "my-arg", args.Get(0))
	require.Equal(t, "-", args.Get(1))
}

func TestCommand_CommandWithNoFlagBeforeTerminator(t *testing.T) {
	var args Args

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "cmd",
				Action: func(_ context.Context, cmd *Command) error {
					args = cmd.Args()
					return nil
				},
			},
		},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"", "cmd", "my-arg", "--", "notAFlagAtAll"}))

	require.NotNil(t, args)
	require.Equal(t, "my-arg", args.Get(0))
	require.Equal(t, "notAFlagAtAll", args.Get(1))
}

func TestCommand_SkipFlagParsing(t *testing.T) {
	var args Args

	cmd := &Command{
		SkipFlagParsing: true,
		Action: func(_ context.Context, cmd *Command) error {
			args = cmd.Args()
			return nil
		},
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "--", "my-arg", "notAFlagAtAll"})

	assert.NotNil(t, args)
	assert.Equal(t, "--", args.Get(0))
	assert.Equal(t, "my-arg", args.Get(1))
	assert.Equal(t, "notAFlagAtAll", args.Get(2))
}

func TestCommand_VisibleCommands(t *testing.T) {
	cmd := &Command{
		Commands: []*Command{
			{
				Name:   "frob",
				Action: func(context.Context, *Command) error { return nil },
			},
			{
				Name:   "frib",
				Hidden: true,
				Action: func(context.Context, *Command) error { return nil },
			},
		},
	}

	cmd.setupDefaults([]string{"test"})
	expected := []*Command{
		cmd.Commands[0],
	}
	actual := cmd.VisibleCommands()
	assert.Len(t, actual, len(expected))
	for i, actualCommand := range actual {
		expectedCommand := expected[i]

		if expectedCommand.Action != nil {
			// comparing func addresses is OK!
			assert.Equal(t, fmt.Sprintf("%p", expectedCommand.Action), fmt.Sprintf("%p", actualCommand.Action))
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

			assert.Equal(t, expectedCommand, actualCommand)
		}()
	}
}

func TestCommand_UseShortOptionHandling(t *testing.T) {
	var one, two bool
	var name string
	expected := "expectedName"

	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	cmd.Flags = []Flag{
		&BoolFlag{Name: "one", Aliases: []string{"o"}},
		&BoolFlag{Name: "two", Aliases: []string{"t"}},
		&StringFlag{Name: "name", Aliases: []string{"n"}},
	}
	cmd.Action = func(_ context.Context, cmd *Command) error {
		one = cmd.Bool("one")
		two = cmd.Bool("two")
		name = cmd.String("name")
		return nil
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "-on", expected})
	assert.True(t, one)
	assert.False(t, two)
	assert.Equal(t, name, expected)
}

func TestCommand_UseShortOptionHandling_missing_value(t *testing.T) {
	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	cmd.Flags = []Flag{
		&StringFlag{Name: "name", Aliases: []string{"n"}},
	}

	err := cmd.Run(buildTestContext(t), []string{"", "-n"})
	assert.EqualError(t, err, "flag needs an argument: -n")
}

func TestCommand_UseShortOptionHandlingCommand(t *testing.T) {
	var (
		one, two bool
		name     string
		expected = "expectedName"
	)

	cmd := &Command{
		Name: "cmd",
		Flags: []Flag{
			&BoolFlag{Name: "one", Aliases: []string{"o"}},
			&BoolFlag{Name: "two", Aliases: []string{"t"}},
			&StringFlag{Name: "name", Aliases: []string{"n"}},
		},
		Action: func(_ context.Context, cmd *Command) error {
			one = cmd.Bool("one")
			two = cmd.Bool("two")
			name = cmd.String("name")
			return nil
		},
		UseShortOptionHandling: true,
		Writer:                 io.Discard,
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"cmd", "-on", expected}))
	require.True(t, one)
	require.False(t, two)
	require.Equal(t, expected, name)
}

func TestCommand_UseShortOptionHandlingCommand_missing_value(t *testing.T) {
	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	command := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{Name: "name", Aliases: []string{"n"}},
		},
	}
	cmd.Commands = []*Command{command}

	require.EqualError(
		t,
		cmd.Run(buildTestContext(t), []string{"", "cmd", "-n"}),
		"flag needs an argument: -n",
	)
}

func TestCommand_UseShortOptionHandlingSubCommand(t *testing.T) {
	var one, two bool
	var name string

	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	cmd.Commands = []*Command{
		{
			Name: "cmd",
			Commands: []*Command{
				{
					Name: "sub",
					Flags: []Flag{
						&BoolFlag{Name: "one", Aliases: []string{"o"}},
						&BoolFlag{Name: "two", Aliases: []string{"t"}},
						&StringFlag{Name: "name", Aliases: []string{"n"}},
					},
					Action: func(_ context.Context, cmd *Command) error {
						one = cmd.Bool("one")
						two = cmd.Bool("two")
						name = cmd.String("name")
						return nil
					},
				},
			},
		},
	}

	expected := "expectedName"

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"", "cmd", "sub", "-on", expected}))
	require.True(t, one)
	require.False(t, two)
	require.Equal(t, expected, name)
}

func TestCommand_UseShortOptionHandlingSubCommand_missing_value(t *testing.T) {
	cmd := buildMinimalTestCommand()
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

	err := cmd.Run(buildTestContext(t), []string{"", "cmd", "sub", "-n"})
	assert.EqualError(t, err, "flag needs an argument: -n")
}

func TestCommand_UseShortOptionAfterSliceFlag(t *testing.T) {
	var one, two bool
	var name string
	var sliceValDest []string
	var sliceVal []string
	expected := "expectedName"

	cmd := buildMinimalTestCommand()
	cmd.UseShortOptionHandling = true
	cmd.Flags = []Flag{
		&StringSliceFlag{Name: "env", Aliases: []string{"e"}, Destination: &sliceValDest},
		&BoolFlag{Name: "one", Aliases: []string{"o"}},
		&BoolFlag{Name: "two", Aliases: []string{"t"}},
		&StringFlag{Name: "name", Aliases: []string{"n"}},
	}
	cmd.Action = func(_ context.Context, cmd *Command) error {
		sliceVal = cmd.StringSlice("env")
		one = cmd.Bool("one")
		two = cmd.Bool("two")
		name = cmd.String("name")
		return nil
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "-e", "foo", "-on", expected})
	assert.Equal(t, sliceVal, []string{"foo"})
	assert.Equal(t, sliceValDest, []string{"foo"})
	assert.True(t, one)
	assert.False(t, two)
	assert.Equal(t, expected, name)
}

func TestCommand_Float64Flag(t *testing.T) {
	var meters float64

	cmd := &Command{
		Flags: []Flag{
			&FloatFlag{Name: "height", Value: 1.5, Usage: "Set the height, in meters"},
		},
		Action: func(_ context.Context, cmd *Command) error {
			meters = cmd.Float("height")
			return nil
		},
	}

	_ = cmd.Run(buildTestContext(t), []string{"", "--height", "1.93"})
	assert.Equal(t, 1.93, meters)
}

func TestCommand_ParseSliceFlags(t *testing.T) {
	var parsedIntSlice []int64
	var parsedStringSlice []string

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "cmd",
				Flags: []Flag{
					&Int64SliceFlag{Name: "p", Value: []int64{}, Usage: "set one or more ip addr"},
					&StringSliceFlag{Name: "ip", Value: []string{}, Usage: "set one or more ports to open"},
				},
				Action: func(_ context.Context, cmd *Command) error {
					parsedIntSlice = cmd.Int64Slice("p")
					parsedStringSlice = cmd.StringSlice("ip")
					return nil
				},
			},
		},
	}

	r := require.New(t)

	r.NoError(cmd.Run(buildTestContext(t), []string{"", "cmd", "-p", "22", "-p", "80", "-ip", "8.8.8.8", "-ip", "8.8.4.4"}))
	r.Equal([]int64{22, 80}, parsedIntSlice)
	r.Equal([]string{"8.8.8.8", "8.8.4.4"}, parsedStringSlice)
}

func TestCommand_ParseSliceFlagsWithMissingValue(t *testing.T) {
	var parsedIntSlice []int64
	var parsedStringSlice []string

	cmd := &Command{
		Commands: []*Command{
			{
				Name: "cmd",
				Flags: []Flag{
					&Int64SliceFlag{Name: "a", Usage: "set numbers"},
					&StringSliceFlag{Name: "str", Usage: "set strings"},
				},
				Action: func(_ context.Context, cmd *Command) error {
					parsedIntSlice = cmd.Int64Slice("a")
					parsedStringSlice = cmd.StringSlice("str")
					return nil
				},
			},
		},
	}

	r := require.New(t)

	r.NoError(cmd.Run(buildTestContext(t), []string{"", "cmd", "-a", "2", "-str", "A"}))
	r.Equal([]int64{2}, parsedIntSlice)
	r.Equal([]string{"A"}, parsedStringSlice)
}

func TestCommand_DefaultStdin(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults([]string{"test"})

	assert.Equal(t, cmd.Reader, os.Stdin, "Default input reader not set.")
}

func TestCommand_DefaultStdout(t *testing.T) {
	cmd := &Command{}
	cmd.setupDefaults([]string{"test"})

	assert.Equal(t, cmd.Writer, os.Stdout, "Default output writer not set.")
}

func TestCommand_SetStdin(t *testing.T) {
	buf := make([]byte, 12)

	cmd := &Command{
		Name:   "test",
		Reader: strings.NewReader("Hello World!"),
		Action: func(_ context.Context, cmd *Command) error {
			_, err := cmd.Reader.Read(buf)
			return err
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"help"})
	require.NoError(t, err)
	assert.Equal(t, "Hello World!", string(buf), "Command did not read input from desired reader.")
}

func TestCommand_SetStdin_Subcommand(t *testing.T) {
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
						Action: func(_ context.Context, cmd *Command) error {
							_, err := cmd.Root().Reader.Read(buf)
							return err
						},
					},
				},
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"test", "command", "subcommand"})
	require.NoError(t, err)
	assert.Equal(t, "Hello World!", string(buf), "Command did not read input from desired reader.")
}

func TestCommand_SetStdout(t *testing.T) {
	var w bytes.Buffer

	cmd := &Command{
		Name:   "test",
		Writer: &w,
	}

	err := cmd.Run(buildTestContext(t), []string{"help"})
	require.NoError(t, err)
	assert.NotZero(t, w.Len(), "Command did not write output to desired writer.")
}

func TestCommand_BeforeFunc(t *testing.T) {
	counts := &opCounts{}
	beforeError := fmt.Errorf("fail")
	var err error

	cmd := &Command{
		Before: func(_ context.Context, cmd *Command) (context.Context, error) {
			counts.Total++
			counts.Before = counts.Total
			s := cmd.String("opt")
			if s == "fail" {
				return nil, beforeError
			}

			return nil, nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(context.Context, *Command) error {
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

	// run with the Before() func succeeding
	err = cmd.Run(buildTestContext(t), []string{"command", "--opt", "succeed", "sub"})
	require.NoError(t, err)

	assert.Equal(t, 1, counts.Before, "Before() not executed when expected")
	assert.Equal(t, 2, counts.SubCommand, "Subcommand not executed when expected")

	// reset
	counts = &opCounts{}

	// run with the Before() func failing
	err = cmd.Run(buildTestContext(t), []string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	assert.ErrorIs(t, err, beforeError, "Run error expected, but not received")
	assert.Equal(t, 1, counts.Before, "Before() not executed when expected")
	assert.Equal(t, 0, counts.SubCommand, "Subcommand executed when NOT expected")

	// reset
	counts = &opCounts{}

	afterError := errors.New("fail again")
	cmd.After = func(context.Context, *Command) error {
		return afterError
	}

	// run with the Before() func failing, wrapped by After()
	err = cmd.Run(buildTestContext(t), []string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	if _, ok := err.(MultiError); !ok {
		t.Errorf("MultiError expected, but not received")
	}

	assert.Equal(t, 1, counts.Before, "Before() not executed when expected")
	assert.Zero(t, counts.SubCommand, "Subcommand executed when NOT expected")
}

func TestCommand_BeforeFuncPersistentFlag(t *testing.T) {
	counts := &opCounts{}
	beforeError := fmt.Errorf("fail")
	var err error

	cmd := &Command{
		Before: func(_ context.Context, cmd *Command) (context.Context, error) {
			counts.Before++
			s := cmd.String("opt")
			if s != "value" {
				return nil, beforeError
			}
			return nil, nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(context.Context, *Command) error {
					counts.SubCommand++
					return nil
				},
			},
		},
		Flags: []Flag{
			&StringFlag{Name: "opt"},
		},
		Writer: io.Discard,
	}

	// Check that --opt value is available in root command Before hook,
	// even when it was set on the subcommand.
	err = cmd.Run(buildTestContext(t), []string{"command", "sub", "--opt", "value"})
	require.NoError(t, err)
	assert.Equal(t, 1, counts.Before, "Before() not executed when expected")
	assert.Equal(t, 1, counts.SubCommand, "Subcommand not executed when expected")
}

func TestCommand_BeforeAfterFuncShellCompletion(t *testing.T) {
	t.Skip("TODO: is '--generate-shell-completion' (flag) still supported?")

	counts := &opCounts{}

	cmd := &Command{
		EnableShellCompletion: true,
		Before: func(context.Context, *Command) (context.Context, error) {
			counts.Total++
			counts.Before = counts.Total
			return nil, nil
		},
		After: func(context.Context, *Command) error {
			counts.Total++
			counts.After = counts.Total
			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(context.Context, *Command) error {
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

	r := require.New(t)

	// run with the Before() func succeeding
	r.NoError(
		cmd.Run(
			buildTestContext(t),
			[]string{
				"command",
				"--opt", "succeed",
				"sub", completionFlag,
			},
		),
	)

	r.Equalf(0, counts.Before, "Before was run")
	r.Equal(0, counts.After, "After was run")
	r.Equal(0, counts.SubCommand, "SubCommand was run")
}

func TestCommand_AfterFunc(t *testing.T) {
	counts := &opCounts{}
	afterError := fmt.Errorf("fail")
	var err error

	cmd := &Command{
		After: func(_ context.Context, cmd *Command) error {
			counts.Total++
			counts.After = counts.Total
			s := cmd.String("opt")
			if s == "fail" {
				return afterError
			}

			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(context.Context, *Command) error {
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
	err = cmd.Run(buildTestContext(t), []string{"command", "--opt", "succeed", "sub"})
	require.NoError(t, err)
	assert.Equal(t, 2, counts.After, "After() not executed when expected")
	assert.Equal(t, 1, counts.SubCommand, "Subcommand not executed when expected")

	// reset
	counts = &opCounts{}

	// run with the Before() func failing
	err = cmd.Run(buildTestContext(t), []string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	assert.ErrorIs(t, err, afterError, "Run error expected, but not received")
	assert.Equal(t, 2, counts.After, "After() not executed when expected")
	assert.Equal(t, 1, counts.SubCommand, "Subcommand not executed when expected")

	/*
		reset
	*/
	counts = &opCounts{}
	// reset the flags since they are set previously
	cmd.Flags = []Flag{
		&StringFlag{Name: "opt"},
	}

	// run with none args
	err = cmd.Run(buildTestContext(t), []string{"command"})

	// should be the same error produced by the Before func
	require.NoError(t, err)

	assert.Equal(t, 1, counts.After, "After() not executed when expected")
	assert.Equal(t, 0, counts.SubCommand, "Subcommand not executed when expected")
}

func TestCommandNoHelpFlag(t *testing.T) {
	oldFlag := HelpFlag
	defer func() {
		HelpFlag = oldFlag
	}()

	HelpFlag = nil

	cmd := &Command{Writer: io.Discard}

	err := cmd.Run(buildTestContext(t), []string{"test", "-h"})

	assert.ErrorContains(t, err, providedButNotDefinedErrMsg, "expected error about missing help flag")
}

func TestRequiredFlagCommandRunBehavior(t *testing.T) {
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
					Action: func(context.Context, *Command) error {
						return nil
					},
				}},
			}},
		},
	}
	for _, test := range tdata {
		t.Run(test.testCase, func(t *testing.T) {
			// setup
			cmd := buildMinimalTestCommand()
			cmd.Flags = test.appFlags
			cmd.Commands = test.appCommands

			// logic under test
			err := cmd.Run(buildTestContext(t), test.appRunInput)

			// assertions
			if test.expectedAnError {
				assert.Error(t, err)
				if _, ok := err.(requiredFlagsErr); test.expectedAnError && !ok {
					t.Errorf("expected a requiredFlagsErr, but got: %s", err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommandHelpPrinter(t *testing.T) {
	oldPrinter := HelpPrinter
	defer func() {
		HelpPrinter = oldPrinter
	}()

	wasCalled := false
	HelpPrinter = func(io.Writer, string, interface{}) {
		wasCalled = true
	}

	cmd := &Command{}

	_ = cmd.Run(buildTestContext(t), []string{"-h"})

	assert.True(t, wasCalled, "Help printer expected to be called, but was not")
}

func TestCommand_VersionPrinter(t *testing.T) {
	oldPrinter := VersionPrinter
	defer func() {
		VersionPrinter = oldPrinter
	}()

	wasCalled := false
	VersionPrinter = func(*Command) {
		wasCalled = true
	}

	cmd := &Command{}
	ShowVersion(cmd)

	assert.True(t, wasCalled, "Version printer expected to be called, but was not")
}

func TestCommand_CommandNotFound(t *testing.T) {
	counts := &opCounts{}
	cmd := &Command{
		CommandNotFound: func(context.Context, *Command, string) {
			counts.Total++
			counts.CommandNotFound = counts.Total
		},
		Commands: []*Command{
			{
				Name: "bar",
				Action: func(context.Context, *Command) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		},
	}

	_ = cmd.Run(buildTestContext(t), []string{"command", "foo"})

	assert.Equal(t, 1, counts.CommandNotFound)
	assert.Equal(t, 0, counts.SubCommand)
	assert.Equal(t, 1, counts.Total)
}

func TestCommand_OrderOfOperations(t *testing.T) {
	buildCmdCounts := func() (*Command, *opCounts) {
		counts := &opCounts{}

		cmd := &Command{
			EnableShellCompletion: true,
			ShellComplete: func(context.Context, *Command) {
				counts.Total++
				counts.ShellComplete = counts.Total
			},
			OnUsageError: func(context.Context, *Command, error, bool) error {
				counts.Total++
				counts.OnUsageError = counts.Total
				return errors.New("hay OnUsageError")
			},
			Writer: io.Discard,
		}

		beforeNoError := func(context.Context, *Command) (context.Context, error) {
			counts.Total++
			counts.Before = counts.Total
			return nil, nil
		}

		cmd.Before = beforeNoError
		cmd.CommandNotFound = func(context.Context, *Command, string) {
			counts.Total++
			counts.CommandNotFound = counts.Total
		}

		afterNoError := func(context.Context, *Command) error {
			counts.Total++
			counts.After = counts.Total
			return nil
		}

		cmd.After = afterNoError
		cmd.Commands = []*Command{
			{
				Name: "bar",
				Action: func(context.Context, *Command) error {
					counts.Total++
					counts.SubCommand = counts.Total
					return nil
				},
			},
		}

		cmd.Action = func(context.Context, *Command) error {
			counts.Total++
			counts.Action = counts.Total
			return nil
		}

		return cmd, counts
	}

	t.Run("on usage error", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		r := require.New(t)

		_ = cmd.Run(buildTestContext(t), []string{"command", "--nope"})
		r.Equal(1, counts.OnUsageError)
		r.Equal(1, counts.Total)
	})

	t.Run("shell complete", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		r := require.New(t)

		_ = cmd.Run(buildTestContext(t), []string{"command", completionFlag})
		r.Equal(1, counts.ShellComplete)
		r.Equal(1, counts.Total)
	})

	t.Run("nil on usage error", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		cmd.OnUsageError = nil

		_ = cmd.Run(buildTestContext(t), []string{"command", "--nope"})
		require.Equal(t, 0, counts.Total)
	})

	t.Run("before after action hooks", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		r := require.New(t)

		_ = cmd.Run(buildTestContext(t), []string{"command", "foo"})
		r.Equal(0, counts.OnUsageError)
		r.Equal(1, counts.Before)
		r.Equal(0, counts.CommandNotFound)
		r.Equal(2, counts.Action)
		r.Equal(3, counts.After)
		r.Equal(3, counts.Total)
	})

	t.Run("before with error", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		cmd.Before = func(context.Context, *Command) (context.Context, error) {
			counts.Total++
			counts.Before = counts.Total
			return nil, errors.New("hay Before")
		}

		r := require.New(t)

		_ = cmd.Run(buildTestContext(t), []string{"command", "bar"})
		r.Equal(0, counts.OnUsageError)
		r.Equal(1, counts.Before)
		r.Equal(2, counts.After)
		r.Equal(2, counts.Total)
	})

	t.Run("nil after", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		cmd.After = nil
		r := require.New(t)

		_ = cmd.Run(buildTestContext(t), []string{"command", "bar"})
		r.Equal(0, counts.OnUsageError)
		r.Equal(1, counts.Before)
		r.Equal(2, counts.SubCommand)
		r.Equal(2, counts.Total)
	})

	t.Run("after errors", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		cmd.After = func(context.Context, *Command) error {
			counts.Total++
			counts.After = counts.Total
			return errors.New("hay After")
		}

		r := require.New(t)

		err := cmd.Run(buildTestContext(t), []string{"command", "bar"})
		r.Error(err)
		r.Equal(0, counts.OnUsageError)
		r.Equal(1, counts.Before)
		r.Equal(2, counts.SubCommand)
		r.Equal(3, counts.After)
		r.Equal(3, counts.Total)
	})

	t.Run("nil commands", func(t *testing.T) {
		cmd, counts := buildCmdCounts()
		cmd.Commands = nil
		r := require.New(t)

		_ = cmd.Run(buildTestContext(t), []string{"command"})
		r.Equal(0, counts.OnUsageError)
		r.Equal(1, counts.Before)
		r.Equal(2, counts.Action)
		r.Equal(3, counts.After)
		r.Equal(3, counts.Total)
	})
}

func TestCommand_Run_CommandWithSubcommandHasHelpTopic(t *testing.T) {
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
				Action:      func(context.Context, *Command) error { return nil },
				Writer:      buf,
			}

			err := cmd.Run(buildTestContext(t), flagSet)
			assert.NoError(t, err)

			output := buf.String()

			assert.NotContains(t, output, "No help topic for", "expect a help topic, got none")

			for _, shouldContain := range []string{
				cmd.Name, cmd.Description,
				subCmdBar.Name, subCmdBar.Usage,
				subCmdBaz.Name, subCmdBaz.Usage,
			} {
				assert.Contains(t, output, shouldContain, "want help to contain %q, did not: \n%q", shouldContain, output)
			}
		})
	}
}

func TestCommand_Run_SubcommandFullPath(t *testing.T) {
	out := &bytes.Buffer{}

	subCmd := &Command{
		Name:      "bar",
		Usage:     "does bar things",
		ArgsUsage: "[arguments...]",
	}

	cmd := &Command{
		Name:        "foo",
		Description: "foo commands",
		Commands:    []*Command{subCmd},
		Writer:      out,
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "bar", "--help"}))

	outString := out.String()
	require.Contains(t, outString, "foo bar - does bar things")
	require.Contains(t, outString, "foo bar [options] [arguments...]")
}

func TestCommand_Run_Help(t *testing.T) {
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
			wantErr:       fmt.Errorf("flag provided but not defined: -help"),
		},
		{
			helpArguments: []string{"boom", "-h"},
			hideHelp:      true,
			wantErr:       fmt.Errorf("flag provided but not defined: -h"),
		},
		{
			helpArguments: []string{"boom", "help"},
			hideHelp:      true,
			wantContains:  "boom I say!",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("checking with arguments %v%v", tt.helpArguments, tt.hideHelp), func(t *testing.T) {
			buf := new(bytes.Buffer)

			cmd := &Command{
				Name:     "boom",
				Usage:    "make an explosive entrance",
				Writer:   buf,
				HideHelp: tt.hideHelp,
				Action: func(context.Context, *Command) error {
					buf.WriteString("boom I say!")
					return nil
				},
			}

			err := cmd.Run(buildTestContext(t), tt.helpArguments)
			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
			}

			output := buf.String()

			assert.Contains(t, output, tt.wantContains, "want help to contain %q, did not: \n%q", "boom - make an explosive entrance", output)
		})
	}
}

func TestCommand_Run_Version(t *testing.T) {
	versionArguments := [][]string{{"boom", "--version"}, {"boom", "-v"}}

	for _, args := range versionArguments {
		t.Run(fmt.Sprintf("checking with arguments %v", args), func(t *testing.T) {
			buf := new(bytes.Buffer)

			cmd := &Command{
				Name:    "boom",
				Usage:   "make an explosive entrance",
				Version: "0.1.0",
				Writer:  buf,
				Action: func(context.Context, *Command) error {
					buf.WriteString("boom I say!")
					return nil
				},
			}

			err := cmd.Run(buildTestContext(t), args)
			assert.NoError(t, err)
			assert.Contains(t, buf.String(), "0.1.0", "want version to contain 0.1.0")
		})
	}
}

func TestCommand_Run_Categories(t *testing.T) {
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

	_ = cmd.Run(buildTestContext(t), []string{"categories"})

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

	require.Equal(t, &expect, cmd.categories)

	output := buf.String()

	assert.Contains(t, output, "1:\n     command1", "want buffer to include category %q, did not: \n%q", "1:\n     command1", output)
}

func TestCommand_VisibleCategories(t *testing.T) {
	cmd := &Command{
		Name:     "visible-categories",
		HideHelp: true,
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
				Hidden:   true,
			},
			{
				Name:     "command2",
				Category: "2",
			},
			{
				Name:     "command3",
				Category: "3",
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

	cmd.setupDefaults([]string{"test"})
	assert.Equal(t, expected, cmd.VisibleCategories())

	cmd = &Command{
		Name:     "visible-categories",
		HideHelp: true,
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
				Hidden:   true,
			},
			{
				Name:     "command2",
				Category: "2",
				Hidden:   true,
			},
			{
				Name:     "command3",
				Category: "3",
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

	cmd.setupDefaults([]string{"test"})
	assert.Equal(t, expected, cmd.VisibleCategories())

	cmd = &Command{
		Name:     "visible-categories",
		HideHelp: true,
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
				Hidden:   true,
			},
			{
				Name:     "command2",
				Category: "2",
				Hidden:   true,
			},
			{
				Name:     "command3",
				Category: "3",
				Hidden:   true,
			},
		},
	}

	cmd.setupDefaults([]string{"test"})
	assert.Empty(t, cmd.VisibleCategories())
}

func TestCommand_Run_SubcommandDoesNotOverwriteErrorFromBefore(t *testing.T) {
	cmd := &Command{
		Commands: []*Command{
			{
				Commands: []*Command{
					{
						Name: "sub",
					},
				},
				Name:   "bar",
				Before: func(context.Context, *Command) (context.Context, error) { return nil, fmt.Errorf("before error") },
				After:  func(context.Context, *Command) error { return fmt.Errorf("after error") },
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "bar"})
	assert.ErrorContains(t, err, "before error")
	assert.ErrorContains(t, err, "after error")
}

func TestCommand_OnUsageError_WithWrongFlagValue_ForSubcommand(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&Int64Flag{Name: "flag"},
		},
		OnUsageError: func(_ context.Context, _ *Command, err error, isSubcommand bool) error {
			assert.False(t, isSubcommand, "Expect subcommand")
			assert.ErrorContains(t, err, "\"wrong\": invalid syntax")
			return errors.New("intercepted: " + err.Error())
		},
		Commands: []*Command{
			{
				Name: "bar",
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "--flag=wrong", "bar"})
	assert.ErrorContains(t, err, "parsing \"wrong\": invalid syntax", "Expect an intercepted error")
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

func (c *customBoolFlag) PreParse() error {
	return nil
}

func (c *customBoolFlag) PostParse() error {
	return nil
}

func (c *customBoolFlag) Get() any {
	dest := false
	return &boolValue{
		destination: &dest,
	}
}

func (c *customBoolFlag) Set(_, _ string) error {
	return nil
}

func (c *customBoolFlag) RunAction(context.Context, *Command) error {
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

	err := cmd.Run(buildTestContext(t), []string{"foo"})
	assert.NoError(t, err, "Run returned unexpected error")
}

func TestCustomFlagsUsed(t *testing.T) {
	cmd := &Command{
		Flags:  []Flag{&customBoolFlag{"custom"}},
		Writer: io.Discard,
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "--custom=bar"})
	assert.NoError(t, err, "Run returned unexpected error")
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

	err := cmd.Run(buildTestContext(t), []string{"foo", "--help-custom=bar"})
	assert.NoError(t, err, "Run returned unexpected error")
}

func TestHandleExitCoder_Default(t *testing.T) {
	app := buildMinimalTestCommand()
	_ = app.handleExitCoder(context.Background(), Exit("Default Behavior Error", 42))

	output := fakeErrWriter.String()
	assert.Contains(t, output, "Default", "Expected Default Behavior from Error Handler")
}

func TestHandleExitCoder_Custom(t *testing.T) {
	cmd := buildMinimalTestCommand()

	cmd.ExitErrHandler = func(context.Context, *Command, error) {
		_, _ = fmt.Fprintln(ErrWriter, "I'm a Custom error handler, I print what I want!")
	}

	_ = cmd.handleExitCoder(context.Background(), Exit("Default Behavior Error", 42))

	output := fakeErrWriter.String()
	assert.Contains(t, output, "Custom", "Expected Custom Behavior from Error Handler")
}

func TestShellCompletionForIncompleteFlags(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&Int64Flag{
				Name: "test-completion",
			},
		},
		EnableShellCompletion: true,
		ShellComplete: func(_ context.Context, cmd *Command) {
			for _, command := range cmd.Commands {
				if command.Hidden {
					continue
				}

				for _, name := range command.Names() {
					_, _ = fmt.Fprintln(cmd.Writer, name)
				}
			}

			for _, fl := range cmd.Flags {
				for _, name := range fl.Names() {
					if name == GenerateShellCompletionFlag.Names()[0] {
						continue
					}

					switch name = strings.TrimSpace(name); len(name) {
					case 0:
					case 1:
						_, _ = fmt.Fprintln(cmd.Writer, "-"+name)
					default:
						_, _ = fmt.Fprintln(cmd.Writer, "--"+name)
					}
				}
			}
		},
		Action: func(context.Context, *Command) error {
			return fmt.Errorf("should not get here")
		},
		Writer: io.Discard,
	}

	err := cmd.Run(buildTestContext(t), []string{"", "--test-completion", completionFlag})
	assert.NoError(t, err, "app should not return an error")
}

func TestWhenExitSubCommandWithCodeThenCommandQuitUnexpectedly(t *testing.T) {
	testCode := 104

	cmd := buildMinimalTestCommand()
	cmd.Commands = []*Command{
		{
			Name: "cmd",
			Commands: []*Command{
				{
					Name: "subcmd",
					Action: func(context.Context, *Command) error {
						return Exit("exit error", testCode)
					},
				},
			},
		},
	}

	// set user function as ExitErrHandler
	exitCodeFromExitErrHandler := int(0)
	cmd.ExitErrHandler = func(_ context.Context, _ *Command, err error) {
		if exitErr, ok := err.(ExitCoder); ok {
			exitCodeFromExitErrHandler = exitErr.ExitCode()
		}
	}

	// keep and restore original OsExiter
	origExiter := OsExiter
	t.Cleanup(func() { OsExiter = origExiter })

	// set user function as OsExiter
	exitCodeFromOsExiter := int(0)
	OsExiter = func(exitCode int) {
		exitCodeFromOsExiter = exitCode
	}

	r := require.New(t)

	r.Error(cmd.Run(buildTestContext(t), []string{
		"myapp",
		"cmd",
		"subcmd",
	}))

	r.Equal(0, exitCodeFromOsExiter)
	r.Equal(testCode, exitCodeFromExitErrHandler)
}

func buildMinimalTestCommand() *Command {
	// reset the help flag because tests may have set it
	HelpFlag.(*BoolFlag).hasBeenSet = false
	return &Command{Writer: io.Discard}
}

func TestSetupInitializesBothWriters(t *testing.T) {
	cmd := &Command{}

	cmd.setupDefaults([]string{"test"})

	assert.Equal(t, cmd.ErrWriter, os.Stderr, "expected a.ErrWriter to be os.Stderr")
	assert.Equal(t, cmd.Writer, os.Stdout, "expected a.Writer to be os.Stdout")
}

func TestSetupInitializesOnlyNilWriters(t *testing.T) {
	wr := &bytes.Buffer{}
	cmd := &Command{
		ErrWriter: wr,
	}

	cmd.setupDefaults([]string{"test"})

	assert.Equal(t, cmd.ErrWriter, wr, "expected a.ErrWriter to be a *bytes.Buffer instance")
	assert.Equal(t, cmd.Writer, os.Stdout, "expected a.Writer to be os.Stdout")
}

func TestFlagAction(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Minute)
	testCases := []struct {
		name string
		args []string
		err  string
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
			err:  "flag needs an argument: --f_string=",
		},
		{
			name: "flag_string_slice",
			args: []string{"app", "--f_string_slice=s1,s2,s3"},
			exp:  "[s1 s2 s3] ",
		},
		{
			name: "flag_string_slice_error",
			args: []string{"app", "--f_string_slice=err"},
			err:  "error string slice",
		},
		{
			name: "flag_bool",
			args: []string{"app", "--f_bool"},
			exp:  "true ",
		},
		{
			name: "flag_bool_error",
			args: []string{"app", "--f_bool=false"},
			err:  "value is false",
		},
		{
			name: "flag_duration",
			args: []string{"app", "--f_duration=1h30m20s"},
			exp:  "1h30m20s ",
		},
		{
			name: "flag_duration_error",
			args: []string{"app", "--f_duration=0"},
			err:  "empty duration",
		},
		{
			name: "flag_float64",
			args: []string{"app", "--f_float64=3.14159"},
			exp:  "3.14159 ",
		},
		{
			name: "flag_float64_error",
			args: []string{"app", "--f_float64=-1"},
			err:  "negative float64",
		},
		{
			name: "flag_float64_slice",
			args: []string{"app", "--f_float64_slice=1.1,2.2,3.3"},
			exp:  "[1.1 2.2 3.3] ",
		},
		{
			name: "flag_float64_slice_error",
			args: []string{"app", "--f_float64_slice=-1"},
			err:  "invalid float64 slice",
		},
		{
			name: "flag_int",
			args: []string{"app", "--f_int=1"},
			exp:  "1 ",
		},
		{
			name: "flag_int_error",
			args: []string{"app", "--f_int=-1"},
			err:  "negative int",
		},
		{
			name: "flag_int_slice",
			args: []string{"app", "--f_int_slice=1,2,3"},
			exp:  "[1 2 3] ",
		},
		{
			name: "flag_int_slice_error",
			args: []string{"app", "--f_int_slice=-1"},
			err:  "invalid int slice",
		},
		{
			name: "flag_timestamp",
			args: []string{"app", "--f_timestamp", now.Format(time.DateTime)},
			exp:  now.UTC().Format(time.RFC3339) + " ",
		},
		{
			name: "flag_timestamp_error",
			args: []string{"app", "--f_timestamp", "0001-01-01 00:00:00"},
			err:  "zero timestamp",
		},
		{
			name: "flag_uint",
			args: []string{"app", "--f_uint=1"},
			exp:  "1 ",
		},
		{
			name: "flag_uint_error",
			args: []string{"app", "--f_uint=0"},
			err:  "zero uint64",
		},
		{
			name: "flag_no_action",
			args: []string{"app", "--f_no_action=xx"},
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
		// TBD
		/*		{
				name: "mixture",
				args: []string{"app", "--f_string=app", "--f_uint=1", "--f_int_slice=1,2,3", "--f_duration=1h30m20s", "c1", "--f_string=c1", "sub1", "--f_string=sub1"},
				exp:  "app 1 [1 2 3] 1h30m20s c1 sub1 ",
			},*/
		{
			name: "flag_string_map",
			args: []string{"app", "--f_string_map=s1=s2,s3="},
			exp:  "map[s1:s2 s3:]",
		},
		{
			name: "flag_string_map_error",
			args: []string{"app", "--f_string_map=err="},
			err:  "error string map",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			out := &bytes.Buffer{}

			newStringFlag := func(local bool) *StringFlag {
				return &StringFlag{
					Local: local,
					Name:  "f_string",
					Action: func(_ context.Context, cmd *Command, v string) error {
						if v == "" {
							return fmt.Errorf("empty string")
						}
						_, err := cmd.Root().Writer.Write([]byte(v + " "))
						return err
					},
				}
			}

			cmd := &Command{
				Writer: out,
				Name:   "app",
				Commands: []*Command{
					{
						Name:   "c1",
						Flags:  []Flag{newStringFlag(true)},
						Action: func(_ context.Context, cmd *Command) error { return nil },
						Commands: []*Command{
							{
								Name:   "sub1",
								Action: func(context.Context, *Command) error { return nil },
								Flags:  []Flag{newStringFlag(true)},
							},
						},
					},
				},
				Flags: []Flag{
					newStringFlag(true),
					&StringFlag{
						Name: "f_no_action",
					},
					&StringSliceFlag{
						Local: true,
						Name:  "f_string_slice",
						Action: func(_ context.Context, cmd *Command, v []string) error {
							if v[0] == "err" {
								return fmt.Errorf("error string slice")
							}
							_, err := fmt.Fprintf(cmd.Root().Writer, "%v ", v)
							return err
						},
					},
					&BoolFlag{
						Name:  "f_bool",
						Local: true,
						Action: func(_ context.Context, cmd *Command, v bool) error {
							if !v {
								return fmt.Errorf("value is false")
							}
							_, err := fmt.Fprintf(cmd.Root().Writer, "%t ", v)
							return err
						},
					},
					&DurationFlag{
						Name:  "f_duration",
						Local: true,
						Action: func(_ context.Context, cmd *Command, v time.Duration) error {
							if v == 0 {
								return fmt.Errorf("empty duration")
							}
							_, err := fmt.Fprintf(cmd.Root().Writer, v.String()+" ")
							return err
						},
					},
					&FloatFlag{
						Name:  "f_float64",
						Local: true,
						Action: func(_ context.Context, cmd *Command, v float64) error {
							if v < 0 {
								return fmt.Errorf("negative float64")
							}
							_, err := fmt.Fprintf(cmd.Root().Writer, strconv.FormatFloat(v, 'f', -1, 64)+" ")
							return err
						},
					},
					&FloatSliceFlag{
						Name:  "f_float64_slice",
						Local: true,
						Action: func(_ context.Context, cmd *Command, v []float64) error {
							if len(v) > 0 && v[0] < 0 {
								return fmt.Errorf("invalid float64 slice")
							}
							_, err := fmt.Fprintf(cmd.Root().Writer, "%v ", v)
							return err
						},
					},
					&Int64Flag{
						Name:  "f_int",
						Local: true,
						Action: func(_ context.Context, cmd *Command, v int64) error {
							if v < 0 {
								return fmt.Errorf("negative int")
							}
							_, err := fmt.Fprintf(cmd.Root().Writer, "%v ", v)
							return err
						},
					},
					&Int64SliceFlag{
						Name:  "f_int_slice",
						Local: true,
						Action: func(_ context.Context, cmd *Command, v []int64) error {
							if len(v) > 0 && v[0] < 0 {
								return fmt.Errorf("invalid int slice")
							}
							_, err := fmt.Fprintf(cmd.Root().Writer, "%v ", v)
							return err
						},
					},
					&TimestampFlag{
						Name:  "f_timestamp",
						Local: true,
						Config: TimestampConfig{
							Timezone: time.UTC,
							Layouts:  []string{time.DateTime},
						},
						Action: func(_ context.Context, cmd *Command, v time.Time) error {
							if v.IsZero() {
								return fmt.Errorf("zero timestamp")
							}

							_, err := cmd.Root().Writer.Write([]byte(v.Format(time.RFC3339) + " "))
							return err
						},
					},
					&Uint64Flag{
						Name:  "f_uint",
						Local: true,
						Action: func(_ context.Context, cmd *Command, v uint64) error {
							if v == 0 {
								return fmt.Errorf("zero uint64")
							}
							_, err := fmt.Fprintf(cmd.Root().Writer, "%v ", v)
							return err
						},
					},
					&StringMapFlag{
						Name:  "f_string_map",
						Local: true,
						Action: func(_ context.Context, cmd *Command, v map[string]string) error {
							if _, ok := v["err"]; ok {
								return fmt.Errorf("error string map")
							}
							_, err := fmt.Fprintf(cmd.Root().Writer, "%v", v)
							return err
						},
					},
				},
				Action: func(context.Context, *Command) error { return nil },
			}

			err := cmd.Run(buildTestContext(t), test.args)

			r := require.New(t)

			if test.err != "" {
				r.EqualError(err, test.err)
				return
			}

			r.NoError(err)
			r.Equal(test.exp, out.String())
		})
	}
}

func TestPersistentFlag(t *testing.T) {
	var topInt, topPersistentInt, subCommandInt, appOverrideInt int64
	var appFlag string
	var appRequiredFlag string
	var appOverrideCmdInt int64
	var appSliceFloat64 []float64
	var persistentCommandSliceInt []int64
	var persistentFlagActionCount int64

	cmd := &Command{
		Flags: []Flag{
			&StringFlag{
				Name:        "persistentCommandFlag",
				Destination: &appFlag,
				Action: func(context.Context, *Command, string) error {
					persistentFlagActionCount++
					return nil
				},
			},
			&Int64SliceFlag{
				Name:        "persistentCommandSliceFlag",
				Destination: &persistentCommandSliceInt,
			},
			&FloatSliceFlag{
				Name:  "persistentCommandFloatSliceFlag",
				Value: []float64{11.3, 12.5},
			},
			&Int64Flag{
				Name:        "persistentCommandOverrideFlag",
				Destination: &appOverrideInt,
			},
			&StringFlag{
				Name:        "persistentRequiredCommandFlag",
				Required:    true,
				Destination: &appRequiredFlag,
			},
		},
		Commands: []*Command{
			{
				Name: "cmd",
				Flags: []Flag{
					&Int64Flag{
						Name:        "cmdFlag",
						Destination: &topInt,
						Local:       true,
					},
					&Int64Flag{
						Name:        "cmdPersistentFlag",
						Destination: &topPersistentInt,
					},
					&Int64Flag{
						Name:        "paof",
						Aliases:     []string{"persistentCommandOverrideFlag"},
						Destination: &appOverrideCmdInt,
						Local:       true,
					},
				},
				Commands: []*Command{
					{
						Name: "subcmd",
						Flags: []Flag{
							&Int64Flag{
								Name:        "cmdFlag",
								Destination: &subCommandInt,
								Local:       true,
							},
						},
						Action: func(_ context.Context, cmd *Command) error {
							appSliceFloat64 = cmd.FloatSlice("persistentCommandFloatSliceFlag")
							return nil
						},
					},
				},
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{
		"app",
		"--persistentCommandFlag", "hello",
		"--persistentCommandSliceFlag", "100",
		"--persistentCommandOverrideFlag", "102",
		"cmd",
		"--cmdFlag", "12",
		"--persistentCommandSliceFlag", "102",
		"--persistentCommandFloatSliceFlag", "102.455",
		"--paof", "105",
		"--persistentRequiredCommandFlag", "hellor",
		"subcmd",
		"--cmdPersistentFlag", "20",
		"--cmdFlag", "11",
		"--persistentCommandFlag", "bar",
		"--persistentCommandSliceFlag", "130",
		"--persistentCommandFloatSliceFlag", "3.1445",
	})

	require.NoError(t, err)

	assert.Equal(t, "bar", appFlag)
	assert.Equal(t, "hellor", appRequiredFlag)
	assert.Equal(t, int64(12), topInt)
	assert.Equal(t, int64(20), topPersistentInt)

	// this should be changed from app since
	// cmd overrides it
	assert.Equal(t, int64(102), appOverrideInt)
	assert.Equal(t, int64(11), subCommandInt)
	assert.Equal(t, int64(105), appOverrideCmdInt)
	assert.Equal(t, []int64{100, 102, 130}, persistentCommandSliceInt)
	assert.Equal(t, []float64{102.455, 3.1445}, appSliceFloat64)
	assert.Equal(t, int64(2), persistentFlagActionCount, "Expected persistent flag action to be called 2 times")
}

func TestPersistentFlagIsSet(t *testing.T) {
	result := ""
	resultIsSet := false

	app := &Command{
		Name: "root",
		Flags: []Flag{
			&StringFlag{
				Name: "result",
			},
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(_ context.Context, cmd *Command) error {
					result = cmd.String("result")
					resultIsSet = cmd.IsSet("result")
					return nil
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"root", "--result", "before", "sub"})
	require.NoError(t, err)
	require.Equal(t, "before", result)
	require.True(t, resultIsSet)

	err = app.Run(context.Background(), []string{"root", "sub", "--result", "after"})
	require.NoError(t, err)
	require.Equal(t, "after", result)
	require.True(t, resultIsSet)
}

func TestRequiredFlagDelayed(t *testing.T) {
	sf := &StringFlag{
		Name:     "result",
		Required: true,
	}

	expectedErr := &errRequiredFlags{
		missingFlags: []string{sf.Name},
	}

	tests := []struct {
		name        string
		args        []string
		errExpected error
	}{
		{
			name:        "leaf help",
			args:        []string{"root", "sub", "-h"},
			errExpected: nil,
		},
		{
			name:        "leaf action",
			args:        []string{"root", "sub"},
			errExpected: expectedErr,
		},
		{
			name:        "leaf flags set",
			args:        []string{"root", "sub", "--if", "10"},
			errExpected: expectedErr,
		},
		{
			name:        "leaf invalid flags set",
			args:        []string{"root", "sub", "--xx"},
			errExpected: expectedErr,
		},
	}

	app := &Command{
		Name: "root",
		Flags: []Flag{
			sf,
		},
		Commands: []*Command{
			{
				Name: "sub",
				Flags: []Flag{
					&Int64Flag{
						Name:     "if",
						Required: true,
					},
				},
				Action: func(ctx context.Context, c *Command) error {
					return nil
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := app.Run(context.Background(), test.args)
			if test.errExpected == nil {
				require.NoError(t, err)
			} else {
				require.ErrorAs(t, err, &test.errExpected)
			}
		})
	}
}

func TestRequiredPersistentFlag(t *testing.T) {
	app := &Command{
		Name: "root",
		Flags: []Flag{
			&StringFlag{
				Name:     "result",
				Required: true,
			},
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(ctx context.Context, c *Command) error {
					return nil
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"root", "sub"})
	require.Error(t, err)

	err = app.Run(context.Background(), []string{"root", "sub", "--result", "after"})
	require.NoError(t, err)
}

func TestFlagDuplicates(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		errExpected bool
	}{
		{
			name: "all args present once",
			args: []string{"foo", "--sflag", "hello", "--isflag", "1", "--isflag", "2", "--fsflag", "2.0", "--iflag", "10", "--bifflag"},
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
		{
			name:        "duplicate bool inverse flag(non duplicatable)",
			args:        []string{"foo", "--bifflag", "--bifflag"},
			errExpected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &Command{
				Flags: []Flag{
					&StringFlag{
						Name:     "sflag",
						OnlyOnce: true,
					},
					&Int64SliceFlag{
						Name: "isflag",
					},
					&FloatSliceFlag{
						Name:     "fsflag",
						OnlyOnce: true,
					},
					&BoolWithInverseFlag{
						Name:     "bifflag",
						OnlyOnce: true,
					},
					&Int64Flag{
						Name: "iflag",
					},
				},
				Action: func(context.Context, *Command) error {
					return nil
				},
			}

			err := cmd.Run(buildTestContext(t), test.args)
			if test.errExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShorthandCommand(t *testing.T) {
	af := func(p *int) ActionFunc {
		return func(context.Context, *Command) error {
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

	err := cmd.Run(buildTestContext(t), []string{"foo", "cth"})
	assert.NoError(t, err)
	assert.True(t, cmd1 == 1 && cmd2 == 0, "Expected command1 to be triggered once")

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(buildTestContext(t), []string{"foo", "cthd"})
	assert.NoError(t, err)
	assert.True(t, cmd1 == 1 && cmd2 == 0, "Expected command1 to be triggered once")

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(buildTestContext(t), []string{"foo", "cthe"})
	assert.NoError(t, err)
	assert.True(t, cmd1 == 1 && cmd2 == 0, "Expected command1 to be triggered once")

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(buildTestContext(t), []string{"foo", "cthert"})
	assert.NoError(t, err)
	assert.True(t, cmd1 == 0 && cmd2 == 1, "Expected command1 to be triggered once")

	cmd1 = 0
	cmd2 = 0

	err = cmd.Run(buildTestContext(t), []string{"foo", "cthet"})
	assert.NoError(t, err)
	assert.True(t, cmd1 == 0 && cmd2 == 1, "Expected command1 to be triggered once")
}

func TestCommand_Int(t *testing.T) {
	pCmd := &Command{
		Flags: []Flag{
			&Int64Flag{
				Name:  "myflag",
				Value: 12,
			},
		},
	}
	cmd := &Command{
		Flags: []Flag{
			&Int64Flag{
				Name:  "top-flag",
				Value: 13,
			},
		},
		parent: pCmd,
	}

	require.Equal(t, int64(12), cmd.Int64("myflag"))
	require.Equal(t, int64(13), cmd.Int64("top-flag"))
}

func TestCommand_Uint(t *testing.T) {
	pCmd := &Command{
		Flags: []Flag{
			&Uint64Flag{
				Name:  "myflagUint",
				Value: 13,
			},
		},
	}
	cmd := &Command{
		Flags: []Flag{
			&Uint64Flag{
				Name:  "top-flag",
				Value: 14,
			},
		},
		parent: pCmd,
	}

	require.Equal(t, uint64(13), cmd.Uint64("myflagUint"))
	require.Equal(t, uint64(14), cmd.Uint64("top-flag"))
}

func TestCommand_Float64(t *testing.T) {
	pCmd := &Command{
		Flags: []Flag{
			&FloatFlag{
				Name:  "myflag",
				Value: 17,
			},
		},
	}
	cmd := &Command{
		Flags: []Flag{
			&FloatFlag{
				Name:  "top-flag",
				Value: 18,
			},
		},
		parent: pCmd,
	}

	r := require.New(t)
	r.Equal(float64(17), cmd.Float("myflag"))
	r.Equal(float64(18), cmd.Float("top-flag"))
}

func TestCommand_Duration(t *testing.T) {
	pCmd := &Command{
		Flags: []Flag{
			&DurationFlag{
				Name:  "myflag",
				Value: 12 * time.Second,
			},
		},
	}
	cmd := &Command{
		Flags: []Flag{
			&DurationFlag{
				Name:  "top-flag",
				Value: 13 * time.Second,
			},
		},
		parent: pCmd,
	}

	r := require.New(t)
	r.Equal(12*time.Second, cmd.Duration("myflag"))
	r.Equal(13*time.Second, cmd.Duration("top-flag"))
}

func TestCommand_Timestamp(t *testing.T) {
	t1 := time.Time{}.Add(12 * time.Second)
	t2 := time.Time{}.Add(13 * time.Second)

	cmd := &Command{
		Name: "hello",
		Flags: []Flag{
			&TimestampFlag{
				Name:  "myflag",
				Value: t1,
			},
		},
		Action: func(ctx context.Context, c *Command) error {
			return nil
		},
	}

	pCmd := &Command{
		Flags: []Flag{
			&TimestampFlag{
				Name:  "top-flag",
				Value: t2,
			},
		},
		Commands: []*Command{
			cmd,
		},
	}

	err := pCmd.Run(context.Background(), []string{"foo", "hello"})
	assert.NoError(t, err)

	r := require.New(t)
	r.Equal(t1, cmd.Timestamp("myflag"))
	r.Equal(t2, cmd.Timestamp("top-flag"))
}

func TestCommand_String(t *testing.T) {
	pCmd := &Command{
		Flags: []Flag{
			&StringFlag{
				Name:  "myflag",
				Value: "hello world",
			},
		},
	}
	cmd := &Command{
		Flags: []Flag{
			&StringFlag{
				Name:  "top-flag",
				Value: "hai veld",
			},
		},
		parent: pCmd,
	}

	r := require.New(t)
	r.Equal("hello world", cmd.String("myflag"))
	r.Equal("hai veld", cmd.String("top-flag"))

	r.Equal("hai veld", cmd.String("top-flag"))
}

func TestCommand_Bool(t *testing.T) {
	pCmd := &Command{
		Flags: []Flag{
			&BoolFlag{
				Name: "myflag",
			},
		},
	}
	cmd := &Command{
		Flags: []Flag{
			&BoolFlag{
				Name:  "top-flag",
				Value: true,
			},
		},
		parent: pCmd,
	}

	r := require.New(t)
	r.False(cmd.Bool("myflag"))
	r.True(cmd.Bool("top-flag"))
}

func TestCommand_Value(t *testing.T) {
	subCmd := &Command{
		Name: "test",
		Flags: []Flag{
			&Int64Flag{
				Name:    "myflag",
				Usage:   "doc",
				Aliases: []string{"m", "mf"},
			},
		},
		Action: func(ctx context.Context, c *Command) error {
			return nil
		},
	}

	cmd := &Command{
		Flags: []Flag{
			&Int64Flag{
				Name:    "top-flag",
				Usage:   "doc",
				Aliases: []string{"t", "tf"},
			},
		},
		Commands: []*Command{
			subCmd,
		},
	}
	t.Run("flag name", func(t *testing.T) {
		r := require.New(t)
		err := cmd.Run(buildTestContext(t), []string{"main", "--top-flag", "13", "test", "--myflag", "14"})

		r.NoError(err)
		r.Equal(int64(13), cmd.Value("top-flag"))
		r.Equal(int64(13), cmd.Value("t"))
		r.Equal(int64(13), cmd.Value("tf"))

		r.Equal(int64(14), subCmd.Value("myflag"))
		r.Equal(int64(14), subCmd.Value("m"))
		r.Equal(int64(14), subCmd.Value("mf"))
	})

	t.Run("flag aliases", func(t *testing.T) {
		r := require.New(t)
		err := cmd.Run(buildTestContext(t), []string{"main", "-tf", "15", "test", "-m", "16"})

		r.NoError(err)
		r.Equal(int64(15), cmd.Value("top-flag"))
		r.Equal(int64(15), cmd.Value("t"))
		r.Equal(int64(15), cmd.Value("tf"))

		r.Equal(int64(16), subCmd.Value("myflag"))
		r.Equal(int64(16), subCmd.Value("m"))
		r.Equal(int64(16), subCmd.Value("mf"))
		r.Nil(cmd.Value("unknown-flag"))
	})
}

func TestCommand_Value_InvalidFlagAccessHandler(t *testing.T) {
	var flagName string
	cmd := &Command{
		InvalidFlagAccessHandler: func(_ context.Context, _ *Command, name string) {
			flagName = name
		},
		Commands: []*Command{
			{
				Name: "command",
				Commands: []*Command{
					{
						Name: "subcommand",
						Action: func(_ context.Context, cmd *Command) error {
							cmd.Value("missing")
							return nil
						},
					},
				},
			},
		},
	}

	r := require.New(t)

	r.NoError(cmd.Run(buildTestContext(t), []string{"run", "command", "subcommand"}))
	r.Equal("missing", flagName)
}

func TestCommand_Args(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&BoolFlag{
				Name: "myflag",
			},
		},
	}
	_ = cmd.Run(context.Background(), []string{"", "--myflag", "bat", "baz"})

	r := require.New(t)
	r.Equal(2, cmd.Args().Len())
	r.True(cmd.Bool("myflag"))
	r.Equal(2, cmd.NArg())
}

func TestCommand_IsSet(t *testing.T) {
	cmd := &Command{
		Name: "frob",
		Flags: []Flag{
			&BoolFlag{
				Name: "one-flag",
			},
			&BoolFlag{
				Name: "two-flag",
			},
			&StringFlag{
				Name:  "three-flag",
				Value: "hello world",
			},
		},
	}
	pCmd := &Command{
		Name: "root",
		Flags: []Flag{
			&BoolFlag{
				Name:  "top-flag",
				Value: true,
			},
		},
		Commands: []*Command{
			cmd,
		},
	}

	r := require.New(t)

	r.NoError(pCmd.Run(context.Background(), []string{"foo", "frob", "--one-flag", "--top-flag", "--two-flag", "--three-flag", "dds"}))

	r.True(cmd.IsSet("one-flag"))
	r.True(cmd.IsSet("two-flag"))
	r.True(cmd.IsSet("three-flag"))
	r.True(cmd.IsSet("top-flag"))
	r.False(cmd.IsSet("bogus"))
}

// XXX Corresponds to hack in context.IsSet for flags with EnvVar field
// Should be moved to `flag_test` in v2
func TestCommand_IsSet_fromEnv(t *testing.T) {
	var (
		timeoutIsSet, tIsSet    bool
		noEnvVarIsSet, nIsSet   bool
		passwordIsSet, pIsSet   bool
		unparsableIsSet, uIsSet bool
	)

	t.Setenv("APP_TIMEOUT_SECONDS", "15.5")
	t.Setenv("APP_PASSWORD", "")

	cmd := &Command{
		Flags: []Flag{
			&FloatFlag{Name: "timeout", Aliases: []string{"t"}, Local: true, Sources: EnvVars("APP_TIMEOUT_SECONDS")},
			&StringFlag{Name: "password", Aliases: []string{"p"}, Local: true, Sources: EnvVars("APP_PASSWORD")},
			&FloatFlag{Name: "unparsable", Aliases: []string{"u"}, Local: true, Sources: EnvVars("APP_UNPARSABLE")},
			&FloatFlag{Name: "no-env-var", Aliases: []string{"n"}, Local: true},
		},
		Action: func(_ context.Context, cmd *Command) error {
			timeoutIsSet = cmd.IsSet("timeout")
			tIsSet = cmd.IsSet("t")
			passwordIsSet = cmd.IsSet("password")
			pIsSet = cmd.IsSet("p")
			unparsableIsSet = cmd.IsSet("unparsable")
			uIsSet = cmd.IsSet("u")
			noEnvVarIsSet = cmd.IsSet("no-env-var")
			nIsSet = cmd.IsSet("n")
			return nil
		},
	}

	r := require.New(t)

	r.NoError(cmd.Run(buildTestContext(t), []string{"run"}))
	r.True(timeoutIsSet)
	r.True(tIsSet)
	r.True(passwordIsSet)
	r.True(pIsSet)
	r.False(noEnvVarIsSet)
	r.False(nIsSet)

	t.Setenv("APP_UNPARSABLE", "foobar")

	r.Error(cmd.Run(buildTestContext(t), []string{"run"}))
	r.False(unparsableIsSet)
	r.False(uIsSet)
}

func TestCommand_NumFlags(t *testing.T) {
	rootCmd := &Command{
		Flags: []Flag{
			&BoolFlag{
				Name:  "myflagGlobal",
				Value: true,
			},
		},
	}
	cmd := &Command{
		Flags: []Flag{
			&BoolFlag{
				Name: "myflag",
			},
			&StringFlag{
				Name:  "otherflag",
				Value: "hello world",
			},
		},
	}

	_ = cmd.Run(context.Background(), []string{"", "--myflag", "--otherflag=foo"})
	_ = rootCmd.Run(context.Background(), []string{"", "--myflagGlobal"})
	require.Equal(t, 2, cmd.NumFlags())
	actualFlags := cmd.LocalFlagNames()
	sort.Strings(actualFlags)

	require.Equal(t, []string{"myflag", "otherflag"}, actualFlags)

	actualFlags = cmd.FlagNames()
	sort.Strings(actualFlags)

	require.Equal(t, []string{"myflag", "otherflag"}, actualFlags)

	cmd.parent = rootCmd
	lineage := cmd.Lineage()

	r := require.New(t)
	r.Equal(2, len(lineage))
	r.Equal(cmd, lineage[0])
	r.Equal(rootCmd, lineage[1])
}

func TestCommand_Set(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&Int64Flag{
				Name:  "int",
				Value: 5,
			},
		},
	}
	r := require.New(t)

	r.False(cmd.IsSet("int"))
	r.NoError(cmd.Set("int", "1"))
	r.Equal(int64(1), cmd.Int64("int"))
	r.True(cmd.IsSet("int"))
}

func TestCommand_Set_InvalidFlagAccessHandler(t *testing.T) {
	var flagName string
	cmd := &Command{
		InvalidFlagAccessHandler: func(_ context.Context, _ *Command, name string) {
			flagName = name
		},
	}

	r := require.New(t)

	r.True(cmd.Set("missing", "") != nil)
	r.Equal("missing", flagName)
}

func TestCommand_lookupFlag(t *testing.T) {
	pCmd := &Command{
		Flags: []Flag{
			&BoolFlag{
				Name:  "top-flag",
				Value: true,
			},
		},
	}
	cmd := &Command{
		Flags: []Flag{
			&BoolFlag{
				Name: "local-flag",
			},
		},
	}
	_ = cmd.Run(context.Background(), []string{"--local-flag"})
	pCmd.Commands = []*Command{cmd}
	_ = pCmd.Run(context.Background(), []string{"--top-flag"})

	r := require.New(t)

	fs := cmd.lookupFlag("top-flag")
	r.Equal(pCmd.Flags[0], fs)

	fs = cmd.lookupFlag("local-flag")
	r.Equal(cmd.Flags[0], fs)
	r.Nil(cmd.lookupFlag("frob"))
}

func TestCommandAttributeAccessing(t *testing.T) {
	tdata := []struct {
		testCase     string
		setBoolInput string
		ctxBoolInput string
		parent       *Command
	}{
		{
			testCase:     "empty",
			setBoolInput: "",
			ctxBoolInput: "",
		},
		{
			testCase:     "empty_with_background_context",
			setBoolInput: "",
			ctxBoolInput: "",
			parent:       &Command{},
		},
		{
			testCase:     "empty_set_bool_and_present_ctx_bool",
			setBoolInput: "",
			ctxBoolInput: "ctx-bool",
		},
		{
			testCase:     "present_set_bool_and_present_ctx_bool_with_background_context",
			setBoolInput: "",
			ctxBoolInput: "ctx-bool",
			parent:       &Command{},
		},
		{
			testCase:     "present_set_bool_and_present_ctx_bool",
			setBoolInput: "ctx-bool",
			ctxBoolInput: "ctx-bool",
		},
		{
			testCase:     "present_set_bool_and_present_ctx_bool_with_background_context",
			setBoolInput: "ctx-bool",
			ctxBoolInput: "ctx-bool",
			parent:       &Command{},
		},
		{
			testCase:     "present_set_bool_and_different_ctx_bool",
			setBoolInput: "ctx-bool",
			ctxBoolInput: "not-ctx-bool",
		},
		{
			testCase:     "present_set_bool_and_different_ctx_bool_with_background_context",
			setBoolInput: "ctx-bool",
			ctxBoolInput: "not-ctx-bool",
			parent:       &Command{},
		},
	}

	for _, test := range tdata {
		t.Run(test.testCase, func(t *testing.T) {
			cmd := &Command{parent: test.parent}

			require.False(t, cmd.Bool(test.ctxBoolInput))
		})
	}
}

func TestCheckRequiredFlags(t *testing.T) {
	tdata := []struct {
		testCase              string
		parseInput            []string
		envVarInput           [2]string
		flags                 []Flag
		expectedAnError       bool
		expectedErrorContents []string
	}{
		{
			testCase: "empty",
		},
		{
			testCase: "optional",
			flags: []Flag{
				&StringFlag{Name: "optionalFlag"},
			},
		},
		{
			testCase: "required",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
			},
			expectedAnError:       true,
			expectedErrorContents: []string{"requiredFlag"},
		},
		{
			testCase: "required_and_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
			},
			parseInput: []string{"--requiredFlag", "myinput"},
		},
		{
			testCase: "required_and_present_via_env_var",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true, Sources: EnvVars("REQUIRED_FLAG")},
			},
			envVarInput: [2]string{"REQUIRED_FLAG", "true"},
		},
		{
			testCase: "required_and_optional",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "optionalFlag"},
			},
			expectedAnError: true,
		},
		{
			testCase: "required_and_optional_and_optional_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "optionalFlag"},
			},
			parseInput:      []string{"--optionalFlag", "myinput"},
			expectedAnError: true,
		},
		{
			testCase: "required_and_optional_and_optional_present_via_env_var",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "optionalFlag", Sources: EnvVars("OPTIONAL_FLAG")},
			},
			envVarInput:     [2]string{"OPTIONAL_FLAG", "true"},
			expectedAnError: true,
		},
		{
			testCase: "required_and_optional_and_required_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "optionalFlag"},
			},
			parseInput: []string{"--requiredFlag", "myinput"},
		},
		{
			testCase: "two_required",
			flags: []Flag{
				&StringFlag{Name: "requiredFlagOne", Required: true},
				&StringFlag{Name: "requiredFlagTwo", Required: true},
			},
			expectedAnError:       true,
			expectedErrorContents: []string{"requiredFlagOne", "requiredFlagTwo"},
		},
		{
			testCase: "two_required_and_one_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "requiredFlagTwo", Required: true},
			},
			parseInput:      []string{"--requiredFlag", "myinput"},
			expectedAnError: true,
		},
		{
			testCase: "two_required_and_both_present",
			flags: []Flag{
				&StringFlag{Name: "requiredFlag", Required: true},
				&StringFlag{Name: "requiredFlagTwo", Required: true},
			},
			parseInput: []string{"--requiredFlag", "myinput", "--requiredFlagTwo", "myinput"},
		},
		{
			testCase: "required_flag_with_short_name",
			flags: []Flag{
				&StringSliceFlag{Name: "names", Aliases: []string{"N"}, Required: true},
			},
			parseInput: []string{"-N", "asd", "-N", "qwe"},
		},
		{
			testCase: "required_flag_with_multiple_short_names",
			flags: []Flag{
				&StringSliceFlag{Name: "names", Aliases: []string{"N", "n"}, Required: true},
			},
			parseInput: []string{"-n", "asd", "-n", "qwe"},
		},
		{
			testCase:              "required_flag_with_short_alias_not_printed_on_error",
			expectedAnError:       true,
			expectedErrorContents: []string{"Required flag \"names\" not set"},
			flags: []Flag{
				&StringSliceFlag{Name: "names", Aliases: []string{"n"}, Required: true},
			},
		},
		{
			testCase:              "required_flag_with_one_character",
			expectedAnError:       true,
			expectedErrorContents: []string{"Required flag \"n\" not set"},
			flags: []Flag{
				&StringFlag{Name: "n", Required: true},
			},
		},
	}

	for _, test := range tdata {
		t.Run(test.testCase, func(t *testing.T) {
			// setup
			if test.envVarInput[0] != "" {
				t.Setenv(test.envVarInput[0], test.envVarInput[1])
			}

			cmd := &Command{
				Name:  "foo",
				Flags: test.flags,
			}
			args := []string{"foo"}
			args = append(args, test.parseInput...)
			_ = cmd.Run(context.Background(), args)

			err := cmd.checkAllRequiredFlags()

			// assertions
			if test.expectedAnError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			for _, errString := range test.expectedErrorContents {
				if err != nil {
					assert.ErrorContains(t, err, errString)
				}
			}
		})
	}
}

func TestCommand_ParentCommand_Set(t *testing.T) {
	cmd := &Command{
		parent: &Command{
			Flags: []Flag{
				&StringFlag{
					Name: "Name",
				},
			},
		},
	}

	err := cmd.Set("Name", "aaa")
	assert.NoError(t, err)
}

func TestCommandStringDashOption(t *testing.T) {
	tests := []struct {
		name                string
		shortOptionHandling bool
		args                []string
	}{
		{
			name: "double dash separate value",
			args: []string{"foo", "--bar", "-", "test"},
		},
		{
			name: "single dash separate value",
			args: []string{"foo", "-bar", "-", "test"},
		},
		/*{
			name:                "single dash combined value",
			args:                []string{"foo", "-b-", "test"},
			shortOptionHandling: true,
		},*/
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &Command{
				Name:                   "foo",
				UseShortOptionHandling: test.shortOptionHandling,
				Flags: []Flag{
					&StringFlag{
						Name:    "bar",
						Aliases: []string{"b"},
					},
				},
				Action: func(ctx context.Context, c *Command) error {
					return nil
				},
			}

			err := cmd.Run(buildTestContext(t), test.args)
			assert.NoError(t, err)

			assert.Equal(t, "-", cmd.String("b"))
		})
	}
}

func TestCommandReadArgsFromStdIn(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		args          []string
		expectedInt   int64
		expectedFloat float64
		expectedSlice []string
		expectError   bool
	}{
		{
			name:          "empty",
			input:         "",
			args:          []string{"foo"},
			expectedInt:   0,
			expectedFloat: 0.0,
			expectedSlice: []string{},
		},
		{
			name: "empty2",
			input: `

			`,
			args:          []string{"foo"},
			expectedInt:   0,
			expectedFloat: 0.0,
			expectedSlice: []string{},
		},
		{
			name:          "intflag-from-input",
			input:         "--if=100",
			args:          []string{"foo"},
			expectedInt:   100,
			expectedFloat: 0.0,
			expectedSlice: []string{},
		},
		{
			name: "intflag-from-input2",
			input: `
			--if

			100`,
			args:          []string{"foo"},
			expectedInt:   100,
			expectedFloat: 0.0,
			expectedSlice: []string{},
		},
		{
			name: "multiflag-from-input",
			input: `
			--if

			100
			--ff      100.1

			--ssf hello
			--ssf

			"hello
  123
44"
			`,
			args:          []string{"foo"},
			expectedInt:   100,
			expectedFloat: 100.1,
			expectedSlice: []string{"hello", "hello\n  123\n44"},
		},
		{
			name: "end-args",
			input: `
			--if

			100
			--
			--ff      100.1

			--ssf hello
			--ssf

			hell02
			`,
			args:          []string{"foo"},
			expectedInt:   100,
			expectedFloat: 0,
			expectedSlice: []string{},
		},
		{
			name: "invalid string",
			input: `
			"
			`,
			args:          []string{"foo"},
			expectedInt:   0,
			expectedFloat: 0,
			expectedSlice: []string{},
		},
		{
			name: "invalid string2",
			input: `
			--if
			"
			`,
			args:        []string{"foo"},
			expectError: true,
		},
		{
			name: "incomplete string",
			input: `
			--ssf
			"
			hello
			`,
			args:          []string{"foo"},
			expectedSlice: []string{"hello"},
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			r := require.New(t)

			fp, err := os.CreateTemp("", "readargs")
			r.NoError(err)
			_, err = fp.Write([]byte(tst.input))
			r.NoError(err)
			fp.Close()

			cmd := buildMinimalTestCommand()
			cmd.ReadArgsFromStdin = true
			cmd.Reader, err = os.Open(fp.Name())
			r.NoError(err)
			cmd.Flags = []Flag{
				&Int64Flag{
					Name: "if",
				},
				&FloatFlag{
					Name: "ff",
				},
				&StringSliceFlag{
					Name: "ssf",
				},
			}

			actionCalled := false
			cmd.Action = func(ctx context.Context, c *Command) error {
				r.Equal(tst.expectedInt, c.Int64("if"))
				r.Equal(tst.expectedFloat, c.Float("ff"))
				r.Equal(tst.expectedSlice, c.StringSlice("ssf"))
				actionCalled = true
				return nil
			}

			err = cmd.Run(context.Background(), tst.args)
			if !tst.expectError {
				r.NoError(err)
				r.True(actionCalled)
			} else {
				r.Error(err)
			}
		})
	}
}

func TestZeroValueCommand(t *testing.T) {
	var cmd Command
	assert.NoError(t, cmd.Run(context.Background(), []string{"foo"}))
}

func TestCommandInvalidName(t *testing.T) {
	var cmd Command
	assert.Equal(t, int64(0), cmd.Int64("foo"))
	assert.Equal(t, uint64(0), cmd.Uint64("foo"))
	assert.Equal(t, float64(0), cmd.Float("foo"))
	assert.Equal(t, "", cmd.String("foo"))
	assert.Equal(t, time.Time{}, cmd.Timestamp("foo"))
	assert.Equal(t, time.Duration(0), cmd.Duration("foo"))

	assert.Equal(t, []int64(nil), cmd.Int64Slice("foo"))
	assert.Equal(t, []uint64(nil), cmd.Uint64Slice("foo"))
	assert.Equal(t, []float64(nil), cmd.FloatSlice("foo"))
	assert.Equal(t, []string(nil), cmd.StringSlice("foo"))
}

func TestCommandCategories(t *testing.T) {
	var cc commandCategories = []*commandCategory{
		{
			name:     "foo",
			commands: []*Command{},
		},
		{
			name:     "bar",
			commands: []*Command{},
		},
		{
			name:     "goo",
			commands: nil,
		},
	}

	sort.Sort(&cc)

	var prev *commandCategory
	for _, c := range cc {
		if prev != nil {
			assert.LessOrEqual(t, prev.name, c.name)
		}
		prev = c
		assert.Equal(t, []*Command(nil), c.VisibleCommands())
	}
}

func TestCommandSliceFlagSeparator(t *testing.T) {
	oldSep := defaultSliceFlagSeparator
	defer func() {
		defaultSliceFlagSeparator = oldSep
	}()

	cmd := &Command{
		SliceFlagSeparator: ";",
		Flags: []Flag{
			&StringSliceFlag{
				Name: "foo",
			},
		},
	}

	r := require.New(t)
	r.NoError(cmd.Run(buildTestContext(t), []string{"app", "--foo", "ff;dd;gg", "--foo", "t,u"}))
	r.Equal([]string{"ff", "dd", "gg", "t,u"}, cmd.Value("foo"))
}

// TestStringFlagTerminator tests the string flag "--flag" with "--" terminator.
func TestStringFlagTerminator(t *testing.T) {
	tests := []struct {
		name         string
		input        []string
		expectFlag   string
		expectArgs   []string
		expectErr    bool
		errorContain string
	}{
		{
			name:       "flag and args after terminator",
			input:      []string{"test", "--flag", "x", "--", "test", "a1", "a2", "a3"},
			expectFlag: "x",
			expectArgs: []string{"test", "a1", "a2", "a3"},
		},
		/*	{
			name:         "missing flag value due to terminator",
			input:        []string{"test", "--flag", "--", "x"},
			expectErr:    true,
			errorContain: "flag needs an argument",
		},*/
		{
			name:       "terminator with no trailing args",
			input:      []string{"test", "--flag", "x", "--"},
			expectFlag: "x",
			expectArgs: []string{},
		},
		{
			name:       "no terminator, only flag",
			input:      []string{"test", "--flag", "x"},
			expectFlag: "x",
			expectArgs: []string{},
		},
		{
			name:       "flag defined after --",
			input:      []string{"test", "--", "x", "--flag=value"},
			expectFlag: "",
			expectArgs: []string{"x", "--flag=value"},
		},
		{
			name:       "flag and without --",
			input:      []string{"test", "--flag", "value", "x"},
			expectFlag: "value",
			expectArgs: []string{"x"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var flagVal string
			var argsVal []string

			// build minimal command with a StringFlag "flag"
			cmd := &Command{
				Name: "test",
				Flags: []Flag{
					&StringFlag{
						Name:        "flag",
						Usage:       "a string flag",
						Destination: &flagVal,
					},
				},
				Action: func(ctx context.Context, c *Command) error {
					argsVal = c.Args().Slice()
					return nil
				},
			}

			err := cmd.Run(context.Background(), tc.input)
			if tc.expectErr {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.errorContain))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectFlag, flagVal)
				assert.Equal(t, tc.expectArgs, argsVal)
			}
		})
	}
}

// TestBoolFlagTerminator tests the bool flag
func TestBoolFlagTerminator(t *testing.T) {
	tests := []struct {
		name         string
		input        []string
		expectFlag   bool
		expectArgs   []string
		expectErr    bool
		errorContain string
	}{
		/*{
			name:         "bool flag with invalid non-bool value",
			input:        []string{"test", "--flag", "x", "--", "test", "a1", "a2", "a3"},
			expectErr:    true,
			errorContain: "invalid syntax",
		},*/
		{
			name:       "bool flag omitted value defaults to true",
			input:      []string{"test", "--flag", "--", "x"},
			expectFlag: true,
			expectArgs: []string{"x"},
		},
		{
			name:       "bool flag explicitly set to false",
			input:      []string{"test", "--flag=false", "--", "x"},
			expectFlag: false,
			expectArgs: []string{"x"},
		},
		{
			name:       "bool flag defined after --",
			input:      []string{"test", "--", "x", "--flag=true"},
			expectFlag: false,
			expectArgs: []string{"x", "--flag=true"},
		},
		{
			name:       "bool flag and without --",
			input:      []string{"test", "--flag=true", "x"},
			expectFlag: true,
			expectArgs: []string{"x"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var flagVal bool
			var argsVal []string

			// build minimal command with a BoolFlag "flag"
			cmd := &Command{
				Name: "test",
				Flags: []Flag{
					&BoolFlag{
						Name:        "flag",
						Usage:       "a bool flag",
						Destination: &flagVal,
					},
				},
				Action: func(ctx context.Context, c *Command) error {
					argsVal = c.Args().Slice()
					return nil
				},
			}

			err := cmd.Run(context.Background(), tc.input)
			if tc.expectErr {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.errorContain))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectFlag, flagVal)
				assert.Equal(t, tc.expectArgs, argsVal)
			}
		})
	}
}

// TestSliceStringFlagParsing tests the StringSliceFlag
func TestSliceStringFlagParsing(t *testing.T) {
	var sliceVal []string

	cmdNoDelimiter := &Command{
		Name: "test",
		Flags: []Flag{
			&StringSliceFlag{
				Name:  "flag",
				Usage: "a string slice flag without delimiter",
			},
		},
		Action: func(ctx context.Context, c *Command) error {
			sliceVal = c.StringSlice("flag")
			return nil
		},
	}

	/*cmdWithDelimiter := &Command{
		Name: "test",
		Flags: []Flag{
			&StringSliceFlag{
				Name:      "flag",
				Usage:     "a string slice flag with delimiter",
				Delimiter: ':',
			},
		},
		Action: func(ctx context.Context, c *Command) error {
			sliceVal = c.StringSlice("flag")
			return nil
		},
	}*/

	tests := []struct {
		name         string
		cmd          *Command
		input        []string
		expectSlice  []string
		expectErr    bool
		errorContain string
	}{
		{
			name:        "single value without delimiter (no split)",
			cmd:         cmdNoDelimiter,
			input:       []string{"test", "--flag", "x"},
			expectSlice: []string{"x"},
		},
		{
			name:        "multiple values with comma (default split)",
			cmd:         cmdNoDelimiter,
			input:       []string{"test", "--flag", "x,y"},
			expectSlice: []string{"x", "y"},
		},
		/*{
			name:        "Case 10: with delimiter specified ':'",
			cmd:         cmdWithDelimiter,
			input:       []string{"test", "--flag", "x:y"},
			expectSlice: []string{"x", "y"},
		},*/
		{
			name:        "without delimiter specified, value remains unsplit",
			cmd:         cmdNoDelimiter,
			input:       []string{"test", "--flag", "x:y"},
			expectSlice: []string{"x:y"},
		},
	}

	for _, tc := range tests {
		// Reset sliceVal
		sliceVal = nil

		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Run(context.Background(), tc.input)
			if tc.expectErr {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.errorContain))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectSlice, sliceVal)
			}
		})
	}
}

func TestJSONExportCommand(t *testing.T) {
	cmd := buildExtendedTestCommand()
	cmd.Arguments = []Argument{
		&IntArgs{
			Name: "fooi",
		},
	}

	out, err := json.Marshal(cmd)
	require.NoError(t, err)

	expected := `{
		"name": "greet",
		"aliases": null,
		"usage": "Some app",
		"usageText": "app [first_arg] [second_arg]",
		"argsUsage": "",
		"version": "",
		"description": "Description of the application.",
		"defaultCommand": "",
		"category": "",
		"commands": [
		  {
			"name": "config",
			"aliases": [
			  "c"
			],
			"usage": "another usage test",
			"usageText": "",
			"argsUsage": "",
			"version": "",
			"description": "",
			"defaultCommand": "",
			"category": "",
			"commands": [
			  {
				"name": "sub-config",
				"aliases": [
				  "s",
				  "ss"
				],
				"usage": "another usage test",
				"usageText": "",
				"argsUsage": "",
				"version": "",
				"description": "",
				"defaultCommand": "",
				"category": "",
				"commands": null,
				"flags": [
				  {
					"name": "sub-flag",
					"category": "",
					"defaultText": "",
					"usage": "",
					"required": false,
					"hidden": false,
					"hideDefault": false,
					"local": false,
					"defaultValue": "",
					"aliases": [
					  "sub-fl",
					  "s"
					],
					"takesFileArg": false,
					"config": {
					  "TrimSpace": false
					},
					"onlyOnce": false,
					"validateDefaults" : false
				  },
				  {
					"name": "sub-command-flag",
					"category": "",
					"defaultText": "",
					"usage": "some usage text",
					"required": false,
					"hidden": false,
					"hideDefault": false,
					"local": false,
					"defaultValue": false,
					"aliases": [
					  "s"
					],
					"takesFileArg": false,
					"config": {
					  "Count": null
					},
					"onlyOnce": false,
					"validateDefaults" : false
				  }
				],
				"hideHelp": false,
				"hideHelpCommand": false,
				"hideVersion": false,
				"hidden": false,
				"authors": null,
				"copyright": "",
				"metadata": null,
				"sliceFlagSeparator": "",
				"disableSliceFlagSeparator": false,
				"useShortOptionHandling": false,
				"suggest": false,
				"allowExtFlags": false,
				"skipFlagParsing": false,
				"prefixMatchCommands": false,
				"mutuallyExclusiveFlags": null,
				"arguments": null,
				"readArgsFromStdin": false
			  }
			],
			"flags": [
			  {
				"name": "flag",
				"category": "",
				"defaultText": "",
				"usage": "",
				"required": false,
				"hidden": false,
				"hideDefault": false,
				"local": false,
				"defaultValue": "",
				"aliases": [
				  "fl",
				  "f"
				],
				"takesFileArg": true,
				"config": {
				  "TrimSpace": false
				},
				"onlyOnce": false,
				"validateDefaults" : false
			  },
			  {
				"name": "another-flag",
				"category": "",
				"defaultText": "",
				"usage": "another usage text",
				"required": false,
				"hidden": false,
				"hideDefault": false,
				"local": false,
				"defaultValue": false,
				"aliases": [
				  "b"
				],
				"takesFileArg": false,
				"config": {
				  "Count": null
				},
				"onlyOnce": false,
				"validateDefaults" : false
			  }
			],
			"hideHelp": false,
			"hideHelpCommand": false,
			"hideVersion": false,
			"hidden": false,
			"authors": null,
			"copyright": "",
			"metadata": null,
			"sliceFlagSeparator": "",
			"disableSliceFlagSeparator": false,
			"useShortOptionHandling": false,
			"suggest": false,
			"allowExtFlags": false,
			"skipFlagParsing": false,
			"prefixMatchCommands": false,
			"mutuallyExclusiveFlags": null,
			"arguments": null,
			"readArgsFromStdin": false
		  },
		  {
			"name": "info",
			"aliases": [
			  "i",
			  "in"
			],
			"usage": "retrieve generic information",
			"usageText": "",
			"argsUsage": "",
			"version": "",
			"description": "",
			"defaultCommand": "",
			"category": "",
			"commands": null,
			"flags": null,
			"hideHelp": false,
			"hideHelpCommand": false,
			"hideVersion": false,
			"hidden": false,
			"authors": null,
			"copyright": "",
			"metadata": null,
			"sliceFlagSeparator": "",
			"disableSliceFlagSeparator": false,
			"useShortOptionHandling": false,
			"suggest": false,
			"allowExtFlags": false,
			"skipFlagParsing": false,
			"prefixMatchCommands": false,
			"mutuallyExclusiveFlags": null,
			"arguments": null,
			"readArgsFromStdin": false
		  },
		  {
			"name": "some-command",
			"aliases": null,
			"usage": "",
			"usageText": "",
			"argsUsage": "",
			"version": "",
			"description": "",
			"defaultCommand": "",
			"category": "",
			"commands": null,
			"flags": null,
			"hideHelp": false,
			"hideHelpCommand": false,
			"hideVersion": false,
			"hidden": false,
			"authors": null,
			"copyright": "",
			"metadata": null,
			"sliceFlagSeparator": "",
			"disableSliceFlagSeparator": false,
			"useShortOptionHandling": false,
			"suggest": false,
			"allowExtFlags": false,
			"skipFlagParsing": false,
			"prefixMatchCommands": false,
			"mutuallyExclusiveFlags": null,
			"arguments": null,
			"readArgsFromStdin": false
		  },
		  {
			"name": "hidden-command",
			"aliases": null,
			"usage": "",
			"usageText": "",
			"argsUsage": "",
			"version": "",
			"description": "",
			"defaultCommand": "",
			"category": "",
			"commands": null,
			"flags": [
			  {
				"name": "completable",
				"category": "",
				"defaultText": "",
				"usage": "",
				"required": false,
				"hidden": false,
				"hideDefault": false,
				"local": false,
				"defaultValue": false,
				"aliases": null,
				"takesFileArg": false,
				"config": {
				  "Count": null
				},
				"onlyOnce": false,
				"validateDefaults": false
			  }
			],
			"hideHelp": false,
			"hideHelpCommand": false,
			"hideVersion": false,
			"hidden": true,
			"authors": null,
			"copyright": "",
			"metadata": null,
			"sliceFlagSeparator": "",
			"disableSliceFlagSeparator": false,
			"useShortOptionHandling": false,
			"suggest": false,
			"allowExtFlags": false,
			"skipFlagParsing": false,
			"prefixMatchCommands": false,
			"mutuallyExclusiveFlags": null,
			"arguments": null,
			"readArgsFromStdin": false
		  },
		  {
			"name": "usage",
			"aliases": [
			  "u"
			],
			"usage": "standard usage text",
			"usageText": "\nUsage for the usage text\n- formatted:  Based on the specified ConfigMap and summon secrets.yml\n- list:       Inspect the environment for a specific process running on a Pod\n- for_effect: Compare 'namespace' environment with 'local'\n\n` + "```\\nfunc() { ... }\\n```" + `\n\nShould be a part of the same code block\n",
			"argsUsage": "",
			"version": "",
			"description": "",
			"defaultCommand": "",
			"category": "",
			"commands": [
			  {
				"name": "sub-usage",
				"aliases": [
				  "su"
				],
				"usage": "standard usage text",
				"usageText": "Single line of UsageText",
				"argsUsage": "",
				"version": "",
				"description": "",
				"defaultCommand": "",
				"category": "",
				"commands": null,
				"flags": [
				  {
					"name": "sub-command-flag",
					"category": "",
					"defaultText": "",
					"usage": "some usage text",
					"required": false,
					"hidden": false,
					"hideDefault": false,
					"local": false,
					"defaultValue": false,
					"aliases": [
					  "s"
					],
					"takesFileArg": false,
					"config": {
					  "Count": null
					},
					"onlyOnce": false,
					"validateDefaults" : false
				  }
				],
				"hideHelp": false,
				"hideHelpCommand": false,
				"hideVersion": false,
				"hidden": false,
				"authors": null,
				"copyright": "",
				"metadata": null,
				"sliceFlagSeparator": "",
				"disableSliceFlagSeparator": false,
				"useShortOptionHandling": false,
				"suggest": false,
				"allowExtFlags": false,
				"skipFlagParsing": false,
				"prefixMatchCommands": false,
				"mutuallyExclusiveFlags": null,
				"arguments": null,
				"readArgsFromStdin": false
			  }
			],
			"flags": [
			  {
				"name": "flag",
				"category": "",
				"defaultText": "",
				"usage": "",
				"required": false,
				"hidden": false,
				"hideDefault": false,
				"local": false,
				"defaultValue": "",
				"aliases": [
				  "fl",
				  "f"
				],
				"takesFileArg": true,
				"config": {
				  "TrimSpace": false
				},
				"onlyOnce": false,
				"validateDefaults" : false
			  },
			  {
				"name": "another-flag",
				"category": "",
				"defaultText": "",
				"usage": "another usage text",
				"required": false,
				"hidden": false,
				"hideDefault": false,
				"local": false,
				"defaultValue": false,
				"aliases": [
				  "b"
				],
				"takesFileArg": false,
				"config": {
				  "Count": null
				},
				"onlyOnce": false,
				"validateDefaults" : false
			  }
			],
			"hideHelp": false,
			"hideHelpCommand": false,
			"hideVersion": false,
			"hidden": false,
			"authors": null,
			"copyright": "",
			"metadata": null,
			"sliceFlagSeparator": "",
			"disableSliceFlagSeparator": false,
			"useShortOptionHandling": false,
			"suggest": false,
			"allowExtFlags": false,
			"skipFlagParsing": false,
			"prefixMatchCommands": false,
			"mutuallyExclusiveFlags": null,
			"arguments": null,
			"readArgsFromStdin": false
		  }
		],
		"flags": [
		  {
			"name": "socket",
			"category": "",
			"defaultText": "",
			"usage": "some 'usage' text",
			"required": false,
			"hidden": false,
			"hideDefault": false,
			"local": false,
			"defaultValue": "value",
			"aliases": [
			  "s"
			],
			"takesFileArg": true,
			"config": {
			  "TrimSpace": false
			},
			"onlyOnce": false,
			"validateDefaults" : false
		  },
		  {
			"name": "flag",
			"category": "",
			"defaultText": "",
			"usage": "",
			"required": false,
			"hidden": false,
			"hideDefault": false,
			"local": false,
			"defaultValue": "",
			"aliases": [
			  "fl",
			  "f"
			],
			"takesFileArg": false,
			"config": {
			  "TrimSpace": false
			},
			"onlyOnce": false,
			"validateDefaults" : false
		  },
		  {
			"name": "another-flag",
			"category": "",
			"defaultText": "",
			"usage": "another usage text",
			"required": false,
			"hidden": false,
			"hideDefault": false,
			"local": false,
			"defaultValue": false,
			"aliases": [
			  "b"
			],
			"takesFileArg": false,
			"config": {
			  "Count": null
			},
			"onlyOnce": false,
			"validateDefaults" : false
		  },
		  {
			"name": "hidden-flag",
			"category": "",
			"defaultText": "",
			"usage": "",
			"required": false,
			"hidden": true,
			"hideDefault": false,
			"local": false,
			"defaultValue": false,
			"aliases": null,
			"takesFileArg": false,
			"config": {
			  "Count": null
			},
			"onlyOnce": false,
			"validateDefaults" : false
		  }
		],
		"hideHelp": false,
		"hideHelpCommand": false,
		"hideVersion": false,
		"hidden": false,
		"authors": [
		  "Harrison <harrison@lolwut.example.com>",
		  {
			"Name": "Oliver Allen",
			"Address": "oliver@toyshop.com"
		  }
		],
		"copyright": "",
		"metadata": null,
		"sliceFlagSeparator": "",
		"disableSliceFlagSeparator": false,
		"useShortOptionHandling": false,
		"suggest": false,
		"allowExtFlags": false,
		"skipFlagParsing": false,
		"prefixMatchCommands": false,
		"mutuallyExclusiveFlags": null,
		"arguments": [
		  {
			"name": "fooi",
			"value": 0,
			"usageText": "",
			"minTimes": 0,
			"maxTimes": 0,
			"config": {
			  "Base": 0
			}
		  }
		],
		"readArgsFromStdin": false
	  }
`
	assert.JSONEq(t, expected, string(out))
}

func TestCommand_ExclusiveFlagsWithAfter(t *testing.T) {
	var called bool
	cmd := &Command{
		Name: "bar",
		MutuallyExclusiveFlags: []MutuallyExclusiveFlags{
			{
				Category: "cat1",
				Flags: [][]Flag{
					{
						&StringFlag{
							Name: "foo",
						},
					},
					{
						&StringFlag{
							Name: "foo2",
						},
					},
				},
			},
		},
		After: func(ctx context.Context, cmd *Command) error {
			called = true
			return nil
		},
	}

	require.Error(t, cmd.Run(buildTestContext(t), []string{
		"bar",
		"--foo", "v1",
		"--foo2", "v2",
	}))
	require.True(t, called)
}
