package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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
	app.Flags = append(app.Flags, &BoolFlag{Name: "é"})

	for _, testCase := range []struct {
		provided, expected string
	}{
		{"", ""},
		{"a", "--another-flag"},
		{"hlp", "--help"},
		{"k", ""},
		{"s", "-s"},
		{"éé", "-é"},
	} {
		// When
		res := suggestFlag(app.Flags, testCase.provided, false)

		// Then
		assert.Equal(t, testCase.expected, res)
	}
}

func TestSuggestFlagHideHelp(t *testing.T) {
	// Given
	app := buildExtendedTestCommand()

	// When
	res := suggestFlag(app.Flags, "hlp", true)

	// Then
	assert.Equal(t, "--fl", res)
}

func TestSuggestFlagHiddenFlag(t *testing.T) {
	// Given
	flags := []Flag{
		&BoolFlag{Name: "internal-debug", Hidden: true},
		&BoolFlag{Name: "verbose"},
	}

	// When
	res := suggestFlag(flags, "internal-debu", true)

	// Then
	assert.Equal(t, "--verbose", res)
}

func TestSuggestFlagHiddenFlagFromRun(t *testing.T) {
	// Given
	var out bytes.Buffer
	cmd := &Command{
		Name:      "app",
		Suggest:   true,
		Writer:    &out,
		ErrWriter: &out,
		Flags: []Flag{
			&BoolFlag{Name: "internal-debug", Hidden: true},
			&BoolFlag{Name: "verbose"},
		},
		Action: func(context.Context, *Command) error { return nil },
	}

	// When
	err := cmd.Run(buildTestContext(t), []string{"app", "--internal-debu"})

	// Then
	assert.Error(t, err)
	assert.NotContains(t, out.String(), "internal-debug")
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

func TestSuggestFlagFromErrorAppliedFlags(t *testing.T) {
	for _, testCase := range []struct {
		name, provided, expected, unexpected string
	}{
		{name: "persistent flag of parent", provided: "--confg", expected: "--config"},
		{name: "mutually exclusive flag", provided: "--jsn", expected: "--json"},
		{name: "local flag of parent", provided: "--local-onli", unexpected: "--local-only"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			// Given
			var errOut bytes.Buffer
			cmd := &Command{
				Name:      "app",
				Writer:    io.Discard,
				ErrWriter: &errOut,
				Flags: []Flag{
					&StringFlag{Name: "config"},
					&StringFlag{Name: "local-only", Local: true},
				},
				Commands: []*Command{
					{
						Name:    "build",
						Suggest: true,
						MutuallyExclusiveFlags: []MutuallyExclusiveFlags{
							{Flags: [][]Flag{{&BoolFlag{Name: "json"}}, {&BoolFlag{Name: "yaml"}}}},
						},
						Action: func(context.Context, *Command) error { return nil },
					},
				},
			}

			// When
			err := cmd.Run(buildTestContext(t), []string{"app", "build", testCase.provided})

			// Then
			assert.Error(t, err)
			if testCase.expected != "" {
				assert.Contains(t, errOut.String(), fmt.Sprintf(SuggestDidYouMeanTemplate, testCase.expected))
			}
			if testCase.unexpected != "" {
				assert.NotContains(t, errOut.String(), fmt.Sprintf(SuggestDidYouMeanTemplate, testCase.unexpected))
			}
		})
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
