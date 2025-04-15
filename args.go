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

type ArgumentBase[T any, C any, VC ValueCreator[T, C]] struct {
	Name        string `json:"name"`      // the name of this argument
	Value       T      `json:"value"`     // the default value of this argument
	Destination *T     `json:"-"`         // the destination point for this argument
	UsageText   string `json:"usageText"` // the usage text to show
	Config      C      `json:"config"`    // config for this argument similar to Flag Config

	value *T
}

func (a *ArgumentBase[T, C, VC]) HasName(s string) bool {
	return s == a.Name
}

func (a *ArgumentBase[T, C, VC]) Usage() string {
	if a.UsageText != "" {
		return a.UsageText
	}

	usageFormat := "%[1]s"
	return fmt.Sprintf(usageFormat, a.Name)
}

func (a *ArgumentBase[T, C, VC]) Parse(s []string) ([]string, error) {
	tracef("calling arg%[1] parse with args %[2]", a.Name, s)

	var vc VC
	var t T
	value := vc.Create(a.Value, &t, a.Config)
	a.value = &t

	tracef("attempting arg%[1] parse", &a.Name)
	if len(s) > 0 {
		if err := value.Set(s[0]); err != nil {
			return s, err
		}
		*a.value = value.Get().(T)
		tracef("set arg%[1] one value", a.Name, *a.value)
	}

	if a.Destination != nil {
		tracef("setting destination")
		*a.Destination = *a.value
	}

	if len(s) > 0 {
		return s[1:], nil
	}
	return s, nil
}

func (a *ArgumentBase[T, C, VC]) Get() any {
	if a.value != nil {
		return *a.value
	}
	return a.Value
}

// ArgumentsBase is a base type for slice arguments
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
		*a.Destination = a.values // append(*a.Destination, a.values...)
	}

	return s[count:], nil
}

func (a *ArgumentsBase[T, C, VC]) Get() any {
	if a.values != nil {
		return a.values
	}
	return []T{}
}

type (
	FloatArg      = ArgumentBase[float64, NoConfig, floatValue]
	IntArg        = ArgumentBase[int64, IntegerConfig, intValue]
	StringArg     = ArgumentBase[string, StringConfig, stringValue]
	StringMapArgs = ArgumentBase[map[string]string, StringConfig, StringMap]
	TimestampArg  = ArgumentBase[time.Time, TimestampConfig, timestampValue]
	UintArg       = ArgumentBase[uint64, IntegerConfig, uintValue]

	FloatArgs     = ArgumentsBase[float64, NoConfig, floatValue]
	IntArgs       = ArgumentsBase[int64, IntegerConfig, intValue]
	StringArgs    = ArgumentsBase[string, StringConfig, stringValue]
	TimestampArgs = ArgumentsBase[time.Time, TimestampConfig, timestampValue]
	UintArgs      = ArgumentsBase[uint64, IntegerConfig, uintValue]
)

func (c *Command) getArgValue(name string) any {
	tracef("command %s looking for args %s", c.Name, name)
	for _, arg := range c.Arguments {
		if arg.HasName(name) {
			tracef("command %s found args %s", c.Name, name)
			return arg.Get()
		}
	}
	tracef("command %s did not find args %s", c.Name, name)
	return nil
}

func (c *Command) StringArg(name string) string {
	val := c.getArgValue(name)
	if a, ok := val.(string); ok {
		return a
	}
	return ""
}

func (c *Command) StringArgs(name string) []string {
	val := c.getArgValue(name)
	if a, ok := val.([]string); ok {
		return a
	}
	return nil
}

func (c *Command) FloatArg(name string) float64 {
	val := c.getArgValue(name)
	if a, ok := val.(float64); ok {
		return a
	}
	return 0
}

func (c *Command) FloatArgs(name string) []float64 {
	val := c.getArgValue(name)
	if a, ok := val.([]float64); ok {
		return a
	}
	return nil
}

func (c *Command) IntArg(name string) int64 {
	val := c.getArgValue(name)
	if a, ok := val.(int64); ok {
		return a
	}
	return 0
}

func (c *Command) IntArgs(name string) []int64 {
	val := c.getArgValue(name)
	if a, ok := val.([]int64); ok {
		return a
	}
	return nil
}

func (c *Command) UintArg(name string) uint64 {
	val := c.getArgValue(name)
	if a, ok := val.(uint64); ok {
		return a
	}
	return 0
}

func (c *Command) UintArgs(name string) []uint64 {
	val := c.getArgValue(name)
	if a, ok := val.([]uint64); ok {
		return a
	}
	return nil
}

func (c *Command) TimestampArg(name string) time.Time {
	val := c.getArgValue(name)
	if a, ok := val.(time.Time); ok {
		return a
	}
	return time.Time{}
}

func (c *Command) TimestampArgs(name string) []time.Time {
	val := c.getArgValue(name)
	if a, ok := val.([]time.Time); ok {
		return a
	}
	return nil
}
