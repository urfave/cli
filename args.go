package cli

import (
	"fmt"
	"time"
)

type Args interface {
	// Get returns the nth argument, or else a blank string
	Get(n int) string
	// First returns the first argument, or else a blank string
	First() string
	// Tail returns the rest of the arguments (not the first one)
	// or else an empty string slice
	Tail() []string
	// Len returns the length of the wrapped slice
	Len() int
	// Present checks if there are any arguments present
	Present() bool
	// Slice returns a copy of the internal slice
	Slice() []string
}

type stringSliceArgs struct {
	v []string
}

func (a *stringSliceArgs) Get(n int) string {
	if len(a.v) > n {
		return a.v[n]
	}
	return ""
}

func (a *stringSliceArgs) First() string {
	return a.Get(0)
}

func (a *stringSliceArgs) Tail() []string {
	if a.Len() >= 2 {
		tail := a.v[1:]
		ret := make([]string, len(tail))
		copy(ret, tail)
		return ret
	}

	return []string{}
}

func (a *stringSliceArgs) Len() int {
	return len(a.v)
}

func (a *stringSliceArgs) Present() bool {
	return a.Len() != 0
}

func (a *stringSliceArgs) Slice() []string {
	ret := make([]string, len(a.v))
	copy(ret, a.v)
	return ret
}

// Argument captures a positional argument that can
// be parsed
type Argument interface {
	// which this argument can be accessed using the given name
	HasName(string) bool

	// Parse the given args and return unparsed args and/or error
	Parse([]string) ([]string, error)

	// The usage template for this argument to use in help
	Usage() string

	// The Value of this Arg
	Get() any
}

// AnyArguments to differentiate between no arguments(nil) vs aleast one
var AnyArguments = []Argument{
	&StringArgs{
		Max: -1,
	},
}

type ArgumentsBase[T any, C any, VC ValueCreator[T, C]] struct {
	Name        string `json:"name"`      // the name of this argument
	Value       T      `json:"value"`     // the default value of this argument
	Destination *[]T   `json:"-"`         // the destination point for this argument
	UsageText   string `json:"usageText"` // the usage text to show
	Min         int    `json:"minTimes"`  // the min num of occurrences of this argument
	Max         int    `json:"maxTimes"`  // the max num of occurrences of this argument, set to -1 for unlimited
	Config      C      `json:"config"`    // config for this argument similar to Flag Config

	values []T
}

func (a *ArgumentsBase[T, C, VC]) HasName(s string) bool {
	return s == a.Name
}

func (a *ArgumentsBase[T, C, VC]) Usage() string {
	if a.UsageText != "" {
		return a.UsageText
	}

	usageFormat := ""
	if a.Min == 0 {
		if a.Max == 1 {
			usageFormat = "[%[1]s]"
		} else {
			usageFormat = "[%[1]s ...]"
		}
	} else {
		usageFormat = "%[1]s [%[1]s ...]"
	}
	return fmt.Sprintf(usageFormat, a.Name)
}

func (a *ArgumentsBase[T, C, VC]) Parse(s []string) ([]string, error) {
	tracef("calling arg%[1] parse with args %[2]", &a.Name, s)
	if a.Max == 0 {
		fmt.Printf("WARNING args %s has max 0, not parsing argument\n", a.Name)
		return s, nil
	}
	if a.Max != -1 && a.Min > a.Max {
		fmt.Printf("WARNING args %s has min[%d] > max[%d], not parsing argument\n", a.Name, a.Min, a.Max)
		return s, nil
	}

	count := 0
	var vc VC
	var t T
	value := vc.Create(a.Value, &t, a.Config)
	a.values = []T{}

	tracef("attempting arg%[1] parse", &a.Name)
	for _, arg := range s {
		if err := value.Set(arg); err != nil {
			return s, err
		}
		tracef("set arg%[1] one value", &a.Name, value.Get().(T))
		a.values = append(a.values, value.Get().(T))
		count++
		if count >= a.Max && a.Max > -1 {
			break
		}
	}
	if count < a.Min {
		return s, fmt.Errorf("sufficient count of arg %s not provided, given %d expected %d", a.Name, count, a.Min)
	}

	if a.Destination != nil {
		tracef("appending destination")
		*a.Destination = append(*a.Destination, a.values...)
	}

	return s[count:], nil
}

func (a *ArgumentsBase[T, C, VC]) Get() any {
	return a.values
}

type (
	FloatArgs     = ArgumentsBase[float64, NoConfig, floatValue]
	IntArgs       = ArgumentsBase[int64, IntegerConfig, intValue]
	StringArgs    = ArgumentsBase[string, StringConfig, stringValue]
	StringMapArgs = ArgumentsBase[map[string]string, StringConfig, StringMap]
	TimestampArgs = ArgumentsBase[time.Time, TimestampConfig, timestampValue]
	UintArgs      = ArgumentsBase[uint64, IntegerConfig, uintValue]
)

func (c *Command) StringArgs(name string) []string {
	for _, arg := range c.Arguments {
		if arg.HasName(name) {
			if a, ok := arg.Get().([]string); ok {
				return a
			} else {
				return nil
			}
		}
	}
	return nil
}

func (c *Command) FloatArgs(name string) []float64 {
	for _, arg := range c.Arguments {
		if arg.HasName(name) {
			if a, ok := arg.Get().([]float64); ok {
				return a
			} else {
				return nil
			}
		}
	}
	return nil
}

func (c *Command) IntArgs(name string) []int64 {
	for _, arg := range c.Arguments {
		if arg.HasName(name) {
			if a, ok := arg.Get().([]int64); ok {
				return a
			} else {
				return nil
			}
		}
	}
	return nil
}

func (c *Command) UintArgs(name string) []uint64 {
	for _, arg := range c.Arguments {
		if arg.HasName(name) {
			if a, ok := arg.Get().([]uint64); ok {
				return a
			} else {
				return nil
			}
		}
	}
	return nil
}

func (c *Command) TimestampArgs(name string) []time.Time {
	for _, arg := range c.Arguments {
		if arg.HasName(name) {
			if a, ok := arg.Get().([]time.Time); ok {
				return a
			} else {
				return nil
			}
		}
	}
	return nil
}
