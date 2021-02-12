package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"
)

// Context is a type that is passed through to
// each Handler action in a cli application. Context
// can be used to retrieve context-specific args and
// parsed command-line options.
type Context struct {
	context.Context
	App           *App
	Command       *Command
	shellComplete bool
	flagSet       *flag.FlagSet
	fromFlagSet   map[string]bool
	parentContext *Context
}

// NewContext creates a new context. For use in when invoking an App or Command action.
func NewContext(app *App, set *flag.FlagSet, parentCtx *Context) *Context {
	c := &Context{App: app, flagSet: set, parentContext: parentCtx}
	if parentCtx != nil {
		c.Context = parentCtx.Context
		c.shellComplete = parentCtx.shellComplete
		if parentCtx.flagSet == nil {
			parentCtx.flagSet = &flag.FlagSet{}
		}
	}

	// pre-compute flag seen on the command line at this context
	c.fromFlagSet = make(map[string]bool)
	if set != nil {
		set.Visit(func(f *flag.Flag) {
			c.fromFlagSet[f.Name] = true
		})
	}

	c.Command = &Command{}

	if c.Context == nil {
		c.Context = context.Background()
	}

	return c
}

// NumFlags returns the number of flags set
func (cCtx *Context) NumFlags() int {
	return cCtx.flagSet.NFlag()
}

// Set sets a context flag to a value.
func (c *Context) Set(name, value string) error {
	err := c.flagSet.Set(name, value)
	if err == nil {
		c.fromFlagSet[name] = true
	}

	return err
}

// IsSet determines if the flag was actually set
func (c *Context) IsSet(name string) bool {
	for ctx := c; ctx != nil; ctx = ctx.parentContext {
		// try flags parsed from command line first
		if ctx.flagSet.Lookup(name) == nil {
			// flag not defined in this context
			continue
		}
		if ctx.flagOnCommandLine(name) {
			return true
		}

		// now see if value was set externally via environment
		definedFlags := ctx.Command.Flags
		if ctx.Command.Name == "" && ctx.App != nil {
			definedFlags = ctx.App.Flags
		}
		for _, ff := range definedFlags {
			for _, fn := range ff.Names() {
				if fn == name {
					if ff.IsSet() {
						return true
					}
					break
				}
			}
		}
	}

	return false
}

func (c *Context) flagOnCommandLine(name string) bool {
	return c.fromFlagSet[name]
}

// LocalFlagNames returns a slice of flag names used in this context.
func (cCtx *Context) LocalFlagNames() []string {
	var names []string
	cCtx.flagSet.Visit(makeFlagNameVisitor(&names))
	// Check the flags which have been set via env or file
	if cCtx.Command != nil && cCtx.Command.Flags != nil {
		for _, f := range cCtx.Command.Flags {
			if f.IsSet() {
				names = append(names, f.Names()...)
			}
		}
	}

	// Sort out the duplicates since flag could be set via multiple
	// paths
	m := map[string]struct{}{}
	var unames []string
	for _, name := range names {
		if _, ok := m[name]; !ok {
			m[name] = struct{}{}
			unames = append(unames, name)
		}
	}

	return unames
}

// FlagNames returns a slice of flag names used by the this context and all of
// its parent contexts.
func (cCtx *Context) FlagNames() []string {
	var names []string
	for _, pCtx := range cCtx.Lineage() {
		names = append(names, pCtx.LocalFlagNames()...)
	}
	return names
}

// Lineage returns *this* context and all of its ancestor contexts in order from
// child to parent
func (cCtx *Context) Lineage() []*Context {
	var lineage []*Context

	for cur := cCtx; cur != nil; cur = cur.parentContext {
		lineage = append(lineage, cur)
	}

	return lineage
}

// Count returns the num of occurences of this flag
func (cCtx *Context) Count(name string) int {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		if cf, ok := fs.Lookup(name).Value.(Countable); ok {
			return cf.Count()
		}
	}
	return 0
}

// Value returns the value of the flag corresponding to `name`
func (cCtx *Context) Value(name string) interface{} {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return fs.Lookup(name).Value.(flag.Getter).Get()
	}
	return nil
}

// Args returns the command line arguments associated with the context.
func (cCtx *Context) Args() Args {
	ret := args(cCtx.flagSet.Args())
	return &ret
}

// NArg returns the number of the command line arguments.
func (cCtx *Context) NArg() int {
	return cCtx.Args().Len()
}

func (cCtx *Context) lookupFlag(name string) Flag {
	for _, c := range cCtx.Lineage() {
		if c.Command == nil {
			continue
		}

		for _, f := range c.Command.Flags {
			for _, n := range f.Names() {
				if n == name {
					return f
				}
			}
		}
	}

	if cCtx.App != nil {
		for _, f := range cCtx.App.Flags {
			for _, n := range f.Names() {
				if n == name {
					return f
				}
			}
		}
	}

	return nil
}

<<<<<<< HEAD
func (cCtx *Context) lookupFlagSet(name string) *flag.FlagSet {
	for _, c := range cCtx.Lineage() {
		if c.flagSet == nil {
			continue
		}
=======
func (c *Context) resolveFlagDeep(name string) *flag.Flag {
	var src *flag.Flag
	for cur := c; cur != nil; cur = cur.parentContext {
		if f := cur.flagSet.Lookup(name); f != nil {
			if cur.flagOnCommandLine(name) {
				// we've found most specific instance on command line, use it
				src = f
				break
			}
			if src == nil {
				// flag was defined, but value is not present among flags of the current context
				// remember the most specific instance of the flag not from command line as fallback
				src = f
			}
		}
	}
	return src
}

func lookupFlagSet(name string, ctx *Context) *flag.FlagSet {
	for _, c := range ctx.Lineage() {
>>>>>>> Allow same flag to be repeated by child commands. Reading the value or checking for presence looks at the most specific subcommand that includes the flag on the command line.
		if f := c.flagSet.Lookup(name); f != nil {
			return c.flagSet
		}
	}
	cCtx.onInvalidFlag(name)
	return nil
}

func (cCtx *Context) checkRequiredFlags(flags []Flag) requiredFlagsErr {
	var missingFlags []string
	for _, f := range flags {
		if rf, ok := f.(RequiredFlag); ok && rf.IsRequired() {
			var flagPresent bool
			var flagName string

			for _, key := range f.Names() {
				flagName = key

				if cCtx.IsSet(strings.TrimSpace(key)) {
					flagPresent = true
				}
			}

			if !flagPresent && flagName != "" {
				missingFlags = append(missingFlags, flagName)
			}
		}
	}

	if len(missingFlags) != 0 {
		return &errRequiredFlags{missingFlags: missingFlags}
	}

	return nil
}

func (cCtx *Context) onInvalidFlag(name string) {
	for cCtx != nil {
		if cCtx.App != nil && cCtx.App.InvalidFlagAccessHandler != nil {
			cCtx.App.InvalidFlagAccessHandler(cCtx, name)
			break
		}
		cCtx = cCtx.parentContext
	}
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
			*names = append(*names, name)
		}
	}
}
