package cli

import (
	"errors"
	"time"
)

// Context is a type that is passed through to each action in a cli application.
// Context can be used to retrieve context-specific Args and parsed command-line
// options.
type Context struct {
	App     *App
	Command *Command

	flagSet       *FlagSet
	parentContext *Context
}

// NewContext creates a new context. For use in when invoking an App or Command action.
func NewContext(app *App, set *FlagSet, parentCtx *Context) *Context {
	return &Context{App: app, flagSet: set, parentContext: parentCtx}
}

// Int looks up the value of a local int flag, returns 0 if no int flag exists
func (c *Context) Int(name string) int {
	if fs := c.lookupFlagSet(name); fs != nil {
		return fs.GetInt(name)
	}
	return 0
}

// Duration looks up the value of a local time.Duration flag, returns 0 if no
// time.Duration flag exists
func (c *Context) Duration(name string) time.Duration {
	if fs := c.lookupFlagSet(name); fs != nil {
		return fs.GetDuration(name)
	}
	return 0
}

// Float64 looks up the value of a local float64 flag, returns 0 if no float64
// flag exists
func (c *Context) Float64(name string) float64 {
	if fs := c.lookupFlagSet(name); fs != nil {
		return fs.GetFloat64(name)
	}
	return 0
}

// Bool looks up the value of a local bool flag, returns false if no bool flag exists
func (c *Context) Bool(name string) bool {
	if fs := c.lookupFlagSet(name); fs != nil {
		return fs.GetBool(name)
	}
	return false
}

// String looks up the value of a local string flag, returns "" if no string flag exists
func (c *Context) String(name string) string {
	if fs := c.lookupFlagSet(name); fs != nil {
		return fs.GetString(name)
	}
	return ""
}

// StringSlice looks up the value of a local string slice flag, returns nil if no
// string slice flag exists
func (c *Context) StringSlice(name string) []string {
	if fs := c.lookupFlagSet(name); fs != nil {
		return fs.GetStringSlice(name)
	}
	return nil
}

// IntSlice looks up the value of a local int slice flag, returns nil if no int
// slice flag exists
func (c *Context) IntSlice(name string) []int {
	if fs := c.lookupFlagSet(name); fs != nil {
		return fs.GetIntSlice(name)
	}
	return nil
}

// Generic looks up the value of a local generic flag, returns nil if no generic
// flag exists
func (c *Context) Generic(name string) interface{} {
	if fs := c.lookupFlagSet(name); fs != nil {
		return fs.GetGeneric(name)
	}
	return nil
}

// NumFlags returns the number of flags set
func (c *Context) NumFlags() int {
	n := 0
	for _, name := range c.FlagNames() {
		if c.IsSet(name) {
			n++
		}
	}
	return n
}

// Set sets a context flag to a string value.
func (c *Context) Set(name, value string) error {
	return c.flagSet.SetString(name, value)
}

// IsSet determines if the flag was actually set
func (c *Context) IsSet(name string) bool {
	if fs := c.lookupFlagSet(name); fs != nil {
		return fs.IsSet(name)
	}
	return false
}

// LocalFlagNames returns a slice of flag names used in this context.
func (c *Context) LocalFlagNames() []string {
	names := []string{}
	c.flagSet.Each(makeFlagNameVisitor(&names))
	return names
}

// FlagNames returns a slice of flag names used by the this context and all of
// its parent contexts.
func (c *Context) FlagNames() []string {
	names := []string{}
	for _, ctx := range c.Lineage() {
		ctx.flagSet.Each(makeFlagNameVisitor(&names))
	}
	return names
}

// Lineage returns *this* context and all of its ancestor contexts in order from
// child to parent
func (c *Context) Lineage() []*Context {
	lineage := []*Context{}

	for cur := c; cur != nil; cur = cur.parentContext {
		lineage = append(lineage, cur)
	}

	return lineage
}

// Args returns the command line arguments associated with the context.
func (c *Context) Args() *Args {
	return &Args{slice: c.flagSet.RemainingArgs()}
}

// NumArgs returns the number of the command line arguments.
func (c *Context) NumArgs() int {
	return c.Args().Len()
}

func (c *Context) lookupFlagSet(name string) *FlagSet {
	for _, ctx := range c.Lineage() {
		if f := ctx.flagSet.Lookup(name); f != nil {
			return ctx.flagSet
		}
	}

	return nil
}

// Args wraps a string slice with some convenience methods
type Args struct {
	slice []string
}

// Get returns the nth argument, or else a blank string
func (a *Args) Get(n int) string {
	if a.Len() > n {
		return a.slice[n]
	}
	return ""
}

// First returns the first argument, or else a blank string
func (a *Args) First() string {
	return a.Get(0)
}

// Tail returns the rest of the arguments (not the first one)
// or else an empty string slice
func (a *Args) Tail() []string {
	if a.Len() >= 2 {
		return a.slice[1:]
	}
	return []string{}
}

// Present checks if there are any arguments present
func (a *Args) Present() bool {
	return a.Len() != 0
}

// Len returns the length of the wrapped slice
func (a *Args) Len() int {
	return len(a.slice)
}

// Swap swaps arguments at the given indexes
func (a *Args) Swap(from, to int) error {
	if from >= len(a.slice) || to >= len(a.slice) {
		return errors.New("index out of range")
	}
	a.slice[from], a.slice[to] = a.slice[to], a.slice[from]
	return nil
}

// Slice returns a copy of the internal slice
func (a *Args) Slice() []string {
	ret := make([]string, len(a.slice))
	copy(ret, a.slice)
	return ret
}

func makeFlagNameVisitor(names *[]string) func(Flag) {
	return func(f Flag) {
		nameParts := f.Names()
		name := nameParts[0]

		for _, part := range nameParts {
			if len(part) > len(name) {
				name = part
			}
		}

		if name != "" {
			(*names) = append(*names, name)
		}
	}
}
