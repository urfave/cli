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
	Destination *Generic

	Aliases []string
	EnvVars []string

	TakesFile bool
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

	TakesFile bool
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

	TakesFile bool
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

	Layout string

	Timezone *time.Location
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

// BoolFlag is a flag with type bool
type BoolFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       bool
	Destination *bool

	Aliases []string
	EnvVars []string
}

// String returns a readable representation of this value (for usage defaults)
func (f *BoolFlag) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *BoolFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *BoolFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *BoolFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *BoolFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *BoolFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *BoolFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *BoolFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *BoolFlag) TakesValue() bool {
	return "BoolFlag" != "BoolFlag"
}

// Float64Flag is a flag with type float64
type Float64Flag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       float64
	Destination *float64

	Aliases []string
	EnvVars []string
}

// String returns a readable representation of this value (for usage defaults)
func (f *Float64Flag) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *Float64Flag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *Float64Flag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *Float64Flag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *Float64Flag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *Float64Flag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *Float64Flag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *Float64Flag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *Float64Flag) TakesValue() bool {
	return "Float64Flag" != "BoolFlag"
}

// IntFlag is a flag with type int
type IntFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       int
	Destination *int

	Aliases []string
	EnvVars []string
}

// String returns a readable representation of this value (for usage defaults)
func (f *IntFlag) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *IntFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *IntFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *IntFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *IntFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *IntFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *IntFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *IntFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *IntFlag) TakesValue() bool {
	return "IntFlag" != "BoolFlag"
}

// Int64Flag is a flag with type int64
type Int64Flag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       int64
	Destination *int64

	Aliases []string
	EnvVars []string
}

// String returns a readable representation of this value (for usage defaults)
func (f *Int64Flag) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *Int64Flag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *Int64Flag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *Int64Flag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *Int64Flag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *Int64Flag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *Int64Flag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *Int64Flag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *Int64Flag) TakesValue() bool {
	return "Int64Flag" != "BoolFlag"
}

// StringFlag is a flag with type string
type StringFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       string
	Destination *string

	Aliases []string
	EnvVars []string

	TakesFile bool
}

// String returns a readable representation of this value (for usage defaults)
func (f *StringFlag) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *StringFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *StringFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *StringFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *StringFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *StringFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *StringFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *StringFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *StringFlag) TakesValue() bool {
	return "StringFlag" != "BoolFlag"
}

// DurationFlag is a flag with type time.Duration
type DurationFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       time.Duration
	Destination *time.Duration

	Aliases []string
	EnvVars []string
}

// String returns a readable representation of this value (for usage defaults)
func (f *DurationFlag) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *DurationFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *DurationFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *DurationFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *DurationFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *DurationFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *DurationFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *DurationFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *DurationFlag) TakesValue() bool {
	return "DurationFlag" != "BoolFlag"
}

// UintFlag is a flag with type uint
type UintFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       uint
	Destination *uint

	Aliases []string
	EnvVars []string
}

// String returns a readable representation of this value (for usage defaults)
func (f *UintFlag) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *UintFlag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *UintFlag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *UintFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *UintFlag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *UintFlag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *UintFlag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *UintFlag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *UintFlag) TakesValue() bool {
	return "UintFlag" != "BoolFlag"
}

// Uint64Flag is a flag with type uint64
type Uint64Flag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       uint64
	Destination *uint64

	Aliases []string
	EnvVars []string
}

// String returns a readable representation of this value (for usage defaults)
func (f *Uint64Flag) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *Uint64Flag) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *Uint64Flag) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *Uint64Flag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *Uint64Flag) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *Uint64Flag) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *Uint64Flag) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *Uint64Flag) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *Uint64Flag) TakesValue() bool {
	return "Uint64Flag" != "BoolFlag"
}

// vim:ro
