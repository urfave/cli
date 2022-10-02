package cli

import (
	"encoding/json"
	"flag"
	"fmt"
)

type flagValue[T any] struct {
	t *T
}

func (v flagValue[T]) String() string {
	if b, err := json.Marshal(v.t); err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (v flagValue[T]) Set(s string) error {
	return json.Unmarshal([]byte(s), v.t)
}

func (v flagValue[T]) Get() any {
	return *v.t
}

// Float64Flag is a flag with type float64
type flagImpl[T any] struct {
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

	value flagValue[T]
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *flagImpl[T]) GetValue() string {
	return fmt.Sprintf("%v", f.Value)
}

// Apply populates the flag given the flag set and environment
func (f *flagImpl[T]) Apply(set *flag.FlagSet) error {
	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		if val != "" {
			f.value.t = new(T)
			if err := f.value.Set(val); err != nil {
				return fmt.Errorf("could not parse %q as %T value from %s for flag %s: %s", val, f.Value, source, f.Name, err)
			}

			f.Value = *f.value.t
			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			f.value.t = f.Destination
			*f.Destination = f.Value
			set.Var(f.value, name, f.Usage)
			continue
		}
		if f.value.t == nil {
			f.value.t = new(T)
			*f.value.t = f.Value
		}
		set.Var(f.value, name, f.Usage)
	}

	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (f *flagImpl[T]) String() string {
	return FlagStringer(f)
}

// IsSet returns whether or not the flag has been set through env or file
func (f *flagImpl[T]) IsSet() bool {
	return f.HasBeenSet
}

// Names returns the names of the flag
func (f *flagImpl[T]) Names() []string {
	return FlagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *flagImpl[T]) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *flagImpl[T]) IsVisible() bool {
	return !f.Hidden
}

// GetCategory returns the category of the flag
func (f *flagImpl[T]) GetCategory() string {
	return f.Category
}

// GetUsage returns the usage string for the flag
func (f *flagImpl[T]) GetUsage() string {
	return f.Usage
}

// GetEnvVars returns the env vars for this flag
func (f *flagImpl[T]) GetEnvVars() []string {
	return f.EnvVars
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *flagImpl[T]) TakesValue() bool {
	return "Float64Flag" != "BoolFlag"
}

// GetDefaultText returns the default text for this flag
func (f *flagImpl[T]) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return f.GetValue()
}

// Get returns the flag’s value in the given Context.
func (f *flagImpl[T]) Get(ctx *Context) T {
	if v, ok := ctx.Value(f.Name).(T); ok {
		return v
	}
	var t T
	return t
}

type Float64Flag = flagImpl[float64]
type IntFlag = flagImpl[int]
type Int64Flag = flagImpl[int64]
type UintFlag = flagImpl[uint]
type Uint64Flag = flagImpl[uint64]

//type StringFlag = flagImpl[string]

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (cCtx *Context) Float64(name string) float64 {
	if v, ok := cCtx.Value(name).(float64); ok {
		return v
	}
	return 0
}

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
func (cCtx *Context) Int64(name string) int64 {
	if v, ok := cCtx.Value(name).(int64); ok {
		return v
	}
	return 0
}

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (cCtx *Context) Uint(name string) uint {
	if v, ok := cCtx.Value(name).(uint); ok {
		return v
	}
	return 0
}

// Int64 looks up the value of a local Int64Flag, returns
// 0 if not found
func (cCtx *Context) Uint64(name string) uint64 {
	if v, ok := cCtx.Value(name).(uint64); ok {
		return v
	}
	return 0
}

// String looks up the value of a local StringFlag, returns
// "" if not found
/*func (cCtx *Context) String(name string) string {
	if v, ok := cCtx.Value(name).(string); ok {
		return v
	}
	return ""
}*/
