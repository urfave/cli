package altsrc

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v2"
)

type testApplyInputSource struct {
	Flag               FlagInputSourceExtension
	FlagName           string
	FlagSetName        string
	Expected           string
	ContextValueString string
	ContextValue       flag.Value
	EnvVarValue        string
	EnvVarName         string
	SourcePath         string
	MapValue           interface{}
}

type racyInputSource struct {
	*MapInputSource
}

func (ris *racyInputSource) isSet(name string) bool {
	if _, ok := ris.MapInputSource.valueMap[name]; ok {
		ris.MapInputSource.valueMap[name] = bogus{0}
	}
	return true
}

func TestGenericApplyInputSourceValue_Alias(t *testing.T) {
	v := &Parser{"abc", "def"}
	tis := testApplyInputSource{
		Flag:     NewGenericFlag(&cli.GenericFlag{Name: "test", Aliases: []string{"test_alias"}, Value: &Parser{}}),
		FlagName: "test_alias",
		MapValue: v,
	}
	c := runTest(t, tis)
	expect(t, v, c.Generic("test_alias"))

	c = runRacyTest(t, tis)
	refute(t, v, c.Generic("test_alias"))
}

func TestGenericApplyInputSourceValue(t *testing.T) {
	v := &Parser{"abc", "def"}
	tis := testApplyInputSource{
		Flag:     NewGenericFlag(&cli.GenericFlag{Name: "test", Value: &Parser{}}),
		FlagName: "test",
		MapValue: v,
	}
	c := runTest(t, tis)
	expect(t, v, c.Generic("test"))

	c = runRacyTest(t, tis)
	refute(t, v, c.Generic("test"))
}

func TestGenericApplyInputSourceMethodContextSet(t *testing.T) {
	p := &Parser{"abc", "def"}
	tis := testApplyInputSource{
		Flag:               NewGenericFlag(&cli.GenericFlag{Name: "test", Value: &Parser{}}),
		FlagName:           "test",
		MapValue:           &Parser{"efg", "hig"},
		ContextValueString: p.String(),
	}
	c := runTest(t, tis)
	expect(t, p, c.Generic("test"))

	c = runRacyTest(t, tis)
	refute(t, p, c.Generic("test"))
}

func TestGenericApplyInputSourceMethodEnvVarSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag: NewGenericFlag(&cli.GenericFlag{
			Name:    "test",
			Value:   &Parser{},
			EnvVars: []string{"TEST"},
		}),
		FlagName:    "test",
		MapValue:    &Parser{"efg", "hij"},
		EnvVarName:  "TEST",
		EnvVarValue: "abc,def",
	}
	c := runTest(t, tis)
	expect(t, &Parser{"abc", "def"}, c.Generic("test"))

	c = runRacyTest(t, tis)
	refute(t, &Parser{"abc", "def"}, c.Generic("test"))
}

func TestStringSliceApplyInputSourceValue_Alias(t *testing.T) {
	dest := cli.NewStringSlice()
	tis := testApplyInputSource{
		Flag:     NewStringSliceFlag(&cli.StringSliceFlag{Name: "test", Aliases: []string{"test_alias"}, Destination: dest}),
		FlagName: "test_alias",
		MapValue: []interface{}{"hello", "world"},
	}
	c := runTest(t, tis)
	expect(t, c.StringSlice("test_alias"), []string{"hello", "world"})
	expect(t, dest.Value(), []string{"hello", "world"})

	// reset dest
	dest = cli.NewStringSlice()
	tis = testApplyInputSource{
		Flag:     NewStringSliceFlag(&cli.StringSliceFlag{Name: "test", Aliases: []string{"test_alias"}, Destination: dest}),
		FlagName: "test_alias",
		MapValue: []interface{}{"hello", "world"},
	}
	c = runRacyTest(t, tis)
	refute(t, c.StringSlice("test_alias"), []string{"hello", "world"})
	refute(t, dest.Value(), []string{"hello", "world"})
}

