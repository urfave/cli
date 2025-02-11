package cli

import (
	"encoding"
	"strings"
)

type TextMarshalUnmarshaller interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

// TextFlag enables you to set types that satisfies [TextMarshalUnmarshaller] using flags such as log levels.
type TextFlag = FlagBase[TextMarshalUnmarshaller, StringConfig, TextValue]

type TextValue struct {
	Value  *TextMarshalUnmarshaller
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

func (f TextValue) Create(v TextMarshalUnmarshaller, p *TextMarshalUnmarshaller, c StringConfig) Value {
	if v != nil {
		*p = v
	}

	return &TextValue{
		Value:  p,
		Config: c,
	}
}

func (f TextValue) ToString(v TextMarshalUnmarshaller) string {
	text, err := v.MarshalText()
	if err != nil {
		return ""
	}

	return string(text)
}
