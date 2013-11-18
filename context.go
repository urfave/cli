package cli

import (
	"flag"
	"strings"
	"strconv"
)

// Context is a type that is passed through to
// each Handler action in a cli application. Context
// can be used to retrieve context-specific Args and
// parsed command-line options.
type Context struct {
	App       *App
	flagSet   *flag.FlagSet
	globalSet *flag.FlagSet
}

// Creates a new context. For use in when invoking an App or Command action.
func NewContext(app *App, set *flag.FlagSet, globalSet *flag.FlagSet) *Context {
	return &Context{app, set, globalSet}
}

// Looks up the value of a local int flag, returns 0 if no int flag exists
func (c *Context) Int(name string) int {
	return lookupInt(name, c.flagSet)
}

// Looks up the value of a local bool flag, returns false if no bool flag exists
func (c *Context) Bool(name string) bool {
	return lookupBool(name, c.flagSet)
}

// Looks up the value of a local string flag, returns "" if no string flag exists
func (c *Context) String(name string) string {
	return lookupString(name, c.flagSet)
}

// Looks up the value of a local string slice flag, returns nil if no string slice flag exists
func (c *Context) StringSlice(name string) []string {
	return lookupStringSlice(name, c.flagSet)
}

// Looks up the value of a local int slice flag, returns nil if no int slice flag exists
func (c *Context) IntSlice(name string) []int {
	return lookupIntSlice(name, c.flagSet)
}

// Looks up the value of a global int flag, returns 0 if no int flag exists
func (c *Context) GlobalInt(name string) int {
	return lookupInt(name, c.globalSet)
}

// Looks up the value of a global bool flag, returns false if no bool flag exists
func (c *Context) GlobalBool(name string) bool {
	return lookupBool(name, c.globalSet)
}

// Looks up the value of a global string flag, returns "" if no string flag exists
func (c *Context) GlobalString(name string) string {
	return lookupString(name, c.globalSet)
}

// Looks up the value of a global string slice flag, returns nil if no string slice flag exists
func (c *Context) GlobalStringSlice(name string) []string {
	return lookupStringSlice(name, c.globalSet)
}

// Looks up the value of a global int slice flag, returns nil if no int slice flag exists
func (c *Context) GlobalIntSlice(name string) []int {
	return lookupIntSlice(name, c.globalSet)
}

// Returns the command line arguments associated with the context.
func (c *Context) Args() []string {
	return c.flagSet.Args()
}

func lookupInt(name string, set *flag.FlagSet) int {
	f := set.Lookup(name)
	if f != nil {
		val, err := strconv.Atoi(f.Value.String())
		if err != nil {
			return 0
		}
		return val
	}

	return 0
}

func lookupString(name string, set *flag.FlagSet) string {
	f := set.Lookup(name)
	if f != nil {
		return f.Value.String()
	}

	return ""
}

func lookupStringSlice(name string, set *flag.FlagSet) []string {
	f := set.Lookup(name)
	if f != nil {
		return (f.Value.(*StringSlice)).Value()

	}

	return nil
}

func lookupIntSlice(name string, set *flag.FlagSet) []int {
	f := set.Lookup(name)
	if f != nil {
		return (f.Value.(*IntSlice)).Value()

	}

	return nil
}

func lookupBool(name string, set *flag.FlagSet) bool {
	f := set.Lookup(name)
	if f != nil {
		val, err := strconv.ParseBool(f.Value.String())
		if err != nil {
			return false
		}
		return val
	}

	return false
}

func normalizeFlags(flags []Flag, set *flag.FlagSet) {
	for _, f := range flags {
		parts := strings.Split(f.GetName(), ", ")
		if len(parts) == 1 {
			continue
		}
		var ff *flag.Flag
		for _, name := range parts {
			ff = set.Lookup(name)
			if ff != nil && ff.Value.String() != "" {
				break
			}
		}
		if ff == nil {
			continue
		}
		for _, name := range parts {
			set.Set(name, ff.Value.String())
		}
	}
}
