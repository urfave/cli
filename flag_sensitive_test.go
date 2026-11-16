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

func TestSensitiveFlagJSON(t *testing.T) {
	b, err := json.Marshal(&StringFlag{Name: "token", Value: "s3cret", Sensitive: true})
	require.NoError(t, err)
	assert.NotContains(t, string(b), "s3cret")
	assert.Contains(t, string(b), `"sensitive":true`)
	assert.Contains(t, string(b), `"name":"token"`)

	f := &StringFlag{Name: "user", Value: "bob"}
	b, err = json.Marshal(f)
	require.NoError(t, err)
	assert.Contains(t, string(b), `"defaultValue":"bob"`)
	assert.Equal(t, "bob", f.Value)
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
