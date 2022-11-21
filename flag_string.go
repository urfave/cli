package cli

import (
	"flag"
	"fmt"
)

// -- string Value
type stringValue string

func (i stringValue) Create(val string, p *string, c FlagConfig) flag.Value {
	*p = val
	return (*stringValue)(p)
}

func (i stringValue) ToString(b string) string {
	if b == "" {
		return b
	}
	return fmt.Sprintf("%q", b)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() any { return string(*s) }

func (s *stringValue) String() string { return string(*s) }

type StringFlag = FlagBase[string, stringValue]

func (cCtx *Context) String(name string) string {
	if v, ok := cCtx.Value(name).(string); ok {
		return v
	}
	return ""
}
