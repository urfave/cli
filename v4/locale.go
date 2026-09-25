package cli

import "strings"

// DetectLocales returns the user's preferred message languages as BCP 47 tags,
// most preferred first, following POSIX and GNU gettext rules: LC_ALL, then
// LC_MESSAGES, then LANG pick the locale, and LANGUAGE, a colon-separated
// list, takes priority unless the locale is C or POSIX. It returns nil when
// the user has asked for the default language.
//
// lookup is usually [os.LookupEnv]; tests pass their own.
func DetectLocales(lookup func(string) (string, bool)) []string {
	var locale string
	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v, ok := lookup(name); ok && v != "" {
			locale = v
			break
		}
	}
	tag := normalize(locale)
	if tag == "" {
		return nil
	}

	var tags []string
	if list, ok := lookup("LANGUAGE"); ok {
		for _, l := range strings.Split(list, ":") {
			if t := normalize(l); t != "" {
				tags = append(tags, t)
			}
		}
	}
	return append(tags, tag)
}

// normalize turns a POSIX locale such as "de_DE.UTF-8@euro" into "de-DE".
// C and POSIX map to "".
func normalize(locale string) string {
	if i := strings.IndexAny(locale, ".@"); i >= 0 {
		locale = locale[:i]
	}
	if locale == "C" || locale == "POSIX" {
		return ""
	}
	return strings.ReplaceAll(locale, "_", "-")
}

// SelectLocalizer returns the catalogue that best matches the user's locale,
// or [English]. An exact tag match wins over a language-only match, so
// "pt-BR" prefers a pt-BR catalogue to a pt one.
func SelectLocalizer(lookup func(string) (string, bool), catalogs ...*Catalog) Localizer {
	for _, want := range DetectLocales(lookup) {
		if c := match(want, catalogs); c != nil {
			return c
		}
	}
	return English
}

func match(want string, catalogs []*Catalog) *Catalog {
	for _, c := range catalogs {
		if strings.EqualFold(c.Tag, want) {
			return c
		}
	}
	base, _, _ := strings.Cut(want, "-")
	for _, c := range catalogs {
		if strings.EqualFold(c.Tag, base) {
			return c
		}
	}
	return nil
}
