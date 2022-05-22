package cli

import (
	"flag"
	"fmt"
	"strconv"
)

// TakesValue returns true of the flag takes a value, otherwise false
func (f *BoolFlag) TakesValue() bool {
	return false
}

// GetUsage returns the usage string for the flag
func (f *BoolFlag) GetUsage() string {
	return f.Usage
}

// GetCategory returns the category for the flag
func (f *BoolFlag) GetCategory() string {
	return f.Category
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *BoolFlag) GetValue() string {
	return ""
}

// GetDefaultText returns the default text for this flag
func (f *BoolFlag) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return fmt.Sprintf("%v", f.Value)
}

// GetEnvVars returns the env vars for this flag
func (f *BoolFlag) GetEnvVars() []string {
	return f.EnvVars
}

// Apply populates the flag given the flag set and environment
func (f *BoolFlag) Apply(set *flag.FlagSet) error {
	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		if val != "" {
			valBool, err := strconv.ParseBool(val)

			if err != nil {
				return fmt.Errorf("could not parse %q as bool value from %s for flag %s: %s", val, source, f.Name, err)
			}

			f.Value = valBool
			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.BoolVar(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.Bool(name, f.Value, f.Usage)
	}

	return nil
}

// Get returns the flag’s value in the given Context.
func (f *BoolFlag) Get(ctx *Context) bool {
	return ctx.Bool(f.Name)
}

// Bool looks up the value of a local BoolFlag, returns
// false if not found
func (cCtx *Context) Bool(name string) bool {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupBool(name, fs)
	}
	return false
}

func lookupBool(name string, set *flag.FlagSet) bool {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := strconv.ParseBool(f.Value.String())
		if err != nil {
			return false
		}
		return parsed
	}
	return false
}
