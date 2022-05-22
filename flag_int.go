package cli

import (
	"flag"
	"fmt"
	"strconv"
)

// TakesValue returns true of the flag takes a value, otherwise false
func (f *IntFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *IntFlag) GetUsage() string {
	return f.Usage
}

// GetCategory returns the category for the flag
func (f *IntFlag) GetCategory() string {
	return f.Category
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *IntFlag) GetValue() string {
	return fmt.Sprintf("%d", f.Value)
}

// GetDefaultText returns the default text for this flag
func (f *IntFlag) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return f.GetValue()
}

// GetEnvVars returns the env vars for this flag
func (f *IntFlag) GetEnvVars() []string {
	return f.EnvVars
}

// Apply populates the flag given the flag set and environment
func (f *IntFlag) Apply(set *flag.FlagSet) error {
	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		if val != "" {
			valInt, err := strconv.ParseInt(val, 0, 64)

			if err != nil {
				return fmt.Errorf("could not parse %q as int value from %s for flag %s: %s", val, source, f.Name, err)
			}

			f.Value = int(valInt)
			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.IntVar(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.Int(name, f.Value, f.Usage)
	}

	return nil
}

// Get returns the flag’s value in the given Context.
func (f *IntFlag) Get(ctx *Context) int {
	return ctx.Int(f.Name)
}

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (cCtx *Context) Int(name string) int {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupInt(name, fs)
	}
	return 0
}

func lookupInt(name string, set *flag.FlagSet) int {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := strconv.ParseInt(f.Value.String(), 0, 64)
		if err != nil {
			return 0
		}
		return int(parsed)
	}
	return 0
}
