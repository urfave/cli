package cli

import (
	"fmt"
	"strings"
)

type StringFlag = FlagBase[string, StringConfig, stringValue]

// StringConfig defines the configuration for string flags
type StringConfig struct {
	// Whether to trim whitespace of parsed value
	TrimSpace bool
}

// -- string Value
type stringValue struct {
	destination *string
	trimSpace   bool
}

// Below functions are to satisfy the ValueCreator interface

func (i stringValue) Create(val string, p *string, c StringConfig) Value {
	*p = val
	return &stringValue{
		destination: p,
		trimSpace:   c.TrimSpace,
	}
}

func (i stringValue) ToString(b string) string {
	if b == "" {
		return b
	}
	return fmt.Sprintf("%q", b)
}

// Below functions are to satisfy the flag.Value interface

func (s *stringValue) Set(val string) error {
	if s.trimSpace {
		val = strings.TrimSpace(val)
	}
	*s.destination = val
	return nil
}

func (s *stringValue) Get() any { return *s.destination }

func (s *stringValue) String() string {
	if s.destination != nil {
		return *s.destination
	}
	return ""
}

func (cCtx *Context) String(name string) string {
	if v, ok := cCtx.Value(name).(string); ok {
		return v
	}
	return ""
}
