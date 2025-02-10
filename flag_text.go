package cli

import (
	"encoding"
	"strings"
)

type TextMarshalUnMarshaller interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

// TextFlag enables you to set types that satisfies [TextMarshalUnMarshaller] using flags such as log levels.
type TextFlag = FlagBase[TextMarshalUnMarshaller, StringConfig, TextValue]

type TextValue struct {
	Value  TextMarshalUnMarshaller
	Config StringConfig
}

func (v TextValue) String() string {
	text, err := v.Value.MarshalText()
	if err != nil {
		return ""
	}

	return string(text)
}

func (v TextValue) Set(s string) error {
	if v.Config.TrimSpace {
		return v.Value.UnmarshalText([]byte(strings.TrimSpace(s)))
	}

	return v.Value.UnmarshalText([]byte(s))
}

func (v TextValue) Get() any {
	return v.Value
}

func (v TextValue) Create(t TextMarshalUnMarshaller, _ *TextMarshalUnMarshaller, c StringConfig) Value {
	return &TextValue{
		Value:  t,
		Config: c,
	}
}

func (v TextValue) ToString(t TextMarshalUnMarshaller) string {
	text, err := t.MarshalText()
	if err != nil {
		return ""
	}

	return string(text)
}
