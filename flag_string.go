package cli

import (
	"flag"
	"fmt"
)

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
	if f.Value == "" {
		return f.Value
	}
	return fmt.Sprintf("%q", f.Value)
}

// Apply populates the flag given the flag set and environment
func (f *StringFlag) Apply(set *flag.FlagSet) error {
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
