package cli

import (
	"flag"
	"strconv"
	"time"
)

// WARNING: This file is generated!

// BoolFlag is a flag with type bool
type BoolFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       bool
	DefaultText string

	Destination *bool
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *BoolFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *BoolFlag) Names() []string {
	return flagNames(f)
}

// Bool looks up the value of a local BoolFlag, returns
// false if not found
func (c *Context) Bool(name string) bool {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupBool(name, fs)
	}
	return false
}

func lookupBool(name string, set *flag.FlagSet) bool {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := strconv.ParseBool(f.Value.String())
		if err != nil {
			return false
		}
		return parsed
	}
	return false
}

// DurationFlag is a flag with type time.Duration (see https://golang.org/pkg/time/#ParseDuration)
type DurationFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       time.Duration
	DefaultText string

	Destination *time.Duration
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *DurationFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *DurationFlag) Names() []string {
	return flagNames(f)
}

// Duration looks up the value of a local DurationFlag, returns
// 0 if not found
func (c *Context) Duration(name string) time.Duration {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupDuration(name, fs)
	}
	return 0
}

func lookupDuration(name string, set *flag.FlagSet) time.Duration {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := time.ParseDuration(f.Value.String())
		if err != nil {
			return 0
		}
		return parsed
	}
	return 0
}

// Float64Flag is a flag with type float64
type Float64Flag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       float64
	DefaultText string

	Destination *float64
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Float64Flag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *Float64Flag) Names() []string {
	return flagNames(f)
}

// Float64 looks up the value of a local Float64Flag, returns
// 0 if not found
func (c *Context) Float64(name string) float64 {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupFloat64(name, fs)
	}
	return 0
}

func lookupFloat64(name string, set *flag.FlagSet) float64 {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := strconv.ParseFloat(f.Value.String(), 64)
		if err != nil {
			return 0
		}
		return parsed
	}
	return 0
}

// GenericFlag is a flag with type Generic
type GenericFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       Generic
	DefaultText string
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *GenericFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *GenericFlag) Names() []string {
	return flagNames(f)
}

// Generic looks up the value of a local GenericFlag, returns
// nil if not found
func (c *Context) Generic(name string) interface{} {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupGeneric(name, fs)
	}
	return nil
}

func lookupGeneric(name string, set *flag.FlagSet) interface{} {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := f.Value, error(nil)
		if err != nil {
			return nil
		}
		return parsed
	}
	return nil
}

// Int64Flag is a flag with type int64
type Int64Flag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       int64
	DefaultText string

	Destination *int64
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Int64Flag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *Int64Flag) Names() []string {
	return flagNames(f)
}

// Int64 looks up the value of a local Int64Flag, returns
// 0 if not found
func (c *Context) Int64(name string) int64 {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupInt64(name, fs)
	}
	return 0
}

func lookupInt64(name string, set *flag.FlagSet) int64 {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := strconv.ParseInt(f.Value.String(), 0, 64)
		if err != nil {
			return 0
		}
		return parsed
	}
	return 0
}

// IntFlag is a flag with type int
type IntFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       int
	DefaultText string

	Destination *int
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *IntFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *IntFlag) Names() []string {
	return flagNames(f)
}

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (c *Context) Int(name string) int {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupInt(name, fs)
	}
	return 0
}

func lookupInt(name string, set *flag.FlagSet) int {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := strconv.ParseInt(f.Value.String(), 0, 64)
		if err != nil {
			return 0
		}
		return int(parsed)
	}
	return 0
}

// IntSliceFlag is a flag with type *IntSlice
type IntSliceFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       *IntSlice
	DefaultText string
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *IntSliceFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *IntSliceFlag) Names() []string {
	return flagNames(f)
}

// IntSlice looks up the value of a local IntSliceFlag, returns
// nil if not found
func (c *Context) IntSlice(name string) []int {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupIntSlice(name, fs)
	}
	return nil
}

func lookupIntSlice(name string, set *flag.FlagSet) []int {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := (f.Value.(*IntSlice)).Value(), error(nil)
		if err != nil {
			return nil
		}
		return parsed
	}
	return nil
}

