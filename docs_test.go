package cli

import (
	"testing"
)

func TestToMarkdownFull(t *testing.T) {
	// Given
	app := testApp()

	// When
	res, err := app.ToMarkdown()

	// Then
	expect(t, err, nil)
	expectFileContent(t, "testdata/expected-doc-full.md", res)
}

func TestToMarkdownNoFlags(t *testing.T) {
	// Given
	app := testApp()
	app.Flags = nil

	// When
	res, err := app.ToMarkdown()

	// Then
	expect(t, err, nil)
	expectFileContent(t, "testdata/expected-doc-no-flags.md", res)
}

func TestToMarkdownNoCommands(t *testing.T) {
	// Given
	app := testApp()
	app.Commands = nil

	// When
	res, err := app.ToMarkdown()

	// Then
	expect(t, err, nil)
	expectFileContent(t, "testdata/expected-doc-no-commands.md", res)
}

func TestToMan(t *testing.T) {
	// Given
	app := testApp()

	// When
	res, err := app.ToMan()

	// Then
	expect(t, err, nil)
	expectFileContent(t, "testdata/expected-doc-full.man", res)
}
