package cli_test

import (
	"flag"
	"reflect"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestConstructedValues(t *testing.T) {
	tests := []struct {
		name     string
		newValue func() cli.Value
		defaults any
		input    string
		want     any
		next     string
		wantNext any
	}{
		{"string slice", func() cli.Value { return cli.NewStringSlice("old") }, []string{"old"}, "new, next", []string{"new", " next"}, "last", []string{"new", " next", "last"}},
		{"int slice", func() cli.Value { return cli.NewIntSlice(1) }, []int{1}, "2,3", []int{2, 3}, "4", []int{2, 3, 4}},
		{"int8 slice", func() cli.Value { return cli.NewInt8Slice(1) }, []int8{1}, "2,3", []int8{2, 3}, "4", []int8{2, 3, 4}},
		{"int16 slice", func() cli.Value { return cli.NewInt16Slice(1) }, []int16{1}, "2,3", []int16{2, 3}, "4", []int16{2, 3, 4}},
		{"int32 slice", func() cli.Value { return cli.NewInt32Slice(1) }, []int32{1}, "2,3", []int32{2, 3}, "4", []int32{2, 3, 4}},
		{"int64 slice", func() cli.Value { return cli.NewInt64Slice(1) }, []int64{1}, "2,3", []int64{2, 3}, "4", []int64{2, 3, 4}},
		{"uint slice", func() cli.Value { return cli.NewUintSlice(1) }, []uint{1}, "2,3", []uint{2, 3}, "4", []uint{2, 3, 4}},
		{"uint8 slice", func() cli.Value { return cli.NewUint8Slice(1) }, []uint8{1}, "2,3", []uint8{2, 3}, "4", []uint8{2, 3, 4}},
		{"uint16 slice", func() cli.Value { return cli.NewUint16Slice(1) }, []uint16{1}, "2,3", []uint16{2, 3}, "4", []uint16{2, 3, 4}},
		{"uint32 slice", func() cli.Value { return cli.NewUint32Slice(1) }, []uint32{1}, "2,3", []uint32{2, 3}, "4", []uint32{2, 3, 4}},
		{"uint64 slice", func() cli.Value { return cli.NewUint64Slice(1) }, []uint64{1}, "2,3", []uint64{2, 3}, "4", []uint64{2, 3, 4}},
		{"float slice", func() cli.Value { return cli.NewFloatSlice(1) }, []float64{1}, "2.5,3.5", []float64{2.5, 3.5}, "4.5", []float64{2.5, 3.5, 4.5}},
		{"float32 slice", func() cli.Value { return cli.NewFloat32Slice(1) }, []float32{1}, "2.5,3.5", []float32{2.5, 3.5}, "4.5", []float32{2.5, 3.5, 4.5}},
		{"float64 slice", func() cli.Value { return cli.NewFloat64Slice(1) }, []float64{1}, "2.5,3.5", []float64{2.5, 3.5}, "4.5", []float64{2.5, 3.5, 4.5}},
		{"string map", func() cli.Value { return cli.NewStringMap(map[string]string{"old": "value"}) }, map[string]string{"old": "value"}, "key=value,other= entry ", map[string]string{"key": "value", "other": " entry "}, "key=last", map[string]string{"key": "last", "other": " entry "}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Run("defaults", func(t *testing.T) {
				if got := test.newValue().Get(); !reflect.DeepEqual(got, test.defaults) {
					t.Fatalf("Get() = %#v, want %#v", got, test.defaults)
				}
			})
			t.Run("set", func(t *testing.T) {
				defer func() {
					if recovered := recover(); recovered != nil {
						t.Fatalf("Set panicked: %v", recovered)
					}
				}()
				value := test.newValue()
				flags := flag.NewFlagSet(test.name, flag.ContinueOnError)
				flags.Var(value, "value", "")
				if err := flags.Set("value", test.input); err != nil {
					t.Fatal(err)
				}
				if got := value.Get(); !reflect.DeepEqual(got, test.want) {
					t.Fatalf("Get() = %#v, want %#v", got, test.want)
				}
				if err := flags.Parse([]string{"-value", test.next}); err != nil {
					t.Fatal(err)
				}
				if got := value.Get(); !reflect.DeepEqual(got, test.wantNext) {
					t.Fatalf("Get() = %#v, want %#v", got, test.wantNext)
				}
			})
		})
	}
}

func TestConstructedValuesEmptyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		value    cli.Value
		defaults any
		input    string
		want     any
	}{
		{"nil slice", cli.NewStringSlice(), []string(nil), "value", []string{"value"}},
		{"empty slice", cli.NewStringSlice([]string{}...), []string{}, "value", []string{"value"}},
		{"nil map", cli.NewStringMap(nil), map[string]string(nil), "key=value", map[string]string{"key": "value"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.value.Get(); !reflect.DeepEqual(got, test.defaults) {
				t.Fatalf("Get() = %#v, want %#v", got, test.defaults)
			}
			if err := test.value.Set(test.input); err != nil {
				t.Fatal(err)
			}
			if got := test.value.Get(); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Get() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestConstructedValuesParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		value cli.Value
		input string
	}{
		{"int8 overflow", cli.NewInt8Slice(), "128"},
		{"negative uint", cli.NewUintSlice(), "-1"},
		{"invalid float", cli.NewFloatSlice(), "not-a-float"},
		{"missing map separator", cli.NewStringMap(nil), "key"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.value.Set(test.input); err == nil {
				t.Fatalf("Set(%q) unexpectedly succeeded", test.input)
			}
		})
	}
}

func TestConstructedValuesSerializedThenSet(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		value := cli.NewStringSlice("old")
		if err := value.Set(cli.NewStringSlice("new").Serialize()); err != nil {
			t.Fatal(err)
		}
		if err := value.Set("last"); err != nil {
			t.Fatal(err)
		}
		if got, want := value.Value(), []string{"new", "last"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("Value() = %#v, want %#v", got, want)
		}
	})
	t.Run("map", func(t *testing.T) {
		value := cli.NewStringMap(map[string]string{"old": "value"})
		if err := value.Set(cli.NewStringMap(map[string]string{"new": "value"}).Serialize()); err != nil {
			t.Fatal(err)
		}
		if err := value.Set("last=entry"); err != nil {
			t.Fatal(err)
		}
		if got, want := value.Value(), map[string]string{"new": "value", "last": "entry"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("Value() = %#v, want %#v", got, want)
		}
	})
}

func TestConstructedValuesDefaultStorage(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		defaults := []string{"old"}
		value := cli.NewStringSlice(defaults...)
		defaults[0] = "changed"
		if got, want := value.Value(), []string{"changed"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("Value() = %#v, want %#v", got, want)
		}
		if err := value.Set("new"); err != nil {
			t.Fatal(err)
		}
		if got := defaults[0]; got != "changed" {
			t.Fatalf("Set modified defaults: %q", got)
		}
	})
	t.Run("map", func(t *testing.T) {
		defaults := map[string]string{"old": "value"}
		value := cli.NewStringMap(defaults)
		defaults["old"] = "changed"
		if got, want := value.Value(), map[string]string{"old": "changed"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("Value() = %#v, want %#v", got, want)
		}
		if err := value.Set("new=entry"); err != nil {
			t.Fatal(err)
		}
		if got := defaults["old"]; got != "changed" {
			t.Fatalf("Set modified defaults: %q", got)
		}
		if _, ok := defaults["new"]; ok {
			t.Fatal("Set added an entry to defaults")
		}
	})
}
