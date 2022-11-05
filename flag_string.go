package cli

import (
	"flag"
)

// -- string Value
type stringValue string

func (i stringValue) Create(val string, p *string) flag.Value {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() any { return string(*s) }

func (s *stringValue) String() string { return string(*s) }

type StringFlag = flagImpl[string, stringValue]

func (cCtx *Context) String(name string) string {
	if v, ok := cCtx.Value(name).(string); ok {
		return v
	}
	return ""
}
