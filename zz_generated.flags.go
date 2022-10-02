// WARNING: this file is generated. DO NOT EDIT

package cli

import "time"

// Float64SliceFlag is a flag with type *Float64Slice
type Float64SliceFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       *Float64Slice
	Destination *Float64Slice

	Aliases []string
	EnvVars []string

	defaultValue *Float64Slice

	Action func(*Context, []float64) error
}

// IsSet returns whether or not the flag has been set through env or file
func (f *Float64SliceFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *Float64SliceFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *Float64SliceFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *Float64SliceFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *Float64SliceFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *Float64SliceFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *Float64SliceFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *Float64SliceFlag) TakesValue() bool {
	return "Float64SliceFlag" != "BoolFlag"
}

// GenericFlag is a flag with type Generic
type GenericFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       Generic
	Destination Generic

	Aliases []string
	EnvVars []string

	defaultValue Generic

	TakesFile bool

	Action func(*Context, interface{}) error
}

// String returns a readable representation of this value (for usage defaults)
func (f *GenericFlag) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *GenericFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *GenericFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *GenericFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *GenericFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *GenericFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *GenericFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *GenericFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *GenericFlag) TakesValue() bool {
	return "GenericFlag" != "BoolFlag"
}

// Int64SliceFlag is a flag with type *Int64Slice
type Int64SliceFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       *Int64Slice
	Destination *Int64Slice

	Aliases []string
	EnvVars []string

	defaultValue *Int64Slice

	Action func(*Context, []int64) error
}

// IsSet returns whether or not the flag has been set through env or file
func (f *Int64SliceFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *Int64SliceFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *Int64SliceFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *Int64SliceFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *Int64SliceFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *Int64SliceFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *Int64SliceFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *Int64SliceFlag) TakesValue() bool {
	return "Int64SliceFlag" != "BoolFlag"
}

// IntSliceFlag is a flag with type *IntSlice
type IntSliceFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       *IntSlice
	Destination *IntSlice

	Aliases []string
	EnvVars []string

	defaultValue *IntSlice

	Action func(*Context, []int) error
}

// IsSet returns whether or not the flag has been set through env or file
func (f *IntSliceFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *IntSliceFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *IntSliceFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *IntSliceFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *IntSliceFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *IntSliceFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *IntSliceFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *IntSliceFlag) TakesValue() bool {
	return "IntSliceFlag" != "BoolFlag"
}

// PathFlag is a flag with type Path
type PathFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       Path
	Destination *Path

	Aliases []string
	EnvVars []string

	defaultValue Path

	TakesFile bool

	Action func(*Context, Path) error
}

// String returns a readable representation of this value (for usage defaults)
func (f *PathFlag) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *PathFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *PathFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *PathFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *PathFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *PathFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *PathFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *PathFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *PathFlag) TakesValue() bool {
	return "PathFlag" != "BoolFlag"
}

// StringSliceFlag is a flag with type *StringSlice
type StringSliceFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       *StringSlice
	Destination *StringSlice

	Aliases []string
	EnvVars []string

	defaultValue *StringSlice

	TakesFile bool

	Action func(*Context, []string) error
}

// IsSet returns whether or not the flag has been set through env or file
func (f *StringSliceFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *StringSliceFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *StringSliceFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *StringSliceFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *StringSliceFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *StringSliceFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *StringSliceFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *StringSliceFlag) TakesValue() bool {
	return "StringSliceFlag" != "BoolFlag"
}

// TimestampFlag is a flag with type *Timestamp
type TimestampFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       *Timestamp
	Destination *Timestamp

	Aliases []string
	EnvVars []string

	defaultValue *Timestamp

	Layout string

	Timezone *time.Location

	Action func(*Context, *time.Time) error
}

// String returns a readable representation of this value (for usage defaults)
func (f *TimestampFlag) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *TimestampFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *TimestampFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *TimestampFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *TimestampFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *TimestampFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *TimestampFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *TimestampFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *TimestampFlag) TakesValue() bool {
	return "TimestampFlag" != "BoolFlag"
}

// Uint64SliceFlag is a flag with type *Uint64Slice
type Uint64SliceFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       *Uint64Slice
	Destination *Uint64Slice

	Aliases []string
	EnvVars []string

	defaultValue *Uint64Slice

	Action func(*Context, []uint64) error
}

// IsSet returns whether or not the flag has been set through env or file
func (f *Uint64SliceFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *Uint64SliceFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *Uint64SliceFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *Uint64SliceFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *Uint64SliceFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *Uint64SliceFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *Uint64SliceFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *Uint64SliceFlag) TakesValue() bool {
	return "Uint64SliceFlag" != "BoolFlag"
}

// UintSliceFlag is a flag with type *UintSlice
type UintSliceFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       *UintSlice
	Destination *UintSlice

	Aliases []string
	EnvVars []string

	defaultValue *UintSlice

	Action func(*Context, []uint) error
}

// IsSet returns whether or not the flag has been set through env or file
func (f *UintSliceFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *UintSliceFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *UintSliceFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *UintSliceFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *UintSliceFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *UintSliceFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *UintSliceFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *UintSliceFlag) TakesValue() bool {
	return "UintSliceFlag" != "BoolFlag"
}

// vim:ro
