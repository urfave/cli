package i18n_test

import (
	"testing"
	"testing/fstest"

	"github.com/urfave/cli/v4"
	"github.com/urfave/cli/v4/i18n"
	"github.com/urfave/cli/v4/i18n/locale/de"
)

func TestLoad(t *testing.T) {
	fsys := fstest.MapFS{"fr.json": {Data: []byte(`{
		"error.unknown_flag": "option inconnue : {flag}",
		"error.too_many_args": {"one": "argument inattendu « {value} »", "other": "{count} arguments inattendus"}
	}`)}}
	// French treats 0 and 1 as singular.
	plural := func(n int) cli.Form {
		if n == 0 || n == 1 {
			return cli.One
		}
		return cli.Other
	}
	c, err := i18n.Load(fsys, "fr.json", "fr", plural)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		key  string
		args cli.Args
		want string
	}{
		{"error.unknown_flag", cli.Args{Flag: "z"}, "option inconnue : z"},
		{"error.too_many_args", cli.Args{Value: "a", Count: 1}, "argument inattendu « a »"},
		{"error.too_many_args", cli.Args{Value: "a", Count: 4}, "4 arguments inattendus"},
		{"error.required_flag", cli.Args{Flag: "name"}, "required flag not set: name"}, // English fallback
	}
	for _, tt := range tests {
		if got := c.Message(tt.key, tt.args); got != tt.want {
			t.Errorf("%s(%d): got %q, want %q", tt.key, tt.args.Count, got, tt.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, data := range []string{`not json`, `{"k": {"one": "x"}}`, `{"k": 3}`} {
		if _, err := i18n.Parse("xx", nil, []byte(data)); err == nil {
			t.Errorf("Parse(%s) returned no error", data)
		}
	}
}

func TestShippedLocalesAreComplete(t *testing.T) {
	for _, c := range []*cli.Catalog{de.Catalog} {
		if missing := i18n.Missing(c); len(missing) > 0 {
			t.Errorf("%s is missing %v", c.Tag, missing)
		}
	}
}
