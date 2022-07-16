//go:build go1.18
// +build go1.18

package cli

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func ExampleMultiStringFlag() {
	run := func(args ...string) {
		// add $0 (the command being run)
		args = append([]string{`-`}, args...)
		type CustomStringSlice []string
		type Config struct {
			FlagOne []string
			Two     CustomStringSlice
		}
		cfg := Config{
			Two: []string{
				`default value 1`,
				`default value 2`,
			},
		}
		if err := (&App{
			Flags: []Flag{
				&MultiStringFlag{
					Target: &StringSliceFlag{
						Name:     `flag-one`,
						Category: `category1`,
						Usage:    `this is the first flag`,
						Aliases:  []string{`1`},
						EnvVars:  []string{`FLAG_ONE`},
					},
					Value:       cfg.FlagOne,
					Destination: &cfg.FlagOne,
				},
				&SliceFlag[*StringSliceFlag, CustomStringSlice, string]{
					Target: &StringSliceFlag{
						Name:     `two`,
						Category: `category2`,
						Usage:    `this is the second flag`,
						Aliases:  []string{`2`},
						EnvVars:  []string{`TWO`},
					},
					Value:       cfg.Two,
					Destination: &cfg.Two,
				},
				&MultiStringFlag{
					Target: &StringSliceFlag{
						Name:     `flag-three`,
						Category: `category1`,
						Usage:    `this is the third flag`,
						Aliases:  []string{`3`},
						EnvVars:  []string{`FLAG_THREE`},
					},
					Value: []string{`some value`},
				},
				&StringSliceFlag{
					Name:     `flag-four`,
					Category: `category2`,
					Usage:    `this is the fourth flag`,
					Aliases:  []string{`4`},
					EnvVars:  []string{`FLAG_FOUR`},
					Value:    NewStringSlice(`d1`, `d2`),
				},
			},
			Action: func(c *Context) error {
				fmt.Printf("Flag names: %q\n", c.FlagNames())
				fmt.Printf("Local flag names: %q\n", c.LocalFlagNames())
				fmt.Println(`Context values:`)
				for _, name := range [...]string{`flag-one`, `two`, `flag-three`, `flag-four`} {
					fmt.Printf("%q=%q\n", name, c.StringSlice(name))
				}
				fmt.Println(`Destination values:`)
				fmt.Printf("cfg.FlagOne=%q\n", cfg.FlagOne)
				fmt.Printf("cfg.Two=%q\n", cfg.Two)
				return nil
			},
			Writer:    os.Stdout,
			ErrWriter: os.Stdout,
			Name:      `app-name`,
		}).Run(args); err != nil {
			panic(err)
		}
	}

	fmt.Printf("Show defaults...\n\n")
	run()

	fmt.Printf("---\nSetting all flags via command line...\n\n")
	allFlagsArgs := []string{
		`-1`, `v 1`,
		`-1`, `v 2`,
		`-2`, `v 3`,
		`-2`, `v 4`,
		`-3`, `v 5`,
		`-3`, `v 6`,
		`-4`, `v 7`,
		`-4`, `v 8`,
	}
	run(allFlagsArgs...)

	func() {
		defer resetEnv(os.Environ())
		os.Clearenv()
		for _, args := range [...][2]string{
			{`FLAG_ONE`, `v 9, v 10`},
			{`TWO`, `v 11, v 12`},
			{`FLAG_THREE`, `v 13, v 14`},
			{`FLAG_FOUR`, `v 15, v 16`},
		} {
			if err := os.Setenv(args[0], args[1]); err != nil {
				panic(err)
			}
		}

		fmt.Printf("---\nSetting all flags via environment...\n\n")
		run()

		fmt.Printf("---\nWith the same environment + args from the previous example...\n\n")
		run(allFlagsArgs...)
	}()

	//output:
	//Show defaults...
	//
	//Flag names: []
	//Local flag names: []
	//Context values:
	//"flag-one"=[]
	//"two"=["default value 1" "default value 2"]
	//"flag-three"=["some value"]
	//"flag-four"=["d1" "d2"]
	//Destination values:
	//cfg.FlagOne=[]
	//cfg.Two=["default value 1" "default value 2"]
	//---
	//Setting all flags via command line...
	//
	//Flag names: ["1" "2" "3" "4" "flag-four" "flag-one" "flag-three" "two"]
	//Local flag names: ["1" "2" "3" "4" "flag-four" "flag-one" "flag-three" "two"]
	//Context values:
	//"flag-one"=["v 1" "v 2"]
	//"two"=["v 3" "v 4"]
	//"flag-three"=["v 5" "v 6"]
	//"flag-four"=["v 7" "v 8"]
	//Destination values:
	//cfg.FlagOne=["v 1" "v 2"]
	//cfg.Two=["v 3" "v 4"]
	//---
	//Setting all flags via environment...
	//
	//Flag names: []
	//Local flag names: []
	//Context values:
	//"flag-one"=["v 9" "v 10"]
	//"two"=["v 11" "v 12"]
	//"flag-three"=["v 13" "v 14"]
	//"flag-four"=["v 15" "v 16"]
	//Destination values:
	//cfg.FlagOne=["v 9" "v 10"]
	//cfg.Two=["v 11" "v 12"]
	//---
	//With the same environment + args from the previous example...
	//
	//Flag names: ["1" "2" "3" "4" "flag-four" "flag-one" "flag-three" "two"]
	//Local flag names: ["1" "2" "3" "4" "flag-four" "flag-one" "flag-three" "two"]
	//Context values:
	//"flag-one"=["v 1" "v 2"]
	//"two"=["v 3" "v 4"]
	//"flag-three"=["v 5" "v 6"]
	//"flag-four"=["v 7" "v 8"]
	//Destination values:
	//cfg.FlagOne=["v 1" "v 2"]
	//cfg.Two=["v 3" "v 4"]
}

