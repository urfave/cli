package cli

import (
	"testing"
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