func TestStringSliceApplyInputSourceValue(t *testing.T) {
	dest := cli.NewStringSlice()
	tis := testApplyInputSource{
		Flag:     NewStringSliceFlag(&cli.StringSliceFlag{Name: "test", Destination: dest}),
		FlagName: "test",
		MapValue: []interface{}{"hello", "world"},
	}
	c := runTest(t, tis)
	expect(t, c.StringSlice("test"), []string{"hello", "world"})
	expect(t, dest.Value(), []string{"hello", "world"})

	// reset dest
	dest = cli.NewStringSlice()
	tis = testApplyInputSource{
		Flag:     NewStringSliceFlag(&cli.StringSliceFlag{Name: "test", Destination: dest}),
		FlagName: "test",
		MapValue: []interface{}{"hello", "world"},
	}
	c = runRacyTest(t, tis)
	refute(t, c.StringSlice("test"), []string{"hello", "world"})
	refute(t, dest.Value(), []string{"hello", "world"})
}

func TestStringSliceApplyInputSourceMethodContextSet(t *testing.T) {
	dest := cli.NewStringSlice()
	c := runTest(t, testApplyInputSource{
		Flag:               NewStringSliceFlag(&cli.StringSliceFlag{Name: "test", Destination: dest}),
		FlagName:           "test",
		MapValue:           []interface{}{"hello", "world"},
		ContextValueString: "ohno",
	})
	expect(t, c.StringSlice("test"), []string{"ohno"})
	expect(t, dest.Value(), []string{"ohno"})
}

func TestStringSliceApplyInputSourceMethodEnvVarSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:        NewStringSliceFlag(&cli.StringSliceFlag{Name: "test", EnvVars: []string{"TEST"}}),
		FlagName:    "test",
		MapValue:    []interface{}{"hello", "world"},
		EnvVarName:  "TEST",
		EnvVarValue: "oh,no",
	}
	c := runTest(t, tis)
	expect(t, c.StringSlice("test"), []string{"oh", "no"})

	c = runRacyTest(t, tis)
	refute(t, c.StringSlice("test"), []string{"oh", "no"})
}

func TestIntSliceApplyInputSourceValue_Alias(t *testing.T) {
	dest := cli.NewIntSlice()
	tis := testApplyInputSource{
		Flag:     NewIntSliceFlag(&cli.IntSliceFlag{Name: "test", Aliases: []string{"test_alias"}, Destination: dest}),
		FlagName: "test_alias",
		MapValue: []interface{}{1, 2},
	}
	c := runTest(t, tis)
	expect(t, c.IntSlice("test_alias"), []int{1, 2})
	expect(t, dest.Value(), []int{1, 2})

	dest = cli.NewIntSlice()
	tis = testApplyInputSource{
		Flag:     NewIntSliceFlag(&cli.IntSliceFlag{Name: "test", Aliases: []string{"test_alias"}, Destination: dest}),
		FlagName: "test_alias",
		MapValue: []interface{}{1, 2},
	}
	c = runRacyTest(t, tis)
	refute(t, c.IntSlice("test_alias"), []int{1, 2})
	refute(t, dest.Value(), []int{1, 2})
}

func TestIntSliceApplyInputSourceValue(t *testing.T) {
	dest := cli.NewIntSlice()
	tis := testApplyInputSource{
		Flag:     NewIntSliceFlag(&cli.IntSliceFlag{Name: "test", Destination: dest}),
		FlagName: "test",
		MapValue: []interface{}{1, 2},
	}
	c := runTest(t, tis)
	expect(t, c.IntSlice("test"), []int{1, 2})
	expect(t, dest.Value(), []int{1, 2})

	// reset dest
	dest = cli.NewIntSlice()
	tis = testApplyInputSource{
		Flag:     NewIntSliceFlag(&cli.IntSliceFlag{Name: "test", Destination: dest}),
		FlagName: "test",
		MapValue: []interface{}{1, 2},
	}
	c = runRacyTest(t, tis)
	refute(t, c.IntSlice("test"), []int{1, 2})
	refute(t, dest.Value(), []int{1, 2})
}

