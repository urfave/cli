package cli_test

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	cli "github.com/urfave/cli/v3"
)

type embeddedSliceFlag struct {
	cli.StringSliceFlag
	calls  []string
	reject bool
}

func (f *embeddedSliceFlag) Set(name, value string) error {
	f.calls = append(f.calls, name+"="+value)
	if f.reject {
		return errors.New("external Set rejected the value")
	}
	return f.StringSliceFlag.Set(name, strings.ToUpper(value))
}

type embeddedMapFlag struct {
	cli.StringMapFlag
	calls  []string
	reject bool
}

func (f *embeddedMapFlag) Set(name, value string) error {
	f.calls = append(f.calls, name+"="+value)
	if f.reject {
		return errors.New("external Set rejected the value")
	}
	return f.StringMapFlag.Set(name, strings.ToUpper(value))
}

type embeddedScalarFlag struct {
	cli.StringFlag
	calls int
}

func (f *embeddedScalarFlag) Set(_, _ string) error {
	f.calls++
	return errors.New("external Set rejected the value")
}

func embeddedFlagCommand(flag cli.Flag, action func(context.Context, *cli.Command) error) *cli.Command {
	return &cli.Command{
		Name: "app", Writer: io.Discard, ErrWriter: io.Discard,
		Flags: []cli.Flag{flag}, Action: action,
	}
}

func TestEmbeddedFlagSetReject(t *testing.T) {
	for _, kind := range []string{"slice", "map"} {
		t.Run(kind, func(t *testing.T) {
			var flag cli.Flag
			var calls *[]string
			if kind == "slice" {
				f := &embeddedSliceFlag{StringSliceFlag: cli.StringSliceFlag{Name: "items"}, reject: true}
				flag, calls = f, &f.calls
			} else {
				f := &embeddedMapFlag{StringMapFlag: cli.StringMapFlag{Name: "items"}, reject: true}
				flag, calls = f, &f.calls
			}
			ran := false
			cmd := embeddedFlagCommand(flag, func(context.Context, *cli.Command) error { ran = true; return nil })
			err := cmd.Run(context.Background(), []string{"app", "--items", "a=b"})
			t.Logf("Set calls=%v; action=%v; error=%v", *calls, ran, err)
			if err == nil || !strings.Contains(err.Error(), "external Set rejected") || ran || len(*calls) != 1 {
				t.Fatalf("custom Set rejection must stop the action; calls=%v action=%v err=%v", *calls, ran, err)
			}
		})
	}
}

func TestEmbeddedFlagSetNormalize(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		f := &embeddedSliceFlag{StringSliceFlag: cli.StringSliceFlag{Name: "items", Aliases: []string{"i"}}}
		cmd := embeddedFlagCommand(f, nil)
		err := cmd.Run(context.Background(), []string{"app", "-i", "first", "--items", "second"})
		t.Logf("Set calls=%v; value=%v; error=%v", f.calls, f.Get(), err)
		if err != nil || len(f.calls) != 2 || !reflect.DeepEqual(f.Get(), []string{"FIRST", "SECOND"}) {
			t.Fatalf("custom Set must normalize both occurrences; calls=%v value=%v err=%v", f.calls, f.Get(), err)
		}
	})
	t.Run("map", func(t *testing.T) {
		f := &embeddedMapFlag{StringMapFlag: cli.StringMapFlag{Name: "items"}}
		cmd := embeddedFlagCommand(f, nil)
		err := cmd.Run(context.Background(), []string{"app", "--items", "first=a", "--items", "second=b"})
		t.Logf("Set calls=%v; value=%v; error=%v", f.calls, f.Get(), err)
		if err != nil || len(f.calls) != 2 || !reflect.DeepEqual(f.Get(), map[string]string{"FIRST": "A", "SECOND": "B"}) {
			t.Fatalf("custom Set must normalize both occurrences; calls=%v value=%v err=%v", f.calls, f.Get(), err)
		}
	})
}

func TestEmbeddedFlagSetControls(t *testing.T) {
	t.Run("scalar_override_rejects", func(t *testing.T) {
		f := &embeddedScalarFlag{StringFlag: cli.StringFlag{Name: "item"}}
		err := embeddedFlagCommand(f, nil).Run(context.Background(), []string{"app", "--item", "first"})
		if err == nil || !strings.Contains(err.Error(), "external Set rejected") || f.calls != 1 {
			t.Fatalf("scalar Set dispatch: calls=%d err=%v", f.calls, err)
		}
	})
	t.Run("programmatic_command_set_dispatches", func(t *testing.T) {
		f := &embeddedSliceFlag{StringSliceFlag: cli.StringSliceFlag{Name: "items"}}
		cmd := embeddedFlagCommand(f, func(_ context.Context, cmd *cli.Command) error { return cmd.Set("items", "first") })
		err := cmd.Run(context.Background(), []string{"app"})
		if err != nil || len(f.calls) != 1 || !reflect.DeepEqual(f.Get(), []string{"FIRST"}) {
			t.Fatalf("programmatic Set dispatch: calls=%v value=%v err=%v", f.calls, f.Get(), err)
		}
	})
	t.Run("embedded_without_set_override_accumulates", func(t *testing.T) {
		f := &struct{ cli.StringSliceFlag }{cli.StringSliceFlag{Name: "items"}}
		err := embeddedFlagCommand(f, nil).Run(context.Background(), []string{"app", "--items", "first", "--items", "second"})
		if err != nil || !reflect.DeepEqual(f.Get(), []string{"first", "second"}) {
			t.Fatalf("embedded slice accumulation: value=%v err=%v", f.Get(), err)
		}
	})
	t.Run("native_slice_accumulates", func(t *testing.T) {
		f := &cli.StringSliceFlag{Name: "items"}
		err := embeddedFlagCommand(f, nil).Run(context.Background(), []string{"app", "--items", "first", "--items", "second"})
		if err != nil || !reflect.DeepEqual(f.Get(), []string{"first", "second"}) {
			t.Fatalf("native slice accumulation: value=%v err=%v", f.Get(), err)
		}
	})
}
