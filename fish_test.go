package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func testApp() *App {
	app := NewApp()
	app.Name = "greet"
	app.Flags = []Flag{
		StringFlag{
			Name:      "socket, s",
			Usage:     "some 'usage' text",
			Value:     "value",
			TakesFile: true,
		},
		StringFlag{Name: "flag, fl, f"},
		BoolFlag{
			Name:  "another-flag, b",
			Usage: "another usage text",
		},
	}
	app.Commands = []Command{{
		Aliases: []string{"c"},
		Flags: []Flag{
			StringFlag{
				Name:      "flag, fl, f",
				TakesFile: true,
			},
			BoolFlag{
				Name:  "another-flag, b",
				Usage: "another usage text",
			},
		},
		Name:  "config",
		Usage: "another usage test",
		Subcommands: []Command{{
			Aliases: []string{"s", "ss"},
			Flags: []Flag{
				StringFlag{Name: "sub-flag, sub-fl, s"},
				BoolFlag{
					Name:  "sub-command-flag, s",
					Usage: "some usage text",
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
	}}
	app.UsageText = "app [first_arg] [second_arg]"
	app.Usage = "Some app"
	app.Author = "Harrison"
	app.Email = "harrison@lolwut.com"
	app.Authors = []Author{{Name: "Oliver Allen", Email: "oliver@toyshop.com"}}
	return app
}

func expectFileContent(t *testing.T, file, expected string) {
	data, err := os.ReadFile(file)
	require.Nil(t, err)

	require.Equal(t, strings.ReplaceAll(string(data), "\r\n", "\n"), expected)
}

func TestFishCompletion(t *testing.T) {
	// Given
	app := testApp()

	// When
	res, err := app.ToFishCompletion()

	// Then
	expect(t, err, nil)
	expectFileContent(t, "testdata/expected-fish-full.fish", res)
}
