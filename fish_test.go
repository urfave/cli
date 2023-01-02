package cli

import (
	"bytes"
	"net/mail"
	"os"
	"testing"
)

func TestFishCompletion(t *testing.T) {
	// Given
	app := testApp()
	app.Flags = append(app.Flags, &StringFlag{
		Name:      "logfile",
		TakesFile: true,
	})

	// When
	res, err := app.ToFishCompletion()

	// Then
	expect(t, err, nil)
	expectFileContent(t, "testdata/expected-fish-full.fish", res)
}

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
		Commands: []*Command{{
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
		Commands: []*Command{{
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
	app.Authors = []any{
		"Harrison <harrison@lolwut.example.com>",
		&mail.Address{Name: "Oliver Allen", Address: "oliver@toyshop.com"},
	}
	return app
}

func expectFileContent(t *testing.T, file, got string) {
	data, err := os.ReadFile(file)
	// Ignore windows line endings
	// TODO: Replace with bytes.ReplaceAll when support for Go 1.11 is dropped
	data = bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)
	expect(t, err, nil)
	expect(t, got, string(data))
}