func TestSliceFlag_Apply_string(t *testing.T) {
	normalise := func(v any) any {
		switch v := v.(type) {
		case *[]string:
			if v == nil {
				return nil
			}
			return *v
		case *StringSlice:
			if v == nil {
				return nil
			}
			return v.Value()
		}
		return v
	}
	expectEqual := func(t *testing.T, actual, expected any) {
		t.Helper()
		actual = normalise(actual)
		expected = normalise(expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("actual: %#v\nexpected: %#v", actual, expected)
		}
	}
	type Config struct {
		Flag        SliceFlagTarget[string]
		Value       *[]string
		Destination **[]string
		Context     *Context
		Check       func()
	}
	for _, tc := range [...]struct {
		Name    string
		Factory func(t *testing.T, f *StringSliceFlag) Config
	}{
		{
			Name: `once`,
			Factory: func(t *testing.T, f *StringSliceFlag) Config {
				v := SliceFlag[*StringSliceFlag, []string, string]{Target: f}
				return Config{
					Flag:        &v,
					Value:       &v.Value,
					Destination: &v.Destination,
					Check: func() {
						expectEqual(t, v.Value, v.Target.Value)
						expectEqual(t, v.Destination, v.Target.Destination)
					},
				}
			},
		},
		{
			Name: `twice`,
			Factory: func(t *testing.T, f *StringSliceFlag) Config {
				v := SliceFlag[*SliceFlag[*StringSliceFlag, []string, string], []string, string]{
					Target: &SliceFlag[*StringSliceFlag, []string, string]{Target: f},
				}
				return Config{
					Flag:        &v,
					Value:       &v.Value,
					Destination: &v.Destination,
					Check: func() {
						expectEqual(t, v.Value, v.Target.Value)
						expectEqual(t, v.Destination, v.Target.Destination)

						expectEqual(t, v.Value, v.Target.Target.Value)
						expectEqual(t, v.Destination, v.Target.Target.Destination)
					},
				}
			},
		},
		{
			Name: `thrice`,
			Factory: func(t *testing.T, f *StringSliceFlag) Config {
				v := SliceFlag[*SliceFlag[*SliceFlag[*StringSliceFlag, []string, string], []string, string], []string, string]{
					Target: &SliceFlag[*SliceFlag[*StringSliceFlag, []string, string], []string, string]{
						Target: &SliceFlag[*StringSliceFlag, []string, string]{Target: f},
					},
				}
				return Config{
					Flag:        &v,
					Value:       &v.Value,
					Destination: &v.Destination,
					Check: func() {
						expectEqual(t, v.Value, v.Target.Value)
						expectEqual(t, v.Destination, v.Target.Destination)

						expectEqual(t, v.Value, v.Target.Target.Value)
						expectEqual(t, v.Destination, v.Target.Target.Destination)

						expectEqual(t, v.Value, v.Target.Target.Target.Value)
						expectEqual(t, v.Destination, v.Target.Target.Target.Destination)
					},
				}
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run(`destination`, func(t *testing.T) {
				c := tc.Factory(t, &StringSliceFlag{
					Name:    `a`,
					EnvVars: []string{`APP_A`},
				})
				defer c.Check()
				vDefault := []string{`one`, ``, ``, `two`, ``}
				var vTarget []string
				*c.Value = vDefault
				*c.Destination = &vTarget
				if err := (&App{Action: func(c *Context) error { return nil }, Flags: []Flag{c.Flag}}).Run([]string{`-`, `--a=`, `--a=three`, `--a=`, `--a=`, `--a=four`, `--a=`, `--a=`}); err != nil {
					t.Fatal(err)
				}
				expectEqual(t, vDefault, []string{`one`, ``, ``, `two`, ``})
				expectEqual(t, vTarget, []string{"", "three", "", "", "four", "", ""})
			})
			t.Run(`context`, func(t *testing.T) {
				c := tc.Factory(t, &StringSliceFlag{
					Name:    `a`,
					EnvVars: []string{`APP_A`},
				})
				defer c.Check()
				vDefault := []string{`one`, ``, ``, `two`, ``}
				*c.Value = vDefault
				var vTarget []string
				if err := (&App{Action: func(c *Context) error {
					vTarget = c.StringSlice(`a`)
					return nil
				}, Flags: []Flag{c.Flag}}).Run([]string{`-`, `--a=`, `--a=three`, `--a=`, `--a=`, `--a=four`, `--a=`, `--a=`}); err != nil {
					t.Fatal(err)
				}
				expectEqual(t, vDefault, []string{`one`, ``, ``, `two`, ``})
				expectEqual(t, vTarget, []string{"", "three", "", "", "four", "", ""})
			})
			t.Run(`context with destination`, func(t *testing.T) {
				c := tc.Factory(t, &StringSliceFlag{
					Name:    `a`,
					EnvVars: []string{`APP_A`},
				})
				defer c.Check()
				vDefault := []string{`one`, ``, ``, `two`, ``}
				*c.Value = vDefault
				var vTarget []string
				var destination []string
				*c.Destination = &destination
				if err := (&App{Action: func(c *Context) error {
					vTarget = c.StringSlice(`a`)
					return nil
				}, Flags: []Flag{c.Flag}}).Run([]string{`-`, `--a=`, `--a=three`, `--a=`, `--a=`, `--a=four`, `--a=`, `--a=`}); err != nil {
					t.Fatal(err)
				}
				expectEqual(t, vDefault, []string{`one`, ``, ``, `two`, ``})
				expectEqual(t, vTarget, []string{"", "three", "", "", "four", "", ""})
				expectEqual(t, destination, []string{"", "three", "", "", "four", "", ""})
			})
			t.Run(`stdlib flag usage with default`, func(t *testing.T) {
				c := tc.Factory(t, &StringSliceFlag{Name: `a`})
				*c.Value = []string{`one`, `two`}
				var vTarget []string
				*c.Destination = &vTarget
				set := flag.NewFlagSet(`flagset`, flag.ContinueOnError)
				var output bytes.Buffer
				set.SetOutput(&output)
				if err := c.Flag.Apply(set); err != nil {
					t.Fatal(err)
				}
				if err := set.Parse([]string{`-h`}); err != flag.ErrHelp {
					t.Fatal(err)
				}
				if s := output.String(); s != "Usage of flagset:\n  -a value\n    \t (default [one two])\n" {
					t.Errorf("unexpected output: %q\n%s", s, s)
				}
			})
			{
				test := func(t *testing.T, value []string) {
					c := tc.Factory(t, &StringSliceFlag{Name: `a`})
					*c.Value = value
					var vTarget []string
					*c.Destination = &vTarget
					set := flag.NewFlagSet(`flagset`, flag.ContinueOnError)
					var output bytes.Buffer
					set.SetOutput(&output)
					if err := c.Flag.Apply(set); err != nil {
						t.Fatal(err)
					}
					if err := set.Parse([]string{`-h`}); err != flag.ErrHelp {
						t.Fatal(err)
					}
					if s := output.String(); s != "Usage of flagset:\n  -a value\n    \t\n" {
						t.Errorf("unexpected output: %q\n%s", s, s)
					}
				}
				t.Run(`stdlib flag usage without default nil`, func(t *testing.T) {
					test(t, nil)
				})
				t.Run(`stdlib flag usage without default empty`, func(t *testing.T) {
					test(t, make([]string, 0))
				})
			}
		})
	}
}

