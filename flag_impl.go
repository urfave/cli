package cli

import (
	"flag"
	"fmt"
)

type ValueCreator[T any] interface {
	Create(t T, d *T) flag.Value
}

// Float64Flag is a flag with type float64
type flagImpl[T any, F ValueCreator[T]] struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       T
	Destination *T

	Aliases []string
	EnvVars []string

	TakesFile bool
	Base      int
	Count     *int

	creator F
	value   flag.Value
	Action  func(*Context, T) error
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *flagImpl[T, V]) GetValue() string {
	return fmt.Sprintf("%v", f.Value)
}

// Apply populates the flag given the flag set and environment
func (f *flagImpl[T, V]) Apply(set *flag.FlagSet) error {
	if f.Destination == nil {
		f.value = f.creator.Create(f.Value, new(T))
	}

	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		if val != "" {
			if err := f.value.Set(val); err != nil {
				return fmt.Errorf("could not parse %q as %T value from %s for flag %s: %s", val, f.Value, source, f.Name, err)
			}
			f.Value = f.value.(flag.Getter).Get().(T)
			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			f.value = f.creator.Create(f.Value, f.Destination)
			set.Var(f.value, name, f.Usage)
			continue
		}
		set.Var(f.value, name, f.Usage)
	}

	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (f *flagImpl[T, V]) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *flagImpl[T, V]) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *flagImpl[T, V]) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *flagImpl[T, V]) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *flagImpl[T, V]) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *flagImpl[T, V]) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *flagImpl[T, V]) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *flagImpl[T, V]) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *flagImpl[T, V]) TakesValue() bool {
	return "Float64Flag" != "BoolFlag"
}

// GetDefaultText returns the default text for this flag
func (f *flagImpl[T, V]) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return f.GetValue()
}

// Get returns the flag’s value in the given Context.
func (f *flagImpl[T, V]) Get(ctx *Context) T {
	if v, ok := ctx.Value(f.Name).(T); ok {
		return v
	}
	var t T
	return t
}
