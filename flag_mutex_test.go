package cli

import (
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
						&BoolFlag{
							Name: "q",
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
			errStr: "option i cannot be set along with option t",
		},
		{
			name:   "set both flags second member",
			args:   []string{"--i", "11", "--q"},
			errStr: "option i cannot be set along with option q",
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
			errStr:   "option i cannot be set along with option t",
			required: true,
		},
		{
			name:     "required both set second member",
			args:     []string{"--i", "11", "--q"},
			errStr:   "option i cannot be set along with option q",
			required: true,
		},
		{
			name:     "required both set second member both groups",
			args:     []string{"--s", "value", "--q"},
			errStr:   "option s cannot be set along with option q",
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
				assert.Contains(t, err.Error(), test.errStr)
			default:
				t.Errorf("got invalid error type %T", err)
			}
		})
	}
}

func TestMutuallyExclusiveFlags_PropagateStringer(t *testing.T) {
	customStringer := func(f Flag) string {
		return "custom:" + f.Names()[0]
	}

	grp := MutuallyExclusiveFlags{
		Stringer: customStringer,
		Flags: [][]Flag{
			{
				&StringFlag{Name: "foo"},
				&BoolWithInverseFlag{Name: "bar"},
			},
			{
				&Int64Flag{Name: "baz"},
			},
		},
	}

	grp.propagateStringer()

	assert.Equal(t, "custom:foo", grp.Flags[0][0].String())
	assert.Contains(t, grp.Flags[0][1].String(), "custom:bar")
	assert.Equal(t, "custom:baz", grp.Flags[1][0].String())
}

func TestMutuallyExclusiveFlags_PropagateStringerNil(t *testing.T) {
	grp := MutuallyExclusiveFlags{
		Flags: [][]Flag{
			{
				&StringFlag{Name: "foo"},
			},
		},
	}

	// should not panic and should leave flags using the default FlagStringer
	grp.propagateStringer()

	assert.NotEqual(t, "", grp.Flags[0][0].String())
}
