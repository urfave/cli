package cli_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/urfave/cli/v4"
)

func TestErrorIs(t *testing.T) {
	cause := errors.New("boom")
	err := &cli.Error{Kind: cli.InvalidValue, Flag: "port", Value: "x", Err: cause}

	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{"same kind", err, cli.ErrInvalidValue, true},
		{"other kind", err, cli.ErrUnknownFlag, false},
		{"cause", err, cause, true},
		{"wrapped with %w", fmt.Errorf("setup: %w", err), cli.ErrInvalidValue, true},
		{"joined", errors.Join(errors.New("x"), err), cli.ErrInvalidValue, true},
		{"inside Exit", cli.Exit(err, 7), cli.ErrInvalidValue, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.target); got != tt.want {
				t.Errorf("errors.Is = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorAs(t *testing.T) {
	err := fmt.Errorf("setup: %w", &cli.Error{Kind: cli.UnknownFlag, Flag: "verbos", Choices: []string{"verbose"}})
	var e *cli.Error
	if !errors.As(err, &e) {
		t.Fatal("errors.As found no *cli.Error")
	}
	if e.Flag != "verbos" || e.Choices[0] != "verbose" {
		t.Errorf("fields lost: %+v", e)
	}
}

func TestEveryKindHasAMessage(t *testing.T) {
	for k := cli.Internal; k <= cli.Version; k++ {
		msg := (&cli.Error{Kind: k, Err: errors.New("cause")}).Error()
		if msg == "" || strings.HasPrefix(msg, "error.") {
			t.Errorf("kind %s has no English message, got %q", k, msg)
		}
	}
}

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		err  *cli.Error
		want string
	}{
		{&cli.Error{Kind: cli.UnknownFlag, Flag: "z"}, "unknown flag: z"},
		{&cli.Error{Kind: cli.TooManyArgs, Value: "a", Count: 1}, `unexpected argument "a"`},
		{&cli.Error{Kind: cli.TooManyArgs, Value: "a", Count: 3}, `3 unexpected arguments, starting with "a"`},
		{&cli.Error{Kind: cli.FlagConflict, Flag: "json", Choices: []string{"yaml", "text"}}, "flag json cannot be used with yaml, text"},
		{&cli.Error{Kind: cli.Internal, Err: errors.New("disk full")}, "disk full"},
	}
	for _, tt := range tests {
		if got := tt.err.Error(); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.err.Kind, got, tt.want)
		}
	}
}
