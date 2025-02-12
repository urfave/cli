package cli

import (
	"encoding"
	"strings"
)

type TextMarshalUnmarshaler interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

// TextFlag enables you to set types that satisfies [TextMarshalUnmarshaler] using flags such as log levels.
type TextFlag = FlagBase[TextMarshalUnmarshaler, StringConfig, TextValue]

type TextValue struct {
	Value  *TextMarshalUnmarshaler
	Config StringConfig
}

func (f TextValue) String() string {
	text, err := (*f.Value).MarshalText()
	if err != nil {
		return ""
	}

	return string(text)
}

func (f TextValue) Set(s string) error {
	if f.Config.TrimSpace {
		s = strings.TrimSpace(s)
	}

	return (*f.Value).UnmarshalText([]byte(s))
}

func (f TextValue) Get() any {
	return *f.Value
}

func (f TextValue) Create(v TextMarshalUnmarshaler, p *TextMarshalUnmarshaler, c StringConfig) Value {
	if v != nil {
		b, _ := v.MarshalText()

		_ = (*p).UnmarshalText(b)
	}

	return &TextValue{
		Value:  p,
		Config: c,
	}
}

func (f TextValue) ToString(v TextMarshalUnmarshaler) string {
	text, err := v.MarshalText()
	if err != nil {
		return ""
	}

	return string(text)
}
