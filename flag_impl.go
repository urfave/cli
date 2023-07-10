package cli

import (
	"context"
	"flag"
	"fmt"
	"reflect"

	"github.com/urfave/argh"
)

var (
	ErrFlagCannotBeDuplicated = fmt.Errorf("flag cannot be duplicated: %[1]w", Err)
)

// Value represents a value as used by cli. It is essentially the
// same interface as flag.Getter, which embeds flag.Value.
type Value interface {
	String() string
	Set(string) error
	Get() any
}

type boolFlag interface {
	IsBoolFlag() bool
}

type fnValue struct {
	fn     func(string) error
	isBool bool
	v      Value
}

func (f *fnValue) Get() any           { return f.v.Get() }
func (f *fnValue) Set(s string) error { return f.fn(s) }
func (f *fnValue) String() string {
	if f.v == nil {
		return ""
	}
	return f.v.String()
}

func (f *fnValue) Serialize() string {
	if s, ok := f.v.(Serializer); ok {
		return s.Serialize()
	}
	return f.v.String()
}

func (f *fnValue) IsBoolFlag() bool { return f.isBool }
func (f *fnValue) Count() int {
	if s, ok := f.v.(Countable); ok {
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

	Category    string // category of the flag, if any
	DefaultText string // default text of the flag for usage purposes
	Usage       string // usage string for help output

	Sources ValueSourceChain // sources to load flag value from

	Required   bool // whether the flag is required or not
	Hidden     bool // whether to hide the flag in help output
	Persistent bool // whether the flag needs to be applied to subcommands as well

	Value       T  // default value for this flag if not set by from any source
	Destination *T // destination pointer for value when set

	Aliases []string // Aliases that are allowed for this flag

	TakesFile bool // whether this flag takes a file argument, mainly for shell completion purposes

	Action func(context.Context, *Command, T) error // Action callback to be called when flag is set

	Config C // Additional/Custom configuration associated with this flag type

	OnlyOnce bool // whether this flag can be duplicated on the command line

	Validator func(T) error // custom function to validate this flag value

	// unexported fields for internal use
	count      int   // number of times the flag has been set
	hasBeenSet bool  // whether the flag has been set from env or file
	applied    bool  // whether the flag has been applied to a flag set already
	creator    VC    // value creator for this flag type
	v          Value // value representing this flag's value
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *FlagBase[T, C, V]) GetValue() string {
	if reflect.TypeOf(f.Value).Kind() == reflect.Bool {
		return ""
	}
	return fmt.Sprintf("%v", f.Value)
}

func (f *FlagBase[T, C, V]) Set(val string) error {
	return f.v.Set(val)
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

		if val, source, found := f.Sources.LookupWithSource(); found {
			tmpVal := f.creator.Create(f.Value, new(T), f.Config)
			if val != "" || reflect.TypeOf(f.Value).Kind() == reflect.String {
				if err := tmpVal.Set(val); err != nil {
					return fmt.Errorf(
						"could not parse %[1]q as %[2]T value from %[3]s for flag %[4]s: %[5]s",
						val, f.Value, source, f.Name, err,
					)
				}
			} else if val == "" && reflect.TypeOf(f.Value).Kind() == reflect.Bool {
				val = "false"
				if err := tmpVal.Set(val); err != nil {
					return fmt.Errorf(
						"could not parse %[1]q as %[2]T value from %[3]s for flag %[4]s: %[5]s",
						val, f.Value, source, f.Name, err,
					)
				}
			}

			newVal = tmpVal.Get().(T)
			f.hasBeenSet = true
		}

		if f.Destination == nil {
			f.v = f.creator.Create(newVal, new(T), f.Config)
		} else {
			f.v = f.creator.Create(newVal, f.Destination, f.Config)
		}

		// Validate the given default or values set from external sources as well
		if f.Validator != nil {
			if v, ok := f.v.Get().(T); !ok {
				return &typeError[T]{
					other: f.v.Get(),
				}
			} else if err := f.Validator(v); err != nil {
				return err
			}
		}
	}

	isBool := false
	if b, ok := f.v.(boolFlag); ok && b.IsBoolFlag() {
		isBool = true
	}

	for _, name := range f.Names() {
		set.Var(&fnValue{
			fn: func(val string) error {
				if f.count == 1 && f.OnlyOnce {
					return fmt.Errorf("cant duplicate this flag")
				}
				f.count++
				if err := f.v.Set(val); err != nil {
					return err
				}
				if f.Validator != nil {
					if v, ok := f.v.Get().(T); !ok {
						return &typeError[T]{
							other: f.v.Get(),
						}
					} else if err := f.Validator(v); err != nil {
						return err
					}
				}
				return nil
			},
			isBool: isBool,
			v:      f.v,
		}, name, f.Usage)
	}

	f.applied = true
	return nil
}

