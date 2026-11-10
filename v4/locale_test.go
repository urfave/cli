package cli_test

import (
	"slices"
	"testing"

	"github.com/urfave/cli/v4"
	"github.com/urfave/cli/v4/i18n/locale/de"
)

func env(kv ...string) func(string) (string, bool) {
	m := map[string]string{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func TestDetectLocales(t *testing.T) {
	tests := []struct {
		name string
		env  func(string) (string, bool)
		want []string
	}{
		{"nothing set", env(), nil},
		{"LANG", env("LANG", "de_DE.UTF-8"), []string{"de-DE"}},
		{"modifier", env("LANG", "de_DE@euro"), []string{"de-DE"}},
		{"LC_MESSAGES beats LANG", env("LANG", "fr_FR", "LC_MESSAGES", "de_AT"), []string{"de-AT"}},
		{"LC_ALL beats both", env("LANG", "fr_FR", "LC_MESSAGES", "de_AT", "LC_ALL", "pt_BR.UTF-8"), []string{"pt-BR"}},
		{"empty LC_ALL is skipped", env("LC_ALL", "", "LANG", "de_DE"), []string{"de-DE"}},
		{"C means default", env("LANG", "C.UTF-8"), nil},
		{"POSIX means default", env("LC_ALL", "POSIX"), nil},
		{"LANGUAGE list first", env("LANG", "en_US", "LANGUAGE", "de:fr_FR"), []string{"de", "fr-FR", "en-US"}},
		{"LANGUAGE ignored under C", env("LANG", "C", "LANGUAGE", "de"), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cli.DetectLocales(tt.env); !slices.Equal(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSelectLocalizer(t *testing.T) {
	deAT := &cli.Catalog{Tag: "de-AT"}
	tests := []struct {
		name string
		env  func(string) (string, bool)
		cats []*cli.Catalog
		want cli.Localizer
	}{
		{"no locale", env(), []*cli.Catalog{de.Catalog}, cli.English},
		{"language match", env("LANG", "de_CH.UTF-8"), []*cli.Catalog{de.Catalog}, de.Catalog},
		{"exact beats language", env("LANG", "de_AT"), []*cli.Catalog{de.Catalog, deAT}, deAT},
		{"no match", env("LANG", "ja_JP"), []*cli.Catalog{de.Catalog}, cli.English},
		{"LANGUAGE order", env("LANG", "ja_JP", "LANGUAGE", "ja:de"), []*cli.Catalog{de.Catalog}, de.Catalog},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cli.SelectLocalizer(tt.env, tt.cats...); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCatalogFallsBackToEnglish(t *testing.T) {
	partial := &cli.Catalog{Tag: "xx", Messages: map[string]cli.Text{}}
	got := partial.Message("error.unknown_flag", cli.Args{Flag: "z"})
	if got != "unknown flag: z" {
		t.Errorf("got %q", got)
	}
}