func TestSliceFlag_Apply_float64(t *testing.T) {
	normalise := func(v any) any {
		switch v := v.(type) {
		case *[]float64:
			if v == nil {
				return nil
			}
			return *v
		case *Float64Slice:
			if v == nil {
				return nil
			}
			return v.Value()
		}
		return v
	}
	expectEqual := func(t *testing.T, actual, expected any) {
		t.Helper()
		actual = normalise(actual)
		expected = normalise(expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("actual: %#v\nexpected: %#v", actual, expected)
		}
	}
	type Config struct {
		Flag        SliceFlagTarget[float64]
		Value       *[]float64
		Destination **[]float64
		Context     *Context
		Check       func()
	}
	for _, tc := range [...]struct {
		Name    string
		Factory func(t *testing.T, f *Float64SliceFlag) Config
	}{
		{
			Name: `once`,
			Factory: func(t *testing.T, f *Float64SliceFlag) Config {
				v := SliceFlag[*Float64SliceFlag, []float64, float64]{Target: f}
				return Config{
					Flag:        &v,
					Value:       &v.Value,
					Destination: &v.Destination,
					Check: func() {
						expectEqual(t, v.Value, v.Target.Value)
						expectEqual(t, v.Destination, v.Target.Destination)
					},
				}
			},
		},
		{
			Name: `twice`,
			Factory: func(t *testing.T, f *Float64SliceFlag) Config {
				v := SliceFlag[*SliceFlag[*Float64SliceFlag, []float64, float64], []float64, float64]{
					Target: &SliceFlag[*Float64SliceFlag, []float64, float64]{Target: f},
				}
				return Config{
					Flag:        &v,
					Value:       &v.Value,
					Destination: &v.Destination,
					Check: func() {
						expectEqual(t, v.Value, v.Target.Value)
						expectEqual(t, v.Destination, v.Target.Destination)

						expectEqual(t, v.Value, v.Target.Target.Value)
						expectEqual(t, v.Destination, v.Target.Target.Destination)
					},
				}
			},
		},
		{
			Name: `thrice`,
			Factory: func(t *testing.T, f *Float64SliceFlag) Config {
				v := SliceFlag[*SliceFlag[*SliceFlag[*Float64SliceFlag, []float64, float64], []float64, float64], []float64, float64]{
					Target: &SliceFlag[*SliceFlag[*Float64SliceFlag, []float64, float64], []float64, float64]{
						Target: &SliceFlag[*Float64SliceFlag, []float64, float64]{Target: f},
					},
				}
				return Config{
					Flag:        &v,
					Value:       &v.Value,
					Destination: &v.Destination,
					Check: func() {
						expectEqual(t, v.Value, v.Target.Value)
						expectEqual(t, v.Destination, v.Target.Destination)

						expectEqual(t, v.Value, v.Target.Target.Value)
						expectEqual(t, v.Destination, v.Target.Target.Destination)

						expectEqual(t, v.Value, v.Target.Target.Target.Value)
						expectEqual(t, v.Destination, v.Target.Target.Target.Destination)
					},
				}
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run(`destination`, func(t *testing.T) {
				c := tc.Factory(t, &Float64SliceFlag{
					Name:    `a`,
					EnvVars: []string{`APP_A`},
				})
				defer c.Check()
				vDefault := []float64{1, 2, 3}
				var vTarget []float64
				*c.Value = vDefault
				*c.Destination = &vTarget
				if err := (&App{Action: func(c *Context) error { return nil }, Flags: []Flag{c.Flag}}).Run([]string{`-`, `--a=4`, `--a=5`}); err != nil {
					t.Fatal(err)
				}
				expectEqual(t, vDefault, []float64{1, 2, 3})
				expectEqual(t, vTarget, []float64{4, 5})
			})
			t.Run(`context`, func(t *testing.T) {
				c := tc.Factory(t, &Float64SliceFlag{
					Name:    `a`,
					EnvVars: []string{`APP_A`},
				})
				defer c.Check()
				vDefault := []float64{1, 2, 3}
				*c.Value = vDefault
				var vTarget []float64
				if err := (&App{Action: func(c *Context) error {
					vTarget = c.Float64Slice(`a`)
					return nil
				}, Flags: []Flag{c.Flag}}).Run([]string{`-`, `--a=4`, `--a=5`}); err != nil {
					t.Fatal(err)
				}
				expectEqual(t, vDefault, []float64{1, 2, 3})
				expectEqual(t, vTarget, []float64{4, 5})
			})
			t.Run(`context with destination`, func(t *testing.T) {
				c := tc.Factory(t, &Float64SliceFlag{
					Name:    `a`,
					EnvVars: []string{`APP_A`},
				})
				defer c.Check()
				vDefault := []float64{1, 2, 3}
				*c.Value = vDefault
				var vTarget []float64
				var destination []float64
				*c.Destination = &destination
				if err := (&App{Action: func(c *Context) error {
					vTarget = c.Float64Slice(`a`)
					return nil
				}, Flags: []Flag{c.Flag}}).Run([]string{`-`, `--a=4`, `--a=5`}); err != nil {
					t.Fatal(err)
				}
				expectEqual(t, vDefault, []float64{1, 2, 3})
				expectEqual(t, vTarget, []float64{4, 5})
				expectEqual(t, destination, []float64{4, 5})
			})
			t.Run(`stdlib flag usage with default`, func(t *testing.T) {
				c := tc.Factory(t, &Float64SliceFlag{Name: `a`})
				*c.Value = []float64{1, 2}
				var vTarget []float64
				*c.Destination = &vTarget
				set := flag.NewFlagSet(`flagset`, flag.ContinueOnError)
				var output bytes.Buffer
				set.SetOutput(&output)
				if err := c.Flag.Apply(set); err != nil {
					t.Fatal(err)
				}
				if err := set.Parse([]string{`-h`}); err != flag.ErrHelp {
					t.Fatal(err)
				}
				if s := output.String(); s != "Usage of flagset:\n  -a value\n    \t (default []float64{1, 2})\n" {
					t.Errorf("unexpected output: %q\n%s", s, s)
				}
			})
			{
				test := func(t *testing.T, value []float64) {
					c := tc.Factory(t, &Float64SliceFlag{Name: `a`})
					*c.Value = value
					var vTarget []float64
					*c.Destination = &vTarget
					set := flag.NewFlagSet(`flagset`, flag.ContinueOnError)
					var output bytes.Buffer
					set.SetOutput(&output)
					if err := c.Flag.Apply(set); err != nil {
						t.Fatal(err)
					}
					if err := set.Parse([]string{`-h`}); err != flag.ErrHelp {
						t.Fatal(err)
					}
					if s := output.String(); s != "Usage of flagset:\n  -a value\n    \t\n" {
						t.Errorf("unexpected output: %q\n%s", s, s)
					}
				}
				t.Run(`stdlib flag usage without default nil`, func(t *testing.T) {
					test(t, nil)
				})
				t.Run(`stdlib flag usage without default empty`, func(t *testing.T) {
					test(t, make([]float64, 0))
				})
			}
		})
	}
}

