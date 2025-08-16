package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand_StopOnNthArg(t *testing.T) {
	tests := []struct {
		name         string
		stopOnNthArg *int
		testArgs     []string
		expectedArgs []string
		expectedFlag string
		expectedBool bool
	}{
		{
			name:         "nil StopOnNthArg - normal parsing",
			stopOnNthArg: nil,
			testArgs:     []string{"cmd", "--flag", "value", "arg1", "--bool", "arg2"},
			expectedArgs: []string{"arg1", "arg2"},
			expectedFlag: "value",
			expectedBool: true,
		},
		{
			name:         "stop after 0 args - all become args",
			stopOnNthArg: intPtr(0),
			testArgs:     []string{"cmd", "--flag", "value", "arg1", "--bool", "arg2"},
			expectedArgs: []string{"--flag", "value", "arg1", "--bool", "arg2"},
			expectedFlag: "",
			expectedBool: false,
		},
		{
			name:         "stop after 1 arg",
			stopOnNthArg: intPtr(1),
			testArgs:     []string{"cmd", "--flag", "value", "arg1", "--bool", "arg2"},
			expectedArgs: []string{"arg1", "--bool", "arg2"},
			expectedFlag: "value",
			expectedBool: false,
		},
		{
			name:         "stop after 2 args",
			stopOnNthArg: intPtr(2),
			testArgs:     []string{"cmd", "--flag", "value", "arg1", "arg2", "--bool", "arg3"},
			expectedArgs: []string{"arg1", "arg2", "--bool", "arg3"},
			expectedFlag: "value",
			expectedBool: false,
		},
		{
			name:         "mixed flags and args - stop after 1",
			stopOnNthArg: intPtr(1),
			testArgs:     []string{"cmd", "--flag", "value", "--bool", "arg1", "--flag2", "value2"},
			expectedArgs: []string{"arg1", "--flag2", "value2"},
			expectedFlag: "value",
			expectedBool: true,
		},
		{
			name:         "args before flags - stop after 1",
			stopOnNthArg: intPtr(1),
			testArgs:     []string{"cmd", "arg1", "--flag", "value", "--bool"},
			expectedArgs: []string{"arg1", "--flag", "value", "--bool"},
			expectedFlag: "",
			expectedBool: false,
		},
		{
			name:         "ssh command example",
			stopOnNthArg: intPtr(1),
			testArgs:     []string{"ssh", "machine-name", "ls", "-la"},
			expectedArgs: []string{"machine-name", "ls", "-la"},
			expectedFlag: "",
			expectedBool: false,
		},
		{
			name:         "with double dash terminator",
			stopOnNthArg: intPtr(1),
			testArgs:     []string{"cmd", "--flag", "value", "--", "arg1", "--not-a-flag"},
			expectedArgs: []string{"arg1", "--not-a-flag"},
			expectedFlag: "value",
			expectedBool: false,
		},
		{
			name:         "stop after large number of args",
			stopOnNthArg: intPtr(100),
			testArgs:     []string{"cmd", "--flag", "value", "arg1", "arg2", "--bool"},
			expectedArgs: []string{"arg1", "arg2"},
			expectedFlag: "value",
			expectedBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args Args
			var flagValue string
			var boolValue bool

			cmd := &Command{
				Name:         "test",
				StopOnNthArg: tt.stopOnNthArg,
				Flags: []Flag{
					&StringFlag{Name: "flag", Destination: &flagValue},
					&StringFlag{Name: "flag2"},
					&BoolFlag{Name: "bool", Destination: &boolValue},
				},
				Action: func(_ context.Context, cmd *Command) error {
					args = cmd.Args()
					return nil
				},
			}

			require.NoError(t, cmd.Run(buildTestContext(t), tt.testArgs))
			assert.Equal(t, tt.expectedArgs, args.Slice())
			assert.Equal(t, tt.expectedFlag, flagValue)
			assert.Equal(t, tt.expectedBool, boolValue)
		})
	}
}

