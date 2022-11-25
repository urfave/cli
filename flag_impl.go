package cli

import (
	"flag"
	"fmt"
	"reflect"
)

type Value interface {
	flag.Value
	flag.Getter
}

// ValueCreator is responsible for creating a flag.Value emulation
// as well as custom formatting
//
//	T specifies the type
//	C specifies the config for the type
type ValueCreator[T any, C any] interface {
	Create(T, *T, C) Value
	ToString(T) string
}

// NoConfig is for flags which dont need a custom configuration
type NoConfig struct{}

// FlagBase[T,C,VC] is a generic flag base which can be used
// as a boilerplate to implement the most common interfaces
// used by urfave/cli.
//
//	T specifies the type
//	C specifies the configuration required(if any for that flag type)
//	VC specifies the value creator which creates the flag.Value emulation
type FlagBase[T any, C any, VC ValueCreator[T, C]] struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	hasBeenSet bool

	Value       T
	Destination *T

	Aliases []string
	EnvVars []string

	TakesFile bool

	Action func(*Context, T) error

	Config C

	creator VC
	value   flag.Value
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *FlagBase[T, C, V]) GetValue() string {
	if reflect.TypeOf(f.Value).Kind() == reflect.Bool {
		return ""
	}
	return fmt.Sprintf("%v", f.Value)
}

// Apply populates the flag given the flag set and environment
func (f *FlagBase[T, C, V]) Apply(set *flag.FlagSet) error {
	newVal := f.Value

	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		tmpVal := f.creator.Create(f.Value, new(T), f.Config)
		if val != "" || reflect.TypeOf(f.Value).Kind() == reflect.String {
			if err := tmpVal.Set(val); err != nil {
				return fmt.Errorf("could not parse %q as %T value from %s for flag %s: %s", val, f.Value, source, f.Name, err)
			}
		} else if val == "" && reflect.TypeOf(f.Value).Kind() == reflect.Bool {
			val = "false"
			if err := tmpVal.Set(val); err != nil {
				return fmt.Errorf("could not parse %q as %T value from %s for flag %s: %s", val, f.Value, source, f.Name, err)
			}
		}

		newVal = tmpVal.(flag.Getter).Get().(T)
		f.hasBeenSet = true
	}

	if f.Destination == nil {
		f.value = f.creator.Create(newVal, new(T), f.Config)
	} else {
		f.value = f.creator.Create(newVal, f.Destination, f.Config)
	}

	for _, name := range f.Names() {
		set.Var(f.value, name, f.Usage)
	}

	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (f *FlagBase[T, C, V]) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *FlagBase[T, C, V]) IsSet() bool {
	return f.hasBeenSet
}

// Names returns the names of the flag
func (f *FlagBase[T, C, V]) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *FlagBase[T, C, V]) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *FlagBase[T, C, V]) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *FlagBase[T, C, V]) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *FlagBase[T, C, V]) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *FlagBase[T, C, V]) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *FlagBase[T, C, V]) TakesValue() bool {
	var t T
	return reflect.TypeOf(t).Kind() != reflect.Bool
}

// GetDefaultText returns the default text for this flag
func (f *FlagBase[T, C, V]) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	var v V
	return v.ToString(f.Value)
}

// Get returns the flag’s value in the given Context.
func (f *FlagBase[T, C, V]) Get(ctx *Context) T {
	if v, ok := ctx.Value(f.Name).(T); ok {
		return v
	}
	var t T
	return t
}

// RunAction executes flag action if set
func (f *FlagBase[T, C, V]) RunAction(ctx *Context) error {
	if f.Action != nil {
		return f.Action(ctx, f.Get(ctx))
	}

	return nil
}

func (f *FlagBase[T, C, VC]) IsSliceFlag() bool {
	// TBD how to specify
	return reflect.TypeOf(f.Value).Kind() == reflect.Slice
}
