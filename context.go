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
	flagSet *flag.FlagSet
}

func NewContext(flagSet *flag.FlagSet) *Context {
	return &Context{flagSet}
}

func (c *Context) Int(name string) int {
	flag := c.flagSet.Lookup(name)
	if flag != nil {
		val, err := strconv.Atoi(flag.Value.String())
		if err != nil {
			panic(err)
		}
		return val
	} else {
		return 0
	}
}

func (c *Context) Bool(name string) bool {
	flag := c.flagSet.Lookup(name)
	if flag != nil {
		val, err := strconv.ParseBool(flag.Value.String())
		if err != nil {
			panic(err)
		}
		return val
	} else {
		return false
	}
}

func (c *Context) String(name string) string {
	flag := c.flagSet.Lookup(name)
	if flag != nil {
		return flag.Value.String()
	} else {
		return ""
	}
}

func (c *Context) Args() []string {
	return c.flagSet.Args()
}
