package cli

import (
	"fmt"
)

type StringFlag = FlagBase[string, NoConfig, stringValue]

// -- string Value
type stringValue string

// Below functions are to satisfy the ValueCreator interface

func (i stringValue) Create(val string, p *string, c NoConfig) Value {
	*p = val
	return (*stringValue)(p)
}

func (i stringValue) ToString(b string) string {
	if b == "" {
		return b
	}
	return fmt.Sprintf("%q", b)
}

// Below functions are to satisfy the flag.Value interface

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() any { return string(*s) }

func (s *stringValue) String() string { return string(*s) }

func (cCtx *Context) String(name string) string {
	if v, ok := cCtx.Value(name).(string); ok {
		return v
	}
	return ""
}