// ApplyWithArgh populates the flag given the command config
func (f *FlagBase[T, C, V]) ApplyWithArgh(cmd *Command) error {
	// TODO move this phase into a separate flag initialization function
	// if flag has been applied previously then it would have already been set
	// from env or file. So no need to apply the env set again. However
	// lots of units tests prior to persistent flags assumed that the
	// flag can be applied to different flag sets multiple times while still
	// keeping the env set.
	if !f.applied || !f.Persistent {
		newVal := f.Value

		if val, source, found := f.Sources.LookupWithSource(); found {
			tmpVal := f.creator.Create(f.Value, new(T), f.Config)
			if val != "" || reflect.TypeOf(f.Value).Kind() == reflect.String {
				if err := tmpVal.Set(val); err != nil {
					return fmt.Errorf(
						"could not parse %[1]q as %[2]T value from %[3]s for flag %[4]s: %[5]w",
						val, f.Value, source, f.Name, err,
					)
				}
			} else if val == "" && reflect.TypeOf(f.Value).Kind() == reflect.Bool {
				val = "false"
				if err := tmpVal.Set(val); err != nil {
					return fmt.Errorf(
						"could not parse %[1]q as %[2]T value from %[3]s for flag %[4]s: %[5]w",
						val, f.Value, source, f.Name, err,
					)
				}
			}

			newVal = tmpVal.Get().(T)
			f.hasBeenSet = true
		}

		if f.Destination == nil {
			f.v = f.creator.Create(newVal, new(T), f.Config)
		} else {
			f.v = f.creator.Create(newVal, f.Destination, f.Config)
		}

		tracef("setting flag value in command value map (flag=%[1]q)", f.CanonicalName())
		cmd.values[f.CanonicalName()] = f.v

		// Validate the given default or values set from external sources as well
		if f.Validator != nil {
			if v, ok := f.v.Get().(T); !ok {
				return &typeError[T]{
					other: f.v.Get(),
				}
			} else if err := f.Validator(v); err != nil {
				return err
			}
		}
	}

	flagNValue := argh.NValue(0)

	if f.TakesValue() {
		flagNValue = argh.NValue(1)
	}

	if dfl, ok := f.v.(DocGenerationMultiValueFlag); ok && dfl.IsMultiValueFlag() {
		flagNValue = argh.OneOrMoreValue
	}

	tracef("using flagNValue=%[1]v (flag=%[2]q)", flagNValue, f.CanonicalName())

	flCfg := &argh.FlagConfig{
		NValue: flagNValue,
		On: func(cf argh.CommandFlag) error {
			tracef("setting via On func (flag=%[1]q)", f.CanonicalName())

			if f.count == 1 && f.OnlyOnce {
				return fmt.Errorf("%[1]q: %[2]w", f.CanonicalName(), ErrFlagCannotBeDuplicated)
			}

			f.count++

			for _, val := range stringMapToSlice(cf.Values) {
				if err := f.v.Set(val); err != nil {
					return err
				}
			}

			if f.Validator != nil {
				if v, ok := f.v.Get().(T); !ok {
					return &typeError[T]{
						other: f.v.Get(),
					}
				} else if err := f.Validator(v); err != nil {
					return err
				}
			}

			return nil
		},
	}

	for _, name := range f.Names() {
		tracef("setting flag config with name=%[1]q cfg=%[2]v", name, flCfg)
		cmd.cfg.SetFlagConfig(name, flCfg)
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

func (f *FlagBase[T, C, V]) CanonicalName() string {
	return f.Name
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
	vals := []string{}

	for _, src := range f.Sources.Chain {
		if v, ok := src.(*envVarValueSource); ok {
			vals = append(vals, v.Key)
		}
	}

	return vals
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

// Get returns the flag’s value in the given Command.
func (f *FlagBase[T, C, V]) Get(cmd *Command) T {
	if v, ok := cmd.Value(f.Name).(T); ok {
		return v
	}
	var t T
	return t
}

// RunAction executes flag action if set
func (f *FlagBase[T, C, V]) RunAction(ctx context.Context, cmd *Command) error {
	if f.Action != nil {
		return f.Action(ctx, cmd, f.Get(cmd))
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
