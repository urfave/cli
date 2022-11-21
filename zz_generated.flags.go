// WARNING: this file is generated. DO NOT EDIT

package cli

import "time"

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

// vim:ro