func TestSliceFlag_Apply_int64(t *testing.T) {
	normalise := func(v any) any {
		switch v := v.(type) {
		case *[]int64:
			if v == nil {
				return nil
			}
			return *v
		case *Int64Slice:
			if v == nil {
				return nil
			}
			return v.Value()
		}
		return v
	}
	expectEqual := func(t *testing.T, actual, expected any) {
		t.Helper()
		actual = normalise(actual)
		expected = normalise(expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("actual: %#v\nexpected: %#v", actual, expected)
		}
	}
	type Config struct {
		Flag        SliceFlagTarget[int64]
		Value       *[]int64
		Destination **[]int64
		Context     *Context
		Check       func()
	}
	for _, tc := range [...]struct {
		Name    string
		Factory func(t *testing.T, f *Int64SliceFlag) Config
	}{
		{
			Name: `once`,
			Factory: func(t *testing.T, f *Int64SliceFlag) Config {
				v := SliceFlag[*Int64SliceFlag, []int64, int64]{Target: f}
				return Config{
					Flag:        &v,
					Value:       &v.Value,
					Destination: &v.Destination,
					Check: func() {
						expectEqual(t, v.Value, v.Target.Value)
						expectEqual(t, v.Destination, v.Target.Destination)
					},
				}
			},
		},
		{
			Name: `twice`,
			Factory: func(t *testing.T, f *Int64SliceFlag) Config {
				v := SliceFlag[*SliceFlag[*Int64SliceFlag, []int64, int64], []int64, int64]{
					Target: &SliceFlag[*Int64SliceFlag, []int64, int64]{Target: f},
				}
				return Config{
					Flag:        &v,
					Value:       &v.Value,
					Destination: &v.Destination,
					Check: func() {
						expectEqual(t, v.Value, v.Target.Value)
						expectEqual(t, v.Destination, v.Target.Destination)

						expectEqual(t, v.Value, v.Target.Target.Value)
						expectEqual(t, v.Destination, v.Target.Target.Destination)
					},
				}
			},
		},
		{
			Name: `thrice`,
			Factory: func(t *testing.T, f *Int64SliceFlag) Config {
				v := SliceFlag[*SliceFlag[*SliceFlag[*Int64SliceFlag, []int64, int64], []int64, int64], []int64, int64]{
					Target: &SliceFlag[*SliceFlag[*Int64SliceFlag, []int64, int64], []int64, int64]{
						Target: &SliceFlag[*Int64SliceFlag, []int64, int64]{Target: f},
					},
				}
				return Config{
					Flag:        &v,
					Value:       &v.Value,
					Destination: &v.Destination,
					Check: func() {
						expectEqual(t, v.Value, v.Target.Value)
						expectEqual(t, v.Destination, v.Target.Destination)

						expectEqual(t, v.Value, v.Target.Target.Value)
						expectEqual(t, v.Destination, v.Target.Target.Destination)

						expectEqual(t, v.Value, v.Target.Target.Target.Value)
						expectEqual(t, v.Destination, v.Target.Target.Target.Destination)
					},
				}
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run(`destination`, func(t *testing.T) {
				c := tc.Factory(t, &Int64SliceFlag{
					Name:    `a`,
					EnvVars: []string{`APP_A`},
				})
				defer c.Check()
				vDefault := []int64{1, 2, 3}
				var vTarget []int64
				*c.Value = vDefault
				*c.Destination = &vTarget
				if err := (&App{Action: func(c *Context) error { return nil }, Flags: []Flag{c.Flag}}).Run([]string{`-`, `--a=4`, `--a=5`}); err != nil {
					t.Fatal(err)
				}
				expectEqual(t, vDefault, []int64{1, 2, 3})
				expectEqual(t, vTarget, []int64{4, 5})
			})
			t.Run(`context`, func(t *testing.T) {
				c := tc.Factory(t, &Int64SliceFlag{
					Name:    `a`,
					EnvVars: []string{`APP_A`},
				})
				defer c.Check()
				vDefault := []int64{1, 2, 3}
				*c.Value = vDefault
				var vTarget []int64
				if err := (&App{Action: func(c *Context) error {
					vTarget = c.Int64Slice(`a`)
					return nil
				}, Flags: []Flag{c.Flag}}).Run([]string{`-`, `--a=4`, `--a=5`}); err != nil {
					t.Fatal(err)
				}
				expectEqual(t, vDefault, []int64{1, 2, 3})
				expectEqual(t, vTarget, []int64{4, 5})
			})
			t.Run(`context with destination`, func(t *testing.T) {
				c := tc.Factory(t, &Int64SliceFlag{
					Name:    `a`,
					EnvVars: []string{`APP_A`},
				})
				defer c.Check()
				vDefault := []int64{1, 2, 3}
				*c.Value = vDefault
				var vTarget []int64
				var destination []int64
				*c.Destination = &destination
				if err := (&App{Action: func(c *Context) error {
					vTarget = c.Int64Slice(`a`)
					return nil
				}, Flags: []Flag{c.Flag}}).Run([]string{`-`, `--a=4`, `--a=5`}); err != nil {
					t.Fatal(err)
				}
				expectEqual(t, vDefault, []int64{1, 2, 3})
				expectEqual(t, vTarget, []int64{4, 5})
				expectEqual(t, destination, []int64{4, 5})
			})
			t.Run(`stdlib flag usage with default`, func(t *testing.T) {
				c := tc.Factory(t, &Int64SliceFlag{Name: `a`})
				*c.Value = []int64{1, 2}
				var vTarget []int64
				*c.Destination = &vTarget
				set := flag.NewFlagSet(`flagset`, flag.ContinueOnError)
				var output bytes.Buffer
				set.SetOutput(&output)
				if err := c.Flag.Apply(set); err != nil {
					t.Fatal(err)
				}
				if err := set.Parse([]string{`-h`}); err != flag.ErrHelp {
					t.Fatal(err)
				}
				if s := output.String(); s != "Usage of flagset:\n  -a value\n    \t (default []int64{1, 2})\n" {
					t.Errorf("unexpected output: %q\n%s", s, s)
				}
			})
			{
				test := func(t *testing.T, value []int64) {
					c := tc.Factory(t, &Int64SliceFlag{Name: `a`})
					*c.Value = value
					var vTarget []int64
					*c.Destination = &vTarget
					set := flag.NewFlagSet(`flagset`, flag.ContinueOnError)
					var output bytes.Buffer
					set.SetOutput(&output)
					if err := c.Flag.Apply(set); err != nil {
						t.Fatal(err)
					}
					if err := set.Parse([]string{`-h`}); err != flag.ErrHelp {
						t.Fatal(err)
					}
					if s := output.String(); s != "Usage of flagset:\n  -a value\n    \t\n" {
						t.Errorf("unexpected output: %q\n%s", s, s)
					}
				}
				t.Run(`stdlib flag usage without default nil`, func(t *testing.T) {
					test(t, nil)
				})
				t.Run(`stdlib flag usage without default empty`, func(t *testing.T) {
					test(t, make([]int64, 0))
				})
			}
		})
	}
}

