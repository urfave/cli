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

// Encoding as JSON shows the default value of a sensitive flag, which is in
// the source or binary anyway, but never the value it was set to
func TestSensitiveFlagJSONHasNoSetValue(t *testing.T) {
	gen := &sensitiveGeneric{s: "default-gen"}
	cmd := &Command{
		Name: "app",
		Flags: []Flag{
			&StringFlag{Name: "str", Value: "default-str", Sensitive: true},
			&StringSliceFlag{Name: "sl", Value: []string{"default-sl"}, Sensitive: true},
			&StringMapFlag{Name: "map", Value: map[string]string{"k": "default-map"}, Sensitive: true},
			&GenericFlag{Name: "gen", Value: gen, Sensitive: true},
			&TextFlag{Name: "txt", Value: &sensitiveText{s: "default-txt"}, Sensitive: true},
		},
	}
	require.NoError(t, cmd.Run(buildTestContext(t), []string{
		"app",
		"--str", "set-s3cret-str",
		"--sl", "set-s3cret-sl",
		"--map", "k=set-s3cret-map",
		"--gen", "set-s3cret-gen",
		"--txt", "set-s3cret-txt",
	}))

	b, err := json.Marshal(cmd)
	require.NoError(t, err)
	out := string(b)
	assert.NotContains(t, out, "set-s3cret")
	for _, def := range []string{"default-str", "default-sl", "default-map", "default-gen", "default-txt"} {
		assert.Contains(t, out, def)
	}

	for _, f := range cmd.Flags {
		b, err := json.Marshal(f)
		require.NoError(t, err)
		assert.NotContains(t, string(b), "set-s3cret")
	}
}

// A sensitive GenericFlag or TextFlag parses into a copy of its default
// value, so the parsed value is read from the command or Destination and
// the caller's default object is left as it was
func TestSensitiveFlagPointerValueIsCopied(t *testing.T) {
	gen := &sensitiveGeneric{s: "default-gen"}
	txt := &sensitiveText{s: "default-txt"}
	var dest TextMarshalUnmarshaler

	cmd := &Command{
		Name: "app",
		Flags: []Flag{
			&GenericFlag{Name: "gen", Value: gen, Sensitive: true},
			&TextFlag{Name: "txt", Value: txt, Destination: &dest, Sensitive: true},
		},
	}
	require.NoError(t, cmd.Run(buildTestContext(t), []string{"app", "--gen", "set-gen", "--txt", "set-txt"}))

	assert.Equal(t, "default-gen", gen.s)
	assert.Equal(t, "default-txt", txt.s)
	assert.Equal(t, "set-gen", cmd.Generic("gen").String())
	assert.Equal(t, &sensitiveText{s: "set-txt"}, cmd.Text("txt"))
	assert.Equal(t, &sensitiveText{s: "set-txt"}, dest)
	assert.Contains(t, cmd.SensitiveValues(), "set-gen")
}

// A flag that is not sensitive still parses into its default value object
func TestNotSensitiveFlagPointerValueIsShared(t *testing.T) {
	gen := &sensitiveGeneric{s: "default-gen"}
	cmd := &Command{
		Name:  "app",
		Flags: []Flag{&GenericFlag{Name: "gen", Value: gen}},
	}
	require.NoError(t, cmd.Run(buildTestContext(t), []string{"app", "--gen", "set-gen"}))
	assert.Equal(t, "set-gen", gen.s)
}

// A type that embeds a Command must keep its own fields when encoded
func TestSensitiveFlagEmbeddedCommandJSON(t *testing.T) {
	type myCommand struct {
		Command
		Extra string `json:"extra"`
	}

	b, err := json.Marshal(&myCommand{Command: Command{Name: "app"}, Extra: "kept"})
	require.NoError(t, err)
	assert.Contains(t, string(b), `"name":"app"`)
	assert.Contains(t, string(b), `"extra":"kept"`)
}

type sensitiveGeneric struct{ s string }

func (g *sensitiveGeneric) Set(s string) error { g.s = s; return nil }
func (g *sensitiveGeneric) String() string     { return g.s }
func (g *sensitiveGeneric) Get() any           { return g }
func (g *sensitiveGeneric) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.s)
}

type sensitiveText struct{ s string }

func (x *sensitiveText) MarshalText() ([]byte, error) { return []byte(x.s), nil }
func (x *sensitiveText) UnmarshalText(b []byte) error { x.s = string(b); return nil }

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
