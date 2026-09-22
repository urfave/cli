package cli

import (
	"bytes"
	"context"
	"testing"
)

func TestCompletionIncludesPersistentFlagsAfterPositionalArgument(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "before positional argument",
			args: []string{"app", "add", "--conf", completionFlag},
			want: "--config:path to config file\n",
		},
		{
			name: "after positional argument",
			args: []string{"app", "add", "value", "--conf", completionFlag},
			want: "--config:path to config file\n",
		},
		{
			name: "short prefix",
			args: []string{"app", "add", "-", completionFlag},
			want: "--config:path to config file\n--help:show help\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			cmd := &Command{
				Name:                  "app",
				EnableShellCompletion: true,
				Writer:                &output,
				Flags: []Flag{
					&StringFlag{Name: "config", Usage: "path to config file"},
				},
				Commands: []*Command{
					{
						Name: "add",
						Arguments: []Argument{
							&StringArg{Name: "arg1"},
						},
					},
				},
			}

			if err := cmd.Run(context.Background(), test.args); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != test.want {
				t.Errorf("completion for %q = %q, want %q", test.args, got, test.want)
			}
		})
	}
}
