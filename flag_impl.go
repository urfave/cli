package cli

import (
	"flag"
	"fmt"
	"reflect"
)

// Value represents a value as used by cli.
// For now it implements the golang flag.Value interface
type Value interface {
	flag.Value
	flag.Getter
}

// simple wrapper to intercept Value operations
// to check for duplicates
type valueWrapper struct {
	value    Value
	count    int
	onlyOnce bool
}

func (v *valueWrapper) String() string {
	if v.value == nil {
		return ""
	}
	return v.value.String()
}

func (v *valueWrapper) Set(s string) error {
	if v.count == 1 && v.onlyOnce {
		return fmt.Errorf("cant duplicate this flag")
	}
	v.count++
	return v.value.Set(s)
}

func (v *valueWrapper) Get() any {
	return v.value.Get()
}

func (v *valueWrapper) IsBoolFlag() bool {
	_, ok := v.value.(*boolValue)
	return ok
}

func (v *valueWrapper) Serialize() string {
	if s, ok := v.value.(Serializer); ok {
		return s.Serialize()
	}
	return v.value.String()
}

func (v *valueWrapper) Count() int {
	if s, ok := v.value.(Countable); ok {
		return s.Count()
	}
	return 0
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
	Name string // name of the flag

	Category    string   // category of the flag, if any
	DefaultText string   // default text of the flag for usage purposes
	FilePaths   []string // file paths to load value from
	Usage       string   // usage string for help output

	Required   bool // whether the flag is required or not
	Hidden     bool // whether to hide the flag in help output
	Persistent bool // whether the flag needs to be applied to subcommands as well

	Value       T  // default value for this flag if not set by from any source
	Destination *T // destination pointer for value when set

	Aliases []string // Aliases that are allowed for this flag
	EnvVars []string // Set of environment variables that can influence this flag's value

	TakesFile bool // whether this flag takes a file argument, mainly for shell completion purposes

	Action func(*Context, T) error // Action callback to be called when flag is set

	Config C // Additional/Custom configuration associated with this flag type

	OnlyOnce bool // whether this flag can be duplicated on the command line

	// unexported fields for internal use
	hasBeenSet bool  // whether the flag has been set from env or file
	applied    bool  // whether the flag has been applied to a flag set already
	creator    VC    // value creator for this flag type
	value      Value // value representing this flag's value
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
	// TODO move this phase into a separate flag initialization function
	// if flag has been applied previously then it would have already been set
	// from env or file. So no need to apply the env set again. However
	// lots of units tests prior to persistent flags assumed that the
	// flag can be applied to different flag sets multiple times while still
	// keeping the env set.
	if !f.applied || !f.Persistent {
		newVal := f.Value

		if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePaths); found {
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

			newVal = tmpVal.Get().(T)
			f.hasBeenSet = true
		}

		if f.Destination == nil {
			f.value = f.creator.Create(newVal, new(T), f.Config)
		} else {
			f.value = f.creator.Create(newVal, f.Destination, f.Config)
		}
	}

	vw := &valueWrapper{
		value:    f.value,
		onlyOnce: f.OnlyOnce,
	}

	for _, name := range f.Names() {
		set.Var(vw, name, f.Usage)
	}

	f.applied = true
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

// IsMultiValueFlag returns true if the value type T can take multiple
// values from cmd line. This is true for slice and map type flags
func (f *FlagBase[T, C, VC]) IsMultiValueFlag() bool {
	// TBD how to specify
	kind := reflect.TypeOf(f.Value).Kind()
	return kind == reflect.Slice || kind == reflect.Map
}

// IsPersistent returns true if flag needs to be persistent across subcommands
func (f *FlagBase[T, C, VC]) IsPersistent() bool {
	return f.Persistent
}
