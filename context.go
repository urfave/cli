package cli

import (
	"flag"
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
	return c.lookupInt(name, c.flagSet)
}

// Looks up the value of a local bool flag, returns false if no bool flag exists
func (c *Context) Bool(name string) bool {
	return c.lookupBool(name, c.flagSet)
}

// Looks up the value of a local string flag, returns "" if no string flag exists
func (c *Context) String(name string) string {
	return c.lookupString(name, c.flagSet)
}

// Looks up the value of a local string slice flag, returns nil if no string slice flag exists
func (c *Context) StringSlice(name string) []string {
	return c.lookupStringSlice(name, c.flagSet)
}

// Looks up the value of a local int slice flag, returns nil if no int slice flag exists
func (c *Context) IntSlice(name string) []int {
	return c.lookupIntSlice(name, c.flagSet)
}

// Looks up the value of a global int flag, returns 0 if no int flag exists
func (c *Context) GlobalInt(name string) int {
	return c.lookupInt(name, c.globalSet)
}

// Looks up the value of a global bool flag, returns false if no bool flag exists
func (c *Context) GlobalBool(name string) bool {
	return c.lookupBool(name, c.globalSet)
}

// Looks up the value of a global string flag, returns "" if no string flag exists
func (c *Context) GlobalString(name string) string {
	return c.lookupString(name, c.globalSet)
}

// Looks up the value of a global string slice flag, returns nil if no string slice flag exists
func (c *Context) GlobalStringSlice(name string) []string {
	return c.lookupStringSlice(name, c.globalSet)
}

// Looks up the value of a global int slice flag, returns nil if no int slice flag exists
func (c *Context) GlobalIntSlice(name string) []int {
	return c.lookupIntSlice(name, c.globalSet)
}

// Returns the command line arguments associated with the context.
func (c *Context) Args() []string {
	return c.flagSet.Args()
}

func (c *Context) lookupInt(name string, set *flag.FlagSet) int {
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

func (c *Context) lookupString(name string, set *flag.FlagSet) string {
	f := set.Lookup(name)
	if f != nil {
		return f.Value.String()
	}

	return ""
}

func (c *Context) lookupStringSlice(name string, set *flag.FlagSet) []string {
	f := set.Lookup(name)
	if f != nil {
		return (f.Value.(*StringSlice)).Value()

	}

	return nil
}

func (c *Context) lookupIntSlice(name string, set *flag.FlagSet) []int {
	f := set.Lookup(name)
	if f != nil {
		return (f.Value.(*IntSlice)).Value()

	}

	return nil
}

func (c *Context) lookupBool(name string, set *flag.FlagSet) bool {
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

// Returns the nth argument, or just a blank string
func (c *Context) GetArg(n int) string {
	args := c.Args()
	if len(args) < n {
		return args[n]
	}
	return ""
}

// Returns the first argument, or just a blank string, for convenience
func (c *Context) FirstArg() string {
	return c.GetArg(0)
}
