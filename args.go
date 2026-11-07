package cli

import (
	"fmt"
	"time"
)

// Args provides access to the unparsed positional arguments.
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
	if n >= 0 && n < len(a.v) {
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

type requiredArgument interface {
	name() string
	required() bool
}

// AnyArguments to differentiate between no arguments(nil) vs aleast one
var AnyArguments = []Argument{
	&StringArgs{
		Max: -1,
	},
}

// ArgumentBase defines a typed positional argument and its parsing configuration.
type ArgumentBase[T any, C any, VC ValueCreator[T, C]] struct {
	Name        string `json:"name"`      // the name of this argument
	Value       T      `json:"value"`     // the default value of this argument
	Destination *T     `json:"-"`         // the destination point for this argument
	UsageText   string `json:"usageText"` // the usage text to show
	Required    bool   `json:"required"`  // whether the argument is required or not
	Config      C      `json:"config"`    // config for this argument similar to Flag Config

	value *T
}

// HasName reports whether s is the name of this argument.
func (a *ArgumentBase[T, C, VC]) HasName(s string) bool {
	return s == a.Name
}

func (a *ArgumentBase[T, C, VC]) name() string {
	return a.Name
}

func (a *ArgumentBase[T, C, VC]) required() bool {
	return a.Required
}

// Usage returns the usage text for this argument.
func (a *ArgumentBase[T, C, VC]) Usage() string {
	if a.UsageText != "" {
		return a.UsageText
	}

	usageFormat := "[%[1]s]"
	if a.Required {
		usageFormat = "%[1]s"
	}
	return fmt.Sprintf(usageFormat, a.Name)
}

// Parse parses the positional arguments and returns any unconsumed arguments.
func (a *ArgumentBase[T, C, VC]) Parse(s []string) ([]string, error) {
	tracef("calling arg%[1] parse with args %[2]", a.Name, s)

	if a.Required && len(s) == 0 {
		return s, &errRequiredArguments{missingArguments: []string{a.Name}}
	}

	var vc VC
	var t T
	value := vc.Create(a.Value, &t, a.Config)
	a.value = &t

	tracef("attempting arg%[1] parse", &a.Name)
	if len(s) > 0 {
		if err := value.Set(s[0]); err != nil {
			return s, fmt.Errorf("invalid value %q for argument %s: %v", s[0], a.Name, err)
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

// Get returns the argument value.
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

// HasName reports whether s is the name of this argument.
func (a *ArgumentsBase[T, C, VC]) HasName(s string) bool {
	return s == a.Name
}

// Usage returns the usage text for this argument.
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

// Parse parses the positional arguments and returns any unconsumed arguments.
func (a *ArgumentsBase[T, C, VC]) Parse(s []string) ([]string, error) {
	tracef("calling arg%[1] parse with args %[2]", &a.Name, s)
	if a.Max == 0 {
		return s, fmt.Errorf("args %s has max 0, not parsing argument", a.Name)
	}
	if a.Max != -1 && a.Min > a.Max {
		return s, fmt.Errorf("args %s has min[%d] > max[%d], not parsing argument", a.Name, a.Min, a.Max)
	}

	count := 0
	var vc VC
	var t T
	value := vc.Create(a.Value, &t, a.Config)
	a.values = []T{}

	tracef("attempting arg%[1] parse", &a.Name)
	for _, arg := range s {
		if err := value.Set(arg); err != nil {
			return s, fmt.Errorf("invalid value %q for argument %s: %v", arg, a.Name, err)
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

// Get returns the argument value.
func (a *ArgumentsBase[T, C, VC]) Get() any {
	if a.values != nil {
		return a.values
	}
	return []T{}
}

type (
	// FloatArg defines a float64 positional argument.
	FloatArg = ArgumentBase[float64, NoConfig, floatValue[float64]]
	// Float32Arg defines a float32 positional argument.
	Float32Arg = ArgumentBase[float32, NoConfig, floatValue[float32]]
	// Float64Arg defines a float64 positional argument.
	Float64Arg = ArgumentBase[float64, NoConfig, floatValue[float64]]
	// IntArg defines a int positional argument.
	IntArg = ArgumentBase[int, IntegerConfig, intValue[int]]
	// Int8Arg defines a int8 positional argument.
	Int8Arg = ArgumentBase[int8, IntegerConfig, intValue[int8]]
	// Int16Arg defines a int16 positional argument.
	Int16Arg = ArgumentBase[int16, IntegerConfig, intValue[int16]]
	// Int32Arg defines a int32 positional argument.
	Int32Arg = ArgumentBase[int32, IntegerConfig, intValue[int32]]
	// Int64Arg defines a int64 positional argument.
	Int64Arg = ArgumentBase[int64, IntegerConfig, intValue[int64]]
	// StringArg defines a string positional argument.
	StringArg = ArgumentBase[string, StringConfig, stringValue]
	// StringMapArgs defines a map[string]string positional argument.
	StringMapArgs = ArgumentBase[map[string]string, StringConfig, StringMap]
	// TimestampArg defines a time.Time positional argument.
	TimestampArg = ArgumentBase[time.Time, TimestampConfig, timestampValue]
	// UintArg defines a uint positional argument.
	UintArg = ArgumentBase[uint, IntegerConfig, uintValue[uint]]
	// Uint8Arg defines a uint8 positional argument.
	Uint8Arg = ArgumentBase[uint8, IntegerConfig, uintValue[uint8]]
	// Uint16Arg defines a uint16 positional argument.
	Uint16Arg = ArgumentBase[uint16, IntegerConfig, uintValue[uint16]]
	// Uint32Arg defines a uint32 positional argument.
	Uint32Arg = ArgumentBase[uint32, IntegerConfig, uintValue[uint32]]
	// Uint64Arg defines a uint64 positional argument.
	Uint64Arg = ArgumentBase[uint64, IntegerConfig, uintValue[uint64]]

	// FloatArgs defines a sequence of float64 positional arguments.
	FloatArgs = ArgumentsBase[float64, NoConfig, floatValue[float64]]
	// Float32Args defines a sequence of float32 positional arguments.
	Float32Args = ArgumentsBase[float32, NoConfig, floatValue[float32]]
	// Float64Args defines a sequence of float64 positional arguments.
	Float64Args = ArgumentsBase[float64, NoConfig, floatValue[float64]]
	// IntArgs defines a sequence of int positional arguments.
	IntArgs = ArgumentsBase[int, IntegerConfig, intValue[int]]
	// Int8Args defines a sequence of int8 positional arguments.
	Int8Args = ArgumentsBase[int8, IntegerConfig, intValue[int8]]
	// Int16Args defines a sequence of int16 positional arguments.
	Int16Args = ArgumentsBase[int16, IntegerConfig, intValue[int16]]
	// Int32Args defines a sequence of int32 positional arguments.
	Int32Args = ArgumentsBase[int32, IntegerConfig, intValue[int32]]
	// Int64Args defines a sequence of int64 positional arguments.
	Int64Args = ArgumentsBase[int64, IntegerConfig, intValue[int64]]
	// StringArgs defines a sequence of string positional arguments.
	StringArgs = ArgumentsBase[string, StringConfig, stringValue]
	// TimestampArgs defines a sequence of time.Time positional arguments.
	TimestampArgs = ArgumentsBase[time.Time, TimestampConfig, timestampValue]
	// UintArgs defines a sequence of uint positional arguments.
	UintArgs = ArgumentsBase[uint, IntegerConfig, uintValue[uint]]
	// Uint8Args defines a sequence of uint8 positional arguments.
	Uint8Args = ArgumentsBase[uint8, IntegerConfig, uintValue[uint8]]
	// Uint16Args defines a sequence of uint16 positional arguments.
	Uint16Args = ArgumentsBase[uint16, IntegerConfig, uintValue[uint16]]
	// Uint32Args defines a sequence of uint32 positional arguments.
	Uint32Args = ArgumentsBase[uint32, IntegerConfig, uintValue[uint32]]
	// Uint64Args defines a sequence of uint64 positional arguments.
	Uint64Args = ArgumentsBase[uint64, IntegerConfig, uintValue[uint64]]
)

// getArgValue returns the value of the named positional argument,
// or nil if it is not found.
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

func arg[T any](name string, c *Command) T {
	val := c.getArgValue(name)
	if a, ok := val.(T); ok {
		return a
	}
	var zero T
	return zero
}

// StringArg returns the string value of the named positional argument,
// or an empty string if the argument is not found or has a different type.
func (c *Command) StringArg(name string) string {
	return arg[string](name, c)
}

// StringArgs returns the []string value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) StringArgs(name string) []string {
	return arg[[]string](name, c)
}

// FloatArg returns the float64 value of the named positional argument,
// or zero if the argument is not found or has a different type.
func (c *Command) FloatArg(name string) float64 {
	return arg[float64](name, c)
}

// FloatArgs returns the []float64 value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) FloatArgs(name string) []float64 {
	return arg[[]float64](name, c)
}

// Float32Arg returns the float32 value of the named positional argument,
// or zero if the argument is not found or has a different type.
func (c *Command) Float32Arg(name string) float32 {
	return arg[float32](name, c)
}

// Float32Args returns the []float32 value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) Float32Args(name string) []float32 {
	return arg[[]float32](name, c)
}

// Float64Arg returns the float64 value of the named positional argument,
// or zero if the argument is not found or has a different type.
func (c *Command) Float64Arg(name string) float64 {
	return arg[float64](name, c)
}

// Float64Args returns the []float64 value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) Float64Args(name string) []float64 {
	return arg[[]float64](name, c)
}

// IntArg returns the int value of the named positional argument,
// or zero if the argument is not found or has a different type.
func (c *Command) IntArg(name string) int {
	return arg[int](name, c)
}

// IntArgs returns the []int value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) IntArgs(name string) []int {
	return arg[[]int](name, c)
}

// Int8Arg returns the int8 value of the named positional argument,
// or zero if the argument is not found or has a different type.
func (c *Command) Int8Arg(name string) int8 {
	return arg[int8](name, c)
}

// Int8Args returns the []int8 value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) Int8Args(name string) []int8 {
	return arg[[]int8](name, c)
}

// Int16Arg returns the int16 value of the named positional argument,
// or zero if the argument is not found or has a different type.
func (c *Command) Int16Arg(name string) int16 {
	return arg[int16](name, c)
}

// Int16Args returns the []int16 value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) Int16Args(name string) []int16 {
	return arg[[]int16](name, c)
}

