package cli

import (
	"errors"
	"flag"
	"strings"
	"time"
)

// Context is a type that is passed through to
// each Handler action in a cli application. Context
// can be used to retrieve context-specific Args and
// parsed command-line options.
type Context struct {
	App            *App
	Command        Command
	flagSetManager FlagSetManager
	parentContext  *Context
}

// Creates a new context. For use in when invoking an App or Command action.
func NewContext(app *App, set *flag.FlagSet, parentCtx *Context) *Context {
	return &Context{App: app, flagSetManager: NewFlagSetManager(set), parentContext: parentCtx}
}

// Looks up the value of a local int flag, returns 0 if no int flag exists
func (c *Context) Int(name string) int {
	return c.flagSetManager.Int(name)
}

// Looks up the value of a local time.Duration flag, returns 0 if no time.Duration flag exists
func (c *Context) Duration(name string) time.Duration {
	return c.flagSetManager.Duration(name)
}

// Looks up the value of a local float64 flag, returns 0 if no float64 flag exists
func (c *Context) Float64(name string) float64 {
	return c.flagSetManager.Float64(name)
}

// Looks up the value of a local bool flag, returns false if no bool flag exists
func (c *Context) Bool(name string) bool {
	return c.flagSetManager.Bool(name)
}

// Looks up the value of a local boolT flag, returns false if no bool flag exists
func (c *Context) BoolT(name string) bool {
	return c.flagSetManager.BoolT(name)
}

// Looks up the value of a local string flag, returns "" if no string flag exists
func (c *Context) String(name string) string {
	return c.flagSetManager.String(name)
}

// Looks up the value of a local string slice flag, returns nil if no string slice flag exists
func (c *Context) StringSlice(name string) []string {
	return c.flagSetManager.StringSlice(name)
}

// Looks up the value of a local int slice flag, returns nil if no int slice flag exists
func (c *Context) IntSlice(name string) []int {
	return c.flagSetManager.IntSlice(name)
}

// Looks up the value of a local generic flag, returns nil if no generic flag exists
func (c *Context) Generic(name string) interface{} {
	return c.flagSetManager.Generic(name)
}

// Looks up the value of a global int flag, returns 0 if no int flag exists
func (c *Context) GlobalInt(name string) int {
	if fsm := lookupGlobalFlagSetManager(name, c); fsm != nil {
		return fsm.Int(name)
	}
	return 0
}

// Looks up the value of a global time.Duration flag, returns 0 if no time.Duration flag exists
func (c *Context) GlobalDuration(name string) time.Duration {
	if fsm := lookupGlobalFlagSetManager(name, c); fsm != nil {
		return fsm.Duration(name)
	}
	return 0
}

// Looks up the value of a global bool flag, returns false if no bool flag exists
func (c *Context) GlobalBool(name string) bool {
	if fsm := lookupGlobalFlagSetManager(name, c); fsm != nil {
		return fsm.Bool(name)
	}
	return false
}

// Looks up the value of a global string flag, returns "" if no string flag exists
func (c *Context) GlobalString(name string) string {
	if fsm := lookupGlobalFlagSetManager(name, c); fsm != nil {
		return fsm.String(name)
	}
	return ""
}

// Looks up the value of a global string slice flag, returns nil if no string slice flag exists
func (c *Context) GlobalStringSlice(name string) []string {
	if fsm := lookupGlobalFlagSetManager(name, c); fsm != nil {
		return fsm.StringSlice(name)
	}
	return nil
}

// Looks up the value of a global int slice flag, returns nil if no int slice flag exists
func (c *Context) GlobalIntSlice(name string) []int {
	if fsm := lookupGlobalFlagSetManager(name, c); fsm != nil {
		return fsm.IntSlice(name)
	}
	return nil
}

// Looks up the value of a global generic flag, returns nil if no generic flag exists
func (c *Context) GlobalGeneric(name string) interface{} {
	if fsm := lookupGlobalFlagSetManager(name, c); fsm != nil {
		return fsm.Generic(name)
	}
	return nil
}

// Returns the number of flags set
func (c *Context) NumFlags() int {
	return c.flagSetManager.NumFlags()
}

// Determines if the flag was actually set
func (c *Context) IsSet(name string) bool {
	return c.flagSetManager.IsSet(name)
}

// Determines if the global flag was actually set
func (c *Context) GlobalIsSet(name string) bool {
	if fsm := lookupGlobalFlagSetManager(name, c); fsm != nil {
		return fsm.IsSet(name)
	}

	return false
}

// Returns a slice of flag names used in this context.
func (c *Context) FlagNames() (names []string) {
	for _, flag := range c.Command.Flags {
		name := strings.Split(flag.getName(), ",")[0]
		if name == "help" {
			continue
		}
		names = append(names, name)
	}
	return
}

// Returns a slice of global flag names used by the app.
func (c *Context) GlobalFlagNames() (names []string) {
	for _, flag := range c.App.Flags {
		name := strings.Split(flag.getName(), ",")[0]
		if name == "help" || name == "version" {
			continue
		}
		names = append(names, name)
	}
	return
}

// Returns the parent context, if any
func (c *Context) Parent() *Context {
	return c.parentContext
}

type Args []string

// Returns the command line arguments associated with the context.
func (c *Context) Args() Args {
	return c.flagSetManager.Args()
}

// Returns the nth argument, or else a blank string
func (a Args) Get(n int) string {
	if len(a) > n {
		return a[n]
	}
	return ""
}

// Returns the first argument, or else a blank string
func (a Args) First() string {
	return a.Get(0)
}

// Return the rest of the arguments (not the first one)
// or else an empty string slice
func (a Args) Tail() []string {
	if len(a) >= 2 {
		return []string(a)[1:]
	}
	return []string{}
}

// Checks if there are any arguments present
func (a Args) Present() bool {
	return len(a) != 0
}

// Swaps arguments at the given indexes
func (a Args) Swap(from, to int) error {
	if from >= len(a) || to >= len(a) {
		return errors.New("index out of range")
	}
	a[from], a[to] = a[to], a[from]
	return nil
}

func lookupGlobalFlagSetManager(name string, ctx *Context) FlagSetManager {
	if ctx.parentContext != nil {
		ctx = ctx.parentContext
	}
	for ; ctx != nil; ctx = ctx.parentContext {
		if ctx.flagSetManager.HasFlag(name) {
			return ctx.flagSetManager
		}
	}
	return nil
}

func copyFlag(name string, ff *flag.Flag, set *flag.FlagSet) {
	switch ff.Value.(type) {
	case *StringSlice:
	default:
		set.Set(name, ff.Value.String())
	}
}

func normalizeFlags(flags []Flag, set *flag.FlagSet) error {
	visited := make(map[string]bool)
	set.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})
	for _, f := range flags {
		parts := strings.Split(f.getName(), ",")
		if len(parts) == 1 {
			continue
		}
		var ff *flag.Flag
		for _, name := range parts {
			name = strings.Trim(name, " ")
			if visited[name] {
				if ff != nil {
					return errors.New("Cannot use two forms of the same flag: " + name + " " + ff.Name)
				}
				ff = set.Lookup(name)
			}
		}
		if ff == nil {
			continue
		}
		for _, name := range parts {
			name = strings.Trim(name, " ")
			if !visited[name] {
				copyFlag(name, ff, set)
			}
		}
	}
	return nil
}
