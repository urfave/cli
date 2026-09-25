package personality_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/urfave/cli/v4"
	"github.com/urfave/cli/v4/clitest"
	"github.com/urfave/cli/v4/personality"
)

func failing(p *cli.Personality, err error) *cli.Command {
	return &cli.Command{
		Name:        "tool",
		Personality: p,
		Action:      func(context.Context, *cli.Command) error { return err },
	}
}

func TestGit(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantErr  string
		wantCode int
	}{
		{
			name:     "unknown command exits 1",
			err:      &cli.Error{Kind: cli.UnknownCommand, Value: "stauts", Choices: []string{"status"}},
			wantErr:  "tool: unknown command \"stauts\"\nThe most similar command is \"status\".\nSee 'tool --help'.\n",
			wantCode: 1,
		},
		{
			name:     "other usage errors exit 129",
			err:      &cli.Error{Kind: cli.UnknownFlag, Flag: "x"},
			wantErr:  "tool: unknown flag: x\nSee 'tool --help'.\n",
			wantCode: 129,
		},
		{
			name:     "action error exits 1",
			err:      errors.New("not a repository"),
			wantErr:  "tool: not a repository\n",
			wantCode: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := clitest.Run(t.Context(), failing(personality.Git, tt.err))
			if r.Stderr != tt.wantErr {
				t.Errorf("stderr = %q, want %q", r.Stderr, tt.wantErr)
			}
			if r.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", r.Code, tt.wantCode)
			}
		})
	}
}

func TestAgent(t *testing.T) {
	err := errors.Join(
		&cli.Error{Kind: cli.InvalidValue, Command: "tool", Flag: "port", Value: "abc", Err: errors.New("not a number")},
		errors.New("disk full"),
	)
	r := clitest.Run(t.Context(), failing(personality.Agent, err))
	if r.Code != 2 {
		t.Errorf("code = %d, want 2", r.Code)
	}

	var got struct {
		ExitCode int `json:"exit_code"`
		Errors   []struct {
			Kind, Message, Command, Flag, Value string
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(r.Stderr), &got); err != nil {
		t.Fatalf("stderr is not JSON: %v\n%s", err, r.Stderr)
	}
	if got.ExitCode != 2 || len(got.Errors) != 2 {
		t.Fatalf("got %+v", got)
	}
	first := got.Errors[0]
	if first.Kind != "invalid_value" || first.Flag != "port" || first.Value != "abc" ||
		first.Message != `invalid value "abc" for flag port: not a number` {
		t.Errorf("first error = %+v", first)
	}
	if got.Errors[1].Kind != "error" || got.Errors[1].Message != "disk full" {
		t.Errorf("second error = %+v", got.Errors[1])
	}
}
