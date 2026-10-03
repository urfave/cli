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
	for _, sourceKind := range []string{"env", "file"} {
		for _, sourceValue := range []string{"", "false", "true"} {
			for _, override := range []bool{false, true} {
				t.Run(sourceKind+"/"+sourceValue+"/override="+map[bool]string{false: "no", true: "yes"}[override], func(t *testing.T) {
					var sources cli.ValueSourceChain
					if sourceKind == "env" {
						t.Setenv("CLI_BOOL_VALIDATOR_TEST", sourceValue)
						sources = cli.EnvVars("CLI_BOOL_VALIDATOR_TEST")
					} else {
						path := filepath.Join(t.TempDir(), "enabled")
						if err := os.WriteFile(path, []byte(sourceValue), 0o600); err != nil {
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
					if override {
						args = append(args, "--enabled=true")
					}
					err := cmd.Run(context.Background(), args)
					wantError := sourceValue != "true" && !override
					if wantError {
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
	}
}
