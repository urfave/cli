package cli

import (
	"flag"
	"fmt"
	"strconv"
)

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *IntFlag) GetValue() string {
	return fmt.Sprintf("%d", f.Value)
}

// Apply populates the flag given the flag set and environment
func (f *IntFlag) Apply(set *flag.FlagSet) error {
	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		if val != "" {
			valInt, err := strconv.ParseInt(val, f.Base, 64)

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

// RunAction executes flag action if set
func (f *IntFlag) RunAction(c *Context) error {
	if f.Action != nil {
		return f.Action(c, c.Int(f.Name))
	}

	return nil
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