func TestCommand_StopOnNthArg_WithSubcommands(t *testing.T) {
	tests := []struct {
		name               string
		parentStopOnNthArg *int
		subStopOnNthArg    *int
		testArgs           []string
		expectedParentArgs []string
		expectedSubArgs    []string
		expectedSubFlag    string
	}{
		{
			name:               "parent normal, subcommand stops after 0",
			parentStopOnNthArg: nil,
			subStopOnNthArg:    intPtr(0),
			testArgs:           []string{"parent", "sub", "--subflag", "value", "subarg", "--not-parsed"},
			expectedParentArgs: []string{},
			expectedSubArgs:    []string{"--subflag", "value", "subarg", "--not-parsed"},
			expectedSubFlag:    "",
		},
		{
			name:               "parent normal, subcommand stops after 1",
			parentStopOnNthArg: nil,
			subStopOnNthArg:    intPtr(1),
			testArgs:           []string{"parent", "sub", "--subflag", "value", "subarg", "--not-parsed"},
			expectedParentArgs: []string{},
			expectedSubArgs:    []string{"subarg", "--not-parsed"},
			expectedSubFlag:    "value",
		},
		{
			name:               "parent normal, subcommand stops after 2",
			parentStopOnNthArg: nil,
			subStopOnNthArg:    intPtr(2),
			testArgs:           []string{"parent", "sub", "--subflag", "value", "subarg1", "subarg2", "--not-parsed"},
			expectedParentArgs: []string{},
			expectedSubArgs:    []string{"subarg1", "subarg2", "--not-parsed"},
			expectedSubFlag:    "value",
		},
		{
			name:               "parent normal, subcommand never stops (high StopOnNthArg)",
			parentStopOnNthArg: nil,
			subStopOnNthArg:    intPtr(100),
			testArgs:           []string{"parent", "sub", "--subflag", "value1", "arg1", "arg2", "--subflag", "value2"},
			expectedParentArgs: []string{},
			expectedSubArgs:    []string{"arg1", "arg2"},
			expectedSubFlag:    "value2", // Should parse the second --subflag since we never hit the stop limit
		},
		{
			// Meaningless, but okay.
			name:               "parent stops after 1, subcommand stops after 1",
			parentStopOnNthArg: intPtr(1),
			subStopOnNthArg:    intPtr(1),
			testArgs:           []string{"parent", "sub", "--subflag", "value", "subarg", "--not-parsed"},
			expectedParentArgs: []string{},
			expectedSubArgs:    []string{"subarg", "--not-parsed"},
			expectedSubFlag:    "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parentArgs, subArgs Args
			var subFlagValue string
			subCalled := false

			subCmd := &Command{
				Name:         "sub",
				StopOnNthArg: tt.subStopOnNthArg,
				Flags: []Flag{
					&StringFlag{Name: "subflag", Destination: &subFlagValue},
				},
				Action: func(_ context.Context, cmd *Command) error {
					subCalled = true
					subArgs = cmd.Args()
					return nil
				},
			}

			parentCmd := &Command{
				Name:         "parent",
				StopOnNthArg: tt.parentStopOnNthArg,
				Commands:     []*Command{subCmd},
				Flags: []Flag{
					&StringFlag{Name: "parentflag"},
				},
				Action: func(_ context.Context, cmd *Command) error {
					parentArgs = cmd.Args()
					return nil
				},
			}

			err := parentCmd.Run(buildTestContext(t), tt.testArgs)

			require.NoError(t, err)

			if tt.expectedSubArgs != nil {
				assert.True(t, subCalled, "subcommand should have been called")
				if len(tt.expectedSubArgs) > 0 {
					haveNonEmptySubArgsSlice := subArgs != nil && subArgs.Slice() != nil && len(subArgs.Slice()) > 0
					assert.True(t, haveNonEmptySubArgsSlice, "subargs.Slice is not nil")
					if haveNonEmptySubArgsSlice {
						assert.Equal(t, tt.expectedSubArgs, subArgs.Slice())
					}
				} else {
					assert.True(t, subArgs == nil || subArgs.Slice() == nil || len(subArgs.Slice()) == 0, "subargs.Slice is not nil")
				}
				assert.Equal(t, tt.expectedSubFlag, subFlagValue)
			} else {
				assert.False(t, subCalled, "subcommand should not have been called")
				assert.Equal(t, tt.expectedParentArgs, parentArgs.Slice())
			}
		})
	}
}

func TestCommand_StopOnNthArg_EdgeCases(t *testing.T) {
	t.Run("negative StopOnNthArg returns error", func(t *testing.T) {
		cmd := &Command{
			Name:         "test",
			StopOnNthArg: intPtr(-1),
			Action: func(_ context.Context, cmd *Command) error {
				return nil
			},
		}

		// Negative value should return an error
		err := cmd.Run(buildTestContext(t), []string{"cmd", "arg1"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "StopOnNthArg must be non-negative")
	})

	t.Run("zero StopOnNthArg with no args", func(t *testing.T) {
		var args Args
		var flagValue string
		cmd := &Command{
			Name:         "test",
			StopOnNthArg: intPtr(0),
			Flags: []Flag{
				&StringFlag{Name: "flag", Destination: &flagValue},
			},
			Action: func(_ context.Context, cmd *Command) error {
				args = cmd.Args()
				return nil
			},
		}

		// All flags should become args
		require.NoError(t, cmd.Run(buildTestContext(t), []string{"cmd", "--flag", "value"}))
		assert.Equal(t, []string{"--flag", "value"}, args.Slice())
		assert.Equal(t, "", flagValue)
	})

	t.Run("StopOnNthArg with only flags", func(t *testing.T) {
		var args Args
		var flagValue string
		var boolValue bool
		cmd := &Command{
			Name:         "test",
			StopOnNthArg: intPtr(1),
			Flags: []Flag{
				&StringFlag{Name: "flag", Destination: &flagValue},
				&BoolFlag{Name: "bool", Destination: &boolValue},
			},
			Action: func(_ context.Context, cmd *Command) error {
				args = cmd.Args()
				return nil
			},
		}

		// Should parse all flags since no args are encountered
		require.NoError(t, cmd.Run(buildTestContext(t), []string{"cmd", "--flag", "value", "--bool"}))
		assert.Equal(t, []string{}, args.Slice())
		assert.Equal(t, "value", flagValue)
		assert.True(t, boolValue)
	})
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