func TestIntSliceApplyInputSourceMethodContextSet(t *testing.T) {
	dest := cli.NewIntSlice()
	tis := testApplyInputSource{
		Flag:               NewIntSliceFlag(&cli.IntSliceFlag{Name: "test", Destination: dest}),
		FlagName:           "test",
		MapValue:           []interface{}{1, 2},
		ContextValueString: "3",
	}
	c := runTest(t, tis)
	expect(t, c.IntSlice("test"), []int{3})
	expect(t, dest.Value(), []int{3})

	// reset dest
	dest = cli.NewIntSlice()
	tis = testApplyInputSource{
		Flag:               NewIntSliceFlag(&cli.IntSliceFlag{Name: "test", Destination: dest}),
		FlagName:           "test",
		MapValue:           []interface{}{1, 2},
		ContextValueString: "3",
	}
	c = runRacyTest(t, tis)
	refute(t, c.IntSlice("test"), []int{3})
	refute(t, dest.Value(), []int{3})
}

func TestIntSliceApplyInputSourceMethodEnvVarSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:        NewIntSliceFlag(&cli.IntSliceFlag{Name: "test", EnvVars: []string{"TEST"}}),
		FlagName:    "test",
		MapValue:    []interface{}{1, 2},
		EnvVarName:  "TEST",
		EnvVarValue: "3,4",
	}
	c := runTest(t, tis)
	expect(t, c.IntSlice("test"), []int{3, 4})

	c = runRacyTest(t, tis)
	refute(t, c.IntSlice("test"), []int{3, 4})
}

func TestInt64SliceFlagApplyInputSourceValue(t *testing.T) {
	dest := cli.NewInt64Slice()
	tis := testApplyInputSource{
		Flag:     NewInt64SliceFlag(&cli.Int64SliceFlag{Name: "test", Destination: dest}),
		FlagName: "test",
		MapValue: []interface{}{int64(1), int64(2)},
	}
	c := runTest(t, tis)
	expect(t, c.Int64Slice("test"), []int64{1, 2})
	expect(t, dest.Value(), []int64{1, 2})

	// reset dest
	dest = cli.NewInt64Slice()
	tis = testApplyInputSource{
		Flag:     NewInt64SliceFlag(&cli.Int64SliceFlag{Name: "test", Destination: dest}),
		FlagName: "test",
		MapValue: []interface{}{int64(1), int64(2)},
	}
	c = runRacyTest(t, tis)
	refute(t, c.IntSlice("test"), []int64{1, 2})
	refute(t, dest.Value(), []int64{1, 2})
}

func TestBoolApplyInputSourceMethodSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:     NewBoolFlag(&cli.BoolFlag{Name: "test"}),
		FlagName: "test",
		MapValue: true,
	}
	c := runTest(t, tis)
	expect(t, true, c.Bool("test"))

	c = runRacyTest(t, tis)
	refute(t, true, c.Bool("test"))
}

func TestBoolApplyInputSourceMethodSet_Alias(t *testing.T) {
	tis := testApplyInputSource{
		Flag:     NewBoolFlag(&cli.BoolFlag{Name: "test", Aliases: []string{"test_alias"}}),
		FlagName: "test_alias",
		MapValue: true,
	}
	c := runTest(t, tis)
	expect(t, true, c.Bool("test_alias"))

	c = runRacyTest(t, tis)
	refute(t, true, c.Bool("test_alias"))
}

func TestBoolApplyInputSourceMethodContextSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:               NewBoolFlag(&cli.BoolFlag{Name: "test"}),
		FlagName:           "test",
		MapValue:           false,
		ContextValueString: "true",
	}
	c := runTest(t, tis)
	expect(t, true, c.Bool("test"))

	c = runRacyTest(t, tis)
	refute(t, true, c.Bool("test"))
}

