package cli

import (
	"testing"
)

// TestRegression tests a regression that was merged between versions 1.20.0 and 1.21.0
// The included app.Run line worked in 1.20.0, and then was broken in 1.21.0.
// Relevant PR: https://github.com/urfave/cli/pull/872
func TestVersionOneTwoOneRegression(t *testing.T) {
	testData := []struct {
		testCase    string
		appRunInput []string
	}{
		// assertion: empty input, when a required flag is present, errors
		{
			testCase:    "with_dash_dash",
			appRunInput: []string{"cli", "command", "--flagone", "flagvalue", "--", "docker", "image", "ls", "--no-trunc"},
		},
		{
			testCase:    "without_dash_dash",
			appRunInput: []string{"cli", "command", "--flagone", "flagvalue", "docker", "image", "ls", "--no-trunc"},
		},
	}
	for _, test := range testData {
		t.Run(test.testCase, func(t *testing.T) {
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
			err := app.Run(test.appRunInput)

			// assertions
			if err != nil {
				t.Errorf("did not expected an error, but there was one: %s", err)
			}
		})
	}
}
