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
