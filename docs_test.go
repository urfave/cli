//go:build !urfave_cli_no_docs
// +build !urfave_cli_no_docs

package cli

import (
	"errors"
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

func TestToMarkdownNoAuthors(t *testing.T) {
	// Given
	app := testApp()
	app.Authors = []*Author{}

	// When
	res, err := app.ToMarkdown()

	// Then
	expect(t, err, nil)
	expectFileContent(t, "testdata/expected-doc-no-authors.md", res)
}

func TestToMarkdownNoUsageText(t *testing.T) {
	// Given
	app := testApp()
	app.UsageText = ""

	// When
	res, err := app.ToMarkdown()

	// Then
	expect(t, err, nil)
	expectFileContent(t, "testdata/expected-doc-no-usagetext.md", res)
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

func TestToManParseError(t *testing.T) {
	// Given
	app := testApp()

	// When
	// temporarily change the global variable for testing
	tmp := MarkdownDocTemplate
	MarkdownDocTemplate = `{{ .App.Name`
	_, err := app.ToMan()
	MarkdownDocTemplate = tmp

	// Then
	expect(t, err, errors.New(`template: cli:1: unclosed action`))
}

func TestToManWithSection(t *testing.T) {
	// Given
	app := testApp()

	// When
	res, err := app.ToManWithSection(8)

	// Then
	expect(t, err, nil)
	expectFileContent(t, "testdata/expected-doc-full.man", res)
}

func Test_prepareUsageText(t *testing.T) {
	t.Run("no UsageText provided", func(t *testing.T) {
		// Given
		cmd := Command{}

		// When
		res := prepareUsageText(&cmd)

		// Then
		expect(t, res, "")
	})

	t.Run("single line UsageText", func(t *testing.T) {
		// Given
		cmd := Command{UsageText: "Single line usage text"}

		// When
		res := prepareUsageText(&cmd)

		// Then
		expect(t, res, ">Single line usage text\n")
	})

	t.Run("multiline UsageText", func(t *testing.T) {
		// Given
		cmd := Command{
			UsageText: `
Usage for the usage text
- Should be a part of the same code block
`,
		}

		// When
		res := prepareUsageText(&cmd)

		// Then
		test := `    Usage for the usage text
    - Should be a part of the same code block
`
		expect(t, res, test)
	})

	t.Run("multiline UsageText has formatted embedded markdown", func(t *testing.T) {
		// Given
		cmd := Command{
			UsageText: `
Usage for the usage text

` + "```" + `
func() { ... }
` + "```" + `

Should be a part of the same code block
`,
		}

		// When
		res := prepareUsageText(&cmd)

		// Then
		test := `    Usage for the usage text
    
    ` + "```" + `
    func() { ... }
    ` + "```" + `
    
    Should be a part of the same code block
`
		expect(t, res, test)
	})
}

func Test_prepareUsage(t *testing.T) {
	t.Run("no Usage provided", func(t *testing.T) {
		// Given
		cmd := Command{}

		// When
		res := prepareUsage(&cmd, "")

		// Then
		expect(t, res, "")
	})

	t.Run("simple Usage", func(t *testing.T) {
		// Given
		cmd := Command{Usage: "simple usage text"}

		// When
		res := prepareUsage(&cmd, "")

		// Then
		expect(t, res, cmd.Usage+"\n")
	})

	t.Run("simple Usage with UsageText", func(t *testing.T) {
		// Given
		cmd := Command{Usage: "simple usage text"}

		// When
		res := prepareUsage(&cmd, "a non-empty string")

		// Then
		expect(t, res, cmd.Usage+"\n\n")
	})
}