func TestSliceFlag_Apply_int(t *testing.T) {
	normalise := func(v any) any {
		switch v := v.(type) {
		case *[]int:
			if v == nil {
				return nil
			}
			return *v
		case *IntSlice:
			if v == nil {
				return nil
			}
			return v.Value()
		}
		return v
	}
	expectEqual := func(t *testing.T, actual, expected any) {
		t.Helper()
		actual = normalise(actual)
		expected = normalise(expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("actual: %#v\nexpected: %#v", actual, expected)
		}
	}
	type Config struct {
		Flag        SliceFlagTarget[int]
		Value       *[]int
		Destination **[]int
		Context     *Context
		Check       func()
	}
	for _, tc := range [...]struct {
		Name    string
		Factory func(t *testing.T, f *IntSliceFlag) Config
	}{
		{
			Name: `once`,
			Factory: func(t *testing.T, f *IntSliceFlag) Config {
				v := SliceFlag[*IntSliceFlag, []int, int]{Target: f}
				return Config{
					Flag:        &v,
					Value:       &v.Value,
					Destination: &v.Destination,
					Check: func() {
						expectEqual(t, v.Value, v.Target.Value)
						expectEqual(t, v.Destination, v.Target.Destination)
					},
				}
			},
		},
		{
			Name: `twice`,
			Factory: func(t *testing.T, f *IntSliceFlag) Config {
				v := SliceFlag[*SliceFlag[*IntSliceFlag, []int, int], []int, int]{
					Target: &SliceFlag[*IntSliceFlag, []int, int]{Target: f},
				}
				return Config{
					Flag:        &v,
					Value:       &v.Value,
					Destination: &v.Destination,
					Check: func() {
						expectEqual(t, v.Value, v.Target.Value)
						expectEqual(t, v.Destination, v.Target.Destination)

						expectEqual(t, v.Value, v.Target.Target.Value)
						expectEqual(t, v.Destination, v.Target.Target.Destination)
					},
				}
			},
		},
		{
			Name: `thrice`,
			Factory: func(t *testing.T, f *IntSliceFlag) Config {
				v := SliceFlag[*SliceFlag[*SliceFlag[*IntSliceFlag, []int, int], []int, int], []int, int]{
					Target: &SliceFlag[*SliceFlag[*IntSliceFlag, []int, int], []int, int]{
						Target: &SliceFlag[*IntSliceFlag, []int, int]{Target: f},
					},
				}
				return Config{
					Flag:        &v,
					Value:       &v.Value,
					Destination: &v.Destination,
					Check: func() {
						expectEqual(t, v.Value, v.Target.Value)
						expectEqual(t, v.Destination, v.Target.Destination)

						expectEqual(t, v.Value, v.Target.Target.Value)
						expectEqual(t, v.Destination, v.Target.Target.Destination)

						expectEqual(t, v.Value, v.Target.Target.Target.Value)
						expectEqual(t, v.Destination, v.Target.Target.Target.Destination)
					},
				}
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run(`destination`, func(t *testing.T) {
				c := tc.Factory(t, &IntSliceFlag{
					Name:    `a`,
					EnvVars: []string{`APP_A`},
				})
				defer c.Check()
				vDefault := []int{1, 2, 3}
				var vTarget []int
				*c.Value = vDefault
				*c.Destination = &vTarget
				if err := (&App{Action: func(c *Context) error { return nil }, Flags: []Flag{c.Flag}}).Run([]string{`-`, `--a=4`, `--a=5`}); err != nil {
					t.Fatal(err)
				}
				expectEqual(t, vDefault, []int{1, 2, 3})
				expectEqual(t, vTarget, []int{4, 5})
			})
			t.Run(`context`, func(t *testing.T) {
				c := tc.Factory(t, &IntSliceFlag{
					Name:    `a`,
					EnvVars: []string{`APP_A`},
				})
				defer c.Check()
				vDefault := []int{1, 2, 3}
				*c.Value = vDefault
				var vTarget []int
				if err := (&App{Action: func(c *Context) error {
					vTarget = c.IntSlice(`a`)
					return nil
				}, Flags: []Flag{c.Flag}}).Run([]string{`-`, `--a=4`, `--a=5`}); err != nil {
					t.Fatal(err)
				}
				expectEqual(t, vDefault, []int{1, 2, 3})
				expectEqual(t, vTarget, []int{4, 5})
			})
			t.Run(`context with destination`, func(t *testing.T) {
				c := tc.Factory(t, &IntSliceFlag{
					Name:    `a`,
					EnvVars: []string{`APP_A`},
				})
				defer c.Check()
				vDefault := []int{1, 2, 3}
				*c.Value = vDefault
				var vTarget []int
				var destination []int
				*c.Destination = &destination
				if err := (&App{Action: func(c *Context) error {
					vTarget = c.IntSlice(`a`)
					return nil
				}, Flags: []Flag{c.Flag}}).Run([]string{`-`, `--a=4`, `--a=5`}); err != nil {
					t.Fatal(err)
				}
				expectEqual(t, vDefault, []int{1, 2, 3})
				expectEqual(t, vTarget, []int{4, 5})
				expectEqual(t, destination, []int{4, 5})
			})
			t.Run(`stdlib flag usage with default`, func(t *testing.T) {
				c := tc.Factory(t, &IntSliceFlag{Name: `a`})
				*c.Value = []int{1, 2}
				var vTarget []int
				*c.Destination = &vTarget
				set := flag.NewFlagSet(`flagset`, flag.ContinueOnError)
				var output bytes.Buffer
				set.SetOutput(&output)
				if err := c.Flag.Apply(set); err != nil {
					t.Fatal(err)
				}
				if err := set.Parse([]string{`-h`}); err != flag.ErrHelp {
					t.Fatal(err)
				}
				if s := output.String(); s != "Usage of flagset:\n  -a value\n    \t (default []int{1, 2})\n" {
					t.Errorf("unexpected output: %q\n%s", s, s)
				}
			})
			{
				test := func(t *testing.T, value []int) {
					c := tc.Factory(t, &IntSliceFlag{Name: `a`})
					*c.Value = value
					var vTarget []int
					*c.Destination = &vTarget
					set := flag.NewFlagSet(`flagset`, flag.ContinueOnError)
					var output bytes.Buffer
					set.SetOutput(&output)
					if err := c.Flag.Apply(set); err != nil {
						t.Fatal(err)
					}
					if err := set.Parse([]string{`-h`}); err != flag.ErrHelp {
						t.Fatal(err)
					}
					if s := output.String(); s != "Usage of flagset:\n  -a value\n    \t\n" {
						t.Errorf("unexpected output: %q\n%s", s, s)
					}
				}
				t.Run(`stdlib flag usage without default nil`, func(t *testing.T) {
					test(t, nil)
				})
				t.Run(`stdlib flag usage without default empty`, func(t *testing.T) {
					test(t, make([]int, 0))
				})
			}
		})
	}
}

type intSliceWrapperDefaultingNil struct {
	*IntSlice
}

func (x intSliceWrapperDefaultingNil) String() string {
	if x.IntSlice != nil {
		return x.IntSlice.String()
	}
	return NewIntSlice().String()
}

func TestFlagValueHook_String_struct(t *testing.T) {
	wrap := func(values ...int) *flagValueHook {
		return &flagValueHook{value: intSliceWrapperDefaultingNil{NewIntSlice(values...)}}
	}
	if s := wrap().String(); s != `` {
		t.Error(s)
	}
	if s := wrap(1).String(); s != `[]int{1}` {
		t.Error(s)
	}
}
