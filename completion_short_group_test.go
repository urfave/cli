package cli

import (
	"context"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestShortGroupCustomCompletionArgs(t *testing.T) {
	for _, path := range [][]string{nil, {"sub"}, {"sub", "nested"}} {
		for _, token := range []string{"-az", "-ab", "-a한", "-b"} {
			t.Run(strings.Join(path, "/")+"/"+token, func(t *testing.T) {
				ran := false
				var got []string
				newCommand := func(name string) *Command {
					return &Command{
						Name:          name,
						Flags:         []Flag{&BoolFlag{Name: "a"}, &StringFlag{Name: "b"}, &StringFlag{Name: "한"}},
						Action:        func(context.Context, *Command) error { ran = true; return nil },
						ShellComplete: func(_ context.Context, c *Command) { got = c.Args().Slice() },
					}
				}
				root, sub, nested := newCommand("app"), newCommand("sub"), newCommand("nested")
				root.Commands = []*Command{sub}
				sub.Commands = []*Command{nested}
				root.EnableShellCompletion = true
				root.UseShortOptionHandling = true
				root.Writer = io.Discard
				root.ErrWriter = io.Discard
				args := append([]string{"app"}, path...)
				args = append(args, "value", token, completionFlag)
				if err := root.Run(t.Context(), args); err != nil {
					t.Fatal(err)
				}
				want := []string{"value", token}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("completion args=%q want=%q", got, want)
				}
				if ran {
					t.Error("completion ran command action")
				}
			})
		}
	}
}

func TestShortGroupParsingControls(t *testing.T) {
	for _, tt := range []struct {
		name      string
		args      []string
		complete  bool
		wantErr   bool
		wantValue string
	}{
		{name: "unknown", args: []string{"-az"}, wantErr: true},
		{name: "missing-value", args: []string{"-ab"}, wantErr: true},
		{name: "missing-unicode-value", args: []string{"-a한"}, wantErr: true},
		{name: "explicit-value", args: []string{"-ab=given"}, wantValue: "given"},
		{name: "separate-value", args: []string{"-ab", "given"}, wantValue: "given"},
		{name: "explicit-empty", args: []string{"-ab="}},
		{name: "completion-explicit-value", args: []string{"-ab=given"}, complete: true, wantValue: "given"},
		{name: "completion-explicit-empty", args: []string{"-ab="}, complete: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			action, completion := false, false
			var got []string
			cmd := &Command{
				Name:                  "app",
				Writer:                io.Discard,
				ErrWriter:             io.Discard,
				EnableShellCompletion: tt.complete, UseShortOptionHandling: true,
				Flags:         []Flag{&BoolFlag{Name: "a"}, &StringFlag{Name: "b"}, &StringFlag{Name: "한"}},
				Action:        func(context.Context, *Command) error { action = true; return nil },
				ShellComplete: func(_ context.Context, c *Command) { completion = true; got = c.Args().Slice() },
			}
			args := append([]string{"app", "value"}, tt.args...)
			if tt.complete {
				args = append(args, completionFlag)
			}
			err := cmd.Run(t.Context(), args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error=%v wantErr=%v", err, tt.wantErr)
			}
			if tt.wantErr {
				if action || completion {
					t.Fatal("invalid invocation ran callback")
				}
				return
			}
			if action == tt.complete || completion != tt.complete {
				t.Fatalf("action=%t completion=%t", action, completion)
			}
			if !cmd.Bool("a") || cmd.String("b") != tt.wantValue {
				t.Fatalf("a=%t b=%q", cmd.Bool("a"), cmd.String("b"))
			}
			if tt.complete && !reflect.DeepEqual(got, []string{"value"}) {
				t.Fatalf("complete group retained as argument: %q", got)
			}
		})
	}
}
