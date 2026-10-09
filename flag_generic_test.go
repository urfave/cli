package cli

import (
	"context"
	"errors"
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
)

type genericPayloadValue struct {
	text     string
	gets     int
	result   Value
	setError error
}

func (v *genericPayloadValue) String() string { return v.text }

func (v *genericPayloadValue) Set(s string) error {
	if v.setError != nil {
		return v.setError
	}
	v.text = s
	return nil
}

func (v *genericPayloadValue) Get() any {
	v.gets++
	if v.result != nil {
		return v.result
	}
	return v.text
}

var _ Value = (*genericPayloadValue)(nil)

func TestGenericFlagTypedCallbacks(t *testing.T) {
	for _, tc := range []struct {
		name string
		make func() (Value, Value)
	}{
		{"payload", func() (Value, Value) {
			v := &genericPayloadValue{text: "before"}
			return v, v
		}},
		{"standard Getter", func() (Value, Value) {
			fs := flag.NewFlagSet("example", flag.ContinueOnError)
			fs.String("text", "before", "")
			v := fs.Lookup("text").Value.(Value)
			return v, v
		}},
		{"self", func() (Value, Value) {
			v := &Parser{"before", "value"}
			return v, v
		}},
		{"delegated Value", func() (Value, Value) {
			result := &Parser{"delegated", "value"}
			return &genericPayloadValue{result: result}, result
		}},
	} {
		for _, mode := range []string{"default", "set", "action"} {
			t.Run(tc.name+"/"+mode, func(t *testing.T) {
				value, want := tc.make()
				calls := 0
				observe := func(got Value) error {
					calls++
					require.Same(t, want, got)
					return nil
				}
				fl := &GenericFlag{Name: "value", Value: value}
				if mode == "action" {
					fl.Action = func(_ context.Context, _ *Command, got Value) error { return observe(got) }
				} else {
					fl.Validator = observe
					fl.ValidateDefaults = mode == "default"
				}
				require.NotPanics(t, func() {
					require.NoError(t, fl.PreParse())
					if mode != "default" {
						require.NoError(t, fl.Set("value", "hello,world"))
					}
					if mode == "action" {
						require.NoError(t, fl.RunAction(buildTestContext(t), &Command{}))
					}
				})
				require.Equal(t, 1, calls)
				if counted, ok := value.(*genericPayloadValue); ok {
					require.Equal(t, 1, counted.gets)
				}
			})
		}
	}
}

func TestGenericFlagPayloadRetrieval(t *testing.T) {
	v := &genericPayloadValue{text: "before"}
	fl := &GenericFlag{Name: "value", Aliases: []string{"v"}, Value: v}
	cmd := &Command{Flags: []Flag{fl}}

	// Before parsing, Get returns the declared default without invoking its Getter.
	require.Same(t, v, fl.Get())
	require.Same(t, v, cmd.Value("value"))
	require.Same(t, v, cmd.Generic("v"))
	require.Zero(t, v.gets)

	require.NoError(t, fl.Set("value", "after"))
	for _, name := range []string{"value", "v"} {
		before := v.gets
		require.Same(t, v, cmd.Generic(name))
		require.Equal(t, before+1, v.gets)
		require.Equal(t, "after", cmd.Value(name))
	}
	require.Equal(t, "after", fl.Get())

	// A Getter that already returns another Value must retain that result.
	other := &Parser{"other", "value"}
	v.result = other
	before := v.gets
	require.Same(t, other, cmd.Generic("value"))
	require.Equal(t, before+1, v.gets)
	require.Same(t, other, cmd.Value("value"))
	require.Same(t, other, fl.Get())
}

