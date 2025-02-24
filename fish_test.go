package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFishCompletion(t *testing.T) {
	// Given
	app := testApp()
	app.Flags = append(app.Flags, &PathFlag{
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
			Name:        "socket",
			Aliases:     []string{"s"},
			Usage:       "some 'usage' text",
			Value:       "value",
			DefaultText: "/some/path",
			TakesFile:   true,
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
	data, err := os.ReadFile(file)
	if err != nil {
		t.FailNow()
	}

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(strings.ReplaceAll(string(data), "\r\n", "\n"))

	if expected != actual {
		t.Logf("file %q content does not match expected", file)

		tryDiff(t, expected, actual)

		t.FailNow()
	}
}

func tryDiff(t *testing.T, a, b string) {
	diff, err := exec.LookPath("diff")
	if err != nil {
		t.Logf("no diff tool available")

		return
	}

	td := t.TempDir()
	aPath := filepath.Join(td, "a")
	bPath := filepath.Join(td, "b")

	if err := os.WriteFile(aPath, []byte(a), 0o0644); err != nil {
		t.Logf("failed to write: %v", err)
		t.FailNow()

		return
	}

	if err := os.WriteFile(bPath, []byte(b), 0o0644); err != nil {
		t.Logf("failed to write: %v", err)
		t.FailNow()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, diff, "-u", aPath, bPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	_ = cmd.Run()
}
