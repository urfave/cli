package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensitiveFlagHidesDefaultInHelp(t *testing.T) {
	f := &StringFlag{Name: "token", Value: "s3cret", Sensitive: true}
	assert.NotContains(t, f.String(), "s3cret")

	f = &StringFlag{Name: "token", Value: "s3cret", Sensitive: true, DefaultText: "from vault"}
	assert.Contains(t, f.String(), "from vault")

	f = &StringFlag{Name: "token", Value: "s3cret"}
	assert.Contains(t, f.String(), "s3cret")
}

func TestSensitiveFlagRedactsArgError(t *testing.T) {
	cmd := &Command{
		Name:  "app",
		Flags: []Flag{&IntFlag{Name: "pin", Sensitive: true}},
	}

	err := cmd.Run(buildTestContext(t), []string{"app", "--pin", "s3cret"})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "s3cret")
	assert.Contains(t, err.Error(), redactedText)
}

func TestSensitiveFlagRedactsSourceError(t *testing.T) {
	t.Setenv("APP_PIN", "s3cret")
	cmd := &Command{
		Name: "app",
		Flags: []Flag{&IntFlag{
			Name:      "pin",
			Sensitive: true,
			Sources:   EnvVars("APP_PIN"),
		}},
	}

	err := cmd.Run(buildTestContext(t), []string{"app"})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "s3cret")
	assert.Contains(t, err.Error(), "APP_PIN")
}

func TestSensitiveFlagRedactsItemsInErrors(t *testing.T) {
	tests := []struct {
		name string
		flag Flag
		arg  string
	}{
		{
			name: "map item without separator",
			flag: &StringMapFlag{Name: "f", Sensitive: true},
			arg:  "a=b,s3cret",
		},
		{
			name: "slice item that fails to parse",
			flag: &IntSliceFlag{Name: "f", Sensitive: true},
			arg:  "1, s3cret",
		},
		{
			name: "map value that fails to parse",
			flag: &StringMapFlag{Name: "f", Sensitive: true, Validator: func(m map[string]string) error {
				return fmt.Errorf("bad value %q", m["k"])
			}},
			arg: "k=s3cret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{Name: "app", Flags: []Flag{tt.flag}}
			err := cmd.Run(buildTestContext(t), []string{"app", "--f", tt.arg})
			require.Error(t, err)
			assert.NotContains(t, err.Error(), "s3cret")
		})
	}
}

func TestSensitiveFlagCommandJSON(t *testing.T) {
	token := &StringFlag{Name: "token", Value: "s3cret-root", Sensitive: true}
	user := &StringFlag{Name: "user", Value: "bob"}
	cmd := &Command{
		Name:  "app",
		Flags: []Flag{token, user},
		Commands: []*Command{{
			Name:  "sub",
			Flags: []Flag{&IntFlag{Name: "pin", Value: 4242, Sensitive: true}},
			MutuallyExclusiveFlags: []MutuallyExclusiveFlags{{
				Flags: [][]Flag{{&StringFlag{Name: "key", Value: "s3cret-group", Sensitive: true}}},
			}},
		}},
	}

	b, err := json.Marshal(cmd)
	require.NoError(t, err)
	out := string(b)
	assert.NotContains(t, out, "s3cret-root")
	assert.NotContains(t, out, "4242")
	assert.NotContains(t, out, "s3cret-group")
	assert.Contains(t, out, `"sensitive":true`)
	assert.Contains(t, out, `"defaultValue":"bob"`)

	// encoding must not change the command's own flags
	assert.Same(t, token, cmd.Flags[0])
	assert.Equal(t, "s3cret-root", token.Value)
}

// Masking is done by Command.MarshalJSON, so a sensitive flag encoded on its
// own still shows its default value, but never the value it was set to
func TestSensitiveFlagAloneJSON(t *testing.T) {
	f := &StringFlag{Name: "token", Value: "default-s3cret", Sensitive: true}
	cmd := &Command{Name: "app", Flags: []Flag{f}}
	require.NoError(t, cmd.Run(buildTestContext(t), []string{"app", "--token", "arg-s3cret"}))

	b, err := json.Marshal(f)
	require.NoError(t, err)
	assert.Contains(t, string(b), `"defaultValue":"default-s3cret"`)
	assert.NotContains(t, string(b), "arg-s3cret")
}

// A type that embeds a flag must keep its own fields when encoded, as it
// did before sensitive flags were added
func TestSensitiveFlagEmbeddedFlagJSON(t *testing.T) {
	type myFlag struct {
		StringFlag
		Extra string `json:"extra"`
	}

	b, err := json.Marshal(&myFlag{StringFlag: StringFlag{Name: "n"}, Extra: "kept"})
	require.NoError(t, err)
	assert.Contains(t, string(b), `"extra":"kept"`)
	assert.Contains(t, string(b), `"name":"n"`)
}

func TestSensitiveValues(t *testing.T) {
	t.Setenv("APP_TOKEN", "root-token")

	var got []string
	cmd := &Command{
		Name: "app",
		Flags: []Flag{
			&StringFlag{Name: "token", Sensitive: true, Sources: EnvVars("APP_TOKEN")},
			&StringFlag{Name: "password", Sensitive: true, Value: "default-password"},
			&StringFlag{Name: "unused", Sensitive: true},
			&StringFlag{Name: "user", Value: "bob"},
		},
		Commands: []*Command{{
			Name: "sub",
			Flags: []Flag{
				&StringSliceFlag{Name: "keys", Sensitive: true},
				&StringMapFlag{Name: "headers", Sensitive: true},
			},
			Action: func(_ context.Context, cmd *Command) error {
				got = cmd.SensitiveValues()
				return nil
			},
		}},
	}

	err := cmd.Run(buildTestContext(t), []string{
		"app", "sub",
		"--keys", "k1,root-token",
		"--headers", "Authorization=Bearer x",
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"default-password", "root-token", "Bearer x", "k1"}, got)
}

func TestSensitiveValuesNone(t *testing.T) {
	cmd := &Command{
		Name:  "app",
		Flags: []Flag{&StringFlag{Name: "user", Value: "bob"}},
	}
	assert.Empty(t, cmd.SensitiveValues())
}
