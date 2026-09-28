package cli

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJaroWinkler(t *testing.T) {
	// Given
	for _, testCase := range []struct {
		a, b     string
		expected float64
	}{
		{"", "", 1},
		{"a", "", 0},
		{"", "a", 0},
		{"a", "a", 1},
		{"a", "b", 0},
		{"aa", "aa", 1},
		{"aa", "bb", 0},
		{"aaa", "aaa", 1},
		{"aa", "ab", 0.6666666666666666},
		{"aa", "ba", 0.6666666666666666},
		{"ba", "aa", 0.6666666666666666},
		{"ab", "aa", 0.6666666666666666},
	} {
		// When
		res := jaroWinkler(testCase.a, testCase.b)

		// Then
		assert.Equal(t, testCase.expected, res)
	}
}

func TestSuggestFlag(t *testing.T) {
	// Given
	app := buildExtendedTestCommand()

	for _, testCase := range []struct {
		provided, expected string
	}{
		{"", ""},
		{"a", "--another-flag"},
		{"hlp", "--help"},
		{"k", ""},
		{"s", "-s"},
	} {
		// When
		res := suggestFlag(app.Flags, testCase.provided, false)

		// Then
		assert.Equal(t, testCase.expected, res)
	}
}

// TestSuggestFlagMultibyteRunePrefix ensures the "-"/"--" prefix chosen for a
// suggested flag name is derived from the rune count, exactly like the prefix
// used when the very same flag is rendered in help output by prefixFor. A
// single (non-ASCII) rune flag such as "é" is a short flag and must be
// suggested as "-é", not "--é".
func TestSuggestFlagMultibyteRunePrefix(t *testing.T) {
	for _, testCase := range []struct {
		name, provided, expected string
	}{
		// Single ASCII rune: short flag.
		{"single-rune-ascii", "ss", "-s"},
		// Single multi-byte rune: still a short flag.
		{"single-rune-multibyte", "éé", "-é"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fl := &BoolFlag{Name: "é"}
			if testCase.name == "single-rune-ascii" {
				fl = &BoolFlag{Name: "s"}
			}

			// The suggestion must agree with how the flag is advertised in
			// help output, which is driven by prefixFor.
			assert.Equal(t, "-"+fl.Names()[0], flagStringForTest(fl))

			assert.Equal(t, testCase.expected, suggestFlag([]Flag{fl}, testCase.provided, true))
		})
	}
}

// TestSuggestFlagMultibyteRuneFromError covers the user-visible output: the
// "Did you mean ...?" hint must name the short flag exactly as help output
// does.
func TestSuggestFlagMultibyteRuneFromError(t *testing.T) {
	cmd := &Command{Flags: []Flag{&BoolFlag{Name: "é"}}}

	res, err := cmd.suggestFlagFromError(
		errors.New(providedButNotDefinedErrMsg+"éé"),
		"",
	)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(SuggestDidYouMeanTemplate+"\n\n", "-é"), res)
	assert.Equal(t, "Did you mean \"-é\"?\n\n", res)
}

// flagStringForTest returns the leading flag-name portion of a flag's help
// rendering, e.g. "-é" or "--verbose".
func flagStringForTest(fl Flag) string {
	s := stringifyFlag(fl)
	if i := strings.IndexByte(s, '\t'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

func TestSuggestFlagHideHelp(t *testing.T) {
	// Given
	app := buildExtendedTestCommand()

	// When
	res := suggestFlag(app.Flags, "hlp", true)

	// Then
	assert.Equal(t, "--fl", res)
}

func TestSuggestFlagFromError(t *testing.T) {
	// Given
	app := buildExtendedTestCommand()

	for _, testCase := range []struct {
		command, provided, expected string
	}{
		{"", "hel", "--help"},
		{"", "soccer", "--socket"},
		{"config", "anot", "--another-flag"},
	} {
		// When
		res, _ := app.suggestFlagFromError(
			errors.New(providedButNotDefinedErrMsg+testCase.provided),
			testCase.command,
		)

		// Then
		assert.Equal(t, fmt.Sprintf(SuggestDidYouMeanTemplate+"\n\n", testCase.expected), res)
	}
}

func TestSuggestFlagFromErrorWrongError(t *testing.T) {
	// Given
	app := buildExtendedTestCommand()

	// When
	_, err := app.suggestFlagFromError(errors.New("invalid"), "")

	// Then
	assert.Error(t, err)
}

func TestSuggestFlagFromErrorWrongCommand(t *testing.T) {
	// Given
	app := buildExtendedTestCommand()

	// When
	_, err := app.suggestFlagFromError(
		errors.New(providedButNotDefinedErrMsg+"flag"),
		"invalid",
	)

	// Then
	assert.Error(t, err)
}

func TestSuggestFlagFromErrorNoSuggestion(t *testing.T) {
	// Given
	app := buildExtendedTestCommand()

	// When
	_, err := app.suggestFlagFromError(
		errors.New(providedButNotDefinedErrMsg+""),
		"",
	)

	// Then
	assert.Error(t, err)
}

func TestSuggestCommand(t *testing.T) {
	// Given
	app := buildExtendedTestCommand()

	for _, testCase := range []struct {
		provided, expected string
	}{
		{"", ""},
		{"conf", "config"},
		{"i", "i"},
		{"information", "info"},
		{"inf", "info"},
		{"con", "config"},
		{"not-existing", "info"},
	} {
		// When
		res := suggestCommand(app.Commands, testCase.provided)

		// Then
		assert.Equal(t, testCase.expected, res)
	}
}
