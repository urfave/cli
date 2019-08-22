package cli

import (
	"testing"
)

func TestFishCompletion(t *testing.T) {
	// Given
	app := testApp()

	// When
	res, err := app.ToFishCompletion()

	// Then
	expect(t, err, nil)
	expectFileContent(t, "testdata/expected-fish-full.fish", res)
}
