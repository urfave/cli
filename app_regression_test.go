package cli

import (
	"testing"
)

// TestRegression tests a regression that was merged between versions 1.20.0 and 1.21.0
// The included app.Run line worked in 1.20.0, and then was broken in 1.21.0.
func TestRegression(t *testing.T) {
	// setup
	app := NewApp()
	app.Commands = []Command{{
		Name: "command",
		Flags: []Flag{
			StringFlag{
				Name: "flagone",
			},
		},
		Action: func(c *Context) error { return nil },
	}}

	// logic under test
	err := app.Run([]string{"cli", "command", "--flagone", "flagvalue", "docker", "image", "ls", "--no-trunc"})

	// assertions
	if err != nil {
		t.Errorf("did not expected an error, but there was one: %s", err)
	}
}