// Int64SliceFlag is a flag with type *Int64Slice
type Int64SliceFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       *Int64Slice
	DefaultText string
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Int64SliceFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *Int64SliceFlag) Names() []string {
	return flagNames(f)
}

// Int64Slice looks up the value of a local Int64SliceFlag, returns
// nil if not found
func (c *Context) Int64Slice(name string) []int64 {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupInt64Slice(name, fs)
	}
	return nil
}

func lookupInt64Slice(name string, set *flag.FlagSet) []int64 {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := (f.Value.(*Int64Slice)).Value(), error(nil)
		if err != nil {
			return nil
		}
		return parsed
	}
	return nil
}

// Float64SliceFlag is a flag with type *Float64Slice
type Float64SliceFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       *Float64Slice
	DefaultText string
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Float64SliceFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *Float64SliceFlag) Names() []string {
	return flagNames(f)
}

// Float64Slice looks up the value of a local Float64SliceFlag, returns
// nil if not found
func (c *Context) Float64Slice(name string) []float64 {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupFloat64Slice(name, fs)
	}
	return nil
}

func lookupFloat64Slice(name string, set *flag.FlagSet) []float64 {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := (f.Value.(*Float64Slice)).Value(), error(nil)
		if err != nil {
			return nil
		}
		return parsed
	}
	return nil
}

// StringFlag is a flag with type string
type StringFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       string
	DefaultText string

	Destination *string
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *StringFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *StringFlag) Names() []string {
	return flagNames(f)
}

// String looks up the value of a local StringFlag, returns
// "" if not found
func (c *Context) String(name string) string {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupString(name, fs)
	}
	return ""
}

func lookupString(name string, set *flag.FlagSet) string {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := f.Value.String(), error(nil)
		if err != nil {
			return ""
		}
		return parsed
	}
	return ""
}

// PathFlag is a flag with type string
type PathFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       string
	DefaultText string

	Destination *string
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *PathFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *PathFlag) Names() []string {
	return flagNames(f)
}

// Path looks up the value of a local PathFlag, returns
// "" if not found
func (c *Context) Path(name string) string {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupPath(name, fs)
	}
	return ""
}

func lookupPath(name string, set *flag.FlagSet) string {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := f.Value.String(), error(nil)
		if err != nil {
			return ""
		}
		return parsed
	}
	return ""
}

// StringSliceFlag is a flag with type *StringSlice
type StringSliceFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       *StringSlice
	DefaultText string
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *StringSliceFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *StringSliceFlag) Names() []string {
	return flagNames(f)
}

// StringSlice looks up the value of a local StringSliceFlag, returns
// nil if not found
func (c *Context) StringSlice(name string) []string {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupStringSlice(name, fs)
	}
	return nil
}

func lookupStringSlice(name string, set *flag.FlagSet) []string {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := (f.Value.(*StringSlice)).Value(), error(nil)
		if err != nil {
			return nil
		}
		return parsed
	}
	return nil
}

// Uint64Flag is a flag with type uint64
type Uint64Flag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       uint64
	DefaultText string

	Destination *uint64
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Uint64Flag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *Uint64Flag) Names() []string {
	return flagNames(f)
}

// Uint64 looks up the value of a local Uint64Flag, returns
// 0 if not found
func (c *Context) Uint64(name string) uint64 {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupUint64(name, fs)
	}
	return 0
}

func lookupUint64(name string, set *flag.FlagSet) uint64 {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := strconv.ParseUint(f.Value.String(), 0, 64)
		if err != nil {
			return 0
		}
		return parsed
	}
	return 0
}

// UintFlag is a flag with type uint
type UintFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Hidden      bool
	Value       uint
	DefaultText string

	Destination *uint
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *UintFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *UintFlag) Names() []string {
	return flagNames(f)
}

// Uint looks up the value of a local UintFlag, returns
// 0 if not found
func (c *Context) Uint(name string) uint {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupUint(name, fs)
	}
	return 0
}

func lookupUint(name string, set *flag.FlagSet) uint {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := strconv.ParseUint(f.Value.String(), 0, 64)
		if err != nil {
			return 0
		}
		return uint(parsed)
	}
	return 0
}
