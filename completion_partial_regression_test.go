package cli

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
)

func partialCompletionCommand(output *bytes.Buffer, ran *bool) *Command {
	command := func(name string) *Command {
		return &Command{
			Name: name,
			Flags: []Flag{
				&StringFlag{Name: "param-" + name + "-one"},
				&StringFlag{Name: "param-" + name + "-two"},
				&StringFlag{Name: "param-" + name + "-hidden", Hidden: true},
				&BoolFlag{Name: "unrelated-" + name},
			},
			Action: func(context.Context, *Command) error { *ran = true; return nil },
		}
	}
	root, child, nested := command("app"), command("sub"), command("nested")
	root.Writer, root.ErrWriter = output, output
	root.EnableShellCompletion = true
	root.Commands = []*Command{child}
	child.Commands = []*Command{nested}
	return root
}

func TestPartialFlagCompletionAfterPositionalArgument(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	for _, scope := range []struct {
		name string
		path []string
	}{
		{"app", nil},
		{"sub", []string{"sub"}},
		{"nested", []string{"sub", "nested"}},
	} {
		for _, positional := range [][]string{nil, {"value"}, {"first", "second"}} {
			t.Run(scope.name+"/"+strings.Join(positional, "_"), func(t *testing.T) {
				var output bytes.Buffer
				ran := false
				cmd := partialCompletionCommand(&output, &ran)
				args := append([]string{"app"}, scope.path...)
				args = append(args, positional...)
				args = append(args, "--pa", completionFlag)
				os.Args = args
				if err := cmd.Run(context.Background(), args); err != nil {
					t.Fatal(err)
				}
				want := "--param-" + scope.name + "-one\n--param-" + scope.name + "-two\n"
				if got := output.String(); got != want {
					t.Errorf("completion for %q = %q, want %q", args, got, want)
				}
				if ran {
					t.Error("completion executed the command action")
				}
			})
		}
	}
}

func TestPartialFlagCompletionAfterDoubleDash(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	for _, path := range [][]string{nil, {"sub"}, {"sub", "nested"}} {
		t.Run(strings.Join(path, "/"), func(t *testing.T) {
			var output bytes.Buffer
			ran := false
			cmd := partialCompletionCommand(&output, &ran)
			args := append([]string{"app"}, path...)
			args = append(args, "--", "--pa", completionFlag)
			os.Args = args
			if err := cmd.Run(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			if output.Len() != 0 || ran {
				t.Errorf("completion after --: output=%q action=%v", output.String(), ran)
			}
		})
	}
}

func TestPartialFlagCustomCompletionArgs(t *testing.T) {
	var output bytes.Buffer
	ran := false
	cmd := partialCompletionCommand(&output, &ran)
	var got []string
	cmd.Commands[0].ShellComplete = func(_ context.Context, cmd *Command) {
		got = append([]string(nil), cmd.Args().Slice()...)
	}
	args := []string{"app", "sub", "value", "--pa", completionFlag}
	if err := cmd.Run(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	if want := []string{"value", "--pa"}; !reflect.DeepEqual(got, want) {
		t.Errorf("custom completion args = %q, want %q", got, want)
	}
	if ran {
		t.Error("custom completion executed the action")
	}
}

func TestPartialFlagCompletionDisabled(t *testing.T) {
	var output bytes.Buffer
	ran := false
	cmd := partialCompletionCommand(&output, &ran)
	cmd.EnableShellCompletion = false
	args := []string{"app", "sub", "value", "--pa", completionFlag}
	if err := cmd.Run(context.Background(), args); err == nil {
		t.Fatal("disabled completion accepted an undefined flag")
	}
	if ran || strings.Contains(output.String(), "--param-sub-one\n--param-sub-two\n") {
		t.Fatalf("disabled completion: output=%q action=%v", output.String(), ran)
	}
}