// Int32Arg returns the int32 value of the named positional argument,
// or zero if the argument is not found or has a different type.
func (c *Command) Int32Arg(name string) int32 {
	return arg[int32](name, c)
}

// Int32Args returns the []int32 value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) Int32Args(name string) []int32 {
	return arg[[]int32](name, c)
}

// Int64Arg returns the int64 value of the named positional argument,
// or zero if the argument is not found or has a different type.
func (c *Command) Int64Arg(name string) int64 {
	return arg[int64](name, c)
}

// Int64Args returns the []int64 value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) Int64Args(name string) []int64 {
	return arg[[]int64](name, c)
}

// UintArg returns the uint value of the named positional argument,
// or zero if the argument is not found or has a different type.
func (c *Command) UintArg(name string) uint {
	return arg[uint](name, c)
}

// Uint8Arg returns the uint8 value of the named positional argument,
// or zero if the argument is not found or has a different type.
func (c *Command) Uint8Arg(name string) uint8 {
	return arg[uint8](name, c)
}

// Uint16Arg returns the uint16 value of the named positional argument,
// or zero if the argument is not found or has a different type.
func (c *Command) Uint16Arg(name string) uint16 {
	return arg[uint16](name, c)
}

// Uint32Arg returns the uint32 value of the named positional argument,
// or zero if the argument is not found or has a different type.
func (c *Command) Uint32Arg(name string) uint32 {
	return arg[uint32](name, c)
}

