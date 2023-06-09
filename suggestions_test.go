package cli

import (
	"errors"
	"fmt"
	"testing"
)

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
		expect(t, res, testCase.expected)
	}
}

func TestSuggestFlagHideHelp(t *testing.T) {
	// Given
	app := buildExtendedTestCommand()

	// When
	res := suggestFlag(app.Flags, "hlp", true)

	// Then
	expect(t, res, "--fl")
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
		expect(t, res, fmt.Sprintf(SuggestDidYouMeanTemplate+"\n\n", testCase.expected))
	}
}

func TestSuggestFlagFromErrorWrongError(t *testing.T) {
	// Given
	app := buildExtendedTestCommand()

	// When
	_, err := app.suggestFlagFromError(errors.New("invalid"), "")

	// Then
	expect(t, true, err != nil)
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
	expect(t, true, err != nil)
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
	expect(t, true, err != nil)
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
		expect(t, res, testCase.expected)
	}
}
