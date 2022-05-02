package genflags

import (
	_ "embed"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	//go:embed generated.gotmpl
	TemplateString string

	//go:embed generated_test.gotmpl
	TestTemplateString string

	titler = cases.Title(language.Und, cases.NoLower)
)

func TypeName(goType string, fc *FlagTypeConfig) string {
	if fc != nil && strings.TrimSpace(fc.TypeName) != "" {
		return strings.TrimSpace(fc.TypeName)
	}

	dotSplit := strings.Split(goType, ".")
	goType = dotSplit[len(dotSplit)-1]

	if strings.HasPrefix(goType, "[]") {
		return titler.String(strings.TrimPrefix(goType, "[]")) + "SliceFlag"
	}

	return titler.String(goType) + "Flag"
}
