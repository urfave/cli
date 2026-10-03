package cli_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cli "github.com/urfave/cli/v3"
)

func TestBoolFlagEmptySourceRunsValidator(t *testing.T) {
	tests := []struct {
		name        string
		sourceKind  string
		sourceValue string
		override    bool
		wantError   bool
	}{
		{name: "env/empty", sourceKind: "env", sourceValue: "", wantError: true},
		{name: "env/empty/override", sourceKind: "env", sourceValue: "", override: true},
		{name: "env/false", sourceKind: "env", sourceValue: "false", wantError: true},
		{name: "env/false/override", sourceKind: "env", sourceValue: "false", override: true},
		{name: "env/true", sourceKind: "env", sourceValue: "true"},
		{name: "env/true/override", sourceKind: "env", sourceValue: "true", override: true},
		{name: "file/empty", sourceKind: "file", sourceValue: "", wantError: true},
		{name: "file/empty/override", sourceKind: "file", sourceValue: "", override: true},
		{name: "file/false", sourceKind: "file", sourceValue: "false", wantError: true},
		{name: "file/false/override", sourceKind: "file", sourceValue: "false", override: true},
		{name: "file/true", sourceKind: "file", sourceValue: "true"},
		{name: "file/true/override", sourceKind: "file", sourceValue: "true", override: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sources cli.ValueSourceChain
			if tt.sourceKind == "env" {
				t.Setenv("CLI_BOOL_VALIDATOR_TEST", tt.sourceValue)
				sources = cli.EnvVars("CLI_BOOL_VALIDATOR_TEST")
			} else {
				path := filepath.Join(t.TempDir(), "enabled")
				if err := os.WriteFile(path, []byte(tt.sourceValue), 0o600); err != nil {
					t.Fatal(err)
				}
				sources = cli.Files(path)
			}
			calls := 0
			actionRan := false
			flagActionRan := false
			destination := true
			rejected := errors.New("enabled must be true")
			flag := &cli.BoolFlag{
				Name: "enabled", Value: true, Destination: &destination, Sources: sources,
				Validator: func(value bool) error {
					calls++
					if !value {
						return rejected
					}
					return nil
				},
				Action: func(context.Context, *cli.Command, bool) error { flagActionRan = true; return nil },
			}
			cmd := &cli.Command{
				Name: "test", Writer: io.Discard, ErrWriter: io.Discard,
				Flags:  []cli.Flag{flag},
				Action: func(context.Context, *cli.Command) error { actionRan = true; return nil },
			}
			args := []string{"test"}
			if tt.override {
				args = append(args, "--enabled=true")
			}
			err := cmd.Run(context.Background(), args)
			if tt.wantError {
				if err == nil || !strings.Contains(err.Error(), rejected.Error()) {
					t.Errorf("Run() error = %v, want validation error", err)
				}
				if actionRan || flagActionRan {
					t.Errorf("actions ran after rejected source: command=%v flag=%v", actionRan, flagActionRan)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !actionRan || !flagActionRan || !destination || !flag.IsSet() {
					t.Errorf("accepted source: action=%v flagAction=%v destination=%v IsSet=%v", actionRan, flagActionRan, destination, flag.IsSet())
				}
			}
			if calls != 1 {
				t.Errorf("validator calls = %d, want 1", calls)
			}
		})
	}
}
