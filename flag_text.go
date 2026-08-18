package cli

import (
	"encoding"
	"strings"
)

// TextMarshalUnmarshaler is the interface implemented by types that can marshal
// themselves to and from a textual form. It is the value type used by TextFlag,
// mirroring the standard library's flag.TextVar, and is satisfied by types such
// as *slog.LevelVar, *net/netip.Addr, and *time.Time.
type TextMarshalUnmarshaler interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

type TextFlag = FlagBase[TextMarshalUnmarshaler, StringConfig, textValue]

// -- TextMarshalUnmarshaler Value
type textValue struct {
	destination TextMarshalUnmarshaler
	trimSpace   bool
}

// Below functions are to satisfy the ValueCreator interface

func (t textValue) Create(val TextMarshalUnmarshaler, p *TextMarshalUnmarshaler, c StringConfig) Value {
	// Only overwrite the target when a non-nil default value is given, so that
	// a Destination pointing at an existing target is preserved (unlike the
	// concrete flag types, T here is an interface whose nil default would
	// otherwise clobber the destination).
	if val != nil {
		*p = val
	}
	return &textValue{
		destination: *p,
		trimSpace:   c.TrimSpace,
	}
}

func (t textValue) ToString(val TextMarshalUnmarshaler) string {
	if val == nil {
		return ""
	}
	text, err := val.MarshalText()
	if err != nil {
		return ""
	}
	return string(text)
}

// Below functions are to satisfy the flag.Value interface

func (t *textValue) Set(val string) error {
	if t.destination == nil {
		return nil
	}
	if t.trimSpace {
		val = strings.TrimSpace(val)
	}
	return t.destination.UnmarshalText([]byte(val))
}

func (t *textValue) Get() any { return t.destination }

func (t *textValue) String() string {
	if t.destination == nil {
		return ""
	}
	text, err := t.destination.MarshalText()
	if err != nil {
		return ""
	}
	return string(text)
}

// Text looks up the value of a local TextFlag, returns nil if not found
func (cmd *Command) Text(name string) TextMarshalUnmarshaler {
	if v, ok := cmd.Value(name).(TextMarshalUnmarshaler); ok {
		tracef("text available for flag name %[1]q with value=%[2]v (cmd=%[3]q)", name, v, cmd.Name)
		return v
	}

	tracef("text NOT available for flag name %[1]q (cmd=%[2]q)", name, cmd.Name)
	return nil
}
