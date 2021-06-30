package cli

import (
	"bytes"
	"errors"
	"io/ioutil"
	"testing"
)

func testApp() *App {
	app := newTestApp()
	app.Name = "greet"
	app.Flags = []Flag{
		&StringFlag{
			Name:      "socket",
			Aliases:   []string{"s"},
			Usage:     "some 'usage' text",
			Value:     "value",
			TakesFile: true,
		},
		&StringFlag{Name: "flag", Aliases: []string{"fl", "f"}},
		&BoolFlag{
			Name:    "another-flag",
			Aliases: []string{"b"},
			Usage:   "another usage text",
		},
		&BoolFlag{
			Name:   "hidden-flag",
			Hidden: true,
		},
	}
	app.Commands = []*Command{{
		Aliases: []string{"c"},
		Flags: []Flag{
			&StringFlag{
				Name:      "flag",
				Aliases:   []string{"fl", "f"},
				TakesFile: true,
			},
			&BoolFlag{
				Name:    "another-flag",
				Aliases: []string{"b"},
				Usage:   "another usage text",
			},
		},
		Name:  "config",
		Usage: "another usage test",
		Subcommands: []*Command{{
			Aliases: []string{"s", "ss"},
			Flags: []Flag{
				&StringFlag{Name: "sub-flag", Aliases: []string{"sub-fl", "s"}},
				&BoolFlag{
					Name:    "sub-command-flag",
					Aliases: []string{"s"},
					Usage:   "some usage text",
				},
			},
			Name:  "sub-config",
			Usage: "another usage test",
		}},
	}, {
		Aliases: []string{"i", "in"},
		Name:    "info",
		Usage:   "retrieve generic information",
	}, {
		Name: "some-command",
	}, {
		Name:   "hidden-command",
		Hidden: true,
	}, {
		Aliases: []string{"u"},
		Flags: []Flag{
			&StringFlag{
				Name:      "flag",
				Aliases:   []string{"fl", "f"},
				TakesFile: true,
			},
			&BoolFlag{
				Name:    "another-flag",
				Aliases: []string{"b"},
				Usage:   "another usage text",
			},
		},
		Name:  "usage",
		Usage: "standard usage text",
		UsageText: `
Usage for the usage text
- formatted:  Based on the specified ConfigMap and summon secrets.yml
- list:       Inspect the environment for a specific process running on a Pod
- for_effect: Compare 'namespace' environment with 'local'

` + "```" + `
func() { ... }
` + "```" + `

Should be a part of the same code block
`,
		Subcommands: []*Command{{
			Aliases: []string{"su"},
			Flags: []Flag{
				&BoolFlag{
					Name:    "sub-command-flag",
					Aliases: []string{"s"},
					Usage:   "some usage text",
				},
			},
			Name:      "sub-usage",
			Usage:     "standard usage text",
			UsageText: "Single line of UsageText",
		}},
	}}
	app.UsageText = "app [first_arg] [second_arg]"
	app.Description = `Description of the application.`
	app.Usage = "Some app"
	app.Authors = []*Author{
		{Name: "Harrison", Email: "harrison@lolwut.com"},
		{Name: "Oliver Allen", Email: "oliver@toyshop.com"},
	}
	return app
}

func expectFileContent(t *testing.T, file, expected string) {
	data, err := ioutil.ReadFile(file)
	// Ignore windows line endings
	// TODO: Replace with bytes.ReplaceAll when support for Go 1.11 is dropped
	data = bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)
	expect(t, err, nil)
	expect(t, string(data), expected)
}

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
