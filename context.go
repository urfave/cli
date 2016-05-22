package cli

import (
	"errors"
	"flag"
	"strconv"
	"strings"
	"time"
)

// Context is a type that is passed through to
// each Handler action in a cli application. Context
// can be used to retrieve context-specific Args and
// parsed command-line options.
type Context struct {
	App     *App
	Command Command

	flagSet       *flag.FlagSet
	parentContext *Context
}

// NewContext creates a new context. For use in when invoking an App or Command action.
func NewContext(app *App, set *flag.FlagSet, parentCtx *Context) *Context {
	return &Context{App: app, flagSet: set, parentContext: parentCtx}
}

// Int looks up the value of a local int flag, returns 0 if no int flag exists
func (c *Context) Int(name string) int {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupInt(name, fs)
	}
	return 0
}

// Duration looks up the value of a local time.Duration flag, returns 0 if no
// time.Duration flag exists
func (c *Context) Duration(name string) time.Duration {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupDuration(name, fs)
	}
	return 0
}

// Float64 looks up the value of a local float64 flag, returns 0 if no float64
// flag exists
func (c *Context) Float64(name string) float64 {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupFloat64(name, fs)
	}
	return 0
}

// Bool looks up the value of a local bool flag, returns false if no bool flag exists
func (c *Context) Bool(name string) bool {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupBool(name, fs)
	}
	return false
}

// String looks up the value of a local string flag, returns "" if no string flag exists
func (c *Context) String(name string) string {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupString(name, fs)
	}
	return ""
}

// StringSlice looks up the value of a local string slice flag, returns nil if no
// string slice flag exists
func (c *Context) StringSlice(name string) []string {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupStringSlice(name, fs)
	}
	return nil
}

// IntSlice looks up the value of a local int slice flag, returns nil if no int
// slice flag exists
func (c *Context) IntSlice(name string) []int {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupIntSlice(name, fs)
	}
	return nil
}

// Generic looks up the value of a local generic flag, returns nil if no generic
// flag exists
func (c *Context) Generic(name string) interface{} {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupGeneric(name, fs)
	}
	return nil
}

// NumFlags returns the number of flags set
func (c *Context) NumFlags() int {
	return c.flagSet.NFlag()
}

// Set sets a context flag to a value.
func (c *Context) Set(name, value string) error {
	return c.flagSet.Set(name, value)
}

// IsSet determines if the flag was actually set
func (c *Context) IsSet(name string) bool {
	if fs := lookupFlagSet(name, c); fs != nil {
		isSet := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name == name {
				isSet = true
			}
		})
		return isSet
	}
	return false
}

// LocalFlagNames returns a slice of flag names used in this context.
func (c *Context) LocalFlagNames() []string {
	names := []string{}
	c.flagSet.Visit(makeFlagNameVisitor(&names))
	return names
}

// FlagNames returns a slice of flag names used by the this context and all of
// its parent contexts.
func (c *Context) FlagNames() []string {
	names := []string{}
	for _, ctx := range c.Lineage() {
		ctx.flagSet.Visit(makeFlagNameVisitor(&names))
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

// Args contains apps console arguments
type Args []string

// Args returns the command line arguments associated with the context.
func (c *Context) Args() Args {
	args := Args(c.flagSet.Args())
	return args
}

// NArg returns the number of the command line arguments.
func (c *Context) NArg() int {
	return len(c.Args())
}

// Get returns the nth argument, or else a blank string
func (a Args) Get(n int) string {
	if len(a) > n {
		return a[n]
	}
	return ""
}

// First returns the first argument, or else a blank string
func (a Args) First() string {
	return a.Get(0)
}

// Tail returns the rest of the arguments (not the first one)
// or else an empty string slice
func (a Args) Tail() []string {
	if len(a) >= 2 {
		return []string(a)[1:]
	}
	return []string{}
}

// Present checks if there are any arguments present
func (a Args) Present() bool {
	return len(a) != 0
}

// Swap swaps arguments at the given indexes
func (a Args) Swap(from, to int) error {
	if from >= len(a) || to >= len(a) {
		return errors.New("index out of range")
	}
	a[from], a[to] = a[to], a[from]
	return nil
}

func lookupFlagSet(name string, ctx *Context) *flag.FlagSet {
	for _, c := range ctx.Lineage() {
		if f := c.flagSet.Lookup(name); f != nil {
			return c.flagSet
		}
	}

	return nil
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

func lookupDuration(name string, set *flag.FlagSet) time.Duration {
	f := set.Lookup(name)
	if f != nil {
		val, err := time.ParseDuration(f.Value.String())
		if err == nil {
			return val
		}
	}

	return 0
}

func lookupFloat64(name string, set *flag.FlagSet) float64 {
	f := set.Lookup(name)
	if f != nil {
		val, err := strconv.ParseFloat(f.Value.String(), 64)
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

func lookupGeneric(name string, set *flag.FlagSet) interface{} {
	f := set.Lookup(name)
	if f != nil {
		return f.Value
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

func copyFlag(name string, ff *flag.Flag, set *flag.FlagSet) {
	switch ff.Value.(type) {
	case Serializeder:
		set.Set(name, ff.Value.(Serializeder).Serialized())
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
		parts := f.Names()
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

func makeFlagNameVisitor(names *[]string) func(*flag.Flag) {
	return func(f *flag.Flag) {
		nameParts := strings.Split(f.Name, ",")
		name := strings.TrimSpace(nameParts[0])

		for _, part := range nameParts {
			part = strings.TrimSpace(part)
			if len(part) > len(name) {
				name = part
			}
		}

		if name != "" {
			(*names) = append(*names, name)
		}
	}
}
