package cli

import (
	"flag"
	"fmt"
)

// TakesValue returns true of the flag takes a value, otherwise false
func (f *StringFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *StringFlag) GetUsage() string {
	return f.Usage
}

// GetCategory returns the category for the flag
func (f *StringFlag) GetCategory() string {
	return f.Category
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *StringFlag) GetValue() string {
	return f.Value
}

// GetDefaultText returns the default text for this flag
func (f *StringFlag) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	if f.defaultValue == "" {
		return f.defaultValue
	}
	return fmt.Sprintf("%q", f.defaultValue)
}

// GetEnvVars returns the env vars for this flag
func (f *StringFlag) GetEnvVars() []string {
	return f.EnvVars
}

// Apply populates the flag given the flag set and environment
func (f *StringFlag) Apply(set *flag.FlagSet) error {
	// set default value so that environment wont be able to overwrite it
	f.defaultValue = f.Value

	if val, _, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		f.Value = val
		f.HasBeenSet = true
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.StringVar(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.String(name, f.Value, f.Usage)
	}

	return nil
}

// Get returns the flag’s value in the given Context.
func (f *StringFlag) Get(ctx *Context) string {
	return ctx.String(f.Name)
}

// RunAction executes flag action if set
func (f *StringFlag) RunAction(c *Context) error {
	if f.Action != nil {
		return f.Action(c, c.String(f.Name))
	}

	return nil
}

// String looks up the value of a local StringFlag, returns
// "" if not found
func (cCtx *Context) String(name string) string {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupString(name, fs)
	}
	return ""
}

func lookupString(name string, set *flag.FlagSet) string {
	f := set.Lookup(name)
	if f != nil {
		parsed := f.Value.String()
		return parsed
	}
	return ""
}
