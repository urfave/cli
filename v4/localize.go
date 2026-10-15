package cli

import (
	"maps"
	"slices"
	"strconv"
	"strings"
)

// Localizer renders a catalogue message. Keys are stable and listed by
// [Keys]; args fill the message's {placeholders}.
type Localizer interface {
	Message(key string, args Args) string
}

// Args are the values a message can refer to. Each field is available as a
// placeholder of the same name in lower case, e.g. {flag} or {count}.
type Args struct {
	Command string
	Flag    string
	Arg     string
	Value   string
	Choices string
	Cause   string
	Count   int
}

// Form is a CLDR plural category.
type Form uint8

const (
	Other Form = iota
	Zero
	One
	Two
	Few
	Many
)

// Text is one message in all the plural forms a language needs. Other is
// required; the rest fall back to it.
type Text struct {
	Zero, One, Two, Few, Many, Other string
}

func (t Text) form(f Form) string {
	var s string
	switch f {
	case Zero:
		s = t.Zero
	case One:
		s = t.One
	case Two:
		s = t.Two
	case Few:
		s = t.Few
	case Many:
		s = t.Many
	}
	if s == "" {
		return t.Other
	}
	return s
}

// Catalog is a map-based [Localizer] for one language.
type Catalog struct {
	Tag      string           // BCP 47 tag, e.g. "de" or "pt-BR"
	Plural   func(n int) Form // nil means every count uses Other
	Messages map[string]Text
	Fallback Localizer // used for missing keys; nil means English
}

// Message renders key, falling back when the catalogue lacks it.
func (c *Catalog) Message(key string, a Args) string {
	t, ok := c.Messages[key]
	if !ok {
		if c.Fallback != nil {
			return c.Fallback.Message(key, a)
		}
		if c != English {
			return English.Message(key, a)
		}
		return key
	}
	f := Other
	if c.Plural != nil {
		f = c.Plural(a.Count)
	}
	return expand(t.form(f), a)
}

// OneOther is the plural rule for English, German, Dutch, Swedish and other
// languages that only distinguish one from many.
func OneOther(n int) Form {
	if n == 1 {
		return One
	}
	return Other
}

func expand(s string, a Args) string {
	if !strings.Contains(s, "{") {
		return s
	}
	return strings.NewReplacer(
		"{command}", a.Command,
		"{flag}", a.Flag,
		"{arg}", a.Arg,
		"{value}", a.Value,
		"{choices}", a.Choices,
		"{cause}", a.Cause,
		"{count}", strconv.Itoa(a.Count),
	).Replace(s)
}

// English is the built-in catalogue and the final fallback for every other.
var English = &Catalog{
	Tag:    "en",
	Plural: OneOther,
	Messages: map[string]Text{
		"error.internal":        {Other: "{cause}"},
		"error.unknown_flag":    {Other: "unknown flag: {flag}"},
		"error.missing_value":   {Other: "flag needs a value: {flag}"},
		"error.invalid_value":   {Other: "invalid value \"{value}\" for flag {flag}: {cause}"},
		"error.required_flag":   {Other: "required flag not set: {flag}"},
		"error.required_arg":    {Other: "required argument not set: {arg}"},
		"error.flag_conflict":   {Other: "flag {flag} cannot be used with {choices}"},
		"error.flag_requires":   {Other: "flag {flag} requires {choices}"},
		"error.unknown_command": {Other: "unknown command \"{value}\""},
		"error.missing_command": {Other: "a command is required"},
		"error.too_many_args": {
			One:   "unexpected argument \"{value}\"",
			Other: "{count} unexpected arguments, starting with \"{value}\"",
		},

		"hint.try_help": {Other: "Try '{command} --help' for more information."},
		"hint.see_help": {Other: "See '{command} --help'."},
		"hint.similar":  {Other: "The most similar command is \"{choices}\"."},
	},
}

// Keys returns every key the library uses, sorted. A translation is complete
// when it covers all of them.
func Keys() []string {
	return slices.Sorted(maps.Keys(English.Messages))
}