func TestBoolApplyInputSourceMethodEnvVarSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:        NewBoolFlag(&cli.BoolFlag{Name: "test", EnvVars: []string{"TEST"}}),
		FlagName:    "test",
		MapValue:    false,
		EnvVarName:  "TEST",
		EnvVarValue: "true",
	}
	c := runTest(t, tis)
	expect(t, true, c.Bool("test"))

	c = runRacyTest(t, tis)
	refute(t, true, c.Bool("test"))
}

func TestStringApplyInputSourceMethodSet_Alias(t *testing.T) {
	tis := testApplyInputSource{
		Flag:     NewStringFlag(&cli.StringFlag{Name: "test", Aliases: []string{"test_alias"}}),
		FlagName: "test_alias",
		MapValue: "hello",
	}
	c := runTest(t, tis)
	expect(t, "hello", c.String("test_alias"))

	c = runRacyTest(t, tis)
	refute(t, "hello", c.String("test_alias"))
}

func TestStringApplyInputSourceMethodSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:     NewStringFlag(&cli.StringFlag{Name: "test"}),
		FlagName: "test",
		MapValue: "hello",
	}
	c := runTest(t, tis)
	expect(t, "hello", c.String("test"))

	c = runRacyTest(t, tis)
	refute(t, "hello", c.String("test"))
}

func TestStringApplyInputSourceMethodContextSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:               NewStringFlag(&cli.StringFlag{Name: "test"}),
		FlagName:           "test",
		MapValue:           "hello",
		ContextValueString: "goodbye",
	}
	c := runTest(t, tis)
	expect(t, "goodbye", c.String("test"))

	c = runRacyTest(t, tis)
	refute(t, "goodbye", c.String("test"))
}

func TestStringApplyInputSourceMethodEnvVarSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:        NewStringFlag(&cli.StringFlag{Name: "test", EnvVars: []string{"TEST"}}),
		FlagName:    "test",
		MapValue:    "hello",
		EnvVarName:  "TEST",
		EnvVarValue: "goodbye",
	}
	c := runTest(t, tis)
	expect(t, "goodbye", c.String("test"))

	c = runRacyTest(t, tis)
	refute(t, "goodbye", c.String("test"))
}

func TestPathApplyInputSourceMethodSet_Alias(t *testing.T) {
	tis := testApplyInputSource{
		Flag:       NewPathFlag(&cli.PathFlag{Name: "test", Aliases: []string{"test_alias"}}),
		FlagName:   "test_alias",
		MapValue:   "hello",
		SourcePath: "/path/to/source/file",
	}
	c := runTest(t, tis)

	expected := "/path/to/source/hello"
	if runtime.GOOS == "windows" {
		var err error
		// Prepend the corresponding drive letter (or UNC path?), and change
		// to windows-style path:
		expected, err = filepath.Abs(expected)
		if err != nil {
			t.Fatal(err)
		}
	}
	expect(t, expected, c.String("test_alias"))

	c = runRacyTest(t, tis)
	refute(t, expected, c.String("test_alias"))
}

func TestPathApplyInputSourceMethodSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:       NewPathFlag(&cli.PathFlag{Name: "test"}),
		FlagName:   "test",
		MapValue:   "hello",
		SourcePath: "/path/to/source/file",
	}
	c := runTest(t, tis)

	expected := "/path/to/source/hello"
	if runtime.GOOS == "windows" {
		var err error
		// Prepend the corresponding drive letter (or UNC path?), and change
		// to windows-style path:
		expected, err = filepath.Abs(expected)
		if err != nil {
			t.Fatal(err)
		}
	}
	expect(t, expected, c.String("test"))

	c = runRacyTest(t, tis)
	refute(t, expected, c.String("test"))
}

func TestPathApplyInputSourceMethodContextSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:               NewPathFlag(&cli.PathFlag{Name: "test"}),
		FlagName:           "test",
		MapValue:           "hello",
		ContextValueString: "goodbye",
		SourcePath:         "/path/to/source/file",
	}
	c := runTest(t, tis)
	expect(t, "goodbye", c.String("test"))

	c = runRacyTest(t, tis)
	refute(t, "goodbye", c.String("test"))
}

func TestPathApplyInputSourceMethodEnvVarSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:        NewPathFlag(&cli.PathFlag{Name: "test", EnvVars: []string{"TEST"}}),
		FlagName:    "test",
		MapValue:    "hello",
		EnvVarName:  "TEST",
		EnvVarValue: "goodbye",
		SourcePath:  "/path/to/source/file",
	}
	c := runTest(t, tis)
	expect(t, "goodbye", c.String("test"))

	c = runRacyTest(t, tis)
	refute(t, "goodbye", c.String("test"))
}

func TestIntApplyInputSourceMethodSet_Alias(t *testing.T) {
	tis := testApplyInputSource{
		Flag:     NewIntFlag(&cli.IntFlag{Name: "test", Aliases: []string{"test_alias"}}),
		FlagName: "test_alias",
		MapValue: 15,
	}
	c := runTest(t, tis)
	expect(t, 15, c.Int("test_alias"))

	c = runRacyTest(t, tis)
	refute(t, 15, c.Int("test_alias"))
}

func TestIntApplyInputSourceMethodSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:     NewIntFlag(&cli.IntFlag{Name: "test"}),
		FlagName: "test",
		MapValue: 15,
	}
	c := runTest(t, tis)
	expect(t, 15, c.Int("test"))

	c = runRacyTest(t, tis)
	refute(t, 15, c.Int("test"))
}

func TestIntApplyInputSourceMethodSetNegativeValue(t *testing.T) {
	tis := testApplyInputSource{
		Flag:     NewIntFlag(&cli.IntFlag{Name: "test"}),
		FlagName: "test",
		MapValue: -1,
	}
	c := runTest(t, tis)
	expect(t, -1, c.Int("test"))

	c = runRacyTest(t, tis)
	refute(t, -1, c.Int("test"))
}

func TestIntApplyInputSourceMethodContextSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:               NewIntFlag(&cli.IntFlag{Name: "test"}),
		FlagName:           "test",
		MapValue:           15,
		ContextValueString: "7",
	}
	c := runTest(t, tis)
	expect(t, 7, c.Int("test"))

	c = runRacyTest(t, tis)
	refute(t, 7, c.Int("test"))
}

func TestIntApplyInputSourceMethodEnvVarSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:        NewIntFlag(&cli.IntFlag{Name: "test", EnvVars: []string{"TEST"}}),
		FlagName:    "test",
		MapValue:    15,
		EnvVarName:  "TEST",
		EnvVarValue: "12",
	}
	c := runTest(t, tis)
	expect(t, 12, c.Int("test"))

	c = runRacyTest(t, tis)
	refute(t, 12, c.Int("test"))
}

func TestDurationApplyInputSourceMethodSet_Alias(t *testing.T) {
	tis := testApplyInputSource{
		Flag:     NewDurationFlag(&cli.DurationFlag{Name: "test", Aliases: []string{"test_alias"}}),
		FlagName: "test_alias",
		MapValue: 30 * time.Second,
	}
	c := runTest(t, tis)
	expect(t, 30*time.Second, c.Duration("test_alias"))

	c = runRacyTest(t, tis)
	refute(t, 30*time.Second, c.Duration("test_alias"))
}

func TestDurationApplyInputSourceMethodSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:     NewDurationFlag(&cli.DurationFlag{Name: "test"}),
		FlagName: "test",
		MapValue: 30 * time.Second,
	}
	c := runTest(t, tis)
	expect(t, 30*time.Second, c.Duration("test"))

	c = runRacyTest(t, tis)
	refute(t, 30*time.Second, c.Duration("test"))
}

