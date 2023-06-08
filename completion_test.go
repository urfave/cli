package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCompletionDisable(t *testing.T) {
	cmd := &Command{}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", completionCommandName})
	if err == nil {
		t.Error("Expected error for no help topic for completion")
	}
}

func TestCompletionEnable(t *testing.T) {
	cmd := &Command{
		EnableShellCompletion: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", completionCommandName})
	if err == nil || !strings.Contains(err.Error(), "no shell provided") {
		t.Errorf("expected no shell provided error instead got [%v]", err)
	}
}

func TestCompletionEnableDiffCommandName(t *testing.T) {
	cmd := &Command{
		EnableShellCompletion:      true,
		ShellCompletionCommandName: "junky",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", "junky"})
	if err == nil || !strings.Contains(err.Error(), "no shell provided") {
		t.Errorf("expected no shell provided error instead got [%v]", err)
	}
}

func TestCompletionShell(t *testing.T) {
	for k := range shellCompletions {
		out := &bytes.Buffer{}

		t.Run(k, func(t *testing.T) {
			cmd := &Command{
				EnableShellCompletion: true,
				Writer:                out,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)

			r := require.New(t)

			r.NoError(cmd.Run(ctx, []string{"foo", completionCommandName, k}))
			r.Containsf(
				k, out.String(),
				"Expected output to contain shell name %[1]q", k,
			)
		})
	}
}

func TestCompletionInvalidShell(t *testing.T) {
	cmd := &Command{
		EnableShellCompletion: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	err := cmd.Run(ctx, []string{"foo", completionCommandName, "junky-sheell"})
	if err == nil {
		t.Error("Expected error for invalid shell")
	}
}
