// Package de is the German catalogue. It is plain Go data, so importing it
// adds no JSON decoder to the binary.
package de

import "github.com/urfave/cli/v4"

var Catalog = &cli.Catalog{
	Tag:    "de",
	Plural: cli.OneOther,
	Messages: map[string]cli.Text{
		"error.internal":        {Other: "{cause}"},
		"error.unknown_flag":    {Other: "unbekanntes Flag: {flag}"},
		"error.missing_value":   {Other: "Flag benötigt einen Wert: {flag}"},
		"error.invalid_value":   {Other: "ungültiger Wert „{value}“ für Flag {flag}: {cause}"},
		"error.required_flag":   {Other: "erforderliches Flag nicht gesetzt: {flag}"},
		"error.required_arg":    {Other: "erforderliches Argument nicht gesetzt: {arg}"},
		"error.flag_conflict":   {Other: "Flag {flag} kann nicht zusammen mit {choices} verwendet werden"},
		"error.flag_requires":   {Other: "Flag {flag} erfordert {choices}"},
		"error.unknown_command": {Other: "unbekannter Befehl „{value}“"},
		"error.missing_command": {Other: "ein Befehl ist erforderlich"},
		"error.too_many_args": {
			One:   "unerwartetes Argument „{value}“",
			Other: "{count} unerwartete Argumente, beginnend mit „{value}“",
		},
		"error.help":    {Other: "Hilfe angefordert"},
		"error.version": {Other: "Version angefordert"},

		"hint.try_help": {Other: "„{command} --help“ gibt weitere Informationen aus."},
		"hint.see_help": {Other: "Siehe „{command} --help“."},
		"hint.similar":  {Other: "Der ähnlichste Befehl ist „{choices}“."},
	},
}