func TestDurationApplyInputSourceMethodSetNegativeValue(t *testing.T) {
	tis := testApplyInputSource{
		Flag:     NewDurationFlag(&cli.DurationFlag{Name: "test"}),
		FlagName: "test",
		MapValue: -30 * time.Second,
	}
	c := runTest(t, tis)
	expect(t, -30*time.Second, c.Duration("test"))

	c = runRacyTest(t, tis)
	refute(t, -30*time.Second, c.Duration("test"))
}

func TestDurationApplyInputSourceMethodContextSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:               NewDurationFlag(&cli.DurationFlag{Name: "test"}),
		FlagName:           "test",
		MapValue:           30 * time.Second,
		ContextValueString: (15 * time.Second).String(),
	}
	c := runTest(t, tis)
	expect(t, 15*time.Second, c.Duration("test"))

	c = runRacyTest(t, tis)
	refute(t, 15*time.Second, c.Duration("test"))
}

func TestDurationApplyInputSourceMethodEnvVarSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:        NewDurationFlag(&cli.DurationFlag{Name: "test", EnvVars: []string{"TEST"}}),
		FlagName:    "test",
		MapValue:    30 * time.Second,
		EnvVarName:  "TEST",
		EnvVarValue: (15 * time.Second).String(),
	}
	c := runTest(t, tis)
	expect(t, 15*time.Second, c.Duration("test"))

	c = runRacyTest(t, tis)
	refute(t, 15*time.Second, c.Duration("test"))
}

func TestFloat64ApplyInputSourceMethodSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:     NewFloat64Flag(&cli.Float64Flag{Name: "test"}),
		FlagName: "test",
		MapValue: 1.3,
	}
	c := runTest(t, tis)
	expect(t, 1.3, c.Float64("test"))

	c = runRacyTest(t, tis)
	refute(t, 1.3, c.Float64("test"))
}

func TestFloat64ApplyInputSourceMethodSetNegativeValue_Alias(t *testing.T) {
	tis := testApplyInputSource{
		Flag:     NewFloat64Flag(&cli.Float64Flag{Name: "test", Aliases: []string{"test_alias"}}),
		FlagName: "test_alias",
		MapValue: -1.3,
	}
	c := runTest(t, tis)
	expect(t, -1.3, c.Float64("test_alias"))

	c = runRacyTest(t, tis)
	refute(t, -1.3, c.Float64("test_alias"))
}

func TestFloat64ApplyInputSourceMethodSetNegativeValue(t *testing.T) {
	tis := testApplyInputSource{
		Flag:     NewFloat64Flag(&cli.Float64Flag{Name: "test"}),
		FlagName: "test",
		MapValue: -1.3,
	}
	c := runTest(t, tis)
	expect(t, -1.3, c.Float64("test"))

	c = runRacyTest(t, tis)
	refute(t, -1.3, c.Float64("test"))
}

func TestFloat64ApplyInputSourceMethodSetNegativeValueNotSet(t *testing.T) {
	c := runTest(t, testApplyInputSource{
		Flag:     NewFloat64Flag(&cli.Float64Flag{Name: "test1"}),
		FlagName: "test1",
		// dont set map value
	})
	expect(t, 0.0, c.Float64("test1"))
}

func TestFloat64ApplyInputSourceMethodContextSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:               NewFloat64Flag(&cli.Float64Flag{Name: "test"}),
		FlagName:           "test",
		MapValue:           1.3,
		ContextValueString: fmt.Sprintf("%v", 1.4),
	}
	c := runTest(t, tis)
	expect(t, 1.4, c.Float64("test"))

	c = runRacyTest(t, tis)
	refute(t, 1.4, c.Float64("test"))
}

