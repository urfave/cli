package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompletionDisable(t *testing.T) {
	a := &App{}
	err := a.Run([]string{"foo", completionCommandName})
	if err == nil {
		t.Error("Expected error for no help topic for completion")
	}
}

func TestCompletionEnable(t *testing.T) {
	a := &App{
		EnableShellCompletion: true,
	}
	err := a.Run([]string{"foo", completionCommandName})
	if err == nil || !strings.Contains(err.Error(), "no shell provided") {
		t.Errorf("expected no shell provided error instead got [%v]", err)
	}
}

func TestCompletionEnableDiffCommandName(t *testing.T) {
	defer func() {
		completionCommand.Name = completionCommandName
	}()

	a := &App{
		EnableShellCompletion:      true,
		ShellCompletionCommandName: "junky",
	}
	err := a.Run([]string{"foo", "junky"})
	if err == nil || !strings.Contains(err.Error(), "no shell provided") {
		t.Errorf("expected no shell provided error instead got [%v]", err)
	}
}

func TestCompletionShell(t *testing.T) {
	for k := range shellCompletions {
		var b bytes.Buffer
		t.Run(k, func(t *testing.T) {
			a := &App{
				EnableShellCompletion: true,
				Writer:                &b,
			}
			err := a.Run([]string{"foo", completionCommandName, k})
			if err != nil {
				t.Error(err)
			}
		})
		output := b.String()
		if !strings.Contains(output, k) {
			t.Errorf("Expected output to contain shell name %v", output)
		}
	}
}

func TestCompletionInvalidShell(t *testing.T) {
	a := &App{
		EnableShellCompletion: true,
	}
	err := a.Run([]string{"foo", completionCommandName, "junky-sheell"})
	if err == nil {
		t.Error("Expected error for invalid shell")
	}
}