func TestGenericFlagPayloadLookup(t *testing.T) {
	v := &genericPayloadValue{text: "parent"}
	fl := &GenericFlag{Name: "value", Aliases: []string{"v"}, Value: v}
	invalidCalls := 0
	parent := &Command{
		Flags: []Flag{fl},
		InvalidFlagAccessHandler: func(_ context.Context, _ *Command, name string) {
			invalidCalls++
			require.Equal(t, "missing", name)
		},
	}
	child := &Command{parent: parent}
	require.NoError(t, fl.PreParse())
	require.Same(t, v, child.Generic("value"))
	require.Same(t, v, child.Generic("v"))
	require.Equal(t, 2, v.gets)

	// A local flag shadows its ancestor even when it is not a GenericFlag.
	child.Flags = []Flag{&StringFlag{Name: "value", Aliases: []string{"v"}, Value: "child"}}
	require.Nil(t, child.Generic("value"))
	require.Nil(t, child.Generic("v"))
	require.Equal(t, 2, v.gets)
	require.Zero(t, invalidCalls)
	require.Nil(t, child.Generic("missing"))
	require.Equal(t, 1, invalidCalls)
}

func TestGenericFlagNilValue(t *testing.T) {
	fl := &GenericFlag{Name: "value"}
	cmd := &Command{Flags: []Flag{fl}}
	require.Nil(t, cmd.Generic("value"))
	require.Nil(t, cmd.Value("value"))
	calls := 0
	fl.Validator = func(v Value) error {
		calls++
		require.Nil(t, v)
		return nil
	}
	fl.ValidateDefaults = true
	fl.Action = func(_ context.Context, _ *Command, v Value) error {
		calls++
		require.Nil(t, v)
		return nil
	}
	require.NotPanics(t, func() {
		require.NoError(t, fl.PreParse())
		require.NoError(t, fl.Set("value", "ignored"))
		require.NoError(t, fl.RunAction(buildTestContext(t), cmd))
	})
	require.Equal(t, 3, calls)
	require.Nil(t, cmd.Generic("value"))
	require.Nil(t, cmd.Value("value"))
	require.Nil(t, fl.Get())
}

func TestGenericFlagPayloadCallbackErrors(t *testing.T) {
	for _, mode := range []string{"default", "set", "action"} {
		t.Run(mode, func(t *testing.T) {
			failure := errors.New("callback rejected value")
			v := &genericPayloadValue{}
			fl := &GenericFlag{Name: "value", Value: v}
			if mode == "action" {
				fl.Action = func(context.Context, *Command, Value) error { return failure }
			} else {
				fl.Validator = func(Value) error { return failure }
				fl.ValidateDefaults = mode == "default"
			}
			require.NotPanics(t, func() {
				if mode == "default" {
					require.ErrorIs(t, fl.PreParse(), failure)
					return
				}
				require.NoError(t, fl.PreParse())
				if mode == "set" {
					require.ErrorIs(t, fl.Set("value", "text"), failure)
				} else {
					require.ErrorIs(t, fl.RunAction(buildTestContext(t), &Command{}), failure)
				}
			})
		})
	}

	t.Run("Set error precedes validation", func(t *testing.T) {
		failure := errors.New("cannot parse value")
		v := &genericPayloadValue{setError: failure}
		fl := &GenericFlag{Name: "value", Value: v, Validator: func(Value) error {
			t.Fatal("validator ran after Set failed")
			return nil
		}}
		require.ErrorIs(t, fl.Set("value", "text"), failure)
		require.Zero(t, v.gets)
	})
}

func TestGenericFlagPayloadCommandRun(t *testing.T) {
	for _, args := range [][]string{{"app"}, {"app", "--value", "hello"}} {
		t.Run(args[len(args)-1], func(t *testing.T) {
			v := &genericPayloadValue{}
			var calls []string
			cmd := &Command{Flags: []Flag{&GenericFlag{
				Name: "value", Value: v,
				Validator: func(got Value) error {
					require.Same(t, v, got)
					calls = append(calls, "validator")
					return nil
				},
				Action: func(_ context.Context, _ *Command, got Value) error {
					require.Same(t, v, got)
					calls = append(calls, "action")
					return nil
				},
			}}}
			require.NotPanics(t, func() { require.NoError(t, cmd.Run(buildTestContext(t), args)) })
			if len(args) == 1 {
				require.Empty(t, calls)
			} else {
				require.Equal(t, []string{"validator", "action"}, calls)
				require.Equal(t, "hello", v.text)
			}
		})
	}
}
