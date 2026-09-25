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
	r := clitest.Run(t.Context(), failing(personality.Agent(cli.POSIX), err))
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

func TestAgentKeepsBaseExitCodes(t *testing.T) {
	err := &cli.Error{Kind: cli.UnknownFlag, Flag: "x"}
	for _, base := range []*cli.Personality{cli.POSIX, personality.Git} {
		t.Run(base.Name, func(t *testing.T) {
			human := clitest.Run(t.Context(), failing(base, err))
			agent := clitest.Run(t.Context(), failing(personality.Agent(base), err))
			if human.Code != agent.Code {
				t.Errorf("exit code changed: %d for people, %d for agents", human.Code, agent.Code)
			}
			if !json.Valid([]byte(agent.Stderr)) {
				t.Errorf("agent stderr is not JSON: %q", agent.Stderr)
			}
		})
	}
}

func TestQuiet(t *testing.T) {
	err := &cli.Error{Kind: cli.UnknownFlag, Flag: "x"}
	r := clitest.Run(t.Context(), failing(personality.Quiet(personality.Git), err))
	if r.Stderr != "tool: unknown flag: x\n" {
		t.Errorf("stderr = %q", r.Stderr)
	}
	if r.Code != 129 {
		t.Errorf("code = %d, want Git's 129", r.Code)
	}
}

func TestAuto(t *testing.T) {
	tests := []struct {
		value string
		set   bool
		agent bool
	}{
		{set: false, agent: false},
		{value: "", set: true, agent: false},
		{value: "0", set: true, agent: false},
		{value: "1", set: true, agent: true},
		{value: "true", set: true, agent: true},
	}
	for _, tt := range tests {
		lookup := func(k string) (string, bool) {
			if k == personality.AgentEnv && tt.set {
				return tt.value, true
			}
			return "", false
		}
		p := personality.Auto(lookup, personality.Git)
		if got := p != personality.Git; got != tt.agent {
			t.Errorf("%s=%q (set %v): agent = %v, want %v", personality.AgentEnv, tt.value, tt.set, got, tt.agent)
		}
	}
}
