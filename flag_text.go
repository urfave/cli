package cli

import (
	"encoding"
)

type TextMarshalUnmarshaler interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

// TextFlag enables you to set types that satisfies [TextMarshalUnmarshaler] using flags such as log levels.
type TextFlag = FlagBase[TextMarshalUnmarshaler, NoConfig, TextValue]

type TextValue struct {
	Value TextMarshalUnmarshaler
}

func (f TextValue) String() string {
	text, err := f.Value.MarshalText()
	if err != nil {
		return ""
	}

	return string(text)
}

func (f TextValue) Set(s string) error {
	return f.Value.UnmarshalText([]byte(s))
}

func (f TextValue) Get() any {
	return f.Value
}

func (f TextValue) Create(v TextMarshalUnmarshaler, p *TextMarshalUnmarshaler, _ NoConfig) Value {
	pp := *p
	if v != nil {
		if b, err := v.MarshalText(); err == nil {
			_ = pp.UnmarshalText(b)
		}
	}

	return &TextValue{
		Value: pp,
	}
}

func (f TextValue) ToString(v TextMarshalUnmarshaler) string {
	text, _ := v.MarshalText()

	return string(text)
}
