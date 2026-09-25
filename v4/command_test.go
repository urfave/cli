package cli_test

import (
	"context"
	"errors"
	"testing"

	"github.com/urfave/cli/v4"
	"github.com/urfave/cli/v4/clitest"
	"github.com/urfave/cli/v4/i18n/locale/de"
)

func hello() *cli.Command {
	return &cli.Command{
		Name: "hello",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_, err := cmd.Out().Write([]byte("Hello\n"))
			return err
		},
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		cmd      func() *cli.Command
		args     []string
		wantOut  string
		wantErr  string
		wantCode int
		wantKind *cli.Error
	}{
		{
			name:    "success",
			cmd:     hello,
			wantOut: "Hello\n",
		},
		{
			name:     "one extra argument",
			cmd:      hello,
			args:     []string{"extra"},
			wantErr:  "hello: unexpected argument \"extra\"\nTry 'hello --help' for more information.\n",
			wantCode: 2,
			wantKind: cli.ErrTooManyArgs,
		},
		{
			name:     "plural",
			cmd:      hello,
			args:     []string{"a", "b"},
			wantErr:  "hello: 2 unexpected arguments, starting with \"a\"\nTry 'hello --help' for more information.\n",
			wantCode: 2,
		},
		{
			name: "German",
			cmd: func() *cli.Command {
				c := hello()
				c.Localizer = cli.SelectLocalizer(env("LANG", "de_DE.UTF-8"), de.Catalog)
				return c
			},
			args:     []string{"extra"},
			wantErr:  "hello: unerwartetes Argument „extra“\n„hello --help“ gibt weitere Informationen aus.\n",
			wantCode: 2,
		},
		{
			name: "unlimited args",
			cmd: func() *cli.Command {
				c := hello()
				c.MaxArgs = -1
				return c
			},
			args:    []string{"a", "b", "c"},
			wantOut: "Hello\n",
		},
		{
			name: "action error",
			cmd: func() *cli.Command {
				return &cli.Command{Name: "app", Action: func(context.Context, *cli.Command) error {
					return errors.New("disk full")
				}}
			},
			wantErr:  "app: disk full\n",
			wantCode: 1,
		},
		{
			name: "action chooses exit code",
			cmd: func() *cli.Command {
				return &cli.Command{Name: "app", Action: func(context.Context, *cli.Command) error {
					return cli.Exit(errors.New("not found"), 4)
				}}
			},
			wantErr:  "app: not found\n",
			wantCode: 4,
		},
		{
			name: "joined errors",
			cmd: func() *cli.Command {
				return &cli.Command{Name: "app", Action: func(context.Context, *cli.Command) error {
					return errors.Join(
						&cli.Error{Kind: cli.RequiredFlag, Flag: "name"},
						&cli.Error{Kind: cli.RequiredFlag, Flag: "port"},
					)
				}}
			},
			wantErr:  "app: required flag not set: name\napp: required flag not set: port\nTry 'app --help' for more information.\n",
			wantCode: 2,
		},
		{
			name: "help exits 0",
			cmd: func() *cli.Command {
				return &cli.Command{Name: "app", Action: func(context.Context, *cli.Command) error {
					return cli.ErrHelp
				}}
			},
			wantCode: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := clitest.Run(t.Context(), tt.cmd(), tt.args...)
			if r.Stdout != tt.wantOut {
				t.Errorf("stdout = %q, want %q", r.Stdout, tt.wantOut)
			}
			if r.Stderr != tt.wantErr {
				t.Errorf("stderr = %q, want %q", r.Stderr, tt.wantErr)
			}
			if r.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", r.Code, tt.wantCode)
			}
			if tt.wantKind != nil && !errors.Is(r.Err, tt.wantKind) {
				t.Errorf("err = %v, want kind %s", r.Err, tt.wantKind.Kind)
			}
		})
	}
}