// Uint64Arg returns the uint64 value of the named positional argument,
// or zero if the argument is not found or has a different type.
func (c *Command) Uint64Arg(name string) uint64 {
	return arg[uint64](name, c)
}

// UintArgs returns the []uint value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) UintArgs(name string) []uint {
	return arg[[]uint](name, c)
}

// Uint8Args returns the []uint8 value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) Uint8Args(name string) []uint8 {
	return arg[[]uint8](name, c)
}

// Uint16Args returns the []uint16 value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) Uint16Args(name string) []uint16 {
	return arg[[]uint16](name, c)
}

// Uint32Args returns the []uint32 value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) Uint32Args(name string) []uint32 {
	return arg[[]uint32](name, c)
}

// Uint64Args returns the []uint64 value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) Uint64Args(name string) []uint64 {
	return arg[[]uint64](name, c)
}

// TimestampArg returns the time.Time value of the named positional argument,
// or the zero time if the argument is not found or has a different type.
func (c *Command) TimestampArg(name string) time.Time {
	return arg[time.Time](name, c)
}

// TimestampArgs returns the []time.Time value of the named positional argument,
// or nil if the argument is not found or has a different type.
func (c *Command) TimestampArgs(name string) []time.Time {
	return arg[[]time.Time](name, c)
}
