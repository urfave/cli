// Package i18n loads message catalogues from JSON. The core already detects
// the user's locale and ships English; import this package only to load
// translations from files, for example ones you embed with go:embed.
//
// A catalogue file maps each key from [cli.Keys] to a string, or to an object
// of plural forms:
//
//	{
//	  "error.unknown_flag": "unbekanntes Flag: {flag}",
//	  "error.too_many_args": {"one": "...", "other": "..."}
//	}
package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/urfave/cli/v4"
)

// Parse builds a catalogue for tag from JSON data. plural picks the form for
// a count; nil means every count uses Other.
func Parse(tag string, plural func(int) cli.Form, data []byte) (*cli.Catalog, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("i18n: parse %s: %w", tag, err)
	}
	c := &cli.Catalog{Tag: tag, Plural: plural, Messages: make(map[string]cli.Text, len(raw))}
	for key, msg := range raw {
		var s string
		if json.Unmarshal(msg, &s) == nil {
			c.Messages[key] = cli.Text{Other: s}
			continue
		}
		var forms struct{ Zero, One, Two, Few, Many, Other string }
		if err := json.Unmarshal(msg, &forms); err != nil {
			return nil, fmt.Errorf("i18n: parse %s: key %q: %w", tag, key, err)
		}
		if forms.Other == "" {
			return nil, fmt.Errorf("i18n: parse %s: key %q has no \"other\" form", tag, key)
		}
		c.Messages[key] = cli.Text(forms)
	}
	return c, nil
}

// Load reads name from fsys and parses it as the catalogue for tag.
func Load(fsys fs.FS, name, tag string, plural func(int) cli.Form) (*cli.Catalog, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("i18n: %w", err)
	}
	return Parse(tag, plural, data)
}

// Missing returns the keys from [cli.Keys] that c does not translate.
func Missing(c *cli.Catalog) []string {
	var out []string
	for _, k := range cli.Keys() {
		if _, ok := c.Messages[k]; !ok {
			out = append(out, k)
		}
	}
	return out
}
