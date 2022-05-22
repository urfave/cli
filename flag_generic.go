package cli

import (
	"flag"
	"fmt"
)

// Generic is a generic parseable type identified by a specific flag
type Generic interface {
	Set(value string) error
	String() string
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *GenericFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *GenericFlag) GetUsage() string {
	return f.Usage
}

// GetCategory returns the category for the flag
func (f *GenericFlag) GetCategory() string {
	return f.Category
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *GenericFlag) GetValue() string {
	if f.Value != nil {
		return f.Value.String()
	}
	return ""
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
func (f GenericFlag) Apply(set *flag.FlagSet) error {
	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		if val != "" {
			if err := f.Value.Set(val); err != nil {
				return fmt.Errorf("could not parse %q from %s as value for flag %s: %s", val, source, f.Name, err)
			}

			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		set.Var(f.Value, name, f.Usage)
	}

	return nil
}

// Get returns the flag’s value in the given Context.
func (f *GenericFlag) Get(ctx *Context) interface{} {
	return ctx.Generic(f.Name)
}

// Generic looks up the value of a local GenericFlag, returns
// nil if not found
func (cCtx *Context) Generic(name string) interface{} {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupGeneric(name, fs)
	}
	return nil
}

func lookupGeneric(name string, set *flag.FlagSet) interface{} {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := f.Value, error(nil)
		if err != nil {
			return nil
		}
		return parsed
	}
	return nil
}
