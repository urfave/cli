package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newCommand() *Command {
	return &Command{
		MutuallyExclusiveFlags: []MutuallyExclusiveFlags{
			{
				Flags: [][]Flag{
					{
						&Int64Flag{
							Name: "i",
						},
						&StringFlag{
							Name:    "s",
							Sources: EnvVars("S_VAR"),
						},
						&BoolWithInverseFlag{
							Name: "b",
						},
					},
					{
						&Int64Flag{
							Name:    "t",
							Aliases: []string{"ai"},
							Sources: EnvVars("T_VAR"),
						},
					},
				},
			},
		},
	}
}

func TestFlagMutuallyExclusiveFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		errStr   string
		required bool
		envs     map[string]string
	}{
		{
			name: "simple",
		},
		{
			name: "set one flag",
			args: []string{"--i", "10"},
		},
		{
			name:   "set both flags",
			args:   []string{"--i", "11", "--ai", "12"},
			errStr: "option i cannot be set along with option ai",
		},
		{
			name:     "required none set",
			required: true,
			errStr:   "one of these flags needs to be provided",
		},
		{
			name:     "required one set",
			args:     []string{"--i", "10"},
			required: true,
		},
		{
			name:     "required both set",
			args:     []string{"--i", "11", "--ai", "12"},
			errStr:   "option i cannot be set along with option ai",
			required: true,
		},
		{
			name:     "set env var",
			required: true,
			envs: map[string]string{
				"S_VAR": "some",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.envs != nil {
				for k, v := range test.envs {
					t.Setenv(k, v)
				}
			}
			cmd := newCommand()
			cmd.MutuallyExclusiveFlags[0].Required = test.required

			err := cmd.Run(buildTestContext(t), append([]string{"foo"}, test.args...))
			if test.errStr == "" {
				assert.NoError(t, err)
				return
			}
			if err == nil {
				t.Error("Expected mutual exclusion error")
				return
			}

			switch err.(type) {
			case (*mutuallyExclusiveGroup), (*mutuallyExclusiveGroupRequiredFlag):
				if !strings.Contains(err.Error(), test.errStr) {
					t.Logf("Invalid error string %v", err)
				}
			default:
				t.Errorf("got invalid error type %T", err)
			}
		})
	}
}
