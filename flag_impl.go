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

//type Float64Flag = flagImpl[float64, jsonValueFactory[float64]]

// type Float64SliceFlag = flagImpl[[]float64, jsonSliceValueFactory[[]float64]]
// type Float64Slice = []float64

//type Int64Flag = flagImpl[int64, jsonValueFactory[int64]]
//type UintFlag = flagImpl[uint, jsonValueFactory[uint]]
//type Uint64Flag = flagImpl[uint64, jsonValueFactory[uint64]]

func newSlice[T any](elem ...T) []T {
	var t []T
	t = append(t, elem...)
	return t
}

/*func NewFloat64Slice(elem ...float64) []float64 {
	return newSlice[float64](elem...)
}*/

//type StringFlag = flagImpl[string]

// Int looks up the value of a local IntFlag, returns
// 0 if not found
/*func (cCtx *Context) Float64(name string) float64 {
	if v, ok := cCtx.Value(name).(float64); ok {
		return v
	}
	return 0
}*/

// Int looks up the value of a local IntFlag, returns
// 0 if not found
/*func (cCtx *Context) Float64Slice(name string) []float64 {
	if v, ok := cCtx.Value(name).([]float64); ok {
		return v
	}
	return nil
}*/

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (cCtx *Context) Int(name string) int {
	if v, ok := cCtx.Value(name).(int); ok {
		return v
	}
	return 0
}

// Int64 looks up the value of a local Int64Flag, returns
// 0 if not found
/*func (cCtx *Context) Int64(name string) int64 {
	if v, ok := cCtx.Value(name).(int64); ok {
		return v
	}
	return 0
}*/

// Int looks up the value of a local IntFlag, returns
// 0 if not found
/*func (cCtx *Context) Uint(name string) uint {
	if v, ok := cCtx.Value(name).(uint); ok {
		return v
	}
	return 0
}*/

// Int64 looks up the value of a local Int64Flag, returns
// 0 if not found
/*func (cCtx *Context) Uint64(name string) uint64 {
	if v, ok := cCtx.Value(name).(uint64); ok {
		return v
	}
	return 0
}*/

// String looks up the value of a local StringFlag, returns
// "" if not found
/*func (cCtx *Context) String(name string) string {
	if v, ok := cCtx.Value(name).(string); ok {
		return v
	}
	return ""
}*/