func TestFloat64ApplyInputSourceMethodEnvVarSet(t *testing.T) {
	tis := testApplyInputSource{
		Flag:        NewFloat64Flag(&cli.Float64Flag{Name: "test", EnvVars: []string{"TEST"}}),
		FlagName:    "test",
		MapValue:    1.3,
		EnvVarName:  "TEST",
		EnvVarValue: fmt.Sprintf("%v", 1.4),
	}
	c := runTest(t, tis)
	expect(t, 1.4, c.Float64("test"))

	c = runRacyTest(t, tis)
	refute(t, 1.4, c.Float64("test"))
}

func TestJsonMissingKey(t *testing.T) {
	type Foo struct {
		IntFlag     int
		Int64Flag   int64
		Float64Flag float64
	}
	f := &Foo{
		IntFlag:     32,
		Int64Flag:   222,
		Float64Flag: 1.001,
	}
	bytes, err := json.Marshal(f)
	if err != nil {
		t.Error(err)
	}
	jis, err := NewJSONSource(bytes)
	if err != nil {
		t.Error(err)
	}

	// define flags with names non existent in the json input source
	flags := []FlagInputSourceExtension{
		NewIntFlag(&cli.IntFlag{Name: "inta"}),
		NewIntSliceFlag(&cli.IntSliceFlag{Name: "intslicea"}),
		NewFloat64Flag(&cli.Float64Flag{Name: "float64a"}),
		NewBoolFlag(&cli.BoolFlag{Name: "boola"}),
		NewStringFlag(&cli.StringFlag{Name: "stringa"}),
		NewStringSliceFlag(&cli.StringSliceFlag{Name: "stringslicea"}),
		NewInt64SliceFlag(&cli.Int64SliceFlag{Name: "int64slicea"}),
		NewDurationFlag(&cli.DurationFlag{Name: "dflag"}),
		NewPathFlag(&cli.PathFlag{Name: "patha"}),
		NewGenericFlag(&cli.GenericFlag{Name: "gflag", Value: &Parser{}}),
	}

	set := flag.NewFlagSet("test", flag.ContinueOnError)
	c := cli.NewContext(nil, set, nil)

	for _, f := range flags {
		if err := f.Apply(set); err != nil {
			t.Error(err)
		}
		if err := f.ApplyInputSourceValue(c, jis); err != nil {
			t.Error(err)
		}
	}
}

func runTest(t *testing.T, test testApplyInputSource) *cli.Context {
	inputSource := &MapInputSource{
		file:     test.SourcePath,
		valueMap: map[interface{}]interface{}{test.FlagName: test.MapValue},
	}
	set := flag.NewFlagSet(test.FlagSetName, flag.ContinueOnError)
	c := cli.NewContext(nil, set, nil)
	if test.EnvVarName != "" && test.EnvVarValue != "" {
		_ = os.Setenv(test.EnvVarName, test.EnvVarValue)
		defer os.Setenv(test.EnvVarName, "")
	}

	_ = test.Flag.Apply(set)
	if test.ContextValue != nil {
		f := set.Lookup(test.FlagName)
		f.Value = test.ContextValue
	}
	if test.ContextValueString != "" {
		_ = set.Set(test.FlagName, test.ContextValueString)
	}
	_ = test.Flag.ApplyInputSourceValue(c, inputSource)

	return c
}

func runRacyTest(t *testing.T, test testApplyInputSource) *cli.Context {
	set := flag.NewFlagSet(test.FlagSetName, flag.ContinueOnError)
	c := cli.NewContext(nil, set, nil)
	_ = test.Flag.ApplyInputSourceValue(c, &racyInputSource{
		MapInputSource: &MapInputSource{
			file:     test.SourcePath,
			valueMap: map[interface{}]interface{}{test.FlagName: test.MapValue},
		},
	})

	return c
}

type Parser [2]string

func (p *Parser) Set(value string) error {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format")
	}

	(*p)[0] = parts[0]
	(*p)[1] = parts[1]

	return nil
}

func (p *Parser) String() string {
	return fmt.Sprintf("%s,%s", p[0], p[1])
}

type bogus [1]uint
