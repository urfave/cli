package cli

import (
	"encoding/json"
	"flag"
	"fmt"
)

// Generic is a generic parseable type identified by a specific flag
type Generic = flag.Value

// JSONWrapGeneric accepts any type and wraps it to provide the
// Generic interface via JSON marshal/unmarshal
func JSONWrapGeneric(v interface{}) Generic {
	return &jsonGenericWrapper{v: v}
}

type jsonGenericWrapper struct {
	v interface{}
}

func (gw *jsonGenericWrapper) Set(value string) error {
	fromJSONStr := ""

	if err := json.Unmarshal([]byte(value), &fromJSONStr); err == nil {
		value = fromJSONStr
	}

	return json.Unmarshal([]byte(value), &gw.v)
}

func (gw *jsonGenericWrapper) String() string {
	vBytes, err := json.Marshal(gw.v)
	if err != nil {
		return fmt.Sprintf("%v", gw.v)
	}

	return string(vBytes)
}

// GenericFlag is a flag with type Generic
type GenericFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	TakesFile   bool
	Value       Generic
	DefaultText string
	HasBeenSet  bool
}

// IsSet returns whether or not the flag has been set through env or file
func (f *GenericFlag) IsSet() bool {
	return f.HasBeenSet
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *GenericFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *GenericFlag) Names() []string {
	return flagNames(f.Name, f.Aliases)
}

// IsRequired returns whether or not the flag is required
func (f *GenericFlag) IsRequired() bool {
	return f.Required
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *GenericFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *GenericFlag) GetUsage() string {
	return f.Usage
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *GenericFlag) GetValue() string {
	if f.Value != nil {
		return f.Value.String()
	}
	return ""
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *GenericFlag) IsVisible() bool {
	return !f.Hidden
}

// GetDefaultText returns the default text for this flag
func (f *GenericFlag) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return f.GetValue()
}

// GetEnvVars returns the env vars for this flag
func (f *GenericFlag) GetEnvVars() []string {
	return f.EnvVars
}

// Apply takes the flagset and calls Set on the generic flag with the value
// provided by the user for parsing by the flag
func (f *GenericFlag) Apply(set *flag.FlagSet) error {
	if val, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		if val != "" {
			if err := f.Value.Set(val); err != nil {
				return fmt.Errorf("could not parse %q as value for flag %s: %s", val, f.Name, err)
			}

			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		set.Var(f.Value, name, f.Usage)
	}

	return nil
}

// Get returns the flag’s value in the given Context. In the
// special case of types that are wrapped as with JSONWrapGeneric,
// the wrapped type value is returned.
func (f *GenericFlag) Get(ctx *Context) interface{} {
	v := ctx.Generic(f.Name)
	if wrapped, ok := v.(*jsonGenericWrapper); ok {
		return wrapped.v
	}

	return v
}

// Generic looks up the value of a local GenericFlag, returning nil
// if not found. In the special case of types that are wrapped as
// with JSONWrapGeneric, the wrapped type value is returned.
func (cCtx *Context) Generic(name string) interface{} {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupGeneric(name, fs)
	}
	return nil
}

func lookupGeneric(name string, set *flag.FlagSet) interface{} {
	f := set.Lookup(name)
	if f != nil {
		if wrapped, ok := f.Value.(*jsonGenericWrapper); ok {
			return wrapped.v
		}

		return f.Value
	}
	return nil
}
